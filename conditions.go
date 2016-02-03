package sed 

import (
	"fmt"
	"regexp"
)

// conditions are what I'm calling the '1,10' in
// commands ike '1,10 d'.  They are the line numbers,
// regexps, and '$' that you can use to control when
// commands execute.

type condition interface {
	isMet(e *Engine) bool
}

// -----------------------------------------------------
type numbercond int // for matching line number conditions

func (n numbercond) isMet(e *Engine) bool {
	return e.lineno == int(n)
}

// -----------------------------------------------------
type eofcond struct{} // for matching the condition '$'

func (_ eofcond) isMet(e *Engine) bool {
	return e.lastl
}

// -----------------------------------------------------
type regexpcond struct {
	re *regexp.Regexp // for matching regexp conditions
}

func (r *regexpcond) isMet(e *Engine) (answer bool) {
	return r.re.MatchString(e.pat)
}

func newRECondition(s string, loc *location) (*regexpcond, error) {
	re, err := regexp.Compile(s)
	if err != nil {
		err = fmt.Errorf("Regexp Error:  %s %v", err.Error(), loc)
	}
	return &regexpcond{re}, err
}
