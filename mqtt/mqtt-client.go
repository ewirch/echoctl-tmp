package mqtt

import (
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"time"
)

const (
	qos                      = 1
	defaultDisconnectQuiesce = 1000
)

type mqttConfigurer struct {
	log    *zap.Logger
	routes map[string]func(mqtt.Client, mqtt.Message)
}

type Routes = map[string]func(mqtt.Client, mqtt.Message)

type mqttLoggerZapAdapter struct {
	print  func(...interface{})
	printf func(string, ...interface{})
}

func NewClient(serverAddress string, clientId string, user string, password string, log *zap.Logger, routes Routes) (mqtt.Client, error) {
	sugaredLogger := log.Sugar()
	mqtt.ERROR = mqttLoggerZapAdapter{print: sugaredLogger.Error, printf: sugaredLogger.Errorf}
	mqtt.CRITICAL = mqttLoggerZapAdapter{print: sugaredLogger.Error, printf: sugaredLogger.Errorf}
	mqtt.WARN = mqttLoggerZapAdapter{print: sugaredLogger.Warn, printf: sugaredLogger.Warnf}
	mqtt.DEBUG = mqttLoggerZapAdapter{print: sugaredLogger.Debug, printf: sugaredLogger.Debugf}

	configurer := &mqttConfigurer{
		log:    log,
		routes: routes,
	}

	clientOptions := getMqttClientOptions(serverAddress, clientId, user, password)
	// When using QOS2 and CleanSession=FALSE, then it is possible that we will receive messages on topics that we have not subscribed to here (if they were previously subscribed to they are part of the session and survive disconnect/reconnect). Adding a DefaultPublishHandler lets us detect this.
	clientOptions.DefaultPublishHandler = configurer.defaultPublisherHandler
	clientOptions.OnConnect = configurer.onConnect

	client := mqtt.NewClient(clientOptions)
	addRoutes(client, routes)

	connectCtx, connectCtxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer connectCtxCancel()

	return client, configurer.connectWithTimeout(client, connectCtx)
}

func (c *mqttConfigurer) connectWithTimeout(client mqtt.Client, ctx context.Context) error {
	token := client.Connect()
	select {
	case <-token.Done():
		if err := token.Error(); err == nil {
			return nil
		} else {
			return fmt.Errorf("mqttClient.Connect(): %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting con client.Connect()")
	}
}

func addRoutes(client mqtt.Client, routes Routes) {
	// If using QOS=2 and CleanSession = FALSE then messages may be transmitted to us before the subscribe operation completes. Adding routes prior to connecting is a way of ensuring that these messages are processed
	for topic, handler := range routes {
		client.AddRoute(topic, handler)
	}
}

func getMqttClientOptions(serverAddress string, clientId string, user string, password string) *mqtt.ClientOptions {
	clientOptions := mqtt.NewClientOptions()
	clientOptions.AddBroker(serverAddress)
	clientOptions.SetClientID(clientId)
	clientOptions.Username = user
	clientOptions.Password = password
	clientOptions.SetOrderMatters(false)
	clientOptions.ConnectTimeout = 1 * time.Second
	clientOptions.WriteTimeout = 1 * time.Second
	clientOptions.KeepAlive = 10
	clientOptions.PingTimeout = time.Second
	clientOptions.ConnectRetry = true
	clientOptions.AutoReconnect = true
	return clientOptions
}

func (c *mqttConfigurer) defaultPublisherHandler(_ mqtt.Client, msg mqtt.Message) {
	c.log.Sugar().Errorf("Unexpected message: %s\n", msg)
}

func (c *mqttConfigurer) onConnect(client mqtt.Client) {
	c.log.Debug("Connection established")

	// Establish the subscription - doing this here means that it will happen every time a connection is established (useful if opts.CleanSession is TRUE or the broker does not reliably store session data)
	for topic, handler := range c.routes {
		t := client.Subscribe(topic, qos, handler)

		// The connection handler is called in a goroutine, so blocking here would not cause an issue. However, as blocking in other handlers does cause problems, it's best to just assume we should not block.
		go func(topic string) {
			select {
			case <-t.Done():
				if t.Error() != nil {
					//FIXME: wait for the subscription to complete, and fail startup if subscription fails.
					c.log.Error("Error subscribing to %s: %s\n", zap.String("topic", topic), zap.Error(t.Error()))
				} else {
					c.log.Debug("subscribed to", zap.String("topic", topic))
				}
			case <-time.After(1 * time.Second):
				//FIXME: wait for the subscription to complete, and fail startup if subscription fails.
				c.log.Error("time out waiting for client.Subscribe()")
			}
		}(topic)
	}
}

func GetQuiesce(ctx context.Context) uint {
	deadline, ok := ctx.Deadline()
	if !ok {
		return defaultDisconnectQuiesce
	}
	duration := deadline.Sub(time.Now()).Milliseconds()
	if duration < 0 {
		return 0
	}
	return uint(duration)
}

func (m mqttLoggerZapAdapter) Println(v ...interface{}) {
	m.print(v...)
}

func (m mqttLoggerZapAdapter) Printf(format string, v ...interface{}) {
	m.printf(format, v...)
}
