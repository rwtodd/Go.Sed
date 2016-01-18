# Go-Sed 

An implementation of sed in Go.  Just because!


## Implementation Notes

To be written.

## Status

It is early in the development.  As of now, the instructions to run are hard-coded for `sed -n`:

    $ d 
    8,11 {
       =
       p
    }
    /import/, /^\)/ d
    p

... which doesn't look impressive, but it illustrates that the basic engine
is working.  I also support inverted conditions, like:

   8,11 !d


## Next Steps

The most important missing piece is a lexer/parser for the sed commands. Right now,
the test program is hard-coded in main.go.

After that, I'll need to add the missing commands (importantly, s// isn't implemented yet!).


## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


