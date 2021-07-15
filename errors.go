package enumhelper

import (
	"errors"
	"fmt"
)

var nullBytes = []byte("null")

// IsNull returns true iff err is an instance of IsNullError.
func IsNull(err error) bool {
	var x IsNullError
	return errors.As(err, &x)
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

// type InvalidBitfieldNameError {{{

// InvalidBitfieldNameError indicates a bitfield bit whose string representation
// could not be recognized.
type InvalidBitfieldNameError struct {
	Type    string
	Name    string
	Allowed []string
}

// Error fulfills the error interface.
func (err InvalidBitfieldNameError) Error() string {
	if len(err.Allowed) == 0 {
		return fmt.Sprintf("invalid %s name %q", err.Type, err.Name)
	}
	return fmt.Sprintf("invalid %s name %q; must be one of %q", err.Type, err.Name, err.Allowed)
}

var _ error = InvalidBitfieldNameError{}

// }}}

// type InvalidBitfieldIndexError {{{

// InvalidBitfieldIndexError indicates an enum whose numeric value is out of range.
type InvalidBitfieldIndexError struct {
	Type  string
	Index uint
	Limit uint
}

// Error fulfills the error interface.
func (err InvalidBitfieldIndexError) Error() string {
	if err.Limit == 0 {
		return fmt.Sprintf("invalid %s value %d", err.Type, err.Index)
	}
	return fmt.Sprintf("invalid %s value %d; must be < %d", err.Type, err.Index, err.Limit)
}

var _ error = InvalidBitfieldIndexError{}

// }}}
