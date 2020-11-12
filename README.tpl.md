# SCE-L4S ECT(1) Test Results

Evidence in opposition to the L4S WGLC

Pete Heist  
Jonathan Morton  

## Table of Contents

1. [Introduction](#introduction)
2. [Key Findings](#key-findings)
3. [Elaboration on Key Findings](#elaboration-on-key-findings)
4. [Full Results](#full-results)
   1. [Scenario 1: RTT Fairness](#scenario-1-rtt-fairness)
   2. [Scenario 2: Codel Rate Step](#scenario-2-codel-rate-step)
   3. [Scenario 3: Codel Variable Rate](#scenario-3-codel-variable-rate)
5. [Appendix](#appendix)
   1. [Scenario 1 Fairness Table](#scenario-1-fairness-table)
   2. [Background](#background)
   3. [Test Setup](#test-setup)

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
  giving some level of priority in dualpi2 queues for L4S traffic over
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
   [dualpi2](https://datatracker.ietf.org/doc/draft-ietf-tsvwg-aqm-dualq-coupled/)
   qdisc introduces a [network bias](#network-bias) for L4S flows over
   existing flows.
2. TCP Prague and dualpi2 exhibit a greater level of
   [RTT unfairness](#rtt-unfairness) than the commonly used CUBIC and pfifo.
3. Due to the incompatible redefinition of CE defined in
   [l4s-id](https://datatracker.ietf.org/doc/draft-ietf-tsvwg-ecn-l4s-id/),
   L4S transports can experience broad
   [intra-flow latency-spikes](#intra-flow-latency-spikes) at RFC 3168
   bottlenecks, particularly upon rate reductions in the widely deployed
   fq_codel.
4. The marking scheme in the dualpi2 qdisc is
   [burst intolerant](#burst-intolerance), causing under-utilization for
   traffic with bursty arrivals.

## Elaboration on Key Findings

### Network Bias

$(plot_inline "CUBIC(20ms) vs Prague(20ms) through dualpi2" "l4s-s1-rttfair" "ns-cubic-vs-prague-dualpi2-10Mbit-20ms-20ms_tcp_delivery_with_rtt.svg")  
*Figure 1*  

TODO

### RTT Unfairness

TODO

$(chart_inline "L4S Network Bias" "s1-charts" "l4s_network_bias.svg")
*Figure 2*

### Intra-flow Latency Spikes

TODO

### Burst Intolerance

## Full Results

In the following results, the links are named as follows:

- _plot_: the plot svg
- _cli.pcap_: the client pcap
- _srv.pcap_: the server pcap
- _teardown_: the teardown log, showing qdisc config and stats

### Scenario 1: RTT Fairness

$(cli_gen_table s1)

### Scenario 2: Rate Steps Down with fq_codel

$(cli_gen_table s2)

### Scenario 3: Variable Rates with fq_codel

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
