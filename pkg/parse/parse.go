package parse

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/gohugoio/hugo/parser"
	"github.com/shurcooL/markdownfmt/markdown"
)

func ParseFile(filename string, in io.ReadSeeker) ([]byte, []byte, error) {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return nil, nil, err
		}
		defer f.Close()
		in = f
	}

	// slurp in the whole file for comparison later
	src, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, nil, err
	}
	in.Seek(0, 0)

	// parse the file with hugo/parser to extract front matter
	page, err := parser.ReadFrom(in)
	if err != nil {
		return nil, nil, err
	}

	md, err := markdown.Process(filename, page.Content(), nil)
	if err != nil {
		return nil, nil, err
	}

	// If we have front matter, insert a newline to separate the front matter
	// from the markdown content.
	sep := ""
	if len(page.FrontMatter()) > 0 {
		sep = "\n"
	}

	res := make([]byte, len(page.FrontMatter())+len(sep)+len(md))
	copy(res, append(append(page.FrontMatter(), []byte(sep)...), md...))

	return src, res, err
}
