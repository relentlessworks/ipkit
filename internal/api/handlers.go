package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/relentlessworks/ipkit/internal/model"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	secret string
}

// NewHandler creates a new API handler.
func NewHandler(secret string) *Handler {
	return &Handler{secret: secret}
}

// Routes returns the HTTP mux with all routes registered.
func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/help", h.help)
	mux.HandleFunc("/.well-known/agent.md", h.help)

	// IP endpoints
	mux.HandleFunc("/ip/", h.ipInfo)       // GET /ip/{addr}
	mux.HandleFunc("/validate", h.validate) // GET /validate?ip={addr}
	mux.HandleFunc("/binary", h.binary)     // GET /binary?ip={addr}
	mux.HandleFunc("/reverse", h.reverse)   // GET /reverse?ip={addr}
	mux.HandleFunc("/toint", h.toInt)       // GET /toint?ip={addr}
	mux.HandleFunc("/fromint", h.fromInt)   // GET /fromint?n={number}

	// CIDR endpoints
	mux.HandleFunc("/cidr/", h.cidrInfo)    // GET /cidr/{cidr}
	mux.HandleFunc("/contains", h.contains) // GET /contains?ip={addr}&cidr={cidr}
	mux.HandleFunc("/subnet", h.subnet)     // GET /subnet?cidr={cidr}&prefix={n}

	// Compare
	mux.HandleFunc("/compare", h.compare) // GET /compare?ip1={a}&ip2={b}

	// MCP
	mux.HandleFunc("/mcp", h.mcp)

	return mux
}

// wantsJSON checks if the client wants JSON response.
func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

// writeError writes a plain text or JSON error response.
func writeError(w http.ResponseWriter, r *http.Request, status int, msg, hint string) {
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{
			"error": msg,
			"hint":  hint,
		})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, "error: %s | hint: %s\n", msg, hint)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// writeText writes a plain text response.
func writeText(w http.ResponseWriter, format string, args ...interface{}) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, format, args...)
}

// --- IP Endpoints ---

func (h *Handler) ipInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	// Extract IP from path: /ip/{addr}
	path := strings.TrimPrefix(r.URL.Path, "/ip/")
	ipStr := strings.TrimPrefix(path, "/")
	if ipStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing IP address", "provide an IP address in the path, e.g. GET /ip/8.8.8.8")
		return
	}

	info, err := model.ParseIP(ipStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide a valid IPv4 or IPv6 address, e.g. 192.168.1.1 or ::1")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, info)
		return
	}

	writeText(w, "ip=%s version=%d type=%s private=%t loopback=%t link_local=%t multicast=%t reserved=%t public=%t binary=%s reverse_dns=%s\n",
		info.IP, info.Version, info.Type, info.IsPrivate, info.IsLoopback, info.IsLinkLocal, info.IsMulticast, info.IsReserved, info.IsPublic, info.Binary, info.ReverseDNS)
}

func (h *Handler) validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	ipStr := r.URL.Query().Get("ip")
	if ipStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing ip parameter", "provide ?ip=192.168.1.1")
		return
	}

	valid := model.ValidateIP(ipStr)
	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{"ip": ipStr, "valid": valid})
		return
	}
	writeText(w, "ip=%s valid=%t\n", ipStr, valid)
}

func (h *Handler) binary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	ipStr := r.URL.Query().Get("ip")
	if ipStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing ip parameter", "provide ?ip=192.168.1.1")
		return
	}

	info, err := model.ParseIP(ipStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide a valid IPv4 or IPv6 address")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, map[string]string{"ip": ipStr, "binary": info.Binary})
		return
	}
	writeText(w, "ip=%s binary=%s\n", ipStr, info.Binary)
}

func (h *Handler) reverse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	ipStr := r.URL.Query().Get("ip")
	if ipStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing ip parameter", "provide ?ip=8.8.8.8")
		return
	}

	result, err := model.ReverseDNS(ipStr)
	if err != nil {
		writeError(w, r, http.StatusNotFound, err.Error(), "the IP may not have reverse DNS records")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, map[string]string{"ip": ipStr, "reverse_dns": result})
		return
	}
	writeText(w, "ip=%s reverse_dns=%s\n", ipStr, result)
}

func (h *Handler) toInt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	ipStr := r.URL.Query().Get("ip")
	if ipStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing ip parameter", "provide ?ip=192.168.1.1")
		return
	}

	result, err := model.IPToInt(ipStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide a valid IPv4 or IPv6 address")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, map[string]string{"ip": ipStr, "integer": result})
		return
	}
	writeText(w, "ip=%s integer=%s\n", ipStr, result)
}

func (h *Handler) fromInt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	nStr := r.URL.Query().Get("n")
	if nStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing n parameter", "provide ?n=3232235777")
		return
	}

	result, err := model.IntToIPv4(nStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide a valid integer (0 to 4294967295)")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, map[string]string{"integer": nStr, "ip": result})
		return
	}
	writeText(w, "integer=%s ip=%s\n", nStr, result)
}

// --- CIDR Endpoints ---

func (h *Handler) cidrInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	// Extract CIDR from path: /cidr/{cidr}
	// CIDR contains a slash, so we need to handle it carefully
	// The path will be /cidr/192.168.1.0/24
	path := strings.TrimPrefix(r.URL.Path, "/cidr/")
	if path == "" {
		writeError(w, r, http.StatusBadRequest, "missing CIDR", "provide a CIDR in the path, e.g. GET /cidr/192.168.1.0/24")
		return
	}

	info, err := model.ParseCIDR(path)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide a valid CIDR notation, e.g. 192.168.1.0/24")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, info)
		return
	}

	writeText(w, "cidr=%s network=%s broadcast=%s mask=%s prefix=%d ip_count=%d first=%s last=%s ipv4=%t\n",
		info.CIDR, info.Network, info.Broadcast, info.Mask, info.PrefixLen, info.IPCount, info.FirstIP, info.LastIP, info.IsIPv4)
}

func (h *Handler) contains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	ipStr := r.URL.Query().Get("ip")
	cidrStr := r.URL.Query().Get("cidr")
	if ipStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing ip parameter", "provide ?ip=192.168.1.5&cidr=192.168.1.0/24")
		return
	}
	if cidrStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing cidr parameter", "provide ?ip=192.168.1.5&cidr=192.168.1.0/24")
		return
	}

	result, err := model.IsInCIDR(ipStr, cidrStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide valid IP and CIDR, e.g. ?ip=192.168.1.5&cidr=192.168.1.0/24")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{"ip": ipStr, "cidr": cidrStr, "contains": result})
		return
	}
	writeText(w, "ip=%s cidr=%s contains=%t\n", ipStr, cidrStr, result)
}

func (h *Handler) subnet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	cidrStr := r.URL.Query().Get("cidr")
	prefixStr := r.URL.Query().Get("prefix")
	if cidrStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing cidr parameter", "provide ?cidr=192.168.1.0/24&prefix=26")
		return
	}
	if prefixStr == "" {
		writeError(w, r, http.StatusBadRequest, "missing prefix parameter", "provide ?cidr=192.168.1.0/24&prefix=26")
		return
	}

	prefix, err := strconv.Atoi(prefixStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid prefix", "provide a numeric prefix length, e.g. prefix=26")
		return
	}

	subnets, err := model.SubnetDivide(cidrStr, prefix)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide valid CIDR and prefix, e.g. ?cidr=192.168.1.0/24&prefix=26")
		return
	}

	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{"cidr": cidrStr, "prefix": prefix, "count": len(subnets), "subnets": subnets})
		return
	}

	var sb strings.Builder
	for _, s := range subnets {
		sb.WriteString(s)
		sb.WriteString("\n")
	}
	writeText(w, "%s", sb.String())
}

// --- Compare ---

func (h *Handler) compare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use GET")
		return
	}

	ip1 := r.URL.Query().Get("ip1")
	ip2 := r.URL.Query().Get("ip2")
	if ip1 == "" {
		writeError(w, r, http.StatusBadRequest, "missing ip1 parameter", "provide ?ip1=192.168.1.1&ip2=192.168.1.2")
		return
	}
	if ip2 == "" {
		writeError(w, r, http.StatusBadRequest, "missing ip2 parameter", "provide ?ip1=192.168.1.1&ip2=192.168.1.2")
		return
	}

	result, err := model.CompareIPs(ip1, ip2)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "provide valid IP addresses")
		return
	}

	var cmp string
	switch {
	case result < 0:
		cmp = "before"
	case result > 0:
		cmp = "after"
	default:
		cmp = "equal"
	}

	if wantsJSON(r) {
		writeJSON(w, map[string]interface{}{"ip1": ip1, "ip2": ip2, "comparison": cmp, "result": result})
		return
	}
	writeText(w, "ip1=%s ip2=%s comparison=%s result=%d\n", ip1, ip2, cmp, result)
}
