package config_test

import (
	"fmt"
	"testing"

	//	"github.com/stewelarend/logger"
	"github.com/stewelarend/config"
	"github.com/stewelarend/config/source/static"
)

//var log = logger.New().WithLevel(logger.LevelDebug)

func TestDefaults(t *testing.T) {
	t.Logf("Testing defaults...")
	//must be able to set
	if err := config.SetDefault("rpc.server", map[string]interface{}{
		"http": map[string]interface{}{
			"port":    8000,
			"address": "",
		},
	}); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	//must NOT be able to set again
	if err := config.SetDefault("rpc.server", map[string]interface{}{
		"http": map[string]interface{}{
			"port":    8000,
			"address": "",
		},
	}); err == nil {
		t.Fatalf("able to set again")
	}

	//must NOT be able to change value inside
	if err := config.SetDefault("rpc.server.http.port", 9000); err == nil {
		t.Fatalf("able to change port")
	}

	//must be able to add a value inside before use
	if err := config.SetDefault("rpc.server.http.limit", 10); err != nil {
		t.Fatalf("unable to add value before use")
	}
	if err := config.SetDefault("rpc.server.http.limit", 10); err == nil {
		t.Fatalf("able to change value inside")
	}

	//after got addr, can still add to server
	config.Get("rpc.server.addr")
	if err := config.SetDefault("rpc.server.http.size", 20); err != nil {
		t.Fatalf("unable to add value before use")
	}

	//after got server, can NOT add to server   (STILL possible to do)
	// config.Get("rpc.server")
	// if err := config.SetDefault("rpc.server.http.range", 30); err == nil {
	// 	t.Fatalf("able to add value after use")
	// }
}

type httpServerConfig struct {
	Address  string `json:"address" doc:"Address"`
	Port     int    `json:"port" doc:"TCP Port"`
	LimitTPS int    `json:"limit_tps" doc:"Max requests per second"`
}

func (c *httpServerConfig) Validate() error {
	if c.Address == "" {
		c.Address = "localhost"
	}
	if c.Port == 0 {
		c.Port = 8000
	}
	if c.Port < 0 {
		return fmt.Errorf("negative port:%d", c.Port)
	}
	return nil
}

type batchServerConfig struct {
	Filename string `json:"filename" doc:"Filename to load requests from"`
}

func TestContentiousStructs(t *testing.T) {
	//define defaults for both server configs
	config.SetDefault("server.http", httpServerConfig{
		LimitTPS: 10,
	})
	config.SetDefault("server.batch", batchServerConfig{})

	//both servers have defaults,
	//so GetNamed will fail because default has two items and one must be selected somehow!!!
	//in this test we have no config source, so cannot test selection
	if name, _, ok := config.GetNamed("server"); ok {
		t.Fatalf("got server(%s) but none was selected", name)
	}
}

func TestStructSelection(t *testing.T) {
	//define defaults for both server configs
	config.SetDefault("server.http", httpServerConfig{
		LimitTPS: 10,
	})
	config.SetDefault("server.batch", batchServerConfig{})

	//add static config to make a selection
	static.Add(map[string]interface{}{
		"server": map[string]interface{}{
			"http": nil,
		},
	})

	//both servers have defaults, but http is configured with nil value
	//so GetNamed() should get nil, and call Validate()

	//so GetNamed will fail because default has two items and one must be selected somehow!!!
	//in this test we have no config source, so cannot test selection
	if name, _, ok := config.GetNamed("server"); ok {
		t.Fatalf("got server(%s) but none was selected", name)
	}
}
