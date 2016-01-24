# Go-Sed 

An implementation of sed in Go.  Just because!


## Implementation Notes

I have never looked at how a "real" implementation of sed is done. I'm just
going by the sed man pages and tutorials.  I will describe this implementation
in this space soon.  To be written.

## Status

  * __Command-Line processing__:  Done. It accepts '-e', '-f', '-n' and long
versions of the same. It takes '-help'.
  * __Lexer__: Complete.
  * __Parser/Engine__:  Has the common commands and conditions, but some are
still missing.  I will be filling these in soon.


## Next Steps

Add missing commands to the parser and execution engine.


## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


