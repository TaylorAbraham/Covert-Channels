package config

import (
	"errors"
	"net"
	"reflect"
	"strconv"
)

type param interface {
	Validate() error
}

type Display struct {
	Description string
	Name        string
	Group       string
	// If GroupToggle is set to true on a bool Param, the truthiness of the bool Param
	// will determine whether the rest of the Group is shown or hidden
	GroupToggle bool
}

type I8Param struct {
	Type    string
	Value   int8
	Range   [2]int8
	Display Display
}

type U16Param struct {
	Type    string
	Value   uint16
	Range   [2]uint16
	Display Display
}

type U64Param struct {
	Type    string
	Value   uint64
	Range   [2]uint64
	Display Display
}

// In Javascript the number type is a floating point integer
// This means that large numbers cannot be represented exactly as integers
// The inprecision of floating point numbers is fine for small numbers,
// for large 64 bit integers not all numbers can be represented.
// The largest safe 64 bit integer is 9007199254740991 (Number.MAX_SAFE_INTEGER, (2^53)-1)
// (see https://www.wikitechy.com/tutorials/javascript/what-is-javascripts-highest-integer-value-that-a-number-can-go-to-without-losing-precision)
// If a large 64 bit integer is input in a number input it will be rounded to the nearest
// whole floating point number when sent to the server.
// This type is used to transport the 64 bit integer as a string in the configuration.
// This allows the exact number to be sent to the server, preventing the number from being changed.
// This param type should be used whenever an exact, large 64 bit number is required (for example,
// encryption keys).
type ExactU64Param struct {
	Type    string
	Value   string
	Display Display
}

type BoolParam struct {
	Type    string
	Value   bool
	Display Display
}

type SelectParam struct {
	Type    string
	Value   string
	Range   []string
	Display Display
}

type IPV4Param struct {
	Type string
	// To support the range of IP addresses, this is a string
	// To convert to the proper IP address that can be used later on use GetValue
	Value   string
	Display Display
}

type HexKeyParam struct {
	Type    string
	Value   []byte
	Range   []int
	Display Display
}

func (p I8Param) Validate() error {
	if p.Value >= p.Range[0] && p.Value <= p.Range[1] {
		return nil
	} else {
		return errors.New("I8 value out of range")
	}
}

func (p U16Param) Validate() error {
	if p.Value >= p.Range[0] && p.Value <= p.Range[1] {
		return nil
	} else {
		return errors.New("U16 value out of range")
	}
}

func (p U64Param) Validate() error {
	if p.Value >= p.Range[0] && p.Value <= p.Range[1] {
		return nil
	} else {
		return errors.New("U64 value out of range")
	}
}

func (p ExactU64Param) Validate() error {
	_, err := p.GetValue()
	return err
}

func (p ExactU64Param) GetValue() (uint64, error) {
	if n, err := strconv.ParseUint(p.Value, 10, 64); err == nil {
		return n, nil
	} else {
		return 0, err
	}
}

func (p BoolParam) Validate() error {
	return nil
}

func (p SelectParam) Validate() error {
	for _, s := range p.Range {
		if s == p.Value {
			return nil
		}
	}
	return errors.New("Select value not in list")
}

func (p IPV4Param) Validate() error {
	_, err := p.GetValue()
	return err
}

func (p IPV4Param) GetValue() ([4]byte, error) {
	var buf [4]byte
	if ip := net.ParseIP(p.Value); ip != nil {
		if ip4 := ip.To4(); ip4 != nil && len(ip4) == 4 {
			copy(buf[:], ip4[:4])
			return buf, nil
		}
	}
	return buf, errors.New("Invalid IPV4 address")
}

func (p HexKeyParam) Validate() error {
	for _, l := range p.Range {
		if len(p.Value) == l {
			return nil
		}
	}
	return errors.New("Invalid key length")
}

func MakeI8(value int8, rng [2]int8, display Display) I8Param {
	return I8Param{"i8", value, rng, display}
}
func MakeU16(value uint16, rng [2]uint16, display Display) U16Param {
	return U16Param{"u16", value, rng, display}
}
func MakeU64(value uint64, rng [2]uint64, display Display) U64Param {
	return U64Param{"u64", value, rng, display}
}
func MakeExactU64(value uint64, display Display) ExactU64Param {
	return ExactU64Param{"exactu64", strconv.FormatUint(value, 10), display}
}
func MakeSelect(value string, rng []string, display Display) SelectParam {
	return SelectParam{"select", value, rng, display}
}
func MakeBool(value bool, display Display) BoolParam {
	return BoolParam{"bool", value, display}
}
func MakeIPV4(value string, display Display) IPV4Param {
	return IPV4Param{"ipv4", value, display}
}

func MakeHexKey(value []byte, rng []int, display Display) HexKeyParam {
	return HexKeyParam{"hexkey", value, rng, display}
}

func Validate(c interface{}) error {
	v := reflect.ValueOf(c)
	// We support pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("Config is not a struct")
	}
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if v.Field(i).CanInterface() {
			if p, ok := v.Field(i).Interface().(param); ok {
				err := p.Validate()
				if err != nil {
					return errors.New(fieldName + " : " + err.Error())
				}
			} else {
				return errors.New(fieldName + " : Invalid struct field type")
			}
		} else {
			return errors.New(fieldName + " : Could not retrieve unexported field")
		}
	}
	return nil
}

// This function analysis a struct containing several config structs
// to ensure that they are all valid
func ValidateConfigSet(c interface{}) error {
	v := reflect.ValueOf(c)
	// We support pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("Config is not a struct")
	}
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if v.Field(i).CanInterface() {
			if err := Validate(v.Field(i).Interface()); err != nil {
				return err
			}
		} else {
			return errors.New(fieldName + " : Could not retrieve unexported field")
		}
	}
	return nil
}

// Copy Every param value from c2 to c1
func CopyValue(c1 interface{}, c2 interface{}) error {
	p1 := reflect.ValueOf(c1)
	v2 := reflect.ValueOf(c2)

	if p1.Kind() != reflect.Ptr {
		return errors.New("Initial config must be pointer")
	}

	v1 := p1.Elem()
	// We support pointers as the second interface for convenience
	if v2.Kind() == reflect.Ptr {
		v2 = v2.Elem()
	}
	if err := validateCopy(v1, v2); err != nil {
		return err
	}
	performCopy(v1, v2)
	return nil
}

func validateCopy(v1 reflect.Value, v2 reflect.Value) error {
	if v1.Type() != v2.Type() {
		return errors.New("Configs must be same type")
	}
	t := v1.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("Configs must be struct")
	}
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		f1 := v1.Field(i)
		f2 := v2.Field(i)
		if t.Field(i).Type.Kind() != reflect.Struct {
			return errors.New(fieldName + " : must be struct")
		}
		if _, ok := t.Field(i).Type.FieldByName("Value"); !ok {
			return errors.New(fieldName + " : struct must contain Value field")
		}
		if !f1.FieldByName("Value").CanSet() {
			return errors.New(fieldName + " : struct Value field must be settable")
		}
		if f1.FieldByName("Value").Type() != f2.FieldByName("Value").Type() {
			return errors.New(fieldName + " : struct Value field must contain compatible types")
		}
	}
	return nil
}

func performCopy(v1 reflect.Value, v2 reflect.Value) {
	t := v1.Type()
	for i := 0; i < t.NumField(); i++ {
		f1 := v1.Field(i)
		f2 := v2.Field(i)
		f1.FieldByName("Value").Set(f2.FieldByName("Value"))
	}
}

// Copy the set of config param values betweem the config structs
func CopyValueSet(c1 interface{}, c2 interface{}, fields []string) error {
	p1 := reflect.ValueOf(c1)
	v2 := reflect.ValueOf(c2)

	if p1.Kind() != reflect.Ptr {
		return errors.New("Initial config must be pointer")
	}

	v1 := p1.Elem()
	// We support pointers as the second interface for convenience
	if v2.Kind() == reflect.Ptr {
		v2 = v2.Elem()
	}

	if v1.Type() != v2.Type() {
		return errors.New("Configs must be same type")
	}

	t := v1.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("Configs must be struct")
	}
	// If nil is supplied we copy all fields
	if fields == nil {
		for i := 0; i < t.NumField(); i++ {
			fields = append(fields, t.Field(i).Name)
		}
	}
	for _, fname := range fields {
		f1 := v1.FieldByName(fname)
		f2 := v2.FieldByName(fname)
		if f1.IsValid() || f2.IsValid() {
			if err := validateCopy(f1, f2); err != nil {
				return errors.New(fname + " : " + err.Error())
			}
		} else {
			return errors.New(fname + " : field not in struct")
		}
	}
	for _, fname := range fields {
		f1 := v1.FieldByName(fname)
		f2 := v2.FieldByName(fname)
		performCopy(f1, f2)
	}
	return nil
}
