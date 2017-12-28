package server

import (
	"github.com/blent/beagle/src/core/discovery/devices"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/core/notification/transport"
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
		zap.String("app", settings.Name),
		zap.String("version", settings.Version),
	))

	if err != nil {
		return nil, err
	}

	// Core
	device, err := devices.NewDevice(logger.Named("device"))

	if err != nil {
		return nil, err
	}

	tracker := tracking.NewTracker(logger.Named("tracker"), device, settings.Tracking)

	// Storage
	storageProvider, err := createStorageProvider(settings.Storage)

	if err != nil {
		return nil, err
	}

	storageManager := storage.NewManager(logger.Named("storage"), storageProvider)

	// Init
	initManager := initialization.NewInitManager(logger.Named("initialization"))
	inits := map[string]initialization.Initializer{
		"storage": initializers.NewDatabaseInitializer(logger.Named("initialization:database"), storageProvider),
	}

	// History
	activityWriter := activity.NewWriter(logger.Named("activity:writer"))

	// Monitoring
	activityService := activityMonitor.NewService(logger.Named("activity:monitor"))

	if err != nil {
		return nil, err
	}

	eventBroker := notification.NewEventBroker(
		logger.Named("broker"),
		notification.NewSender(
			logger.Named("sender"),
			transport.NewHttpTransport(logger.Named("transport")),
		),
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
		webServer = http.NewServer(logger.Named("server"), settings.Http)

		monitoringRoute := routes.NewMonitoringRoute(
			path.Join(settings.Http.Api.Route, "monitoring"),
			logger.Named("route:monitoring"),
			activityService,
			systemMonitor.NewService(logger.Named("service:monitoring:system")),
		)

		peripheralsRoute := routes.NewPeripheralsRoute(
			path.Join(settings.Http.Api.Route, "registry"),
			logger.Named("route:peripherals"),
			storageManager,
		)

		endpointsRoute := routes.NewEndpointsRoute(
			path.Join(settings.Http.Api.Route, "registry"),
			logger.Named("route:endpoints"),
			storageManager,
		)

		inits["routes"] = initializers.NewRoutesInitializer(
			logger.Named("initialization:routes"),
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
