#!/bin/bash

out="${1:-irtt.json.gz}"

irtt client -i 10ms -d 1m -q apu2c -o "$out"
