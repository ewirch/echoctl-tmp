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
	valueCodes := subscription.Command.ValueCode
	e := entity{
		Device:                    daikinAltherma(),
		ObjectId:                  strPtr(id),
		UniqueId:                  strPtr("daikin/" + id),
		Name:                      localize(id, subscription.Command.Name, lang, log),
		StateTopic:                strPtr("daikin/" + id),
		UnitOfMeasurement:         mapUnit(unit),
		Icon:                      mapIcon(unit),
		DeviceClass:               mapDeviceClass(unit, valueCodes),
		StateClass:                mapStateClass(unit),
		ExpiresAfter:              int64Ptr(expiresAfter(subscription)),
		SuggestedDisplayPrecision: suggestedDisplayPrecision(unit),
	}
	return json.Marshal(e)
}

func suggestedDisplayPrecision(unit conf.Unit) *int {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg, conf.UnitBar, conf.UnitWh, conf.UnitW:
		return intPtr(2)
	case conf.UnitKw, conf.UnitKwh:
		return intPtr(3)
	case conf.UnitLh:
		return intPtr(1)
	case conf.UnitPercent, conf.UnitSec, conf.UnitMin, conf.UnitHour:
		return intPtr(0)
	case conf.UnitNone:
		return nil
	default:
		return nil
	}
}

func mapDeviceClass(unit conf.Unit, valueCodes map[string]int) *string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg:
		return strPtr("temperature")
	case conf.UnitBar:
		return strPtr("pressure")
	case conf.UnitWh, conf.UnitKwh:
		return strPtr("energy")
	case conf.UnitW, conf.UnitKw:
		return strPtr("power")
	case conf.UnitNone:
		if len(valueCodes) > 0 {
			return strPtr("enum")
		} else {
			return nil
		}
	case conf.UnitLh, conf.UnitPercent, conf.UnitSec, conf.UnitMin, conf.UnitHour:
		return nil
	default:
		return nil
	}
}

func mapStateClass(unit conf.Unit) *string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg, conf.UnitBar, conf.UnitW, conf.UnitKw, conf.UnitLh, conf.UnitPercent, conf.UnitSec, conf.UnitMin, conf.UnitHour:
		return strPtr("measurement")
	case conf.UnitWh, conf.UnitKwh:
		return strPtr("total_increasing")
	case conf.UnitNone:
		return nil
	default:
		return nil
	}
}

func mapUnit(unit conf.Unit) *string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg:
		return strPtr("°C")
	case conf.UnitBar:
		return strPtr("bar")
	case conf.UnitLh:
		return strPtr("L/h")
	case conf.UnitPercent:
		return strPtr("%")
	case conf.UnitWh:
		return strPtr("Wh")
	case conf.UnitKwh:
		return strPtr("kWh")
	case conf.UnitW:
		return strPtr("W")
	case conf.UnitKw:
		return strPtr("kW")
	case conf.UnitSec:
		return strPtr("s")
	case conf.UnitMin:
		return strPtr("min")
	case conf.UnitHour:
		return strPtr("h")
	case conf.UnitNone:
		return nil
	default:
		return nil
	}
}

func mapIcon(unit conf.Unit) *string {
	if !unit.IsAUnit() {
		panic(fmt.Sprintf("%v is not a conf.Unit", unit))
	}
	switch unit {
	case conf.UnitDeg:
		return strPtr("mdi:thermometer")
	case conf.UnitBar:
		return strPtr("mdi:car-brake-low-pressure")
	case conf.UnitWh, conf.UnitKwh, conf.UnitKw, conf.UnitW:
		return strPtr("mdi:lightning-bolt")
	case conf.UnitLh, conf.UnitPercent, conf.UnitSec, conf.UnitMin, conf.UnitHour, conf.UnitNone:
		return nil
	default:
		return nil
	}
}

func localize(id string, dict map[string]string, lang string, log *zap.Logger) *string {
	text, ok := dict[lang]
	if !ok {
		log.Error("dict is missing translation", zap.String("id", id), zap.String("lang", lang))
		text = maps.Values(dict)[0]
	}
	return &text
}

// expiresAfter returns the duration in seconds after the last update, after which the sensor can be considered as unavailable.
func expiresAfter(subscription *can.Subscription) int64 {
	// We allow twice the update duration. If the sensor was not updated at that time, probably something is wrong, and the sensor should be considered unavailable.
	return int64(subscription.Delay.Seconds() * 2)
}

func int64Ptr(v int64) *int64 { return &v }

func intPtr(v int) *int { return &v }

func strPtr(v string) *string { return &v }
