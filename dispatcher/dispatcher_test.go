package dispatcher_test

import (
	"echoctl/conf"
	"echoctl/dispatcher"
	"github.com/go-daq/canbus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestBlockingFlow(t *testing.T) {
	t.Run("Exits on blocking inbound", func(t *testing.T) {
		t.Parallel()
		inbound := make(chan canbus.Frame, 1)
		toRequestor := make(chan conf.Command, 1)
		toMqttPublisher := make(chan dispatcher.CommandValue, 1)
		d := NewDispatcherWithChannels(inbound, toRequestor, toMqttPublisher, []conf.Command{})

		tmb := d.Dispatch()
		tmb.Kill(nil)
		select {
		case <-tmb.Dead():
		case <-time.After(time.Second):
			t.Log("Dispatcher failed to shut down in 1s")
		}
	})

	t.Run("Exits on blocking toRequestor", func(t *testing.T) {
		t.Parallel()
		inbound := make(chan canbus.Frame, 1)
		toRequestor := make(chan conf.Command)
		toMqttPublisher := make(chan dispatcher.CommandValue, 1)
		sendToInboundAndKill(t, inbound, toRequestor, toMqttPublisher)
	})

	t.Run("Exits on blocking toMqttPublisher", func(t *testing.T) {
		t.Parallel()
		inbound := make(chan canbus.Frame, 1)
		toRequestor := make(chan conf.Command, 1)
		toMqttPublisher := make(chan dispatcher.CommandValue)
		sendToInboundAndKill(t, inbound, toRequestor, toMqttPublisher)
	})
}

func TestFunction(t *testing.T) {
	t.Run("Passes matched command to requestor", func(t *testing.T) {
		t.Parallel()
		d, inbound, toRequestor, _ := NewDispatcher([]conf.Command{
			{
				Id: "001",
				Response: conf.RequestCommand{
					CanId:        123,
					CommandBytes: []byte{3, 7, 5},
				},
			},
		})

		startAndRun(t, d, func() {
			inbound <- canbus.Frame{ID: 123, Data: []byte{3, 7, 5, 0, 0}}
			select {
			case cmd := <-toRequestor:
				assert.Equal(t, "001", cmd.Id, "ID is different. Wrong match?")
			case <-time.After(time.Second):
				t.Log("Timeout waiting for data from toRequestor.")
			}
		})
	})

	t.Run("Matches command by id and CommandBytes", func(t *testing.T) {
		t.Parallel()
		d, inbound, toRequestor, _ := NewDispatcher([]conf.Command{
			{
				Id: "001",
				Response: conf.RequestCommand{
					CanId:        456,
					CommandBytes: []byte{3, 7, 5},
				},
			},
			{
				Id: "002",
				Response: conf.RequestCommand{
					CanId:        123,
					CommandBytes: []byte{4, 8, 5},
				},
			},
			{
				Id: "003",
				Response: conf.RequestCommand{
					CanId:        123,
					CommandBytes: []byte{3, 7, 5},
				},
			},
		})

		startAndRun(t, d, func() {
			inbound <- canbus.Frame{ID: 123, Data: []byte{3, 7, 5, 0, 0}}
			select {
			case cmd := <-toRequestor:
				assert.Equal(t, "003", cmd.Id, "ID is different. Wrong match?")
			case <-time.After(time.Second):
				t.Log("Timeout waiting for data from toRequestor.")
			}
		})
	})

	t.Run("Passes inbound value to mqttPublisher", func(t *testing.T) {
		t.Parallel()
		d, inbound, _, toMqttPublisher := NewDispatcher([]conf.Command{
			{
				Id: "001",
				Response: conf.RequestCommand{
					CanId:        123,
					CommandBytes: []byte{3, 7, 5},
				},
			},
		})

		startAndRun(t, d, func() {
			inbound <- canbus.Frame{ID: 123, Data: []byte{3, 7, 5, 4, 3}}
			select {
			case commValue := <-toMqttPublisher:
				assert.Equal(t, int16(4*256+3), commValue.Value, "ID is different. Wrong match?")
			case <-time.After(time.Second):
				t.Log("Timeout waiting for data from toRequestor.")
			}
		})
	})

	t.Run("Interpets inbound value as int16", func(t *testing.T) {
		t.Parallel()
		d, inbound, _, toMqttPublisher := NewDispatcher([]conf.Command{
			{
				Id: "001",
				Response: conf.RequestCommand{
					CanId:        123,
					CommandBytes: []byte{1, 1, 1},
				},
			},
		})

		startAndRun(t, d, func() {
			inbound <- canbus.Frame{ID: 123, Data: []byte{1, 1, 1, 0xff, 0xff}}
			select {
			case commValue := <-toMqttPublisher:
				assert.Equal(t, int16(-1), commValue.Value, "ID is different. Wrong match?")
			case <-time.After(time.Second):
				t.Log("Timeout waiting for data from toRequestor.")
			}
		})
	})

	t.Run("Continues on unknown command", func(t *testing.T) {
		t.Parallel()
		d, inbound, toRequestor, _ := NewDispatcher([]conf.Command{
			{
				Id: "001",
				Response: conf.RequestCommand{
					CanId:        123,
					CommandBytes: []byte{3, 7, 5},
				},
			},
		})

		startAndRun(t, d, func() {
			inbound <- canbus.Frame{ID: 123, Data: []byte{1, 2, 3, 0, 0}}
			inbound <- canbus.Frame{ID: 123, Data: []byte{3, 7, 5, 4, 3}}
			select {
			case cmd := <-toRequestor:
				assert.Equal(t, "001", cmd.Id, "ID is different. Wrong match?")
			case <-time.After(time.Second):
				t.Log("Timeout waiting for data from toRequestor.")
			}
		})
	})

}

func sendToInboundAndKill(t *testing.T, inbound chan canbus.Frame, toRequestor chan conf.Command, toMqttPublisher chan dispatcher.CommandValue) {
	d := NewDispatcherWithChannels(inbound, toRequestor, toMqttPublisher, []conf.Command{
		{
			Response: conf.RequestCommand{
				CanId:        123,
				CommandBytes: []byte{3, 7, 5},
			},
		},
	})
	tmb := d.Dispatch()
	inbound <- canbus.Frame{ID: 123, Data: []byte{3, 7, 5, 0, 0}}

	tmb.Kill(nil)
	select {
	case <-tmb.Dead():
	case <-time.After(time.Second):
		t.Log("Dispatcher failed to shut down in 1s")
	}
}

func NewDispatcherWithChannels(inbound <-chan canbus.Frame, toRequestor chan<- conf.Command, toMqttPublisher chan<- dispatcher.CommandValue, commands []conf.Command) (d dispatcher.Dispatcher) {
	d = dispatcher.NewDispatcher(inbound, commands, toRequestor, toMqttPublisher, zap.NewNop())
	return
}

func NewDispatcher(commands []conf.Command) (d dispatcher.Dispatcher, inbound chan canbus.Frame, toRequestor chan conf.Command, toMqttPublisher chan dispatcher.CommandValue) {
	inbound = make(chan canbus.Frame, 1)
	toRequestor = make(chan conf.Command, 1)
	toMqttPublisher = make(chan dispatcher.CommandValue, 1)
	d = dispatcher.NewDispatcher(inbound, commands, toRequestor, toMqttPublisher, zap.NewNop())
	return
}

func startAndRun(t *testing.T, d dispatcher.Dispatcher, f func()) {
	tmb := d.Dispatch()

	f()

	tmb.Kill(nil)
	select {
	case <-tmb.Dead():
	case <-time.After(time.Second):
		t.Log("Dispatcher failed to shut down in 1s")
	}
}
