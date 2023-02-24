// Code generated by "enumer -type=ValueType -json -trimprefix=Type -transform lower"; DO NOT EDIT.

package conf

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _ValueTypeName = "notypevaluelongintfloat"

var _ValueTypeIndex = [...]uint8{0, 6, 11, 18, 23}

const _ValueTypeLowerName = "notypevaluelongintfloat"

func (i ValueType) String() string {
	if i < 0 || i >= ValueType(len(_ValueTypeIndex)-1) {
		return fmt.Sprintf("ValueType(%d)", i)
	}
	return _ValueTypeName[_ValueTypeIndex[i]:_ValueTypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _ValueTypeNoOp() {
	var x [1]struct{}
	_ = x[TypeNoType-(0)]
	_ = x[TypeValue-(1)]
	_ = x[TypeLongint-(2)]
	_ = x[TypeFloat-(3)]
}

var _ValueTypeValues = []ValueType{TypeNoType, TypeValue, TypeLongint, TypeFloat}

var _ValueTypeNameToValueMap = map[string]ValueType{
	_ValueTypeName[0:6]:        TypeNoType,
	_ValueTypeLowerName[0:6]:   TypeNoType,
	_ValueTypeName[6:11]:       TypeValue,
	_ValueTypeLowerName[6:11]:  TypeValue,
	_ValueTypeName[11:18]:      TypeLongint,
	_ValueTypeLowerName[11:18]: TypeLongint,
	_ValueTypeName[18:23]:      TypeFloat,
	_ValueTypeLowerName[18:23]: TypeFloat,
}

var _ValueTypeNames = []string{
	_ValueTypeName[0:6],
	_ValueTypeName[6:11],
	_ValueTypeName[11:18],
	_ValueTypeName[18:23],
}

// ValueTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ValueTypeString(s string) (ValueType, error) {
	if val, ok := _ValueTypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _ValueTypeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to ValueType values", s)
}

// ValueTypeValues returns all values of the enum
func ValueTypeValues() []ValueType {
	return _ValueTypeValues
}

// ValueTypeStrings returns a slice of all String values of the enum
func ValueTypeStrings() []string {
	strs := make([]string, len(_ValueTypeNames))
	copy(strs, _ValueTypeNames)
	return strs
}

// IsAValueType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ValueType) IsAValueType() bool {
	for _, v := range _ValueTypeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for ValueType
func (i ValueType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for ValueType
func (i *ValueType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ValueType should be a string, got %s", data)
	}

	var err error
	*i, err = ValueTypeString(s)
	return err
}
