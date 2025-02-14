package netx

import (
	"io"
	"log"
	"net/http"
	"testing"
)

func TestGetPublicIP(t *testing.T) {
	resp, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(ip))
}
