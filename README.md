# phosport

phosport (So*phos* Ex*port*) is a Go utility to query the built-in confd storage of [Sophos UTM](https://www.sophos.com/en-us/products/unified-threat-management.aspx) appliances and export 
the set of network objects and packet filter rules along with the respective network and service objects to a JSON format. The internal
reference objects are resolved and output as plain IPs or network/netmask combinations. 

## Usage

There are two query modes implemented: If a `-host` flag is given that is not `localhost`, phosport will connect to the
target host via SSH (repeatedly; please set-up public key authentication beforehand).

Otherwise, `confd-client.plx` will be executed on the local machine, assuming you have copied the binary to the target UTM (e.g.
via scp). Local mode is obviously fastest, but with [SSH Multiplexing](http://en.wikibooks.org/wiki/OpenSSH/Cookbook/Multiplexing) remote execution is pretty fast, too. (It all depends on the size of the confd-database, of course.)

If you specify the `-v` command-line flag, all commands executed by phosport will be echoed to the terminal as they're triggerd. This will mainly consist of `get_object` calls as phosport initially fetches all packet filter rules and then starts resolving objects as it goes along.

~~~
  -host="localhost": UTM Hostname
  -v=false: Output all commands executed on UTM
~~~

Output modes:

~~~
  -hostgroups    This output mode exports all host groups
  -hosts         This output mode exports all hosts
  -rules         This output mode exports all packetfilter rules
  -servicegroups This output mode exports all service groups
  -services      This output mode exports all services
~~~

The special `-resolve` argument only works with the `-rules` output mode and resolves all internal REFs to the corresponding objects and also all groups to their respective members.

## Caveats

* If your rulebase uses dynamic user objects those will be output with their respective names as there is no IP attached to those objects

## Sample Output for `-rules -resolve`

~~~json
[
  {
    "action": "accept",
    "comment": "",
    "destinations": [
      "0.0.0.0"
    ],
    "group": "Outbound",
    "interface": "",
    "name": "Any from Topmedia Gast (Network) to Internet",
    "services": [
      "1:65535"
    ],
    "sources": [
      "172.16.28.0/24"
    ],
    "status": 1
  }
]
~~~

## Sample Output for `-hosts`

~~~
[
  {
    "address": "172.17.198.3",
    "comment": "",
    "hostnames": [
      "xn2a8124f79158b7"
    ],
    "macs": [
      "6c:98:eb:00:0b:40"
    ],
    "name": "Ocedo G50",
    "netmask": 0,
    "interface": "",
    "ref": "REF_NetHosOcedoG50"
  }
]
~~~
 
## Sample Output for `-services`

~~~
  {
    "dst_high": 3306,
    "dst_low": 3306,
    "comment": "MySQL Database Service",
    "name": "MySQL",
    "src_high": 65535,
    "src_low": 1,
    "ref": "REF_sDdolLNxJK",
    "protocol": "tcp"
  }
]
~~~

## Building

Assuming a proper `GOPATH` is set, building phosport is mainly a matter of executing:

~~~
go get github.com/topmedia/phosport
~~~

If you cloned the repository manually, fire off a build via:

~~~
go build
~~~

This will create an executable called `phosport` in the current directory, built for the current architecture. 

If you'd like to build phosport for Linux on a non-Linux platform (because you want to use local mode for example), set-up 
Go for [cross-compilation](http://dave.cheney.net/2015/03/03/cross-compilation-just-got-a-whole-lot-better-in-go-1-5) and
run the following in the checked out repository:

~~~
GOOS=linux go build
~~~

Then use `scp` to transfer the resulting `phosport` binary to the target UTM.
