# Config for Micro-Services

Support for different config sources and default values.

Out of the box, just call config.Get("...") to get a value from the defaults.
Import config/env to also read environment settings
Import config/file to also read from a file
Write your own engine to read from any other source

## Getting Started
Import the config library where you define default and/or where you use config values:
```
import "github.com/stewelarend/config"
```

Define default values with:
```
config.SetDefault("abc", 123)

config.SetDefault("my.name", "Jan")

config.SetDefault("server.http", map[string]interface{}{
    "port":8000,
    "address":"localhost",
})
```

Retrieve config (default values for now) with:
```
abc,ok := config.GetValue("abc").(int)

myName,ok := config.GetValue("my.name").(string)

httpServer,ok := config.GetValue("server.http").(map[string]interface{})
```

You can also retrieve the server fields one by one:
```
port,ok := config.GetValue("server.http.port").(int)
```

All of this code is in example/main.go

## Config in the ENVIRONMENT
When you run the example, it prints the default values.
To change a value with the env, define it in the console or container:
```
export abc=456
```
Run the example again as is. It will still print the default values.
```
examples$ go run main.go
abc    = (int)123
myName = (string)Jan
server = (map[string]interface {})map[address:localhost port:8000]
```

Now change example/main.go to do an anonymous import of config/source/env and run again to retrieve the value from env.

**NOTE:** _Import the config source libraries only into your main program (or as close as possible to it)_
```
import _ "github.com/stewelarend/config/source/env"
...
abc,ok := config.GetValue("abc").(int)
```
It will now fail on the type assertion because ENV only stores strings. To make this safer, use GetInt() to do the conversion automatically:
```
abc,ok := config.GetInt("abc")
```
You can only define top-level values in ENV. Any use of dotted notation named will fail.

## Config from a file
Example has a JSON file ./config.json
```
{
    "server":{
        "http":{
            "port":9000
        }
    }
}
```
Then import the config/source/configfile library and load the file with:
```
configfile.Add("./config.json")
```

**NOTE:** _Import the config engine libraries only into your main program (or as close as possible to it)_
```
import _ "github.com/stewelarend/config/file"
...
port,ok := config.GetInt("server.http.port")
```
## Config from multiple sources
Import both env and configfile in the order they should apply.

The first match is used. If env is imported first, and it defines the value, that value will be used.

Define a different value in ENV and File, then import both config engines, then change the import order and test again. The first match will always apply. Generally, env should be before file, and env config can be used to change the name of the file.
```
import _ "github.com/stewelarend/config/env"
import _ "github.com/stewelarend/config/file"
...
```
## Config Structs
You can define a config struct with validation, e.g. for your HTTP server, the struct has a field for address and port:
```
type httpServerConfig struct {
    Address string `json:"address"`
    Port int       `json:"port"`
}

func (c httpServerConfig) Validate() error {
    if c.Address == "" {
        return fmt.Errorf("missing address")
    }
    if c.Port <= 0 {
        return fmt.Errorf("negative port")
    }
    return nil
}
```
Retrieve the struct with:
```
c,err := config.GetStruct("server.http", httpServerConfig{})
if err != nil {
    panic(fmt.Errorf("invalid config: %v", err))
}
fmt.Printf("address: %s\n", c.Address)
fmt.Printf("port:    %d\n", c.Port)
```
The Validate() method may also have a pointer receiver if you want to apply some values as part of validation:
```
func (c *httpServerConfig) Validate() error {
    if c.Address == "" {
        c.Address = "localhost" //<<-- use this as default address -->>
    }
    if c.Port <= 0 {
        return fmt.Errorf("negative port")
    }
    return nil
}
```
## Named Config
Example: When your server can be either HTTP or ZMQ, define a config struct for HTTP and another struct for ZMQ and register both defaults as "server.http" and "server.zmq" respectively.

Configure either:
```
{"server":{"http":{...}}}
```
or
```
{"server":{"zmq":{...}}}
```
Then when creating the server, get the named struct to tell you which option is configured:
```
serverName,serverConfig,ok := config.GetNamed("server")
```
If !ok, server is not configured correctly, or has multiple named items.
Then switch on serverName or do lookup in a list to create the correct type of server or fail for unknown serverName.

In example/main-server.go you will see how this can be used to support
multiple implementation. If your implementation register itself, you can add new implementations by importing them and not changing the rest of the code at all, just import and configure.

## Config Changes
No changes are allowed to config at run-time. As soon as a default is set or a value is used, that value cannot be changed in the code again, and run-time changes from sources are not loaded.

<!-- TODO!
## Config Documentation
To document your config, the values can be written to a file, or you can serve them as HTML etc. It is possible to see the default values and the actual values applied in this instance.

You should define defaults for all config to have proper documentation. Values without defaults will only appear in documentation after they have been retrieved.


Use either:
* config.Defaults() to get all the defaults defined in the code
* config.Defined() to get all values used so far
* config.Documented() to get full documentation for each known value

Use structs to define default values with human readable documentation in a doc-tag that will be available in config.Documented() output.
```
type httpServerConfig struct {
    Address string `json:"address" doc:"Interface address"`
    Port int       `json:"port"    doc:"TCP port to listen on"`
}
``` -->
