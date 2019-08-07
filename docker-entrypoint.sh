#!/bin/sh

set -e

if [ -z "$ADMIN_USER" -o -z "$ADMIN_PASSWORD" ]; then
	ADMIN_USER=admin
	ADMIN_PASSWORD=$(dd if="/dev/urandom" bs=8 count=1 status=none | xxd -p)
	ADMIN_PASSWORD=$(</dev/urandom tr -dc '0123456789' | head -c10; echo "")

	cat <<EOF
You didn't specify the ADMIN_USER and ADMIN_PASSWORD environment varibles.
I've generated some for you:

Admin username: $ADMIN_USER
Admin password: $ADMIN_PASSWORD

EOF
fi

if [ -z "$LISTEN_ADDRESS" ]; then
	LISTEN_ADDRESS="0.0.0.0:1412"
fi

EXTRA_ARGS=

if [ -n "$LOAD_CSV" ]; then
	echo "The following CSV file will be loaded at startup: $LOAD_CSV"
	echo ""

	EXTRA_ARGS="-loadCSV $LOAD_CSV"
fi

exec /home/server/server -listen "$LISTEN_ADDRESS" -adminUser "$ADMIN_USER" -adminPassword "$ADMIN_PASSWORD" $EXTRA_ARGS
