package enumhelper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var nullBytes = []byte("null")

// IsNull returns true iff err is an instance of IsNullError.
func IsNull(err error) bool {
	var x IsNullError
	return errors.As(err, &x)
}

// EnumData holds data about one particular enum value.
type EnumData struct {
	// GoName is the Go constant name for this enum value.
	GoName string

	// Name is the string representation of this enum value.
	Name string

	// JSON is the JSON representation of this enum value.
	//
	// Optional; it is inferred from Name if not set.
	JSON []byte

	// Aliases is a list of zero or more aliases for this enum value.
	//
	// Optional.
	Aliases []string
}

// MakeAllowedEnumNames returns the list of canonical string representations
// for this enum.
func MakeAllowedEnumNames(enumData []EnumData) []string {
	out := make([]string, len(enumData))
	for i, row := range enumData {
		out[i] = row.Name
	}
	return out
}

// DereferenceEnumData returns enumData[value] or panics with InvalidEnumValueError.
func DereferenceEnumData(enumName string, enumData []EnumData, value uint) EnumData {
	if limit := uint(len(enumData)); value >= limit {
		panic(InvalidEnumValueError{
			Type:  enumName,
			Value: value,
			Limit: limit,
		})
	}
	return enumData[value]
}

// MarshalEnumToJSON marshals this enum value to JSON.  It may panic with
// InvalidEnumValueError if the enum value is out of range.
func MarshalEnumToJSON(enumName string, enumData []EnumData, value uint) ([]byte, error) {
	row := DereferenceEnumData(enumName, enumData, value)
	if row.JSON == nil {
		return json.Marshal(row.Name)
	}
	return row.JSON, nil
}

// ParseEnum parses an enum value.  Returns InvalidEnumNameError if the string
// cannot be parsed.
func ParseEnum(enumName string, enumData []EnumData, str string) (uint, error) {
	for index, row := range enumData {
		if strings.EqualFold(str, row.Name) || strings.EqualFold(str, row.GoName) {
			return uint(index), nil
		}
		for _, alias := range row.Aliases {
			if strings.EqualFold(str, alias) {
				return uint(index), nil
			}
		}
	}

	return 0, InvalidEnumNameError{
		Type:    enumName,
		Name:    str,
		Allowed: MakeAllowedEnumNames(enumData),
	}
}

// UnmarshalEnumFromJSON unmarshals an enum value from JSON.  Returns
// IsNullError, InvalidEnumNameError, or InvalidEnumValueError if a JSON value
// was parsed but could not be unmarshaled as an enum value.
func UnmarshalEnumFromJSON(enumName string, enumData []EnumData, raw []byte) (uint, error) {
	if raw == nil {
		panic(errors.New("[]byte is nil"))
	}

	if bytes.Equal(raw, nullBytes) {
		return 0, IsNullError{}
	}

	for index, row := range enumData {
		if row.JSON != nil && bytes.Equal(raw, row.JSON) {
			return uint(index), nil
		}
	}

	var str string
	err0 := json.Unmarshal(raw, &str)
	if err0 == nil {
		return ParseEnum(enumName, enumData, str)
	}

	var num uint
	err1 := json.Unmarshal(raw, &num)
	limit := uint(len(enumData))
	if err1 == nil && num >= limit {
		return 0, InvalidEnumValueError{
			Type:  enumName,
			Value: num,
			Limit: limit,
		}
	}
	if err1 == nil {
		return num, nil
	}

	return 0, err0

}

// type IsNullError {{{

// IsNullError indicates that a JSON null value was parsed.
type IsNullError struct{}

// Error fulfills the error interface.
func (IsNullError) Error() string {
	return "JSON value is null"
}

var _ error = IsNullError{}

// }}}

// type InvalidEnumNameError {{{

// InvalidEnumNameError indicates an enum whose string representation could not
// be recognized.
type InvalidEnumNameError struct {
	Type    string
	Name    string
	Allowed []string
}

// Error fulfills the error interface.
func (err InvalidEnumNameError) Error() string {
	if len(err.Allowed) == 0 {
		return fmt.Sprintf("invalid %s name %q", err.Type, err.Name)
	}
	return fmt.Sprintf("invalid %s name %q; must be one of %q", err.Type, err.Name, err.Allowed)
}

var _ error = InvalidEnumNameError{}

// }}}

// type InvalidEnumValueError {{{

// InvalidEnumValueError indicates an enum whose numeric value is out of range.
type InvalidEnumValueError struct {
	Type  string
	Value uint
	Limit uint
}

// Error fulfills the error interface.
func (err InvalidEnumValueError) Error() string {
	if err.Limit == 0 {
		return fmt.Sprintf("invalid %s value %d", err.Type, err.Value)
	}
	return fmt.Sprintf("invalid %s value %d; must be < %d", err.Type, err.Value, err.Limit)
}

var _ error = InvalidEnumValueError{}

// }}}
