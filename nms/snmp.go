package nms

import (
	"fmt"
	"net"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/mehrdadrad/mylg/cli"
)

var (
	OID = map[string]string{
		"sysDescr":             "1.3.6.1.2.1.1.1.0",
		"ifDescr":              "1.3.6.1.2.1.2.2.1.2",
		"ifHCInOctets":         "1.3.6.1.2.1.31.1.1.1.6",
		"ifHCInUcastPkts":      "1.3.6.1.2.1.31.1.1.1.7",
		"ifHCOutOctets":        "1.3.6.1.2.1.31.1.1.1.10",
		"ifHCOutUcastPkts":     "1.3.6.1.2.1.31.1.1.1.11",
		"ifHCInMulticastPkts":  "1.3.6.1.2.1.31.1.1.1.8",
		"ifHCOutMulticastPkts": "1.3.6.1.2.1.31.1.1.1.12",
		"ifHCInBroadcastPkts":  "1.3.6.1.2.1.31.1.1.1.9",
		"ifHCOutBroadcastPkts": "1.3.6.1.2.1.31.1.1.1.13",
		"ifInDiscards":         "1.3.6.1.2.1.2.2.1.13",
		"ifInErrors":           "1.3.6.1.2.1.2.2.1.14",
		"ifOutDiscards":        "1.3.6.1.2.1.2.2.1.19",
		"ifOutErrors":          "1.3.6.1.2.1.2.2.1.20",
	} // OID holds essentiall OIDs that need for each interface
)

// SNMPClient represents all necessary SNMP parameters
type SNMPClient struct {
	Args     snmpgo.SNMPArguments
	Host     string
	SysDescr string
}

// NewSNMP sets and validates SNMP parameters
func NewSNMP(a string, cfg cli.Config) (*SNMPClient, error) {
	var (
		host, flag = cli.Flag(a)
		community  = cli.SetFlag(flag, "c", cfg.Snmp.Community).(string)
		timeout    = cli.SetFlag(flag, "t", cfg.Snmp.Timeout).(string)
		version    = cli.SetFlag(flag, "v", cfg.Snmp.Version).(string)
		retries    = cli.SetFlag(flag, "r", cfg.Snmp.Retries).(int)
		port       = cli.SetFlag(flag, "p", cfg.Snmp.Port).(int)
	)

	tDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return &SNMPClient{}, err
	}

	args := snmpgo.SNMPArguments{
		Timeout:   tDuration,
		Address:   net.JoinHostPort(host, fmt.Sprintf("%d", port)),
		Retries:   uint(retries),
		Community: community,
	}

	// set SNMP version
	switch version {
	case "1":
		args.Version = snmpgo.V1
	case "2", "2c":
		args.Version = snmpgo.V2c
	case "3":
		args.Version = snmpgo.V3
	default:
		return &SNMPClient{}, fmt.Errorf("wrong version")
	}

	if args.Version == snmpgo.V3 {
		checkAuth()
	}

	return &SNMPClient{
		Args:     args,
		Host:     host,
		SysDescr: "",
	}, nil
}

func checkAuth() {
	//TODO
}

// BulkWalk retrieves a subtree of management values
func (c *SNMPClient) BulkWalk(oid ...string) ([]*snmpgo.VarBind, error) {
	var r []*snmpgo.VarBind
	snmp, err := snmpgo.NewSNMP(c.Args)
	if err != nil {
		return r, err
	}

	oids, err := snmpgo.NewOids(oid)
	if err != nil {
		return r, err
	}

	if err = snmp.Open(); err != nil {
		return r, err
	}
	defer snmp.Close()
	pdu, err := snmp.GetBulkWalk(oids, 0, 100)
	if err != nil {
		return r, err
	}
	r = pdu.VarBinds()
	return r, nil
}

// GetOIDs retrieves values based on the oid(s)
func (c *SNMPClient) GetOIDs(oid ...string) ([]*snmpgo.VarBind, error) {
	var r []*snmpgo.VarBind
	snmp, err := snmpgo.NewSNMP(c.Args)
	if err != nil {
		return r, err
	}

	oids, err := snmpgo.NewOids(oid)
	if err != nil {
		return r, err
	}

	if err = snmp.Open(); err != nil {
		return r, err
	}
	defer snmp.Close()

	pdu, err := snmp.GetRequest(oids)
	if err != nil {
		return r, err
	}
	r = pdu.VarBinds()

	return r, nil
}
