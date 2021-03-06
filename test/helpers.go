package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

// OutputStub implements a simple utils.Runnable
type OutputStub struct {
	Out []byte
}

func (s OutputStub) Output() ([]byte, error) {
	return s.Out, nil
}

func (s OutputStub) Run() error {
	return nil
}

type TempGitRepo struct {
	Remote   string
	TearDown func()
}

func UseTempGitRepo() *TempGitRepo {
	pwd, _ := os.Getwd()
	oldEnv := make(map[string]string)
	overrideEnv := func(name, value string) {
		oldEnv[name] = os.Getenv(name)
		os.Setenv(name, value)
	}

	remotePath := filepath.Join(pwd, "..", "test", "fixtures", "test.git")
	home, err := ioutil.TempDir("", "test-repo")
	if err != nil {
		panic(err)
	}

	overrideEnv("HOME", home)
	overrideEnv("XDG_CONFIG_HOME", "")
	overrideEnv("XDG_CONFIG_DIRS", "")

	targetPath := filepath.Join(home, "test.git")
	cmd := exec.Command("git", "clone", remotePath, targetPath)
	if output, err := cmd.Output(); err != nil {
		panic(fmt.Errorf("error running %s\n%s\n%s", cmd, err, output))
	}

	if err = os.Chdir(targetPath); err != nil {
		panic(err)
	}

	// Our libs expect the origin to be a github url
	cmd = exec.Command("git", "remote", "set-url", "origin", "https://github.com/github/FAKE-GITHUB-REPO-NAME")
	if output, err := cmd.Output(); err != nil {
		panic(fmt.Errorf("error running %s\n%s\n%s", cmd, err, output))
	}

	tearDown := func() {
		if err := os.Chdir(pwd); err != nil {
			panic(err)
		}
		for name, value := range oldEnv {
			os.Setenv(name, value)
		}
		if err = os.RemoveAll(home); err != nil {
			panic(err)
		}
	}

	return &TempGitRepo{Remote: remotePath, TearDown: tearDown}
}

func ExpectLines(t *testing.T, output string, lines ...string) {
	var r *regexp.Regexp
	for _, l := range lines {
		r = regexp.MustCompile(l)
		if !r.MatchString(output) {
			t.Errorf("output did not match regexp /%s/\n> output\n%s\n", r, output)
			return
		}
	}
}
