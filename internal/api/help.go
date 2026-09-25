package api

import (
	"fmt"
	"net/http"
)

func (h *Handler) help(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, helpText)
}

const helpText = `# ipkit — Agentic-First IP Address & Network Utility

ipkit is a headless service for IP address parsing, validation, classification, CIDR calculations, and network utilities. Designed for AI agents to call over plain HTTP.

## Endpoints

### IP Address Operations
  GET /ip/{addr}          Parse and classify an IP address (type, private/public, binary, reverse DNS)
  GET /validate?ip=       Validate if a string is a valid IP address
  GET /binary?ip=         Get binary representation of an IP address
  GET /reverse?ip=        Reverse DNS lookup for an IP address
  GET /toint?ip=          Convert IP address to integer
  GET /fromint?n=         Convert integer to IPv4 address

### CIDR / Network Operations
  GET /cidr/{cidr}        Parse CIDR notation (network, broadcast, mask, IP count, range)
  GET /contains?ip=&cidr= Check if an IP is within a CIDR range
  GET /subnet?cidr=&prefix= Divide a CIDR into smaller subnets

### Comparison
  GET /compare?ip1=&ip2= Compare two IP addresses (before/after/equal)

### MCP
  POST /mcp              Model Context Protocol JSON-RPC 2.0 endpoint

### Meta
  GET /help               This help page
  GET /.well-known/agent.md  Same as /help

## Response Format
  Plain text by default (space-separated key=value pairs, one record per line).
  JSON on demand: send Accept: application/json header or ?format=json query param.

## Errors
  Errors are plain text: error: message | hint: what to do next
  JSON errors: {"error":"message","hint":"what to do next"}

## Examples
  GET /ip/192.168.1.1
  GET /ip/::1
  GET /cidr/10.0.0.0/24
  GET /contains?ip=10.0.0.5&cidr=10.0.0.0/24
  GET /subnet?cidr=192.168.1.0/24&prefix=26
  GET /compare?ip1=10.0.0.1&ip2=10.0.0.2
  GET /toint?ip=192.168.1.1
  GET /fromint?n=3232235777

## Configuration
  IPKIT_ADDR    Listen address (default :7700)
  IPKIT_SECRET  Auth token secret (auto-generated if empty)
  -addr         Listen address flag
  -secret       Auth token secret flag

## Build
  make build    CGO_ENABLED=0 go build -trimpath ./cmd/ipkit
  make test     go test -race ./...
  make vet      go vet ./...
`
