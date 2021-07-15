package enumhelper

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// BitfieldData holds data about one particular bitfield bit.
type BitfieldData struct {
	// GoName is the Go constant name for this bit.
	GoName string

	// Name is the string representation of this bit.
	Name string

	// Aliases is a list of zero or more aliases for this bit.
	//
	// Optional.
	Aliases []string
}

// AnnotatedBitfieldData extends BitfieldData with some auto-populated fields.
type AnnotatedBitfieldData struct {
	BitfieldData

	// Index is the index of this bit; always between 0 and 63.
	Index uint

	// Bit is the value of this bit; always equal to (1 << Index).
	Bit uint64
}

// BitfieldType holds data about a bitfield type.
type BitfieldType struct {
	// Type gives the Go name for this bitfield type.
	Type string

	// Data lists the data for all known bits.
	Data []*AnnotatedBitfieldData

	// Names holds some valid example names for the bitfield bits, if any.
	Names []string

	// ByName maps valid names to the data for the corresponding bit.
	ByName map[string]*AnnotatedBitfieldData
}

// MakeBitfieldType initializes and returns a BitfieldType.
func MakeBitfieldType(typeName string, in []BitfieldData) BitfieldType {
	length := uint(len(in))
	if length > 64 {
		length = 64
	}

	out := BitfieldType{
		Type:   typeName,
		Data:   make([]*AnnotatedBitfieldData, 64),
		Names:  make([]string, 0, length),
		ByName: make(map[string]*AnnotatedBitfieldData, 4*length),
	}

	for index := uint(0); index < 64; index++ {
		var data BitfieldData
		if index < length {
			data = in[index]
		}

		ptr := &AnnotatedBitfieldData{
			BitfieldData: data,
			Index:        index,
			Bit:          (1 << index),
		}

		out.Data[index] = ptr
		if data.GoName == "" && data.Name == "" {
			continue
		}

		name := data.Name
		if name == "" {
			name = data.GoName
		}
		out.Names = append(out.Names, name)

		if data.GoName != "" {
			out.ByName[data.GoName] = ptr
			out.ByName[strings.ToLower(data.GoName)] = ptr
		}

		if data.Name != "" {
			out.ByName[data.Name] = ptr
			out.ByName[strings.ToLower(data.Name)] = ptr
		}

		for _, alias := range data.Aliases {
			out.ByName[alias] = ptr
			out.ByName[strings.ToLower(alias)] = ptr
		}
	}
	return out
}

// Get returns bitfield.Data[index] or panics with InvalidBitfieldIndexError.
func (bitfield BitfieldType) Get(index uint) AnnotatedBitfieldData {
	if index >= 64 {
		panic(InvalidBitfieldIndexError{
			Type:  bitfield.Type,
			Index: index,
			Limit: 64,
		})
	}
	return *bitfield.Data[index]
}

// ForEach iterates over bitfield.Data with the given callback function.
func (bitfield BitfieldType) ForEach(fn func(data AnnotatedBitfieldData)) {
	for index := uint(0); index < 64; index++ {
		fn(*bitfield.Data[index])
	}
}

func (bitfield BitfieldType) toStringImpl(
	value uint64,
	fn1 func(data AnnotatedBitfieldData) string,
	fn2 func(remnant uint64) string,
) string {
	pieces := make([]string, 0, 64)
	remnant := uint64(0)
	bitfield.ForEach(func(data AnnotatedBitfieldData) {
		if (value & data.Bit) != 0 {
			str := fn1(data)
			if str == "" {
				remnant |= data.Bit
			} else {
				pieces = append(pieces, str)
			}
		}
	})
	if remnant != 0 || len(pieces) == 0 {
		if str := fn2(remnant); str != "" {
			pieces = append(pieces, str)
		}
	}
	return strings.Join(pieces, "|")
}

// ToGoString generates a Go string representation for the given bitfield value.
func (bitfield BitfieldType) ToGoString(value uint64) string {
	return bitfield.toStringImpl(
		value,
		func(data AnnotatedBitfieldData) string {
			return data.GoName
		},
		func(remnant uint64) string {
			if remnant == 0 {
				return bitfield.Type + "(0)"
			}
			return bitfield.Type + "(0x" + strconv.FormatUint(remnant, 16) + ")"
		},
	)
}

// ToString generates a string representation for the given bitfield value.
func (bitfield BitfieldType) ToString(value uint64) string {
	return bitfield.toStringImpl(
		value,
		func(data AnnotatedBitfieldData) string {
			return data.Name
		},
		func(remnant uint64) string {
			if remnant == 0 {
				return "0"
			}
			return "0x" + strconv.FormatUint(remnant, 16)
		},
	)
}

// ToJSON marshals this bitfield value to JSON.
func (bitfield BitfieldType) ToJSON(value uint64) ([]byte, error) {
	return json.Marshal(bitfield.ToString(value))
}

func (bitfield BitfieldType) parseItem(str string) (uint64, bool) {
	strPrefix := bitfield.Type + "("
	strSuffix := ")"
	if strings.HasPrefix(str, strPrefix) && strings.HasSuffix(str, strSuffix) {
		i := uint(len(strSuffix))
		j := uint(len(str)) - uint(len(strSuffix))
		str = str[i:j]
	}

	if data, found := bitfield.ByName[str]; found {
		return data.Bit, true
	}

	strLower := strings.ToLower(str)
	if data, found := bitfield.ByName[strLower]; found {
		return data.Bit, true
	}

	if str == "0" {
		return 0, true
	}

	if u64, err := strconv.ParseUint(str, 0, 64); err == nil {
		return u64, true
	}

	return 0, false
}

// FromString parses the string representation of a bitfield value.  Returns
// InvalidBitfieldNameError if the string cannot be parsed.
func (bitfield BitfieldType) FromString(str string) (uint64, error) {
	if u64, ok := bitfield.parseItem(str); ok {
		return u64, nil
	}

	accum := uint64(0)
	pieces := strings.Split(str, "|")
	errors := []error(nil)
	for _, piece := range pieces {
		if u64, ok := bitfield.parseItem(piece); ok {
			accum |= u64
		} else {
			errors = append(errors, InvalidBitfieldNameError{
				Type:    bitfield.Type,
				Name:    piece,
				Allowed: bitfield.Names,
			})
		}
	}

	if len(errors) == 0 {
		return accum, nil
	}

	if len(errors) == 1 {
		return 0, errors[0]
	}

	return 0, &multierror.Error{Errors: errors}
}

// FromJSON unmarshals a bitfield value from JSON.  Returns IsNullError or
// InvalidBitfieldNameError if a JSON value was parsed but could not be
// unmarshaled as an bitfield value.
func (bitfield BitfieldType) FromJSON(raw []byte) (uint64, error) {
	if raw == nil {
		panic(errors.New("[]byte is nil"))
	}

	if bytes.Equal(raw, nullBytes) {
		return 0, IsNullError{}
	}

	var str string
	err0 := json.Unmarshal(raw, &str)
	if err0 == nil {
		return bitfield.FromString(str)
	}

	var u64 uint64
	err1 := json.Unmarshal(raw, &u64)
	if err1 == nil {
		return u64, nil
	}

	return 0, err0
}
