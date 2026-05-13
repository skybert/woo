package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var Version = "dev"
var showVersion bool
var sourceDir = "/tmp/src"
var targetDir = "/tmp/target"

func init() {
	pflag.BoolVarP(&showVersion, "version", "v", false, "Show version")
	pflag.StringVar(&sourceDir, "src-dir", "/tmp/src", "Source directory")
	pflag.StringVar(&targetDir, "target-dir", "/tmp/output", "Target directory")
}

func main() {
	pflag.Parse()
	if showVersion {
		fmt.Printf("woo version: %v\n", Version)
		os.Exit(0)
	}
	w := NewWoo()
	err := w.dir2html(sourceDir)
	if err != nil {
		panic(err)
	}
}

func (w Woo) dir2html(sourceDir string) error {
	// 1.  read files in source dir
	// 2.  iterate through these
	filepath.WalkDir(
		sourceDir,
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			if !strings.HasSuffix(path, ".md") {
				return nil
			}

			fmt.Printf("f=%v, d=%v err=%v\n", path, d, err)
			convErr := w.file2html(path)
			if convErr != nil {
				return convErr
			}

			return nil
		})
	return nil
}

func (w Woo) file2html(fn string) error {
	fmt.Printf("fn=%v\n", fn)

	var source []byte

	source, err := os.ReadFile(fn)
	if err != nil {
		return nil
	}

	// 3.1 Create target targetDir based on the title meta field
	targetDir, err := w.createTargetDir(fn)
	targetFile, err := w.createTargetFile(fn)
	if err != nil {
		panic(err)
	}

	// 3.2 Convert to HTML and write to target dir
	var buf bytes.Buffer
	if err := w.Converter.Convert(source, &buf); err != nil {
		panic(err)
	}
	// 4.  Insert header and footer from template dir
	s, err := w.addHeaderAndFooter(buf.String())
	if err != nil {
		panic(err)
	}

	// 5.  Copy any image from the directory of the markdown file
	//     to the corresponding target dir.
	w.copyFilesToTargetDir(fn, targetDir)
	fmt.Printf("Writing: %v\n", targetFile)

	if err = os.WriteFile(targetFile, []byte(s), 0644); err != nil {
		panic(err)
	}

	return nil
}

// createTargetFile returns the target file name corresponding to fn
// (source file). It will be a filename without a dir prefix.
func (w Woo) createTargetFile(fn string) (string, error) {
	s, ok := strings.CutSuffix(filepath.Base(fn), ".md")
	if !ok {
		return "", errors.New("Not an md file")
	}
	return s + ".html", nil
}

// copyFilesToTargetDir copies non-markdown files in the same
// directory as fn (fn, being the full path to a Markdown file) to
// target dir.
func (w Woo) copyFilesToTargetDir(fn string, dir string) {
	fmt.Printf("TODO Copy all files in %v to %v\n", fn, dir)
}

func (w Woo) addHeaderAndFooter(s string) (string, error) {
	fmt.Printf("TODO add header and footer\n")
	return s, nil

}

func (w Woo) createTargetDir(fn string) (dir string, err error) {
	fmt.Printf("TODO Create target dir for fn=%v\n", fn)

	return targetDir + filepath.Base(filepath.Dir(fn)), nil
}

type Woo struct {
	Converter goldmark.Markdown
}

func NewWoo() Woo {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
		),
		goldmark.WithExtensions(
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols([]string{
					"http:",
					"https:",
				}),
			),
		),
	)

	return Woo{
		Converter: markdown,
	}
}
