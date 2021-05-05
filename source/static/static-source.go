package static

import (
	"github.com/stewelarend/config"
)

//add a static value to config
func Add(value map[string]interface{}) {
	config.AddSource(config.NewValues("static", value))
}
