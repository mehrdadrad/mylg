// Package nms provides network monitoring system through
// different various protocols such as SNMP, SSH
package nms

import (
	"fmt"
	"os"
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
		help()
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
			println(err.Error())
		} else {
			descr := r[0].Variable.(*snmpgo.OctetString).String()
			printEff(trim("Connected: "+descr, 80))
		}
	}
	return client, err
}

// ShowInterface shows interface(s) information based on
// Specific portocol (SNMP/SSH/...) for now it support only SNMP
func (c *Client) ShowInterface() {
	c.snmpShowInterface()
}

func (c *Client) snmpShowInterface() {
	var (
		data [][][]string
		once sync.Once
		spin = spinner.New(spinner.CharSets[26], 220*time.Millisecond)
	)

	for range []int{0, 1} {
		sample, err := c.snmpGetInterfaces()
		if err != nil {
			return
		}
		data = append(data, sample)
		once.Do(
			func() {
				fmt.Printf("* %d interfaces (physical/logical) has been found\n", len(sample)-1)
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
}

func normalize(time0, time1 []string, t int) []string {
	var f = []int{8, 8, 1, 1, 1, 1, 1, 1}

	for _, i := range []int{1, 2, 3, 4} {
		n, _ := strconv.Atoi(time0[i])
		n = n * f[i-1]
		m, _ := strconv.Atoi(time1[i])
		m = m * f[i-1]
		time1[i] = fmt.Sprintf("%d", (m-n)/t)
	}
	return time1
}

func (c *Client) snmpGetInterfaces() ([][]string, error) {
	var (
		data   = make([][]string, 100)
		maxIdx = 0
		oids   []string
		cols   [][]string
	)

	cols = append(cols, []string{"Interface", "ifDescr"})
	cols = append(cols, []string{"Traffic In", "ifHCInOctets"})
	cols = append(cols, []string{"Traffic Out", "ifHCOutOctets"})
	cols = append(cols, []string{"PPS In", "ifHCInUcastPkts"})
	cols = append(cols, []string{"PPS Out", "ifHCOutUcastPkts"})
	cols = append(cols, []string{"Discard In", "ifInDiscards"})
	cols = append(cols, []string{"Discard Out", "ifOutDiscards"})
	cols = append(cols, []string{"Error In", "ifInErrors"})
	cols = append(cols, []string{"Error Out", "ifOutErrors"})

	for _, c := range cols {
		oids = append(oids, OID[c[1]])
		data[0] = append(data[0], c[0])
	}

	r, err := c.SNMP.BulkWalk(oids...)
	if err != nil {
		return data, err
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
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	return data[:maxIdx+1], nil
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

func help() {
	fmt.Println(`
        Usage:
              connect host [options]

        Options:
        Example:
              connect 127.0.0.1 -c public
		`)
}
