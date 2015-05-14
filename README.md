# phosport

phosport (So*phos* Ex*port*) is a Go utility to query the built-in confd storage of [Sophos UTM](https://www.sophos.com/en-us/products/unified-threat-management.aspx) appliances and export 
the set of packet filter rules along with the respective network and service objects to a JSON format. The internal
reference objects are resolved and output as plain IPs or network/netmask combinations. 

## Usage

There are two modes implemented: If a `-host` flag is given that is not `localhost`, phosport will connect to the
target host via SSH (repeatedly; please set-up public key authentication beforehand).

Otherwise, `confd-client.plx` will be executed on the local machine, assuming you have copied the binary to the target UTM (e.g.
via scp). Local mode is obviously fastest, but with [SSH Multiplexing](http://en.wikibooks.org/wiki/OpenSSH/Cookbook/Multiplexing) remote execution is pretty fast, too. (It all depends on the size of the confd-database, of course.)

If you specify the `-v` command-line flag, all commands executed by phosport will be echoed to the terminal as they're triggerd. This will mainly consist of `get_object` calls as phosport initially fetches all packet filter rules and then starts resolving objects as it goes along.

~~~
  -host="localhost": UTM Hostname
  -v=false: Output all commands executed on UTM
~~~

## Caveats

* If your rulebase uses dynamic user objects those will be output with their respective names as there is no IP attached to those objects

## Sample Output

~~~json
[
  {
    "sources": [
      "10.0.0.200"
    ],
    "destinations": [
      "192.168.1.12"
    ],
    "services": [
      "5060"
    ]
  },
  {
    "sources": [
      "192.168.1.0/24",
      "172.16.198.0/24"
    ],
    "destinations": [
      "192.168.1.0/24",
      "172.17.198.0/24",
      "172.21.16.0/24",
      "172.16.198.0/24"
    ],
    "services": [
      "1:65535"
    ]
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
