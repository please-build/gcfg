package writer

import (
	"io/ioutil"
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
	filename := "test_file.txt"
	err := ioutil.WriteFile(filename, []byte(config), 0755)
	if err != nil {
		log.Fatalf("Failed to create test file")
	}
	defer os.Remove(filename)

	file := readIntoStruct("test_file.txt")
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
	filename := "test_file.txt"
	err := ioutil.WriteFile(filename, []byte(config), 0755)
	if err != nil {
		log.Fatalf("Failed to create test file")
	}
	defer os.Remove(filename)

	file := readIntoStruct("test_file.txt")
	if file.NumSections != 2 {
		t.Errorf("Expected number of sections == 2. Got %v", file.NumSections)
	}
	if !(file.Sections[0].Line == 0 && file.Sections[1].Line == 6) {
		t.Errorf("Expected sections at lines 0 and 6 but got %v and %v", file.Sections[0].Line, file.Sections[1].Line)
	}

}
