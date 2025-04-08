FROM golang:1.23-alpine AS builder

WORKDIR /
COPY . .
COPY .docker/entrypoint.sh /entrypoint.sh
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o maestro-server ./cmd/api/main.go

FROM linuxserver/wireguard
RUN apk add --no-cache wireguard-tools

RUN apk add --no-cache \
    openrc \
    postgresql \
    su-exec \
    redis \
    curl && \
    curl -sL https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xz -C /usr/local/bin

RUN mkdir -p /var/lib/postgresql/data \
    && mkdir -p /var/run/postgresql \
    && chown -R postgres:postgres /var/lib/postgresql \
    && chmod 700 /var/lib/postgresql/data

RUN chown postgres:postgres /var/lib/postgresql/data
RUN chown postgres:postgres /run/postgresql

COPY --from=builder /migrations /migrations-server
COPY --from=builder maestro-server .
COPY --from=builder /entrypoint.sh /entrypoint.sh

RUN mkdir -p /config/wg_confs
RUN chmod +x /maestro-server
RUN chmod +x /entrypoint.sh

RUN chmod -R a+r /migrations-server
RUN chmod -R a+r /migrations-server/*.sql

ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=Docker
ENV POSTGRES_DB=maestro_db
ENV DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable
ENV REDIS_HOST=localhost
ENV REDIS_PORT=6379

ENV SERVERPORT=51825
ENV PEERS=root
ENV PEERDNS=10.10.0.1
ENV INTERNAL_SUBNET=10.10.0.0
ENV ALLOWEDIPS=10.10.0.0/24
ENV LOG_CONFS=false

RUN echo $HOST_ADDRESS

VOLUME /var/lib/postgresql/data

ENTRYPOINT ["/entrypoint.sh"]

