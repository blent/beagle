package initializers

import (
	"fmt"
	"github.com/blent/beagle/src/server/storage"
	"go.uber.org/zap"
)

var (
	MSG_ERR_DATABASE             = "failed to initialize database"
	MSG_ERR_DATABASE_TRANSACTION = "failed to initialize transaction for database initialization"
)

type (
	DatabaseInitializer struct {
		logger   *zap.Logger
		provider storage.Provider
	}
)

func NewDatabaseInitializer(logger *zap.Logger, provider storage.Provider) *DatabaseInitializer {
	return &DatabaseInitializer{logger, provider}
}

func (init *DatabaseInitializer) Run() error {
	initializer := init.provider.GetInitializer()

	if initializer == nil {
		return nil
	}

	tx, err := init.provider.GetConnection().Begin()

	if err != nil {
		init.logger.Error(MSG_ERR_DATABASE_TRANSACTION)
		return err
	}

	created, err := initializer(tx)

	if err != nil {
		if rollbackFailure := tx.Rollback(); rollbackFailure != nil {
			err = fmt.Errorf("%s: %s", err.Error(), rollbackFailure.Error())
		}

		init.logger.Error(MSG_ERR_DATABASE)
		init.logger.Error(err.Error())
		return err
	}

	err = tx.Commit()

	if err != nil {
		init.logger.Error(MSG_ERR_DATABASE)
		init.logger.Error(err.Error())
		return err
	}

	if created {
		init.logger.Info("successfully initialized database")
	}

	return nil
}
