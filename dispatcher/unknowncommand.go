package dispatcher

import (
	"encoding/binary"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type unknownCommandCollector struct {
	commands []unknownCommand
	log      *zap.Logger
}
type unknownCommand struct {
	canId uint32
	code  []byte
	value uint16
}

func newUnknownCommandCollector(log *zap.Logger) *unknownCommandCollector {
	return &unknownCommandCollector{
		log: log,
	}
}

func (c *unknownCommandCollector) addCommand(canId uint32, data []byte) {
	if len(data) < 4 {
		c.log.Error("unknown command is too short: ", zap.Object("command", &unknownCommand{canId, data, 0}))
		return
	}
	cmd := convert(canId, data)
	c.log.Debug(fmt.Sprintf("UNKNOWN COMMAND id: 0x%02X, code: % X, value: %d\n", cmd.canId, cmd.code, cmd.value))

	//if c.exists(cmd) {
	//	return
	//}
	//c.commands = append(c.commands, cmd)
	//c.logCommands()
}

func convert(canId uint32, data []byte) unknownCommand {
	value := binary.BigEndian.Uint16(data[len(data)-2:])
	return unknownCommand{canId, data[:len(data)-2], value}
}

func (c *unknownCommandCollector) exists(cmd unknownCommand) bool {
	for i := range c.commands {
		if cmd.canId == c.commands[i].canId && slices.Equal(cmd.code, c.commands[i].code) {
			return true
		}
	}
	return false
}

func (c *unknownCommandCollector) logCommands() {
	var bytes []byte
	fmt.Appendf(bytes, "======== UNKNOWN COMMANDS ==========\n")
	for i := range c.commands {
		cmd := &c.commands[i]
		bytes = fmt.Appendf(bytes, "id: 0x%02X, code: % X, value: %d\n", cmd.canId, cmd.code, cmd.value)
	}
	c.log.Debug(string(bytes))
}
