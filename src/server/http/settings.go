package http

type (
	Settings struct {
		Port     int
		Headless bool
		Enabled  bool
		Api      *ApiSettings
		Static   *StaticSettings
	}

	ApiSettings struct {
		Route string
	}

	StaticSettings struct {
		Route     string
		Directory string
	}
)
