package main

// conditions are what I'm calling the '1,10' in
// commands ike '1,10 d'.  They are the line numbers,
// regexps, and '$' that you can use to control when
// commands execute.

type condition interface {
	isMet(e *engine) bool
}

type numbercond int

func (n numbercond) isMet(e *engine) bool {
	return e.lineno == int(n)
}
