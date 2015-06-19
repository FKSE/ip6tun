Automatic IPv6 Tunnelbroker
===========================

**Version 0.0.1**

Overview
--------

Usecase-Scenario: You have a IPv6-Only or Dual-Stack-Lite connection at home and you want to access your devices at home from remote. This works unless you are using an IPv4-Only connection.

This tool is a lightweight solution for this problem. The only thing you need is a real Dual-Stack-Server. The tool works as it follows:

1. Client sends a "Tunnel-Request" with a local and a remote port
2. Server tries to create a 4in6 tunnel
3. Server tells client the result

###Example

Somewhere you have a server which is reachable over IPv4 and IPv6 (Host: *thats.your.server*). At home you have only Dual Stack Lite, which means you have no exclusive public IPv4 (If you are here you should already know this). But now you want to make your You have a Raspberry-PI running a Website