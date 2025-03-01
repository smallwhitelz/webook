package config

import (
	"log"
	"os/user"
	"testing"
)

type Req struct {
	ClusterIds *string
}

func Test_normal(t *testing.T) {
	//m := make(map[int]int)
	//for i := 0; i < 100; i++ {
	//	m[i] = i
	//	fmt.Printf("i=%d, len(m)=%d\n", i, len(m))
	//}

	current, err := user.Current()
	if err != nil {
		log.Println(err)
	}

	log.Println(current.Username)
}
