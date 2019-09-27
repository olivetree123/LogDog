package common

import (
	//"github.com/go-redis/redis"
	"sync"
)

var RedisConnections sync.Map

//type RedisConnections struct {
//	sync.Map
//	lock        sync.Mutex
//	connections map[string]*redis.Client
//}
