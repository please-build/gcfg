package writer

import (
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

	if strings.Contains(section.Title, "[") {
		t.Errorf("Expected section title not to contain brackets. Got %v", section.Title)
	}
	if strings.Contains(section.Title, "]") {
		t.Errorf("Expected section title not to contain brackets. Got %v", section.Title)
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

func TestCanonicaliseSectionTitle(t *testing.T) {
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
	if file.Sections[0].Title != "foobar" {
		t.Errorf("Expected foobar. Got %v", file.Sections[0].Title)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

}

func TestInjectFieldIntoAST(t *testing.T) {
	config := `[hallmark]
christmas = merry

[Rosaceae]
Malus domestica = apple
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
		Key:   "newyear",
		Value: "happy",
	}
	injectField(file, field, "rosaceae")

	if file.NumSections != 2 {
		t.Errorf("Expected 2 sections in config. Got %v", file.NumSections)
	} else if !(file.Sections[0].Title == "hallmark" && file.Sections[1].Title == "rosaceae") {
		t.Errorf("Expected sections \"hallmark\" and \"rosaceae\". Got \"%v\" and \"%v\"", file.Sections[0].Title, file.Sections[1].Title)
	} else if len(file.Sections[1].Fields) != 2 {
		t.Errorf("Expected section Rosaceae to have 2 fields. Got %v", len(file.Sections[1].Fields))
	} else if file.Sections[1].Fields[1].Key != field.Key {
		t.Errorf("Expected injected field to have key %v. Got %v", field.Key, file.Sections[1].Fields[1].Key)
	} else if file.Sections[1].Fields[1].Value != field.Value {
		t.Errorf("Expected injected field to have value %v. Got %v", field.Value, file.Sections[1].Fields[1].Value)
	}
}
