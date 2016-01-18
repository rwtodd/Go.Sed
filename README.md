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
is working.


## Next Steps

Next I'll implement inverted conditions, like: 

    /^NODEL/ !d

After I have those working, I'll start working on the lexer/parser. Until then, the 
`main` funciton will just hard-code the sed-program. 


## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


