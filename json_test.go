package gcfg

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawJson1(t *testing.T) {
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
	expectedResult := `{
    "foo": {
        "bar": "value1",
        "baz": "value2",
        "f--bar": true,
        "bfoo": false,
        "bazbar": 0,
        "foobaz": -5,
        "bar-foo": 10,
        "baz-baz": [
            "value3",
            "value4"
        ],
        "bar-key": "bar-value",
        "ǂbar": "",
        "xfoo": ""
    }
}`

	v, err := RawJSON(config)
	assert.NoError(t, err)

	var out bytes.Buffer
	err = json.Indent(&out, v, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, out.String())
}

func TestRawJson2(t *testing.T) {
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
	expectedResult := `{
    "foo": {
        "sub": {
            "bar": "value1",
            "baz": "value2",
            "f--bar": true,
            "bfoo": false,
            "bazbar": 0,
            "foobaz": -5,
            "bar-foo": 10,
            "baz-baz": [
                "value3",
                "value4"
            ],
            "bar-key": "bar-value",
            "ǂbar": "",
            "xfoo": ""
        }
    }
}`

	v, err := RawJSON(config)
	assert.NoError(t, err)

	var out bytes.Buffer
	err = json.Indent(&out, v, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, out.String())
}

func TestRawJson3(t *testing.T) {
	config := &struct {
		Foo map[string]string
	}{
		Foo: map[string]string{"key": "value"},
	}
	expectedResult := `{
    "foo": {
        "key": "value"
    }
}`

	v, err := RawJSON(config)
	assert.NoError(t, err)

	var out bytes.Buffer
	err = json.Indent(&out, v, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, out.String())
}

func TestRawJson4(t *testing.T) {
	config := &struct {
		Foo map[string]string
	}{
		Foo: map[string]string{"sub key": "value"},
	}
	expectedResult := `{
    "foo": {
        "sub key": "value"
    }
}`

	v, err := RawJSON(config)
	assert.NoError(t, err)

	var out bytes.Buffer
	err = json.Indent(&out, v, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, out.String())
}

func TestRawJson5(t *testing.T) {
	config := &struct {
		Bar subtypeStructNoMarshaler
		Baz extraValuesStruct
	}{
		Bar: subtypeStructNoMarshaler{Value: "value1"},
		Baz: extraValuesStruct{
			ExtraValues: map[string][]string{"key1": {"value2", "value3"}},
		},
	}
	expectedResult := `{
    "bar": {
        "value": "value1"
    },
    "baz": {
        "key1": [
            "value2",
            "value3"
        ]
    }
}`

	v, err := RawJSON(config)
	assert.NoError(t, err)

	var out bytes.Buffer
	err = json.Indent(&out, v, "", "    ")
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, out.String())
}
