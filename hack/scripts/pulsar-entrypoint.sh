#!/bin/sh
set -e

# Pulsar Manager host, default for docker-compose network.
# You can override when running locally:
#   PM_HOST=localhost PM_PORT=17750 sh pulsar_init.sh
PM_HOST=${PM_HOST:-pulsar-manager}
PM_PORT=${PM_PORT:-7750}

BASE_URL="http://${PM_HOST}:${PM_PORT}/pulsar-manager"

printf "==> Waiting for pulsar-manager ...\n"
until curl -s "${BASE_URL}/csrf-token" > /dev/null; do
  sleep 5
done
printf "\n"

printf "==> Extracting CSRF Token...\n"
TOKEN=$(curl -sf "${BASE_URL}/csrf-token")
printf "CSRF_TOKEN = %s\n\n" "${TOKEN}"

if [ -z "$TOKEN" ]; then
  echo "Failed to get CSRF token"
  exit 1
fi

printf "==> Create administrator...\n"
curl -sf -X PUT "${BASE_URL}/users/superuser" \
  -H "X-XSRF-TOKEN: $TOKEN" \
  -H "Cookie: XSRF-TOKEN=$TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"admin123","description":"dev","email":"any@any.com"}'
printf "\n* username: admin"
printf "\n* password: admin123\n\n"

#printf "==> Create environments ...\n"
printf "Pulsar manager initialized\n"

#创建环境
#echo '{"name":"confx","broker":"http://pulsar:8080","bookie":"pulsar://pulsar:6650"}' | http put http://localhost:19527/pulsar-manager/environments/environment \
#token:$TOKEN \
#X-XSRF-TOKEN:$CSRF \
#username:admin \
#Cookie:'Admin-Token=$TOKEN; JSESSIONID=$JSESSIONID; XSRF-TOKEN=e385dd76-a41b-4849-88dd-3e8b5ccf5fdc; e385dd76-a41b-4849-88dd-3e8b5ccf5fdc=e385dd76-a41b-4849-88dd-3e8b5ccf5fdc'

#登陆
#echo '{"username":"admin","password":"admin123"}' | http post http://localhost:19527/pulsar-manager/login \
#Cookie:'XSRF-TOKEN=318f3737-9107-43a9-a724-47d04def1d44; JSESSIONID=AC5237080D4C8C30457A82BA33DE1343' \
#X-XSRF-TOKEN:318f3737-9107-43a9-a724-47d04def1d44