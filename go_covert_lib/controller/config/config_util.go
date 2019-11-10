package config

import (
	"errors"
	"net"
	"reflect"
)

type param interface {
	Validate() error
}

type U16Param struct {
	Type        string
	Value       uint16
	Range       [2]uint16
	Description string
}

type U64Param struct {
	Type        string
	Value       uint64
	Range       [2]uint64
	Description string
}

type BoolParam struct {
	Type        string
	Value       bool
	Description string
}

type SelectParam struct {
	Type        string
	Value       string
	Range       []string
	Description string
}

type IPV4Param struct {
	Type string
	// To support the range of IP addresses, this is a string
	// To convert to the proper IP address that can be used later on use GetValue
	Value       string
	Description string
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

func (p *IPV4Param) GetValue() ([4]byte, error) {
	var buf [4]byte
	if ip := net.ParseIP(p.Value); ip != nil {
		if ip4 := ip.To4(); ip4 != nil && len(ip4) == 4 {
			copy(buf[:], ip4[:4])
			return buf, nil
		}
	}
	return buf, errors.New("Invalid IPV4 address")
}

func MakeIPV4(value string, description string) IPV4Param {
	return IPV4Param{"ipv4", value, description}
}
func MakeU16(value uint16, rng [2]uint16, description string) U16Param {
	return U16Param{"u16", value, rng, description}
}
func MakeU64(value uint64, rng [2]uint64, description string) U64Param {
	return U64Param{"u64", value, rng, description}
}
func MakeSelect(value string, rng []string, description string) SelectParam {
	return SelectParam{"select", value, rng, description}
}
func MakeBool(value bool, description string) BoolParam {
	return BoolParam{"bool", value, description}
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

	if v1.Type() != v2.Type() {
		return errors.New("Configs must be same type")
	}

	t := v1.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("Configs must be struct")
	}
	// We run the loop twice, the first time to validate the structure
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
	// The second time is to update the fields
	// This way no updates happen unless all updates are valid
	for i := 0; i < t.NumField(); i++ {
		f1 := v1.Field(i)
		f2 := v2.Field(i)
		f1.FieldByName("Value").Set(f2.FieldByName("Value"))
	}
	return nil
}
