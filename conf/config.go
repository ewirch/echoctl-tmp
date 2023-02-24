package conf

import "time"

type Configuration struct {
	Can           Can
	Mqtt          Mqtt
	Subscriptions []Subscription
	Lang          string
	Homeassistant Homeassistant
}

type Can struct {
	Iface string
}

type Mqtt struct {
	Server           string
	ClientId         string `yaml:"client-id"`
	ValueTopicPrefix string `yaml:"value-topic-prefix"`
}

type Subscription struct {
	Command string
	Delay   time.Duration
}

type Homeassistant struct {
	DiscoveryTopicPrefix string `yaml:"discovery-topic-prefix"`
}
