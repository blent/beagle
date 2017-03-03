package tracking

type Target struct {
	Id          uint64        `json:"id"`
	Key         string        `json:"key"`
	Name        string        `json:"name"`
	Kind        string        `json:"kind"`
	Enabled     bool          `json:"enabled"`
	Subscribers []*Subscriber `json:"subscribers"`
}
