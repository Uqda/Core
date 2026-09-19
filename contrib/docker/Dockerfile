FROM docker.io/golang:alpine as builder

COPY . /src
WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git && ./build && go build -o /src/genkeys cmd/genkeys/main.go

FROM docker.io/alpine

LABEL org.opencontainers.image.title="Uqda Core"

COPY --from=builder /src/uqda /usr/bin/uqda
COPY --from=builder /src/uqdactl /usr/bin/uqdactl
COPY --from=builder /src/genkeys /usr/bin/genkeys
COPY contrib/docker/entrypoint.sh /usr/bin/entrypoint.sh
COPY contrib/packaging/install-config.sh /usr/bin/uqda-install-config
RUN chmod 755 /usr/bin/entrypoint.sh /usr/bin/uqda-install-config

VOLUME [ "/etc/uqda" ]

ENTRYPOINT [ "/usr/bin/entrypoint.sh" ]
