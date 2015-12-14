mdfmt
=====

Like `gofmt`, but for Markdown with front matter.

`mdfmt` is a fork of [markdownfmt](https://github.com/shurchooL/markdownfmt) that simply adds support for TOML, YAML, and JSON front matter.

Usage
-----

```
usage: mdfmt [flags] [path ...]
  -d    display diffs instead of rewriting files
  -l    list files whose formatting differs from mdfmt's
  -w    write result to (source) file instead of stdout
```

Acknowledgements
----------------

-	[markdownfmt](https://github.com/shurchooL/markdownfmt) for the actual markdown formatting.
-	[Hugo](https://gohugo.io) for front matter parsing.
-	[Go Authors](https://golang.org) for writing `gofmt`.

License
-------

Released under the MIT License. See LICENSE file for details.
