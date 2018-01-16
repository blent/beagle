package notification

type (
	Subscriber struct {
		Id       uint64    `json:"id"`
		Name     string    `json:"name"`
		Event    string    `json:"event"`
		Endpoint *Endpoint `json:"endpoint"`
		Enabled  bool      `json:"enabled"`
	}
)
