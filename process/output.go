package process

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	. "logDog/common"
	"logDog/entity"
	"net/http"
)

func OutputHttp(output entity.Output, message *entity.Message) error {
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

func OutputRedis(output entity.Output, label string, message *entity.Message) error {
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

func OutputMySQL(output entity.Output, label string, message *entity.Message) error {
	if output.Type != "mysql" {
		err := errors.New("output type is not mysql.")
		return err
	}
	Logger.Info(message)
	var err error
	var client *sql.DB
	client1, found := MySQLConnections.Load(label)
	if !found || client1 == nil {
		client, err = sql.Open("mysql", "root:gaojian@/logdog_data")
		if err != nil {
			return err
		}
		MySQLConnections.Store(label, client)
	} else {
		client = client1.(*sql.DB)
	}
	tbName := message.Table
	sqlShow := fmt.Sprintf("SHOW COLUMNS FROM %s", tbName)
	rows, err := client.Query(sqlShow)
	if err != nil {
		return err
	}
	Logger.Info(rows.Columns())
	fields := make(map[string]map[string]string)
	for rows.Next() {
		var fieldName, fieldType, fieldNull string
		var d4, d5, d6 interface{}
		err = rows.Scan(&fieldName, &fieldType, &fieldNull, &d4, &d5, &d6)
		if err != nil {
			return err
		}
		Logger.Info(fieldName, fieldType, fieldNull)
		field := make(map[string]string)
		field["name"] = fieldName
		field["type"] = fieldType
		field["null"] = fieldNull
		fields[fieldName] = field
	}
	Logger.Info(fields)
	condFields := "("
	condValues := ""
	var values []interface{}
	for key, val := range message.Message.(map[string]interface{}) {
		if field, found := fields[key]; found {
			condFields += field["name"] + ","
			condValues += "?,"
			values = append(values, val)
		}
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s values %s", condFields, condValues)
	Logger.Info(insertSQL)
	stmtIns, err := client.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = stmtIns.Exec(values...)
	if err != nil {
		return err
	}
	return nil
}
