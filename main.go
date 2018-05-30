package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const (
	prefix            = "[rewrite-args] "
	defaultConfigFile = "~/.rewrite-args.conf"
)

const usage = `
# If the following is true
alias ssh='rewrite-args ssh -X'

# and ~/.rewrite-args.conf contains
{
  "debug": false,
  "rewrites": [
    {
      "match": ".use1",
      "replace": ".prod.us-east-1.postgun.com"
    }
  ]
}

# Given the following command
ssh worker-n01.use1 

# Will expand too
/usr/bin/ssh -X worker-n01.prod.us-east-1.postgun.com
`

type Replacement struct {
	Match    string
	Compiled *regexp.Regexp
	Replace  string
}

type Config struct {
	Rewrites []Replacement
	Debug    bool
}

func main() {
	// Look for our config file
	conf, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, prefix+"%s\n", err)
		os.Exit(1)
	}

	// look for matches in os.Args and replace
	args := os.Args
	for i, arg := range args {
		var replaced string
		for _, item := range conf.Rewrites {
			replaced = item.Compiled.ReplaceAllString(arg, item.Replace)
			if replaced != arg {
				args[i] = replaced
			}
		}
	}

	if len(args) == 1 {
		fmt.Println(usage)

		fmt.Printf("\n%s\n", strings.Join(args, " "))
		os.Exit(1)
	}

	// Find the requested executable in the path
	path, err := exec.LookPath(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: command not found", args[1])
		os.Exit(1)
	}

	if conf.Debug {
		fmt.Printf(prefix+"exec: [%s] %s\n", path, args[1:])
	}

	// os.Exec with replaced arguments
	err = syscall.Exec(path, args[1:], os.Environ())
	if err != nil {
		fmt.Fprintf(os.Stderr, prefix+"exec: %s\n", err)
		os.Exit(1)
	}
}

func loadConfig() (*Config, error) {
	configFile := expandTilde(defaultConfigFile)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, errors.Errorf("config '%s' missing....\n"+usage, configFile)
	}

	fd, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, errors.Errorf("while reading config '%s' - %s", configFile, err)
	}

	var conf Config
	err = json.Unmarshal(contents, &conf)
	if err != nil {
		return nil, errors.Errorf("while un-marshalling config '%s' - %s", configFile, err)
	}

	for i, item := range conf.Rewrites {
		conf.Rewrites[i].Compiled, err = regexp.Compile(item.Match)
		if err != nil {
			return nil, errors.Errorf("failed to compile regex '%s' - %s", item.Match, err)
		}
	}
	return &conf, nil
}

func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	findHome := func() string {
		switch runtime.GOOS {
		case "windows":
			home := filepath.Join(os.Getenv("HomeDrive"), os.Getenv("HomePath"))
			if home == "" {
				home = os.Getenv("UserProfile")
			}
			return home

		default:
			return os.Getenv("HOME")
		}
	}

	home := findHome()
	if home == "" {
		home = "~"
	}

	return strings.Replace(path, "~", home, -1)
}

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}

// A modified version of exec.LookPath which ignores our executable if found on the path
//
// We do this because if we have `rewrite-args` linked as ~/bin/ssh
// and it's the first thing in our path then `LookPath()` will find it first
// and infinite loop death spiral will begin
func LookPath(file, ignore string) (string, error) {
	if strings.Contains(file, "/") {
		err := findExecutable(file)
		if err == nil {
			return file, nil
		}
		return "", &exec.Error{file, err}
	}

	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := filepath.Join(dir, file)
		if err := findExecutable(path); err == nil {
			if path != ignore {
				return path, nil
			}
		}
	}
	return "", &exec.Error{file, exec.ErrNotFound}
}
