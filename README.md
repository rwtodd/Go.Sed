# Go-Sed 

An implementation of sed in Go.  Just because!


## Implementation Notes

I have never looked at how a "real" implementation of sed is done. I'm just
going by the sed man pages and tutorials.  I will describe this implementation
in this space soon.  To be written.

## Status

It is early in the development.  There is a basic engine in place, which can run
a few of the commands.  The lexer is almost complete (missing the 'y' command).
The parser is pretty much working now.


## Next Steps

Add code to run the missing commands (notably I need to code the 's'ubstitution command).

Also, right now it looks for a hard-coded 'program.sed', and doesn't yet support options
like '-e' or '-n'.  I'll need to add those soon.

## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


