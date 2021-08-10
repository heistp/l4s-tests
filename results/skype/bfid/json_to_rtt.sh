#!/bin/bash

jq -r '.round_trips[] | (.delay.rtt/1000000)'
