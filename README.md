testdns
=======

testdns is a tool to help identify resolvers that give the correct response for a given hostname.

It's useful for:

* Finding authoritative DNS servers for a particular domain (detailed explanation
  here https://github.com/Abss0x7tbh/bass/blob/master/contrib/testdns/README.md)
* Finding/validating open resolvers. 

## Usage

The tool accepts the following flags:

- `-n` : Hostname to resolve. Default is "test-12-34-56-78.nip.io".
- `-w` : Number of concurrent workers. Default is 25.
- `-r` : Trusted resolver. Default is "1.1.1.1:53".
- `-s` : Show resolvers only. Default is is to include response time
- `-t` : Timeout for queries. Default is 2 seconds.
- `-i` : Show failing servers 

Use hostnames that resolve consistently or things will become confused.