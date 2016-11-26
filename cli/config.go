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
		"interval" : "0s",
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
	},
	"trace" : {
		"wait"  : "2s",
		"theme" : "dark"
	},
	"snmp" : {
		"community"     : "public",
		"timeout"       : "1s",
		"version"       : "2c",
		"retries"       : 1,
		"port"          : 161,
		"securitylevel" : "noauthnopriv",
		"authpass"      : "nopass",
		"authproto"     : "sha",
		"Privacypass"   : "nopass",
		"Privacyproto"  : "aes"
	}
}`

// Config represents configuration
type Config struct {
	Ping  Ping  `json:"ping"`
	Hping HPing `json:"hping"`
	Web   Web   `json:"web"`
	Scan  Scan  `json:"scan"`
	Trace Trace `json:"trace"`
	Snmp  SNMP  `json:"snmp"`
}

// Ping represents ping command options
type Ping struct {
	Timeout  string `json:"timeout" tag:"lower"`
	Interval string `json:"interval" tag:"lower"`
	Count    int    `json:"count"`
}

// HPing represents ping command options
type HPing struct {
	Timeout  string `json:"timeout" tag:"lower"`
	Interval string `json:"interval" tag:"lower"`
	Method   string `json:"method" tag:"upper"`
	Data     string `json:"data"`
	Count    int    `json:"count"`
}

// Web represents web command options
type Web struct {
	Port    int    `json:"port"`
	Address string `json:"address"`
}

// Scan represents scan command options
type Scan struct {
	Port string `json:"port"`
}

// Trace represents trace command options
type Trace struct {
	Wait  string `json:"wait" tag:"lower"`
	Theme string `json:"theme" tag:"lower"`
}

// SNMP represents nms command options
type SNMP struct {
	Community     string `json:"community"`
	Timeout       string `json:"timeout" tag:"lower"`
	Version       string `json:"version" tag:"lower"`
	Retries       int    `json:"retries"`
	Port          int    `json:"port"`
	Securitylevel string `json:"securitylevel"`
	Authpass      string `json:"authpass"`
	Authproto     string `json:"authproto" tag:"lower"`
	Privacypass   string `json:"privacypass"`
	Privacyproto  string `json:"privacyproto" tag:"lower"`
}

// WriteConfig write config to disk
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

// GetOptions returns option(s)/value(s) for specific command
func GetOptions(s interface{}, key string) ([]string, []interface{}) {
	var (
		opts []string
		vals []interface{}
	)
	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Name == key {
			f := v.Field(i)
			ft := f.Type()
			for j := 0; j < f.NumField(); j++ {
				vals = append(vals, f.Field(j))
				opts = append(opts, ft.Field(j).Name)
			}
			break
		}
	}
	return opts, vals
}

// GetCMDNames returns command line names
func GetCMDNames(s interface{}) []string {
	var fields []string

	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	return fields
}

// UpgradeConfig adds / removes new command(s)/option(s)
func UpgradeConfig(cfg *Config) error {
	var (
		conf  map[string]interface{}
		cConf Config
	)

	b := make([]byte, 2048)
	f, err := cfgFile()
	if err != nil {
		return err
	}

	h, err := os.Open(f)
	if err != nil {
		return err
	}
	n, _ := h.Read(b)
	b = b[:n]
	// load saved/old config to conf
	json.Unmarshal(b, &conf)
	// load default config to cConf
	json.Unmarshal([]byte(defaultConfig), &cConf)

	for _, cmd := range GetCMDNames(cConf) {
		opts, vals := GetOptions(cConf, cmd)
		for i, opt := range opts {
			if v, ok := conf[strings.ToLower(cmd)].(interface{}); ok {
				if _, ok = v.(map[string]interface{})[strings.ToLower(opt)]; !ok {
					args := fmt.Sprintf("%s %s %v", cmd, opt, vals[i])
					SetConfig(args, cfg)
				}
			} else {
				// there is new command
				args := fmt.Sprintf("%s %s %v", cmd, opt, vals[i])
				SetConfig(args, cfg)
			}
		}
	}
	return nil
}

// LoadConfig loads configuration
func LoadConfig() Config {
	var cfg Config

	cfg = ReadConfig()
	UpgradeConfig(&cfg)
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

//
func optionProp(v reflect.Value, opt string) (string, string) {
	opt = strings.Title(opt)
	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).Name == opt {
			field := v.Type().Field(i)
			return field.Tag.Get("tag"), field.Type.Name()
		}
	}
	return "", "nan"
}

// SetConfig handles update option's value
func SetConfig(args string, s *Config) error {
	var (
		v       reflect.Value
		integer int64
		float   float64
		err     error
	)

	//args = strings.ToLower(args)
	f := strings.Fields(args)
	if len(f) < 3 {
		helpSet()
		return fmt.Errorf("syntax error")
	}

	cmd, opt, val := f[0], f[1], f[2]
	opt = strings.Title(opt)

	v = reflect.ValueOf(s)
	v = reflect.Indirect(v)
	v = v.FieldByName(strings.Title(cmd))

	if !v.IsValid() {
		return fmt.Errorf("invalid command")
	}

	tags, valType := optionProp(v, opt)
	val = applyTag(val, tags)

	switch valType {
	case "string":
		// string
		err = SetValue(v.Addr(), opt, fmt.Sprintf("%v", val))
	case "int":
		// integer
		if integer, err = strconv.ParseInt(val, 10, 64); err == nil {
			err = SetValue(v.Addr(), strings.Title(opt), integer)
		} else {
			err = fmt.Errorf("the value should be integer")
		}
	case "float":
		// float
		if float, err = strconv.ParseFloat(val, 64); err == nil {
			err = SetValue(v.Addr(), opt, float)
		} else {
			err = fmt.Errorf("the value should be float")
		}
	default:
		err = fmt.Errorf("invalid option")
	}

	if err != nil {
		return err
	}

	// save config
	err = WriteConfig(*s)
	return err
}

// SetValue set optioni's value
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
			fmt.Printf("set %-8s %-15s %v\n", cmd, subCmd, value)
		}
	}
}

func applyTag(val string, typ string) string {
	switch {
	case strings.Contains(typ, "lower"):
		return strings.ToLower(val)
	case strings.Contains(typ, "upper"):
		return strings.ToUpper(val)
	default:
		return val
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
