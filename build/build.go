package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var destOS string
var destArch string

func init() {
	flag.StringVar(&destOS, "os", "linux", "Set the target GOOS for cross compilation.")
	flag.StringVar(&destArch, "arch", "amd64", "Set the target GOARCH for cross compilation")
}

func main() {
	flag.Parse()
	curDir, _ := os.Getwd()
	os.Mkdir("geobin", 0777)
	buildDir := fmt.Sprintf("%s/geobin", curDir)
	gopath := os.Getenv("GOPATH")
	geobinPath := fmt.Sprintf("%s/src/github.com/esripdx/geobin.io", gopath)
	command(true, "go", "get", "-u", "github.com/esripdx/geobin.io")
	cd(geobinPath)
	command(true, "go", "get", "-t")
	command(true, "npm", "install")
	env("GOOS", destOS)
	env("GOARCH", destArch)
	command(true, "go", "build", "-o", fmt.Sprintf("%s/geobin", buildDir))
	command(true, "cp", "-R", fmt.Sprintf("%s/static", geobinPath), buildDir)
	command(true, "cp", fmt.Sprintf("%s/config.json.dist", geobinPath), fmt.Sprintf("%s/config.json", buildDir))
	version := getVersion()
	cd(curDir)

	filename := fmt.Sprintf("geobin-%s_%s-%s.tar.gz", version, destOS, destArch)
	os.Setenv("COPYFILE_DISABLE", "true")
	command(true, "tar", "-zcf", filename, "./geobin")

	fmt.Println("Created tar", filename)
}

func command(fatal bool, name string, arg ...string) (out []byte, err error) {
	fmt.Println("Running command:", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	out, err = cmd.Output()
	if fatal && err != nil {
		os.Exit(1)
	}
	return
}

func cd(dir string) {
	fmt.Println("Changing to:", dir)
	if err := os.Chdir(dir); err != nil {
		fmt.Println("Could not cd to", dir, err)
		os.Exit(1)
	}
}

func env(key, val string) {
	fmt.Printf("Setting %s : %s\n", key, val)
	if err := os.Setenv(key, val); err != nil {
		fmt.Println("Could not set", key, err)
		os.Exit(1)
	}
}

func getVersion() (version string) {
	v, err := command(false, "git", "describe")
	if err != nil {
		fmt.Println("No named tag found. Using 'tip'")
		version = "tip"
	} else {
		version = strings.TrimSpace(string(v))
	}
	return
}
