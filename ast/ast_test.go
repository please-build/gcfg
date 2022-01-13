package ast

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSectionKey(t *testing.T) {
	config := `[FOObar]
bar = value1
`
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, file.Sections[0].Key, "foobar")
}

func TestInjectFieldIntoAST(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, file.numLines(), 6)

	field := Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = InjectField(file, field, "rosaceae", true)

	require.Equal(t, file.numLines(), 7)
	require.Equal(t, len(file.Sections), 2)
	require.Equal(t, file.Sections[0].Key, "hallmark")
	require.Equal(t, file.Sections[1].Key, "rosaceae")
	require.Equal(t, len(file.Sections[0].Fields), 3)
	require.Equal(t, len(file.Sections[1].Fields), 2)
	require.Equal(t, file.Sections[1].Fields[1].Key, field.Key+" ")
	require.Equal(t, file.Sections[1].Fields[1].Value, " "+field.Value)
}

func TestWriteASTToFile(t *testing.T) {
	config := `; Some preamble
[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	// Read config into an ast.File
	file := Read(strings.NewReader(config), "test")
	Write(file, file.Name)
	defer os.Remove(file.Name)

	if err := os.WriteFile("expected", []byte(config), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("test", "expected"), string(convertASTToBytes(file)))
}

func TestReadInjectWrite(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	file := Read(strings.NewReader(config), "test")
	field := Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = InjectField(file, field, "rosaceae", true)
	Write(file, file.Name)
	defer os.Remove(file.Name)

	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
Malus prunifolia = Chinese crabapple
`
	if err := os.WriteFile("expected", []byte(expectedResult), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("test", "expected"))
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
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, len(file.Sections[1].Fields), 3)

	field := Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = InjectField(file, field, "rosaceae", true)
	Write(file, file.Name)
	defer os.Remove(file.Name)

	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
Malus prunifolia = Chinese crabapple
`
	if err := os.WriteFile("expected", []byte(expectedResult), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.Equal(t, len(file.Sections[1].Fields), 4)
	require.True(t, deepCompare("test", "expected"))
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
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, len(file.Sections), 2)
	require.Equal(t, file.Sections[1].Key, "rosaceae \"subsection\"")
	require.Equal(t, file.Sections[1].Header, "[Rosaceae \"subsection\"]")

	field := Field{
		Key:   "Malus prunifolia",
		Value: "Chinese crabapple",
	}
	file = InjectField(file, field, "rosaceae \"subsection\"", true)
	Write(file, file.Name)
	defer os.Remove(file.Name)

	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
Malus prunifolia = Chinese crabapple
`
	if err := os.WriteFile("expected", []byte(expectedResult), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("test", "expected"))
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
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, len(file.Sections), 3)
	require.Equal(t, file.Sections[0].Key, "_preamble")
	require.Equal(t, len(file.Sections[0].Fields), 4)
	require.Equal(t, file.numLines(), 12)
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
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, file.numLines(), 3)

	file = Read(strings.NewReader(config1), "test")
	require.Equal(t, file.numLines(), 4)

	file = Read(strings.NewReader(config2), "test")
	require.Equal(t, file.numLines(), 8)
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
	file := Read(strings.NewReader(config), "test")
	field := makeField("e = mc2")
	file = InjectField(file, field, "newSectION", true)
	Write(file, file.Name)
	defer os.Remove(file.Name)

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
	if err := os.WriteFile("expected", []byte(expected), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("test", "expected"))
}

func TestInjectNonRepeatableField(t *testing.T) {
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
	file := Read(strings.NewReader(config), "test")
	field := makeField("newyear = sad")
	file = InjectField(file, field, "hallmark", false)
	Write(file, file.Name)
	defer os.Remove(file.Name)

	expected := `orange = naranja
red = rojo
; This is a preamble

[hallMaRk]
christmas = merry
newyear = sad

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
Malus domestica = Orchard apple
`
	if err := os.WriteFile("expected", []byte(expected), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("test", "expected"))
}

func TestFileIsNotModified(t *testing.T) {
	config := `

[Section "blah"  ]   

; this is a comment 

; This is another 

Value =    blah

`
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, len(file.Sections[0].Fields), 2)
	require.Equal(t, len(file.Sections), 2)
	require.Equal(t, file.Sections[1].Header, "[Section \"blah\"  ]   ")
	require.Equal(t, config, string(convertASTToBytes(file)))
}

func TestDeleteSection(t *testing.T) {
	config := `[foo]
; a comment
key =   value   
[bar]

[baz]
; keep this bit
keep = true



[foo]
# this and the first section should be deleted
`
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, 4, len(file.Sections))

	file = DeleteSection(file, "foo")
	require.Equal(t, 2, len(file.Sections))

	expected := `[bar]

[baz]
; keep this bit
keep = true



`
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

const chunkSize = 64000

func deepCompare(file1, file2 string) bool {
	f1, err := os.Open(file1)
	if err != nil {
		log.Fatal(err)
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}
