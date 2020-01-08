package config

import (
	"strconv"
	"testing"
	"bytes"
)

type testCase struct {
	data     interface{}
	error    bool
	errorMsg string
}

type s1 struct {
	Prm U16Param
}

type s2 struct {
	Prm U64Param
}

type s3 struct {
	Prm BoolParam
}

type s4 struct {
	Prm SelectParam
}

type s5 struct {
	Prm IPV4Param
}

type s6 struct {
	Prm1 SelectParam
	Prm2 IPV4Param
	Prm3 BoolParam
	Prm4 U16Param
}

type s7 struct {
	Prm1 SelectParam
	prm2 IPV4Param
}

type s8 struct {
	Prm1 SelectParam
	Prm2 s1
}

type s9 struct{}

type s10 struct {
	Prm1 SelectParam
	Prm2 int
}

var ptr *s1 = &s1{Prm: MakeU16(5, [2]uint16{0, 10}, Display{})}

var tests []testCase = []testCase{
	testCase{1, true, "Config is not a struct"},
	testCase{"abc", true, "Config is not a struct"},
	testCase{[]byte{1, 2, 3}, true, "Config is not a struct"},
	testCase{&ptr, true, "Config is not a struct"},

	// Ensure it works on pointers
	testCase{&s1{Prm: MakeU16(5, [2]uint16{0, 10}, Display{})}, false, ""},

	testCase{s1{Prm: MakeU16(5, [2]uint16{0, 10}, Display{})}, false, ""},
	testCase{s1{Prm: MakeU16(5, [2]uint16{7, 10}, Display{})}, true, "Prm : U16 value out of range"},

	testCase{s2{Prm: MakeU64(5, [2]uint64{0, 10}, Display{})}, false, ""},
	testCase{s2{Prm: MakeU64(5, [2]uint64{7, 10}, Display{})}, true, "Prm : U64 value out of range"},

	testCase{s3{Prm: MakeBool(true, Display{})}, false, ""},
	testCase{s3{Prm: MakeBool(false, Display{})}, false, ""},

	testCase{s4{Prm: MakeSelect("yes", []string{"yes", "no"}, Display{})}, false, ""},
	testCase{s4{Prm: MakeSelect("yes", []string{}, Display{})}, true, "Prm : Select value not in list"},
	testCase{s4{Prm: MakeSelect("not", []string{"yes", "no"}, Display{})}, true, "Prm : Select value not in list"},

	testCase{s5{Prm: MakeIPV4("1.2.3.4", Display{})}, false, ""},
	testCase{s5{Prm: MakeIPV4("1.2.3.4.5", Display{})}, true, "Prm : Invalid IPV4 address"},

	testCase{s6{Prm1: MakeSelect("yes", []string{"yes", "no"}, Display{}),
		Prm2: MakeIPV4("1.2.3.4", Display{}),
		Prm3: MakeBool(false, Display{}),
		Prm4: MakeU16(5, [2]uint16{0, 10}, Display{})}, false, ""},
	testCase{s6{Prm1: MakeSelect("yes", []string{"yes", "no"}, Display{}),
		Prm2: MakeIPV4("1:2:3:4:5:6", Display{}),
		Prm3: MakeBool(false, Display{}),
		Prm4: MakeU16(5, [2]uint16{0, 10}, Display{})}, true, "Prm2 : Invalid IPV4 address"},

	testCase{s7{Prm1: MakeSelect("yes", []string{"yes", "no"}, Display{}),
		prm2: MakeIPV4("1:2:3:4:5:6", Display{})}, true, "prm2 : Could not retrieve unexported field"},

	testCase{s8{Prm1: MakeSelect("yes", []string{"yes", "no"}, Display{})}, true, "Prm2 : Invalid struct field type"},

	testCase{s9{}, false, ""},

	testCase{s10{Prm1: MakeSelect("yes", []string{"yes", "no"}, Display{})}, true, "Prm2 : Invalid struct field type"},
}

func TestValidate(t *testing.T) {
	for i, v := range tests {
		if err := Validate(v.data); v.error && err == nil {
			t.Errorf("Case %d : Expected error %s", i, v.errorMsg)
		} else if v.error && err != nil && v.errorMsg != err.Error() {
			t.Errorf("Case %d : Expected error %s: Found %s", i, v.errorMsg, err.Error())
		} else if !v.error && err != nil {
			t.Errorf("Case %d : Expected no error: Found %s", i, err.Error())
		}
	}
}

type copyTestCase struct {
	c1       interface{}
	c2       interface{}
	error    bool
	errorMsg string
}

type ValStruct struct {
	Value int
}

type ValStructInter struct {
	Value interface{}
}

type NoValStruct struct {
	NotValue int
}

func TestCopyValueErrors(t *testing.T) {

	type st1 struct {
		p1 int
	}

	type st2 struct {
		p1 int
	}

	type st3 struct {
		p1 NoValStruct
	}

	type st4 struct {
		p1 ValStruct
	}

	type st5 struct {
		P1 ValStruct
	}

	type st6 struct {
		P1 ValStructInter
	}

	type st7 struct{}

	var (
		intVal int
		sVal1  st1
		sVal3  st3
		sVal4  st4
		sVal5  st5
		sVal6  st6 = st6{P1: ValStructInter{123}}
		sVal7  st7
	)

	var copyTests []copyTestCase = []copyTestCase{
		copyTestCase{1, 2, true, "Initial config must be pointer"},
		copyTestCase{&intVal, "abc", true, "Configs must be same type"},
		copyTestCase{&sVal1, st2{}, true, "Configs must be same type"},
		copyTestCase{&intVal, intVal, true, "Configs must be struct"},
		copyTestCase{&sVal1, st1{}, true, "p1 : must be struct"},
		copyTestCase{&sVal3, st3{}, true, "p1 : struct must contain Value field"},
		copyTestCase{&sVal4, st4{}, true, "p1 : struct Value field must be settable"},
		copyTestCase{&sVal5, st5{}, false, ""},
		// Ensure that types stored by interfaces can be swapped
		copyTestCase{&sVal6, st6{P1: ValStructInter{"abc"}}, false, ""},
		copyTestCase{&sVal7, st7{}, false, ""},
	}

	for i, v := range copyTests {
		if err := CopyValue(v.c1, v.c2); v.error && err == nil {
			t.Errorf("Case %d : Expected error %s", i, v.errorMsg)
		} else if v.error && err != nil && v.errorMsg != err.Error() {
			t.Errorf("Case %d : Expected error %s: Found %s", i, v.errorMsg, err.Error())
		} else if !v.error && err != nil {
			t.Errorf("Case %d : Expected no error: Found %s", i, err.Error())
		}
	}
}

func TestCopyValueU16(t *testing.T) {
	var sVal1 s1 = s1{Prm: MakeU16(5, [2]uint16{0, 10}, Display{})}
	var sVal2 s1 = s1{Prm: MakeU16(6, [2]uint16{0, 10}, Display{})}

	if sVal1.Prm.Value == sVal2.Prm.Value {
		t.Errorf("Expected values to not match : Found %d", sVal2.Prm.Value)
	}
	err := CopyValue(&sVal1, sVal2)
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm.Value != sVal2.Prm.Value {
		t.Errorf("Expected %d: Found %d", sVal1.Prm.Value, sVal2.Prm.Value)
	}
}

func TestCopyValueU16Ptr(t *testing.T) {
	var sVal1 s1 = s1{Prm: MakeU16(5, [2]uint16{0, 10}, Display{})}
	var sVal2 s1 = s1{Prm: MakeU16(6, [2]uint16{0, 10}, Display{})}

	if sVal1.Prm.Value == sVal2.Prm.Value {
		t.Errorf("Expected values to not match : Found %d", sVal2.Prm.Value)
	}
	err := CopyValue(&sVal1, &sVal2)
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm.Value != sVal2.Prm.Value {
		t.Errorf("Expected %d: Found %d", sVal1.Prm.Value, sVal2.Prm.Value)
	}
}

func TestCopyValueSameValue(t *testing.T) {
	var sVal1 s1 = s1{Prm: MakeU16(5, [2]uint16{0, 10}, Display{})}
	var sVal2 s1 = s1{Prm: MakeU16(5, [2]uint16{0, 10}, Display{})}

	err := CopyValue(&sVal1, sVal2)
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm.Value != sVal2.Prm.Value {
		t.Errorf("Expected %d: Found %d", sVal1.Prm.Value, sVal2.Prm.Value)
	}
}

func TestCopyValueU64(t *testing.T) {
	var sVal1 s2 = s2{Prm: MakeU64(0, [2]uint64{0, 10}, Display{})}
	var sVal2 s2 = s2{Prm: MakeU64(10, [2]uint64{0, 10}, Display{})}

	if sVal1.Prm.Value == sVal2.Prm.Value {
		t.Errorf("Expected values to not match : Found %d", sVal2.Prm.Value)
	}
	err := CopyValue(&sVal1, sVal2)
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm.Value != sVal2.Prm.Value {
		t.Errorf("Expected %d: Found %d", sVal1.Prm.Value, sVal2.Prm.Value)
	}
}

func TestCopyValueBool(t *testing.T) {
	var sVal1 s3 = s3{Prm: MakeBool(true, Display{})}
	var sVal2 s3 = s3{Prm: MakeBool(false, Display{})}

	if sVal1.Prm.Value == sVal2.Prm.Value {
		t.Errorf("Expected values to not match : Found %s", strconv.FormatBool(sVal2.Prm.Value))
	}
	err := CopyValue(&sVal1, sVal2)
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm.Value != sVal2.Prm.Value {
		t.Errorf("Expected %s: Found %s", strconv.FormatBool(sVal1.Prm.Value), strconv.FormatBool(sVal2.Prm.Value))
	}
}

func TestCopyValueSelect(t *testing.T) {
	var sVal1 s4 = s4{Prm: MakeSelect("yes", []string{"yes", "no"}, Display{})}
	var sVal2 s4 = s4{Prm: MakeSelect("no", []string{"yes", "no"}, Display{})}

	if sVal1.Prm.Value == sVal2.Prm.Value {
		t.Errorf("Expected values to not match : Found %s", sVal2.Prm.Value)
	}
	err := CopyValue(&sVal1, sVal2)
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm.Value != sVal2.Prm.Value {
		t.Errorf("Expected %s: Found %s", sVal1.Prm.Value, sVal2.Prm.Value)
	}
}

func TestCopyValueIPV4(t *testing.T) {
	var sVal1 s5 = s5{Prm: MakeIPV4("1.2.3.4", Display{})}
	var sVal2 s5 = s5{Prm: MakeIPV4("4.3.2.1", Display{})}

	if sVal1.Prm.Value == sVal2.Prm.Value {
		t.Errorf("Expected values to not match : Found %s", sVal2.Prm.Value)
	}
	err := CopyValue(&sVal1, sVal2)
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm.Value != sVal2.Prm.Value {
		t.Errorf("Expected %s: Found %s", sVal1.Prm.Value, sVal2.Prm.Value)
	}
}

func TestCopyValueMultiValue(t *testing.T) {
	var sVal1 s6 = s6{Prm1: MakeSelect("yes", []string{"yes", "no"}, Display{}),
		Prm2: MakeIPV4("1.2.3.4", Display{}),
		Prm3: MakeBool(true, Display{}),
		Prm4: MakeU16(5, [2]uint16{0, 10}, Display{})}
	var sVal2 s6 = s6{Prm1: MakeSelect("no", []string{"yes", "no"}, Display{}),
		Prm2: MakeIPV4("4.3.2.1", Display{}),
		Prm3: MakeBool(false, Display{}),
		Prm4: MakeU16(6, [2]uint16{0, 10}, Display{})}

	err := CopyValue(&sVal1, sVal2)
	// We test the values explicitely here
	// to double check that CopyValue is not altering sVal2
	if err != nil {
		t.Errorf("Expected no error: Found %s", err.Error())
	} else if sVal1.Prm1.Value != "no" {
		t.Errorf("Expected %s: Found %s", sVal1.Prm1.Value, "no")
	} else if sVal1.Prm2.Value != "4.3.2.1" {
		t.Errorf("Expected %s: Found %s", sVal1.Prm2.Value, "4.3.2.1")
	} else if sVal1.Prm3.Value != false {
		t.Errorf("Expected %s: Found %s", strconv.FormatBool(sVal1.Prm3.Value), strconv.FormatBool(false))
	} else if sVal1.Prm4.Value != 6 {
		t.Errorf("Expected %d: Found %d", sVal1.Prm4.Value, 6)
	}
}

func TestNoUpdateUnlessAllValid(t *testing.T) {
	type ValStruct struct {
		Value int
	}
	type NoValStruct struct {
		NotValue int
	}
	type s struct {
		P1 ValStruct
		P2 NoValStruct
		P3 ValStruct
	}
	var s1 s = s{P1: ValStruct{1}, P2: NoValStruct{2}, P3: ValStruct{3}}
	var s2 s = s{P1: ValStruct{4}, P2: NoValStruct{5}, P3: ValStruct{6}}

	err := CopyValue(&s1, s2)
	if err == nil {
		t.Errorf("Expected error")
	} else if err.Error() != "P2 : struct must contain Value field" {
		t.Errorf("Expected error %s: Found %s", "P2 : struct must contain Value field", err.Error())
	} else if s1.P1.Value != 1 {
		t.Errorf("Expected %d: Found %d", s1.P1.Value, 1)
	} else if s1.P2.NotValue != 2 {
		t.Errorf("Expected %d: Found %d", s1.P2.NotValue, 2)
	} else if s1.P3.Value != 3 {
		t.Errorf("Expected %d: Found %d", s1.P3.Value, 3)
	}
}

type copySetTestCase struct {
	c1       interface{}
	c2       interface{}
	fields   []string
	error    bool
	errorMsg string
}

func TestCopySet(t *testing.T) {

	type conf1 struct {
		P1 ValStruct
	}

	type conf2 struct {
		P2 ValStruct
	}

	type noConf struct {
		P3 NoValStruct
	}

	type st1 struct{}
	type st2 struct {
		C1 int
	}
	type st3 struct {
		C1 conf1
	}
	type st4 struct {
		C1 conf1
		C2 conf2
	}
	type st5 struct {
		C1 conf1
		C2 conf2
		C3 noConf
	}

	var (
		intVal int
		sVal1  st1
		sVal2  st2
		sVal3  st3
		sVal4  st4
		sVal5  st5
	)

	var copyTests []copySetTestCase = []copySetTestCase{
		copySetTestCase{1, 2, nil, true, "Initial config must be pointer"},
		copySetTestCase{&intVal, 2, nil, true, "Configs must be struct"},
		copySetTestCase{sVal1, sVal1, nil, true, "Initial config must be pointer"},
		copySetTestCase{&sVal1, sVal1, nil, false, ""},
		copySetTestCase{&sVal1, sVal1, []string{"C1"}, true, "C1 : field not in struct"},
		copySetTestCase{&sVal2, sVal2, []string{}, false, ""},
		copySetTestCase{&sVal2, sVal2, nil, true, "C1 : Configs must be struct"},
		copySetTestCase{&sVal3, sVal3, nil, false, ""},
		copySetTestCase{&sVal3, sVal3, []string{"C1"}, false, ""},
		copySetTestCase{&sVal3, sVal3, []string{"C1", "C2"}, true, "C2 : field not in struct"},
		copySetTestCase{&sVal4, sVal4, nil, false, ""},
		copySetTestCase{&sVal4, sVal4, []string{"C2"}, false, ""},
		copySetTestCase{&sVal4, sVal4, []string{"C1", "C2"}, false, ""},
		copySetTestCase{&sVal3, sVal4, nil, true, "Configs must be same type"},
		copySetTestCase{&sVal5, sVal5, []string{}, false, ""},
		copySetTestCase{&sVal5, sVal5, []string{"C1", "C2"}, false, ""},
		copySetTestCase{&sVal5, sVal5, nil, true, "C3 : P3 : struct must contain Value field"},
	}

	for i, v := range copyTests {
		if err := CopyValueSet(v.c1, v.c2, v.fields); v.error && err == nil {
			t.Errorf("Case %d : Expected error %s", i, v.errorMsg)
		} else if v.error && err != nil && v.errorMsg != err.Error() {
			t.Errorf("Case %d : Expected error %s: Found %s", i, v.errorMsg, err.Error())
		} else if !v.error && err != nil {
			t.Errorf("Case %d : Expected no error: Found %s", i, err.Error())
		}
	}
}

func TestSetCopy(t *testing.T) {

	type conf1 struct {
		P1 ValStruct
	}
	type conf2 struct {
		P1 ValStruct
	}

	type st struct {
		C1 conf1
		C2 conf2
	}

	var s1 st = st{C1: conf1{P1: ValStruct{Value: 1}}, C2: conf2{P1: ValStruct{Value: 1}}}
	var s2 st = st{C1: conf1{P1: ValStruct{Value: 2}}, C2: conf2{P1: ValStruct{Value: 2}}}

	err := CopyValueSet(&s1, s2, []string{})
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	} else if s1.C1.P1.Value != 1 && s2.C1.P1.Value != 2 {
		t.Errorf("Expected %d: Found %d", 1, s1.C1.P1.Value)
	} else if s1.C2.P1.Value != 1 && s2.C2.P1.Value != 2 {
		t.Errorf("Expected %d: Found %d", 1, s1.C1.P1.Value)
	}

	err = CopyValueSet(&s1, s2, nil)
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	} else if s1.C1.P1.Value != s2.C1.P1.Value {
		t.Errorf("Expected %d: Found %d", s2.C1.P1.Value, s1.C1.P1.Value)
	} else if s1.C2.P1.Value != s2.C2.P1.Value {
		t.Errorf("Expected %d: Found %d", s2.C2.P1.Value, s1.C1.P1.Value)
	}

	s1 = st{C1: conf1{P1: ValStruct{Value: 1}}, C2: conf2{P1: ValStruct{Value: 1}}}
	s2 = st{C1: conf1{P1: ValStruct{Value: 2}}, C2: conf2{P1: ValStruct{Value: 2}}}

	err = CopyValueSet(&s1, s2, []string{"C1"})
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	} else if s1.C1.P1.Value != s2.C1.P1.Value {
		t.Errorf("Expected %d: Found %d", 1, s1.C1.P1.Value)
	} else if s1.C2.P1.Value != 1 && s2.C2.P1.Value != 2 {
		t.Errorf("Expected %d: Found %d", 1, s1.C1.P1.Value)
	}

	s1 = st{C1: conf1{P1: ValStruct{Value: 1}}, C2: conf2{P1: ValStruct{Value: 1}}}
	s2 = st{C1: conf1{P1: ValStruct{Value: 2}}, C2: conf2{P1: ValStruct{Value: 2}}}

	err = CopyValueSet(&s1, s2, []string{"C1", "C2"})
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
	} else if s1.C1.P1.Value != 1 && s2.C1.P1.Value != 2 {
		t.Errorf("Expected %d: Found %d", 1, s1.C1.P1.Value)
	} else if s1.C2.P1.Value != 1 && s2.C2.P1.Value != 2 {
		t.Errorf("Expected %d: Found %d", 1, s1.C1.P1.Value)
	}
}

func TestSetNoUpdateUnlessAllValid(t *testing.T) {

	type conf1 struct {
		P1 ValStruct
	}
	type noConf struct {
		P2 NoValStruct
	}
	type conf2 struct {
		P1 ValStruct
	}

	type st struct {
		C1 conf1
		C2 noConf
		C3 conf2
	}

	var s1 st = st{C1: conf1{P1: ValStruct{Value: 1}}}
	var s2 st = st{C1: conf1{P1: ValStruct{Value: 2}}}

	err := CopyValueSet(&s1, s2, []string{"C1", "C2"})
	if err == nil {
		t.Errorf("Expected error")
	} else if err.Error() != "C2 : P2 : struct must contain Value field" {
		t.Errorf("Expected error %s: Found %s", "C2 : P2 : struct must contain Value field", err.Error())
	} else if s1.C1.P1.Value != 1 {
		t.Errorf("Expected %d: Found %d", s1.C1.P1.Value, 1)
	} else if s1.C2.P2.NotValue != 0 {
		t.Errorf("Expected %d: Found %d", s1.C2.P2.NotValue, 0)
	} else if s2.C1.P1.Value != 2 {
		t.Errorf("Expected %d: Found %d", s2.C1.P1.Value, 3)
	}
}

func TestExactU64(t *testing.T) {

	eU64 := MakeExactU64(12345, Display{})

	if eU64.Value != "12345" {
		t.Errorf("Expected %s; found %s", "12345", eU64.Value)
	} else if eU64.Validate() != nil {
		t.Errorf("Expected successful validation")
	} else if n, err := eU64.GetValue(); err != nil {
		t.Errorf("Expected no error; found %s", err.Error())
	} else if n != 12345 {
		t.Errorf("Expected %d; found %d", 12345, n)
	}

	eU64.Value = "123451234512345123451234512345"

	if eU64.Validate() == nil {
		t.Errorf("Expected unsuccessful validation")
	} else if n, err := eU64.GetValue(); err == nil {
		t.Errorf("Expected error from GetValue")
	} else if err.Error() != "strconv.ParseUint: parsing \"123451234512345123451234512345\": value out of range" {
		t.Errorf("Expected error %s; found %s", "strconv.ParseUint: parsing \"123451234512345123451234512345\": value out of range", err.Error())
	} else if n != 0 {
		t.Errorf("Expected %d for invalid value; found %d", 0, n)
	}

}

func TestHexKey(t *testing.T) {

	hKey := MakeHexKey([]byte{1,2,3,4}, []int{2, 4}, Display{})

	if hKey.Validate() != nil {
		t.Errorf("Expected no error; found %s", hKey.Validate().Error())
	} else if !bytes.Equal(hKey.Value, []byte{1,2,3,4}) {
		t.Errorf("Expected found %v; found %v", []byte{1,2,3,4}, hKey.Value)
	}

	hKey.Value = []byte{1,2}

	if hKey.Validate() != nil {
		t.Errorf("Expected no error; found %s", hKey.Validate().Error())
	} else if !bytes.Equal(hKey.Value, []byte{1,2}) {
		t.Errorf("Expected found %v; found %v", []byte{1,2}, hKey.Value)
	}

	hKey.Value = []byte{1,2,3}
	if hKey.Validate() == nil {
		t.Errorf("Expected error")
	} else if hKey.Validate().Error() != "Invalid key length" {
		t.Errorf("Expected error %s; found %s", "Invalid key length", hKey.Validate().Error())
	}
}
