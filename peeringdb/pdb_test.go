package peeringdb_test

import (
	"github.com/mehrdadrad/mylg/peeringdb"
	"gopkg.in/h2non/gock.v0"
	"testing"
)

func TestIsASN(t *testing.T) {
	if peeringdb.IsASN("mylg") {
		t.Error("expected none ASN but it recognized incorrectly")
	}
}

func TestGetNetIXLAN(t *testing.T) {
	peer1 := peeringdb.Peer{Name: "a", ASN: 1}
	peer2 := peeringdb.Peer{Name: "b", ASN: 2}
	gock.New(peeringdb.APINetIXLAN).
		Reply(200).
		JSON(map[string][]peeringdb.Peer{"data": {peer1, peer2}})
	ix, err := peeringdb.GetNetIXLAN()
	if err != nil {
		t.Error("")
	}
	// test the data return correctly
	data := ix.(peeringdb.Peers)
	for k, v := range data.Data {
		if k == 0 && v.Name != "a" {
			t.Error("")
		}
		if k == 1 && v.Name != "b" {
			t.Error("")
		}
	}
	// test none 200 HTTP code
	gock.New(peeringdb.APINetIXLAN).
		Reply(403)
	_, err = peeringdb.GetNetIXLAN()
	if err == nil {
		t.Error("")
	}

}
