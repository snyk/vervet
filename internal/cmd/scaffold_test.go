package cmd_test

import (
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/urfave/cli/v2"

	"github.com/snyk/vervet/v6/internal/cmd"
	"github.com/snyk/vervet/v6/testdata"
)

var vervetConfigFile = "./.vervet.yaml"

func appWithSubcommand(subcmd *cli.Command, prompt cmd.VervetPrompt) *cmd.VervetApp {
	newCli := cmd.CLIApp
	newCli.Commands = append(newCli.Commands, subcmd)
	return cmd.NewApp(&newCli, cmd.VervetParams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Prompt: prompt,
	})
}

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
	content, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), "bad wolf"), nil
}

func TestScaffold(t *testing.T) {
	c := qt.New(t)
	dstDir := c.TempDir()
	cd(c, dstDir)

	prompt := &testPrompt{}
	testScaffoldCmd := appWithSubcommand(&cmd.Scaffold, prompt)

	// Running init creates the project files.
	err := testScaffoldCmd.Run([]string{"vervet", "scaffold", "init", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)

	// Rerunning init asks the user if they want to overwrite; if they say no
	// the command ends...
	prompt.ReturnConfirm = false
	err = markTestFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	err = testScaffoldCmd.Run([]string{"vervet", "scaffold", "init", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)
	fileMarked, err := markInFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	c.Assert(fileMarked, qt.IsTrue)

	// ...if the user selects yes, it will overwrite the project files.
	prompt.ReturnConfirm = true
	err = testScaffoldCmd.Run([]string{"vervet", "scaffold", "init", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)
	fileMarked, err = markInFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	c.Assert(fileMarked, qt.IsFalse)

	// Rerunning init with the force option overwrites the project files.
	prompt.ReturnConfirm = false
	err = markTestFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	err = testScaffoldCmd.Run([]string{"vervet", "scaffold", "init", "--force", testdata.Path("test-scaffold")})
	c.Assert(err, qt.IsNil)
	fileMarked, err = markInFile(vervetConfigFile)
	c.Assert(err, qt.IsNil)
	c.Assert(fileMarked, qt.IsFalse)
}
