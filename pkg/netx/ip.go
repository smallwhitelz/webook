package netx

import (
	"io"
	"net"
	"net/http"
	"strings"
)

// GetOutboundIP 获得对外发送消息的 IP 地址，获取局域网 IP 地址
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func GetPublicIP() string {
	resp, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		return "Error fetching IP"
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error reading response"
	}

	return strings.TrimSpace(string(ip))
}
