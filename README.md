# Go-Sed 

An implementation of sed in Go.  Just because!


## Implementation Notes

To be written.

## Status

It is early in the development.  As of now, the instructions to run are hard-coded for `sed -n`:

    =
    p

... which doesn't look impressive, but it illustrates that the basic engine
is working, and includes under-the-covers functionality which amounts to the
`n` and `d` commands.


## Next Steps

Next I'll implement conditional guards for the commands, like:

    1,10 d

After I have those working, I'll start working on the lexer/parser. Until then, the 
`main` funciton will just hard-code the sed-program. 


## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


