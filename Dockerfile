# SPDX-FileCopyrightText: 2024-2025 Pagefault Games
#
# SPDX-License-Identifier: AGPL-3.0-or-later
ARG GO_VERSION=1.22

FROM docker.io/library/golang:${GO_VERSION} AS builder

WORKDIR /src

COPY ./go.mod /src/
COPY ./go.sum /src/

RUN go mod download && go mod verify

COPY . /src/

RUN CGO_ENABLED=0 \
    go build -tags=devsetup -o rogueserver

RUN chmod +x /src/rogueserver

# ---------------------------------------------

FROM scratch

WORKDIR /app

COPY --from=builder /src/rogueserver .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8001

ENTRYPOINT ["./rogueserver"]
