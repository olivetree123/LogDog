package common

import (
	//"github.com/go-redis/redis"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

func init() {
	var err error
	//RedisConnections = make(map[string]*redis.Client)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	DockerClient, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
}
