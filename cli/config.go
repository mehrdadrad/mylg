package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"
)

var defaultConfig = `{
	"ping" : {
		"timeout" : "2s",
		"interval": "1s",
		"count":	4
	},	
	"hping" : {
		"timeout"  : "2s",
		"method"   : "HEAD",
		"data"	   : "mylg",
		"count"	   : 5
	},
	"web" : {
		"port"	   : 8080,
		"address"  : "127.0.0.1"
	},
	"scan" : {
		"port"     : "1-500"		
	}
}`

// Config represents configuration
type Config struct {
	Ping  Ping  `json:"ping"`
	Hping HPing `json:"hping"`
	Web   Web   `json:"web"`
	Scan  Scan  `json:"scan"`
}

// Ping represents ping command options
type Ping struct {
	Timeout  string `json:"timeout"`
	Interval string `json:"interval"`
	Count    int    `json:"count"`
}

// HPing represents ping command options
type HPing struct {
	Timeout string `json:"timeout"`
	Method  string `json:"method"`
	Data    string `json:"data"`
	Count   int    `json:"count"`
}

// Web represents web command options
type Web struct {
	Port    int    `json:port`
	Address string `json:address`
}

// Scan represents scan command options
type Scan struct {
	Port string `json:port`
}

// UpdateConfig
func WriteConfig(cfg Config) error {
	f, err := cfgFile()
	if err != nil {
		return err
	}

	h, err := os.Create(f)
	if err != nil {
		return err
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = h.Write(b)
	if err != nil {
		return err
	}
	h.Close()

	return nil
}

// UpgradeConfig adds / removes new command(s)/option(s)
func UpgradeConfig() {
	// TODO
	var conf map[string]interface{}
	b := make([]byte, 2048)
	f, err := cfgFile()
	if err != nil {

	}
	h, err := os.Open(f)
	n, _ := h.Read(b)
	b = b[:n]

	json.Unmarshal(b, &conf)
	if v, ok := conf["ping"].(interface{}); ok {
		if _, ok = v.(map[string]interface{})["timeout"]; !ok {
			// there is new option
		}
	} else {
		// there is new command
	}

}

// LoadConfig loads configuration
func LoadConfig() Config {
	var cfg Config

	cfg = ReadConfig()

	return cfg
}

// InitConfig creates new config file
func InitConfig(f string) ([]byte, error) {
	h, err := os.Create(f)
	if err != nil {
		return []byte(""), err
	}

	h.Chmod(os.FileMode(int(0600)))
	h.WriteString(defaultConfig)
	h.Close()

	return []byte(defaultConfig), nil
}

// ReadConfig reads configuration from existing
// or default configuration
func ReadConfig() Config {
	var (
		b    = make([]byte, 2048)
		conf Config
		err  error
	)
	f, err := cfgFile()
	if err != nil {

	}

	h, err := os.Open(f)

	if err != nil {
		switch {
		case os.IsNotExist(err):
			if b, err = InitConfig(f); err != nil {
				println(err.Error())
			}
		case os.IsPermission(err):
			println("cannot read configuration file due to insufficient permissions")
			b = []byte(defaultConfig)
		default:
			println(err.Error())
			b = []byte(defaultConfig)
		}
	} else {
		n, _ := h.Read(b)
		b = b[:n]
	}

	err = json.Unmarshal(b, &conf)
	if err != nil {
		println(err.Error())
		b = []byte(defaultConfig)
		json.Unmarshal(b, &conf)
	}

	return conf
}

// ReadDefaultConfig returns default configuration
func ReadDefaultConfig() (Config, error) {
	var (
		b    = make([]byte, 2048)
		conf Config
	)
	b = []byte(defaultConfig)
	err := json.Unmarshal(b, &conf)
	return conf, err
}

// cfgFile returns config file
func cfgFile() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return user.HomeDir + "/.mylg.config", nil
}

// SetConfig handles update option's value
func SetConfig(args string, s *Config) {
	var (
		v     reflect.Value
		i     int64
		float float64
		err   error
	)

	args = strings.ToLower(args)
	f := strings.Fields(args)
	if len(f) < 1 {
		helpSet()
		return
	}

	v = reflect.ValueOf(s)
	v = reflect.Indirect(v)
	v = v.FieldByName(strings.Title(f[0]))

	if v.IsValid() {
		if i, err = strconv.ParseInt(f[2], 10, 64); err == nil {
			// integer
			err = SetValue(v.Addr(), strings.Title(f[1]), i)
		} else if float, err = strconv.ParseFloat(f[2], 64); err == nil {
			// float
			err = SetValue(v.Addr(), strings.Title(f[1]), float)
		} else {
			// string
			err = SetValue(v.Addr(), strings.Title(f[1]), f[2])
		}
	} else {
		err = fmt.Errorf("invalid")
	}

	if err != nil {
		println(err.Error())
	} else {
		if err = WriteConfig(*s); err != nil {
			println(err.Error())
		}
	}

}

// SetConfig set optioni's value
func SetValue(v reflect.Value, rec string, val interface{}) error {

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer value")
	}

	v = reflect.Indirect(v)
	switch v.Kind() {
	case reflect.Int:
		if value, ok := val.(int64); ok {
			v.SetInt(value)
		} else {
			return fmt.Errorf("the value should be integer")
		}
	case reflect.Float64:
		if value, ok := val.(float64); ok {
			v.SetFloat(value)
		} else {
			return fmt.Errorf("the value should be float")
		}
	case reflect.String:
		if value, ok := val.(string); ok {
			v.SetString(value)
		} else {
			return fmt.Errorf("the value shouldn't be number")
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name == rec {
				err := SetValue(v.Field(i).Addr(), rec, val)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ShowConfig prints the configuration
func ShowConfig(s *Config) {
	var v reflect.Value

	v = reflect.ValueOf(s)
	v = reflect.Indirect(v)

	for i := 0; i < v.NumField(); i++ {
		cmd := v.Type().Field(i).Name
		cmd = strings.ToLower(cmd)

		vv := v.Field(i).Addr()
		vv = reflect.Indirect(vv)

		for j := 0; j < vv.NumField(); j++ {
			subCmd := vv.Type().Field(j).Name
			subCmd = strings.ToLower(subCmd)
			value := vv.Field(j)
			fmt.Printf("set %-8s %-10s %v\n", cmd, subCmd, value)
		}
	}
}

// helpSet shows set command
func helpSet() {
	println(`
          usage:
               set command option value
          example:
               set ping timeout 2s		  
	`)

}
