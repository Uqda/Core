# UQDA Latency and Performance

UQDA should keep overlay overhead small and choose good available paths, but it
cannot guarantee a fixed Internet round-trip time. Physical distance, the
underlay route, access technology, Wi-Fi contention, congestion, and the
remote host all contribute to RTT.

## A measurable objective

For two devices in the same home, venue, or nearby metro area, a useful target
is zero packet loss, idle p95 RTT at or below 20 ms, low successive-sample
jitter, and no large RTT increase while the WAN is loaded. Use p95 over at
least 20 packets; one successful ping does not describe stability.

For geographically remote nodes, record the measured baseline and set a
realistic target instead of labeling unavoidable distance delay as a software
failure.

## Measure

Stable releases starting with v0.1.2 publish the helper as a checksum-covered
release asset. It can also be installed directly from the reviewed source:

```sh
curl -fsSLo uqda-latency \
  https://raw.githubusercontent.com/Uqda/Core/main/contrib/performance/uqda-latency
chmod +x uqda-latency
sudo install -m 0755 uqda-latency /usr/local/bin/uqda-latency
```

Measure an UQDA address:

```sh
uqda-latency 21c:f3b0:e941:88bc:2938:a694:9012:37f2
```

It reports loss, minimum, average, p50, p95, maximum and jitter. It succeeds
only when every sample returns and p95 meets the default 20 ms target. Use a
different explicit target for a remote site:

```sh
uqda-latency 21c:f3b0:e941:88bc:2938:a694:9012:37f2 --count 50 --target-ms 40
```

List direct peers from lowest measured RTT to highest:

```sh
sudo uqda-latency peers
# equivalent:
sudo uqdactl getPeers sort=latency
```

## How routing already uses latency

The embedded Ironwood router measures direct-link RTT with signed request and
response traffic, smooths it with an exponentially weighted average, and uses
the resulting link cost when selecting a loop-free next hop. New links receive
a temporary stability penalty so a tiny sample change does not continuously
move routes.

Adding many distant peers can make performance worse. Prefer a small set of
reliable peers near the device or directly peer the two sites that need low
latency. The `priority` URI option selects between duplicate links to the same
public key; it is not an end-to-end latency guarantee.

## Diagnose a missed target

1. Measure while the Internet connection is idle.
2. Compare with `uqdactl getPeers sort=latency`.
3. Measure again while uploading and downloading heavily.
4. Test by Ethernet to remove Wi-Fi contention.
5. Compare ordinary underlay RTT to the direct peer host with UQDA RTT.

| Observation | Likely cause | Action |
|---|---|---|
| Direct peer RTT is already above target | distance or ISP route | use a closer peer/direct site or raise the target |
| Idle is good, loaded p95 is high | bufferbloat | enable SQM/CAKE on the actual WAN router using measured rates |
| Ethernet is good, Wi-Fi is bad | radio contention | use 5/6 GHz, a clearer channel, better AP placement, or Ethernet |
| Direct peer is good, destination is slow | overlay path/transit | add a reliable direct peer where operationally possible |
| Loss appears first | link quality or overload | fix signal, cabling, CPU saturation, or upstream loss |

Do not apply a guessed bandwidth shaper. SQM must sit at the real bottleneck,
and its rates should be slightly below measured stable capacity. A shaper on
the wrong interface can reduce throughput without controlling the queue.

## Gateway recommendations

- use Ethernet for WAN and reserve Wi-Fi for clients;
- avoid CPU-heavy logging on small routers;
- keep the validated platform MTU unless packet-size tests prove a PMTU issue;
- use a wired management path during tuning; and
- compare idle and loaded p95 after firmware or ISP changes.

For OpenWrt, use its maintained SQM packages and configure CAKE on the true WAN
interface. For Raspberry Pi/Linux behind another router, shaping the Pi cannot
fully control a queue already building in the upstream router.
