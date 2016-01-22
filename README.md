# Go-Sed 

An implementation of sed in Go.  Just because!


## Implementation Notes

I have never looked at how a "real" implementation of sed is done. I'm just
going by the sed man pages and tutorials.  I will describe this implementation
in this space soon.  To be written.

## Status

It is early in the development.  There is a basic engine in place, which can run
a few of the commands.  The lexer is almost complete (missing the 'y' command).
The parser is just now under development. It can only parse a few commands. That's
what I'm working on now.


## Next Steps

Finish the parser/compiler from tokens to instructions for the execution
engine.

After that, I'll need to add code to run the missing commands.


## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


