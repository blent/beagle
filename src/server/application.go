package server

import (
	"golang.org/x/net/context"
)

type Application struct {
	container *Container
}

func NewApplication(settings *Settings) (*Application, error) {
	container, err := NewContainer(settings)

	if err != nil {
		return nil, err
	}

	return &Application{container}, nil
}

func (app *Application) Run() error {
	var err error
	err = app.container.GetInitManager().Run(app.container.GetAllInitializers())

	if err != nil {
		return err
	}

	ctx, stop := context.WithCancel(context.Background())
	stream, err := app.container.GetTracker().Track(ctx)

	if err != nil {
		return err
	}

	// Closes db connection
	defer app.container.GetStorageProvider().Close()

	app.container.GetEventBroker().Use(stream)

	app.container.GetActivityWriter().Use(app.container.GetEventBroker())
	app.container.GetActivityService().Use(app.container.GetEventBroker())

	err = app.container.GetServer().Run(ctx)

	if err != nil {
		stop()

		return err
	}

	return nil
}
