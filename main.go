package main

import (
	"fmt"
	"os"

	"time"

	"github.com/jessevdk/go-flags"
	"github.com/sahilm/jaal/jaal"
	"github.com/sahilm/jaal/ssh"
	"github.com/sirupsen/logrus"
)

var version = "latest"

func main() {
	var opts struct {
		SSHHostKeyFile string `long:"ssh-host-key-file" description:"path to the ssh host key file"`
		SSHPort        uint   `long:"ssh-port" description:"port to listen on for ssh traffic" default:"22"`
		Version        func() `long:"version" description:"print version and exit"`
	}

	opts.Version = func() {
		fmt.Fprintf(os.Stderr, "%v\n", version)
		os.Exit(0)
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	logger := logrus.New()
	logger.Out = os.Stderr
	logger.Formatter = &jaal.UTCFormatter{Formatter: &logrus.JSONFormatter{}}

	eventLogger := jaal.NewEventLogger(os.Stdout)

	sshServer := ssh.Server{
		Addr:        fmt.Sprintf(":%v", opts.SSHPort),
		HostKeyFile: opts.SSHHostKeyFile,
		IdleTimeout: 10 * time.Second,
		MaxTimeout:  1 * time.Hour,
		Logger:      logger,
	}

	jaal.ListenAndLog(eventLogger, logger, sshServer)
}
