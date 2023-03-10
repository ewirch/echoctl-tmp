package can

import (
	"echoctl/conf"
	"echoctl/flowcontrol"
	"echoctl/schedule"
	"errors"
	"github.com/go-daq/canbus"
	"go.uber.org/zap"
	"gopkg.in/tomb.v2"
	"syscall"
	"time"
)

const RetryDelay = 100 * time.Millisecond

type Subscription struct {
	Command conf.Command
	Delay   time.Duration
}

type poller struct {
	socket        Socket
	subscriptions []Subscription
	inbound       <-chan conf.Command
	tomb          *tomb.Tomb
	log           *zap.Logger
	scheduler     schedule.Scheduler[Subscription]
}

// Poller sends periodic commands to a can-bus socket, following the specified schedule. Poller does not wait for a reply. It relies on Reader to read the reply from can-bus. The Reader passes the received frame to the Dispatcher, and the Dispatcher passes it on to Poller.
type Poller interface {
	Poll() *tomb.Tomb
}

var _ Poller = (*poller)(nil)

func NewPoller(socket Socket, subscriptions []Subscription, inbound <-chan conf.Command, scheduler schedule.Scheduler[Subscription], log *zap.Logger) Poller {
	return &poller{
		socket:        socket,
		subscriptions: subscriptions,
		inbound:       inbound,
		tomb:          new(tomb.Tomb),
		log:           log,
		scheduler:     scheduler,
	}
}

func (poller *poller) Poll() *tomb.Tomb {
	poller.tomb.Go(poller.poll)
	poller.tomb.Go(func() error {
		poller.scheduler.Run(poller.tomb.Dying())
		return nil
	})
	return poller.tomb
}

func (poller *poller) poll() error {
	poller.createSchedule(poller.subscriptions)
	for {
		select {
		case trigger := <-poller.scheduler.Next():
			if err := poller.processTrigger(trigger); err != nil {
				return err
			}

		case <-poller.inbound:
			// Ignore inbound commands for now.

		case <-poller.tomb.Dying():
			return tomb.ErrDying
		}
	}
}

func (poller *poller) createSchedule(subscriptions []Subscription) {
	for _, subscription := range subscriptions {
		poller.scheduler.Schedule() <- schedule.Request[Subscription]{Data: &subscription, TriggerIn: subscription.Delay}
	}
}

func (poller *poller) processTrigger(trigger schedule.Trigger[Subscription]) error {
	err := poller.sendCommand(trigger.Data.Command)

	if flowcontrol.IsShouldRetry(err) {
		// Retry sending, but delay a bit, to not directly fail again on retry.
		poller.scheduler.Schedule() <- schedule.Request[Subscription]{Data: trigger.Data, TriggerIn: RetryDelay}
		return nil
	}
	if err != nil {
		return err
	}

	// Command sent successfully, reschedule the next sending.
	poller.scheduler.Schedule() <- schedule.Request[Subscription]{Data: trigger.Data, TriggerIn: trigger.Data.Delay}
	return nil
}

func (poller *poller) sendCommand(command conf.Command) error {
	poller.log.Debug("sending", zap.String("command", command.Id))
	_, err := poller.socket.Send(toFrame(command.Request))
	if errors.Is(err, syscall.ENOBUFS) {
		poller.log.Debug("sending failed. send buffer full. retrying.")
		return sendBufferFullError{}
	}
	return err
}

func toFrame(request conf.RequestCommand) canbus.Frame {
	return canbus.Frame{
		ID:   uint32(request.CanId),
		Data: request.CommandBytes,
	}
}
