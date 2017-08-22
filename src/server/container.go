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
	activity2 "github.com/blent/beagle/src/server/monitoring/activity"
	system2 "github.com/blent/beagle/src/server/monitoring/system"
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
	activityService *activity2.Service
	activityWriter  *activity.Writer
	server          *http.Server
}

func NewContainer(settings *Settings) (*Container, error) {
	var err error

	// Core
	device, err := createDevice()

	if err != nil {
		return nil, err
	}

	tracker, err := createTracker(device, settings.Tracking)

	if err != nil {
		return nil, err
	}

	sender, err := createSender()

	if err != nil {
		return nil, err
	}

	// Storage
	storageProvider, err := createStorageProvider(settings.Storage)

	if err != nil {
		return nil, err
	}

	storageManager, err := createStorageManager(storageProvider)

	if err != nil {
		return nil, err
	}

	// Init
	initManager, err := createInitManager()

	if err != nil {
		return nil, err
	}

	inits := make(map[string]initialization.Initializer)

	databaseInitializer, err := createDatabaseInitializer(storageProvider)

	if err != nil {
		return nil, err
	}

	inits["storage"] = databaseInitializer

	// History
	activityWriter, err := createActivityWriter()

	if err != nil {
		return nil, err
	}

	// Monitoring
	activityService, err := createActivityMonitoring()

	if err != nil {
		return nil, err
	}

	eventBroker, err := createEventBroker(sender, storageManager)

	if err != nil {
		return nil, err
	}

	// Http
	var webServer *http.Server

	if settings.Http.Enabled {
		webServer, err = createWebServer(settings.Http)

		if err != nil {
			return nil, err
		}

		systemService, err := createSystemMonitoring()

		if err != nil {
			return nil, err
		}

		monitoringRoute, err := createMonitoringRoute(settings.Http.Api, activityService, systemService)

		if err != nil {
			return nil, err
		}

		peripheralsRoute, err := createPeripheralsRoute(settings.Http.Api, storageManager)

		if err != nil {
			return nil, err
		}

		endpointsRoute, err := createEndpointsRoute(settings.Http.Api, storageManager)

		if err != nil {
			return nil, err
		}

		routesInitializer, err := createRoutesInitializer(
			webServer,
			[]http.Route{monitoringRoute, peripheralsRoute, endpointsRoute},
		)

		if err != nil {
			return nil, err
		}

		inits["routes"] = routesInitializer
	}

	appLogger, err := createLogger("application")

	if err != nil {
		return nil, err
	}

	return &Container{
		settings,
		appLogger,
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

func createLogger(name string) (*zap.Logger, error) {
	logger, err := zap.NewProduction()

	if err != nil {
		return nil, err
	}

	return logger.Named(name), nil
}

func createDevice() (devices.Device, error) {
	logger, err := createLogger("device")

	if err != nil {
		return nil, err
	}

	return devices.NewDevice(logger)
}

func createTracker(device devices.Device, settings *tracking.Settings) (*tracking.Tracker, error) {
	logger, err := createLogger("tracker")

	if err != nil {
		return nil, err
	}

	return tracking.NewTracker(logger, device, settings), nil
}

func createEventBroker(sender *notification.Sender, storageManager *storage.Manager) (*notification.EventBroker, error) {
	logger, err := createLogger("broker")

	if err != nil {
		return nil, err
	}

	return notification.NewEventBroker(
		logger,
		sender,
		storageManager.GetPeripheralByKey,
		func(targetId uint64, events ...string) ([]*notification.Subscriber, error) {
			return storageManager.GetPeripheralSubscribersByEvent(
				targetId,
				events,
				storage.PERIPHERAL_STATUS_ENABLED,
			)
		},
	), nil
}

func createSender() (*notification.Sender, error) {
	logger, err := createLogger("sender")

	if err != nil {
		return nil, err
	}

	return notification.NewSender(logger, transports.NewHttpTransport()), nil
}

func createStorageProvider(settings *storage.Settings) (storage.Provider, error) {
	switch settings.Provider {
	case "sqlite3":
		return sqlite.NewSQLiteProvider(settings.ConnectionString)
	default:
		return nil, errors.New("Not supported storage provider")
	}
}

func createStorageManager(provider storage.Provider) (*storage.Manager, error) {
	logger, err := createLogger("storage")

	if err != nil {
		return nil, err
	}

	return storage.NewManager(logger, provider), nil
}

func createActivityWriter() (*activity.Writer, error) {
	logger, err := createLogger("history")

	if err != nil {
		return nil, err
	}

	return activity.NewWriter(logger), nil
}

func createActivityMonitoring() (*activity2.Service, error) {
	logger, err := createLogger("monitoring.activity")

	if err != nil {
		return nil, err
	}

	return activity2.NewService(logger), nil
}

func createSystemMonitoring() (*system2.Service, error) {
	logger, err := createLogger("monitoring.system")

	if err != nil {
		return nil, err
	}

	return system2.NewService(logger), nil
}

func createWebServer(settings *http.Settings) (*http.Server, error) {
	logger, err := createLogger("server")

	if err != nil {
		return nil, err
	}

	return http.NewServer(logger, settings), nil
}

func createMonitoringRoute(settings *http.ApiSettings, activity *activity2.Service, system *system2.Service) (*routes.MonitoringRoute, error) {
	logger, err := createLogger("route.monitoring")

	if err != nil {
		return nil, err
	}

	return routes.NewMonitoringRoute(
		path.Join(settings.Route, "monitoring"),
		logger,
		activity,
		system,
	), nil
}

func createPeripheralsRoute(settings *http.ApiSettings, storage *storage.Manager) (*routes.PeripheralsRoute, error) {
	logger, err := createLogger("route.registry.peripherals")

	if err != nil {
		return nil, err
	}

	return routes.NewPeripheralsRoute(
		path.Join(settings.Route, "registry"),
		logger,
		storage,
	), nil
}

func createEndpointsRoute(settings *http.ApiSettings, storage *storage.Manager) (*routes.EndpointsRoute, error) {
	logger, err := createLogger("route.registry.endpoints")

	if err != nil {
		return nil, err
	}

	return routes.NewEndpointsRoute(
		path.Join(settings.Route, "registry"),
		logger,
		storage,
	), nil
}

func createInitManager() (*initialization.InitManager, error) {
	logger, err := createLogger("initialization")

	if err != nil {
		return nil, err
	}

	return initialization.NewInitManager(logger), nil
}

func createDatabaseInitializer(storageProvider storage.Provider) (*initializers.DatabaseInitializer, error) {
	logger, err := createLogger("initialization.database")

	if err != nil {
		return nil, err
	}

	return initializers.NewDatabaseInitializer(logger, storageProvider), nil
}

func createRoutesInitializer(webServer *http.Server, routes []http.Route) (*initializers.RoutesInitializer, error) {
	logger, err := createLogger("initialization.routes")

	if err != nil {
		return nil, err
	}

	return initializers.NewRoutesInitializer(
		logger,
		webServer,
		routes,
	), nil
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

func (c *Container) GetActivityService() *activity2.Service {
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
