package server

import (
	"github.com/blent/beagle/src/core/discovery/devices"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/core/notification/delivery"
	"github.com/blent/beagle/src/core/notification/delivery/transports"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/history/activity"
	"github.com/blent/beagle/src/server/http"
	"github.com/blent/beagle/src/server/http/routes"
	"github.com/blent/beagle/src/server/initialization"
	"github.com/blent/beagle/src/server/initialization/initializers"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/sqlite"
	"github.com/pkg/errors"
	"path"
)

type Container struct {
	settings       *Settings
	initManager    *initialization.InitManager
	initializers   map[string]initialization.Initializer
	tracker        *tracking.Tracker
	eventBroker    *notification.EventBroker
	activityWriter *activity.Writer
	server         *http.Server
}

func NewContainer(settings *Settings) (*Container, error) {
	log := logging.DefaultOutput

	var err error

	// Core
	device, err := devices.NewBleDevice(logging.NewLogger("device", log))

	if err != nil {
		return nil, err
	}

	tracker := tracking.NewTracker(logging.NewLogger("tracker", log), device, settings.Tracking)
	sender := delivery.NewSender(logging.NewLogger("sender", log), transports.NewHttpTransport())

	// History
	activityWriter := activity.NewWriter(logging.NewLogger("history", log))

	// Storage
	storageProvider, err := getStorageProvider(settings.Storage)

	if err != nil {
		return nil, err
	}

	targetRepository := storageProvider.GetTargetRepository()

	// Init
	initManager := initialization.NewInitManager(logging.NewLogger("initialization", log))

	inits := map[string]initialization.Initializer{
		"database": initializers.NewDatabaseInitializer(
			logging.NewLogger("initialization:database", log),
			storageProvider,
		),
	}

	// Http
	var server *http.Server

	if settings.Http.Enabled {
		server = http.NewServer(logging.NewLogger("server", log), settings.Http)

		inits["routes"] = initializers.NewRoutesInitializer(
			logging.NewLogger("initialization:routes", log),
			server,
			[]http.Route{
				routes.NewActivityRoute(
					settings.Http.Api.Route,
					logging.NewLogger("route:activity", log),
					activityWriter,
				),
				routes.NewTargetRoute(
					path.Join(settings.Http.Api.Route, "registry"),
					logging.NewLogger("route:registry:target", log),
					targetRepository,
				),
			},
		)
	}

	eventBroker := notification.NewEventBroker(
		logging.NewLogger("broker", log),
		sender,
		func(key string) (*tracking.Target, error) {
			return targetRepository.GetByKey(key)
		},
	)

	return &Container{
		settings,
		initManager,
		inits,
		tracker,
		eventBroker,
		activityWriter,
		server,
	}, nil
}

func getStorageProvider(settings *storage.Settings) (storage.Provider, error) {
	switch settings.Provider {
	case "sqlite3":
		return sqlite.NewSQLiteProvider(settings.ConnectionString)
	default:
		return nil, errors.New("Not supported storage provider")
	}
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

func (c *Container) GetActivityWriter() *activity.Writer {
	return c.activityWriter
}

func (c *Container) GetTracker() *tracking.Tracker {
	return c.tracker
}

func (c *Container) GetServer() *http.Server {
	return c.server
}
