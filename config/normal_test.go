package config

import (
	"fmt"
	"testing"
	"time"
)

type Req struct {
	ClusterIds *string
}

func Test_normal(t *testing.T) {
	now := time.Now().UnixMilli()
	fmt.Println(now)
	milli := time.UnixMilli(now)
	fmt.Println(milli)
}
