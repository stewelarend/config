package config

import (
	"fmt"
)

var (
	defaults = NewValues("defaults", nil)
)

//Set a default value to use if value is not found in any config engine
//Fails when already defined (which may be from a config engine or default previously set)
func SetDefault(name string, defaultValue interface{}) error {
	if definedValue, ok := defined.Get(name); ok {
		return fmt.Errorf("cannot set default for %s=(%T)%+v (already defined)", name, definedValue, definedValue)
	}
	if err := defaults.Set(name, defaultValue); err != nil {
		return fmt.Errorf("failed to set default in defaults: %v", err)
	}
	return nil
}
