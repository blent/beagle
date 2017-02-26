package tracking

import (
	"github.com/blent/beagle/src/core/discovery"
	"github.com/blent/beagle/src/core/discovery/devices"
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"golang.org/x/net/context"
	"time"
)

const bufferSize = 500

type (
	TrackerError error

	Tracker struct {
		logger    *logging.Logger
		device    devices.Device
		settings  *Settings
		tracks    map[string]*Track
		isRunning bool
	}
)

func NewTracker(logger *logging.Logger, device devices.Device, settings *Settings) *Tracker {
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
	streamIsClosed := false

	ticker := time.NewTicker(tracker.settings.Heartbeat)

	for {
		if streamIsClosed {
			return
		}

		select {
		case <-ticker.C:
			tracker.heartbeat(inLost)
		case peripheral, isOpen := <-stream.Data():
			streamIsClosed = !isOpen

			if peripheral != nil {
				tracker.push(peripheral, inFound)
			}
		case err, _ := <-stream.Error():
			streamIsClosed = true

			ticker.Stop()

			if err != nil {
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

	alive := make(map[string]*Track)

	for key, record := range tracker.tracks {
		if !record.IsLost() {
			alive[key] = record
		} else {
			inLost <- record.Peripheral()
			tracker.logger.Infof("Lost a peripheral with key %s", record.Peripheral().UniqueKey())
		}
	}

	tracker.tracks = alive
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
		tracker.logger.Infof("Found a new peripheral with id %s", key)
	}
}
