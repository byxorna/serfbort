# serfbort

Screwing around with serf for a more secure, distributed deploy system for app code

![serfbort](https://36.media.tumblr.com/98bcbec7cca08c51a417680069316eb9/tumblr_nu25k8iIWZ1upb1bro1_1280.jpg)

***NOTE*** This is a WIP, and I am most likely going to get bored of it and forget about it after 6 hours of work, and leave it in a broken, halfassed, unfinished state. Whatevs :)

# Ideal features

- support labeling hosts (weba, webb, canary)
- support deploy to individual nodes
- support deploy to tagged subsets
- rotatable asymetric keys for deploy (i.e. if coordination node is compromised)
- report version for app via external command (git rev-parse, whatever)
- trigger deploy only from host holding master keys (i.e. coordination node listens for RPC, but not webs)
- support multiple "applications" (app, config, creds) as separate deploys

# Examples

TODO

# TODO

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

## With docker-compose

On OSX, if you are using `docker-machine` and have `docker-compose` installed...

Create a docker-compose.yml with the helper script:
```
$ ./create-docker-compose.rb 4 > docker-compose.yml
```
Now run `docker-compose`:
```
$ eval "$(docker-machine env dev)"
$ docker-compose build
...
$ docker-compose up
...
```

If you want to hook into the network namespace of this compose cluster, make sure to add `--net container:serfbort_serfnet_1` to your `docker run` command, like this:
```
$ docker run -ti --net container:serfbort_serfnet_1 serfbort_master cluster-status
4/4 hosts matching map[] []
Name    Addr    Tags                    Status
agent0  :::7947 map[]                   alive
master  :::7946 map[master:true]        alive
agent2  :::7949 map[]                   alive
agent1  :::7948 map[]                   alive
$ docker run -ti --net container:serfbort_serfnet_1 serfbort_master -config examples/config.json verify app 123
...
```

