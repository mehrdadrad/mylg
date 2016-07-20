# myLG [![Build Status](https://travis-ci.org/mehrdadrad/mylg.svg?branch=master)](https://travis-ci.org/mehrdadrad/mylg) [![Go Report Card](https://goreportcard.com/badge/github.com/mehrdadrad/mylg)](https://goreportcard.com/report/github.com/mehrdadrad/mylg) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/mehrdadrad/mylg?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) 

###Command line Network Diagnostic Tool
[![IMAGE ALT TEXT HERE](https://img.youtube.com/vi/jQJWcnLH3Fg/0.jpg)](https://www.youtube.com/watch?v=jQJWcnLH3Fg)

## Features
* Popular looking glasses (ping/trace/bgp) like Telia, Level3
* More than 200 countries DNS Lookup information 
* Local fast ping and trace
* Local HTTP/HTTPS ping (HEAD, GET)
* RIPE information (ASN, IP/CIDR)
* PeeringDB information
* Port scanning fast
* Support vi and emacs mode, almost all basic features
* CLI auto complete and history features

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
================== myLG v0.1.8 ==================
local> hping www.google.com -c 5
HPING www.google.com (172.217.4.164), Method: HEAD, DNSLookup: 19.0237 ms
HTTP Response seq=0, proto=HTTP/1.1, status=200, time=90.134 ms
HTTP Response seq=1, proto=HTTP/1.1, status=200, time=62.478 ms
HTTP Response seq=2, proto=HTTP/1.1, status=200, time=65.311 ms
HTTP Response seq=3, proto=HTTP/1.1, status=200, time=70.106 ms
HTTP Response seq=4, proto=HTTP/1.1, status=200, time=62.913 ms
local> 
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
80/tcp open
443/tcp open
Scan done: 2 opened port(s) found in 5.438 seconds
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

local> peering 577
The data provided from www.peeringdb.com
+----------------------+---------+------+--------------------+------+
|         NAME         | TRAFFIC | TYPE |      WEB SITE      | NOTE |
+----------------------+---------+------+--------------------+------+
| Bell Canada Backbone |         | NSP  | http://www.bell.ca |      |
+----------------------+---------+------+--------------------+------+
+-------------------+--------+-------+-----------------+------------------------+
|       NAME        | STATUS | SPEED |    IPV4 ADDR    |       IPV6 ADDR        |
+-------------------+--------+-------+-----------------+------------------------+
| Equinix Ashburn   | ok     | 20000 | 206.126.236.203 | 2001:504:0:2::577:1    |
| Equinix Chicago   | ok     | 20000 | 206.223.119.66  | 2001:504:0:4::577:1    |
| Equinix Palo Alto | ok     | 10000 | 198.32.176.94   | 2001:504:d::5e         |
| Equinix New York  | ok     | 10000 | 198.32.118.113  | 2001:504:f::577:1      |
| SIX Seattle       | ok     | 10000 | 206.81.80.217   | 2001:504:16::241       |
| NYIIX             | ok     | 10000 | 198.32.160.36   | 2001:504:1::a500:577:1 |
+-------------------+--------+-------+-----------------+------------------------+
```

## Contribute 
Welcomes any kind of contribution, please follow the next steps:

- Fork the project on github.com.
- Create a new branch.
- Commit changes to the new branch.
- Send a pull request.
