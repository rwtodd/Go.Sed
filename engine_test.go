package sed // import "go.waywardcode.com/sed"

import (
	"bytes"
	"testing"
)

// a driver for running a program against input, and checking the output
func runprog(t *testing.T, prog, input, expected string) {
	engine, err := New(bytes.NewBufferString(prog))
	if err != nil {
		t.Fatalf("Couldn't parse program <%s>, %s", prog, err.Error())
	}

	result, err := engine.RunString(input)
	if err != nil {
		t.Fatalf("Couldn't run program, %s", err.Error())
	}

	if result != expected {
		t.Fatalf("Program got result <%s> instead of <%s>", result, expected)
	}

}

func TestCommify(t *testing.T) {
	prog := `
# a program to commify numbers
:loop 
s/(.*\d)(\d\d\d)/$1,$2/
t loop
`
	runprog(t, prog,
		"12345\n",
		"12,345\n")
	runprog(t, prog,
		"12345678910\nthe best 1234.56\n",
		"12,345,678,910\nthe best 1,234.56\n")
}

func TestDelete(t *testing.T) {
	runprog(t, "d", "12345\n12345", "")
}

func TestSubst(t *testing.T) {
	runprog(t, `
# test a few features of s/pattern/replacement/flags
s:(\d)(\d)(\d):$1\t$2\t$3:  # put tabs between 3 digits
s/[a-z]/X/3g                # replace lowercase letters with an X, starting with the 3rd one
`,
		"a 234 is the Way\n12345 ONE two three\n",
		"a 2\t3\t4 iX XXX WXX\n1\t2\t345 ONE twX XXXXX\n")
}

func TestG(t *testing.T) {
	runprog(t, "$ !G",
		"one\ntwo\nthree\n",
		"one\n\ntwo\n\nthree\n")
}

func TestRemoveTags(t *testing.T) {
	runprog(t, `
# remove all the tags from an xml/html document
/</{
  :loop
  s/<[^<]*>//g
  /</ {
    N
    b loop
  }
  /^\s*$/d  # skip the line if it was all tags
}`,
		`<html><body>
<table
border=2><tr><td valign=top
align=right>1.</td>
<td>Line 1 Column 2</
td>
</table>
</body></html>`,
		"1.\nLine 1 Column 2\n")
}

func TestCatS(t *testing.T) {
	runprog(t, `
# Write non-empty lines.
/./ {
    p
    d
    }
# Write a single empty line, then look for more empty lines.
/^$/    p
# Get next line, discard the held <newline> (empty line),
# and look for more empty lines.
:Empty
/^$/    {
    N
    s/(?s).//
    b Empty
    }
# Write the non-empty line before going back to search
# for the first in a set of empty lines.
    p
    d
`,
		"one\n\n\n\ntwo\n\n\n\nthree\n",
		"one\n\ntwo\n\nthree\n")
}
