package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/user"
	"path/filepath"
)

var (
	cluster = kingpin.Flag("cluster", "Name of Kubernetes context").Default(currentContext()).String()
	authURL = kingpin.Flag("auth-url", "URL to authentication app").URL()
	port    = kingpin.Flag("local-port", "Port to run in-process helper app on").Default("8976").String()
)

func defaultConfigPath() string {
	p, err := homePath(".kube", "config")
	if err != nil {
		fatal(fmt.Errorf("error determining default config path: %s", err.Error()))
	}
	return p
}

func homePath(paths ...string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	parts := append([]string{u.HomeDir}, paths...)
	return filepath.Join(parts...), nil
}

func currentContext() string {
	cfg, err := ReadKubeConfig(defaultConfigPath())
	if err != nil {
		fatal(fmt.Errorf("error reading configuration %s: %s", "", err.Error()))
	}
	return cfg.CurrentContext
}

func fatal(err error) {
	fmt.Println("ERROR", err.Error())
	os.Exit(2)
}

func main() {
	kingpin.Parse()

	err := ExecAuth(*cluster, *authURL, *port)
	if err != nil {
		fatal(err)
	}
}
