package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/ipkit/internal/model"
)

// MCPRequest is a JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// MCPResponse is a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCError    `json:"error,omitempty"`
}

// MCError is a JSON-RPC 2.0 error.
type MCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPTool represents a tool in tools/list.
type MCPTool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema MCPInputSchema    `json:"inputSchema"`
}

// MCPInputSchema describes tool parameters.
type MCPInputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]MCPProperty `json:"properties"`
	Required   []string               `json:"required"`
}

// MCPProperty describes a single parameter.
type MCPProperty struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// MCPToolCallParams holds the arguments for a tools/call.
type MCPToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func (h *Handler) mcp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", "use POST for MCP")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeMCPError(w, nil, -32700, "parse error")
		return
	}
	defer r.Body.Close()

	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeMCPError(w, nil, -32700, "parse error")
		return
	}

	switch req.Method {
	case "initialize":
		writeMCPResult(w, req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "ipkit",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{"listChanged": false},
			},
		})

	case "tools/list":
		writeMCPResult(w, req.ID, map[string]interface{}{
			"tools": mcpTools(),
		})

	case "tools/call":
		var params MCPToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			writeMCPError(w, req.ID, -32602, "invalid params")
			return
		}
		result := h.handleMCPToolCall(params)
		writeMCPResult(w, req.ID, result)

	default:
		writeMCPError(w, req.ID, -32601, "method not found: "+req.Method)
	}
}

func writeMCPResult(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeMCPError(w http.ResponseWriter, id interface{}, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &MCError{Code: code, Message: msg},
	})
}

func mcpTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "parse_ip",
			Description: "Parse and classify an IP address. Returns version, type (public/private/loopback/etc), flags, binary representation, and reverse DNS.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"ip": {Type: "string", Description: "IP address to parse (IPv4 or IPv6)"},
				},
				Required: []string{"ip"},
			},
		},
		{
			Name:        "validate_ip",
			Description: "Validate if a string is a valid IP address.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"ip": {Type: "string", Description: "String to validate as IP address"},
				},
				Required: []string{"ip"},
			},
		},
		{
			Name:        "ip_binary",
			Description: "Get the binary representation of an IP address.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"ip": {Type: "string", Description: "IP address"},
				},
				Required: []string{"ip"},
			},
		},
		{
			Name:        "reverse_dns",
			Description: "Perform a reverse DNS lookup on an IP address.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"ip": {Type: "string", Description: "IP address to look up"},
				},
				Required: []string{"ip"},
			},
		},
		{
			Name:        "ip_to_int",
			Description: "Convert an IP address to its integer representation.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"ip": {Type: "string", Description: "IP address to convert"},
				},
				Required: []string{"ip"},
			},
		},
		{
			Name:        "int_to_ip",
			Description: "Convert an integer to an IPv4 address.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"n": {Type: "string", Description: "Integer (0 to 4294967295)"},
				},
				Required: []string{"n"},
			},
		},
		{
			Name:        "parse_cidr",
			Description: "Parse CIDR notation. Returns network address, broadcast, mask, prefix length, IP count, and first/last IP.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"cidr": {Type: "string", Description: "CIDR notation, e.g. 192.168.1.0/24"},
				},
				Required: []string{"cidr"},
			},
		},
		{
			Name:        "cidr_contains",
			Description: "Check if an IP address is within a CIDR range.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"ip":   {Type: "string", Description: "IP address to check"},
					"cidr": {Type: "string", Description: "CIDR range, e.g. 192.168.1.0/24"},
				},
				Required: []string{"ip", "cidr"},
			},
		},
		{
			Name:        "subnet_divide",
			Description: "Divide a CIDR block into smaller subnets of the given prefix length.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"cidr":   {Type: "string", Description: "Parent CIDR, e.g. 192.168.1.0/24"},
					"prefix": {Type: "integer", Description: "New prefix length, e.g. 26"},
				},
				Required: []string{"cidr", "prefix"},
			},
		},
		{
			Name:        "compare_ips",
			Description: "Compare two IP addresses. Returns before/after/equal.",
			InputSchema: MCPInputSchema{
				Type: "object",
				Properties: map[string]MCPProperty{
					"ip1": {Type: "string", Description: "First IP address"},
					"ip2": {Type: "string", Description: "Second IP address"},
				},
				Required: []string{"ip1", "ip2"},
			},
		},
	}
}

func (h *Handler) handleMCPToolCall(params MCPToolCallParams) interface{} {
	getStr := func(key string) string {
		if v, ok := params.Arguments[key].(string); ok {
			return v
		}
		return ""
	}
	getInt := func(key string) int {
		if v, ok := params.Arguments[key].(float64); ok {
			return int(v)
		}
		return 0
	}

	switch params.Name {
	case "parse_ip":
		ipStr := getStr("ip")
		info, err := model.ParseIP(ipStr)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		return mcpTextResult(fmt.Sprintf("ip=%s version=%d type=%s private=%t loopback=%t link_local=%t multicast=%t reserved=%t public=%t binary=%s reverse_dns=%s",
			info.IP, info.Version, info.Type, info.IsPrivate, info.IsLoopback, info.IsLinkLocal, info.IsMulticast, info.IsReserved, info.IsPublic, info.Binary, info.ReverseDNS))

	case "validate_ip":
		ipStr := getStr("ip")
		valid := model.ValidateIP(ipStr)
		return mcpTextResult(fmt.Sprintf("ip=%s valid=%t", ipStr, valid))

	case "ip_binary":
		ipStr := getStr("ip")
		info, err := model.ParseIP(ipStr)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		return mcpTextResult(fmt.Sprintf("ip=%s binary=%s", ipStr, info.Binary))

	case "reverse_dns":
		ipStr := getStr("ip")
		result, err := model.ReverseDNS(ipStr)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		return mcpTextResult(fmt.Sprintf("ip=%s reverse_dns=%s", ipStr, result))

	case "ip_to_int":
		ipStr := getStr("ip")
		result, err := model.IPToInt(ipStr)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		return mcpTextResult(fmt.Sprintf("ip=%s integer=%s", ipStr, result))

	case "int_to_ip":
		nStr := getStr("n")
		result, err := model.IntToIPv4(nStr)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		return mcpTextResult(fmt.Sprintf("integer=%s ip=%s", nStr, result))

	case "parse_cidr":
		cidrStr := getStr("cidr")
		info, err := model.ParseCIDR(cidrStr)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		return mcpTextResult(fmt.Sprintf("cidr=%s network=%s broadcast=%s mask=%s prefix=%d ip_count=%d first=%s last=%s ipv4=%t",
			info.CIDR, info.Network, info.Broadcast, info.Mask, info.PrefixLen, info.IPCount, info.FirstIP, info.LastIP, info.IsIPv4))

	case "cidr_contains":
		ipStr := getStr("ip")
		cidrStr := getStr("cidr")
		result, err := model.IsInCIDR(ipStr, cidrStr)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		return mcpTextResult(fmt.Sprintf("ip=%s cidr=%s contains=%t", ipStr, cidrStr, result))

	case "subnet_divide":
		cidrStr := getStr("cidr")
		prefix := getInt("prefix")
		subnets, err := model.SubnetDivide(cidrStr, prefix)
		if err != nil {
			return mcpErrorResult(err.Error())
		}
		var sb strings.Builder
		for _, s := range subnets {
			sb.WriteString(s)
			sb.WriteString("\n")
		}
		return mcpTextResult(sb.String())

	case "compare_ips":
		ip1 := getStr("ip1")
		ip2 := getStr("ip2")
		result, err := model.CompareIPs(ip1, ip2)
		if err != nil {
			return mcpErrorResult(err.Error())
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
		return mcpTextResult(fmt.Sprintf("ip1=%s ip2=%s comparison=%s result=%d", ip1, ip2, cmp, result))

	default:
		return mcpErrorResult("unknown tool: " + params.Name)
	}
}

func mcpTextResult(text string) interface{} {
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": text},
		},
	}
}

func mcpErrorResult(msg string) interface{} {
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": "error: " + msg},
		},
		"isError": true,
	}
}
