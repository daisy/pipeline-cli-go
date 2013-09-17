package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

var files = []struct {
	Name, Body string
}{
	{"readme.txt", "This archive contains some text files."},
	{"fold1/gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
	{"fold1/fold2/todo.txt", "Get animal handling licence.\nWrite more examples."},
}

func createZipFile(t *testing.T) []byte {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add some files to the archive.
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
	}

	// Make sure to check the error on Close.
	err := w.Close()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	return buf.Bytes()
}

func TestDumpFiles(t *testing.T) {
	data := createZipFile(t)
	folder := filepath.Join(os.TempDir(), "pipeline_commands_test")
	err := os.MkdirAll(folder, 0755)
	visited := make(map[string]bool)
	for _, f := range files {
		visited[filepath.Clean(f.Name)] = false
	}

	defer func() {
		os.RemoveAll(folder)
	}()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	err = dumpZippedData(data, folder)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	filepath.Walk(folder, func(path string, inf os.FileInfo, err error) error {
		entry, err := filepath.Rel(folder, path)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		visited[entry] = true
		return nil
	})
	for _, f := range files {
		if !visited[filepath.Clean(f.Name)] {
			t.Errorf("%v was not visited", filepath.Clean(f.Name))
		}
	}

}
