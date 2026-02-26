#!/bin/sh
set -e

# Pulsar Manager host, default for docker-compose network.
# You can override when running locally:
#   PM_HOST=localhost PM_PORT=17750 sh pulsar_init.sh
PM_HOST=${PM_HOST:-pulsar-manager}
PM_PORT=${PM_PORT:-7750}
USERNAME=${USERNAME:-admin}
PASSWORD=${PASSWORD:-pulsar}
BASE_URL="http://${PM_HOST}:${PM_PORT}/pulsar-manager"

printf "==> envs:\n"
printf "\t%s\n"   "$BASE_URL"
printf "\t%s\n"   "$USERNAME"
printf "\t%s\n\n" "$PASSWORD"

printf "==> Waiting for pulsar-manager ...\n"
until curl -s "${BASE_URL}/csrf-token" > /dev/null; do
  sleep 5
done
printf "\n"

printf "==> Extracting CSRF Token...\n"
CSRF_TOKEN=$(curl -sf "${BASE_URL}/csrf-token")
printf "CSRF_TOKEN = %s\n\n" "${CSRF_TOKEN}"

printf "==> Create administrator...\n"
PAYLOAD=$(printf '{"name":"%s","password":"%s","description":"dev","email":"any@any.com"}' "$USERNAME" "$PASSWORD")
curl -sf -X PUT "${BASE_URL}/users/superuser" \
  -H "X-XSRF-TOKEN: $CSRF_TOKEN" \
  -H "Cookie: XSRF-TOKEN=$CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD"
printf "\n* username: %s" "$USERNAME"
printf "\n* password: %s\n\n" "$PASSWORD"

printf "==> Logging in...\n\n"
PAYLOAD=$(printf '{"username":"%s","password":"%s"}' "$USERNAME" "$PASSWORD")
RESP=$(curl -s -i -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json;charset=UTF-8" \
  -H "Cookie: XSRF-TOKEN=$CSRF_TOKEN" \
  -H "X-XSRF-TOKEN: $CSRF_TOKEN" \
  -d "$PAYLOAD")
JSESSIONID=$(echo "$RESP" | awk -F'[=;]' '/Set-Cookie: JSESSIONID/ {print $2}')
TOKEN=$(echo "$RESP" | awk '/^token:/ {print $2}' | tr -d '\r')

printf "==> Create environment\n"
curl -X PUT "$BASE_URL/environments/environment" \
  -H "Content-Type: application/json" \
  -H "token: $TOKEN" \
  -H "X-XSRF-TOKEN: $CSRF_TOKEN" \
  -H "username: $USERNAME" \
  -H "Cookie: Admin-Token=$TOKEN; JSESSIONID=$JSESSIONID; XSRF-TOKEN=$CSRF_TOKEN; $CSRF_TOKEN=$CSRF_TOKEN" \
  -d '{"name":"confx","broker":"http://pulsar:8080","bookie":"pulsar://pulsar:6650"}'
