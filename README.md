# Go-Sed 

An implementation of sed in Go.  Just because!

## Status

  * __Command-Line processing__:  Done. It accepts '-e', '-f', '-n' and long
versions of the same. It takes '-help'.
  * __Lexer__: Complete.
  * __Parser/Engine__:  Has every command in a typical sed except 'w' now. 
 It has:  a\, i\, c\, d, D, p, P, g, G, x, h, H, r, s, y, b, t, :label, n, N, q, =.

The only thing you really have to keep in mind when using it, is I use
Go's "regexp" package. Therefore, you have to use that syntax for the
regular expressions.  In particular, the biggest difference between 
go-sed and a typical sed with extended regexps, is that replacements 
are on `$1` and `$2` instead of `\1` and `\2`:

    /trigger/ {
        s/a(bc*)d/$1/g
    }


## Implementation Notes

I have never looked at how a "real" implementation of sed is done. I'm just
going by the sed man pages and tutorials.  I will describe this implementation
in this space soon.  To be written.

I will note that in speed comparisons, go-sed outperforms Mac OS X's sed on my
iMac, as long as the input isn't tiny.  That's despite the fact that I work 
on strings internally... eating the cost of conversions from []byte on input.

## Next Steps

Add the final missing command ('w') to the parser and execution engine.  Document
the design a little.


## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


