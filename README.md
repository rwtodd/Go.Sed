# Go-Sed 

An implementation of sed in Go.  Just because!

## Status

  * __Command-Line processing__:  Done. It accepts '-e', '-f', '-n' and long
versions of the same. It takes '-help'.
  * __Lexer__: Complete.
  * __Parser/Engine__:  Has every command in a typical sed now. 
 It has:  a\, i\, c\, d, D, p, P, g, G, x, h, H, r, w, s, y, b, t, :label, n, N, q, =.

## Differences from Standard Sed

__Regexps__: The only thing you really have to keep in mind when using 
go-sed, is I use Go's "regexp" package. Therefore, you have to use that 
syntax for the regular expressions.  In particular, the biggest difference 
between go-sed and a typical sed with extended regexps, is that replacements 
are on `$1` and `$2` instead of `\1` and `\2`:

    /trigger/ {
        s/a(bc*)d/$1/g
    }

There are a few niceties though, such as I interpret '\t' and '\n' in 
replacement strings:

    s/\w+\s*/\t$0\n/g

You can also escape the newline like in a typical sed, if you want.

__Loser Syntax__: Go-sed is a little more user-friendly when it comes to
syntax.  In a normal sed, you have to use one (and ONLY one)
space between a `r` or `w` and the filename. Go-sed eats whitespace until it
sees the filename.

Also, in a typical sed, you need seemingly-extraneous semicolons like the one after the `d` below: 

    sed -e '/re/ { p ; d; }' < in > out

... but go-sed is much nicer about it:

    go-sed -e '/re/ { p ; d }' < in > out 


## Implementation Notes

I have never looked at how a "real" implementation of sed is done. I'm just
going by the sed man pages and tutorials.  I will describe this implementation
in this space soon.  To be written.

I will note that in speed comparisons, go-sed outperforms Mac OS X's sed on my
iMac, as long as the input isn't tiny.  That's despite the fact that I work 
on strings internally... eating the cost of conversions from []byte on input.

The way the script is compiled to an array of closures makes the inner 
loop of the interpreter very compact:

    for err == nil {
       err = e.ins[e.ip](e)
    }


## Next Steps

The program is done now, barring bug fixes!  Next I want to document
the design a little.


## Go Get

You can get the code/executable by saying:

    go get github.com/waywardcode/go-sed


