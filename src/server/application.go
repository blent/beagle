package server

import (
	"golang.org/x/net/context"
)

type Application struct {
	container *Container
}

func NewDefaultApplication() (*Application, error) {
	return NewApplication(NewDefaultSettings())
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

	ctx := context.Background()
	_, err = app.container.GetTracker().Track(ctx)

	if err != nil {
		return err
	}

	// engine.getEventBroker().Use(stream)

	app.container.GetActivityWriter().Use(app.container.GetEventBroker())

	return app.container.GetServer().Run(ctx)
}
