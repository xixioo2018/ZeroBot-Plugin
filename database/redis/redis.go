package redis

import (
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/database"
	"github.com/FloatTech/ZeroBot-Plugin/utils"
	"time"

	"github.com/gomodule/redigo/redis"
	logger "github.com/sirupsen/logrus"
)

const (
	setIfNotExist     = "NX" // 不存在则执行
	setWithExpireTime = "PX" // 过期时间(秒)  PX 毫秒
)

var (
	RdPool *redis.Pool
	Nil    = redis.ErrNil
)

func InitRedis(config database.Config) {
	uri := fmt.Sprintf(
		"%s:%d",
		config.Redis.Hostname,
		config.Redis.Port,
	)
	RdPool = &redis.Pool{
		Dial: func() (conn redis.Conn, e error) {
			c, err := redis.Dial("tcp", uri)
			if err != nil {
				return nil, err
			}
			if config.Redis.Password != "" {
				if _, err := c.Do("AUTH", config.Redis.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		MaxIdle:     10,
		MaxActive:   20,
		IdleTimeout: 1000,
	}
}

func Test() {
	conn := RdPool.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	utils.PanicNotNil(err)
}

func DelKey(key string) {
	conn := RdPool.Get()
	defer conn.Close()
	conn.Do("DEL", key)
}

func SetString(key string, value string, duration time.Duration) bool {
	conn := RdPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value, setWithExpireTime, duration.Milliseconds())
	if err != nil {
		logger.Info(err)
	}
	return err == nil
}

func GetString(key string) string {
	conn := RdPool.Get()
	defer conn.Close()
	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		logger.Info(err)
		return ""
	}
	return value
}

func GetStringError(key string) (string, error) {
	conn := RdPool.Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", key))
}

func SetByteArray(key string, value []byte, duration time.Duration) error {
	conn := RdPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value, setWithExpireTime, duration.Milliseconds())
	return err
}

func GetByteArray(key string) []byte {
	conn := RdPool.Get()
	defer conn.Close()
	value, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		logger.Info(err)
		return nil
	}
	return value
}

func GetByteArrayError(key string) ([]byte, error) {
	conn := RdPool.Get()
	defer conn.Close()
	return redis.Bytes(conn.Do("GET", key))
}

func SetInt(key string, value int, duration time.Duration) bool {
	conn := RdPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value, setWithExpireTime, duration.Milliseconds())
	if err != nil {
		logger.Info(err)
	}
	return err == nil
}

func GetInt(key string) (int, error) {
	conn := RdPool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("GET", key))
}

func SetBool(key string, value bool, duration time.Duration) error {
	conn := RdPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value, setWithExpireTime, duration.Milliseconds())
	return err
}

func GetBoolErr(key string) (bool, error) {
	conn := RdPool.Get()
	defer conn.Close()
	return redis.Bool(conn.Do("GET", key))
}
