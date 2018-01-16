package system

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
	"runtime"
)

type Monitoring struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) *Monitoring {
	return &Monitoring{logger}
}

func (s *Monitoring) GetStats() (*Stats, error) {
	stats := &Stats{}

	info, err := host.Info()

	if err != nil {
		return nil, err
	}

	stats.OS = info.OS
	stats.Arch = runtime.GOARCH
	stats.Kernel = info.KernelVersion
	stats.Platform = info.Platform
	stats.Hostname = info.Hostname

	cpuStats, err := s.getCpuStats()

	if err != nil {
		return nil, err
	}

	stats.Cpu = cpuStats

	memoryStats, err := s.getMemoryStats()

	if err != nil {
		return nil, err
	}

	stats.Memory = memoryStats

	storageStats, err := s.getStorageStats()

	if err != nil {
		return nil, err
	}

	stats.Storage = storageStats

	return stats, nil
}

func (s *Monitoring) getCpuStats() ([]float64, error) {
	info, err := cpu.Percent(0, true)

	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *Monitoring) getMemoryStats() (*Memory, error) {
	info, err := mem.VirtualMemory()

	if err != nil {
		return nil, err
	}

	return &Memory{
		Total:       info.Total,
		Available:   info.Available,
		Used:        info.Used,
		UsedPercent: info.UsedPercent,
	}, nil
}

func (s *Monitoring) getStorageStats() ([]*Storage, error) {
	partitions, err := disk.Partitions(false) // only physical

	if err != nil {
		return nil, err
	}

	result := make([]*Storage, 0, len(partitions))

	for _, partition := range partitions {
		info, err := disk.Usage(partition.Mountpoint)

		if err != nil {
			return nil, err
		}

		result = append(result, &Storage{
			Total:       info.Total,
			Available:   info.Free,
			Used:        info.Used,
			UsedPercent: info.UsedPercent,
			Fstype:      info.Fstype,
			Path:        info.Path,
		})
	}

	return result, nil
}
