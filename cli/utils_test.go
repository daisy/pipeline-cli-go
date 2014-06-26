package cli

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"os/user"
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

//Creates a fake key file
func createKeyFile(keyFile, key string) (file *os.File, err error) {
	path := filepath.Join(os.TempDir(), keyFile)
	file, err = os.Create(path)
	if err != nil {
		return
	}
	_, err = file.Write([]byte(key))
	if err != nil {
		return
	}
	if file.Close() != nil {
		return
	}
	return
}

//Test that the key is correctly loaded
func TestLoadKey(t *testing.T) {
	backup := keyFile
	defer func() {
		keyFile = backup
	}()
	expected := "dondeestanlasllavesmatarile"
	keyFile = "fakeKey"
	file, err := createKeyFile(keyFile, expected)
	defer os.Remove(file.Name())
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

//Test that check priority recognises the allowed priorities
func TestCheckPriorityOk(t *testing.T) {
	if !checkPriority("high") {
		t.Errorf("high wasn't recognised as priority")
	}
	if !checkPriority("medium") {
		t.Errorf("medium wasn't recognised as priority")
	}
	if !checkPriority("low") {
		t.Errorf("low wasn't recognised as priority")
	}
}

//Test that check priority discards non-allowed values
func TestCheckPriorityNotOk(t *testing.T) {
	if checkPriority("asdfasdf") {
		t.Errorf("non-recognised value passed checkPriority")
	}
}

func TestGetLastId(t *testing.T) {
	oldSep := pathSeparator
	oldHome := homePath
	homePath = "home"
	defer func() {
		pathSeparator = oldSep
		homePath = oldHome
	}()
	//for linux
	pathSeparator = '/'
	path := getLastIdPath("linux")
	if "home/.daisy-pipeline/dp2/lastid" != path {
		t.Errorf("Lastid path for linux is wrong %v", path)
	}

	//for windows
	os.Setenv("APPDATA", "windows")
	path = getLastIdPath("windows")
	pathSeparator = '\\'
	if path != "windows\\DAISY Pipeline 2\\dp2\\lastid" {
		t.Errorf("Lastid path for windows is wrong %v", path)
	}
	//for darwin
	pathSeparator = '/'
	path = getLastIdPath("darwin")
	if "home/Library/Application Support/DAISY Pipeline 2/dp2/lastid" != path {
		t.Errorf("Lastid path for darwin is wrong %v", path)
	}
}

func TestUnknownOs(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Expecting panic didn't happend")
		}
	}()
	getLastIdPath("myos")

}

func TestMustUserError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Expecting panic didn't happend")
		}
	}()

	mustUser(nil, fmt.Errorf("erroring"))

}

func TestMustUser(t *testing.T) {
	usr := &user.User{}
	res := mustUser(usr, nil)
	if usr != res {
		t.Errorf("Different users")
	}
}
