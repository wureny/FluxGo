#flux_go:
#	. .env && export GO111MODULE=on && go build -v $(LDFLAGS) ./cmd/fluxgo
#
#test: flux_go
#	source .env && export GO111MODULE=on && ./fluxgo TryitOut
#
#.PHONY: flux_go test

include .env
export $(shell sed 's/=.*//' .env)

flux_go:
	. .env && export GO111MODULE=on && go build -v $(LDFLAGS) ./cmd/fluxgo && ./fluxgo TryitOut