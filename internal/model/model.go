package model

import (
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
)

// IPInfo holds parsed information about an IP address.
type IPInfo struct {
	IP          string `json:"ip"`
	Version     int    `json:"version"`
	Type        string `json:"type"`
	IsPrivate   bool   `json:"is_private"`
	IsLoopback  bool   `json:"is_loopback"`
	IsLinkLocal bool   `json:"is_link_local"`
	IsMulticast bool   `json:"is_multicast"`
	IsReserved  bool   `json:"is_reserved"`
	IsPublic    bool   `json:"is_public"`
	Binary      string `json:"binary"`
	ReverseDNS  string `json:"reverse_dns"`
}

// CIDRInfo holds parsed information about a CIDR block.
type CIDRInfo struct {
	CIDR      string `json:"cidr"`
	Network   string `json:"network"`
	Broadcast string `json:"broadcast"`
	Mask      string `json:"mask"`
	PrefixLen int    `json:"prefix_len"`
	IPCount   uint64 `json:"ip_count"`
	FirstIP   string `json:"first_ip"`
	LastIP    string `json:"last_ip"`
	IsIPv4   bool   `json:"is_ipv4"`
}

// ParseIP parses an IP address string and returns detailed info.
func ParseIP(ipStr string) (*IPInfo, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	version := 4
	if ip.To4() == nil {
		version = 6
	}

	isPrivate := ip.IsPrivate()
	isLoopback := ip.IsLoopback()
	isLinkLocal := ip.IsLinkLocalUnicast()
	isMulticast := ip.IsMulticast()
	isUnspecified := ip.IsUnspecified()

	isReserved := isUnspecified
	isPublic := !isPrivate && !isLoopback && !isLinkLocal && !isMulticast && !isReserved

	// Determine type
	ipType := "public"
	if isLoopback {
		ipType = "loopback"
	} else if isPrivate {
		ipType = "private"
	} else if isLinkLocal {
		ipType = "link-local"
	} else if isMulticast {
		ipType = "multicast"
	} else if isReserved {
		ipType = "reserved"
	}

	binary := ipToBinary(ip, version)

	// Reverse DNS (best effort)
	reverseDNS := ""
	names, err := net.LookupAddr(ipStr)
	if err == nil && len(names) > 0 {
		reverseDNS = names[0]
	}

	return &IPInfo{
		IP:          ipStr,
		Version:     version,
		Type:        ipType,
		IsPrivate:   isPrivate,
		IsLoopback:  isLoopback,
		IsLinkLocal: isLinkLocal,
		IsMulticast: isMulticast,
		IsReserved:  isReserved,
		IsPublic:    isPublic,
		Binary:      binary,
		ReverseDNS:  reverseDNS,
	}, nil
}

// ipToBinary converts an IP address to its binary string representation.
func ipToBinary(ip net.IP, version int) string {
	var bits []string
	if version == 4 {
		ip4 := ip.To4()
		for _, b := range ip4 {
			bits = append(bits, fmt.Sprintf("%08b", b))
		}
		return strings.Join(bits, ".")
	}
	for _, b := range ip.To16() {
		bits = append(bits, fmt.Sprintf("%08b", b))
	}
	return strings.Join(bits, ":")
}

// ValidateIP returns true if the string is a valid IP address.
func ValidateIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}

// ParseCIDR parses a CIDR notation string and returns detailed info.
func ParseCIDR(cidrStr string) (*CIDRInfo, error) {
	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR notation: %s", cidrStr)
	}

	ones, _ := ipNet.Mask.Size()
	isIPv4 := ipNet.IP.To4() != nil

	var ipCount uint64
	var firstIP, lastIP, broadcast string

	if isIPv4 {
		mask := ipNet.Mask
		ip4 := ipNet.IP.To4()

		firstIP = ipNet.IP.String()

		broadcastIP := make(net.IP, 4)
		for i := 0; i < 4; i++ {
			broadcastIP[i] = ip4[i] | ^mask[i]
		}
		broadcast = broadcastIP.String()
		lastIP = broadcast

		hostBits := 32 - ones
		if hostBits >= 64 {
			ipCount = 0
		} else {
			ipCount = 1 << uint(hostBits)
		}
	} else {
		mask := ipNet.Mask
		ip6 := ipNet.IP.To16()

		firstIP = ipNet.IP.String()

		broadcastIP := make(net.IP, 16)
		for i := 0; i < 16; i++ {
			broadcastIP[i] = ip6[i] | ^mask[i]
		}
		broadcast = broadcastIP.String()
		lastIP = broadcast

		hostBits := 128 - ones
		if hostBits >= 64 {
			ipCount = 0
		} else {
			ipCount = 1 << uint(hostBits)
		}
	}

	return &CIDRInfo{
		CIDR:      cidrStr,
		Network:   firstIP,
		Broadcast: broadcast,
		Mask:      net.IP(ipNet.Mask).String(),
		PrefixLen:  ones,
		IPCount:   ipCount,
		FirstIP:   firstIP,
		LastIP:    lastIP,
		IsIPv4:    isIPv4,
	}, nil
}

// IsInCIDR checks if an IP address is within a CIDR range.
func IsInCIDR(ipStr, cidrStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR notation: %s", cidrStr)
	}

	return ipNet.Contains(ip), nil
}

// IPToInt converts an IP address to its integer representation.
func IPToInt(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	if ip.To4() != nil {
		ip4 := ip.To4()
		val := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
		return strconv.FormatUint(uint64(val), 10), nil
	}

	// IPv6 — use big.Int
	ip6 := ip.To16()
	bigVal := new(big.Int).SetBytes(ip6)
	return bigVal.String(), nil
}

// IntToIPv4 converts a uint32 integer to an IPv4 address string.
func IntToIPv4(valStr string) (string, error) {
	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid integer: %s", valStr)
	}
	if val > 0xFFFFFFFF {
		return "", fmt.Errorf("value too large for IPv4: %d", val)
	}
	ip := net.IPv4(byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
	return ip.String(), nil
}

// SubnetDivide divides a CIDR into smaller subnets of the given prefix length.
func SubnetDivide(cidrStr string, newPrefix int) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR notation: %s", cidrStr)
	}

	ones, bits := ipNet.Mask.Size()
	if newPrefix < ones {
		return nil, fmt.Errorf("new prefix length %d must be >= current prefix length %d", newPrefix, ones)
	}
	if newPrefix > bits {
		return nil, fmt.Errorf("new prefix length %d exceeds maximum %d", newPrefix, bits)
	}

	shift := newPrefix - ones
	if shift > 20 {
		return nil, fmt.Errorf("too many subnets (shift=%d), max 20", shift)
	}
	count := 1 << uint(shift)

	var subnets []string
	if bits == 32 {
		baseInt := uint32(ipNet.IP.To4()[0])<<24 | uint32(ipNet.IP.To4()[1])<<16 | uint32(ipNet.IP.To4()[2])<<8 | uint32(ipNet.IP.To4()[3])
		hostBits := 32 - newPrefix
		subnetSize := uint32(1 << uint(hostBits))
		for i := 0; i < count; i++ {
			offset := uint32(i) * subnetSize
			subnetIP := net.IPv4(byte((baseInt+offset)>>24), byte((baseInt+offset)>>16), byte((baseInt+offset)>>8), byte(baseInt+offset))
			subnets = append(subnets, fmt.Sprintf("%s/%d", subnetIP.String(), newPrefix))
		}
	} else {
		baseBytes := ipNet.IP.To16()
		hostBits := 128 - newPrefix
		subnetSize := new(big.Int).Lsh(big.NewInt(1), uint(hostBits))
		baseInt := new(big.Int).SetBytes(baseBytes)

		for i := 0; i < count; i++ {
			offset := new(big.Int).Mul(big.NewInt(int64(i)), subnetSize)
			subnetInt := new(big.Int).Add(baseInt, offset)
			subnetBytes := subnetInt.Bytes()
			padded := make([]byte, 16)
			copy(padded[16-len(subnetBytes):], subnetBytes)
			subnetIP := net.IP(padded)
			subnets = append(subnets, fmt.Sprintf("%s/%d", subnetIP.String(), newPrefix))
		}
	}

	return subnets, nil
}

// ReverseDNS performs a reverse DNS lookup on an IP address.
func ReverseDNS(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}
	names, err := net.LookupAddr(ipStr)
	if err != nil {
		return "", fmt.Errorf("no reverse DNS found for %s", ipStr)
	}
	if len(names) == 0 {
		return "", fmt.Errorf("no reverse DNS records for %s", ipStr)
	}
	return strings.Join(names, ", "), nil
}

// CompareIPs compares two IP addresses. Returns -1, 0, or 1.
func CompareIPs(ip1Str, ip2Str string) (int, error) {
	ip1 := net.ParseIP(ip1Str)
	if ip1 == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip1Str)
	}
	ip2 := net.ParseIP(ip2Str)
	if ip2 == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip2Str)
	}

	// Normalize to 16-byte representation
	ip1 = ip1.To16()
	ip2 = ip2.To16()

	for i := 0; i < 16; i++ {
		if ip1[i] < ip2[i] {
			return -1, nil
		}
		if ip1[i] > ip2[i] {
			return 1, nil
		}
	}
	return 0, nil
}
