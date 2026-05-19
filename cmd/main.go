package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func (w Woo) file2html(sourceFile string) error {
	fmt.Printf("fn=%v\n", sourceFile)

	var source []byte

	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	// 3.1 Create target targetDir based on the title meta field
	targetDir, err := w.createTargetDir(sourceFile)
	if err != nil {
		return err
	}
	targetFile, err := w.getTargetFile(targetDir, sourceFile)
	if err != nil {
		return err
	}

	// 3.2 Convert to HTML and write to target dir
	var buf bytes.Buffer
	if err := w.Converter.Convert(source, &buf); err != nil {
		panic(err)
	}

	// 4.  Insert header and footer from template dir
	values := map[string]string{
		"TITLE": title(string(source)),
	}
	s, err := w.addHeaderAndFooter(buf.String(), values)
	if err != nil {
		panic(err)
	}

	// 5.  Copy any image from the directory of the markdown file
	//     to the corresponding target dir.
	w.copyFilesToTargetDir(sourceFile, targetDir)

	if err = os.WriteFile(targetFile, []byte(s), 0644); err != nil {
		return err
	}

	fmt.Printf("Wrote: %v\n", targetFile)
	return nil
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

// copyFilesToTargetDir copies non-markdown files in the same
// directory as fn (fn, being the full path to a Markdown file) to
// target dir.
func (w Woo) copyFilesToTargetDir(srcFileName string, targetDir string) error {
	srcDir := filepath.Dir(srcFileName)
	dfi, derr := os.Stat(targetDir)
	if derr != nil {
		return derr
	}
	err := filepath.WalkDir(
		srcDir,
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			sfi, err := os.Stat(path)
			if err != nil {
				return err
			}
			if !sfi.Mode().Perm().IsRegular() {
				fmt.Printf("Not a regular file: %v\n", sfi)
				return nil
			}
			if os.SameFile(dfi, sfi) {
				return nil
			}
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			targetFileName := targetDir +
				string(filepath.Separator) +
				filepath.Base(path)

			dstFile, err := os.Create(targetFileName)
			defer func() error {
				if err := dstFile.Close(); err != nil {
					return err
				}
				return nil
			}()

			if err != nil {
				return err
			}

			bytes, err := io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}
			err = dstFile.Sync()
			if err != nil {
				return err
			}

			fmt.Printf(
				"Copied %v to %v, %v bytes in total\n",
				srcFile.Name(),
				dstFile.Name(),
				bytes)

			return nil
		},
	)

	fmt.Printf("TODO Copy all files in %v to %v\n", srcDir, targetDir)
	return err
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
