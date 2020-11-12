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

$(chart_inline "L4S Network Bias 20ms" "s1-charts" "rttfair_cc_qdisc_20ms_20ms.svg")
*Figure 1*

Measurements show that DualPI2 consistently gives TCP Prague flows a throughput advantage over conventional CUBIC flows, 
where both flows run over the same path RTT.  In the above plot, we compare the typical status quo in the form of a 
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
On the right side of the chart, you can see that CUBIC and TCP Prague are given very nearly equal access to the link, with 
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

One of the socalled "Prague Requirements" adopted by L4S is to reduce the dependence on path RTT for flow throughput.  
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

This can be clearly seen in the above chart, in which a comparatively extreme ratio of path RTTs has been introduced between 
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
