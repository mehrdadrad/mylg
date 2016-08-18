package icmp_test

import (
	"testing"

	"github.com/mehrdadrad/mylg/icmp"
)

func TestNewTrace(t *testing.T) {
	_, err := icmp.NewTrace("google.com -n -nr -m 30")
	if err != nil {
		t.Error("unexpected error. expected %v, actual %v", nil, err)
	}
}
