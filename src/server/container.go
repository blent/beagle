package server

import (
	"github.com/blent/beagle/src/core/discovery/devices"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/core/notification/transports"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/history/activity"
	"github.com/blent/beagle/src/server/http"
	"github.com/blent/beagle/src/server/http/routes"
	"github.com/blent/beagle/src/server/initialization"
	"github.com/blent/beagle/src/server/initialization/initializers"
	activityMonitor "github.com/blent/beagle/src/server/monitoring/activity"
	systemMonitor "github.com/blent/beagle/src/server/monitoring/system"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/providers/sqlite"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"path"
)

type Container struct {
	settings        *Settings
	logger          *zap.Logger
	initManager     *initialization.InitManager
	initializers    map[string]initialization.Initializer
	tracker         *tracking.Tracker
	eventBroker     *notification.EventBroker
	storageProvider storage.Provider
	activityService *activityMonitor.Service
	activityWriter  *activity.Writer
	server          *http.Server
}

func NewContainer(settings *Settings) (*Container, error) {
	var err error

	logger, err := zap.NewProduction(zap.Fields(
		zap.String("name", settings.Name),
		zap.String("version", settings.Version),
	))

	if err != nil {
		return nil, err
	}

	// Core
	device, err := devices.NewDevice(logger)

	if err != nil {
		return nil, err
	}

	tracker := tracking.NewTracker(logger, device, settings.Tracking)

	// Storage
	storageProvider, err := createStorageProvider(settings.Storage)

	if err != nil {
		return nil, err
	}

	storageManager := storage.NewManager(logger, storageProvider)

	// Init
	initManager := initialization.NewInitManager(logger)
	inits := map[string]initialization.Initializer{
		"storage": initializers.NewDatabaseInitializer(logger, storageProvider),
	}

	// History
	activityWriter := activity.NewWriter(logger)

	// Monitoring
	activityService := activityMonitor.NewService(logger)

	if err != nil {
		return nil, err
	}

	eventBroker := notification.NewEventBroker(
		logger,
		notification.NewSender(logger, transports.NewHttpTransport()),
		storageManager.GetPeripheralByKey,
		func(targetId uint64, events ...string) ([]*notification.Subscriber, error) {
			return storageManager.GetPeripheralSubscribersByEvent(
				targetId,
				events,
				storage.PERIPHERAL_STATUS_ENABLED,
			)
		},
	)

	// Http
	var webServer *http.Server

	if settings.Http.Enabled {
		webServer = http.NewServer(logger, settings.Http)

		monitoringRoute := routes.NewMonitoringRoute(
			path.Join(settings.Http.Api.Route, "monitoring"),
			logger,
			activityService,
			systemMonitor.NewService(logger),
		)

		peripheralsRoute := routes.NewPeripheralsRoute(
			path.Join(settings.Http.Api.Route, "registry"),
			logger,
			storageManager,
		)

		endpointsRoute := routes.NewEndpointsRoute(
			path.Join(settings.Http.Api.Route, "registry"),
			logger,
			storageManager,
		)

		inits["routes"] = initializers.NewRoutesInitializer(
			logger,
			webServer,
			[]http.Route{monitoringRoute, peripheralsRoute, endpointsRoute},
		)
	}

	if err != nil {
		return nil, err
	}

	return &Container{
		settings,
		logger,
		initManager,
		inits,
		tracker,
		eventBroker,
		storageProvider,
		activityService,
		activityWriter,
		webServer,
	}, nil
}

func createStorageProvider(settings *storage.Settings) (storage.Provider, error) {
	switch settings.Provider {
	case "sqlite3":
		return sqlite.NewSQLiteProvider(settings.ConnectionString)
	default:
		return nil, errors.New("Not supported storage provider")
	}
}

func (c *Container) GetLogger() *zap.Logger {
	return c.logger
}

func (c *Container) GetInitManager() *initialization.InitManager {
	return c.initManager
}

func (c *Container) GetAllInitializers() map[string]initialization.Initializer {
	return c.initializers
}

func (c *Container) GetEventBroker() *notification.EventBroker {
	return c.eventBroker
}

func (c *Container) GetStorageProvider() storage.Provider {
	return c.storageProvider
}

func (c *Container) GetActivityService() *activityMonitor.Service {
	return c.activityService
}

func (c *Container) GetActivityWriter() *activity.Writer {
	return c.activityWriter
}

func (c *Container) GetTracker() *tracking.Tracker {
	return c.tracker
}

func (c *Container) GetServer() *http.Server {
	return c.server
}
