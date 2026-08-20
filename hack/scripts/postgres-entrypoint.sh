#!/bin/bash
set -euo pipefail

# Postgres rejects SSL keys with group/world access. Testdata keys are often 0644
# (and bind mounts on Docker Desktop may look world-readable), so copy into a
# private path before starting.
TLS_SRC=${TLS_SRC:-/tls}
TLS_DST=${TLS_DST:-/var/lib/postgresql/tls}

mkdir -p "$TLS_DST"
cp "$TLS_SRC/server.crt" "$TLS_SRC/server.key" "$TLS_SRC/ca.crt" "$TLS_DST/"
chmod 600 "$TLS_DST/server.key"
chmod 644 "$TLS_DST/server.crt" "$TLS_DST/ca.crt"
chown -R postgres:postgres "$TLS_DST"

exec docker-entrypoint.sh postgres \
	-c ssl=on \
	-c ssl_cert_file="$TLS_DST/server.crt" \
	-c ssl_key_file="$TLS_DST/server.key" \
	-c ssl_ca_file="$TLS_DST/ca.crt" \
	"$@"
