package env

import (
	"os"

	"github.com/stewelarend/config"
	"github.com/stewelarend/logger"
)

var log = logger.New()

func init() {
	config.AddSource(envSource{})
}

type envSource struct{}

func (e envSource) Get(name string) (interface{}, bool) {
	s := os.Getenv(name)
	log.Debugf("Get(%s)=(%T)\"%s\"", name, s, s)
	if s == "" {
		return nil, false
	}
	return s, true
}
