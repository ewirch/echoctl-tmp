package homeassistant

import "time"

type sensor struct {
	id            string
	deviceClass   string
	stateClass    string
	friendlyName  string
	unit          string
	icon          string
	valueTemplate string
	stats         sensorStats
}
type sensorStats struct {
	stateCharacteristic string
	samplingSize        int
	maxAge              time.Duration
}

type device struct {
	Identifiers  []string `json:"identifiers,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty"`
	Name         string   `json:"name,omitempty"`
}

type entity struct {
	Device            device `json:"device"`
	ObjectId          string `json:"object_id,omitempty"`
	UniqueId          string `json:"unique_id,omitempty"`
	Name              string `json:"name,omitempty"`
	StateTopic        string `json:"state_topic,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	Icon              string `json:"icon,omitempty"`
	DeviceClass       string `json:"device_class,omitempty"`
	StateClass        string `json:"state_class,omitempty"`
	ValueTemplate     string `json:"value_template,omitempty"`
}

func daikinAltherma() device {
	return device{
		Identifiers:  []string{"daikin-0123456789"},
		Manufacturer: "Daikin",
		Model:        "EKHWMX500C",
		Name:         "Altherma M ECH₂O",
	}
}

func (sensor *sensor) asEntity() entity {
	return entity{
		Device:            daikinAltherma(),
		ObjectId:          sensor.id,
		UniqueId:          "daikin/" + sensor.id,
		Name:              sensor.friendlyName,
		StateTopic:        "daikin/" + sensor.id,
		UnitOfMeasurement: sensor.unit,
		Icon:              sensor.icon,
		DeviceClass:       sensor.deviceClass,
		StateClass:        sensor.stateClass,
		ValueTemplate:     sensor.valueTemplate,
	}
}
