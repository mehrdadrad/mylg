#[![Build Status](https://travis-ci.org/mehrdadrad/mylg.svg?branch=master)](https://travis-ci.org/mehrdadrad/mylg) [![Go Report Card](https://goreportcard.com/badge/github.com/mehrdadrad/mylg)](https://goreportcard.com/report/github.com/mehrdadrad/mylg) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/mehrdadrad/mylg?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) 

![IMAGE](http://mylg.io/wp-content/uploads/2016/08/logo_mylgio_xxsmall.png)
###myLG, Command line Network Diagnostic Tool
my looking glass is open source software utility which combines the functions of the different network probes in one network diagnostic tool.

## Features
* Popular looking glasses (ping/trace/bgp) like Telia, Level3
* More than 200 countries DNS Lookup information 
* Local fast ping and trace
* Packet analyzer - TCP/IP and other packets 
* Local HTTP/HTTPS ping (GET, POST, HEAD)
* RIPE information (ASN, IP/CIDR)
* PeeringDB information
* Port scanning fast
* Network LAN Discovery
* Web dashboard
* Support vi and emacs mode, almost all basic features
* CLI auto complete and history features

![IMAGE ALT TEXT HERE](http://mylg.io/img/packet_analyzer.png)
![IMAGE ALT TEXT HERE](http://mylg.io/img/mylg_dashboard.png)

### Usage

```
=================================================	
                          _    ___ 
                _ __ _  _| |  / __|
               | '  \ || | |_| (_ |
               |_|_|_\_, |____\___|
                      |__/          
	
                 My Looking Glass
           Free Network Diagnostic Tool
             www.facebook.com/mylg.io
                  http://mylg.io
================== myLG v0.2.0 ==================

local> hping google.com -c 5
HPING google.com (216.58.216.14), Method: HEAD, DNSLookup: 44.3652 ms
HTTP Response seq=0, proto=HTTP/1.1, status=200, time=476.465 ms
HTTP Response seq=1, proto=HTTP/1.1, status=200, time=279.403 ms
HTTP Response seq=2, proto=HTTP/1.1, status=200, time=186.390 ms
HTTP Response seq=3, proto=HTTP/1.1, status=200, time=279.451 ms
HTTP Response seq=4, proto=HTTP/1.1, status=200, time=249.007 ms

--- google.com HTTP ping statistics --- 
5 requests transmitted, 5 replies received, 0% timeout
HTTP Round-trip min/avg/max = 186.39/264.91/476.46 ms
HTTP Code [200] responses : [████████████████████] 100.00%
 
local> whois 577
BACOM - Bell Canada, CA
+--------------------+-----------+
|      LOCATION      | COVERED % |
+--------------------+-----------+
| Canada - ON        |   61.3703 |
| Canada             |   36.2616 |
| Canada - QC        |    1.3461 |
| United States - MA |    0.7160 |
| Canada - BC        |    0.1766 |
| Canada - AB        |    0.0811 |
| United States      |    0.0195 |
| United States - NJ |    0.0143 |
| Belgium            |    0.0048 |
| United States - NC |    0.0048 |
| United States - TX |    0.0048 |
| Canada - NB        |    0.0000 |
| Canada - NS        |    0.0000 |
+--------------------+-----------+

local> scan www.google.com -p 1-500
+----------+------+--------+-------------+
| PROTOCOL | PORT | STATUS | DESCRIPTION |
+----------+------+--------+-------------+
| TCP      |   80 | Open   |             |
| TCP      |  443 | Open   |             |
+----------+------+--------+-------------+
Scan done: 2 opened port(s) found in 5.605 seconds

lg/telia/los angeles> bgp 8.8.8.0/24
Telia Carrier Looking Glass - show route protocol bgp 8.8.8.0/24 table inet.0

Router: Los Angeles

Command: show route protocol bgp 8.8.8.0/24 table inet.0

inet.0: 661498 destinations, 5564401 routes (657234 active, 509 holddown, 194799 hidden)
+ = Active Route, - = Last Active, * = Both

8.8.8.0/24         *[BGP/170] 33w0d 01:36:06, MED 0, localpref 200
                      AS path: 15169 I, validation-state: unverified
                    > to 62.115.36.170 via ae4.0
                    [BGP/170] 8w3d 11:19:40, MED 0, localpref 200, from 80.91.255.95
                      AS path: 15169 I, validation-state: unverified
                      to 62.115.119.84 via xe-1/1/0.0
                      to 62.115.119.88 via xe-1/2/0.0
                      to 62.115.119.90 via xe-11/0/3.0
                      to 62.115.119.102 via xe-9/0/0.0
                      to 62.115.119.92 via xe-9/0/2.0
                    > to 62.115.119.86 via xe-9/1/2.0
                      to 62.115.119.98 via xe-9/2/2.0
                      to 62.115.119.100 via xe-9/2/3.0
                      to 62.115.119.94 via xe-9/3/1.0
                      to 62.115.119.96 via xe-9/3/3.0

ns/united states/redwood city> dig yahoo.com
Trying to query server: 204.152.184.76 united states redwood city
;; opcode: QUERY, status: NOERROR, id: 19850
;; flags: qr rd ra;
yahoo.com.	728	IN	MX	1 mta6.am0.yahoodns.net.
yahoo.com.	728	IN	MX	1 mta5.am0.yahoodns.net.
yahoo.com.	728	IN	MX	1 mta7.am0.yahoodns.net.
yahoo.com.	143013	IN	NS	ns4.yahoo.com.
yahoo.com.	143013	IN	NS	ns6.yahoo.com.
yahoo.com.	143013	IN	NS	ns2.yahoo.com.
yahoo.com.	143013	IN	NS	ns5.yahoo.com.
yahoo.com.	143013	IN	NS	ns1.yahoo.com.
yahoo.com.	143013	IN	NS	ns3.yahoo.com.

;; ADDITIONAL SECTION:
ns1.yahoo.com.	561456	IN	A	68.180.131.16
ns2.yahoo.com.	27934	IN	A	68.142.255.16
ns3.yahoo.com.	532599	IN	A	203.84.221.53
ns4.yahoo.com.	532599	IN	A	98.138.11.157
ns5.yahoo.com.	532599	IN	A	119.160.247.124
ns6.yahoo.com.	143291	IN	A	121.101.144.139
ns1.yahoo.com.	51624	IN	AAAA	2001:4998:130::1001
ns2.yahoo.com.	51624	IN	AAAA	2001:4998:140::1002
ns3.yahoo.com.	51624	IN	AAAA	2406:8600:b8:fe03::1003
ns6.yahoo.com.	143291	IN	AAAA	2406:2000:108:4::1006
;; Query time: 1204 ms

;; CHAOS CLASS BIND
version.bind.	0	CH	TXT	"9.10.4-P1"
hostname.bind.	0	CH	TXT	"fred.isc.org"

local> peering 6327
The data provided from www.peeringdb.com
+-------------------+---------------+---------------+--------------------+------+
|       NAME        |    TRAFFIC    |     TYPE      |      WEB SITE      | NOTE |
+-------------------+---------------+---------------+--------------------+------+
| Shaw Cablesystems | 500-1000 Gbps | Cable/DSL/ISP | http://www.shaw.ca |      |
+-------------------+---------------+---------------+--------------------+------+
+------------------+--------+--------+-----------------+-------------------------+
|       NAME       | STATUS | SPEED  |    IPV4 ADDR    |        IPV6 ADDR        |
+------------------+--------+--------+-----------------+-------------------------+
| Equinix Ashburn  | ok     |  20000 | 206.126.236.20  | 2001:504:0:2::6327:1    |
| Equinix Ashburn  | ok     |  20000 | 206.223.115.20  |                         |
| Equinix Chicago  | ok     |  30000 | 206.223.119.20  | 2001:504:0:4::6327:1    |
| Equinix San Jose | ok     |  30000 | 206.223.116.20  | 2001:504:0:1::6327:1    |
| Equinix Seattle  | ok     |  20000 | 198.32.134.4    | 2001:504:12::4          |
| Equinix New York | ok     |  10000 | 198.32.118.16   | 2001:504:f::10          |
| SIX Seattle      | ok     | 100000 | 206.81.80.54    | 2001:504:16::18b7       |
| NYIIX            | ok     |  20000 | 198.32.160.86   | 2001:504:1::a500:6327:1 |
| TorIX            | ok     |  10000 | 206.108.34.12   |                         |
| PIX Vancouver    | ok     |  10000 | 206.223.127.2   |                         |
| PIX Toronto      | ok     |   1000 | 206.223.127.132 |                         |
| Equinix Toronto  | ok     |  10000 | 198.32.181.50   | 2001:504:d:80::6327:1   |
+------------------+--------+--------+-----------------+-------------------------+

local> disc
Network LAN Discovery
+---------------+-------------------+------+-----------+--------------------------------+
|      IP       |        MAC        | HOST | INTERFACE |       ORGANIZATION NAME        |
+---------------+-------------------+------+-----------+--------------------------------+
| 192.168.0.1   | a4:2b:b0:eb:9a:b4 | NA   | en0       | TP-LINK TECHNOLOGIES CO.,LTD.  |
| 192.168.0.103 | ac:bc:32:b4:33:23 | NA   | en0       | Apple, Inc.                    |
| 192.168.0.105 | 40:b8:9a:60:55:9e | NA   | en0       | Hon Hai Precision Ind.         |
| 224.0.0.251   | 1:0:5e:0:0:fb     | NA   | en0       | NA                             |
+---------------+-------------------+------+-----------+--------------------------------+

local> whois 8.8.8.8
+------------+-------+--------------------------+
|   PREFIX   |  ASN  |          HOLDER          |
+------------+-------+--------------------------+
| 8.8.8.0/24 | 15169 | GOOGLE - Google Inc., US |
+------------+-------+--------------------------+

local> dump -d
+---------+-------------------+--------+-------+-----------+-----------+--------------+----------+
|  NAME   |        MAC        | STATUS |  MTU  | MULTICAST | BROADCAST | POINTTOPOINT | LOOPBACK |
+---------+-------------------+--------+-------+-----------+-----------+--------------+----------+
| lo0     |                   | UP     | 16384 | ✓         |           |              | ✓        |
| gif0    |                   | DOWN   |  1280 | ✓         |           | ✓            |          |
| stf0    |                   | DOWN   |  1280 |           |           |              |          |
| en0     | ac:bc:32:b4:33:23 | UP     |  1500 | ✓         | ✓         |              |          |
| en1     | 4a:00:03:9c:8d:60 | UP     |  1500 |           | ✓         |              |          |
| en2     | 4a:00:03:9c:8d:61 | UP     |  1500 |           | ✓         |              |          |
| p2p0    | 0e:bc:32:b4:33:23 | UP     |  2304 | ✓         | ✓         |              |          |
| awdl0   | 16:fe:c4:ab:2a:f9 | UP     |  1484 | ✓         | ✓         |              |          |
| bridge0 | ae:bc:32:4b:10:00 | UP     |  1500 | ✓         | ✓         |              |          |
| gpd0    | 02:50:41:00:01:01 | DOWN   |  1400 | ✓         | ✓         |              |          |
| utun0   |                   | UP     |  1500 | ✓         |           | ✓            |          |
+---------+-------------------+--------+-------+-----------+-----------+--------------+----------+

local> dump 
20:29:36.415 IPv4/TCP  ec2-52-73-80-145.compute-1.amazonaws.com.:443(https) > 192.168.0.104:61479 [P.], win 166, len: 33
20:29:36.416 IPv4/TCP  192.168.0.104:61479 > ec2-52-73-80-145.compute-1.amazonaws.com.:443(https) [.], win 4094, len: 0
20:29:36.417 IPv4/TCP  192.168.0.104:61479 > ec2-52-73-80-145.compute-1.amazonaws.com.:443(https) [P.], win 4096, len: 37
20:29:36.977 IPv4/UDP  192.168.0.104:62733 > 192.168.0.1:53(domain) , len: 0
20:29:37.537 IPv4/TCP  ec2-54-86-120-119.compute-1.amazonaws.com.:443(https) > 192.168.0.104:61302 [.], win 124, len: 0
20:29:38.125 IPv4/TCP  192.168.0.104:61304 > ec2-52-23-213-161.compute-1.amazonaws.com.:443(https) [P.], win 4096, len: 85
20:29:38.126 IPv4/TCP  ec2-52-23-213-161.compute-1.amazonaws.com.:443(https) > 192.168.0.104:61304 [.], win 1048, len: 0
20:29:38.760 IPv4/TCP  ec2-54-165-12-100.compute-1.amazonaws.com.:443(https) > 192.168.0.104:61296 [.], win 2085, len: 0
20:29:39.263 IPv4/ICMP 192.168.0.104 > ir1.fp.vip.ne1.yahoo.com.: EchoRequest id 20859, seq 27196, len: 56
20:29:39.265 IPv4/UDP  192.168.0.1:53(domain) > 192.168.0.104:62733 , len: 0

local> dump tcp and port 443 -c 10
23:26:56.026 IPv4/TCP  192.168.0.104:64686 > 192.0.80.242:443(https) [F.], win 8192, len: 0
23:26:56.045 IPv4/TCP  192.168.0.104:64695 > i2.wp.com.:443(https) [F.], win 8192, len: 0
23:26:56.048 IPv4/TCP  i2.wp.com.:443(https) > 192.168.0.104:64695 [F.], win 62, len: 0
23:26:56.081 IPv4/TCP  192.168.0.104:63692 > ec2-54-88-144-213.compute-1.amazonaws.com.:443(https) [P.], win 4096, len: 37
23:26:56.082 IPv4/TCP  192.168.0.104:64695 > i2.wp.com.:443(https) [.], win 8192, len: 0
23:26:56.083 IPv4/TCP  192.0.80.242:443(https) > 192.168.0.104:64686 [.], win 64, len: 0
23:26:56.150 IPv4/TCP  ec2-54-88-144-213.compute-1.amazonaws.com.:443(https) > 192.168.0.104:63692 [.], win 166, len: 0
23:26:56.259 IPv4/TCP  ec2-54-172-56-148.compute-1.amazonaws.com.:443(https) > 192.168.0.104:63623 [P.], win 1316, len: 85
23:26:56.260 IPv4/TCP  192.168.0.104:63623 > ec2-54-172-56-148.compute-1.amazonaws.com.:443(https) [.], win 4093, len: 0
23:26:56.820 IPv4/TCP  192.168.0.104:64691 > 192.30.253.116:443(https) [.], win 4096, len: 0

local> dump -s http -x
22:10:15.770 IPv4/TCP  151.101.44.143:443(https) > 10.0.9.9:50771 [P.], win 59, len: 156
00000000  16 03 03 00 64 02 00 00  60 03 03 a2 32 19 4b 78  |....d...`...2.Kx|
00000010  77 ed 40 75 f6 4c 55 74  43 1d b7 6c f2 59 f8 d8  |w.@u.LUtC..l.Y..|
00000020  09 8a 3e 03 62 56 38 45  d2 bc 02 20 bd 52 8a 42  |..>.bV8E... .R.B|
00000030  5b 01 33 7d 2b 0b 41 da  eb 38 87 79 f1 37 62 5c  |[.3}+.A..8.y.7b\|
00000040  f3 ed 5a 7c 07 6c e9 28  9b fe fa 76 c0 2f 00 00  |..Z|.l.(...v./..|
00000050  18 ff 01 00 01 00 00 05  00 00 00 10 00 0b 00 09  |................|
00000060  08 68 74 74 70 2f 31 2e  31 14 03 03 00 01 01 16  |.http/1.1.......|
00000070  03 03 00 28 fc 20 2d 6f  1a 94 78 53 55 0f 8c 05  |...(. -o..xSU...|
00000080  3e ae 12 34 79 af d2 a9  bd 22 e5 3f b1 2b f5 36  |>..4y....".?.+.6|
00000090  ba 51 31 37 f5 0b e6 d2  40 fb 88 a5              |.Q17....@...    |

local> dump !udp -w /home/user1/mypcap -c 100000

local> ping google.com -6
PING google.com (2607:f8b0:400b:80a::200e): 56 data bytes
64 bytes from 2607:f8b0:400b:80a::200e icmp_seq=0 time=23.193988 ms
64 bytes from 2607:f8b0:400b:80a::200e icmp_seq=1 time=21.265492 ms
64 bytes from 2607:f8b0:400b:80a::200e icmp_seq=2 time=24.521306 ms
64 bytes from 2607:f8b0:400b:80a::200e icmp_seq=3 time=25.313072 ms

local> trace google.com
trace route to google.com (172.217.4.142), 30 hops max
1  192.168.0.1 4.705 ms 1.236 ms 0.941 ms 
2  142.254.236.25 [ASN 20001/ROADRUNNER-WEST] 13.941 ms 13.504 ms 12.303 ms 
3  agg59.snmncaby01h.socal.rr.com. (76.167.31.241) [ASN 20001/ROADRUNNER-WEST] 14.834 ms 11.625 ms 13.050 ms 
4  agg20.lamrcadq01r.socal.rr.com. (72.129.10.128) [ASN 20001/ROADRUNNER-WEST] 17.617 ms 18.064 ms 15.612 ms 
5  agg28.lsancarc01r.socal.rr.com. (72.129.9.0) [ASN 20001/ROADRUNNER-WEST] 16.291 ms 24.079 ms 20.456 ms 
6  bu-ether26.lsancarc0yw-bcr00.tbone.rr.com. (66.109.3.230) [ASN 7843/TWCABLE-BACKBONE] 18.339 ms 23.278 ms 23.434 ms 
7  216.0.6.25 [ASN 2828/XO-AS15] 19.842 ms 21.025 ms 35.105 ms 
8  216.0.6.42 [ASN 2828/XO-AS15] 16.666 ms 18.252 ms 18.872 ms 
9  209.85.245.199 [ASN 15169/GOOGLE] 14.358 ms 17.478 ms 
   209.85.246.125 [ASN 15169/GOOGLE] 18.593 ms 
10 72.14.239.121 [ASN 15169/GOOGLE] 21.635 ms 
   72.14.238.213 [ASN 15169/GOOGLE] 16.133 ms 
   72.14.239.121 [ASN 15169/GOOGLE] 21.541 ms 
11 lax17s14-in-f14.1e100.net. (172.217.4.142) [ASN 15169/GOOGLE] 18.127 ms 17.151 ms 18.892 ms 
```
## Build
It can be built for Linux and Darwin. there is libpcap dependency:
```
- LINUX
apt-get install libpcap-dev
- OSX
brew install libpcap
```

Given that the Go Language compiler is installed, you can build and run:

```
git clone https://github.com/mehrdadrad/mylg.git
cd mylg
go get
go build mylg
```

## Licence
This project is licensed under MIT license. Please read the LICENSE file.


## Contribute 
Welcomes any kind of contribution, please follow the next steps:

- Fork the project on github.com.
- Create a new branch.
- Commit changes to the new branch.
- Send a pull request.
