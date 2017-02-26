package initialization

import (
	"fmt"
	"github.com/blent/beagle/src/core/logging"
)

type (
	Initializer interface {
		Run() error
	}

	InitManager struct {
		logger *logging.Logger
	}
)

func NewInitManager(logger *logging.Logger) *InitManager {
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
