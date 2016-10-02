// Package nms provides network monitoring system through
// different various protocols such as SNMP, SSH
package nms

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/k-sone/snmpgo"
	"github.com/olekukonko/tablewriter"

	"github.com/mehrdadrad/mylg/cli"
)

// Client represents NMS client
type Client struct {
	SNMP *SNMPClient
	Host string
}

// NewClient makes new NMS client
func NewClient(args string, cfg cli.Config) (Client, error) {
	var (
		client Client
		err    error
	)

	_, flags := cli.Flag(args)
	if _, ok := flags["help"]; ok {
		help(cfg)
		return client, nil
	}

	switch {

	default:
		if client.SNMP, err = NewSNMP(args, cfg); err != nil {
			return client, err
		}
		client.Host = client.SNMP.Host

		r, err := client.SNMP.GetOIDs(OID["sysDescr"])
		if err != nil {
			return client, err
		}
		switch r[0].Variable.(type) {
		case *snmpgo.NoSucheObject, *snmpgo.NoSucheInstance, *snmpgo.EndOfMibView:
			return client, fmt.Errorf("no such object / instance")
		case *snmpgo.OctetString:
			descr := r[0].Variable.(*snmpgo.OctetString).String()
			printEff(trim("Connected: "+descr, 80))
		}
	}
	return client, err
}

// ShowInterface prints out interface(s) information based on
// specific portocol (SNMP/SSH/...) for now it supports only SNMP
func (c *Client) ShowInterface(filter string) error {
	if c.SNMP == nil {
		return fmt.Errorf("snmp not connected, try connect help")
	}
	if err := c.snmpShowInterface(filter); err != nil {
		return err
	}
	return nil
}

// snmpGetIdx finds SNMP index(es) based on the filter
func (c *Client) snmpGetIdx(filter string) []int {
	var res []int

	filter = fmt.Sprintf("^%s$", filter)
	filter = strings.Replace(filter, "*", ".*", -1)
	re := regexp.MustCompile(filter)

	r, _ := c.SNMP.BulkWalk(OID["ifDescr"])
	for _, v := range r {
		a := strings.Split(v.Oid.String(), ".")
		if re.MatchString(v.Variable.String()) {
			idx, _ := strconv.Atoi(a[len(a)-1])
			res = append(res, idx)
		}
	}
	return res
}

func (c *Client) snmpShowInterface(filter string) error {
	var (
		data [][][]string
		once sync.Once
		idxs []int
		spin = spinner.New(spinner.CharSets[26], 220*time.Millisecond)
	)

	if len(strings.TrimSpace(filter)) > 1 {
		idxs = c.snmpGetIdx(filter)
	}

	for range []int{0, 1} {
		sample, err := c.snmpGetInterfaces(idxs)
		if err != nil {
			return err
		}
		if len(sample)-1 < 1 {
			return fmt.Errorf("could not find any interface")
		}

		data = append(data, sample)
		once.Do(
			func() {
				fmt.Printf("%d interfaces has been found\n", len(sample)-1)
				spin.Prefix = "please wait "
				spin.Start()
				time.Sleep(10 * time.Second)
				spin.Stop()
			},
		)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(data[0][0])

	data[0] = data[0][1:] // remove title row
	data[1] = data[1][1:] // remove title row

	for i := range data[0] {
		row := normalize(data[0][i], data[1][i], 10)
		table.Append(row)
	}
	table.Render()

	return nil
}

func (c *Client) snmpGetInterfaces(filter []int) ([][]string, error) {
	var (
		data = make(map[int][]string, 1000)
		oids []string
		cols [][]string
		res  [][]string
		idxs []int
		r    []*snmpgo.VarBind
		err  error
	)

	cols = append(cols, []string{"Interface", "ifDescr"})
	cols = append(cols, []string{"Status", "ifOperStatus"})
	cols = append(cols, []string{"Traffic In", "ifHCInOctets"})
	cols = append(cols, []string{"Traffic Out", "ifHCOutOctets"})
	cols = append(cols, []string{"PPS In", "ifHCInUcastPkts"})
	cols = append(cols, []string{"PPS Out", "ifHCOutUcastPkts"})
	cols = append(cols, []string{"Discard In", "ifInDiscards"})
	cols = append(cols, []string{"Discard Out", "ifOutDiscards"})
	cols = append(cols, []string{"Error In", "ifInErrors"})
	cols = append(cols, []string{"Error Out", "ifOutErrors"})

	if len(filter) < 1 {
		for _, c := range cols {
			oids = append(oids, OID[c[1]])
			data[0] = append(data[0], c[0])
		}

		r, err = c.SNMP.BulkWalk(oids...)
		if err != nil {
			return [][]string{}, err
		}
	} else {
		for _, c := range cols {
			for _, idx := range filter {
				oids = append(oids, fmt.Sprintf("%s.%d", OID[c[1]], idx))
			}
			data[0] = append(data[0], c[0])
		}

		r, err = c.SNMP.GetOIDs(oids...)
		if err != nil {
			return [][]string{}, err
		}
	}

	for _, v := range r {
		a := strings.Split(v.Oid.String(), ".")
		idx, _ := strconv.Atoi(a[len(a)-1])
		if len(data[idx]) < 1 {
			data[idx] = make([]string, len(cols))
		}

		colNum := 0
		for _, c := range cols {
			if OID[c[1]] == strings.Join(a[:len(a)-1], ".") {
				data[idx][colNum] = v.Variable.String()
				break
			}
			colNum++
		}
	}

	// put snmp indexes to idxs & sort them
	for k := range data {
		idxs = append(idxs, k)
	}
	sort.Ints(idxs)

	// convert map (data)  to slice (res)
	for i := range idxs {
		res = append(res, data[i])
	}

	return res, nil
}

func normalize(time0, time1 []string, t int) []string {
	var f = []int{8, 8, 1, 1, 1, 1, 1, 1}

	for _, i := range []int{2, 3, 4, 5} {
		n, _ := strconv.Atoi(time0[i])
		n = n * f[i-1]
		m, _ := strconv.Atoi(time1[i])
		m = m * f[i-1]
		time1[i] = fmt.Sprintf("%d", (m-n)/t)
	}
	// interface status
	time1[1] = ifStatus(time1[1])

	return time1
}

func ifStatus(s string) string {
	switch s {
	case "1":
		return "Up"
	case "2":
		return "Down"
	default:
		return "Unknown"
	}
}

func trim(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n] + " ..."
}

func printEff(s string) {
	for _, c := range s {
		fmt.Printf("%s", string(c))
		time.Sleep(3 * time.Millisecond)
	}
	println("")
}

func help(cfg cli.Config) {
	fmt.Printf(`
        SNMP Usage:
              connect host [options]

        Options:
              -v version           Specifies the protocol version: 1/2c/3 (default: %s)
              -c community         Community string for SNMPv1/v2c transactions (default: %s)
              -t timeout           Specify a timeout in format "ms", "s", "m" (default: %s)
              -p port              Specify SNMP port number (default: %d)
              -r retries           Specifies the number of retries (default:%d)
              -l security level    Security level (NoAuthNoPriv|AuthNoPriv|AuthPriv) (default: %s)
              -a auth protocol     Authentication protocol (MD5|SHA) (default: %s)
              -A auth password     Authentication protocol pass phrase
              -x privacy protocol  Privacy protocol (DES|AES) (default: %s)
              -X privacy password  Privacy protocol pass phrase

        Example:
              connect 127.0.0.1 -c public
		`,
		cfg.Snmp.Version,
		cfg.Snmp.Community,
		cfg.Snmp.Timeout,
		cfg.Snmp.Port,
		cfg.Snmp.Retries,
		cfg.Snmp.Securitylevel,
		cfg.Snmp.Authproto,
		cfg.Snmp.Privacyproto)
}
