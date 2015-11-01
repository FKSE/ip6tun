IPv6 Tunnel Broker
===========================

**Version 1.0.0-alpha**

Overview
--------

Usecase-Scenario: You have a IPv6-Only or Dual-Stack-Lite connection at home and you want to access your devices at home from remote. This works unless you are using an IPv4-Only connection.

This tool is a lightweight solution for this problem. The only thing you need is a "real" Dual-Stack-Server.

This app does more or less the same as https://www.sixxs.net/main/

### Use-Case

Without Tunnel

```
[Work] --- IPv4 ----> [Internet] --- IPv4 ---> [ISP] <No IPv4 port forwarning> [Home]
```

Using a Tunnel

```
[Work] --- IPv4 ----> [Internet] --- IPv4 ---> [Your dual stack server]--- IPv6 ---> [ISP] --- IPv6 ---> [Home]
```

But having a dynamic IPv6-Prefix means you have to reconfigure the tunnel on your server every time your prefix changes (In my case every 24 hours).

This is pretty annoying so I started developing `ip6tun` which is a tunnel broker proving a tiny rest interface allowing

At home you access the internet using a dual stack lite connection. Now you wan't to access you NAS from work where you only have IPv4.

Usage
----

Example script for running the server:

```
#!/bin/bash

# Enable debug
export IP6TUN_DEBUG=true
# Port
export IP6TUN_PORT=8080
# Servername
export IP6TUN_SERVERNAME=myserver
# Api access key
export IP6TUN_APIKEY=notverysecurekey
# https config
export IP6TUN_TLS_CERT=cert.pem
export IP6TUN_TLS_KEY=key.pem
# run server
./ip6tun-server
```

Client:

```
IP6TUN_KEY=notverysecurekey IP6TUN_HOST=localhost IP6TUN_PORT=8080 ip6tun-client mynas 80 10001
```

This will create a tunnel from `myserver:100001` to `[whatever:ipv6:your:nas:has]:80`.

TODOS
-----

* More tests
* Improve tunnel updating on server side
* Better client implementation
* Persist tunnels (Server should recrate tunnels after crash or reboot)

License
-------

MIT see LICENSE