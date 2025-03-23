package config

import (
	"fmt"
	"testing"
)

type Req struct {
	ClusterIds *string
}

func Test_normal(t *testing.T) {
	a := make(map[int]bool)
	a[1] = false
	a[2] = true
	a[3] = false
	for k, v := range a {
		if v {
			a[10+k] = true
			fmt.Println(v)
		}
	}
	fmt.Println(a)
}
