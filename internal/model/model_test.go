package model

import (
	"testing"
)

func TestParseIP_IPv4(t *testing.T) {
	info, err := ParseIP("192.168.1.1")
	if err != nil {
		t.Fatalf("ParseIP failed: %v", err)
	}
	if info.Version != 4 {
		t.Errorf("expected version 4, got %d", info.Version)
	}
	if info.Type != "private" {
		t.Errorf("expected type private, got %s", info.Type)
	}
	if !info.IsPrivate {
		t.Error("expected IsPrivate=true")
	}
	if info.IsPublic {
		t.Error("expected IsPublic=false")
	}
	if info.Binary != "11000000.10101000.00000001.00000001" {
		t.Errorf("unexpected binary: %s", info.Binary)
	}
}

func TestParseIP_IPv4Loopback(t *testing.T) {
	info, err := ParseIP("127.0.0.1")
	if err != nil {
		t.Fatalf("ParseIP failed: %v", err)
	}
	if info.Type != "loopback" {
		t.Errorf("expected type loopback, got %s", info.Type)
	}
	if !info.IsLoopback {
		t.Error("expected IsLoopback=true")
	}
}

func TestParseIP_IPv4Public(t *testing.T) {
	info, err := ParseIP("8.8.8.8")
	if err != nil {
		t.Fatalf("ParseIP failed: %v", err)
	}
	if info.Type != "public" {
		t.Errorf("expected type public, got %s", info.Type)
	}
	if !info.IsPublic {
		t.Error("expected IsPublic=true")
	}
	if info.IsPrivate {
		t.Error("expected IsPrivate=false")
	}
}

func TestParseIP_IPv6Loopback(t *testing.T) {
	info, err := ParseIP("::1")
	if err != nil {
		t.Fatalf("ParseIP failed: %v", err)
	}
	if info.Version != 6 {
		t.Errorf("expected version 6, got %d", info.Version)
	}
	if info.Type != "loopback" {
		t.Errorf("expected type loopback, got %s", info.Type)
	}
}

func TestParseIP_IPv6Private(t *testing.T) {
	info, err := ParseIP("fc00::1")
	if err != nil {
		t.Fatalf("ParseIP failed: %v", err)
	}
	if info.Version != 6 {
		t.Errorf("expected version 6, got %d", info.Version)
	}
	if !info.IsPrivate {
		t.Error("expected IsPrivate=true for fc00::1")
	}
}

func TestParseIP_IPv6LinkLocal(t *testing.T) {
	info, err := ParseIP("fe80::1")
	if err != nil {
		t.Fatalf("ParseIP failed: %v", err)
	}
	if info.Type != "link-local" {
		t.Errorf("expected type link-local, got %s", info.Type)
	}
}

func TestParseIP_Invalid(t *testing.T) {
	_, err := ParseIP("not-an-ip")
	if err == nil {
		t.Error("expected error for invalid IP")
	}
	_, err = ParseIP("999.999.999.999")
	if err == nil {
		t.Error("expected error for invalid IP")
	}
}

func TestValidateIP(t *testing.T) {
	if !ValidateIP("192.168.1.1") {
		t.Error("expected 192.168.1.1 to be valid")
	}
	if !ValidateIP("::1") {
		t.Error("expected ::1 to be valid")
	}
	if ValidateIP("not-an-ip") {
		t.Error("expected not-an-ip to be invalid")
	}
}

func TestParseCIDR_IPv4(t *testing.T) {
	info, err := ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}
	if info.Network != "192.168.1.0" {
		t.Errorf("expected network 192.168.1.0, got %s", info.Network)
	}
	if info.Broadcast != "192.168.1.255" {
		t.Errorf("expected broadcast 192.168.1.255, got %s", info.Broadcast)
	}
	if info.PrefixLen != 24 {
		t.Errorf("expected prefix 24, got %d", info.PrefixLen)
	}
	if info.IPCount != 256 {
		t.Errorf("expected ip_count 256, got %d", info.IPCount)
	}
	if !info.IsIPv4 {
		t.Error("expected IsIPv4=true")
	}
}

func TestParseCIDR_IPv4Small(t *testing.T) {
	info, err := ParseCIDR("10.0.0.0/30")
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}
	if info.IPCount != 4 {
		t.Errorf("expected ip_count 4, got %d", info.IPCount)
	}
	if info.FirstIP != "10.0.0.0" {
		t.Errorf("expected first 10.0.0.0, got %s", info.FirstIP)
	}
	if info.LastIP != "10.0.0.3" {
		t.Errorf("expected last 10.0.0.3, got %s", info.LastIP)
	}
}

func TestParseCIDR_IPv6(t *testing.T) {
	info, err := ParseCIDR("2001:db8::/32")
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}
	if info.IsIPv4 {
		t.Error("expected IsIPv4=false")
	}
	if info.PrefixLen != 32 {
		t.Errorf("expected prefix 32, got %d", info.PrefixLen)
	}
}

func TestParseCIDR_Invalid(t *testing.T) {
	_, err := ParseCIDR("not-a-cidr")
	if err == nil {
		t.Error("expected error for invalid CIDR")
	}
}

func TestIsInCIDR_True(t *testing.T) {
	result, err := IsInCIDR("192.168.1.5", "192.168.1.0/24")
	if err != nil {
		t.Fatalf("IsInCIDR failed: %v", err)
	}
	if !result {
		t.Error("expected 192.168.1.5 to be in 192.168.1.0/24")
	}
}

func TestIsInCIDR_False(t *testing.T) {
	result, err := IsInCIDR("10.0.0.5", "192.168.1.0/24")
	if err != nil {
		t.Fatalf("IsInCIDR failed: %v", err)
	}
	if result {
		t.Error("expected 10.0.0.5 to NOT be in 192.168.1.0/24")
	}
}

func TestIsInCIDR_IPv6(t *testing.T) {
	result, err := IsInCIDR("2001:db8::1", "2001:db8::/32")
	if err != nil {
		t.Fatalf("IsInCIDR failed: %v", err)
	}
	if !result {
		t.Error("expected 2001:db8::1 to be in 2001:db8::/32")
	}
}

func TestIPToInt_IPv4(t *testing.T) {
	result, err := IPToInt("192.168.1.1")
	if err != nil {
		t.Fatalf("IPToInt failed: %v", err)
	}
	if result != "3232235777" {
		t.Errorf("expected 3232235777, got %s", result)
	}
}

func TestIPToInt_IPv6(t *testing.T) {
	result, err := IPToInt("::1")
	if err != nil {
		t.Fatalf("IPToInt failed: %v", err)
	}
	if result != "1" {
		t.Errorf("expected 1, got %s", result)
	}
}

func TestIntToIPv4(t *testing.T) {
	result, err := IntToIPv4("3232235777")
	if err != nil {
		t.Fatalf("IntToIPv4 failed: %v", err)
	}
	if result != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", result)
	}
}

func TestIntToIPv4_Zero(t *testing.T) {
	result, err := IntToIPv4("0")
	if err != nil {
		t.Fatalf("IntToIPv4 failed: %v", err)
	}
	if result != "0.0.0.0" {
		t.Errorf("expected 0.0.0.0, got %s", result)
	}
}

func TestIntToIPv4_TooLarge(t *testing.T) {
	_, err := IntToIPv4("4294967296")
	if err == nil {
		t.Error("expected error for value > 2^32-1")
	}
}

func TestSubnetDivide_IPv4(t *testing.T) {
	subnets, err := SubnetDivide("192.168.1.0/24", 26)
	if err != nil {
		t.Fatalf("SubnetDivide failed: %v", err)
	}
	if len(subnets) != 4 {
		t.Errorf("expected 4 subnets, got %d", len(subnets))
	}
	if subnets[0] != "192.168.1.0/26" {
		t.Errorf("expected first subnet 192.168.1.0/26, got %s", subnets[0])
	}
	if subnets[1] != "192.168.1.64/26" {
		t.Errorf("expected second subnet 192.168.1.64/26, got %s", subnets[1])
	}
	if subnets[2] != "192.168.1.128/26" {
		t.Errorf("expected third subnet 192.168.1.128/26, got %s", subnets[2])
	}
	if subnets[3] != "192.168.1.192/26" {
		t.Errorf("expected fourth subnet 192.168.1.192/26, got %s", subnets[3])
	}
}

func TestSubnetDivide_SamePrefix(t *testing.T) {
	subnets, err := SubnetDivide("10.0.0.0/24", 24)
	if err != nil {
		t.Fatalf("SubnetDivide failed: %v", err)
	}
	if len(subnets) != 1 {
		t.Errorf("expected 1 subnet, got %d", len(subnets))
	}
}

func TestSubnetDivide_SmallerPrefix(t *testing.T) {
	_, err := SubnetDivide("192.168.1.0/24", 16)
	if err == nil {
		t.Error("expected error for prefix < current prefix")
	}
}

func TestCompareIPs_Before(t *testing.T) {
	result, err := CompareIPs("10.0.0.1", "10.0.0.2")
	if err != nil {
		t.Fatalf("CompareIPs failed: %v", err)
	}
	if result >= 0 {
		t.Error("expected 10.0.0.1 < 10.0.0.2")
	}
}

func TestCompareIPs_Equal(t *testing.T) {
	result, err := CompareIPs("10.0.0.1", "10.0.0.1")
	if err != nil {
		t.Fatalf("CompareIPs failed: %v", err)
	}
	if result != 0 {
		t.Error("expected 10.0.0.1 == 10.0.0.1")
	}
}

func TestCompareIPs_After(t *testing.T) {
	result, err := CompareIPs("10.0.0.2", "10.0.0.1")
	if err != nil {
		t.Fatalf("CompareIPs failed: %v", err)
	}
	if result <= 0 {
		t.Error("expected 10.0.0.2 > 10.0.0.1")
	}
}

func TestCompareIPs_IPv6(t *testing.T) {
	result, err := CompareIPs("2001:db8::1", "2001:db8::2")
	if err != nil {
		t.Fatalf("CompareIPs failed: %v", err)
	}
	if result >= 0 {
		t.Error("expected 2001:db8::1 < 2001:db8::2")
	}
}

func TestCompareIPs_Invalid(t *testing.T) {
	_, err := CompareIPs("not-an-ip", "10.0.0.1")
	if err == nil {
		t.Error("expected error for invalid IP")
	}
}
