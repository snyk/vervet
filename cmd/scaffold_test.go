package cmd_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/snyk/vervet/v3/cmd"
	"github.com/snyk/vervet/v3/testdata"
)

var vervetConfigFile = "./.vervet.yaml"

type testPrompt struct {
	ReturnConfirm bool
	ReturnSelect  string
	ReturnEntry   string
}

func (tp *testPrompt) Confirm(label string) (bool, error) {
	return tp.ReturnConfirm, nil
}

func (tp *testPrompt) Entry(label string) (string, error) {
	return tp.ReturnEntry, nil
}

func (tp *testPrompt) Select(label string, items []string) (string, error) {
	return tp.ReturnSelect, nil
}

var filemark = "bad wolf"

// markFile adds a string to a file so we can check if that file is being overwritten.
func markTestFile(filename string) error {
	// Write a string to the file; we should see this string removed when
	f, err := os.OpenFile(filename, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err != nil {
		return err
	}
	_, err = f.Write([]byte(filemark))
	return err
}

// markInFile checks if the filemark is present, determining if the file has been
// overwritten.
func markInFile(filename string) (bool, error) {
	content, err := ioutil.ReadFile(vervetConfigFile)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), "bad wolf"), nil
}

func TestScaffold(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	cd(c, dstDir)

	prompt := testPrompt{}
	testApp := cmd.NewApp(cmd.VervetParams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Prompt: &prompt,
	})

	// Running init creates the project files.
	err := testApp.Run([]string{"vervet", "scaffold", "init", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)

	// Rerunning init asks the user if they want to overwrite; if they say no
	// the command ends...
	prompt.ReturnConfirm = false
	err = markTestFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	err = testApp.Run([]string{"vervet", "scaffold", "init", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)
	fileMarked, err := markInFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	c.Assert(fileMarked, qt.IsTrue)

	// ...if the user selects yes, it will overwrite the project files.
	prompt.ReturnConfirm = true
	err = testApp.Run([]string{"vervet", "scaffold", "init", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)
	fileMarked, err = markInFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	c.Assert(fileMarked, qt.IsFalse)

	// Rerunning init with the force option overwrites the project files.
	prompt.ReturnConfirm = false
	err = markTestFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	err = testApp.Run([]string{"vervet", "scaffold", "init", "--force", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)
	fileMarked, err = markInFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	c.Assert(fileMarked, qt.IsFalse)

	// A new resource version can be generated in the project after initialization has completed.
	err = testApp.Run([]string{"vervet", "version", "new", "--version", "2021-10-01", "v3", "foo"})
	c.Assert(err, qt.IsNil)
	for _, item := range []string{".vervet/templates/README.tmpl", ".vervet.yaml", ".vervet/extras/foo", ".vervet/extras/bar/bar"} {
		_, err = os.Stat(item)
		c.Assert(err, qt.IsNil)
	}
	readme, err := ioutil.ReadFile("v3/resources/foo/2021-10-01/README")
	c.Assert(err, qt.IsNil)
	c.Assert(string(readme), qt.Equals, `
This is a generated scaffold for version 2021-10-01~wip of the
foo resource in API v3.

`[1:])
}
