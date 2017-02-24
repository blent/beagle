package system

import (
	"github.com/blent/beagle/src/core/discovery/devices"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/core/notification/delivery"
	"github.com/blent/beagle/src/core/notification/delivery/transports"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/system/history/activity"
	"github.com/blent/beagle/src/system/http"
	"github.com/blent/beagle/src/system/http/routes"
	"github.com/blent/beagle/src/system/initialization"
	"github.com/blent/beagle/src/system/initialization/initializers"
	"github.com/blent/beagle/src/system/storage"
	"github.com/blent/beagle/src/system/storage/sqlite"
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

	// Http
	server := http.NewServer(logging.NewLogger("server", log), settings.Http)

	var activityRoute *routes.ActivityRoute

	if settings.Http.Enabled {
		activityRoute = routes.NewActivityRoutes(
			settings.Http.Api.Route,
			logging.NewLogger("route:activity", log),
			activityWriter,
		)
	}

	// Init
	initManager := initialization.NewInitManager(logging.NewLogger("init", log))

	db, err := storage.NewDatabase(settings.Storage)

	if err != nil {
		return nil, err
	}

	inits := map[string]initialization.Initializer{
		"database": initializers.NewDatabaseInitializer(
			logging.NewLogger("init:database", log),
			db,
		),
	}

	if settings.Http.Enabled {
		inits["routes"] = initializers.NewRoutesInitializer(
			logging.NewLogger("init:routes", log),
			server,
			[]http.Route{activityRoute},
		)
	}

	targetRepo := sqlite.NewSQLiteTargetRepository(db)

	eventBroker := notification.NewEventBroker(
		logging.NewLogger("broker", log),
		sender,
		func(key string) (*tracking.Target, error) {
			return targetRepo.FindByKey(key)
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
