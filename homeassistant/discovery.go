package homeassistant

import (
	"echoctl/can"
	"echoctl/conf"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"gopkg.in/tomb.v2"
)

const (
	qos = 1
)

type discovery struct {
	subscriptions        []can.Subscription
	discoveryTopicPrefix string
	lang                 string
	log                  *zap.Logger
	tomb                 *tomb.Tomb
	client               mqtt.Client
}

type DiscoveryAnnouncer interface {
	Announce() *tomb.Tomb
}

var _ DiscoveryAnnouncer = (*discovery)(nil)

func NewDiscoveryAnnouncer(subscriptions []can.Subscription, discoveryTopicPrefix string, lang string, client mqtt.Client, log *zap.Logger) DiscoveryAnnouncer {
	p := &discovery{
		discoveryTopicPrefix: discoveryTopicPrefix,
		subscriptions:        subscriptions,
		lang:                 lang,
		client:               client,
		log:                  log,
		tomb:                 new(tomb.Tomb),
	}
	return p
}

func (p *discovery) Announce() *tomb.Tomb {
	p.tomb.Go(p.announce)
	return p.tomb
}

func (p *discovery) announce() error {
	err := p.publishNodeConfigurations()
	if err != nil {
		return err
	}
	// FIXME subscribe to discovery topic
	return nil
}

func (p *discovery) publishNodeConfigurations() error {
	for i := range p.subscriptions {
		err := p.publishNodeConf(&p.subscriptions[i].Command)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *discovery) publishNodeConf(command *conf.Command) error {
	json, err := AsEntityJson(command, p.lang, p.log)
	if err != nil {
		return fmt.Errorf("publish node configuration for command %s: %w", command.Id, err)
	}

	token := p.client.Publish(
		p.discoveryTopicPrefix+"/sensor/daikin_altherma/"+command.Id+"/config",
		qos,
		true,
		json,
	)

	select {
	case <-p.tomb.Dying():
		return tomb.ErrDying
	case <-token.Done():
		return token.Error()
	}
}
