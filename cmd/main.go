package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
var theme = "sweet"

func themeDir() string {
	return sourceDir + string(filepath.Separator) +
		"themes" + string(filepath.Separator) +
		theme
}

func headerFile() string {
	return themeDir() + string(filepath.Separator) + "header.html"
}

func footerFile() string {
	return themeDir() + string(filepath.Separator) + "footer.html"
}

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
	err := filepath.WalkDir(
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

	return err
}

// getTargetFile returns the target file name corresponding to fn
// (source file). It will be a filename with a dir prefix.
func (w Woo) getTargetFile(targetDir, fn string) (string, error) {
	s, ok := strings.CutSuffix(filepath.Base(fn), ".md")
	if !ok {
		return "", errors.New("Not an md file")
	}
	return targetDir + string(filepath.Separator) + s + ".html", nil
}

func (w Woo) addHeaderAndFooter(s string, values map[string]string) (string, error) {
	fmt.Printf("Adding header and footer\n")
	f := headerFile()
	var header []byte
	if _, err := os.Stat(f); err == nil {
		header, err = os.ReadFile(f)
		if err != nil {
			return "", err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Warn: %v\n", err)
	}

	f = footerFile()
	var footer []byte
	if _, err := os.Stat(f); err == nil {
		footer, err = os.ReadFile(f)
		if err != nil {
			return "", err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Warn: %v\n", err)
	}
	return expandTemplate(header, values) +
			"\n" + s +
			"\n" +
			expandTemplate(footer, values),
		nil
}

func title(s string) string {
	re := regexp.MustCompile(`title:\s*(.*)`)
	matches := re.FindAllStringSubmatch(s, -1)

	for _, match := range matches {
		if len(match) > 1 {
			// match[0] is the full line matching
			// match[1] is the group, i.e. the value we want
			fmt.Println("Found title:", match[1])
			return match[1]
		}
	}

	return ""
}

// expandTemplate replaces any supported variable inside the (HTML)
// template file with an actual value.
func expandTemplate(template []byte, values map[string]string) string {
	fmt.Printf("TODO dynamic variable map\n")
	key := "TITLE"
	return strings.ReplaceAll(string(template), "{{TITLE}}", values[key])
}

func (w Woo) createTargetDir(fn string) (string, error) {
	dir := getTargetDir(fn)
	fmt.Printf("Generated target dir=%v\n", dir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	fmt.Printf("Created target dir for fn=%v => %v\n", fn, dir)
	return dir, nil
}

func getTargetDir(fn string) string {
	return targetDir + string(filepath.Separator) + filepath.Base(filepath.Dir(fn))
}

type Woo struct {
	Converter  goldmark.Markdown
	DirsCopied []string
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
