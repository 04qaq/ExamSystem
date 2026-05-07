package discovery

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const (
	DiscoveryPort = 43210
	reqMsg        = "EXAM_SERVER_DISCOVERY"
)

var stopChan chan struct{}

// Start 启动 UDP + TCP 服务发现，监听局域网广播
func Start(serverPort string) {
	stopChan = make(chan struct{})

	go startUDP(serverPort)
	go startTCP(serverPort)
}

func startUDP(serverPort string) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		log.Printf("[发现服务-UDP] 地址解析失败: %v", err)
		return
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Printf("[发现服务-UDP] 端口 %d 监听失败: %v", DiscoveryPort, err)
		return
	}

	log.Printf("[发现服务-UDP] 已启动，监听 UDP 端口 %d", DiscoveryPort)

	defer conn.Close()
	buf := make([]byte, 256)
	for {
		select {
		case <-stopChan:
			log.Println("[发现服务-UDP] 已停止")
			return
		default:
			conn.SetReadDeadline(time.Now().Add(time.Second))
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			msg := strings.TrimSpace(string(buf[:n]))
			if msg == reqMsg {
				conn.WriteToUDP([]byte(serverPort), remoteAddr)
				log.Printf("[发现服务-UDP] 响应客户端 %s", remoteAddr.String())
			}
		}
	}
}

// startTCP 启动 TCP 发现监听（解决手机热点阻断 UDP 广播的问题）
func startTCP(serverPort string) {
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		log.Printf("[发现服务-TCP] 地址解析失败: %v", err)
		return
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		log.Printf("[发现服务-TCP] 端口 %d 监听失败: %v", DiscoveryPort, err)
		return
	}

	log.Printf("[发现服务-TCP] 已启动，监听 TCP 端口 %d", DiscoveryPort)

	defer listener.Close()
	for {
		select {
		case <-stopChan:
			log.Println("[发现服务-TCP] 已停止")
			return
		default:
			listener.SetDeadline(time.Now().Add(time.Second))
			conn, err := listener.AcceptTCP()
			if err != nil {
				continue
			}
			conn.Write([]byte(serverPort))
			conn.Close()
		}
	}
}

// Stop 停止服务发现
func Stop() {
	if stopChan != nil {
		close(stopChan)
	}
}
