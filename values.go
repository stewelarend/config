package config

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

//any config value returned to a caller, is stored here
//config cannot change during run-time
//so if a new engine is added, or the value is defined after the initial use
//the same old value will be returned until it is specifically deleted

//configValues is a flat name-value pair that may contain nested values in some names
type values struct {
	sync.Mutex
	name   string //full name in dotted notation
	value  map[string]interface{}
	locked bool //set true when not allowed to change
}

func NewValues(name string, value map[string]interface{}) *values {
	v := &values{
		name:   name,
		value:  map[string]interface{}{},
		locked: false,
	}
	for fieldName, fieldValue := range value {
		if err := v.Set(fieldName, fieldValue); err != nil {
			panic(fmt.Errorf("failed to set init %s.%s: %v", v.name, fieldName, err))
		}
	}
	return v
}

//names may only consist only of alpha-numerics with '_' and '-' in the middle of the name
//names use '.' to nest
const namePattern = `[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])*`

var nameRegex = regexp.MustCompile("^" + namePattern + "$")

//Set a named config value
//name may be dot-notation for nesting
func (v *values) Set(name string, value interface{}) error {
	v.Lock()
	defer v.Unlock()
	nameParts := strings.SplitN(name, ".", 2)
	if len(nameParts) == 0 {
		return fmt.Errorf("missing name")
	}

	if !nameRegex.MatchString(nameParts[0]) {
		return fmt.Errorf("invalid name \"%s\" in \"%s\"", nameParts[0], name)
	}

	if len(nameParts) == 1 {
		if v.locked {
			return fmt.Errorf("%s.%s is locked, cannot change", v.name, nameParts[0])
		}
		sub, ok := v.value[name]
		if ok {
			return fmt.Errorf("%s.%s=(%T)%v cannot be set to (%T)%v", v.name, nameParts[0], sub, sub, value, value)
		}

		if obj, ok := value.(map[string]interface{}); ok {
			sub := NewValues(nameParts[0], nil)
			if v.name != "" {
				sub.name = v.name + "." + nameParts[0]
			}
			for fieldName, fieldValue := range obj {
				if err := sub.Set(fieldName, fieldValue); err != nil {
					return fmt.Errorf("failed to set %s.%s: %v", v.name, fieldName, err)
				}
			}
			v.value[nameParts[0]] = sub
			return nil
		}
		v.value[nameParts[0]] = value
		return nil
	}

	if sub, ok := v.value[nameParts[0]]; ok {
		if subValues := sub.(*values); ok {
			return subValues.Set(nameParts[1], value)
		}
		return fmt.Errorf("%s.%s=(%T)%v cannot set %s=(%T)%v", v.name, nameParts[0], sub, sub, nameParts[1], value, value)
	}
	subValues := NewValues(nameParts[0], nil)
	if v.name != "" {
		subValues.name = v.name + "." + nameParts[0]
	}

	v.value[nameParts[0]] = subValues
	return subValues.Set(nameParts[1], value)
}

//Get a named config value
//name may be dot-notation for nesting
func (v *values) Get(name string) (value interface{}, ok bool) {
	value, err := v.GetWithLock(name, false)
	if err != nil {
		return nil, false
	}
	return value, true
}

func (v *values) GetAndLock(name string) (value interface{}, err error) {
	return v.GetWithLock(name, true)
}

func (v *values) GetWithLock(name string, setLocked bool) (value interface{}, err error) {
	v.Lock()
	defer v.Unlock()
	nameParts := strings.SplitN(name, ".", 2)
	if len(nameParts) == 0 {
		return nil, fmt.Errorf("missing name")
	}

	if !nameRegex.MatchString(nameParts[0]) {
		return nil, fmt.Errorf("invalid name \"%s\" in \"%s\"", nameParts[0], name)
	}

	if len(nameParts) == 1 {
		sub, ok := v.value[name]
		if !ok {
			return nil, fmt.Errorf("%s.%s is not defined", v.name, nameParts[0])
		}
		if subValues, ok := sub.(*values); ok {
			if setLocked {
				subValues.SetLock()
			}
			return subValues.Value(), nil
		}
		return sub, nil
	}

	if sub, ok := v.value[nameParts[0]]; ok {
		if subValues, ok := sub.(*values); ok {
			return subValues.GetWithLock(nameParts[1], setLocked)
		}
		return nil, fmt.Errorf("cannot find \"%s\" inside value (%T)%v", nameParts[1], sub, sub)
		//return sub, nil
	}
	return nil, fmt.Errorf("%s.%s is not defined", v.name, nameParts[0])
}

func (v *values) Value() map[string]interface{} {
	value := map[string]interface{}{}
	for n, v := range v.value {
		if subValues, ok := v.(*values); ok {
			value[n] = subValues.Value()
		} else {
			value[n] = v
		}
	}
	return value
}

func (v *values) SetLock() {
	v.locked = true
	for _, v := range v.value {
		if subValues, ok := v.(*values); ok {
			subValues.SetLock()
		}
	}
}

func (v *values) Merge(b *values) {
	if b != nil {
		for bn, bv := range b.value {
			v.Del(bn)
			if err := v.Set(bn, bv); err != nil {
				panic(fmt.Errorf("cannot merge v(%s) %s=(%T)%+v: %v", v.name, bn, bv, bv, err))
			}
		}
	}
}

func (v *values) Del(name string) error {
	v.Lock()
	defer v.Unlock()
	nameParts := strings.SplitN(name, ".", 2)
	if len(nameParts) == 0 {
		return fmt.Errorf("missing name")
	}

	if !nameRegex.MatchString(nameParts[0]) {
		return fmt.Errorf("invalid name \"%s\" in \"%s\"", nameParts[0], name)
	}

	if len(nameParts) == 1 {
		if v.locked {
			return fmt.Errorf("%s.%s is locked, cannot delete", v.name, nameParts[0])
		}
		delete(v.value, name)
		return nil
	}

	if sub, ok := v.value[nameParts[0]]; ok {
		if subValues := sub.(*values); ok {
			return subValues.Del(nameParts[1])
		}
		return fmt.Errorf("%s.%s=(%T)%v cannot del(%s)", v.name, nameParts[0], sub, sub, nameParts[1])
	}
	return nil
}
