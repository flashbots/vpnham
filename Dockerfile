# stage: prefetch ------------------------------------------------------

FROM golang:1.22-alpine AS prefetch

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /go/src/github.com/flashbots/vpnham

COPY go.* ./
RUN go mod download

# stage: build ---------------------------------------------------------

FROM prefetch AS build

COPY . .

RUN go build -o bin/vpnham -ldflags "-s -w" github.com/flashbots/vpnham/cmd

# stage: run -----------------------------------------------------------

FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /go/src/github.com/flashbots/vpnham/bin/vpnham ./vpnham

ENTRYPOINT ["/app/vpnham"]
