# myLG [![Build Status](https://travis-ci.org/mehrdadrad/mylg.svg?branch=master)](https://travis-ci.org/mehrdadrad/mylg) [![Go Report Card](https://goreportcard.com/badge/github.com/mehrdadrad/mylg)](https://goreportcard.com/report/github.com/mehrdadrad/mylg) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/mehrdadrad/mylg?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) 

###Command line Network Diagnostic Tool
myLG, my looking glass is software utility which combines the functions of the different network probes in one network diagnostic tool.


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

local> hping www.google.com -c 5
HPING www.google.com (172.217.4.164), Method: HEAD, DNSLookup: 19.0237 ms
HTTP Response seq=0, proto=HTTP/1.1, status=200, time=90.134 ms
HTTP Response seq=1, proto=HTTP/1.1, status=200, time=62.478 ms
HTTP Response seq=2, proto=HTTP/1.1, status=200, time=65.311 ms
HTTP Response seq=3, proto=HTTP/1.1, status=200, time=70.106 ms
HTTP Response seq=4, proto=HTTP/1.1, status=200, time=62.913 ms
 
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

ns/united kingdom/manchester> dig yahoo.com
Trying to query server: 80.84.72.20 united kingdom manchester
Query time: 0.2369 ms
yahoo.com.	103	IN	AAAA	2001:4998:c:a06::2:4008
yahoo.com.	103	IN	AAAA	2001:4998:44:204::a7
yahoo.com.	103	IN	AAAA	2001:4998:58:c02::a9
yahoo.com.	518	IN	A	98.139.183.24
yahoo.com.	518	IN	A	206.190.36.45
yahoo.com.	518	IN	A	98.138.253.109
yahoo.com.	161748	IN	NS	ns1.yahoo.com.
yahoo.com.	161748	IN	NS	ns2.yahoo.com.
yahoo.com.	161748	IN	NS	ns3.yahoo.com.
yahoo.com.	161748	IN	NS	ns4.yahoo.com.
yahoo.com.	161748	IN	NS	ns5.yahoo.com.
yahoo.com.	161748	IN	NS	ns6.yahoo.com.

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

local> whois 8.8.8.8
+------------+-------+--------------------------+
|   PREFIX   |  ASN  |          HOLDER          |
+------------+-------+--------------------------+
| 8.8.8.0/24 | 15169 | GOOGLE - Google Inc., US |
+------------+-------+--------------------------+

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

```

## Contribute 
Welcomes any kind of contribution, please follow the next steps:

- Fork the project on github.com.
- Create a new branch.
- Commit changes to the new branch.
- Send a pull request.
