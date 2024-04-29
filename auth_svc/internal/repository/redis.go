package repository

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/config"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/models"
	"sync"
)

type IRedisStore interface {
	// user
	RedisUserSet(k string, d models.UserRedisDatas) error
	RedisUserGet(k string) (models.UserRedisDatas, error)
	RedisUserDelete(k string) error
	RedisUserCount() int
	RedisUserReset()
	// APIservice
	RedisAPIserverSet(k string, d models.APIserverRedisDatas) error
	RedisAPIserverGet(k string) (models.APIserverRedisDatas, error)
	RedisAPIserverDelete(k string) error
	RedisAPIserverCount() int
	RedisAPIserverReset()
}

type RedisStore struct {
	redisPool *redis.Pool
	// userStr   map[string]models.UserRedisDatas
	sync.RWMutex
}

func NewRedisStore(conf *config.Config) *RedisStore {
	// create the redis pool
	rPool := &redis.Pool{
		MaxIdle:   conf.RdsGetMaxIdle(),
		MaxActive: conf.RdsGetMaxActive(), // max number of connections
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%s", conf.RdsGetAddress(), conf.RdsGetPort()))
		},
	}

	return &RedisStore{
		redisPool: rPool,
		// userStr:   make(map[string]models.UserRedisDatas),
	}
}

// RedisAPIservice
func (rs *RedisStore) RedisAPIserverSet(k string, d models.APIserverRedisDatas) error {
	conn := rs.redisPool.Get()
	defer conn.Close()

	k = models.AuthAPIserverKey(k)

	_, err := conn.Do("HSET", redis.Args{}.Add(k).AddFlat(d)...)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RedisStore) RedisAPIserverGet(k string) (models.APIserverRedisDatas, error) {
	conn := rs.redisPool.Get()
	defer conn.Close()

	fk := models.AuthAPIserverKey(k)

	values, err := redis.Values(conn.Do("HGETALL", fk))
	if err != nil {
		return models.APIserverRedisDatas{}, err
	}

	res := models.APIserverRedisDatas{}

	redis.ScanStruct(values, &res)

	return rs.derializeAPIserver(k, res), nil
}

func (rs *RedisStore) derializeAPIserver(clientID string, apiServer models.APIserverRedisDatas) models.APIserverRedisDatas {
	apiServer.ServiceID = clientID
	return apiServer
}

func (rs *RedisStore) RedisAPIserverDelete(k string) error {
	conn := rs.redisPool.Get()
	defer conn.Close()

	k = models.AuthAPIserverKey(k)

	_, err := conn.Do("DEL", k)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RedisStore) RedisAPIserverCount() int {
	return 0
}

func (rs *RedisStore) RedisAPIserverReset() {
}

// RedisUser
func (rs *RedisStore) RedisUserSet(k string, d models.UserRedisDatas) error {
	conn := rs.redisPool.Get()
	defer conn.Close()

	k = models.AuthUserKey(k)

	_, err := conn.Do("HSET", redis.Args{}.Add(k).AddFlat(d)...)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RedisStore) RedisUserGet(k string) (models.UserRedisDatas, error) {
	conn := rs.redisPool.Get()
	defer conn.Close()

	fk := models.AuthUserKey(k)

	values, err := redis.Values(conn.Do("HGETALL", fk))
	if err != nil {
		return models.UserRedisDatas{}, err
	}

	res := models.UserRedisDatas{}

	redis.ScanStruct(values, &res)

	return rs.derializeUser(k, res), nil
}

func (rs *RedisStore) derializeUser(email string, user models.UserRedisDatas) models.UserRedisDatas {
	user.Email = email
	return user
}

func (rs *RedisStore) RedisUserDelete(k string) error {
	conn := rs.redisPool.Get()
	defer conn.Close()

	k = models.AuthUserKey(k)

	_, err := conn.Do("DEL", k)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RedisStore) RedisUserCount() int {
	return 0
}

func (rs *RedisStore) RedisUserReset() {
}
