package can

import (
	"echoctl/conf"
	"echoctl/flowcontrol"
	"errors"
	"github.com/benbjohnson/clock"
	"github.com/go-daq/canbus"
	"go.uber.org/zap"
	"gopkg.in/tomb.v2"
	"syscall"
	"time"
)

type Subscription struct {
	Command conf.Command
	Delay   time.Duration
}

type poller struct {
	socket        Socket
	subscriptions []Subscription
	inbound       <-chan conf.Command
	tomb          *tomb.Tomb
	clock         clock.Clock
	log           *zap.Logger
}

// Poller sends periodic commands to a can-bus socket, following the specified schedule. Poller does not wait for a reply. It relies on Reader to read the reply from can-bus. The Reader passes the received frame to the Dispatcher, and the Dispatcher passes it on to Poller.
type Poller interface {
	Poll() *tomb.Tomb
}

var _ Poller = (*poller)(nil)

type job struct {
	nextTime     time.Time
	subscription Subscription
}

func NewPoller(socket Socket, subscriptions []Subscription, inbound <-chan conf.Command, clck clock.Clock, log *zap.Logger) Poller {
	return &poller{
		socket:        socket,
		subscriptions: subscriptions,
		inbound:       inbound,
		tomb:          new(tomb.Tomb),
		clock:         clck,
		log:           log,
	}
}

func (poller *poller) Poll() *tomb.Tomb {
	poller.tomb.Go(poller.poll)
	return poller.tomb
}

func (poller *poller) poll() error {
	schedule := poller.createSchedule(poller.subscriptions)
	for {
		job := getNextJob(schedule)
		timer := poller.clock.Timer(poller.clock.Until(job.nextTime))
		select {
		case <-poller.tomb.Dying():
			return tomb.ErrDying

		case <-timer.C:
			err := poller.sendCommand(job.subscription.Command)
			if flowcontrol.IsShouldRetry(err) {
				// Retry sending command, but delay a bit, to not directly fail again on retry.
				poller.schedule(job, 100*time.Millisecond)
			} else if err != nil {
				return err
			} else {
				poller.schedule(job, job.subscription.Delay)
			}

		case cmd := <-poller.inbound:
			job := findJob(schedule, cmd)
			if job != nil {
				poller.schedule(job, job.subscription.Delay)
			}

		}
		timer.Stop()
	}
}

func (poller *poller) closeSocket(socket *canbus.Socket) {
	if err := socket.Close(); err != nil {
		poller.tomb.Kill(err)
	}
}

func findJob(schedule []job, cmd conf.Command) *job {
	for i := range schedule {
		if schedule[i].subscription.Command.Id == cmd.Id {
			return &schedule[i]
		}
	}
	return nil
}

func (poller *poller) sendCommand(command conf.Command) error {
	poller.log.Debug("sending", zap.String("command", command.Id))
	requestFrame := canbus.Frame{
		ID:   uint32(command.Request.CanId),
		Data: command.Request.CommandBytes,
	}

	_, err := poller.socket.Send(requestFrame)
	if errors.Is(err, syscall.ENOBUFS) {
		poller.log.Debug("sending failed, send buffer full. retrying.")
		return sendBufferFullError{}
	}
	if err != nil {
		return err
	}
	return nil
}

func getNextJob(schedule []job) (nextJob *job) {
	nextJob = &schedule[0]
	for i := range schedule {
		if schedule[i].nextTime.Before(nextJob.nextTime) {
			nextJob = &schedule[i]
		}
	}
	return
}

func (poller *poller) createSchedule(subscriptions []Subscription) []job {
	schedule := make([]job, len(subscriptions))
	now := poller.clock.Now()
	for i, subscription := range subscriptions {
		schedule[i].subscription = subscription
		schedule[i].nextTime = now
	}
	return schedule
}

func (poller *poller) schedule(job *job, duration time.Duration) {
	job.nextTime = poller.clock.Now().Add(duration)
}
