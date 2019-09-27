package common

import (
	//"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

func init() {
	//RedisConnections = make(map[string]*redis.Client)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
