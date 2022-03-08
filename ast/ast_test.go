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
	file := Read(strings.NewReader(config))
	require.Equal(t, "foobar", file.Sections[0].Key)
}

func TestInjectFieldIntoAST(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
MalusDomestica = "Orchard apple"
`
	file := Read(strings.NewReader(config))
	require.Equal(t, "hallmark", file.Sections[0].Name)
	require.Equal(t, 2, file.Sections[0].numFields())
	require.Equal(t, 1, file.Sections[1].numFields())
	require.Equal(t, 6, file.numLines())

	key := "MalusPrunifolia"
	value := "\"Chinese crabapple\""
	section := "rosaceae"

	file = InjectField(file, key, value, section, "", true)

	require.Equal(t, 7, file.numLines())
	require.Equal(t, 2, len(file.Sections))
	require.Equal(t, "hallmark", file.Sections[0].Key)
	require.Equal(t, "rosaceae", file.Sections[1].Key)
	require.Equal(t, 2, len(file.Sections[0].Fields))
	require.Equal(t, 2, len(file.Sections[1].Fields))
	require.Equal(t, 0, len(file.Sections[0].CommentsBefore))
	require.Equal(t, 1, len(file.Sections[1].CommentsBefore))
	require.Equal(t, key, file.Sections[1].Fields[1].Name)
	require.Equal(t, value, file.Sections[1].Fields[1].Value)
}

func TestWriteASTToFile(t *testing.T) {
	config := `; Some preamble
[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
MalusDomestica = "Orchard apple"
`
	// Read config into an ast.File
	file := Read(strings.NewReader(config))
	Write(file, "actual")
	defer os.Remove("actual")

	if err := os.WriteFile("expected", []byte(config), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("actual", "expected"))
}

func TestReadInjectWrite(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
MalusDomestica = "Orchard apple"
`
	file := Read(strings.NewReader(config))
	require.Equal(t, 2, len(file.Sections))
	require.Equal(t, 2, len(file.Sections[0].Fields))
	require.Equal(t, 1, len(file.Sections[1].Fields))
	fieldName := "MalusPrunifolia"
	value := "\"Chinese crabapple\""
	section := "rosaceae"
	file = InjectField(file, fieldName, value, section, "", true)
	require.Equal(t, 2, len(file.Sections))
	require.Equal(t, 2, len(file.Sections[0].Fields))
	require.Equal(t, 2, len(file.Sections[1].Fields))
	Write(file, "actual")
	defer os.Remove("actual")

	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
MalusDomestica = "Orchard apple"
MalusPrunifolia = "Chinese crabapple"
`
	if err := os.WriteFile("expected", []byte(expectedResult), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("actual", "expected"))
}

func TestHandleComments(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
MalusDomestica = "Orchard apple"
`
	file := Read(strings.NewReader(config))
	require.Equal(t, 1, len(file.Sections[1].Fields))

	fieldName := "MalusPrunifolia"
	value := "\"Chinese crabapple\""
	section := "rosaceae"
	file = InjectField(file, fieldName, value, section, "", true)
	Write(file, "actual")
	defer os.Remove("actual")

	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
MalusDomestica = "Orchard apple"
MalusPrunifolia = "Chinese crabapple"
`
	if err := os.WriteFile("expected", []byte(expectedResult), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.Equal(t, 2, len(file.Sections[1].Fields))
	require.True(t, deepCompare("actual", "expected"))
}

func TestHandleSubsections(t *testing.T) {
	config := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
MalusDomestica = "Orchard apple"
`
	file := Read(strings.NewReader(config))
	require.Equal(t, 2, len(file.Sections))
	require.Equal(t, "rosaceae&subsection", file.Sections[1].Key)
	require.Equal(t, "[Rosaceae \"subsection\"]", file.Sections[1].HeadingStr)

	fieldName := "MalusPrunifolia"
	value := "\"Chinese crabapple\""
	section := "rosaceae"
	subsection := "subsection"
	file = InjectField(file, fieldName, value, section, subsection, true)
	Write(file, "actual")
	defer os.Remove("actual")

	expectedResult := `[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
MalusDomestica = "Orchard apple"
MalusPrunifolia = "Chinese crabapple"
`
	if err := os.WriteFile("expected", []byte(expectedResult), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("actual", "expected"))
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
MalusDomestica = "Orchard apple"
`
	file := Read(strings.NewReader(config))
	require.Equal(t, 2, len(file.Sections))
	require.Equal(t, "hallmark", file.Sections[0].Key)
	require.Equal(t, 2, len(file.Sections[0].Fields))
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
	file := Read(strings.NewReader(config))
	require.Equal(t, 3, file.numLines())

	file = Read(strings.NewReader(config1))
	require.Equal(t, 4, file.numLines())

	file = Read(strings.NewReader(config2))
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
MalusDomestica = "Orchard apple"
`
	file := Read(strings.NewReader(config))
	key := "e"
	value := "mc2"
	section := "newSectION"
	subsection := ""
	file = InjectField(file, key, value, section, subsection, true)
	require.Equal(t, 1, len(file.Sections[1].Fields))
	require.Equal(t, 1, len(file.Sections[2].Fields))
	Write(file, "actual")
	defer os.Remove("actual")

	expected := `orange = naranja
red = rojo
; This is a preamble

[hallMaRk]
christmas = merry
newyear = happy

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
MalusDomestica = "Orchard apple"

[newSectION]
e = mc2
`
	if err := os.WriteFile("expected", []byte(expected), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("actual", "expected"))
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
MalusDomestica = "Orchard apple"
`
	file := Read(strings.NewReader(config))
	key := "newyear"
	value := "sad"
	section := "hallmark"
	subsection := ""
	file = InjectField(file, key, value, section, subsection, false)
	Write(file, "actual")
	defer os.Remove("actual")

	expected := `orange = naranja
red = rojo
; This is a preamble

[hallMaRk]
christmas = merry
newyear = sad

[Rosaceae "subsection"]
; Malus is a genus of small deciduous
; trees in the Rosaceae family
MalusDomestica = "Orchard apple"
`
	if err := os.WriteFile("expected", []byte(expected), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("actual", "expected"))
}

func TestFileIsNotModified(t *testing.T) {
	config := `

[Section "blah"  ]   

; this is a comment 

; This is another 

Value =    blah

`
	file := Read(strings.NewReader(config))
	require.Equal(t, 1, len(file.Sections[0].Fields))
	require.Equal(t, 1, len(file.Sections))
	require.Equal(t, "[Section \"blah\"  ]   ", file.Sections[0].HeadingStr)
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
	file := Read(strings.NewReader(config))
	require.Equal(t, 5, len(file.Sections))

	require.Equal(t, 1, len(file.Sections[0].Fields))

	file = DeleteSection(file, "foo", "")
	file = DeleteSection(file, "section", "withsubsection")
	require.Equal(t, 2, len(file.Sections))

	expected := `[bar]

[baz]
; keep this bit
keep = true
`
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestSectionLineHasTrailingComment(t *testing.T) {
	config := `[foo  ] ; a comment containing an '='
key =   value   
[bar]
`
	file := Read(strings.NewReader(config))
	require.Equal(t, 2, len(file.Sections))
	require.Equal(t, config, string(convertASTToBytes(file)))
}

func TestFieldHasTrailingComment(t *testing.T) {
	config := `[foo  ] ; a comment containing an '='
key =   value   ; a TraiLIng coMment
[bar]

another = field ; with another comment


`
	file := Read(strings.NewReader(config))
	require.Equal(t, config, string(convertASTToBytes(file)))

	file = InjectField(file, "key", "Zanzibar", "foo", "", false)
	expected := `[foo  ] ; a comment containing an '='
key = Zanzibar ; a TraiLIng coMment
[bar]

another = field ; with another comment


`
	require.Equal(t, expected, string(convertASTToBytes(file)))

}

func TestFieldValueHasDoubleQuotes(t *testing.T) {
	config := `[foo]
key = "value"

[bar]
another = field
`
	file := Read(strings.NewReader(config))
	require.Equal(t, config, string(convertASTToBytes(file)))
}

func TestDeleteAllFieldsWithName(t *testing.T) {
	config := `[foo]
key = "value"
key = foo
key  = bar 
key   = baz ;  comment
yankee = doodle

[bar]
another = field
`
	file := Read(strings.NewReader(config))
	file = DeleteAllFieldsWithName(file, "key", "foo", "")
	expected := `[foo]
yankee = doodle

[bar]
another = field
`
	Write(file, "actual")
	defer os.Remove("actual")

	if err := os.WriteFile("expected", []byte(expected), 0644); err != nil {
		t.Errorf("Error writing file to disk")
	}
	defer os.Remove("expected")

	require.True(t, deepCompare("actual", "expected"), string(convertASTToBytes(file)))

}

func TestMergeDuplicateSections(t *testing.T) {
	config := `[fruits]
fruit = apple

[foo  "  Sub"]
bar = baz

[Fruits]
fruit = banana

; a comment
[  foo "sub "]
baz =  bar

[ Fruits  ]
 fruit =   cherry ; comment
`
	file := Read(strings.NewReader(config))
	file = MergeAllDuplicateSections(file)

	expected := `

[fruits]
fruit = apple
fruit = banana
 fruit =   cherry ; comment


; a comment
[foo  "  Sub"]
bar = baz
baz =  bar
`
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestAppendFieldToSection(t *testing.T) {
	config := `[fruits]
fruit = apple ;; a comment

[food "vegetables"]
broccoli = green
`
	file := Read(strings.NewReader(config))
	file = AppendFieldToSection(file, "onion", "brown", "food", "vegetables")
	file = AppendFieldToSection(file, "fruit", "mango", "fruits", "")
	expected := `[fruits]
fruit = apple ;; a comment
fruit = mango

[food "vegetables"]
broccoli = green
onion = brown
`
	require.Equal(t, 2, file.Sections[1].numFields())
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestDeleteFieldWithValue(t *testing.T) {
	config := `[fruits]
fruit = apple ;; a comment
fruit = papaya

[food "vegetables"]
broccoli = green
broccoli = red
`
	file := Read(strings.NewReader(config))
	file = DeleteFieldWithValue(file, "broccoli", "red", "food", "vegetables")
	file = DeleteFieldWithValue(file, "fruit", "papaya", "fruits", "")
	expected := `[fruits]
fruit = apple ;; a comment

[food "vegetables"]
broccoli = green
`
	require.Equal(t, 1, file.Sections[0].numFields())
	require.Equal(t, 1, file.Sections[1].numFields())
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestAppendBlankLineToFile(t *testing.T) {
	config := `[fruits]
fruit = apple ;; a comment
fruit = papaya

[food "vegetables"]
broccoli = green
broccoli = red
`
	file := Read(strings.NewReader(config))
	file = AppendBlankLineToFile(file)
	expected := `[fruits]
fruit = apple ;; a comment
fruit = papaya

[food "vegetables"]
broccoli = green
broccoli = red

`
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestAppendBlankLineToCommentFile(t *testing.T) {
	config := `;[fruits]
;fruit = apple ;; a comment
;fruit = papaya
; comment
; preamble
`
	file := Read(strings.NewReader(config))
	file = AppendBlankLineToFile(file)
	file = AppendBlankLineToFile(file)
	file = AppendBlankLineToFile(file)
	expected := `;[fruits]
;fruit = apple ;; a comment
;fruit = papaya
; comment
; preamble



`
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestAppendBlankLineToSection(t *testing.T) {
	config := `[fruits]
fruit = apple ;; a comment
fruit = papaya
; comment
; preamble
[vegetables]
vegetable = broccoli
veg = aubergine
`
	file := Read(strings.NewReader(config))
	file, ok := AppendBlankLineToSection(file, "fruits", "")
	require.True(t, ok)
	file, ok = AppendBlankLineToSection(file, "vegetables", "")
	require.True(t, ok)
	expected := `[fruits]
fruit = apple ;; a comment
fruit = papaya

; comment
; preamble
[vegetables]
vegetable = broccoli
veg = aubergine

`
	require.Equal(t, expected, string(convertASTToBytes(file)))
}

func TestInjectFieldIntoCommentFile(t *testing.T) {
	config := `; comment
; preamble
`
	file := Read(strings.NewReader(config))
	file = InjectField(file, "foo", "bar", "Section", "baz", false)
	expected := `; comment
; preamble

[Section "baz"]
foo = bar
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
