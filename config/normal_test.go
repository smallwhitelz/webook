package config

import (
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"testing"
)

type Req struct {
	ClusterIds *string
}

func Test_normal(t *testing.T) {
	mp := make(map[string]string)
	mp["1"] = "zhangsan"
	mp["2"] = "lisi"
	mp["3"] = "wangwu"
	marshal, err := json.Marshal(mp)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(mp)
	fmt.Println(string(marshal))
}
