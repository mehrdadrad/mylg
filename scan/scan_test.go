package scan_test

import (
	"github.com/mehrdadrad/mylg/scan"
	"testing"
)

var s scan.Scan

func TestIsCIDR(t *testing.T) {
	var err error
	s, err = scan.NewScan("8.8.8.0/24")
	if err != nil {
		t.Error("NewScan failed")
	}
	if !s.IsCIDR() {
		t.Error("IsCIDR failed")
	}
	s, err = scan.NewScan("8.8.8.8")
	if err != nil {
		t.Error("NewScan failed")
	}
	if s.IsCIDR() {
		t.Error("IsCIDR failed")
	}
}
