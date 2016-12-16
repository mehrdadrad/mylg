// Package speedtest interfaces for testing internet bandwidth through HTTP by speedtest.net
package speedtest

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Server struct {
}

type ST struct {
	cfg     Settings
	servers []Server
}

type Settings struct {
	Download struct {
		TestLength    int `xml:"testlength,attr"`
		ThreadPperURL int `xml:"threadsperurl,attr"`
	} `xml:"download"`
	Upload struct {
		Ratio int `xml:"ratio,attr"`
	} `xml:"upload"`
	Client struct {
		ISP string `xml:"isp,attr"`
	} `xml:"client"`
	ServerCfg struct {
		IgnoreIds string `xml:"ignoreids,attr"`
	} `xml:"server-config"`
}

func Run() error {
	st := ST{}
	if err := st.getConfig(); err != nil {
		return err
	}
	return nil
}

func getData(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("can not connect")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (st *ST) getConfig() error {
	b, err := getData("http://www.speedtest.net/speedtest-config.php")
	if err != nil {
		return err
	}
	xml.Unmarshal(b, &st.cfg)
	return nil
}

func (st *ST) getServers() error {
	var (
		servers = []string{
			"http://www.speedtest.net/speedtest-servers-static.php",
			"http://c.speedtest.net/speedtest-servers-static.php",
		}
	)
	for _, server := range servers {
		getData(server)
	}
	return nil
}
