package homeassistant

type device struct {
	Identifiers  []string `json:"identifiers,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty"`
	Name         string   `json:"name,omitempty"`
}

type entity struct {
	Device                    device `json:"device"`
	ObjectId                  string `json:"object_id,omitempty"`
	UniqueId                  string `json:"unique_id,omitempty"`
	Name                      string `json:"name,omitempty"`
	StateTopic                string `json:"state_topic,omitempty"`
	UnitOfMeasurement         string `json:"unit_of_measurement,omitempty"`
	Icon                      string `json:"icon,omitempty"`
	DeviceClass               string `json:"device_class,omitempty"`
	StateClass                string `json:"state_class,omitempty"`
	ValueTemplate             string `json:"value_template,omitempty"`
	ExpiresAfter              int64  `json:"expires_after"`
	SuggestedDisplayPrecision int    `json:"suggested_display_precision"`
}

func daikinAltherma() device {
	return device{
		Identifiers:  []string{"daikin-0123456789"},
		Manufacturer: "Daikin",
		Model:        "EKHWMX500C",
		Name:         "Altherma M ECH₂O",
	}
}
