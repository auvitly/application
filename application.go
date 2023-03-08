package application

import (
	"context"
	"errors"
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
	log Logger

	// The channel defining initialization status
	initCh chan types.OperationResult
	// The channel that determines the application's exit status
	shutdownCh chan types.OperationResult
	// The channel that determines whether all services are running and the application has started
	runCh chan struct{}
}

var defaultTerminateSyscall = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

// The channel is created to negotiate application termination via system calls
var exitCh = make(chan os.Signal)

// A channel that allows you to intercept the error of one service
var errCh = make(chan error)

// New - creating an application instance
func New(config *Config) *Application {
	app := &Application{
		config:     config,
		log:        log.Default(),
		initCh:     make(chan types.OperationResult),
		shutdownCh: make(chan types.OperationResult),
		runCh:      make(chan struct{}),
	}

	return app
}

// SetLogger sets the logger for package output
func (app *Application) SetLogger(logger Logger) {
	if logger != nil {
		app.log = logger
	}
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
func (app *Application) Init(ctx context.Context, signals ...os.Signal) (err error) {
	if app.state != StateInit {
		return ErrWrongState
	}

	var (
		initCtx       context.Context
		initCtxCancel context.CancelFunc
	)
	if app.config.InitialisationTimeout != 0 {
		initCtx, initCtxCancel = context.WithTimeout(context.Background(), app.config.InitialisationTimeout)
	} else {
		initCtx, initCtxCancel = context.WithCancel(context.Background())
	}
	defer initCtxCancel()

	go app.init(ctx, signals...)

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
			case <-exitCh:
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

func (app *Application) init(ctx context.Context, signals ...os.Signal) {
	defer Recover()

	for i := range app.constructors {
		var service Service
		var err error
		service, err = app.constructors[i](ctx, app)
		if err != nil {
			app.initCh <- types.ResultError
		}
		app.services = append(app.services, service)
	}

	if len(signals) == 0 {
		signal.Notify(exitCh, defaultTerminateSyscall...)
	} else {
		signal.Notify(exitCh, signals...)
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
		case signal := <-exitCh:
			if signal == types.SIGPANIC {
				err = <-errCh
				if app.config.EnableDebugStack {
					app.log.Println(err, string(debug.Stack()))
				} else {
					app.log.Println(err)
				}
				return ErrRunPanic
			}
			return nil
		case <-ctx.Done():
			return ErrRunContextDeadline
		case err = <-errCh:
			return err
		default:
		}
	}

}

func (app *Application) run() {
	// Start all services with error handling
	for i := range app.services {
		go func() {
			defer Recover()
			if err := app.services[i].Serve(); err != nil {
				errCh <- err
			}
		}()
	}
}

// Shutdown - shutdown the application
func (app *Application) Shutdown() (err error) {
	app.state = StateShutdown

	var (
		shutdownCtx    context.Context
		shutdownCancel context.CancelFunc
	)
	if app.config.InitialisationTimeout != 0 {
		shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), app.config.TerminationTimeout)
	} else {
		shutdownCtx, shutdownCancel = context.WithCancel(context.Background())
	}
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

// Recover - global method for catching application panics
func Recover() {
	if panicMsg := recover(); panicMsg != nil {
		exitCh <- types.SIGPANIC
		errCh <- errors.New(panicMsg.(string))
	}
}
