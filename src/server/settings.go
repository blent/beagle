package server

import (
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/http"
	"github.com/blent/beagle/src/server/storage"
	"time"
)

type Settings struct {
	Version  string
	Name     string
	Http     *http.Settings
	Storage  *storage.Settings
	Tracking *tracking.Settings
}

func NewDefaultSettings() *Settings {
	return &Settings{
		Name: "beagle",
		Http: &http.Settings{
			Port:     8080,
			Enabled:  true,
			Headless: true,
			Api: &http.ApiSettings{
				Route: "/api",
			},
			Static: &http.StaticSettings{
				Route:     "/public",
				Directory: "",
			},
		},
		Storage: &storage.Settings{
			ConnectionString: "/var/lib/beagle/database.db",
			Provider:         "sqlite3",
		},
		Tracking: &tracking.Settings{
			Heartbeat: time.Second * 5,
			Ttl:       time.Second * 5,
		},
	}
}
