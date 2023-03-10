// Code generated by "enumer -type=Unit -json -trimprefix=Unit -transform lower"; DO NOT EDIT.

package conf

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _UnitName = "nonedegbarlhpercentwhkwwsecminhour"

var _UnitIndex = [...]uint8{0, 4, 7, 10, 12, 19, 21, 23, 24, 27, 30, 34}

const _UnitLowerName = "nonedegbarlhpercentwhkwwsecminhour"

func (i Unit) String() string {
	if i < 0 || i >= Unit(len(_UnitIndex)-1) {
		return fmt.Sprintf("Unit(%d)", i)
	}
	return _UnitName[_UnitIndex[i]:_UnitIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _UnitNoOp() {
	var x [1]struct{}
	_ = x[UnitNone-(0)]
	_ = x[UnitDeg-(1)]
	_ = x[UnitBar-(2)]
	_ = x[UnitLh-(3)]
	_ = x[UnitPercent-(4)]
	_ = x[UnitWh-(5)]
	_ = x[UnitKw-(6)]
	_ = x[UnitW-(7)]
	_ = x[UnitSec-(8)]
	_ = x[UnitMin-(9)]
	_ = x[UnitHour-(10)]
}

var _UnitValues = []Unit{UnitNone, UnitDeg, UnitBar, UnitLh, UnitPercent, UnitWh, UnitKw, UnitW, UnitSec, UnitMin, UnitHour}

var _UnitNameToValueMap = map[string]Unit{
	_UnitName[0:4]:        UnitNone,
	_UnitLowerName[0:4]:   UnitNone,
	_UnitName[4:7]:        UnitDeg,
	_UnitLowerName[4:7]:   UnitDeg,
	_UnitName[7:10]:       UnitBar,
	_UnitLowerName[7:10]:  UnitBar,
	_UnitName[10:12]:      UnitLh,
	_UnitLowerName[10:12]: UnitLh,
	_UnitName[12:19]:      UnitPercent,
	_UnitLowerName[12:19]: UnitPercent,
	_UnitName[19:21]:      UnitWh,
	_UnitLowerName[19:21]: UnitWh,
	_UnitName[21:23]:      UnitKw,
	_UnitLowerName[21:23]: UnitKw,
	_UnitName[23:24]:      UnitW,
	_UnitLowerName[23:24]: UnitW,
	_UnitName[24:27]:      UnitSec,
	_UnitLowerName[24:27]: UnitSec,
	_UnitName[27:30]:      UnitMin,
	_UnitLowerName[27:30]: UnitMin,
	_UnitName[30:34]:      UnitHour,
	_UnitLowerName[30:34]: UnitHour,
}

var _UnitNames = []string{
	_UnitName[0:4],
	_UnitName[4:7],
	_UnitName[7:10],
	_UnitName[10:12],
	_UnitName[12:19],
	_UnitName[19:21],
	_UnitName[21:23],
	_UnitName[23:24],
	_UnitName[24:27],
	_UnitName[27:30],
	_UnitName[30:34],
}

// UnitString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func UnitString(s string) (Unit, error) {
	if val, ok := _UnitNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _UnitNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Unit values", s)
}

// UnitValues returns all values of the enum
func UnitValues() []Unit {
	return _UnitValues
}

// UnitStrings returns a slice of all String values of the enum
func UnitStrings() []string {
	strs := make([]string, len(_UnitNames))
	copy(strs, _UnitNames)
	return strs
}

// IsAUnit returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Unit) IsAUnit() bool {
	for _, v := range _UnitValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for Unit
func (i Unit) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for Unit
func (i *Unit) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Unit should be a string, got %s", data)
	}

	var err error
	*i, err = UnitString(s)
	return err
}
