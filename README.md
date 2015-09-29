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

* agents arent responding to messages properly. makes me think that the master's agent ipc isnt sending queries along... (it does forward events!)

* make agents respond to deploy messages properly
* make agents respond to verify messages properly
* convert deploy/verify from events to queries so they can return statuses
* implement tag filtering for messages from deploy messages (encode target tags in payload?)
* hook shutdown properly for master+agent so they send leave messages
* make agents rejoin cluster (tune settings for rejoin?)


* slaves get batched messages, and delayed by a bunch of time
* lots of `[ERR] memberlist: Failed to send ping: write udp [::]:7947->[::]:7946: sendto: no route to host` on agents
* make agents and masters use keys for encryption


# Devving

## Locally

You need go 1.4.2, because `serf` isnt happy with 1.5.x (yet!). To build, install dependent go tools and do the build with `make`:

```
$ make setup
$ make all
```

## With Otto

Install `otto`: https://www.ottoproject.io/downloads.html

```
$ otto compile
$ otto dev
```
