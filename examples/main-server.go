package main

import (
	"fmt"

	"github.com/stewelarend/config"
	"github.com/stewelarend/config/source/configfile"
)

func main() {
	if err := configfile.Add("./config.json"); err != nil {
		panic(err)
	}

	//retrieve a named struct with optional validation:
	//e.g. we want the server and it could be http or zmq:
	//change the config file "http" to "zmq" to get the other option
	//then run again...
	options := map[string]interface{}{
		"http": httpServerConfig{},
		"zmq":  zmqServerConfig{},
	}
	if named, cfg, err := config.GetNamedStruct("server", options); err != nil {
		panic(err)
	} else {
		fmt.Printf("named server(%s) config struct: (%T)%+v\n", named, cfg, cfg)

		//now you can create the configured type of server
		//as long as both config structures implements the constructor interface:
		server, err := cfg.(IServerConstructor).Create()
		if err != nil {
			panic(err)
		}
		if err := server.Serve(); err != nil {
			panic(err)
		}

	}
}

type httpServerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

func (c httpServerConfig) Create() (IServer, error) {
	return httpServer{config: c}, nil
}

type zmqServerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

func (c zmqServerConfig) Create() (IServer, error) {
	return zmqServer{config: c}, nil
}

type IServerConstructor interface {
	Create() (IServer, error)
}

type IServer interface {
	Serve() error
}

type httpServer struct {
	config httpServerConfig
}

func (s httpServer) Serve() error {
	return fmt.Errorf("NYI")
}

type zmqServer struct {
	config zmqServerConfig
}

func (s zmqServer) Serve() error {
	return fmt.Errorf("NYI")
}
