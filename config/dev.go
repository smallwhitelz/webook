//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(43.154.97.245:13316)/webook",
	},
	Redis: RedisConfig{
		Addr: "43.154.97.245:6379",
	},
}
