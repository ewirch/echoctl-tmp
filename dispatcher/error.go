package dispatcher

import (
	"echoctl/conf"
	"echoctl/flowcontrol"
	"fmt"
)

type commandNotFoundError struct {
	canId        uint32
	commandBytes conf.CommandBytes
}

var _ flowcontrol.CanSkip = commandNotFoundError{}
var _ flowcontrol.ShouldLog = commandNotFoundError{}
var _ error = commandNotFoundError{}

type commandIsRequestError struct {
}

var _ flowcontrol.CanSkip = commandIsRequestError{}
var _ error = commandIsRequestError{}

func (err commandNotFoundError) CanSkip() bool {
	return true
}

func (err commandNotFoundError) ShouldLog() bool {
	return true
}

func (err commandNotFoundError) Error() string {
	return fmt.Sprintf("command (ID: 0x%X, Data (hex): % X) not found", err.canId, err.commandBytes)
}

func (c commandIsRequestError) CanSkip() bool {
	return true
}

func (c commandIsRequestError) Error() string {
	return ""
}
