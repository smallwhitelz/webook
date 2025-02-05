package config

import (
	"fmt"
	"testing"
)

type Req struct {
	ClusterIds *string
}

func Test_normal(t *testing.T) {
	mod := "12331"
	var req = Req{
		ClusterIds: &mod,
	}
	fmt.Println(*req.ClusterIds)
}
