package homeassistant

import (
	"echoctl/can"
	"echoctl/conf"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

func AsEntityJson(subscription *can.Subscription, lang string, log *zap.Logger) ([]byte, error) {
	id := subscription.Command.Id
	unit := subscription.Command.Unit
	e := entity{
		Device:                    daikinAltherma(),
		ObjectId:                  id,
		UniqueId:                  "daikin/" + id,
		Name:                      localize(id, subscription.Command.Name, lang, log),
		StateTopic:                "daikin/" + id,
		UnitOfMeasurement:         mapUnit(unit),
		Icon:                      mapIcon(unit),
		DeviceClass:               mapDeviceClass(unit),
		StateClass:                mapStateClass(unit),
		ExpiresAfter:              expiresAfter(subscription),
		SuggestedDisplayPrecision: suggestedDisplayPrecision(unit),
	}
	return json.Marshal(e)
}

func suggestedDisplayPrecision(unit conf.Unit) int {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg, conf.UnitBar, conf.UnitWh, conf.UnitW:
		return 2
	case conf.UnitKw, conf.UnitKwh:
		return 3
	case conf.UnitLh:
		return 1
	case conf.UnitPercent, conf.UnitNone, conf.UnitSec, conf.UnitMin, conf.UnitHour:
		return 0
	default:
		return 0
	}
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
	case conf.UnitWh, conf.UnitKwh:
		return "energy"
	case conf.UnitW, conf.UnitKw:
		return "power"
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
	switch unit {
	case conf.UnitDeg, conf.UnitBar, conf.UnitW, conf.UnitKw, conf.UnitLh, conf.UnitPercent, conf.UnitSec, conf.UnitMin, conf.UnitHour:
		return "measurement"
	case conf.UnitWh, conf.UnitKwh:
		return "total_increasing"
	case conf.UnitNone:
		return ""
	default:
		return ""
	}
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
	case conf.UnitKwh:
		return "kWh"
	case conf.UnitW:
		return "W"
	case conf.UnitKw:
		return "kW"
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
	case conf.UnitWh, conf.UnitKwh, conf.UnitKw, conf.UnitW:
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

// expiresAfter returns the duration in seconds after the last update, after which the sensor can be considered as unavailable.
func expiresAfter(subscription *can.Subscription) int64 {
	// We allow twice the update duration. If the sensor was not updated at that time, probably something is wrong, and the sensor should be considered unavailable.
	return int64(subscription.Delay.Seconds() * 2)
}
