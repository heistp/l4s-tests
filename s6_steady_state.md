| Rate | qdisc | CC1 (RTT) | D<sub>SS</sub>1 | CC2 (RTT) | D<sub>SS</sub>2 | Ratio |
| ---- | ----- | --------- | --------------- | --------- | ----------------| -- |
| 50Mbit | fq_codel(1q) | prague(20ms) | 45.70 | cubic-ecn(20ms) | 1.94 | 24:1 |
| 50Mbit | fq_codel(1q,1ms/20ms) | prague(20ms) | 45.94 | cubic-ecn(20ms) | 1.68 | 27:1 |
| 50Mbit | pie | prague(20ms) | 45.69 | cubic-ecn(20ms) | 1.98 | 23:1 |
| 50Mbit | pie(100p/5ms) | prague(20ms) | 45.41 | cubic-ecn(20ms) | 2.23 | 20:1 |
| 50Mbit | red(150000b) | prague(20ms) | 45.93 | cubic-ecn(20ms) | 1.69 | 27:1 |
| 50Mbit | red(400000b) | prague(20ms) | 45.62 | cubic-ecn(20ms) | 2.05 | 22:1 |
| 50Mbit | fq_codel(1q) | prague(20ms) | 45.81 | cubic-noecn(20ms) | 1.81 | 25:1 |
| 50Mbit | fq_codel(1q,1ms/20ms) | prague(20ms) | 45.76 | cubic-noecn(20ms) | 1.83 | 25:1 |
| 50Mbit | pie | prague(20ms) | 45.57 | cubic-noecn(20ms) | 2.01 | 23:1 |
| 50Mbit | pie(100p/5ms) | prague(20ms) | 45.63 | cubic-noecn(20ms) | 2.00 | 23:1 |
| 50Mbit | red(150000b) | prague(20ms) | 46.02 | cubic-noecn(20ms) | 1.57 | 29:1 |
| 50Mbit | red(400000b) | prague(20ms) | 45.80 | cubic-noecn(20ms) | 1.81 | 25:1 |
| 50Mbit | fq_codel(1q) | prague(20ms) | 46.60 | reno-ecn(20ms) | 1.10 | 42:1 |
| 50Mbit | fq_codel(1q,1ms/20ms) | prague(20ms) | 46.34 | reno-ecn(20ms) | 1.25 | 37:1 |
| 50Mbit | pie | prague(20ms) | 46.94 | reno-ecn(20ms) | 0.78 | 60:1 |
| 50Mbit | pie(100p/5ms) | prague(20ms) | 45.86 | reno-ecn(20ms) | 1.77 | 26:1 |
| 50Mbit | red(150000b) | prague(20ms) | 45.86 | reno-ecn(20ms) | 1.73 | 27:1 |
| 50Mbit | red(400000b) | prague(20ms) | 45.90 | reno-ecn(20ms) | 1.80 | 26:1 |
| 50Mbit | fq_codel(1q) | prague(20ms) | 46.49 | reno-noecn(20ms) | 1.17 | 40:1 |
| 50Mbit | fq_codel(1q,1ms/20ms) | prague(20ms) | 46.01 | reno-noecn(20ms) | 1.59 | 29:1 |
| 50Mbit | pie | prague(20ms) | 46.67 | reno-noecn(20ms) | 0.96 | 48:1 |
| 50Mbit | pie(100p/5ms) | prague(20ms) | 46.36 | reno-noecn(20ms) | 1.32 | 35:1 |
| 50Mbit | red(150000b) | prague(20ms) | 46.27 | reno-noecn(20ms) | 1.34 | 35:1 |
| 50Mbit | red(400000b) | prague(20ms) | 45.74 | reno-noecn(20ms) | 1.87 | 24:1 |
