package mqtt

import (
	"echoctl/conf"
	"echoctl/dispatcher"
	"echoctl/flowcontrol"
	"fmt"
)

type mappingNotFoundError struct {
	code uint16
}

var _ flowcontrol.CanSkip = mappingNotFoundError{}
var _ error = mappingNotFoundError{}

type convertError struct {
	cmd   dispatcher.CommandValue
	cause error
}

var _ flowcontrol.CanSkip = convertError{}
var _ error = convertError{}

type notAValueTypeError struct {
	valueType conf.ValueType
	cmd       conf.Command
}

var _ flowcontrol.CanSkip = notAValueTypeError{}
var _ error = notAValueTypeError{}

type valueTypeNotImplementedError struct {
	valueType conf.ValueType
}

var _ flowcontrol.CanSkip = valueTypeNotImplementedError{}
var _ error = valueTypeNotImplementedError{}

func (err mappingNotFoundError) CanSkip() bool {
	return true
}

func (err mappingNotFoundError) Error() string {
	return fmt.Sprintf("no mapping found for code %d", err.code)
}

func (err convertError) CanSkip() bool {
	return true
}

func (err convertError) Error() string {
	return fmt.Sprintf("failed converting command %v: %s", err.cmd, err.cause)
}

func (err notAValueTypeError) CanSkip() bool {
	return true
}

func (err notAValueTypeError) Error() string {
	return fmt.Sprintf("type %d is not a valid ValueType constant, command %v", err.valueType, err.cmd)
}

func (err valueTypeNotImplementedError) CanSkip() bool {
	return true
}

func (err valueTypeNotImplementedError) Error() string {
	return fmt.Sprintf("ValueType constant %d not implemented", err.valueType)
}
