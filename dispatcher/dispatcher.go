package dispatcher

import (
	"echoctl/conf"
	"echoctl/flowcontrol"
	"encoding/binary"
	"github.com/go-daq/canbus"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	tombPkg "gopkg.in/tomb.v2"
)

type CommandValue struct {
	Cmd   conf.Command
	Value int16
}

type dispatcher struct {
	inbound         <-chan canbus.Frame
	commands        []conf.Command
	toRequestor     chan<- conf.Command
	toMqttPublisher chan<- CommandValue
	tomb            *tombPkg.Tomb
	log             *zap.Logger
	unknownCommands *unknownCommandCollector
}

type Dispatcher interface {
	Dispatch() *tombPkg.Tomb
}

var _ Dispatcher = (*dispatcher)(nil)

func NewDispatcher(inbound <-chan canbus.Frame, commands []conf.Command, toRequestor chan<- conf.Command, toMqttPublisher chan<- CommandValue, log *zap.Logger) Dispatcher {
	tomb := new(tombPkg.Tomb)

	return &dispatcher{
		inbound:         inbound,
		commands:        commands,
		toRequestor:     toRequestor,
		toMqttPublisher: toMqttPublisher,
		tomb:            tomb,
		log:             log,
		unknownCommands: newUnknownCommandCollector(log),
	}
}

func (d *dispatcher) Dispatch() *tombPkg.Tomb {
	d.tomb.Go(d.dispatch)
	return d.tomb
}

func (d *dispatcher) dispatch() error {
	for {
		select {
		case frame := <-d.inbound:
			cmd, err := d.findCmd(frame)
			if flowcontrol.IsCanSkip(err) {
				d.logNotFound(err)
				continue
			}
			if err != nil {
				return err
			}

			d.publishToRequester(cmd)
			d.publishToMqttPublisher(cmd, extractValue(cmd, frame.Data))
		case <-d.tomb.Dying():
			return tombPkg.ErrDying
		}
	}
}

func (d *dispatcher) publishToRequester(cmd conf.Command) {
	select {
	case d.toRequestor <- cmd:
	case <-d.tomb.Dying():
	}
}

func (d *dispatcher) publishToMqttPublisher(cmd conf.Command, value int16) {
	select {
	case d.toMqttPublisher <- CommandValue{cmd, value}:
	case <-d.tomb.Dying():
	}
}

func extractValue(cmd conf.Command, data []byte) int16 {
	lenCommandBytes := len(cmd.Response.CommandBytes)
	return int16(binary.BigEndian.Uint16(data[lenCommandBytes:]))
}

// findCmd searches for a command matching frame in dispatcher.commands.
// Returns found command. Returned errors implement flowcontrol.CanSkip attribute.
func (d *dispatcher) findCmd(frame canbus.Frame) (conf.Command, error) {
	for _, cmd := range d.commands {
		if equals(frame, cmd.Response) {
			return cmd, nil
		}
		if equals(frame, cmd.Request) {
			return conf.Command{}, commandIsRequestError{}
		}
	}
	return conf.Command{}, commandNotFoundError{frame.ID, frame.Data}
}

func (d *dispatcher) logNotFound(err error) {
	if notFoundError, isNotFound := err.(commandNotFoundError); isNotFound {
		d.unknownCommands.addCommand(notFoundError.canId, notFoundError.commandBytes)
		//if len(notFoundError.commandBytes) < 2 {
		//	d.log.Debug("======== UNKNOWN COMMAND ==========", zap.String("id", fmt.Sprintf("0x%X", notFoundError.canId)), zap.String("data", fmt.Sprintf("% X", notFoundError.commandBytes)))
		//} else {
		//	data := notFoundError.commandBytes
		//	value := binary.BigEndian.Uint16(data[len(data)-2:])
		//	d.log.Debug("======== UNKNOWN COMMAND ==========", zap.String("id", fmt.Sprintf("0x%X", notFoundError.canId)), zap.String("data", fmt.Sprintf("% X", notFoundError.commandBytes)), zap.Uint16("value", value))
		//}
	}
}

func equals(frame canbus.Frame, cmd conf.RequestCommand) bool {
	return frame.ID == uint32(cmd.CanId) && slices.Equal(frame.Data[:len(cmd.CommandBytes)], cmd.CommandBytes)
}
