FROM docker.io/golang:alpine as builder

COPY . /src
WORKDIR /src

ENV CGO_ENABLED=0

RUN apk add git && ./build && go build -o /src/genkeys cmd/genkeys/main.go

FROM docker.io/alpine

COPY --from=builder /src/uqda /usr/bin/uqda
COPY --from=builder /src/uqdactl /usr/bin/uqdactl
COPY --from=builder /src/genkeys /usr/bin/genkeys
COPY contrib/docker/entrypoint.sh /usr/bin/entrypoint.sh

# RUN addgroup -g 1000 -S uqda-network \
#  && adduser -u 1000 -S -g 1000 --home /etc/uqda-network uqda-network
#
# USER uqda-network
# TODO: Make running unprivileged work

VOLUME [ "/etc/uqda-network" ]

ENTRYPOINT [ "/usr/bin/entrypoint.sh" ]
