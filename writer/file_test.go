package writer

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/please-build/gcfg/ast"
)

func TestReadFile(t *testing.T) {
	config := `[foo "sub"]
bar = value1
baz = value2
f--bar = true
bfoo = false
bazbar = 0
foobaz = -5
bar-foo = 10
baz-baz = value3
baz-baz = value4
bar-key = bar-value
ǂbar = 
xfoo = 
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}
	file := readIntoStruct(f)
	if file.Name == "" {
		t.Error("file.Name == \"\"")
	}
	if file.NumSections != 1 {
		t.Errorf("Expected number of sections == 1. Got %v", file.NumSections)
	}

}

func TestMakeSection(t *testing.T) {
	s := "[foo \"sub\"]"
	section := ast.MakeSection(s, 0)

	if strings.Contains(section.Key, "[") {
		t.Errorf("Expected section title not to contain brackets. Got %v", section.Key)
	}
	if strings.Contains(section.Key, "]") {
		t.Errorf("Expected section title not to contain brackets. Got %v", section.Key)
	}
}

func TestSectionLineNumbers(t *testing.T) {
	config := `[foo "sub"]
bar = value1
baz = value2
f--bar = true
bfoo = false
bazbar = 0
[curried crab]
foobaz = -5
bar-foo = 10
baz-baz = value3
baz-baz = value4
bar-key = bar-value
ǂbar = 
xfoo = 
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}
	file := readIntoStruct(f)
	if file.NumSections != 2 {
		t.Errorf("Expected number of sections == 2. Got %v", file.NumSections)
	}
	if !(file.Sections[0].Line == 0 && file.Sections[1].Line == 6) {
		t.Errorf("Expected sections at lines 0 and 6 but got %v and %v", file.Sections[0].Line, file.Sections[1].Line)
	}

}

func TestGetSectionKey(t *testing.T) {
	config := `[FOObar]
bar = value1
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	if file.Sections[0].Key != "foobar" {
		t.Errorf("Expected foobar. Got %v", file.Sections[0].Key)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

}

func TestInjectFieldIntoAST(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	field := ast.Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = injectField(file, field, "rosaceae")

	if file.Lines != 6 {
		t.Errorf("Expected read file to have 7 lines. Got %v", file.Lines)
	} else if file.NumSections != 2 {
		t.Errorf("Expected 2 sections in config. Got %v", file.NumSections)
	} else if !(file.Sections[0].Key == "hallmark" && file.Sections[1].Key == "rosaceae") {
		t.Errorf("Expected sections \"hallmark\" and \"rosaceae\". Got \"%v\" and \"%v\"", file.Sections[0].Key, file.Sections[1].Key)
	} else if len(file.Sections[0].Fields) != 1 {
		t.Errorf("Expected section hallmark to have 1 field. Got %v", len(file.Sections[0].Fields))
	} else if len(file.Sections[1].Fields) != 2 {
		t.Errorf("Expected section rosaceae to have 2 fields. Got %v", len(file.Sections[1].Fields))
	} else if file.Sections[1].Fields[1].Key != field.Key {
		t.Errorf("Expected injected field to have key %v. Got %v", field.Key, file.Sections[1].Fields[1].Key)
	} else if file.Sections[1].Fields[1].Value != field.Value {
		t.Errorf("Expected injected field to have value %v. Got %v", field.Value, file.Sections[1].Fields[1].Value)
	}
}

func TestWriteASTToFile(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)

	// Convert to bytes
	data := convertASTToBytes(file)

	if bytes.Compare([]byte(config), data) != 0 {
		t.Errorf("config and data not the same. config=\n%v\ndata=\n%v", config, data)
	}
}
