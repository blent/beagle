package server

import (
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Application struct {
	container *Container
}

func New(settings *Settings) (*Application, error) {
	container, err := NewContainer(settings)

	if err != nil {
		return nil, err
	}

	return &Application{container}, nil
}

func (app *Application) Run() error {
	var err error

	logger := app.container.GetLogger()

	logger.Info("Starting the application")

	err = app.container.GetInitManager().Run(app.container.GetAllInitializers())

	if err != nil {
		logger.Error(
			"Failed to initialize the system",
			zap.Error(err),
		)

		return err
	}

	ctx, stop := context.WithCancel(context.Background())
	stream, err := app.container.GetTracker().Track(ctx)

	if err != nil {
		logger.Error(
			"Failed to start the tracker",
			zap.Error(err),
		)

		return err
	}

	// Closes db connection
	defer app.container.GetStorageProvider().Close()

	app.container.GetEventBroker().Use(stream)

	app.container.GetActivityWriter().Use(app.container.GetEventBroker())
	app.container.GetActivityService().Use(app.container.GetEventBroker())

	err = app.container.GetServer().Run(ctx)

	if err != nil {
		logger.Error(
			"Failed to start the server",
			zap.Error(err),
		)

		stop()

		return err
	}

	return nil
}
