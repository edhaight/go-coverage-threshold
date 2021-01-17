package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/edhaight/go-coverage-threshold/pkg/cover"
)

const (
	thresholdDefault = 80.0
	thresholdUsage   = "threshold that coverage must exceed"
)

var (
	threshold float64
	profile   bool
	packages  Packages
)

// Packages is a custom type for passing packages on the command
// line to test.
type Packages []string

// String implements the flag.Value interface's String method.
func (p *Packages) String() string {
	return strings.Join(*p, " ")
}

// Set implements the flag.Value interface's Set method.
func (p *Packages) Set(s string) error {
	if s == "" {
		return nil
	}
	*p = strings.Split(s, " ")
	return nil
}

func config(s string) *cover.Config {
	if len(os.Args) >= 2 && threshold != thresholdDefault {
		// user specified -t or -threshold
		// args take precedence over .cover.toml files
		return &cover.Config{
			Threshold: threshold,
		}
	}

	config, err := cover.Load(s)
	if err != nil {
		fmt.Printf("no arguments specified, and unable to load .cover.toml file: %v\n", err)
		return &cover.Config{
			Threshold: thresholdDefault,
		}
	}
	return config
}

func flags() {
	flag.Float64Var(&threshold, "threshold", thresholdDefault, thresholdUsage)
	flag.Float64Var(&threshold, "t", thresholdDefault, thresholdUsage+" (shorthand)")
	flag.BoolVar(&profile, "profile", false, "to generate profile file cover.out in current directory")
	flag.Var(&packages, "packages", "space seperated list of packages to test")
	flag.Parse()
}

func goPath() (string, string, error) {
	gopath, ok := os.LookupEnv("GOPATH")
	if ok {
		return gopath, "", nil
	}

	// check if the project uses go modules by running go mod why
	// for a project that does not use go modules this command will fail
	cmd := exec.Command("go", "list", "-m") // nolint: gas,gosec
	if out, err := cmd.CombinedOutput(); err == nil {
		// go modules are active - use current working directory
		moduleName := strings.TrimSpace(string(out))

		// look for working directory - this is non optional
		// panic if no PWD is not set
		if pwd, err := os.Getwd(); err == nil {
			return pwd, moduleName, nil
		}
		log.Fatalf("PWD is not set in ENV, when using modules it is necessary to know current working directory in order to build package path")
	}

	home, ok := os.LookupEnv("HOME")
	if !ok {
		return "", "", errors.New("no GOPATH or HOME in environment")
	}
	stat, err := os.Stat(path.Join(home, "go", "src"))
	if err != nil || !stat.IsDir() {
		return "", "", errors.New("$HOME/go is not a valid GOPATH")
	}
	return path.Join(home, "go"), "", nil
}

func main() {
	flags()

	output, err := cover.Run(profile, packages...)
	if err != nil {
		log.Fatalf("cover failed %v - %v", err, string(output))
	}

	gp, module, err := goPath()
	if err != nil {
		log.Fatalf("gopath failed %v", err)
	}

	exitCode := 0
	for _, e := range cover.Parse(output) {
		realPath := ""
		if len(module) == 0 {
			realPath = path.Join(gp, "src", e.Path)
		} else {
			realPath = path.Join(gp, strings.ReplaceAll(e.Path, module, ""))
		}
		cfg := config(realPath)

		e.Threshold = cfg.Threshold

		if e.Failed() {
			exitCode = 1
		}
		fmt.Println(e.String())
	}
	os.Exit(exitCode)
}
