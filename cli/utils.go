package cli

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/capitancambio/go-subcommand"
)

var keyFile = "dp2key.txt"

//testing multienv is a pain
var pathSeparator = os.PathSeparator
var homePath = mustUser(user.Current()).HomeDir

//Filter the user error and panics if the error is present
func mustUser(user *user.User, err error) *user.User {
	if err != nil {
		panic("Current user not found")
	}
	return user
}

//Checks that a string defines a priority value
func checkPriority(priority string) bool {

	return priority == "high" || priority == "medium" ||
		priority == "low"

}

//loads the halt key
func loadKey() (key string, err error) {
	//get temp dir
	path := filepath.Join(os.TempDir(), keyFile)
	file, err := os.Open(path)
	if err != nil {
		errors.New("Could not find the key file, is the webservice running in this machine?")
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	key = string(bytes)
	return
}

//Checks if the job id is present when the command was called
func checkId(lastId bool, command string, args ...string) (id string, err error) {
	if len(args) != 1 && !lastId {
		return id, fmt.Errorf("Command %v needs a job id", command)
	}
	//got it from file
	if lastId {
		id, err = getLastId()
		return
	} else {
		//first arg otherwise
		id = args[0]
		return
	}
}

//Adds the last id switch to the command
func addLastId(cmd *subcommand.Command, lastId *bool) {
	cmd.AddSwitch("lastid", "l", "Get id from the last executed job instead of JOB_ID", func(string, string) error {
		*lastId = true
		return nil
	})
	cmd.SetArity(-1, "[JOB_ID]")
}

//Calculates the absolute path in base of cwd and creates the directory
func createAbsoluteFolder(folder string) (absPath string, err error) {
	absPath, err = filepath.Abs(folder)
	if err != nil {
		return
	}
	return absPath, mkdir(absPath)
}

//mkdir -p
func mkdir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	return nil
}

//Writes the  data to the specified folder
func dumpZippedData(data []byte, folder string) error {
	buff := bytes.NewReader(data)
	reader, err := zip.NewReader(buff, int64(len(data)))
	if err != nil {
		return err
	}
	// Iterate through the files in the archive,
	//and store the results
	for _, f := range reader.File {
		//Get the path of the new file
		path := filepath.Join(folder, filepath.Clean(f.Name))
		if err := mkdir(filepath.Dir(path)); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		dest, err := os.Create(path)
		if err != nil {
			return err
		}

		if _, err = io.Copy(dest, rc); err != nil {
			return err
		}

		if err := dest.Close(); err != nil {
			return err
		}

		if err := rc.Close(); err != nil {
			return err
		}

	}
	return nil
}

//Creates the folder and dumps the zippped data
func zippedDataToFolder(data []byte, folder string) (absPath string, err error) {
	//Create folder
	absPath, err = createAbsoluteFolder(folder)
	filepath.Abs(folder)
	err = dumpZippedData(data, absPath)
	return
}

//Creates the folder and dumps the zippped data
func zippedDataToFile(data []byte, file string) (absPath string, err error) {
	//Create folder
	absPath, err = filepath.Abs(file)
	if err != nil {
		return
	}
	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer func() {
		f.Close()
	}()
	if _, err = f.Write(data); err != nil {
		return absPath, err
	}
	return
}

//gets the path for last id file
func getLastIdPath(currentOs string) string {
	var path string
	switch currentOs {
	case "linux":
		path = homePath + "/.daisy-pipeline/dp2/lastid"
	case "windows":
		path = os.Getenv("APPDATA") + "\\DAISY Pipeline 2\\dp2\\lastid"
	case "darwin":
		path = homePath + "/Library/Application Support/DAISY Pipeline 2/dp2/lastid"
	default:
		panic(fmt.Sprintf("Platform not recognised %v", currentOs))
	}
	return path
}
