package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// copyFilesToTargetDir copies non-markdown files in the same
// directory as fn (fn, being the full path to a Markdown file) to
// target dir.
func (w *Woo) copyFilesToTargetDir(srcFileName string, targetDir string) error {
	srcDir := filepath.Dir(srcFileName)
	fmt.Printf("dirs copied before: %v\n", w.DirsCopied)
	if w.hasCopiedFilesFor(srcDir) {
		fmt.Printf("Already copied files from dir: %v\n", srcDir)
		return nil
	} else {
		w.DirsCopied = append(w.DirsCopied, srcDir)
	}
	fmt.Printf("dirs copied after: %v\n", w.DirsCopied)

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

	fmt.Printf("Copied all files in %v to %v\n", srcDir, targetDir)
	return err
}

func (w *Woo) hasCopiedFilesFor(srcFileName string) bool {
	return slices.Contains(w.DirsCopied, srcFileName)
}
