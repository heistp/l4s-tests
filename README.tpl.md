# L4S Tests

Evidence in opposition to the L4S WGLC

Pete Heist  
Jonathan Morton  

## Table of Contents

1. [Introduction](#introduction)
2. [Key Findings](#key-findings)
3. [Elaboration on Key Findings](#elaboration-on-key-findings)
   1. [Network Bias](#network-bias)
   2. [RTT Unfairness](#rtt-unfairness)
   3. [Intra-flow Latency Spikes](#intra-flow-latency-spikes)
   4. [Burst Intolerance](#burst-intolerance)
   5. [Unsafety in Tunnels Through RFC3168 Bottlenecks](#unsafety-in-tunnels-through-rfc3168-bottlenecks)
4. [Full Results](#full-results)
   1. [Scenario 1: RTT Fairness](#scenario-1-rtt-fairness)
   2. [Scenario 2: Codel Rate Step](#scenario-2-codel-rate-step)
   3. [Scenario 3: Codel Variable Rate](#scenario-3-codel-variable-rate)
   4. [Scenario 4: Bi-directional Traffic, Asymmetric Rates](#scenario-4-bi-directional-traffic-asymmetric-rates)
   5. [Scenario 5: Tunnels](#scenario-5-tunnels)
5. [Appendix](#appendix)
   1. [Scenario 1 Fairness Table](#scenario-1-fairness-table)
   2. [Background](#background)
   3. [Deployments of fq_codel](#deployments-of-fq_codel)
   4. [Test Setup](#test-setup)

## Introduction

The Transport Area Working Group
([TSVWG](https://datatracker.ietf.org/group/tsvwg/about/)) will undergo a WGLC
(working group last call) for [L4S](https://riteproject.eu/dctth/), which
proposes to use the available ECT(1) codepoint for two purposes:

* to redefine the existing CE codepoint as a high-fidelity congestion control
  signal, which is incompatible with the present definition of CE in
  [RFC3168](https://tools.ietf.org/html/rfc3168) and
  [RFC8511](https://tools.ietf.org/html/rfc8511)
* as a PHB (per-hop behavior) to select alternate treatment in bottlenecks,
  giving some level of priority in DualPI2 queues for L4S traffic over
  existing Internet traffic

While safety concerns for existing flows have been covered before and have
not been addressed in the implementation (see
[issues](https://trac.ietf.org/trac/tsvwg/report/1) listed in TSVWG's trac),
these tests raise new fairness and performance concerns that should be
understood before L4S is considered for WGLC.

Readers wishing for a quick background in high-fidelity congestion control
may wish to read the [Background](#background) section, while those already
familiar with the topic can proceed to the [Key Findings](#key-findings).

## Key Findings

1. The
   [DualPI2](https://datatracker.ietf.org/doc/draft-ietf-tsvwg-aqm-dualq-coupled/)
   qdisc introduces a [network bias](#network-bias) for L4S flows over
   existing flows.
2. TCP Prague and DualPI2 exhibit a greater level of
   [RTT unfairness](#rtt-unfairness) than the commonly used CUBIC and pfifo.
3. L4S transports can experience broad
   [intra-flow latency spikes](#intra-flow-latency-spikes) at RFC 3168
   bottlenecks, particularly with the widely deployed fq_codel.
4. The marking scheme in the DualPI2 qdisc is
   [burst intolerant](#burst-intolerance), causing under-utilization for
   traffic with bursty arrivals.

## Elaboration on Key Findings

### Network Bias

$(chart_inline "L4S Network Bias 20ms" "s1-charts" "rttfair_cc_qdisc_20ms_20ms.svg")
*Figure 1*

Measurements show that DualPI2 consistently gives TCP Prague flows a throughput advantage over conventional CUBIC flows, 
where both flows run over the same path RTT.  In *Figure 1* above, we compare the typical status quo in the form of a 
250ms-sized dumb FIFO (middle) to DualPI2 (left) and an Approximate Fairness AQM (right) which actively considers queue 
occupancy of each flow.  The baseline path RTT for both flows is 20ms, which is in the range expected for CDN to consumer 
traffic.  Both flows start simultaneously and run for 3 minutes, with the throughput figures being taken from the final 
minute of the run as an approximation of reaching steady-state.

It is well-known that CUBIC outperforms NewReno on high-BDP paths where the polynomial curve grows faster than the linear 
one; the 250ms queue depth of the dumb FIFO and the relatively high throughput of the link puts the middle chart firmly in 
that regime.  Because no AQM is present at the bottleneck, TCP Prague behaves approximately like NewReno and, as expected, 
is outperformed by CUBIC.  It is difficult, incidentally, to see where L4S' "scalable throughput" claim is justified here, 
as CUBIC clearly scales up in throughput better in today's typical Internet environment.

L4S assumes that an L4S-aware AQM is present at the bottleneck.  The left-hand chart shows what happens when DualPI2, which 
is claimed to implement L4S in the network, is indeed present there.  In a stark reversal from the dumb FIFO scenario, TCP 
Prague is seen to have a large throughput advantage over CUBIC, in more than a 2:1 ratio.  This cannot be explained by 
CUBIC's sawtooth behaviour, as that would leave much less than 50% of available capacity unused.  We believe that several 
effects, both explicit and accidental, in DualPI's design are giving TCP Prague an unfair advantage in throughput.

The CodelAF results are presented as an example of what can easily be achieved by actively equalising queue occupancy across 
flows through differential AQM activity, which compensates for differing congestion control algorithms and path 
characteristics.  CodelAF was initially developed as part of SCE, but the version used here is purely RFC-3168 compliant.
On the right side of *Figure 2*, you can see that CUBIC and TCP Prague are given very nearly equal access to the link, with 
considerably less queuing than in the dumb FIFO.

$(chart_inline "L4S Network Bias 80ms" "s1-charts" "rttfair_cc_qdisc_80ms_80ms.svg")
*Figure 2*

These results also hold on 10ms and 80ms paths, with only minor variations; most notably, at 80ms CUBIC loses a bit of 
throughput in CodelAF due to its sawtooth behaviour, but is still not disadvantaged to the extent that DualPI2 imposes.  We 
also see very similar results to CodelAF when the current state-of-the-art fq_codel and CAKE qdiscs are used.  Hence we show 
that DualPI2 represents a regression in behaviour from both the currently typical deployment and the state of the art, with 
respect to throughput fairness on a common RTT.  We could even hypothesise from this data that a deliberate attempt to 
introduce a "fast lane" is in evidence here.

### RTT Unfairness

One of the so-called "Prague Requirements" adopted by L4S is to reduce the dependence on path RTT for flow throughput.
Conventional single-queue AQM tends to result in a consistent average cwnd across flows sharing the bottleneck, and since 
BDP == cwnd * MTU == throughput * RTT, the throughput of each flow is inversely proportional to the effective RTT 
experienced by that flow, which in turn is the baseline path RTT plus the queue delay.

However, DualPI2 is designed to perpetuate this equalising of average cwnd, not only between flows in the same queue, but 
between the two classes of traffic it manages (L4S and conventional).  Further, the effective RTT differs between the two 
classes of traffic due to the different AQM target in each, and the queue depth in the L4S class is limited to a very small 
value.  The result is that the ratio of effective RTTs is not diluted by queue depth, as it would be in a deeper queue, and 
also not compensated for by differential per-flow AQM action, as it would be in FQ or AF AQMs which are already deployed to
some extent.

$(chart_inline "L4S RTT Bias 10/160ms" "s1-charts" "rttfair_cc_qdisc_10ms_160ms.svg")
*Figure 3*

This can be clearly seen in *Figure 3* above, in which a comparatively extreme ratio of path RTTs has been introduced between 
two flows to illustrate the effect.  In the middle, the 250ms dumb FIFO is clearly seen to dilute the effect (the effective 
RTTs are 260ms and 410ms respectively) to the point where, except for two CUBIC flows competing against each other, other 
effects dominate the result in terms of steady-state throughput.  On the right, the AF AQM clearly reduces the RTT bias 
effect to almost parity, with the exception of the pair of CUBIC flows which are still slightly improved over the dumb FIFO.  

But on the left, when the bottleneck is managed by DualPI2, the shorter-RTT flow has a big throughput advantage in every 
case - even overcoming the throughput advantage that DualPI2 normally gives to TCP Prague, as shown previously.  Indeed the 
only case where DualPI2 shows better elimination of RTT bias than the dumb FIFO is entirely due to this bias in favour of 
TCP Prague.  Additionally, in the pure-L4S scenario in which both flows are TCP Prague, the ratio of throughput actually 
exceeds the nominal 16:1 ratio of path RTTs.

We conclude that DualPI2 does not represent "running code" supporting the L4S specification in respect of the "reduce RTT 
dependence" element of the Prague Requirements.  Observing that the IETF standardisation process is predicated upon "rough 
consensus and running code", we strongly suggest that this deficiency be remedied before a WGLC process is considered.

### Intra-flow Latency Spikes

Intra-flow latency refers to the delay experienced within a single flow, and for
TCP is typically measured using TCP RTT. Increases in intra-flow latency lead to
delays experienced by the user, for example when HTTP/2 requests are
multiplexed over a single TCP or QUIC flow that is building a queue.

Due to the redefinition of the CE codepoint
[[l4s-id](https://datatracker.ietf.org/doc/draft-ietf-tsvwg-ecn-l4s-id/)], L4S
transports underreact to CE signals sent by existing
[RFC3168](https://tools.ietf.org/html/rfc3168) AQMs, causing them to inflate
queues where these AQMs are deployed. We usually discuss this in the context of
safety for non-L4S flows in the same RFC3168 queue, but the added delay that L4S
flows can induce on themselves is also an important consideration.

For a practical example, we'll look at the transient behavior of fq_codel. Rate
reductions in particular can lead to intra-flow latency spikes. They occur
routinely in fq_codel, both due to flow arrivals at the bottleneck, and rate
changes in wireless links, which occur on timescales of tens to hundreds of
milliseconds. (For more information on the approximate scope of fq_codel
deployments, see [Deployments of fq_codel](#deployments-of-fq_codel)).

First, let's look at what happens when a standard **CUBIC** flow experiences a
routine 50% rate reduction in an fq_codel queue, from 50Mbps to 25Mbps (see
*Figure 10*).

$(plot_inline "Rate Reduction for CUBIC with fq_codel, 50 -> 25Mbit at 80ms" "l4s-s2-codel-rate-step" "ns-clean-cubic-fq_codel-50Mbit-25mbit-80ms_tcp_delivery_with_rtt.svg")  
*Figure 10*

In *Figure 10* above, we can see a brief spike in intra-flow latency (TCP RTT)
at around T=30, as Codel's estimator detects the queue, and the flow is
signaled to slow down. CUBIC reacts with the expected 50% multiplicative
decrease.

Next, let's look at the result when an L4S **TCP Prague** flow experiences the
same 50% rate reduction (see *Figure 11* below):

$(plot_inline "Rate Reduction for Prague with fq_codel, 50 -> 25Mbit at 80ms" "l4s-s2-codel-rate-step" "ns-clean-prague-fq_codel-50Mbit-25mbit-80ms_tcp_delivery_with_rtt.svg")  
*Figure 11*

Comparing *Figure 10* and *Figure 11*, we can see that the induced latency spike
has a much longer duration for TCP Prague than CUBIC. Note that although the
spike may appear small in magnitude due to the plot scale, 100ms is a
significant induced delay when targets in the L4S queue are around 1ms, and
further, we can see that the spike lasts around 5 seconds. This occurs because
TCP Prague mis-interprets the CE signal as coming from an L4S instead of an
RFC3168 queue. Prague reacts with a small linear cwnd reduction instead of the
expected multiplicative decrease, building excessive queue until Codel's
signaling eventually gets it under control.

The consequences of L4S transports underreacting to RFC3168 CE signals can be
more severe as the rate reductions get larger. See *Figure 12* and *Figure 13*
below for what happens to TCP Prague flows when reduced from 50Mbps to 5Mbps and
1Mbps, respectively. These larger reductions may be encountered, for example, as
wireless devices with fq_codel in the driver change rates in areas of
intermittent AP coverage.

$(plot_inline "Rate Reduction for Prague with fq_codel, 50 -> 5Mbit at 80ms" "l4s-s2-codel-rate-step" "ns-clean-prague-fq_codel-50Mbit-5mbit-80ms_tcp_delivery_with_rtt.svg")  
*Figure 12*

$(plot_inline "Rate Reduction for Prague with fq_codel, 50 -> 1Mbit at 80ms" "l4s-s2-codel-rate-step" "ns-clean-prague-fq_codel-50Mbit-1mbit-80ms_tcp_delivery_with_rtt.svg")  
*Figure 13*

In *Figure 13* above, we see a latency spike that has exceeded the fixed scale
of our plot. However, a review of the
$(batch_link ".flent.gz file" "l4s-s2-codel-rate-step" "ns-clean-prague-fq_codel-50Mbit-1mbit-80ms.flent.gz")
shows the maximum TCP RTT to be **4346ms**, and we can see that the spike lasts
for over **30 seconds**. This behavior is something we need to be aware of
before introducing an ambiguous definition of the CE signal in the presence of
[fq_codel deployments](#deployments-of-fq_codel).

See the [Scenario 3](#scenario-3-codel-variable-rate) results, in
particular for TCP Prague through fq_codel, to look at what happens when
rates vary several times over the course of a flow.

### Burst Intolerance

The default marking scheme used in the DualPI2 L queue begins at a shallow, sub
1 ms threshold, which while effective for keeping queues shorter, causes
excessive marking for bursty packet arrivals. This results in link
under-utilization for the typically bursty Internet traffic. Burstiness can
come from the link layer, for example with WiFi, where bursts of up to about
4ms are sent, or just from cross-flow traffic through shared bottlenecks.

Note that burstiness is distinguished from jitter in general, which is
associated with a variance in inter-packet gaps, but does not necessarily
consist of well-defined bursts of packets at line rate. In any case, both
well-paced and bursty flows can be expected on the Internet.

[Scenario 2](#scenario-2-codel-rate-step) and
[Scenario 3](#scenario-3-codel-variable-rate) both include runs with netem
simulated bursts of approximately 4ms in duration. In *Figure 14*, we can
see how **CUBIC through fq_codel** handles such bursts.

$(plot_inline "Rate Reduction for CUBIC with fq_codel, 50Mbps -> 25Mbps with Bursty Traffic" "l4s-s2-codel-rate-step" "ns-bursty-cubic-fq_codel-50Mbit-25mbit-20ms_tcp_delivery_with_rtt.svg")  
*Figure 14*

Next, in *Figure 15* we see how **TCP Prague through DualPI2** handles the same
bursts:

$(plot_inline "Rate Reduction for Prague with dualpi2, 50Mbps -> 25Mbps with Bursty Traffic" "l4s-s2-codel-rate-step" "ns-bursty-prague-dualpi2-50Mbit-25mbit-20ms_tcp_delivery_with_rtt.svg")  
*Figure 15*

In *Figure 14* and *Figure 15* we can see that the lower TCP RTT of TCP Prague
comes with a tradeoff of about a 50% reduction in link utilization. While this
may be appropriate for low-latency traffic, capacity seeking bulk downloads may
prefer increased utilization at the expense of some intra-flow delay. We raise
this point merely to help set the expectation that maintaining strictly low
delays at bottlenecks comes at the expense of some link utilization for typical
Internet traffic.

### Unsafety in Tunnels Through RFC3168 Bottlenecks

When tunneled traffic traverses an RFC 3168 bottleneck, including those with FQ
(such as fq_codel), it can lose the flow isolation that L4S depends on for flow
safety. When this happens, L4S flows dominate the non-L4S flows in the tunnel,
whether the non-L4S flows are ECN capable or not.

This is expected to happen to any tunneled traffic whose encapsulated packets
use a fixed 5-tuple (most of them), at any RFC3168 bottleneck, with or without
FQ. Here is a common sample topology:

\`\`\`
    -------------------    ------------    -------------------
    | Tunnel Endpoint |----| fq_codel |----| Tunnel Endpoint |
    -------------------    ------------    -------------------
\`\`\`

In *Figure 16* below, we can see how an L4S **Prague** flow (the red trace)
dominates a standard **CUBIC** flow (the green trace) in the same
[Wireguard](https://www.wireguard.com/) tunnel:

$(plot_inline "wireguard Tunnel, Prague vs CUBIC" "l4s-s5-tunnel" "phys-wireguard-prague-vs-cubic-fq_codel-50Mbit-20ms_tcp_delivery_with_rtt.svg")  
*Figure 16*


The following table shows the 60-second median throughputs of the tested flows
(reported by netperf, and in the .flent.gz files):

| Tunnel                                    | CC algo 1 | CC algo 2 | Throughput 1 | Throughput 2 | Ratio |
| ----------------------------------------- | --------- | --------- | ------------ | ------------ | ----- |
| [Wireguard](https://www.wireguard.com/)   | Prague    | CUBIC     | 43.75 Mbps   | 2.41 Mbps    | 18:1  |
| [Wireguard](https://www.wireguard.com/)   | Prague    | Reno      | 43.27 Mbps   | 3.91 Mbps    | 11:1  |
| [ipfou](https://lwn.net/Articles/614348/) | Prague    | CUBIC     | 44.81 Mbps   | 2.54 Mbps    | 18:1  |
| [ipfou](https://lwn.net/Articles/614348/) | Prague    | Reno      | 44.27 Mbps   | 3.06 Mbps    | 14:1  |

See [Scenario 5](#scenario-5-tunnels) in the Appendix for links to these
results, which are expected to be similar with most any tunnel.

*Note #1* In testing this scenario, it was discovered that the [Foo over
UDP](https://lwn.net/Articles/614348/) tunnel has the ability to use an
automatic source port (\`encap-sport auto\`), which restores flow isolation by
using a different source port for each inner flow. However, this is tunnel
dependent, and secure tunnels like VPNs are not likely to support this option,
as doing so would be a security risk.

*Note #2* Also in testing, we found that when using a netns (network namespaces)
environment, the Linux kernel (5.4 at least) tracks a tunnel's inner flows even
as their encapsulated packets cross namespace boundaries, making the results not
representative of what typically happens in the real world. Flows not only get
their own hash, but that hash can actually change across the lifetime of the
flow, resulting in an unexpected AQM response. To avoid this problem, make sure
the client, middlebox and server all run on different kernels when testing
tunnels.

## Full Results

In the following results, the links are named as follows:

- _plot_: the plot svg
- _cli.pcap_: the client pcap
- _srv.pcap_: the server pcap
- _teardown_: the teardown log, showing qdisc config and stats

### Scenario 1: RTT Fairness

$(cli_gen_table s1)

### Scenario 2: Codel Rate Step

$(cli_gen_table s2)

### Scenario 3: Codel Variable Rate

$(cli_gen_table s3)

### Scenario 4: Bi-directional Traffic, Asymmetric Rates

### Scenario 5: Tunnels

$(cli_gen_table s5)

## Appendix

### Scenario 1 Fairness Table

**D<sub>SS</sub>** Delivery rate (throughput) at steady state (mean of 60 second window ending 2 seconds before end of test)

$(<s1_table.md)

### Background

Conventional congestion control is based on the
[AIMD](https://en.wikipedia.org/wiki/Additive_increase/multiplicative_decrease)
(Additive Increase, Multiplicate Decrease) principle.  This exhibits a
characteristic sawtooth pattern in which the congestion window grows slowly,
then reduces rapidly on receipt of a congestion signal.  This was introduced to
solve the problem of congestion collapse.  However, it is incapable of finding
and settling on the ideal congestion window, which is approximately equal to the
bandwidth-delay product (BDP) plus a jitter margin.

High Fidelity Congestion Control is an attempt to solve this problem by
implementing a finer-grained control loop between the network and the transport
layer.  Hence, instead of oscillating around the ideal (at best), the transport
can keep the ideal amount of traffic in the network, simultaneously maximising
throughput and minimising latency.

### Deployments of fq_codel

The [fq_codel](https://tools.ietf.org/html/rfc8290) qdisc has been in the Linux
kernel since version 3.6 (late 2012) and is now in widespread use in commercial
routers (e.g. Ubiquiti EdgeMAX and UniFi products), CPE devices and some ISP
backhauls (see [Preseem](https://preseem.com/qoe-optimized-shaping/)). It has
also been integrated into the ath9k, ath10k, mt76 and iwl WiFi drivers, and is
used in Google WiFi and OpenWrt, as well as vendor products that depend on
OpenWrt, such as Open Mesh products. Since fq_codel uses RFC3168 ECN signaling
by default, it is important for safety and performance that new congestion
control mechanisms take existing RFC3168 bottlenecks into account.

### Test Setup

The test setup consists of a dumbbell configuration (client, middlebox and
server) using network namespaces. [Flent](https://flent.org/) was used for all
tests. The L4S kernel tested was
[L4STeam/linux@b256daedc7672b2188f19e8b6f71ef5db7afc720](https://github.com/L4STeam/linux/tree/b256daedc7672b2188f19e8b6f71ef5db7afc720)
(from Aug 4, 2020).

The single **fl** script performs the following functions:
- updates itself onto the management server and clients
- runs tests (./fl run), plot results (./fl plot) and pushes them to a server
- acts as a harness for flent, setting up and tearing down the test config
- generates this README.md from a template

If there are more questions, feel free to file an
[issue](https://github.com/heistp/l4s-tests/issues).
