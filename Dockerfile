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

EXPOSE 8001

ENTRYPOINT ["./rogueserver"]
