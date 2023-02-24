package homeassistant

import (
	"echoctl/conf"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

func AsEntityJson(cmd *conf.Command, lang string, log *zap.Logger) ([]byte, error) {
	id := cmd.Id
	unit := cmd.Unit
	e := entity{
		Device:            daikinAltherma(),
		ObjectId:          id,
		UniqueId:          "daikin/" + id,
		Name:              localize(id, cmd.Name, lang, log),
		StateTopic:        "daikin/" + id,
		UnitOfMeasurement: mapUnit(unit),
		Icon:              mapIcon(unit),
		DeviceClass:       mapDeviceClass(unit),
		StateClass:        mapStateClass(unit),
	}
	return json.Marshal(e)
}

func mapDeviceClass(unit conf.Unit) string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg:
		return "temperature"
	case conf.UnitBar:
		return "pressure"
	case conf.UnitWh:
		return "energy"
	case conf.UnitW:
		return ""
	case conf.UnitKw:
		return ""
	case conf.UnitNone:
		return ""
	case conf.UnitLh:
		return ""
	case conf.UnitPercent:
		return ""
	case conf.UnitSec:
		return ""
	case conf.UnitMin:
		return ""
	case conf.UnitHour:
		return ""
	default:
		return ""
	}
}

func mapStateClass(unit conf.Unit) string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	if unit == conf.UnitWh {
		return "total_increasing"
	}
	return "measurement"
}

func mapUnit(unit conf.Unit) string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg:
		return "°C"
	case conf.UnitBar:
		return "bar"
	case conf.UnitLh:
		return "L/h"
	case conf.UnitPercent:
		return "%"
	case conf.UnitWh:
		return "Wh"
	case conf.UnitKw:
		return "kW"
	case conf.UnitW:
		return "W"
	case conf.UnitSec:
		return "s"
	case conf.UnitMin:
		return "min"
	case conf.UnitHour:
		return "h"
	case conf.UnitNone:
		return ""
	default:
		return ""
	}
}

func mapIcon(unit conf.Unit) string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg:
		return "mdi:thermometer"
	case conf.UnitBar:
		return "mdi:car-brake-low-pressure"
	case conf.UnitWh:
		return "mdi:lightning-bolt"
	case conf.UnitKw:
		return "mdi:lightning-bolt"
	case conf.UnitW:
		return "mdi:lightning-bolt"
	case conf.UnitLh:
		return ""
	case conf.UnitPercent:
		return ""
	case conf.UnitSec:
		return ""
	case conf.UnitMin:
		return ""
	case conf.UnitHour:
		return ""
	case conf.UnitNone:
		return ""
	default:
		return ""
	}
}

func localize(id string, dict map[string]string, lang string, log *zap.Logger) string {
	text, ok := dict[lang]
	if !ok {
		log.Error("dict is missing translation", zap.String("id", id), zap.String("lang", lang))
		text = maps.Values(dict)[0]
	}
	return text
}
