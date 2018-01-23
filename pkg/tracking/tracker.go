package tracking

import (
	"context"
	"time"

	"github.com/blent/beagle/pkg/discovery"
	"github.com/blent/beagle/pkg/discovery/devices"
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"go.uber.org/zap"
)

const bufferSize = 500

type (
	TrackerError error

	Tracker struct {
		logger    *zap.Logger
		device    devices.Device
		settings  *Settings
		tracks    map[string]*Track
		isRunning bool
	}
)

func NewTracker(logger *zap.Logger, device devices.Device, settings *Settings) *Tracker {
	return &Tracker{
		logger:    logger,
		device:    device,
		settings:  settings,
		tracks:    make(map[string]*Track),
		isRunning: false,
	}
}

func (tracker *Tracker) IsRunning() bool {
	return tracker.isRunning
}

func (tracker *Tracker) Track(ctx context.Context) (*Stream, error) {
	if tracker.isRunning {
		return nil, ErrStart
	}

	if tracker.device.IsScanning() {
		return nil, devices.ErrStartScanning
	}

	inFound := make(chan peripherals.Peripheral, bufferSize)
	inLost := make(chan peripherals.Peripheral, bufferSize)
	inError := make(chan error)

	output, err := tracker.device.Scan(ctx)

	if err != nil {
		return nil, err
	}

	tracker.isRunning = true

	go tracker.start(ctx, output, inFound, inLost, inError)
	go tracker.stopOnDone(ctx, inFound, inLost, inError)

	return NewStream(inFound, inLost, inError), nil
}

func (tracker *Tracker) start(ctx context.Context, stream *discovery.Stream, inFound chan<- peripherals.Peripheral, inLost chan<- peripherals.Peripheral, inError chan<- error) {
	tracker.logger.Info("Started tracking")

	done := false
	ticker := time.NewTicker(tracker.settings.Heartbeat)

	for {
		if done {
			ticker.Stop()
			tracker.logger.Info("Stopped tracking")
			return
		}

		select {
		case <-ctx.Done():
			done = true
		case <-ticker.C:
			tracker.heartbeat(inLost)
		case peripheral, isOpen := <-stream.Data():
			done = !isOpen

			if done == false {
				tracker.push(peripheral, inFound)
			}
		case err, _ := <-stream.Error():
			done = true

			if err != nil {
				tracker.logger.Error(
					"Error occurred in device stream",
					zap.Error(err),
				)

				inError <- err
			}
		}
	}
}

func (tracker *Tracker) stopOnDone(ctx context.Context, inFound chan peripherals.Peripheral, inLost chan peripherals.Peripheral, inError chan error) {
	<-ctx.Done()
	tracker.isRunning = false
	close(inFound)
	close(inLost)
	close(inError)
}

func (tracker *Tracker) heartbeat(inLost chan<- peripherals.Peripheral) {
	if len(tracker.tracks) == 0 {
		return
	}

	active := make(map[string]*Track)

	for key, record := range tracker.tracks {
		if record.IsActive() {
			active[key] = record
		} else {
			inLost <- record.Peripheral()

			tracker.logger.Info(
				"Lost a peripheral",
				zap.String("key", record.Peripheral().UniqueKey()),
			)
		}
	}

	tracker.tracks = active
}

func (tracker *Tracker) push(peripheral peripherals.Peripheral, inFound chan<- peripherals.Peripheral) {
	if peripheral == nil {
		return
	}

	key := peripheral.UniqueKey()

	found, ok := tracker.tracks[key]

	if ok {
		found.Update()
	} else {
		tracker.tracks[key] = NewTrack(peripheral, tracker.settings.Ttl)
		inFound <- peripheral

		tracker.logger.Info(
			"Found a peripheral",
			zap.String("key", key),
		)
	}
}
