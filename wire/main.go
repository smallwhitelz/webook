package wire

import "fmt"

// UseRepository 这里演示使用wire生成的代码，因为没有wire的标签，这里使用的是wire_gen.go
func UseRepository() {
	repo := InitUserRepository()
	fmt.Println(repo)
}
