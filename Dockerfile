FROM golang:1.26 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/dns-control \
    ./cmd/dns-control


FROM gcr.io/distroless/static-debian12

COPY --from=builder /out/dns-control /dns-control

COPY configs/config.yaml /etc/dns-control/domains.yaml

EXPOSE 53/udp
EXPOSE 53/tcp
EXPOSE 8080/tcp

ENTRYPOINT ["/dns-control"]