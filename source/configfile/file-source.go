package configfile

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/stewelarend/config"
	"gopkg.in/yaml.v2"
)

func Add(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file(%s): %v", filename, err)
	}
	defer f.Close()

	if strings.HasSuffix(filename, ".json") {
		var data map[string]interface{}
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			return fmt.Errorf("cannot read JSON object from file(%s): %v", filename, err)
		}
		config.AddSource(config.NewValues(filename, data))
		return nil
	}

	if strings.HasSuffix(filename, ".xml") {
		var data map[string]interface{}
		if err := xml.NewDecoder(f).Decode(&data); err != nil {
			return fmt.Errorf("cannot read XML object from file(%s): %v", filename, err)
		}
		config.AddSource(config.NewValues(filename, data))
		return nil
	}

	if strings.HasSuffix(filename, ".yaml") {
		var data map[string]interface{}
		if err := yaml.NewDecoder(f).Decode(&data); err != nil {
			return fmt.Errorf("cannot read YAML object from file(%s): %v", filename, err)
		}
		config.AddSource(config.NewValues(filename, data))
		return nil
	}
	return fmt.Errorf("unknown suffix in filename(%s) expecting json|xml|yaml", filename)
}
