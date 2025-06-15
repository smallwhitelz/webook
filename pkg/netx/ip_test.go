package netx

import "testing"

func TestIP(t *testing.T) {
	ip := GetOutboundIP()
	t.Log(ip)
}
