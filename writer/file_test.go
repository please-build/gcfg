package writer

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestReadFileFoo(t *testing.T) {
	err := ioutil.WriteFile("test_file.txt", []byte("Hello\n[section_one]\nI\nAm\nA\nTest\nFile"), 0755)
	if err != nil {
		log.Fatalf("Unable to create test file")
	}
	file := readIntoStruct("test_file.txt")

	if file.name == "" {
		t.Fail()
	} else {
		t.Logf("got file name %v", file.name)
	}
	if file.numSections != 1 {
		t.Fail()
	}
}
