package application

import (
	"context"
	"github.com/Auvitly/application/internal/types"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

// Application - implements the start of services and their completion
type Application struct {
	// contains a list of registered constructors
	constructors []Constructor
	// contains a list of started services
	services []Service
	// contains a list of started resources
	resources []io.Closer
	// current application state
	state State
	// application launch configuration
	config *Config
	// log for application
	log *log.Logger

	// The channel defining initialization status
	initCh chan types.OperationResult
	// The channel that determines the application's exit status
	shutdownCh chan types.OperationResult
	// The channel that determines whether all services are running and the application has started
	runCh chan struct{}
	// A channel that allows you to intercept the error of one service
	errCh chan error

	// The channel is created to negotiate application termination via system calls
	exitCh chan os.Signal
}

func New(config *Config) *Application {
	app := &Application{
		config:     config,
		log:        log.Default(),
		initCh:     make(chan types.OperationResult),
		shutdownCh: make(chan types.OperationResult),
		runCh:      make(chan struct{}),
		errCh:      make(chan error),
		exitCh:     make(chan os.Signal),
	}

	signal.Notify(app.exitCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	return app
}

// RegistrationService - registering Constructor with internally initialized dependencies
func (app *Application) RegistrationService(constructors ...Constructor) (err error) {
	if app.state != StateInit {
		return ErrWrongState
	}
	app.constructors = append(app.constructors, constructors...)
	return nil
}

// RegistrationResource - registering resource Destructors
func (app *Application) RegistrationResource(resources ...io.Closer) (err error) {
	if app.state != StateInit {
		return ErrWrongState
	}

	for i := range resources {
		var isContain bool
		for j := range app.resources {
			if resources[i] == app.resources[j] {
				isContain = true
				break
			}
		}
		if !isContain {
			app.resources = append(app.resources, resources[i])
		}
	}

	return nil
}

// Init - performs initialization of registered constructors
func (app *Application) Init(ctx context.Context) (err error) {
	if app.state != StateInit {
		return ErrWrongState
	}

	initCtx, initCtxCancel := context.WithTimeout(context.Background(), app.config.InitialisationTimeout)
	defer initCtxCancel()

	go app.init(ctx)

	err = func() error {
		for {
			select {
			case result := <-app.initCh:
				switch result {
				case types.ResultSuccess:
					return nil
				case types.ResultError:
					return ErrInitFailure
				default:
				}
			case <-ctx.Done():
				return ErrInitContextDeadline
			case <-initCtx.Done():
				return ErrInitTimeout
			case <-app.exitCh:
				return ErrInitConstructorPanic
			}
		}
	}()
	if err != nil {
		return err
	}
	close(app.initCh)

	app.state = StateReady
	return nil
}

func (app *Application) init(ctx context.Context) {
	defer app.recover()

	for i := range app.constructors {
		var service Service
		var err error
		service, err = app.constructors[i](ctx)
		if err != nil {
			app.initCh <- types.ResultError
		}
		app.services = append(app.services, service)
	}
	app.initCh <- types.ResultSuccess
}

// Run - launching the ready application
func (app *Application) Run(ctx context.Context) (err error) {
	if app.state != StateReady {
		return ErrWrongState
	}

	go app.run()
	defer func() {
		go app.Shutdown()
	}()

	app.state = StateRunning

	for {
		select {
		case signal := <-app.exitCh:
			if signal == types.SIGPANIC {
				return ErrRunPanic
			}
			return nil
		case <-ctx.Done():
			return ErrRunContextDeadline
		case err = <-app.errCh:
			return err
		default:
		}
	}

}

func (app *Application) run() {
	// Start all services with error handling
	for i := range app.services {
		go func() {
			defer app.recover()
			if err := app.services[i].Serve(); err != nil {
				app.errCh <- err
			}
		}()
	}
}

// recover - panic detection and processing system
func (app *Application) recover() {
	if err := recover(); err != nil {
		if app.config.DebugStack {
			app.log.Println(err, string(debug.Stack()))
		} else {
			app.log.Println(err)
		}
		app.exitCh <- types.SIGPANIC
	}
}

// Shutdown - shutdown the application
func (app *Application) Shutdown() (err error) {
	app.state = StateShutdown

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), app.config.TerminationTimeout)
	defer shutdownCancel()

	go app.shutdown()

	err = func() error {
		for {
			select {
			case <-app.shutdownCh:
				return nil
			case <-shutdownCtx.Done():
				return ErrTerminateTimeout
			}
		}
	}()

	app.state = StateOff
	return err
}

func (app *Application) shutdown() {
	for i := range app.services {
		err := app.services[i].Close()
		if err != nil {
			app.log.Println(err)
			continue
		}
	}
	for i := range app.resources {
		err := app.resources[i].Close()
		if err != nil {
			app.log.Println(err)
			continue
		}
	}
	app.shutdownCh <- types.ResultSuccess
}
