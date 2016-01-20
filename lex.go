package main

// the lexer for SED.  The point of the lexer is to
// reliably transform the input into a series of token structs.
// These structs know the source location, and the token type, and
// any arguments to the token (e.g., a regexp's '/' argument is the
// regular expression itself).
//
// The lexer also simplifies and regularises the input, for instance
// by eliminating comments, and making every command appear to the
// parser as if it is inside braces:
//    1,10 d   ==>   1,10 { d }

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
	TOK_CMD
	TOK_LABEL
)

type token struct {
	location
	tok  int
	args []string
}
