package main

import (
	"flag"
	"fmt"
	"github.com/blent/beagle/src/core"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server"
	"github.com/blent/beagle/src/server/http"
	"github.com/blent/beagle/src/server/storage"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"
)

var DefaultSettings = server.NewDefaultSettings()

var (
	ErrInvalidName              = errors.New("name value must be non-empty string")
	ErrRouteCollision           = errors.New("routes collision detected")
	ErrStaticRoute              = errors.New("static route must be non-empty string")
	ErrInvalidTtlDuration       = errors.New("ttl value must be greater than 0")
	ErrInvalidHeartbeatInterval = errors.New("heartbeat value must be greater than 0")
	ErrInvalidStorageConnection = errors.New("storage connection value must be non-empty string")
)

var (
	help = flag.Bool(
		"help",
		false,
		"show this list",
	)
	version = flag.Bool(
		"version",
		false,
		"show version",
	)
	name = flag.String(
		"name",
		DefaultSettings.Name,
		"application name",
	)
	httpEnable = flag.Bool(
		"http",
		DefaultSettings.Http.Enabled,
		"enables http server",
	)
	httpPort = flag.Int(
		"http-port",
		DefaultSettings.Http.Port,
		"http server port number",
	)
	httpApiRoute = flag.String(
		"http-api-route",
		DefaultSettings.Http.Api.Route,
		"http server api route",
	)
	httpStaticsDir = flag.String("http-static-dir",
		DefaultSettings.Http.Static.Directory,
		"http server static files directory",
	)
	httpStaticsRoute = flag.String(
		"http-static-route",
		DefaultSettings.Http.Static.Route,
		"http server static files route",
	)
	trackingTtl = flag.Int(
		"tracking-ttl",
		int(DefaultSettings.Tracking.Ttl/time.Second),
		"peripheral ttl duration in seconds",
	)
	trackingHeartbeat = flag.Int(
		"tracking-heartbeat",
		int(DefaultSettings.Tracking.Heartbeat/time.Second),
		"peripheral heartbeat interval in seconds",
	)
	storageConnection = flag.String(
		"storage-connection",
		DefaultSettings.Storage.ConnectionString,
		"storage connection string",
	)
)

func setHttpSettings(settings *http.Settings) error {
	settings.Enabled = *httpEnable

	if !settings.Enabled {
		return nil
	}

	settings.Api.Route = strings.TrimSpace(*httpApiRoute)
	settings.Port = *httpPort
	settings.Headless = true

	httpStaticsDirVal := strings.TrimSpace(*httpStaticsDir)

	if httpStaticsDirVal != "" {
		settings.Static.Directory = httpStaticsDirVal
		settings.Static.Route = strings.TrimSpace(*httpStaticsRoute)

		if settings.Static.Route == settings.Api.Route {
			return ErrRouteCollision
		}

		if settings.Static.Route == "" {
			return ErrStaticRoute
		}
	}

	return nil
}

func setTrackingSettings(settings *tracking.Settings) error {
	trackingTtlVal := *trackingTtl

	if trackingTtlVal < 0 {
		return ErrInvalidTtlDuration
	}

	trackingHeartbeat := *trackingHeartbeat

	if trackingHeartbeat < 0 {
		return ErrInvalidHeartbeatInterval
	}

	settings.Ttl = time.Second * time.Duration(trackingTtlVal)
	settings.Heartbeat = time.Second * time.Duration(trackingHeartbeat)

	return nil
}

func setStorageSettings(settings *storage.Settings) error {
	settings.ConnectionString = strings.TrimSpace(*storageConnection)

	if settings.ConnectionString == "" {
		return ErrInvalidStorageConnection
	}

	return nil
}

func createSettings() (*server.Settings, error) {
	res := server.NewDefaultSettings()

	res.Name = strings.TrimSpace(*name)

	if res.Name == "" {
		return nil, ErrInvalidName
	}

	if err := setHttpSettings(res.Http); err != nil {
		return nil, err
	}

	if err := setTrackingSettings(res.Tracking); err != nil {
		return nil, err
	}

	if err := setStorageSettings(res.Storage); err != nil {
		return nil, err
	}

	return res, nil
}

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
		return
	}

	if *version {
		fmt.Println(core.Version)
		os.Exit(0)
		return
	}

	if os.Geteuid() != 0 {
		fmt.Println(os.ErrPermission.Error())
		os.Exit(1)
		return
	}

	settings, err := createSettings()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}

	app, err := server.NewApplication(settings)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}

	if err := app.Run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}
}
