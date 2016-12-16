// Package speedtest interfaces for testing internet bandwidth through HTTP by speedtest.net
package speedtest

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

type byDistance []Server
type ST struct {
	cfg     Settings
	servers []Server
}
type Client struct {
	IP  string  `xml:"ip,attr"`
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
	ISP string  `xml:"isp,attr"`
}
type Server struct {
	Name     string  `xml:"name,attr"`
	Sponsor  string  `xml:"sponsor,attr"`
	Country  string  `xml:"country,attr"`
	URL      string  `xml:"url,attr"`
	URL2     string  `xml:"url2,attr"`
	Lat      float64 `xml:"lat,attr"`
	Lon      float64 `xml:"lon,attr"`
	Distance float64
}

type Hosts struct {
	Server []Server `xml:"servers>server"`
}

type Settings struct {
	Download struct {
		TestLength    int `xml:"testlength,attr"`
		ThreadsPerURL int `xml:"threadsperurl,attr"`
	} `xml:"download"`
	Upload struct {
		Ratio         int `xml:"ratio,attr"`
		MaxChunkCount int `xml:"maxchunkcount,attr"`
		Threads       int `xml:"threads,attr"`
		TestLength    int `xml:"testlength,attr"`
	} `xml:"upload"`
	ServerCfg struct {
		IgnoreIds string `xml:"ignoreids,attr"`
	} `xml:"server-config"`
	Client struct {
		Client
	} `xml:"client"`
}

func Run() error {
	st := ST{}
	if err := st.getConfig(); err != nil {
		return err
	}
	if err := st.getServers(); err != nil {
		return err
	}
	server := st.bestServer()
	if server.Distance == 0 {
		return fmt.Errorf("could not find a server")
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
	req.Header.Set("Cache-Control", "no-cache")

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
		stServers = []string{
			"http://www.speedtest.net/speedtest-servers-static.php",
			"http://c.speedtest.net/speedtest-servers-static.php",
		}
		hosts Hosts
	)

	for _, server := range stServers {
		b, err := getData(server)
		if err != nil {
			continue
		}

		xml.Unmarshal(b, &hosts)

		st.cfg.Client.Lon = st.cfg.Client.Lon * math.Pi / 180
		st.cfg.Client.Lat = st.cfg.Client.Lat * math.Pi / 180

		for i, server := range hosts.Server {
			hosts.Server[i].Distance = distance(st.cfg.Client.Lon, st.cfg.Client.Lat, server)
		}

		sort.Sort(byDistance(hosts.Server))
		st.servers = hosts.Server
		break
	}
	return nil
}

func (st *ST) bestServer() Server {
	var (
		latency float64
		sum     float64
		server  Server
	)
	latency = 1000
	for i := range []int{1, 2, 3, 4} {
		base := st.servers[i].URL[:strings.LastIndex(st.servers[i].URL, "/")]
		url := base + "/latency.txt"
		sum = 0
		for range []int{1, 2} {
			ts := time.Now()
			_, err := getData(url)
			if err != nil {
				sum = 100000.0
				break
			}
			elapsed := time.Since(ts)
			sum += elapsed.Seconds()
		}
		if sum/2 < latency {
			server = st.servers[i]
		}
	}
	return server
}

func (st *ST) download() {

}

func distance(cLon, cLat float64, server Server) float64 {
	server.Lon = server.Lon * math.Pi / 180
	server.Lat = server.Lat * math.Pi / 180

	deltaLon := server.Lon - cLon
	deltaLat := server.Lat - cLat

	hsinLat := math.Pow(math.Sin(deltaLat/2), 2)
	hsinLon := math.Pow(math.Sin(deltaLon/2), 2)

	a := hsinLat + math.Cos(server.Lat)*math.Cos(cLat)*hsinLon
	c := 2 * 3961 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return c
}

func (a byDistance) Len() int           { return len(a) }
func (a byDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }
