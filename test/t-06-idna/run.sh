#!/bin/bash

set -e
. "$(dirname "$0")/../util/lib.sh"

init
check_hostaliases

rm -rf .data-A .data-B .mail

# Build with the DNS override, so we can fake DNS records.
export GOTAGS="dnsoverride"

# Launch minidns in the background using our configuration.
minidns_bg --addr=":9053" -zones=zones >> .minidns.log 2>&1

# Two servers:
# A - listens on :1025, hosts srv-ñ
# B - listens on :2015, hosts srv-ü

CONFDIR=A generate_certs_for srv-ñ
CONFDIR=A add_user ñangapirí@srv-ñ antaño
CONFDIR=A add_user nadaa@nadaA nadaA

CONFDIR=B generate_certs_for srv-ü
CONFDIR=B add_user pingüino@srv-ü velóz
CONFDIR=B add_user nadab@nadaB nadaB

mkdir -p .logs-A .logs-B

chasquid -v=2 --logfile=.logs-A/chasquid.log --config_dir=A \
	--testing__dns_addr=127.0.0.1:9053 \
	--testing__outgoing_smtp_port=2025 &
chasquid -v=2 --logfile=.logs-B/chasquid.log --config_dir=B \
	--testing__dns_addr=127.0.0.1:9053 \
	--testing__outgoing_smtp_port=1025 &

wait_until_ready 1465
wait_until_ready 2465
wait_until_ready 9053

# Send from A to B.
smtpc --addr=localhost:1465 --user=nadaA@nadaA --password=nadaA \
	--server_cert=A/certs/srv-ñ/fullchain.pem \
	pingüino@srv-ü < from_A_to_B

wait_for_file .mail/pingüino@srv-ü
mail_diff from_A_to_B .mail/pingüino@srv-ü

# Send from B to A.
smtpc --addr=localhost:2465 --user=nadaB@nadaB --password=nadaB \
	--server_cert=B/certs/srv-ü/fullchain.pem \
	ñangapirí@srv-ñ < from_B_to_A

wait_for_file .mail/ñangapirí@srv-ñ
mail_diff from_B_to_A .mail/ñangapirí@srv-ñ

success
