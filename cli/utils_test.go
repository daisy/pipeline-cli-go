package cli

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

//Creates a new zip file to test the dump function
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

//Test the zip dumping functionality
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

//Test that the key is correctly loaded
func TestLoadKey(t *testing.T) {
	backup := keyFile
	defer func() {
		keyFile = backup
	}()
	keyFile = "fakeKey"
	expected := "dondeestanlasllavesmatarile"
	path := filepath.Join(os.TempDir(), keyFile)
	file, err := os.Create(path)
	defer os.Remove(file.Name())
	if err != nil {
		t.Errorf("Unexpected error opening file%v", err)
	}
	file.Write([]byte(expected))
	if file.Close() != nil {
		t.Errorf("Unexpected error closing file%v", err)
	}
	key, err := loadKey()
	if err != nil {
		t.Errorf("Unexpected error loading key %v", err)
	}
	if expected != key {
		t.Errorf("The stored key doesn't correspond with the loaded key '%s'!='%s'", expected, key)
	}
}

//Test that the error is propagated if the file doesn't exist
func TestLoadKeyOpenError(t *testing.T) {
	backup := keyFile
	defer func() {
		keyFile = backup
	}()
	keyFile = "thiskeyfiledoensntexist"
	_, err := loadKey()
	if err == nil {
		t.Errorf("Expected error loading key didn't occur", err)
	}
}
