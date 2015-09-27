# serfbort

Screwing around with serf for a more secure, distributed deploy system for app code

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

