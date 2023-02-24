package main

import (
	"echoctl/can"
	"echoctl/conf"
	"echoctl/dispatcher"
	"echoctl/homeassistant"
	"echoctl/mqtt"
	"fmt"
	"github.com/benbjohnson/clock"
	"github.com/docopt/docopt-go"
	phaoMqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-daq/canbus"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"os"
)

/*
sudo ip link add dev vcan0 type vcan && sudo ip link set up vcan0
candump -tz vcan0
cansend vcan0 123#00FFAA5501020304
cansend vcan0 '180#3210FAC0F60100'
cansend vcan0 '180#3210FA01800100'
cansend vcan0 '180#3210FAC0F60300'
*/

const version = "1.0"
const usage = `Altherma ECH₂O Control.

Usage:
  echoctl [options]

Options:
  -h --help     Show this screen.
  --version     Show version.
  --debug       Turn on debug logging [default: false].
`

type commandLineOptions struct {
	Debug bool
}

func main() {
	cliOpts := parseArgs()

	commands, err := conf.ReadCommands("commands_hpsu.json")
	if err != nil {
		panic(err)
	}
	configuration, err := conf.ReadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	subscriptions := attachCommand(configuration.Subscriptions, commands)

	dispatcherToRequestor := make(chan conf.Command, 10)
	dispatcherToMqttPublisher := make(chan dispatcher.CommandValue, 10)
	canReaderToDispatcher := make(chan canbus.Frame, 10)

	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.Named("fx")}
		}),
		fx.Provide(
			func() (can.Socket, error) {
				return can.NewSocket(configuration.Can.Iface)
			},
			func(socket can.Socket, log *zap.Logger) can.Poller {
				return can.NewPoller(socket, subscriptions, dispatcherToRequestor, clock.New(), log.Named("poller"))
			},
			func(log *zap.Logger) dispatcher.Dispatcher {
				return dispatcher.NewDispatcher(canReaderToDispatcher, maps.Values(commands), dispatcherToRequestor, dispatcherToMqttPublisher, log.Named("disp"))
			},
			func(log *zap.Logger) (phaoMqtt.Client, error) {
				return mqtt.NewClient(configuration.Mqtt.Server, configuration.Mqtt.ClientId, log.Named("mqtt"), nil)
			},
			func(client phaoMqtt.Client, log *zap.Logger) mqtt.Publisher {
				return mqtt.NewPublisher(configuration.Mqtt.ValueTopicPrefix, dispatcherToMqttPublisher, client, log.Named("publ"))
			},
			func(client phaoMqtt.Client, log *zap.Logger) homeassistant.DiscoveryAnnouncer {
				return homeassistant.NewDiscoveryAnnouncer(subscriptions, configuration.Homeassistant.DiscoveryTopicPrefix, configuration.Lang, client, log.Named("anou"))
			},
			func(socket can.Socket, log *zap.Logger) can.Reader {
				return can.NewReader(socket, canReaderToDispatcher, log.Named("reader"))
			},
			getLogConfig(cliOpts.Debug).Build,
		),

		fx.Invoke(func(publisher mqtt.Publisher, poller can.Poller, dispatcher dispatcher.Dispatcher, reader can.Reader, shutdowner fx.Shutdowner, discoveryAnnouncer homeassistant.DiscoveryAnnouncer, lc fx.Lifecycle, client phaoMqtt.Client, socket can.Socket, log *zap.Logger) {
			daemonize(
				lc,
				shutdowner,
				fx.DefaultTimeout,
				log,
				client,
				socket,
				publisher.Publish(),
				poller.Poll(),
				dispatcher.Dispatch(),
				reader.Read(),
				discoveryAnnouncer.Announce(),
			)
		}),
	).Run()
}

func parseArgs() commandLineOptions {
	arguments, _ := docopt.ParseArgs(usage, os.Args[1:], version)
	var cliOpts commandLineOptions
	err := arguments.Bind(&cliOpts)
	if err != nil {
		panic(err)
	}
	return cliOpts
}

func getLogConfig(debugLogging bool) zap.Config {
	config := zap.NewDevelopmentConfig()
	if debugLogging {
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		config.Level.SetLevel(zap.InfoLevel)
	}
	return config
}

func attachCommand(subscriptions []conf.Subscription, commands map[string]conf.Command) []can.Subscription {
	result := make([]can.Subscription, len(subscriptions))
	for i := range subscriptions {
		var ok bool
		result[i].Command, ok = commands[subscriptions[i].Command]
		if !ok {
			panic(fmt.Errorf("error parsing configuration file: command '%s' not found in commands_hpsu.json", subscriptions[i].Command))
		}
		result[i].Delay = subscriptions[i].Delay
	}

	return result
}
