package nms

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/k-sone/snmpgo"

	"github.com/mehrdadrad/mylg/cli"
)

var (
	// OID holds essentiall OIDs that need for each interface
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
	}
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
		host, flag    = cli.Flag(a)
		community     = cli.SetFlag(flag, "c", cfg.Snmp.Community).(string)
		timeout       = cli.SetFlag(flag, "t", cfg.Snmp.Timeout).(string)
		version       = cli.SetFlag(flag, "v", cfg.Snmp.Version).(string)
		retries       = cli.SetFlag(flag, "r", cfg.Snmp.Retries).(int)
		port          = cli.SetFlag(flag, "p", cfg.Snmp.Port).(int)
		securityLevel = cli.SetFlag(flag, "l", cfg.Snmp.Securitylevel).(string)
		privacyProto  = cli.SetFlag(flag, "x", cfg.Snmp.Privacyproto).(string)
		privacyPass   = cli.SetFlag(flag, "X", cfg.Snmp.Privacypass).(string)
		authProto     = cli.SetFlag(flag, "a", cfg.Snmp.Authproto).(string)
		authPass      = cli.SetFlag(flag, "A", cfg.Snmp.Authpass).(string)
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
		authProto = strings.ToUpper(authProto)
		privacyProto = strings.ToUpper(privacyProto)
		// set SNMP3 configuration
		args.Version = snmpgo.V3
		args.AuthProtocol = snmpgo.AuthProtocol(authProto)
		args.AuthPassword = authPass
		args.PrivProtocol = snmpgo.PrivProtocol(privacyProto)
		args.PrivPassword = privacyPass
	default:
		return &SNMPClient{}, fmt.Errorf("wrong version")
	}

	// set security level - SNMP3
	switch strings.ToLower(securityLevel) {
	case "noauthnopriv":
		args.SecurityLevel = snmpgo.NoAuthNoPriv
	case "authnopriv":
		args.SecurityLevel = snmpgo.AuthNoPriv
	case "authpriv":
		args.SecurityLevel = snmpgo.AuthPriv
	}

	// check args before try to connect
	if err = argsValidate(&args); err != nil {
		return &SNMPClient{}, err
	}

	return &SNMPClient{
		Args:     args,
		Host:     host,
		SysDescr: "",
	}, nil
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
	pdu, err := snmp.GetBulkWalk(oids, 0, 10)
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
	if pdu.ErrorStatus() != snmpgo.NoError {
		return r, fmt.Errorf("%d %d", pdu.ErrorStatus(), pdu.ErrorIndex())
	}

	r = pdu.VarBinds()

	return r, nil
}

// inspired from snmpgo/client.go -> validate()
func argsValidate(a *snmpgo.SNMPArguments) error {
	// check version
	v := a.Version
	if v != snmpgo.V1 && v != snmpgo.V2c && v != snmpgo.V3 {
		return fmt.Errorf("unknown SNMP version")
	}
	// check SNMPv3
	if v == snmpgo.V3 {
		// RFC3414 Section 5
		if l := len(a.UserName); l < 1 || l > 32 {
			return fmt.Errorf("username length is range 1..32")
		}
		if a.SecurityLevel > snmpgo.NoAuthNoPriv {
			// RFC3414 Section 11.2
			if len(a.AuthPassword) < 8 {
				return fmt.Errorf("authpass is at least 8 characters in length")
			}
			if p := a.AuthProtocol; p != snmpgo.Md5 && p != snmpgo.Sha {
				return fmt.Errorf("illegal authproto")
			}
		}
		if a.SecurityLevel > snmpgo.AuthNoPriv {
			// RFC3414 Section 11.2
			if len(a.PrivPassword) < 8 {
				return fmt.Errorf("privacypass is at least 8 characters in length")
			}
			if p := a.PrivProtocol; p != snmpgo.Des && p != snmpgo.Aes {
				return fmt.Errorf("illegal privacyproto")
			}
		}
	}
	return nil
}
