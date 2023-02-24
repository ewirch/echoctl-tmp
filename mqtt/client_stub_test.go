package mqtt_test

import mqtt "github.com/eclipse/paho.mqtt.golang"

type Frame struct {
	topic    string
	qos      byte
	retained bool
	payload  interface{}
}

type ClientStub struct {
	published chan Frame
}

func NewClientStub() *ClientStub {
	c := new(ClientStub)
	c.published = make(chan Frame, 1)
	return c
}

func (c *ClientStub) IsConnected() bool {
	return true
}

func (c *ClientStub) IsConnectionOpen() bool {
	return true
}

func (c *ClientStub) Connect() mqtt.Token {
	panic("not implemented")
}

func (c *ClientStub) Disconnect(uint) {
	panic("not implemented")
}

func (c *ClientStub) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	c.published <- Frame{
		topic:    topic,
		qos:      qos,
		retained: retained,
		payload:  payload,
	}
	return new(mqtt.DummyToken)
}

func (c *ClientStub) Subscribe(string, byte, mqtt.MessageHandler) mqtt.Token {
	panic("not implemented")
}

func (c *ClientStub) SubscribeMultiple(map[string]byte, mqtt.MessageHandler) mqtt.Token {
	panic("not implemented")
}

func (c *ClientStub) Unsubscribe(...string) mqtt.Token {
	panic("not implemented")
}

func (c *ClientStub) AddRoute(string, mqtt.MessageHandler) {
	panic("not implemented")
}

func (c *ClientStub) OptionsReader() mqtt.ClientOptionsReader {
	panic("not implemented")
}

func (c *ClientStub) GetPublished() <-chan Frame {
	return c.published
}

var _ mqtt.Client = (*ClientStub)(nil)
