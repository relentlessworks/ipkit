package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func doRequest(method, path string) *httptest.ResponseRecorder {
	h := NewHandler("test-secret")
	mux := h.Routes()
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func doRequestJSON(method, path string) *httptest.ResponseRecorder {
	h := NewHandler("test-secret")
	mux := h.Routes()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func TestIPInfo_PlainText(t *testing.T) {
	w := doRequest("GET", "/ip/192.168.1.1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "ip=192.168.1.1") {
		t.Errorf("expected ip=192.168.1.1 in response, got: %s", body)
	}
	if !strings.Contains(body, "version=4") {
		t.Errorf("expected version=4 in response, got: %s", body)
	}
	if !strings.Contains(body, "type=private") {
		t.Errorf("expected type=private in response, got: %s", body)
	}
}

func TestIPInfo_JSON(t *testing.T) {
	w := doRequestJSON("GET", "/ip/192.168.1.1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var info map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if info["ip"] != "192.168.1.1" {
		t.Errorf("expected ip=192.168.1.1, got %v", info["ip"])
	}
}

func TestIPInfo_IPv6(t *testing.T) {
	w := doRequest("GET", "/ip/::1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "version=6") {
		t.Errorf("expected version=6, got: %s", body)
	}
	if !strings.Contains(body, "type=loopback") {
		t.Errorf("expected type=loopback, got: %s", body)
	}
}

func TestIPInfo_Invalid(t *testing.T) {
	w := doRequest("GET", "/ip/not-an-ip")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "error:") {
		t.Errorf("expected error in response, got: %s", body)
	}
	if !strings.Contains(body, "hint:") {
		t.Errorf("expected hint in response, got: %s", body)
	}
}

func TestIPInfo_MissingIP(t *testing.T) {
	w := doRequest("GET", "/ip/")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidate_Valid(t *testing.T) {
	w := doRequest("GET", "/validate?ip=192.168.1.1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "valid=true") {
		t.Errorf("expected valid=true, got: %s", w.Body.String())
	}
}

func TestValidate_Invalid(t *testing.T) {
	w := doRequest("GET", "/validate?ip=not-an-ip")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "valid=false") {
		t.Errorf("expected valid=false, got: %s", w.Body.String())
	}
}

func TestValidate_MissingParam(t *testing.T) {
	w := doRequest("GET", "/validate")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBinary(t *testing.T) {
	w := doRequest("GET", "/binary?ip=192.168.1.1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "binary=11000000.10101000.00000001.00000001") {
		t.Errorf("expected binary representation, got: %s", w.Body.String())
	}
}

func TestToInt(t *testing.T) {
	w := doRequest("GET", "/toint?ip=192.168.1.1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "integer=3232235777") {
		t.Errorf("expected integer=3232235777, got: %s", w.Body.String())
	}
}

func TestFromInt(t *testing.T) {
	w := doRequest("GET", "/fromint?n=3232235777")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ip=192.168.1.1") {
		t.Errorf("expected ip=192.168.1.1, got: %s", w.Body.String())
	}
}

func TestCIDRInfo(t *testing.T) {
	w := doRequest("GET", "/cidr/192.168.1.0/24")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "network=192.168.1.0") {
		t.Errorf("expected network=192.168.1.0, got: %s", body)
	}
	if !strings.Contains(body, "broadcast=192.168.1.255") {
		t.Errorf("expected broadcast=192.168.1.255, got: %s", body)
	}
	if !strings.Contains(body, "prefix=24") {
		t.Errorf("expected prefix=24, got: %s", body)
	}
}

func TestCIDRInfo_Invalid(t *testing.T) {
	w := doRequest("GET", "/cidr/not-a-cidr")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestContains_True(t *testing.T) {
	w := doRequest("GET", "/contains?ip=192.168.1.5&cidr=192.168.1.0/24")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "contains=true") {
		t.Errorf("expected contains=true, got: %s", w.Body.String())
	}
}

func TestContains_False(t *testing.T) {
	w := doRequest("GET", "/contains?ip=10.0.0.5&cidr=192.168.1.0/24")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "contains=false") {
		t.Errorf("expected contains=false, got: %s", w.Body.String())
	}
}

func TestContains_MissingIP(t *testing.T) {
	w := doRequest("GET", "/contains?cidr=192.168.1.0/24")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSubnet(t *testing.T) {
	w := doRequest("GET", "/subnet?cidr=192.168.1.0/24&prefix=26")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "192.168.1.0/26") {
		t.Errorf("expected 192.168.1.0/26, got: %s", body)
	}
	if !strings.Contains(body, "192.168.1.64/26") {
		t.Errorf("expected 192.168.1.64/26, got: %s", body)
	}
	if !strings.Contains(body, "192.168.1.128/26") {
		t.Errorf("expected 192.168.1.128/26, got: %s", body)
	}
	if !strings.Contains(body, "192.168.1.192/26") {
		t.Errorf("expected 192.168.1.192/26, got: %s", body)
	}
}

func TestSubnet_JSON(t *testing.T) {
	w := doRequestJSON("GET", "/subnet?cidr=192.168.1.0/24&prefix=26")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if int(result["count"].(float64)) != 4 {
		t.Errorf("expected count=4, got %v", result["count"])
	}
}

func TestCompare(t *testing.T) {
	w := doRequest("GET", "/compare?ip1=10.0.0.1&ip2=10.0.0.2")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "comparison=before") {
		t.Errorf("expected comparison=before, got: %s", w.Body.String())
	}
}

func TestCompare_Equal(t *testing.T) {
	w := doRequest("GET", "/compare?ip1=10.0.0.1&ip2=10.0.0.1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "comparison=equal") {
		t.Errorf("expected comparison=equal, got: %s", w.Body.String())
	}
}

func TestHelp(t *testing.T) {
	w := doRequest("GET", "/help")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "ipkit") {
		t.Errorf("expected ipkit in help, got: %s", body)
	}
	if !strings.Contains(body, "/ip/") {
		t.Errorf("expected /ip/ in help, got: %s", body)
	}
}

func TestAgentMD(t *testing.T) {
	w := doRequest("GET", "/.well-known/agent.md")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMCP_Initialize(t *testing.T) {
	h := NewHandler("test-secret")
	mux := h.Routes()
	body := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	req := httptest.NewRequest("POST", "/mcp", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp MCPResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
}

func TestMCP_ToolsList(t *testing.T) {
	h := NewHandler("test-secret")
	mux := h.Routes()
	body := strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	req := httptest.NewRequest("POST", "/mcp", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatal("expected result in response")
	}
	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("expected tools array in response")
	}
	if len(tools) != 10 {
		t.Errorf("expected 10 tools, got %d", len(tools))
	}
}

func TestMCP_ToolCall_ParseIP(t *testing.T) {
	h := NewHandler("test-secret")
	mux := h.Routes()
	body := strings.NewReader(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"parse_ip","arguments":{"ip":"192.168.1.1"}}}`)
	req := httptest.NewRequest("POST", "/mcp", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatal("expected result in response")
	}
	content, ok := result["content"].([]interface{})
	if !ok {
		t.Fatal("expected content array in result")
	}
	if len(content) == 0 {
		t.Fatal("expected at least one content item")
	}
}

func TestMCP_ToolCall_ValidateIP(t *testing.T) {
	h := NewHandler("test-secret")
	mux := h.Routes()
	body := strings.NewReader(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"validate_ip","arguments":{"ip":"192.168.1.1"}}}`)
	req := httptest.NewRequest("POST", "/mcp", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMCP_ToolCall_CIDRContains(t *testing.T) {
	h := NewHandler("test-secret")
	mux := h.Routes()
	body := strings.NewReader(`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"cidr_contains","arguments":{"ip":"192.168.1.5","cidr":"192.168.1.0/24"}}}`)
	req := httptest.NewRequest("POST", "/mcp", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMCP_UnknownMethod(t *testing.T) {
	h := NewHandler("test-secret")
	mux := h.Routes()
	body := strings.NewReader(`{"jsonrpc":"2.0","id":6,"method":"unknown/method"}`)
	req := httptest.NewRequest("POST", "/mcp", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp MCPResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
}

func TestMCP_GetNotAllowed(t *testing.T) {
	w := doRequest("GET", "/mcp")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
