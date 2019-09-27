package process

import (
	"bytes"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	. "logDog/common"
	"logDog/entity"
	"net/http"
)

func OutputHttp(output entity.Output, message map[string]interface{}) error {
	if output.Type != "http" {
		err := errors.New("output type is not HTTP.")
		return err
	}
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", output.HttpURL, bytes.NewBuffer(msgBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, val := range output.HttpHeaders {
		req.Header.Set(key, val)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		Logger.Error("HTTP StatusCode = ?, URL = ?", resp.StatusCode, output.HttpURL)
		return err
	}
	return nil
}

func OutputRedis(output entity.Output, label string, message map[string]interface{}) error {
	if output.Type != "redis" {
		err := errors.New("output type is not redis.")
		return err
	}
	var client *redis.Client
	client1, found := RedisConnections.Load(label)
	if !found {
		client = redis.NewClient(&redis.Options{
			Addr:     output.RedisAddr,
			Password: "",             // no password set
			DB:       output.RedisDB, // use default DB
		})
		_, err := client.Ping().Result()
		if err != nil {
			return err
		}
		RedisConnections.Store(label, client)
	} else {
		client = client1.(*redis.Client)
	}
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	Logger.Info("msgBytes = ", string(msgBytes))
	err = client.LPush(output.RedisKey, string(msgBytes)).Err()
	if err != nil {
		return err
	}
	return nil
}
