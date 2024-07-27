ARG GO_VERSION=1.22

FROM golang:${GO_VERSION} AS builder

WORKDIR /src

COPY ./go.mod /src/
COPY ./go.sum /src/

RUN go mod download && go mod verify

COPY . /src/

RUN CGO_ENABLED=0 \
    go build -o rogueserver

RUN chmod +x /src/rogueserver

# ---------------------------------------------

FROM scratch

WORKDIR /app

COPY --from=builder /src/rogueserver .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8001

ENTRYPOINT ["./rogueserver"]
