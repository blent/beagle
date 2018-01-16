package system

type (
	Memory struct {
		Total       uint64  `json:"total"`
		Available   uint64  `json:"available"`
		Used        uint64  `json:"used"`
		UsedPercent float64 `json:"usedPercent"`
	}

	Storage struct {
		Total       uint64  `json:"total"`
		Available   uint64  `json:"available"`
		Used        uint64  `json:"used"`
		UsedPercent float64 `json:"usedPercent"`
		Path        string  `json:"path"`
		Fstype      string  `json:"fstype"`
	}

	Stats struct {
		OS       string     `json:"os"`
		Kernel   string     `json:"kernel"`
		Platform string     `json:"platform"`
		Hostname string     `json:"hostname"`
		Arch     string     `json:"arch"`
		Cpu      []float64  `json:"cpu"`
		Memory   *Memory    `json:"memory"`
		Storage  []*Storage `json:"storage"`
	}
)
