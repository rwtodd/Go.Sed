package main

// the lexer for SED.  The point of the lexer is to
// reliably transform the input into a series of token structs.
// These structs know the source location, and the token type, and
// any arguments to the token (e.g., a regexp's '/' argument is the
// regular expression itself).
//
// The lexer also simplifies and regularises the input, for instance
// by eliminating comments.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode"
)

type location struct {
	line int
	pos  int
}

const (
	TOK_NUM = iota
	TOK_RX
	TOK_COMMA
	TOK_BANG
	TOK_LBRACE
	TOK_RBRACE
	TOK_EOL
	TOK_CMD
	TOK_LABEL
)

type token struct {
	location
	tok  int
	args []string
}

// ----------------------------------------------------------
//  Location-tracking reader
// ----------------------------------------------------------
type loc_reader struct {
	location
	eol bool // state for end of line, true when last rune was '\n'
	r   io.RuneScanner
}

func (lr *loc_reader) ReadRune() (rune, int, error) {
	r, i, err := lr.r.ReadRune()

	lr.pos++

	if lr.eol {
		lr.pos = 1
		lr.line++
		lr.eol = false
	}
	if r == '\n' {
		lr.eol = true
	}

	return r, i, err
}

func (lr *loc_reader) UnreadRune() error {
	return lr.r.UnreadRune()
}

// ----------------------------------------------------------
// lexer functions
// ----------------------------------------------------------
func skipComment(r io.RuneReader) (rune, error) {
	var err error
	var cur rune = ' '
	for (cur != '\n') && (err == nil) {
		cur, _, err = r.ReadRune()
	}
	return ';', err
}

func skipWS(r io.RuneReader) (rune, error) {
	var err error
	var cur rune = ' '
	for {
		switch {
		case cur == '\n':
			return ';', err
		case cur == '#':
			return skipComment(r)
		case !unicode.IsSpace(cur):
			return cur, err
		}
		cur, _, err = r.ReadRune()
	}
}

func readNumber(r io.RuneScanner, character rune) (string, error) {
	var buffer bytes.Buffer

	var err error
	for (err == nil) && unicode.IsDigit(character) {
		buffer.WriteRune(character)
		character, _, err = r.ReadRune()
	}

	if err == nil {
		err = r.UnreadRune()
	}

	return buffer.String(), err
}

// readDelimited reads until it finds the delimter character,
// returning the string (not including the delimiter). It does
// allow the delimiter to be escaped by a backslash ('\').
// It is an error to reach EOL while looking for the delimiter.
func readDelimited(r io.RuneScanner, delimiter rune) (string, error) {
	var buffer bytes.Buffer

	var err error
	var character rune
	var previous rune

	character, _, err = r.ReadRune()
	for (err == nil) &&
		(character != '\n') &&
		((character != delimiter) || (previous == '\\')) {
		buffer.WriteRune(character)
		previous = character
		character, _, err = r.ReadRune()
	}

	if character == '\n' {
		err = fmt.Errorf("end-of-line while looking for %v", delimiter)
	}
	return buffer.String(), err
}

// readIdentifier skips any whitespace, and then reads until it
// finds either a ';' or a non-alphanumeric character.  It
// returns the string it reads.
func readIdentifier(r io.RuneScanner) (string, error) {
	var buffer bytes.Buffer

	var err error
	var character rune

	character, err = skipWS(r)
	for (err == nil) && (character != ';') && !unicode.IsSpace(character) {
		buffer.WriteRune(character)
		character, _, err = r.ReadRune()
	}

	if err == nil {
		err = r.UnreadRune()
	}
	return buffer.String(), err
}

func readSubstitution(r io.RuneScanner) ([]string, error) {
	var ans = []string{"s"}
	var err error

	// step 1.: get the delimiter character for substitutions
	var delimiter rune
	delimiter, _, err = r.ReadRune()
	if err != nil {
		return ans, err
	}

	// step 2.: read the regexp
	var part1 string
	part1, err = readDelimited(r, delimiter)
	if err != nil {
		return ans, err
	}

	// step 3.: read the replacement
	var part2 string
	part2, err = readDelimited(r, delimiter)
	if err != nil {
		return ans, err
	}

	// step 4.: read the modifiers
	var mods string
	mods, err = readIdentifier(r)

	return append(ans, part1, part2, mods), err
}

func lex(r io.RuneScanner, ch chan *token) {
	defer close(ch)

	rdr := loc_reader{}
	rdr.r = r
	rdr.eol = true

	var err error
	var cur rune

	for err == nil {
		cur, err = skipWS(&rdr)
		if err != nil {
			break
		}
		switch {
		case cur == ';':
			ch <- &token{rdr.location, TOK_EOL, nil}
		case cur == ',':
			ch <- &token{rdr.location, TOK_COMMA, nil}
		case cur == '{':
			ch <- &token{rdr.location, TOK_LBRACE, nil}
		case cur == '}':
			ch <- &token{rdr.location, TOK_RBRACE, nil}
		case cur == '!':
			ch <- &token{rdr.location, TOK_BANG, nil}
		case cur == '/':
			var rx string
			rx, err = readDelimited(&rdr, '/')
			ch <- &token{rdr.location, TOK_RX, []string{rx}}
		case cur == ':':
			var label string
			label, err = readIdentifier(&rdr)
			ch <- &token{rdr.location, TOK_LABEL, []string{label}}
		case cur == 'b', cur == 't': // branches...
			var label string
			label, err = readIdentifier(&rdr)
			ch <- &token{rdr.location, TOK_CMD, []string{string(cur), label}}
		case cur == 's': // substitution
			var args []string
			args, err = readSubstitution(&rdr)
			ch <- &token{rdr.location, TOK_CMD, args}
		case unicode.IsDigit(cur):
			var num string
			num, err = readNumber(&rdr, cur)
			ch <- &token{rdr.location, TOK_NUM, []string{num}}
		default:
			ch <- &token{rdr.location, TOK_CMD, []string{string(cur)}}
		}
	}

	if err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error reading... <%s> near <%+v>", err.Error(), rdr.location)
	}
}
