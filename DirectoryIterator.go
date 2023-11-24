package main


import (
	"os"
	"path/filepath"
	"sort"
	"io"
	"fmt"
)



type DirectoryIterator struct {
	file *os.File
}

func (d *DirectoryIterator) Next() error {
	currentFileInfo, err := d.file.Stat()
	if err != nil {
		return err
	}
	currentModTime := currentFileInfo.ModTime()
	fmt.Println("Current file", d.file.Name())
	directory, err := os.ReadDir(filepath.Dir(d.file.Name()))
	if err != nil {
		return err
	}

	var relevantFiles []os.FileInfo
	for _, dirEntry := range directory {
		fileInfo, err := dirEntry.Info()
		if err != nil {
			return err
		}

		if !dirEntry.IsDir() && filepath.Ext(dirEntry.Name()) == ".sst" {
			relevantFiles = append(relevantFiles, fileInfo)
		}
	}

	sort.Slice(relevantFiles, func(i, j int) bool {
		return relevantFiles[i].ModTime().Before(relevantFiles[j].ModTime())
	})

	var prevFile os.FileInfo
	for _, file := range relevantFiles {
		if file.ModTime().Before(currentModTime) {
			prevFile = file
		}
	}

	if prevFile != nil {
		d.file, err = os.Open(filepath.Join(filepath.Dir(d.file.Name()), prevFile.Name()))
		if err != nil {
			return err
		}
	}else{
		return io.EOF
	}
	return nil
}

