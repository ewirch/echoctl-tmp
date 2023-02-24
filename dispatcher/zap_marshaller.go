package dispatcher

import (
	"fmt"
	"go.uber.org/zap/zapcore"
)

type unknownCommands []unknownCommand

func (commands unknownCommands) MarshalLogArray(encoder zapcore.ArrayEncoder) error {
	for i := range commands {
		if err := encoder.AppendObject(&commands[i]); err != nil {
			return err
		}
	}
	return nil
}

func (u *unknownCommand) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("canId", fmt.Sprintf("0x%X", u.canId))
	encoder.AddString("code", fmt.Sprintf("% X", u.code))
	encoder.AddUint16("value", u.value)
	return nil
}
