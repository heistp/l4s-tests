#!/bin/bash

# script location
SCRIPT_DIR='~/src/l4s-tests'

# batch config
BATCH=l4s
BATCH_FILE=${BATCH}.batch

# output file spec
BATCH_OUT_SPEC="$BATCH-????-??-?????????"

# architectures
ARCHS=(l4s)

# management config
MGMT_SSH=

# SCE config
SCE_CLI_SSH=
SCE_CLI_RIGHT=
SCE_MID_SSH=
SCE_MID_LEFT=
SCE_MID_RIGHT=
SCE_SRV_SSH=
SCE_SRV_LEFT=

# L4S config
L4S_CLI_SSH=
L4S_CLI_RIGHT=
L4S_MID_SSH=
L4S_MID_LEFT=
L4S_MID_RIGHT=
L4S_SRV_SSH=
L4S_SRV_LEFT=

# netns config
NS_CLI_RIGHT=
NS_MID_LEFT=
NS_MID_RIGHT=
NS_SRV_LEFT=

# wireguard config
NS_SRV_WG_PRIV="WKYMm45p+0R0knXkp7e3rvxRSUwMg6pYGUa1R5ITVW0="
NS_SRV_WG_PUB="gA+0M6RBxpEqB9et57D7Im6mxyHQEiQGMISmEkPUIHw="
NS_CLI_WG_PRIV="KJ1e43Dz5xT82L3/ZVVA0iybHRvANCnIeRNKvHNVRFo="
NS_CLI_WG_PUB="5xW8bLo4+m/CrfDA96JjEei/lYVegaGZicmRSenX8H0="
NS_WG_IFACE=wg1
NS_WG_PORT=51821

# ipsec config
NS_IPSEC_KEY1="0x08ad0b0eb4e89a0f9be52676de065986b84e7d5dbc664b5d0431ae833613a4b9"
NS_IPSEC_KEY2="0xa0112cfa4777c377f76e891a6fe079aa573e4cc5392da55fc0f953f59b7dcbbe"
NS_IPSEC_ID="0x601559c5"
NS_IPSEC_REQID="$NS_IPSEC_ID"
NS_IPSEC_IFACE=ipsec1

# fou config
NS_FOU_IFACE=fou1
NS_FOU_PORT=5556

# all nodes to clear before each netns test
CLEAR_NODES=(cli mid srv)

# all ssh dests for physical hosts to clear before each phys test
CLEAR_SSH_DESTS=(c2 m1 m3 s2)

# push config
ARCHIVE_DIR=""
ARCHIVE_URL=""
PUSH_SSH_DEST=""

# tc config
TC_DIR="/usr/local/bin"

# Pushover config
PUSHOVER_SOUND_SUCCESS=""
PUSHOVER_SOUND_FAILURE=""
PUSHOVER_USER=""
PUSHOVER_TOKEN=""

# plot colors
# bright
#COLORS="\
#	'#1AC938',\
#	'#E8080A',\
#	'#8B2BE2',\
#	'#9F4800',\
#	'#F14CC1',\
#	'#A3A3A3',\
#	"
# dark
#COLORS="\
#	'#12711C',\
#	'#8C0800',\
#	'#591E71',\
#	'#592F0E',\
#	'#A23582',\
#	'#3C3C3C',\
#	"
# flent default
# COLORS="\
# '#1B9E77',\
# '#D95F02',\
# '#7570B3',\
# '#E7298A',\
# '#66A61E',\
# '#E6AB02',\
# '#A6761D',\
# '#666666',\
# "
# color set from 8-color set at colorbrewer2.org
# but changed FF7F00 to EB7500 for projector use
COLORS="\
'#E41A1C',\
'#377EB8',\
'#4DAF4A',\
'#984EA3',\
'#EB7500',\
'#FFFF33',\
'#A65628',\
'#F781BF',\
"

# plot size
PLOT_WIDTH=12
PLOT_HEIGHT=9
#PLOT_WIDTH=9
#PLOT_HEIGHT=7.5

# plot format
PLOT_FORMAT=svg

# compression config
COMPRESS=xz

# browser settings
BROWSER= # set to browser command, if not Linux or Mac

# results directories
RESULTS_URL="http://sce.dnsmgr.net/results"
RESULTS_DIR="l4s-2020-11-11T120000-final"

# harness config
DEBUG=0
TMPDIR="/tmp/l4s-tests"

# namespaces config
NS_OFFLOADS=off
NS_CLI_IP=10.9.9.1
NS_CLI_NET=$NS_CLI_IP/24
NS_SRV_IP=10.9.9.2
NS_SRV_NET=$NS_SRV_IP/24
NS_CLI_WG_IP=10.9.99.1
NS_CLI_WG_NET=$NS_CLI_WG_IP/24
NS_SRV_WG_IP=10.9.99.2
NS_SRV_WG_NET=$NS_SRV_WG_IP/24
NS_CLI_FOU_IP=10.9.98.1
NS_CLI_FOU_NET=$NS_CLI_FOU_IP/24
NS_SRV_FOU_IP=10.9.98.2
NS_SRV_FOU_NET=$NS_SRV_FOU_IP/24
NS_CLI_IPSEC_IP=10.9.97.1
NS_CLI_IPSEC_NET=$NS_CLI_IPSEC_IP/24
NS_SRV_IPSEC_IP=10.9.97.2
NS_SRV_IPSEC_NET=$NS_SRV_IPSEC_IP/24

# data_dir emits the data directory for a node
data_dir() {
	local node=$1
	# end of params

	echo "$TMPDIR/$node/data"
}

# log_dir emits the log directory for a node
log_dir() {
	local node=$1
	# end of params

	echo "$TMPDIR/$node/log"
}

# arch_tc emits the tc executable name for the architecture
arch_tc() {
	local arch=$1
	# end of params

	echo tc-${arch}
}

# node_ssh emits the ssh destination for a node
node_ssh() {
	local arch=$1
	local node=$2
	# end of params

	local v=${arch^^}_${node^^}_SSH
	if [[ ${!v} ]]; then
		echo ${!v}
	else
		echo "not_defined:$v"
	fi
}

# node_devs emits the node's interfaces for a direction
node_devs() {
	local arch=$1
	local net=$2
	local node=$3
	local dir=$4
	# end of params

	echov() {
		[[ ${!1} ]] && echo ${!1}
	}

	# select prefix from network
	local p
	case $net in
		phys)
			p=${arch^^}
			;;
		ns)
			p="NS"
			;;
		*)
			echo "unknown_net:$net"
			return 1
	esac

	# output based on direction
	local ok=false
	if [[ $dir == "left" ]] || [[ $dir == "bidir" ]]; then
			echov ${p}_${node^^}_LEFT
			ok=true
	fi
	if [[ $dir == "right" ]] || [[ $dir == "bidir" ]]; then
			echov ${p}_${node^^}_RIGHT
			ok=true
	fi

	# check dir value
	if [[ $ok == false ]]; then
		echo "unknown_dir:$dir"
		return 1
	fi
}
