package gcfg

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInvalidFields(t *testing.T) {
	config := &struct{}{}

	_, err := Get(config, "foo")
	assert.Error(t, err)

	_, err = Get(config, "foo.bar.baz")
	assert.Error(t, err)
}

func TestGet1(t *testing.T) {
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

	// Valid
	res, err := Get(config, "foo.bar")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value1"}, res)

	res, err = Get(config, "foo.baz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value2"}, res)

	res, err = Get(config, "foo.f--bar")
	assert.NoError(t, err)
	assert.Equal(t, []string{"true"}, res)

	res, err = Get(config, "foo.bfoo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"false"}, res)

	res, err = Get(config, "foo.bazbar")
	assert.NoError(t, err)
	assert.Equal(t, []string{"0"}, res)

	res, err = Get(config, "foo.foobaz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"-5"}, res)

	res, err = Get(config, "foo.baz-baz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value3", "value4"}, res)

	res, err = Get(config, "foo.bar-key")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar-value"}, res)

	res, err = Get(config, "foo.ǂbar")
	assert.NoError(t, err)
	assert.Equal(t, []string{""}, res)

	res, err = Get(config, "foo.xfoo")
	assert.NoError(t, err)
	assert.Equal(t, []string{""}, res)

	// Invalid
	_, err = Get(config, "nosection.foo")
	assert.Error(t, err)

	_, err = Get(config, "foo.novar")
	assert.Error(t, err)
}

func TestGet2(t *testing.T) {
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

	// Valid
	res, err := Get(config, "foo.sub.bar")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value1"}, res)

	res, err = Get(config, "foo.sub.baz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value2"}, res)

	res, err = Get(config, "foo.sub.f--bar")
	assert.NoError(t, err)
	assert.Equal(t, []string{"true"}, res)

	res, err = Get(config, "foo.sub.bfoo")
	assert.NoError(t, err)
	assert.Equal(t, []string{"false"}, res)

	res, err = Get(config, "foo.sub.bazbar")
	assert.NoError(t, err)
	assert.Equal(t, []string{"0"}, res)

	res, err = Get(config, "foo.sub.foobaz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"-5"}, res)

	res, err = Get(config, "foo.sub.baz-baz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value3", "value4"}, res)

	res, err = Get(config, "foo.sub.bar-key")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar-value"}, res)

	res, err = Get(config, "foo.sub.ǂbar")
	assert.NoError(t, err)
	assert.Equal(t, []string{""}, res)

	res, err = Get(config, "foo.sub.xfoo")
	assert.NoError(t, err)
	assert.Equal(t, []string{""}, res)

	// Invalid
	_, err = Get(config, "foo.nosub.novar")
	assert.Error(t, err)

	_, err = Get(config, "foo.sub.nonexisting")
	assert.Error(t, err)
}

func TestGet3(t *testing.T) {
	config := &struct {
		Foo map[string]string
	}{
		Foo: map[string]string{"key": "value"},
	}

	res, err := Get(config, "foo.key")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value"}, res)
}

func TestGet4(t *testing.T) {
	config := &struct {
		Foo map[string]string
	}{
		Foo: map[string]string{"sub key": "value"},
	}

	res, err := Get(config, "foo.sub.key")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value"}, res)
}

func TestGet5(t *testing.T) {
	config := &struct {
		Bar subtypeStructNoMarshaler
		Baz extraValuesStruct
	}{
		Baz: extraValuesStruct{
			ExtraValues: map[string][]string{"key1": {"value2", "value3"}},
		},
	}

	res, err := Get(config, "baz.key1")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value2", "value3"}, res)
}
