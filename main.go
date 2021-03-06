// mdfmt formats Markdown content containing frontmatter.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/scanner"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gohugoio/hugo/parser"
	"github.com/shurcooL/markdownfmt/markdown"
)

var (
	// Main operation modes.
	list   = flag.Bool("l", false, "list files whose formatting differs from mdfmt's")
	write  = flag.Bool("w", false, "write result to (source) file instead of stdout")
	doDiff = flag.Bool("d", false, "display diffs instead of rewriting files")

	exitCode = 0
)

func report(err error) {
	scanner.PrintError(os.Stderr, err)
	exitCode = 2
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: mdfmt [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func isMarkdownFile(f os.FileInfo) bool {
	name := f.Name()
	ext := filepath.Ext(name)
	return !f.IsDir() && !strings.HasPrefix(name, ".") && (ext == "md" || ext == "markdown")
}

func processFile(filename string, in io.ReadSeeker, out io.Writer, stdin bool) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	// slurp in the whole file for comparison later
	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	in.Seek(0, 0)

	// parse the file with hugo/parser to extract front matter
	page, err := parser.ReadFrom(in)
	if err != nil {
		return err
	}

	md, err := markdown.Process(filename, page.Content(), nil)
	if err != nil {
		return err
	}

	// If we have front matter, insert a newline to separate the front matter
	// from the markdown content.
	sep := ""
	if len(page.FrontMatter()) > 0 {
		sep = "\n"
	}

	res := make([]byte, len(page.FrontMatter())+len(sep)+len(md))
	copy(res, append(append(page.FrontMatter(), []byte(sep)...), md...))

	if !bytes.Equal(src, res) {
		// formatting has changed
		if *list {
			fmt.Fprintln(out, filename)
		}
		if *write {
			err = ioutil.WriteFile(filename, res, 0)
			if err != nil {
				return err
			}
		}
		if *doDiff {
			data, err := diff(src, res)
			if err != nil {
				return fmt.Errorf("computing diff: %s", err)
			}
			fmt.Printf("diff %s mdfmt/%s\n", filename, filename)
			out.Write(data)
		}
	}

	if !*list && !*write && !*doDiff {
		_, err = out.Write(res)
	}

	return err
}

func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isMarkdownFile(f) {
		err = processFile(path, nil, os.Stdout, false)
	}
	if err != nil {
		report(err)
	}
	return nil
}

func walkDir(path string) {
	filepath.Walk(path, visitFile)
}

func main() {
	// call mdfmtMain in a separate function
	// so that it can use defer and have them
	// run before the exit.
	mdfmtMain()
	os.Exit(exitCode)
}

func mdfmtMain() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		if err := processFile("<standard input>", os.Stdin, os.Stdout, true); err != nil {
			report(err)
		}
		return
	}

	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			walkDir(path)
		default:
			if err := processFile(path, nil, os.Stdout, false); err != nil {
				report(err)
			}
		}
	}
}

func diff(b1, b2 []byte) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "mdfmt")
	if err != nil {
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "mdfmt")
	if err != nil {
		return
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(b1)
	f2.Write(b2)

	data, err = exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return
}
