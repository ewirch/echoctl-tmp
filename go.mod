module echoctl

go 1.19

require github.com/go-daq/canbus v0.2.0

require github.com/dmarkham/enumer v1.5.7

replace github.com/go-daq/canbus v0.2.0 => ../canbus

require (
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/eclipse/paho.mqtt.golang v1.4.2
	github.com/stretchr/testify v1.7.1
	go.uber.org/fx v1.18.2
	go.uber.org/zap v1.16.0
	golang.org/x/exp v0.0.0-20221211140036-ad323defaf05
	golang.org/x/sys v0.2.0
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pascaldekloe/name v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/dig v1.15.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/tools v0.2.0 // indirect
)
