# Go-Sed 

An implementation of sed in Go.  Just because!

## Status

  * __Command-Line processing__:  Done. It accepts '-e', '-f', '-n' and long
versions of the same. It takes '-help'.
  * __Lexer__: Complete.
  * __Parser/Engine__:  Has every command in a typical sed now. 
 It has:  a\, i\, c\, d, D, p, P, g, G, x, h, H, r, w, s, y, b, t, :label, n, N, q, =.

This `sed` engine can be embedded in your program, wrapping any `io.Reader` so that
the stream is lazily processed as you read from it.  Of course I also have a command-line
driver program here (in package `go.waywardcode.com/sed/cmd/go-sed`).

## Differences from Standard Sed

__Regexps__: The only thing you really have to keep in mind when using 
go-sed, is I use Go's "regexp" package. Therefore, you have to use that 
syntax for the regular expressions.  The main differences I've noticed 
in practice are: 

| Go-sed          |  Traditional RE   | Notes                             |
| --------------- | ----------------- | --------------------------------- |
|  s/a(bc*)d/$1/g |  s/a\\(bc*\\)d/\1/g | Don't escape (); Use $1, $2, etc. |
|  s/(?s).//      |  s/.//            | If you want dot to match \n, use (?s) flag.  |

Go's regexps have many rich options, which you can see [here](https://github.com/google/re2/wiki/Syntax).

There are a few niceties though, such as I interpret '\t' and '\n' in 
replacement strings:

    s/\w+\s*/\t$0\n/g

You can also escape the newline like in a typical sed, if you want.

__Slightly Friendlier Syntax__: Go-sed is a little more user-friendly when it comes to
syntax.  In a normal sed, you have to use one (and ONLY one)
space between a `r` or `w` and the filename. Go-sed eats whitespace until it
sees the filename.

Also, in a typical sed, you need seemingly-extraneous semicolons like the one after the `d` below: 

    sed -e '/re/ { p ; d; }' < in > out

... but go-sed is nicer about it:

    go-sed -e '/re/ { p ; d }' < in > out 

__Unicode__: Original sed had no idea about unicode, but modern ones do, and 
Go-sed is unicode-friendly:

    go-sed -e 'y/go/世界/' < in > out

## Go Get

You can get the code/executable by saying:

    go get go.waywardcode.com/sed/cmd/go-sed

And, if you want to embed a sed engine in your own program, you can import:

    import "go.waywardcode.com/sed"

The package docs are up at [https://godoc.org/go.waywardcode.com/sed](https://godoc.org/go.waywardcode.com/sed).

## Implementation Notes

I have never looked at how a "real" implementation of sed is done. I'm just
going by the sed man pages and tutorials.  I will note that in speed comparisons, 
go-sed outperforms Mac OS X's sed on my iMac, as long as the input isn't tiny.  So, I
think the architecture here is pretty good.

The library is spread out among several files:

  * _lex.go_: Lexes the input into tokens. Skips over comments. These are pretty
  large-grained tokens. For example, when it reads a 's'ubstitution command, it
  pulls in the arguments and modifiers and packages them into a single token.  This makes
  the parser simpler.  The tokens are sent over a channel so that the lexer can run concurrently
  with the parser. This is more of a design win than a performance win, but it is more than
  fast enough.
  * _parse.go_: Takes tokens from the lexer and parses the sed program. It's really a
  parser+compiler, becuase the output is an array of instructions for the VM to 
  interpret.  Because the tokens are designed to be pretty self-contained, this parser
  doesn't ever need to backtrack.  I always like it when I can achieve that.

  When branch targets (_e.g._, `:loop`) are parsed, a branch instruction to that location is
  stored off, along with the name.  Then, after the initial parse, each branch
  is fixed up against the proper target.

  * _instructions.go_: This file holds all of the VM instructions except for substitution and 
  translation, which are in _substitution.go_.  An instruction for the VM I've built is just a 
  `func (svm *vm) error`.   This turned out to be a very flexible arrangement.  Most instructions
  have a 1:1 mapping to sed commands, but others (like the 'n' command) are broken up at parse time
  into multiple engine instructions.

  Simple instructions, like `cmd_get` (which handles the `g` sed command), are package functions, which
  can get inserted into the instruction stream context-free.  A few instructions have state, such
  as the string to insert in an `i\` command. For those, a closure holds the state:

      func cmd_newInserter(text string) instruction {
           return func(svm *vm) error {
               vm.ip++
               _,err := svm.output.WriteString(text)
               return err
           }
      }
  
  A couple commands have so much state that a simple closure would be unwieldy, so those get a struct
  and an associated `run` method. That `run` method pointer becomes the instruction.

  You can see in the example above that each instruction is responsible for incrementing the IP (_instruction pointer_)
  in the VM. That's flexible because many of the instructions branch, and they can set the IP to whatever
  they need. However, this was the __number one__ cause of bugs during development: I'd add a new command, and forget
  to increment the IP, leading to an infinite loop on that instruction.  So, I possibly should have had the 
  VM auto-increment the IP, and have the branching instructions account for that when setting IP.  It was a
  trade-off between keeping the inner loop as tight as possible and keeping the instructions as simple as 
  possible.  I might have made the wrong choice there. 

  * _engine.go_: This is the sed-VM, and this file also has the entire public interface to the library. It 
  is arranged for simplicity. You have one function to create an Engine from a sed program, and you can
  use that Engine to wrap an `io.Reader`. The same engine can be re-used against multiple inputs. 

  The inner loop of the interpreter very compact:

      for err == nil {
         err = svm.ins[svm.ip](svm)
      }

  * _conditions.go_: Conditions are what I call the guards around commands (like the `1,10` in `1,10d`). The
  sed man pages act like the conditions are part of the command, but in my engine VM, they are commands themselves.
  In _instructions.go_, `simplecond` and `twocond` are the commands that make use of the conditions.  The `condition`
  interface defines just one methon, `isMet`, which can inspect an engine and determine if the condition in question
  is met or not. The code in _instructions.go_ does the rest.
   

