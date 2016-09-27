package disc_test

import (
	"reflect"
	"testing"

	"github.com/mehrdadrad/mylg/disc"
)

func TestWalkIP(t *testing.T) {
	var result []string
	CIDR := "192.168.1.0/30"
	expectedIPs := []string{
		"192.168.1.0",
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
	}

	for ip := range disc.WalkIP(CIDR) {
		result = append(result, ip)
	}

	if eq := reflect.DeepEqual(result, expectedIPs); !eq {
		t.Error("WalkIP returns unexpected IP address(es)")
	}
}
