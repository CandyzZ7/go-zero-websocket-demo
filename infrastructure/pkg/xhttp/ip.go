package xhttp

import (
	"errors"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}
	TrustedProxies  = []string{"0.0.0.0/0"} // 配置你的可信代理 IP 地址范围
	LocalNetworks   = []string{"127.0.0.0/8", "10.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "172.0.0.0/8", "192.168.0.0/16"}
	WhitelistIPs    = []string{""} // 配置白名单 IP 地址或网络
)

var (
	localNetworks     []*net.IPNet
	localNetworksOnce sync.Once
	whitelistIPs      []*net.IPNet
	whitelistOnce     sync.Once
)

func getLocalNetworks() []*net.IPNet {
	localNetworksOnce.Do(func() {
		localNetworks = make([]*net.IPNet, 0, len(LocalNetworks))
		for _, sNetwork := range LocalNetworks {
			_, network, err := net.ParseCIDR(sNetwork)
			if err == nil {
				localNetworks = append(localNetworks, network)
			}
		}
	})
	return localNetworks
}

// getWhitelistIPs 获取白名单 IP 列表
func getWhitelistIPs() []*net.IPNet {
	whitelistOnce.Do(func() {
		whitelistIPs = make([]*net.IPNet, 0, len(WhitelistIPs))
		for _, sNetwork := range WhitelistIPs {
			_, network, err := net.ParseCIDR(sNetwork)
			if err == nil {
				whitelistIPs = append(whitelistIPs, network)
			}
		}
	})
	return whitelistIPs
}

// IsWhitelistedIP 判断 IP 是否在白名单中
func IsWhitelistedIP(ip string) bool {
	for _, network := range getWhitelistIPs() {
		if network.Contains(net.ParseIP(ip)) {
			return true
		}
	}
	return false
}

// IsLocalAddrIP 检查 IP 是否是本地地址
func IsLocalAddrIP(ip string) bool {
	return IsLocalIP(net.ParseIP(ip))
}

// IsLocalIP 判断 IP 是否是本地网络的地址
func IsLocalIP(ip net.IP) bool {
	for _, network := range getLocalNetworks() {
		if network.Contains(ip) {
			return true
		}
	}
	if ip.String() == "0.0.0.0" {
		return true
	}
	return ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast()
}

func getTrustedIP(r *http.Request, remoteIP string) string {
	ip := remoteIP
	// 白名单中的 IP 直接返回
	if IsWhitelistedIP(ip) {
		return ip
	}
	// 遍历所有可能的代理头，确保通过代理时从可信的 IP 地址中获取客户端 IP
	for _, header := range RemoteIPHeaders {
		val := r.Header.Get(header)
		ips := parseHeadersIP(val)

		// 逐个验证是否是可信代理的 IP
		for _, ipInHeader := range ips {
			if !IsLocalAddrIP(ipInHeader) && !inTrustedProxies(ipInHeader) {
				ip = ipInHeader
				break
			}
		}
	}
	return ip
}

// getRemoteIP 从请求中获取客户端的 IP 地址
func getRemoteIP(r *http.Request) []string {
	ips := make([]string, 0)
	ip := RemoteIP(r)
	trusted := ip == ""
	// 先检查白名单
	if !IsWhitelistedIP(ip) {
		// 如果 IP 不在白名单中，检查是否属于可信代理 IP 地址
		if !trusted {
			for _, proxy := range TrustedProxies {
				if inNetwork(ip, proxy) {
					trusted = true
					break
				}
			}
		}
	}

	// 如果请求经过了可信代理，则解析 X-Forwarded-For 或 X-Real-IP 中的 IP 地址
	if trusted {
		for _, header := range RemoteIPHeaders {
			val := r.Header.Get(header)
			ips = append(ips, parseHeadersIP(val)...)
		}
	}
	return append(ips, ip)
}

// parseHeadersIP 解析 X-Forwarded-For 等头部中的 IP 地址
func parseHeadersIP(val string) []string {
	if val == "" {
		return []string{}
	}
	// 将多个 IP 地址通过逗号分隔
	str := strings.Split(val, ",")
	l := len(str)
	ips := make([]string, l)
	for i := l - 1; i >= 0; i-- {
		ips[l-1-i] = strings.TrimSpace(str[i])
	}
	return ips
}

// GetClientIP 获取客户端 IP 地址
func GetClientIP(r *http.Request) string {
	ips := getRemoteIP(r)
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	return ""
}

// GetClientPublicIP 获取公共 IP 地址
func GetClientPublicIP(r *http.Request) string {
	var ip string
	ips := getRemoteIP(r)
	for _, ip = range ips {
		if ip != "" && !IsLocalAddrIP(ip) {
			return ip
		}
	}
	return ""
}

// RemoteIP 获取请求的真实客户端 IP 地址
func RemoteIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// IPToLong 将 IP 地址转换为长整型
func IPToLong(ip string) (uint, error) {
	return NetIPToLong(net.ParseIP(ip))
}

// LongToIP 将长整型转换为 IP 地址
func LongToIP(i uint) (string, error) {
	ip, err := LongToNetIP(i)
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

// NetIPToLong 将 net.IP 转换为长整型
func NetIPToLong(ip net.IP) (uint, error) {
	b := ip.To4()
	if b == nil {
		return 0, errors.New("invalid ipv4 format")
	}
	i := uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24
	return i, nil
}

// LongToNetIP 将长整型转换为 net.IP
func LongToNetIP(i uint) (net.IP, error) {
	if i > math.MaxUint32 {
		return nil, errors.New("beyond the scope of ipv4")
	}
	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)
	return ip, nil
}

// IsIP 判断字符串是否是有效的 IP 地址
func IsIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// GetIPv 获取 IP 类型（IPv4 或 IPv6）
func GetIPv(s string) int {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return 4
		case ':':
			return 6
		}
	}
	return 0
}

// inNetwork 检查 IP 是否在指定的网络中
func inNetwork(ip, network string) bool {
	n, err := netCIDR(network)
	if err != nil {
		return false
	}
	netIP := net.ParseIP(ip)
	return n.Contains(netIP)
}

// netCIDR 解析 CIDR 网络
func netCIDR(network string) (*net.IPNet, error) {
	_, n, err := net.ParseCIDR(network)
	if err != nil && IsIP(network) {
		_, n, err = net.ParseCIDR(network + "/24")
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

// inTrustedProxies 检查 IP 是否在受信任代理范围内
func inTrustedProxies(ip string) bool {
	for _, proxy := range TrustedProxies {
		if inNetwork(ip, proxy) {
			return true
		}
	}
	return false
}

// GetServerIp 获取服务端 IP
// 问题：我在本地多网卡机器上，运行分布式场景，此函数返回的ip有误导致rpc连接失败。 遂google结果如下：
// 1、https://www.jianshu.com/p/301aabc06972
// 2、https://www.cnblogs.com/chaselogs/p/11301940.html
func GetServerIp() string {
	ip, err := externalIP()
	if err != nil {
		return ""
	}
	return ip.String()
}

func externalIP() (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, err
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}
