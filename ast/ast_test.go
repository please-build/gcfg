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
	require.Equal(t, "foobar", file.sections[0].key)
}

func TestInjectFieldIntoAST(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
Malus domestica = Orchard apple
`
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, 6, file.numLines())

	key := "Malus prunifolia"
	value := "Chinese crabapple"
	section := "rosaceae"

	file.PrintDebug()
	file = InjectField(file, key, value, section, "", true)
	file.PrintDebug()

	require.Equal(t, 7, file.numLines())
	require.Equal(t, 2, len(file.sections))
	require.Equal(t, "hallmark", file.sections[0].key)
	require.Equal(t, "rosaceae", file.sections[1].key)
	require.Equal(t, 3, len(file.sections[0].fields))
	require.Equal(t, 2, len(file.sections[1].fields))
	require.Equal(t, key, file.sections[1].fields[1].key)
	require.Equal(t, value, file.sections[1].fields[1].value)
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
	Write(file, file.name)
	defer os.Remove(file.name)

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
	require.Equal(t, 2, len(file.sections))
	require.Equal(t, 3, len(file.sections[0].fields))
	require.Equal(t, 1, len(file.sections[1].fields))
	key := "Malus prunifolia"
	value := "Chinese crabapple"
	section := "rosaceae"
	file = InjectField(file, key, value, section, "", true)
	require.Equal(t, 2, len(file.sections))
	require.Equal(t, 3, len(file.sections[0].fields))
	require.Equal(t, 2, len(file.sections[1].fields))
	Write(file, file.name)
	defer os.Remove(file.name)

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
	require.Equal(t, 3, len(file.sections[1].fields))

	key := "Malus prunifolia"
	value := "Chinese crabapple"
	section := "rosaceae"
	file = InjectField(file, key, value, section, "", true)
	Write(file, file.name)
	defer os.Remove(file.name)

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

	require.Equal(t, 4, len(file.sections[1].fields))
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
	require.Equal(t, 2, len(file.sections))
	require.Equal(t, "rosaceae&subsection", file.sections[1].key)
	require.Equal(t, "[Rosaceae \"subsection\"]", file.sections[1].str)

	key := "Malus prunifolia"
	value := "Chinese crabapple"
	section := "rosaceae"
	subsection := "subsection"
	file = InjectField(file, key, value, section, subsection, true)
	Write(file, file.name)
	// defer os.Remove(file.name)

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
	// defer os.Remove("expected")

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
	require.Equal(t, 3, len(file.sections))
	require.Equal(t, "_preamble", file.sections[0].key)
	require.Equal(t, 4, len(file.sections[0].fields))
	require.Equal(t, 12, file.numLines())
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
	require.Equal(t, 3, file.numLines())

	file = Read(strings.NewReader(config1), "test")
	require.Equal(t, 4, file.numLines())

	file = Read(strings.NewReader(config2), "test")
	require.Equal(t, 8, file.numLines())
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
	key := "e"
	value := "mc2"
	section := "newSectION"
	subsection := ""
	file = InjectField(file, key, value, section, subsection, true)
	Write(file, file.name)
	// defer os.Remove(file.name)

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
	// defer os.Remove("expected")

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
	key := "newyear"
	value := "sad"
	section := "hallmark"
	subsection := ""
	file.PrintDebug()
	file = InjectField(file, key, value, section, subsection, false)
	log.Printf("file.sections[1].fields[1].key=%v", file.sections[1].fields[1].key)
	log.Printf("file.sections[1].fields[1].value=%v", file.sections[1].fields[1].value)
	file.PrintDebug()
	Write(file, file.name)
	defer os.Remove(file.name)

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
	require.Equal(t, 2, len(file.sections[0].fields))
	require.Equal(t, 2, len(file.sections))
	require.Equal(t, "[Section \"blah\"  ]   ", file.sections[1].str)
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

[section "withsubsection"]
; stuff
stuff = stuff
`
	file := Read(strings.NewReader(config), "test")
	require.Equal(t, 5, len(file.sections))

	file = DeleteSection(file, "foo", "")
	file = DeleteSection(file, "section", "withsubsection")
	require.Equal(t, 2, len(file.sections))

	expected := `[bar]

[baz]
; keep this bit
keep = true



`
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestMakeSectionFromString(t *testing.T) {
	s := ` [ section   "subsection"  ] `
	section := makeSectionFromString(s)
	require.Equal(t, "section", section.name)
	require.Equal(t, "subsection", section.subsection)
	require.Equal(t, s, section.str)

	s = `[  [ badsection ]`
	// require.Panics(t, makeSectionFromString(s))
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
