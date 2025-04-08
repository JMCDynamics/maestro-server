#!/bin/bash

export SERVERURL=$HOST_ADDRESS
export WIREGUARD_ENDPOINT="$HOST_ADDRESS:$SERVERPORT"

echo "initializing postgres..."
if [ ! -d "/var/lib/postgresql/data/base" ]; then
  su postgres -c "initdb -D /var/lib/postgresql/data"
fi

su - postgres -c "postgres -D /var/lib/postgresql/data -k /run/postgresql" &

echo "initializing redis..."
redis-server &

sleep 5

echo "Enabling Redis expiration notifications..."
redis-cli config set notify-keyspace-events Ex

echo "waiting for services to start..."
sleep 10

echo "checking database database..."
psql -U postgres -c "CREATE USER ${POSTGRES_USER} WITH PASSWORD '${POSTGRES_PASSWORD}';"
psql -U postgres -c "CREATE DATABASE ${POSTGRES_DB};"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${POSTGRES_DB} TO ${POSTGRES_USER};"

echo "executing migrations..."
/usr/local/bin/migrate.linux-amd64 -path /migrations-server -database "${DATABASE_URL}" up

echo "Iniciando o Maestro Server..."

/maestro-server &

exec /init