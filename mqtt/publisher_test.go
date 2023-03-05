package mqtt_test

import (
	"echoctl/conf"
	"echoctl/dispatcher"
	"echoctl/mqtt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestNewPublisher(t *testing.T) {
	t.Run("Publish exits before timeout", func(t *testing.T) {
		t.Parallel()

		_, _, publisher := NewPublisher("")

		tmb := publisher.Publish()
		tmb.Kill(nil)
		select {
		case <-time.After(time.Second):
			t.Errorf("Publisher did not exit after 1s")
		case <-tmb.Dead():
		}
	})

	t.Run("Publishes TypeLongint command", func(t *testing.T) {
		t.Parallel()

		toPublisher, mqttClient, publisher := NewPublisher("topic_prfx")

		startAndRun(t, publisher, func() {
			toPublisher <- NewLongIntCommand("temp_ext", 16, 1)
			readWithTimeout(t, mqttClient.GetPublished(), func(frame Frame) {
				assert.Equal(t, "topic_prfx/temp_ext", frame.topic, "Topic should be the same")
				assert.Equal(t, "16", frame.payload, "Payload should be the same")
			})
		})
	})

	t.Run("Publishes valueType command", func(t *testing.T) {
		t.Parallel()

		toPublisher, mqttClient, publisher := NewPublisher("")

		startAndRun(t, publisher, func() {
			toPublisher <- NewValueCommand("mode", 4, map[string]int{
				"cooling":   2,
				"defrost":   3,
				"heating":   1,
				"standby":   0,
				"hot water": 4,
			})
			select {
			case <-time.After(time.Second):
				t.Errorf("Expected frame was not published in 1s")
			case frame := <-mqttClient.GetPublished():
				assert.Equal(t, "hot water", frame.payload, "Payload should be the same")
			}
		})
	})
	t.Run("Publishes TypeFloat command", func(t *testing.T) {
		t.Parallel()

		toPublisher, mqttClient, publisher := NewPublisher("")

		startAndRun(t, publisher, func() {
			toPublisher <- NewFloatCommand("temp", 12345, 10000)
			select {
			case <-time.After(time.Second):
				t.Errorf("Expected frame was not published in 1s")
			case frame := <-mqttClient.GetPublished():
				assert.Equal(t, "1.2345", frame.payload, "Payload should be the same")
			}
		})
	})
}

func startAndRun(t *testing.T, publisher mqtt.Publisher, f func()) {
	tmb := publisher.Publish()

	f()

	tmb.Kill(nil)
	select {
	case <-tmb.Dead():
	case <-time.After(time.Second):
		t.Log("Publisher failed to shut down in 1s")
	}
}

func readWithTimeout(t *testing.T, inChan <-chan Frame, consumer func(frame Frame)) {
	select {
	case <-time.After(time.Second):
		t.Errorf("Expected frame was not published in 1s")
	case frame := <-inChan:
		consumer(frame)
	}
}

func NewPublisher(topicPrefix string) (chan dispatcher.CommandValue, *ClientStub, mqtt.Publisher) {
	toPublisher := make(chan dispatcher.CommandValue, 1)
	log := zap.NewNop()
	mqttClient := NewClientStub()
	publisher := mqtt.NewPublisher(topicPrefix, toPublisher, mqttClient, log)
	return toPublisher, mqttClient, publisher
}

func NewLongIntCommand(id string, value uint16, divisor float32) dispatcher.CommandValue {
	return dispatcher.CommandValue{
		Cmd: conf.Command{
			Id:      id,
			Type:    conf.TypeLongint,
			Divisor: divisor,
		},
		Value: value,
	}
}

func NewValueCommand(id string, value uint16, labelMap map[string]int) dispatcher.CommandValue {
	return dispatcher.CommandValue{
		Cmd: conf.Command{
			Id:        id,
			Type:      conf.TypeValue,
			ValueCode: labelMap,
		},
		Value: value,
	}
}

func NewFloatCommand(id string, value uint16, divisor float32) dispatcher.CommandValue {
	return dispatcher.CommandValue{
		Cmd: conf.Command{
			Id:      id,
			Type:    conf.TypeFloat,
			Divisor: divisor,
		},
		Value: value,
	}
}
