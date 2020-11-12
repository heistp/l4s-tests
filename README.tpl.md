# SCE-L4S ECT(1) Test Results

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
4. [Full Results](#full-results)
   1. [Scenario 1: RTT Fairness](#scenario-1-rtt-fairness)
   2. [Scenario 2: Codel Rate Step](#scenario-2-codel-rate-step)
   3. [Scenario 3: Codel Variable Rate](#scenario-3-codel-variable-rate)
5. [Appendix](#appendix)
   1. [Scenario 1 Fairness Table](#scenario-1-fairness-table)
   2. [Background](#background)
   3. [Deployments of fq_codel](#deployments-of-fq_codel)
   4. [Test Setup](#test-setup)

## Introduction

The Transport Area Working Group
([TSVWG](https://datatracker.ietf.org/group/tsvwg/about/)) will undergo a WGLC
(working group last call) for [L4S](https://riteproject.eu/dctth/). L4S proposes
to use the available ECT(1) codepoint for two purposes:

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
   bottlenecks, particularly in the widely deployed fq_codel.
4. The marking scheme in the DualPI2 qdisc is
   [burst intolerant](#burst-intolerance), causing under-utilization for
   traffic with bursty arrivals.

## Elaboration on Key Findings

### Network Bias

$(plot_inline "CUBIC(20ms) vs Prague(20ms) through DualPI2" "l4s-s1-rttfair" "ns-cubic-vs-prague-dualpi2-10Mbit-20ms-20ms_tcp_delivery_with_rtt.svg")  
*Figure 1*  

TODO

### RTT Unfairness

TODO

$(chart_inline "L4S Network Bias" "s1-charts" "l4s_network_bias.svg")
*Figure 2*

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
routers (e.g. Ubiquiti), CPE devices and some ISP backhauls (e.g.
[Preseem](https://preseem.com/qoe-optimized-shaping/)). It has also been
integrated into the ath9k, ath10k, mt76 and iwl WiFi drivers, and is used in
Google WiFi and OpenWrt, as well as vendor products that depend on OpenWrt, such
as Open Mesh products. Since fq_codel uses RFC3168 ECN signaling by default, it
is important for safety and performance that new congestion control mechanisms
take RFC3168 ECN into account.

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
