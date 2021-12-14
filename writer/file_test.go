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
		log.Fatalf("Unable to create test file")
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
	section := ast.MakeSection(s)

	if strings.Contains(section.Title, "[") {
		t.Errorf("Expected section title not to contain brackets. Got %v", section.Title)
	}
	if strings.Contains(section.Title, "]") {
		t.Errorf("Expected section title not to contain brackets. Got %v", section.Title)
	}
}
