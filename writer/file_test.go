package writer

import (
	"bytes"
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
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}
	file := readIntoStruct(f)
	if file.Name == "" {
		t.Error("file.Name == \"\"")
	}
	if len(file.Sections) != 1 {
		t.Errorf("Expected number of sections == 1. Got %v", len(file.Sections))
	}

}

func TestMakeSection(t *testing.T) {
	s := "[foo \"sub\"]"
	section := ast.MakeSection(s, 0)

	if strings.Contains(section.Key, "[") {
		t.Errorf("Expected section key not to contain brackets. Got %v", section.Key)
	}
	if strings.Contains(section.Key, "]") {
		t.Errorf("Expected section key not to contain brackets. Got %v", section.Key)
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
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}
	file := readIntoStruct(f)
	if len(file.Sections) != 2 {
		t.Errorf("Expected number of sections == 2. Got %v", len(file.Sections))
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
	defer os.Remove(f.Name())
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
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	field := ast.Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = injectField(file, field, "rosaceae")

	if file.Lines != 7 {
		t.Errorf("Expected read file to have 7 lines. Got %v", file.Lines)
	} else if len(file.Sections) != 2 {
		t.Errorf("Expected 2 sections in config. Got %v", len(file.Sections))
	} else if !(file.Sections[0].Key == "hallmark" && file.Sections[1].Key == "rosaceae") {
		t.Errorf("Expected sections \"hallmark\" and \"rosaceae\". Got \"%v\" and \"%v\"", file.Sections[0].Key, file.Sections[1].Key)
	} else if len(file.Sections[0].Fields) != 3 {
		t.Errorf("Expected section hallmark to have 2 fields. Got %v", len(file.Sections[0].Fields))
	} else if len(file.Sections[1].Fields) != 2 {
		t.Errorf("Expected section rosaceae to have 2 fields. Got %v", len(file.Sections[1].Fields))
	} else if file.Sections[1].Fields[1].Key != field.Key+" " {
		t.Errorf("Expected injected field to have key \"%v\". Got \"%v\"", field.Key+" ", file.Sections[1].Fields[1].Key)
	} else if file.Sections[1].Fields[1].Value != " "+field.Value {
		t.Errorf("Expected injected field to have value \"%v\". Got \"%v\"", " "+field.Value, file.Sections[1].Fields[1].Value)
	}
}

func TestWriteASTToFile(t *testing.T) {
	config := `; Some preamble
[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)

	// Convert to bytes
	data := convertASTToBytes(file)
	f1, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f1.Name())

	writeBytesToFile(data, f1.Name())

	expectedBytes, err := ioutil.ReadFile(f.Name())
	resultBytes, err := ioutil.ReadFile(f1.Name())

	if bytes.Compare(expectedBytes, resultBytes) != 0 {
		t.Errorf("config and data not the same.\nconfig:\n%v\ndata:\n%v", expectedBytes, resultBytes)
	}
}

func TestReadInjectWrite(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
Malus prunifolia = Chinese crabapple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	field := ast.Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = injectField(file, field, "rosaceae")

	// Convert to bytes
	data := convertASTToBytes(file)

	f1, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f1.Name())

	writeBytesToFile(data, f1.Name())

	expectedBytes := []byte(expectedResult)
	resultBytes, err := ioutil.ReadFile(f1.Name())

	if bytes.Compare(expectedBytes, resultBytes) != 0 {
		t.Errorf("Result and expected not the same.\nconfig:\n%v\ndata:\n%v", expectedBytes, resultBytes)
	}
}

func TestHandleComments(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
`
	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
Malus prunifolia = Chinese crabapple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	if len(file.Sections[1].Fields) != 3 {
		t.Errorf("Expected section \"rosaceae\" to have 3 fields. Got %v", len(file.Sections[1].Fields))
	}

	field := ast.Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = injectField(file, field, "rosaceae")

	// Convert to bytes
	data := convertASTToBytes(file)

	f1, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f1.Name())

	writeBytesToFile(data, f1.Name())

	expectedBytes := []byte(expectedResult)
	resultBytes, err := ioutil.ReadFile(f1.Name())

	if bytes.Compare(expectedBytes, resultBytes) != 0 {
		t.Errorf("Result and expected not the same.\nconfig:\n%v\ndata:\n%v", expectedBytes, resultBytes)
	}
}

func TestHandleSubsections(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
`
	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
Malus prunifolia = Chinese crabapple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	if len(file.Sections) != 2 {
		t.Errorf("Expected 2 sections . Got %v", len(file.Sections))
	}
	if file.Sections[1].Key != "rosaceae \"subsection\"" {
		t.Errorf("Expected section key 'rosaceae \"subsection\"'. Got '%v'", file.Sections[1].Key)
	} else if file.Sections[1].Header != "[Rosaceae \"subsection\"]" {
		t.Errorf("Expected section header '[Rosaceae \"subsection\"]'. Got '%v'", file.Sections[1].Header)
	}

	field := ast.Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = injectField(file, field, "rosaceae \"subsection\"")

	// Convert to bytes
	data := convertASTToBytes(file)

	f1, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f1.Name())

	writeBytesToFile(data, f1.Name())

	expectedBytes := []byte(expectedResult)
	resultBytes, err := ioutil.ReadFile(f1.Name())

	if bytes.Compare(expectedBytes, resultBytes) != 0 {
		t.Errorf("Result and expected not the same.\ninput config:\n%v\noutput config:\n%v", expectedBytes, resultBytes)
	}
}

func TestPreambleSection(t *testing.T) {
	config := `orange = naranja
red = rojo
; This is a preamble

[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)

	if len(file.Sections) != 3 {
		t.Errorf("Expected 3 sections including preamble. Got %v", len(file.Sections))
	} else if file.Sections[0].Key != "_preamble" {
		t.Errorf("Expected first section key to be called \"_preamble\". Got \"%v\"", file.Sections[0].Key)
	} else if len(file.Sections[0].Fields) != 4 {
		t.Errorf("Expected preamble to have 4 fields. Got %v", len(file.Sections[0].Fields))
	} else if file.Lines != 12 {
		t.Errorf("Expected read file to contain 12 lines. Got %v", file.Lines)
	}

}

func TestLineCounts(t *testing.T) {
	config := `orange = naranja
red = rojo
; This is a preamble
`
	config1 := `orange = naranja
red = rojo
; This is a preamble
; Elbmaerp a si siht
`
	config2 := `; This is a preamble

[hallMaRk]
christmas = merry
; This is a comment

[anotherSeCTion]
field = value
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	f1, err := os.CreateTemp("", "test1")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f1.Name())
	if _, err := f1.Write([]byte(config1)); err != nil {
		log.Fatal(err)
	}

	f2, err := os.CreateTemp("", "test2")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f2.Name())
	if _, err := f2.Write([]byte(config2)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	file1 := readIntoStruct(f1)
	file2 := readIntoStruct(f2)

	if file.Lines != 3 {
		t.Errorf("Expected file to have 3 lines. Got %v", file.Lines)
	} else if file1.Lines != 4 {
		t.Errorf("Expected file to have 4 lines. Got %v", file1.Lines)
	} else if file2.Lines != 8 {
		t.Errorf("Expected file to have 8 lines. Got %v", file2.Lines)
	}

}

func TestInjectIntoNewSection(t *testing.T) {
	config := `orange = naranja
red = rojo
; This is a preamble

[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
`
	f, err := os.CreateTemp("", "test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte(config)); err != nil {
		log.Fatal(err)
	}

	file := readIntoStruct(f)
	field := ast.MakeField("e = mc2")
	injectField(file, field, "newSectION")

	expected := `orange = naranja
red = rojo
; This is a preamble

[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple

[newSectION]
e = mc2
`
	resultsFile, err := os.CreateTemp("", "result")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(resultsFile.Name())
	if _, err := f.Write([]byte(expected)); err != nil {
		log.Fatal(err)
	}

	file.PrintDebug()

	expectedBytes, err := ioutil.ReadFile(f.Name())
	resultBytes, err := ioutil.ReadFile(resultsFile.Name())
	if bytes.Compare(expectedBytes, resultBytes) != 0 {
		t.Errorf("config and data not the same.\nconfig:\n%v\ndata:\n%v", expectedBytes, resultBytes)
	}
}
