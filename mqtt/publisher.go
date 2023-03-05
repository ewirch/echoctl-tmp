package mqtt

import (
	"echoctl/conf"
	"echoctl/dispatcher"
	"echoctl/flowcontrol"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"gopkg.in/tomb.v2"
	"math"
	"strconv"
)

type publisher struct {
	topicPrefix string
	inbound     <-chan dispatcher.CommandValue
	log         *zap.Logger
	tomb        *tomb.Tomb
	client      mqtt.Client
}

type Publisher interface {
	Publish() *tomb.Tomb
}

var _ Publisher = (*publisher)(nil)

func NewPublisher(topicPrefix string, inbound <-chan dispatcher.CommandValue, client mqtt.Client, log *zap.Logger) Publisher {
	p := &publisher{
		topicPrefix: topicPrefix,
		client:      client,
		inbound:     inbound,
		log:         log,
		tomb:        new(tomb.Tomb),
	}
	return p
}

func (p *publisher) Publish() *tomb.Tomb {
	p.tomb.Go(p.publish)
	return p.tomb
}

func (p *publisher) publish() error {
	for {
		select {
		case cmd := <-p.inbound:
			if err := p.publishCmd(cmd); err != nil {
				if err == tomb.ErrDying {
					return err
				} else if flowcontrol.IsCanSkip(err) {
					p.log.Error("publishing", zap.Error(err))
				} else {
					return fmt.Errorf("publishing to mqtt server: %w", err)
				}
			}
		case <-p.tomb.Dying():
			return tomb.ErrDying
		}
	}
}

func (p *publisher) publishCmd(cmd dispatcher.CommandValue) error {
	value, err := convert(cmd)
	if err != nil {
		return convertError{cmd, err}
	}
	token := p.publishCmdValue(cmd, value)
	select {
	case <-token.Done():
		if err := token.Error(); err == nil {
			return nil
		} else {
			return fmt.Errorf("Publisher.mqttClient.Publish(): %w", err)
		}

	case <-p.tomb.Dying():
		return tomb.ErrDying
	}
}

func (p *publisher) publishCmdValue(cmd dispatcher.CommandValue, value string) mqtt.Token {
	topic := p.getSensorTopic(cmd)
	p.log.Debug(
		"mqtt: publishing",
		zap.String("topic", topic),
		zap.String("value", value),
		zap.String("id", cmd.Cmd.Id),
		zap.Uint16("orig_value", cmd.Value),
		zap.Float32("divisor", cmd.Cmd.Divisor),
		zap.Stringer("unit", cmd.Cmd.Type),
	)
	return p.client.Publish(topic, qos, false, value)
}

func (p *publisher) getSensorTopic(cmd dispatcher.CommandValue) string {
	return p.topicPrefix + "/" + cmd.Cmd.Id
}

func convert(commandValue dispatcher.CommandValue) (string, error) {
	valueType := commandValue.Cmd.Type
	if !conf.ValueType.IsAValueType(valueType) {
		return "", notAValueTypeError{valueType, commandValue.Cmd}
	}
	switch valueType {
	case conf.TypeValue:
		return getLabel(commandValue.Value, commandValue.Cmd.ValueCode)
	case conf.TypeLongint:
		assertNonZeroDivisor(commandValue)
		return strconv.Itoa(int(math.Round(float64(applyDivisor(commandValue.Value, commandValue.Cmd.Divisor))))), nil
	case conf.TypeFloat:
		assertNonZeroDivisor(commandValue)
		return strconv.FormatFloat(float64(applyDivisor(commandValue.Value, commandValue.Cmd.Divisor)), 'f', 4, 32), nil
	case conf.TypeNoType:
		fallthrough
	default:
		return "", valueTypeNotImplementedError{valueType}
	}
}

func assertNonZeroDivisor(commandValue dispatcher.CommandValue) {
	if commandValue.Cmd.Divisor == 0 {
		panic(fmt.Sprintf("Divisor must not be 0: %v", commandValue))
	}
}

func applyDivisor(value uint16, divisor float32) float32 {
	return float32(value) / divisor
}

func getLabel(code uint16, labelMap map[string]int) (string, error) {
	for label, labelCode := range labelMap {
		if int(code) == labelCode {
			return label, nil
		}
	}
	return strconv.Itoa(int(code)), nil
}
