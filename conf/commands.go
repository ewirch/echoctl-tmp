package conf

import (
	"encoding/json"
	"strconv"
	"strings"
)

type CommandBytes []byte

func (c *CommandBytes) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	parts := strings.Split(s, " ")
	if len(parts) == 0 {
		*c = CommandBytes{}
		return nil
	}

	var result []byte
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		num, err := strconv.ParseInt(part, 16, 9)
		if err != nil {
			return err
		}
		result = append(result, byte(num))

	}
	*c = result
	return nil
}

//go:generate go run github.com/dmarkham/enumer -type=Unit -json -trimprefix=Unit -transform lower
type Unit int

const (
	UnitNone Unit = iota
	UnitDeg
	UnitBar
	UnitLh
	UnitPercent
	UnitWh
	UnitKwh
	UnitW
	UnitKw
	UnitSec
	UnitMin
	UnitHour
)

//go:generate go run github.com/dmarkham/enumer -type=ValueType -json -trimprefix=Type -transform lower
type ValueType int

const (
	TypeNoType ValueType = iota
	TypeValue
	TypeLongint
	TypeFloat
)

type CanId uint32

func (c *CanId) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	num, err := strconv.ParseInt(s, 16, 17)
	if err != nil {
		return err
	}

	*c = CanId(num)
	return nil
}

type RequestCommand struct {
	CanId        CanId        `json:"can_id"`
	CommandBytes CommandBytes `json:"command"`
}

type Command struct {
	Id          string            `json:"id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	Request     RequestCommand    `json:"request"`
	Response    RequestCommand    `json:"response"`
	Divisor     float32           `json:"divisor"`
	Writable    bool              `json:"writable"`
	Unit        Unit              `json:"unit"`
	Type        ValueType         `json:"type"`
	ValueCode   map[string]int    `json:"value_code"`
}
