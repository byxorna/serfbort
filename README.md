# serfbort

Screwing around with serf for a more secure, distributed deploy system for app code

***NOTE*** This is a WIP, and I am most likely going to get bored of it and forget about it after 6 hours of work, and leave it in a broken, halfassed, unfinished state. Whatevs :)

# Ideal features

- support labeling hosts (weba, webb, canary)
- support deploy to individual nodes
- support deploy to tagged subsets
- rotatable asymetric keys for deploy (i.e. if coordination node is compromised)
- report version for app via external command (git rev-parse, whatever)
- trigger deploy only from host holding master keys (i.e. coordination node listens for RPC, but not webs)
- support multiple "applications" (app, config, creds) as separate deploys

## Behaviors

`deploy target [selector or hostnames] version` - tells agents matching the selector or hostnames to deploy `version` of application `target`.
`check target [version]` - Queries the deploy at `target` and either returns its version, or "OK" if the optional `version` matches or ...?
`list [selector of hostnames]` - Check status of hosts in cluster matching selector

# TODO

* slaves get batched messages, and delayed by a bunch of time
* lots of `[ERR] memberlist: Failed to send ping: write udp [::]:7947->[::]:7946: sendto: no route to host` on agents
* make agents and masters use keys for encryption

# Notes

```
-> $ godep go build && ./serfbort  master
2015/09/27 15:20:36 [WARN] memberlist: Binding to public address without encryption!
2015/09/27 15:20:36 [INFO] serf: EventMemberJoin: tumblr-MacBookPro-b8cf72.gateway.pace.com ::
2015/09/27 15:20:36 1 nodes currently in cluster:
2015/09/27 15:20:36   tumblr-MacBookPro-b8cf72.gateway.pace.com :::7946 map[role:web env:dev] alive
2015/09/27 15:20:36 Running...

-> $ ./serfbort -master localhost -listen localhost:7947
2015/09/27 15:20:38 [WARN] memberlist: Binding to public address without encryption!
2015/09/27 15:20:38 [INFO] serf: EventMemberJoin: tumblr-MacBookPro-b8cf72.gateway.pace.com ::
2015/09/27 15:20:38 Joining localhost
2015/09/27 15:20:38 [DEBUG] memberlist: TCP connection from: [::1]:57915
2015/09/27 15:20:38 [DEBUG] memberlist: Initiating push/pull sync with: [::1]:7947
2015/09/27 15:20:38 [DEBUG] memberlist: TCP connection from: 127.0.0.1:57916
2015/09/27 15:20:38 [DEBUG] memberlist: Initiating push/pull sync with: 127.0.0.1:7947
2015/09/27 15:20:38 joined cluster with master localhost and 2 nodes
2015/09/27 15:20:38 1 nodes currently in cluster:
2015/09/27 15:20:38   tumblr-MacBookPro-b8cf72.gateway.pace.com :::7947 map[role:web env:dev] alive
2015/09/27 15:20:38 Running...
```
