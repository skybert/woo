package main

import (
	"bytes"
	"fmt"
	"os"
)

func (w *Woo) file2html(sourceFile string) error {
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
