// Package lanfind 使用 UDP 广播 + TCP 43210 扫描查找局域网内的 exam-server，再通过 HTTP /api/ping 校验。
package lanfind

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	discoveryUDPPort = 43210
	reqMsg           = "EXAM_SERVER_DISCOVERY"
	examService      = "exam-server"
)

type endpoint struct {
	host    string
	apiPort int
}

// Discover 依次尝试：本机 HTTP → UDP 广播 → TCP:43210 → 网段 HTTP 扫描（与浏览器兜底一致）。
func Discover(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	client := newDiscoverHTTPClient(1150 * time.Millisecond)

	for _, base := range []string{
		"http://127.0.0.1:8080",
		"http://127.0.0.1:80",
		"http://localhost:8080",
		"http://[::1]:8080",
		"http://[::1]:80",
	} {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if pingExam(ctx, client, base) {
			return strings.TrimRight(base, "/"), nil
		}
	}

	var ordered []endpoint
	seen := map[string]bool{}

	add := func(list []endpoint) {
		for _, e := range list {
			k := e.host + ":" + strconv.Itoa(e.apiPort)
			if seen[k] {
				continue
			}
			seen[k] = true
			ordered = append(ordered, e)
		}
	}

	add(udpPhase(ctx))
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	prefixes := mergedSweepPrefixes()
	add(tcpPhase(ctx, prefixes))

	for _, e := range ordered {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		base := fmt.Sprintf("http://%s:%d", e.host, e.apiPort)
		if pingExam(ctx, client, base) {
			return strings.TrimRight(base, "/"), nil
		}
	}

	// 许多网络会拦截 UDP 广播或对 43210 无响应，直接按网段扫 HTTP（与前端逻辑一致）
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if hit := httpSweepPhase(ctx, prefixes, []int{8080, 80}); hit != "" {
		return hit, nil
	}

	return "", nil
}

// 系统环境变量 HTTP_PROXY 等会把请求发给代理；访问本机/私网 exam-server 时必须直连，否则同机也发现不了。
func newDiscoverHTTPClient(timeout time.Duration) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = func(req *http.Request) (*url.URL, error) {
		host := req.URL.Hostname()
		if shouldBypassProxyForHost(host) {
			return nil, nil
		}
		return http.ProxyFromEnvironment(req)
	}
	return &http.Client{Timeout: timeout, Transport: tr}
}

func shouldBypassProxyForHost(host string) bool {
	host = strings.Trim(host, "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

func pingExam(ctx context.Context, c *http.Client, base string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/api/ping", nil)
	if err != nil {
		return false
	}
	resp, err := c.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var wrap struct {
		Code int `json:"code"`
		Data struct {
			Service string `json:"service"`
		} `json:"data"`
	}
	if json.NewDecoder(resp.Body).Decode(&wrap) != nil {
		return false
	}
	return wrap.Code == 0 && wrap.Data.Service == examService
}

func parseDiscoveryReply(data []byte, remoteHost string) (host string, httpPort int, ok bool) {
	data = bytes.TrimSpace(data)
	var j struct {
		V    int    `json:"v"`
		Svc  string `json:"svc"`
		HTTP int    `json:"http"`
		TLS  bool   `json:"tls"`
	}
	if json.Unmarshal(data, &j) == nil && j.Svc == examService && j.HTTP > 0 && j.HTTP < 65536 {
		return remoteHost, j.HTTP, true
	}
	s := string(data)
	if strings.HasPrefix(s, "EXAM|") {
		p := strings.TrimSpace(strings.TrimPrefix(s, "EXAM|"))
		port, err := strconv.Atoi(p)
		if err == nil && port > 0 && port < 65536 {
			return remoteHost, port, true
		}
	}
	return "", 0, false
}

func ipv4Broadcast(ip net.IP, mask net.IPMask) net.IP {
	ip = ip.To4()
	if ip == nil || len(mask) != net.IPv4len {
		return nil
	}
	out := make(net.IP, net.IPv4len)
	for i := 0; i < net.IPv4len; i++ {
		out[i] = ip[i] | ^mask[i]
	}
	return out
}

func udpBroadcastTargets() []*net.UDPAddr {
	var out []*net.UDPAddr
	seen := map[string]bool{}
	ifaces, err := net.Interfaces()
	if err != nil {
		ifaces = nil
	}
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagUp == 0 || ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			ip := ipnet.IP.To4()
			mask := ipnet.Mask
			if len(mask) != 4 {
				continue
			}
			bcast := ipv4Broadcast(ip, mask)
			if bcast == nil {
				continue
			}
			key := bcast.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, &net.UDPAddr{IP: bcast, Port: discoveryUDPPort})
		}
	}
	out = append(out, &net.UDPAddr{IP: net.IPv4(255, 255, 255, 255), Port: discoveryUDPPort})
	return out
}

func udpGather(ctx context.Context, pc net.PacketConn, totalWait time.Duration) []endpoint {
	deadline := time.Now().Add(totalWait)
	var mu sync.Mutex
	var list []endpoint
	buf := make([]byte, 1024)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			break
		}
		block := time.Until(deadline)
		if block > 130*time.Millisecond {
			block = 130 * time.Millisecond
		}
		if block <= 0 {
			break
		}
		_ = pc.SetReadDeadline(time.Now().Add(block))
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}
		ua, ok := addr.(*net.UDPAddr)
		if !ok {
			continue
		}
		host, port, ok := parseDiscoveryReply(buf[:n], ua.IP.String())
		if !ok {
			continue
		}
		mu.Lock()
		list = append(list, endpoint{host: host, apiPort: port})
		mu.Unlock()
	}
	return dedupeEndpoints(list)
}

func udpPhase(ctx context.Context) []endpoint {
	pc, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil
	}
	defer pc.Close()

	targets := udpBroadcastTargets()
	payload := []byte(reqMsg)
	for round := 0; round < 3; round++ {
		if ctx.Err() != nil {
			return nil
		}
		for _, dst := range targets {
			_, _ = pc.WriteTo(payload, dst)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(95 * time.Millisecond):
		}
	}
	return udpGather(ctx, pc, 720*time.Millisecond)
}

func interfacePrefixes() []string {
	var out []string
	seen := map[string]bool{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ifaces {
		if ifi.Flags&net.FlagUp == 0 || ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			ip := ipnet.IP.To4()
			pre := fmt.Sprintf("%d.%d.%d", ip[0], ip[1], ip[2])
			if seen[pre] {
				continue
			}
			seen[pre] = true
			out = append(out, pre)
		}
	}
	return out
}

func fallbackPrefixes() []string {
	return []string{
		"192.168.1",
		"192.168.0",
		"192.168.31",
		"192.168.43",
		"192.168.137",
		"172.20.10",
		"172.16.0",
		"10.0.0",
	}
}

// 本机网卡推导的前缀 + 常见私网段（热点/路由），避免只扫到其中一类而漏掉服务端所在网段。
func mergedSweepPrefixes() []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range interfacePrefixes() {
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	for _, p := range fallbackPrefixes() {
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func httpSweepPhase(ctx context.Context, prefixes []string, ports []int) string {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := newDiscoverHTTPClient(780 * time.Millisecond)
	jobs := make(chan string, 4096)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var winner string

	workers := 56
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for base := range jobs {
				if ctx.Err() != nil {
					return
				}
				if pingExam(ctx, client, base) {
					mu.Lock()
					if winner == "" {
						winner = strings.TrimRight(base, "/")
						cancel()
					}
					mu.Unlock()
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, pre := range prefixes {
			for i := 1; i <= 254; i++ {
				ip := fmt.Sprintf("%s.%d", pre, i)
				for _, port := range ports {
					select {
					case <-ctx.Done():
						return
					case jobs <- fmt.Sprintf("http://%s:%d", ip, port):
					}
				}
			}
		}
	}()

	wg.Wait()
	return winner
}

func tcpPhase(ctx context.Context, prefixes []string) []endpoint {
	jobs := make(chan string, 512)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var found []endpoint

	workers := 56
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				if ctx.Err() != nil {
					return
				}
				addr := net.JoinHostPort(ip, strconv.Itoa(discoveryUDPPort))
				conn, err := (&net.Dialer{Timeout: 210 * time.Millisecond}).DialContext(ctx, "tcp", addr)
				if err != nil {
					continue
				}
				_ = conn.SetReadDeadline(time.Now().Add(380 * time.Millisecond))
				buf := make([]byte, 512)
				n, _ := conn.Read(buf)
				_ = conn.Close()
				host, port, ok := parseDiscoveryReply(buf[:n], ip)
				if !ok {
					continue
				}
				mu.Lock()
				found = append(found, endpoint{host: host, apiPort: port})
				mu.Unlock()
			}
		}()
	}

outer:
	for _, pre := range prefixes {
		for i := 1; i <= 254; i++ {
			select {
			case <-ctx.Done():
				break outer
			case jobs <- fmt.Sprintf("%s.%d", pre, i):
			}
		}
	}
	close(jobs)
	wg.Wait()
	return dedupeEndpoints(found)
}

func dedupeEndpoints(in []endpoint) []endpoint {
	seen := map[string]bool{}
	var out []endpoint
	for _, e := range in {
		k := e.host + ":" + strconv.Itoa(e.apiPort)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, e)
	}
	return out
}
