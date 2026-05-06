package testapp

import (
	"github.com/doron-cohen/klee"
	"github.com/doron-cohen/klee/log"
	"github.com/doron-cohen/klee/version"
	"github.com/urfave/cli/v3"
)

func init() {
	version.Version = "1.0.0-test"
	version.Commit = "abc1234"
	version.BuildDate = "2026-04-27"
}

// Config is the testapp configuration struct.
type Config struct {
	Host string `yaml:"host" env:"TESTAPP_HOST" default:"localhost"`
	Port int    `yaml:"port" env:"TESTAPP_PORT" default:"8080"`
	log.Config `yaml:"log"`
}

// NewApp builds a klee.App for use in tests.
func NewApp() *klee.App[Config] {
	return klee.New[Config]("testapp", version.String(), []*cli.Command{
		echoCmd,
		flagsCmd,
		failCmd,
		logCmd,
		msgCmd,
	})
}
