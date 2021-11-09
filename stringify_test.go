package gcfg

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type subtypeStruct1 struct {
	Bar     string
	Baz     subtypeStructWithMarshaler
	FBar    bool `gcfg:"f--bar"`
	BFoo    bool
	Bazbar  uint
	FooBaz  int
	Bar_Foo *big.Int
	Baz_Baz []string
	F_Bar   map[string]string `gcfg:"extra_values"`
	// Private field is here to make sure they aren't picked up.
	private bool
}

type subtypeStructWithMarshaler struct {
	Value string
}

func (s subtypeStructWithMarshaler) MarshalText() ([]byte, error) {
	return []byte(s.Value), nil
}

type subtypeStruct2 struct {
	Foobar subtypeStructNoMarshaler
}

type subtypeStructNoMarshaler struct {
	Value string
}

type extraValuesStruct struct {
	ExtraValues map[string][]string `gcfg:"extra_values"`
}

func TestStringify1(t *testing.T) {
	config := &struct {
		Foo subtypeStruct1
	}{
		Foo: subtypeStruct1{
			Bar:     "value1",
			Baz:     subtypeStructWithMarshaler{Value: "value2"},
			FBar:    true,
			FooBaz:  -5,
			Bar_Foo: big.NewInt(10),
			Baz_Baz: []string{"value3", "value4"},
			F_Bar:   map[string]string{"bar-key": "bar-value"},
		},
	}
	expectedResult := `[foo]
bar=value1
baz=value2
f--bar=true
bfoo=false
bazbar=0
foobaz=-5
bar-foo=10
baz-baz=value3
baz-baz=value4
bar-key=bar-value

`

	res, err := Stringify(config)
	assert.NoError(t, err)
	assert.Equal(t, res, expectedResult)
}

func TestStringify2(t *testing.T) {
	config := &struct {
		Foo subtypeStruct2
	}{}

	_, err := Stringify(config)
	assert.Errorf(t, err, "Third-level down structs must fail if they don't implement encoding.TextMarshaler: %v", config)
}

func TestStringify3(t *testing.T) {
	config := &struct {
		Foo map[string]*subtypeStruct1
	}{
		Foo: map[string]*subtypeStruct1{
			"sub": {
				Bar:     "value1",
				Baz:     subtypeStructWithMarshaler{Value: "value2"},
				FBar:    true,
				FooBaz:  -5,
				Bar_Foo: big.NewInt(10),
				Baz_Baz: []string{"value3", "value4"},
				F_Bar:   map[string]string{"bar-key": "bar-value"},
			},
		},
	}
	expectedResult := `[foo "sub"]
bar=value1
baz=value2
f--bar=true
bfoo=false
bazbar=0
foobaz=-5
bar-foo=10
baz-baz=value3
baz-baz=value4
bar-key=bar-value

`

	res, err := Stringify(config)
	assert.NoError(t, err)
	assert.Equal(t, res, expectedResult)
}

func TestStringify4(t *testing.T) {
	config := &struct {
		Foo map[string]string
	}{
		Foo: map[string]string{"key": "value"},
	}
	expectedResult := `[foo]
key=value

`

	res, err := Stringify(config)
	assert.NoError(t, err)
	assert.Equal(t, res, expectedResult)
}

func TestStringify5(t *testing.T) {
	config := &struct {
		Foo map[string]string
	}{
		Foo: map[string]string{"sub key": "value"},
	}
	expectedResult := `[foo "sub"]
key=value

`

	res, err := Stringify(config)
	assert.NoError(t, err)
	assert.Equal(t, res, expectedResult)
}

func TestStringify6(t *testing.T) {
	config := &struct {
		Bar subtypeStructNoMarshaler
		Baz extraValuesStruct
	}{
		Bar: subtypeStructNoMarshaler{Value: "value1"},
		Baz: extraValuesStruct{
			ExtraValues: map[string][]string{"key1": {"value2", "value3"}},
		},
	}
	expectedResult := `[bar]
value=value1

[baz]
key1=value2
key1=value3

`

	res, err := Stringify(config)
	assert.NoError(t, err)
	assert.Equal(t, res, expectedResult)
}
