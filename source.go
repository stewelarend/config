package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/stewelarend/logger"
)

var log = logger.New()

type ISource interface {
	//get a key-value pair from this source
	//the name may use dotted notation for nesting
	Get(name string) (value interface{}, ok bool)

	//get a named item and its value
	//e.g. with config {"rpc":{"server":{"http":{"port":8000}}}}
	//     GetNamed("rpc.server") will return:
	//			named="http"
	//			value={"port":8000}
	//GetNamed(name string) (named string, value interface{}, ok bool)
}

type ISourceConstructor interface {
	Create() (ISource, error)
}

func RegisterSource(name string, constructor ISourceConstructor) {
	sourceConstructors[name] = constructor
}

func AddSource(s ISource) {
	if s != nil {
		sourcesMutex.Lock()
		defer sourcesMutex.Unlock()
		sources = append(sources, s)
	}
}

var (
	sourcesMutex       sync.Mutex
	sourceConstructors = map[string]ISourceConstructor{}
	sources            = []ISource{}
	defined            = NewValues("defined", nil)
)

//GetValue() is same as Get() but only returns the value if defined else nil
func GetValue(name string) interface{} {
	if v, ok := Get(name); ok {
		return v
	}
	return nil
}

func GetInt(name string) (int, bool) {
	v, ok := Get(name)
	if !ok {
		return 0, false
	}
	if i, ok := v.(int); ok {
		return i, true
	}
	i64, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
	if err != nil {
		return 0, false
	}
	return int(i64), true
}

func GetString(name string) (string, bool) {
	v, ok := Get(name)
	if !ok {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	return fmt.Sprintf("%v", v), true
}

func Get(name string) (interface{}, bool) {
	log.Debugf("Get(%s)...", name)
	//if already defined, use that value
	if v, err := defined.GetAndLock(name); err == nil {
		return v, true
	}

	//not yet defined, try to retrieve from sources
	sourcesMutex.Lock()
	defer sourcesMutex.Unlock()
	for _, s := range sources {
		if v, ok := s.Get(name); ok {
			//found in this source
			//if this is an object and we also have defaults
			//for the object, then need to merge
			//e.g. if defaults has server:{address:"localhost", port:8000}
			//      and source has server:{port:9000}
			//      then we define server:{address:"localhost", port:9000}
			if sourceObj, ok := v.(map[string]interface{}); ok {
				if defaultValue, ok := defaults.Get(name); ok {
					if defaultObj, ok := defaultValue.(map[string]interface{}); ok {
						//has source and default obj
						//start with default and add source values into it
						v = mergedObj(defaultObj, sourceObj)
					}
				}
			}

			//copy to defined and lock
			if err := defined.Set(name, v); err != nil {
				panic(fmt.Errorf("failed to define config: %v", err))
			}
			v, _ = defined.GetAndLock(name)
			return v, true
		}
	}

	//still not defined, try to retrieve from defaults
	if v, ok := defaults.Get(name); ok {
		if err := defined.Set(name, v); err != nil {
			panic(fmt.Errorf("failed to apply default value: %v", err))
		}
		v, _ := defined.GetAndLock(name)
		return v, true
	}

	//config is undefined
	return nil, false
} //Get()

//template must be a struct
func GetStruct(name string, tmpl interface{}) (interface{}, error) {
	value, ok := Get(name)
	if !ok {
		return nil, fmt.Errorf("%s not defined", name)
	}
	tmplType := reflect.TypeOf(tmpl)
	if tmplType.Kind() == reflect.Ptr {
		tmplType = tmplType.Elem()
	}
	if tmplType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s template is %v != struct", name, tmplType)
	}
	newStructPtrValue := reflect.New(tmplType)
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%s value cannot encode to JSON: %v", name, err)
	}
	if err := json.Unmarshal(jsonValue, newStructPtrValue.Interface()); err != nil {
		return nil, fmt.Errorf("%s value cannot decode into %v: %v", name, tmplType, err)
	}
	if validator, ok := newStructPtrValue.Interface().(IValidator); ok {
		if err := validator.Validate(); err != nil {
			return nil, fmt.Errorf("%s invalid: %v", name, err)
		}
	}
	if reflect.TypeOf(tmpl).Kind() == reflect.Ptr {
		return newStructPtrValue.Interface(), nil
	}
	return newStructPtrValue.Elem().Interface(), nil
} //GetStruct()

func GetNamed(name string) (string, interface{}, bool) {
	value, ok := Get(name)
	if !ok {
		return "", nil, false //name not defined
	}
	if obj, ok := value.(map[string]interface{}); ok {
		if len(obj) != 1 {
			return "", nil, false //not exactly one named item
		}
		for named, value := range obj {
			return named, value, true
		}
	}
	return "", nil, false //name value is not an object
} //GetNamed()

//Get named config into a struct
//templates must be named structs to get type of and parse value into struct
func GetNamedStruct(name string, templates map[string]interface{}) (string, interface{}, error) {
	named, value, ok := GetNamed(name)
	if !ok {
		return "", nil, fmt.Errorf("%s is not defined", name)
	}
	tmpl, ok := templates[named]
	if !ok {
		return "", nil, fmt.Errorf("unknown %s.%s (no template, expecting %s)", name, named, strings.Join(names(templates), "|"))
	}
	tmplType := reflect.TypeOf(tmpl)
	if tmplType.Kind() == reflect.Ptr {
		tmplType = tmplType.Elem()
	}
	if tmplType.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("%s template[%s] is %v != struct", name, named, tmplType)
	}
	newStructPtrValue := reflect.New(tmplType)
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return "", nil, fmt.Errorf("%s.%s value cannot encode to JSON: %v", name, named, err)
	}
	if err := json.Unmarshal(jsonValue, newStructPtrValue.Interface()); err != nil {
		return "", nil, fmt.Errorf("%s.%s value cannot decode into %v: %v", name, named, tmplType, err)
	}
	if validator, ok := newStructPtrValue.Interface().(IValidator); ok {
		if err := validator.Validate(); err != nil {
			return "", nil, fmt.Errorf("%s.%s invalid: %v", name, named, err)
		}
	}
	if reflect.TypeOf(tmpl).Kind() == reflect.Ptr {
		return named, newStructPtrValue.Interface(), nil
	}
	return named, newStructPtrValue.Elem().Interface(), nil
} //GetNamedStruct()

func names(items map[string]interface{}) []string {
	s := []string{}
	for n := range items {
		s = append(s, n)
	}
	return s
}

type IValidator interface {
	Validate() error
}

func mergedObj(a, b map[string]interface{}) map[string]interface{} {
	va := NewValues("a", a)
	vb := NewValues("b", b)
	va.Merge(vb)
	return va.Value()
}
