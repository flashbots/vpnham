# vpnham

High-availability monitor for VPN tunnels.

## TL;DR

```text
Side A                                               Side B

    +----------+           active          +----------+
    |  Active  | <=======================> |  Active  |
    +----------o                           x----------+
       .           o                   x           .
       .               o           x               .
       .                   o   x                   .
       .                   x   o                   .
       .     backup    x           o    backup     .
       .           x                   o           .
    +----------x                           o----------+
    |  Backup  | < - - - - - - - - - - - > |  Backup  |
    +----------+           backup          +----------+
```

## Details

Typical setup would be:

- 2x VPN hosts on each side of the bridge.
- One host on each side configured as `active`, the other as `standby`.
- Each host on each side has VPN tunnels configured to _both_ of the other side
  hosts.
- Only `active`-`active` tunnel is used, the others are there for backup.

### Logic

In the event of the downtime (tunnel is broken, one of active hosts is down),
the monitor would:

- Promote the remaining tunnel to become `active`.
- Trigger custom command scripts to adjust for the situation (to re-configure
  the routes, for example).

In order to achieve that, `vpnham` does the following:

- Regularly send UDP-datagrams to the each of the peers on the other side
  (connections `<==>`, `< - >`, `< x >`, and `< o >` on the diagram above).
  - The peer adds its bit into the datagram and sends it back.
  - This is how `vpnham` determines the `up`/`down` status of the tunnel (it
    accounts for the sent probes with their sequence numbers, and expects them
    to come back).

- Regularly poll the partner's bridge (i.e. the `active` bridge polls the
  `standby` one, and vice versa;  connections `< . >` above).
  - Failure to poll means the partner is `down`.
  - If both of the tunnels are `down`, the bridge marks itself `down` as well
    and reports itself accordingly to its partner.

- If `active` tunnel is `down`, the `standby` is promoted to `active`.

- If `active` bridge is `down`, the `standby` is promoted to `active`.

- Once the `active` (by configuration) tunnel gets back `up`, it will reclaim
  the `active` status (in other words, the ties are broken via configuration).

- Similar approach with the bridges.

### Scripts

There are configurable scripts (per bridge, or globally):

- `bridge_activate` is triggered when a bridge is promoted to `active`.
  - Recognised placeholders are:
    - `${proto}`
    - `${bridge_peer_cidr}`
    - `${bridge_interface}`
    - `${bridge_interface_ip}`

- `tunnel_activate` is triggered when a tunnel is marked `active`.
  - Recognised placeholders are the same as for `bridge_activate`, plus:
    - `${tunnel_interface}`
    - `${tunnel_interface_ip}`

- `tunnel_deactivate` is triggered when the tunnel's `active` mark is removed.
  - Recognised placeholders are the same as for `tunnel_activate`

### Metrics

In addition, there is a metrics endpoint where `vpnham` reports the following:

- `vpnham_bridge_active` is a gauge for count of active bridges.
  - `0` is when no connectivity to the other side (bad).
  - `1` is when all is good (yay).
  - `2` is when both us and the partner consider themselves `active`
    (this means bug).

- `vpnham_bridge_up` is a gauge for the count of online bridges
  (from `0` to `2`, the more the merrier).

- `vpnham_tunnel_interface_active` is a gauge for count of active tunnels.

- `vpnham_tunnel_interface_up` is a gauge for count of online tunnels.

Also (since we have that info at our fingertips through probing), the following
metrics are exposed:

- `vpnham_probes_sent_total` is a counter for probes sent.

- `vpnham_probes_returned_total` is a counter for probes returned.

- `vpnham_probes_failed_total` is a counter for probes failed to send, or to
  receive.

- `vpnham_probes_latency_forward_microseconds` is a histogram for the probes
  forward latency (on their trip "there").

- `vpnham_probes_latency_return_microseconds` is a histogram for the probe
  return latency (trip "back")

## Example

```yaml
bridges:
  vpnham-dev-lft:  # bridge name (must match one at the partner's side)
    role: active   # our role (`active` or `standby`)

    bridge_interface: eth0  # interface on which the bridge connects to VPC
    peer_cidr: 10.1.0.0/16  # CIDR range of the VPC we are bridging into

    status_addr: 10.0.0.2:8080                 # address where our partner polls our status
    partner_status_url: http://10.0.0.3:8080/  # url where we poll the status of the partner

    probe_interval: 1s           # interval between UDP probes or status polls
    probe_location: left/active  # location label for the latency metrics

    tunnel_interfaces:
      eth1:           # interface on which VPN tunnel is running
        role: active  # tunnel role (`active` or `standby`)

        addr: 192.168.255.2:3003        # address where we respond to UDP probes
        probe_addr: 192.168.255.3:3003  # address where we send the UDP probes to

        threshold_down: 5  # count of failed probes/polls to mark peer/partner "down"
        threshold_up: 3    # count of successful probes/polls to mark peer/partner "up"

      eth2:
        role: standby
        addr: 192.168.255.18:3003
        probe_addr: 192.168.255.19:3003
        threshold_down: 7
        threshold_up: 5

    scripts_timeout: 5s  # max amount of time for script commands to finish

metrics:
  listen_addr: 0.0.0.0:8000  # where we expose the metrics (at `/metrics` path)

  latency_buckets_count: 33  # count of histogram buckets for latency metrics

  max_latency_us: 1000000  # max latency bucket in [us]; the buckets are computed
                           # exponentially, so that
                           # max_latency == pow(min_latency, buckets_count)

default_scripts:    # default scripts (complement the `scripts` on bridge config)
  bridge_activate:  # script that we will run when bridge becomes `active`
    - ["sh", "-c", "echo ${bridge_interface} ${bridge_interface_ip} ${bridge_peer_cidr}"]
    - ["sleep", "15"]

  interface_activate:  # script that we will run when tunnel becomes `active`
    - ["sh", "-c", "echo 'activate ${tunnel_interface} ${tunnel_interface_ip}'"]

  interface_deactivate:  # script that we will run when tunnel becomes `inactive`
    - ["sh", "-c", "echo 'deactivate ${tunnel_interface} ${tunnel_interface_ip}'"]
```

>
> [!NOTE]
>
> See the following files for the full example:
>
> - [docker-compose.yaml](./docker-compose.yaml)
> - [br-lft-act.yaml](./.dev/br-lft-act.yaml)
> - [br-lft-sby.yaml](./.dev/br-lft-sby.yaml)
> - [br-rgt-act.yaml](./.dev/br-rgt-act.yaml)
> - [br-rgt-sby.yaml](./.dev/br-rgt-sby.yaml)
>
> Also: `make docker-compose`

## CLI

`vpnham` takes only one cli-parameter `--config` that should point to the
yaml-file with full configuration.  By default it will seek `.vpnham.yaml` file
in the working directory.
