package main

import (
	"fmt"

	"github.com/stewelarend/config"
	//uncomment this line to load config from ENV:
	//_ "github.com/stewelarend/config/source/env"
	//uncomment this line to load config from FILES:
	"github.com/stewelarend/config/source/configfile"
)

func main() {
	//define devault values
	//this is usually done in the modules that use the values
	config.SetDefault("abc", 123)

	config.SetDefault("myName", "Jan")

	config.SetDefault("server.http", map[string]interface{}{
		"port":    8000,
		"address": "localhost",
	})

	//make your program read a config file (supports XML, JSON and YAML)
	if err := configfile.Add("./config.json"); err != nil {
		panic(err)
	}

	//retrieve the config values:
	if abc, ok := config.GetInt("abc"); !ok {
		panic("did not get expected config")
	} else {
		fmt.Printf("abc    = (%T)%+v\n", abc, abc)
	}

	if myName, ok := config.GetString("myName"); !ok {
		panic("did not get expected config")
	} else {
		fmt.Printf("myName = (%T)%+v\n", myName, myName)
	}

	if httpServer, ok := config.GetValue("server.http").(map[string]interface{}); !ok {
		panic("did not get expected config")
	} else {
		fmt.Printf("server = (%T)%+v\n", httpServer, httpServer)
	}

	//retrieve elements inside default map:
	if port, ok := config.GetInt("server.http.port"); !ok {
		panic("did not get expected config")
	} else {
		fmt.Printf("port  = %d\n", port)
	}

	//retrieve a struct with validation:
	if cfg, err := config.GetStruct("server.http", httpServerConfig{}); err != nil {
		panic(err)
	} else {
		fmt.Printf("http server config struct: (%T)%+v\n", cfg, cfg)
	}
}

type httpServerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

//change receiver to ptr if want to change values in Validate() method...
func (c httpServerConfig) Validate() error {
	fmt.Printf("Validating (%T)%+v ...\n", c, c)
	if c.Port <= 0 {
		return fmt.Errorf("invalid port:%d", c.Port)
	}
	if c.Address == "localhost" {
		c.Address = "myhost" //just to show you can change value, ONLY if receiver is changed to ptr
	}
	return nil
}
