package sed

import (
	"bufio"
	"bytes"
	"testing"
)

func TestCommify(t *testing.T) {
	prog := `
# a program to commify numbers
:loop 
s/(.*\d)(\d\d\d)/$1,$2/
t loop
`
	engine, err := New(bufio.NewReader(bytes.NewBufferString(prog)))
	if err != nil {
		t.Fatalf("Couldn't parse commify, %s", err.Error())
	}

	result, err := engine.RunString("12345\n")
	if err != nil {
		t.Fatalf("Couldn't run commify, %s", err.Error())
	}

	if result != "12,345\n" {
		t.Fatalf("Commify got result <%s> instead of 12,345", result)
	}
}

func TestDelete(t *testing.T) {
	prog := "d"
	engine, err := New(bufio.NewReader(bytes.NewBufferString(prog)))
	if err != nil {
		t.Fatalf("Couldn't parse delete prog, %s", err.Error())
	}

	result, err := engine.RunString("12345\n12345")
	if err != nil {
		t.Fatalf("Couldn't run delete prog, %s", err.Error())
	}

	if result != "" {
		t.Fatalf("Delete prog got result <%s> instead of an empty string", result)
	}

}
