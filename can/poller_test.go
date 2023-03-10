package can_test

import (
	"echoctl/can"
	"echoctl/conf"
	"echoctl/schedule"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"syscall"
	"testing"
	"time"
)

func TestBlockingBehavior(t *testing.T) {
	t.Run("exits on kill", func(t *testing.T) {
		t.Parallel()

		poller, _, _, _ := NewPoller()
		runAndKillPoller(t, poller, func() {})
	})
}

func TestSending(t *testing.T) {
	t.Run("send command when triggered", func(t *testing.T) {
		t.Parallel()

		poller, socket, _, nextTrigger := NewPoller()

		runAndKillPoller(t, poller, func() {
			nextTrigger <- newTrigger(123, time.Second)
			cmd := readWithTimeout(t, socket.Outbound())
			assert.Equal(t, uint32(123), cmd.ID, "Looks like we received the wrong command")
		})
	})

	t.Run("reschedule successful sends with configured delay", func(t *testing.T) {
		t.Parallel()
		poller, _, scheduleRequests, nextTrigger := NewPoller()

		runAndKillPoller(t, poller, func() {
			nextTrigger <- newTrigger(123, 3*time.Second)
			scheduleRequest := readWithTimeout(t, scheduleRequests)
			assert.Equal(t, 3*time.Second, scheduleRequest.TriggerIn, "the schedule request should have TriggerIn=3s")
		})
	})

	t.Run("reschedule failed sends with short delay", func(t *testing.T) {
		t.Parallel()
		poller, socket, scheduleRequests, nextTrigger := NewPoller()

		runAndKillPoller(t, poller, func() {
			socket.NextSendError(syscall.ENOBUFS)
			nextTrigger <- newTrigger(123, 3*time.Second)
			scheduleRequest := readWithTimeout(t, scheduleRequests)
			assert.Equal(t, can.RetryDelay, scheduleRequest.TriggerIn, "the schedule request should have a short TriggerIn")
		})
	})
}

func newTrigger(canId conf.CanId, delay time.Duration) schedule.Trigger[can.Subscription] {
	return schedule.Trigger[can.Subscription]{
		Data: &can.Subscription{
			Command: NewCommand(canId),
			Delay:   delay,
		},
		TriggeredAt: time.Now(),
	}
}

func readWithTimeout[T any](t *testing.T, ch <-chan T) *T {
	select {
	case value := <-ch:
		return &value
	case <-time.After(time.Second):
		assert.Fail(t, "Poller failed to send command in 1s")
		return nil
	}
}

func NewPoller() (can.Poller, SocketMock, chan schedule.Request[can.Subscription], chan schedule.Trigger[can.Subscription]) {
	socket := NewSocketMock()
	inbound := make(chan conf.Command)
	scheduleRequests := make(chan schedule.Request[can.Subscription], 20)
	nextTrigger := make(chan schedule.Trigger[can.Subscription])
	scheduler := schedule.NewImmediatelyScheduler(scheduleRequests, nextTrigger)
	poller := can.NewPoller(socket, []can.Subscription{}, inbound, scheduler, zap.NewNop())
	return poller, socket, scheduleRequests, nextTrigger
}

func runAndKillPoller(t *testing.T, poller can.Poller, f func()) {
	tmb := poller.Poll()
	f()
	tmb.Kill(nil)
	select {
	case <-tmb.Dead():
		// success
	case <-time.After(time.Second):
		assert.Fail(t, "Poller failed to exit in 1s")
	}
}

func NewCommand(canId conf.CanId) conf.Command {
	return conf.Command{
		Id: "001",
		Request: conf.RequestCommand{
			CanId:        canId,
			CommandBytes: []byte{3, 7, 5},
		},
	}
}
