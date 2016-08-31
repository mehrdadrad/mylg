package cli_test

import (
	"reflect"
	"testing"

	"github.com/mehrdadrad/mylg/cli"
)

func TestDefaultConfig(t *testing.T) {
	_, err := cli.ReadDefaultConfig()
	if err != nil {
		t.Error("default configuration failed")
	}
}

func TestGetOptions(t *testing.T) {
	type CMD1 struct {
		key1 string
		key2 int
	}
	type CMD2 struct {
		key3 string
		key4 int
	}

	s := struct {
		CMD1
		CMD2
	}{
		CMD1{"value1", 1},
		CMD2{"value2", 2},
	}
	k, v := cli.GetOptions(s, "CMD1")
	if k[0] != "key1" || k[1] != "key2" {
		t.Error("unexpected key(s) - GetOptions")
	}
	if v[0].(reflect.Value).String() != "value1" {
		t.Error("unexpected value - GetOptions")
	}
	if v[1].(reflect.Value).Int() != 1 {
		t.Error("unexpected value - GetOptions")
	}

	k, v = cli.GetOptions(s, "CMD2")
	if k[0] != "key3" || k[1] != "key4" {
		t.Error("unexpected key(s) - GetOptions")
	}
	if v[0].(reflect.Value).String() != "value2" {
		t.Error("unexpected value - GetOptions")
	}
	if v[1].(reflect.Value).Int() != 2 {
		t.Error("unexpected value i GetOptions")
	}
}

func TestGetCMDNames(t *testing.T) {
	type CMD1 struct {
		key1 string
		key2 int
	}
	type CMD2 struct {
		key3 string
		key4 int
	}

	s := struct {
		CMD1
		CMD2
	}{
		CMD1{"value1", 1},
		CMD2{"value2", 2},
	}

	c := cli.GetCMDNames(s)
	if c[0] != "CMD1" || c[1] != "CMD2" {
		t.Error("unexpected value - GetCMDNames")
	}
}
