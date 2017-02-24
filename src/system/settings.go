package system

import (
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/system/http"
	"github.com/blent/beagle/src/system/storage"
	"time"
)

type Settings struct {
	Name     string
	Http     *http.Settings
	Storage  *storage.Settings
	Tracking *tracking.Settings
}

func NewDefaultSettings() *Settings {
	return &Settings{
		Name: "beacon-tracker",
		Http: &http.Settings{
			Port:     8080,
			Enabled:  true,
			Headless: true,
			Api: &http.ApiSettings{
				Route: "/api",
			},
		},
		Storage: &storage.Settings{
			ConnectionString: "/tmp/beagle.db",
			Dialect:          "sqlite3",
		},
		Tracking: &tracking.Settings{
			Heartbeat: time.Second * 30,
			Ttl:       time.Second * 30,
		},
	}
}
