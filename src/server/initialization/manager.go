package initialization

import (
	"fmt"
	"go.uber.org/zap"
)

type (
	Initializer interface {
		Run() error
	}

	InitManager struct {
		logger *zap.Logger
	}
)

func NewInitManager(logger *zap.Logger) *InitManager {
	return &InitManager{
		logger: logger,
	}
}

func (manager *InitManager) Run(initializers map[string]Initializer) error {
	var err error

	for name, init := range initializers {
		initError := init.Run()

		if initError != nil {
			err = fmt.Errorf("Error occured during %s initializer: %s", name, initError.Error())
			manager.logger.Error(err.Error())
			break
		}
	}

	return err
}
