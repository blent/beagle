package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/system"
	"github.com/blent/beagle/src/system/http"
	"github.com/blent/beagle/src/system/storage"
	"os"
	"strings"
	"time"
)

var (
	ErrInvalidName              = errors.New("name value must be non-empty string")
	ErrRouteCollision           = errors.New("routes collision detected")
	ErrInvalidTtlDuration       = errors.New("ttl value must be greater than 0")
	ErrInvalidHeartbeatInterval = errors.New("heartbeat value must be greater than 0")
	ErrInvalidStorageConnection = errors.New("storage connection value must be non-empty string")
)

var (
	name              = flag.String("name", "beagle", "application name")
	help              = flag.Bool("help", false, "show this list")
	httpEnable        = flag.Bool("http", true, "enables http server")
	httpPort          = flag.Int("http-port", 8080, "htpp server port number")
	httpApiRoute      = flag.String("http-api-route", "/api", "http server api route")
	httpStaticsDir    = flag.String("http-static-dir", "", "http server static files directory")
	httpStaticsRoute  = flag.String("http-static-route", "/static", "http server static files route")
	trackingTtl       = flag.Int("tracking-ttl", 30, "peripheral ttl duration (seconds)")
	trackingHeartbeat = flag.Int("tracking-heartbeat", 30, "peripheral heartbeat interval (seconds)")
	storageConnection = flag.String("storage-connection", "/tmp/beagle.db", "storage connection string (sqlite)")
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
		settings.Static = &http.StaticSettings{
			Directory: httpStaticsDirVal,
			Route:     strings.TrimSpace(*httpStaticsRoute),
		}

		if settings.Static.Route == settings.Api.Route {
			return ErrRouteCollision
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

func createSettings() (*system.Settings, error) {
	res := system.NewDefaultSettings()

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

	app, err := system.NewApplication(settings)

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
