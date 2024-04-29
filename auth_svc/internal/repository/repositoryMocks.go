package repository

import (
	"errors"
	"gitlab.com/grpasr/asonrythme/auth_svc/internal/models"
	"sync"
)

/************
* Mock the UserStore, implement the IUserStore
*************/
type UserStoreMock struct {
	str map[string]models.UserDatas
	sync.RWMutex
}

func NewUserStoreMock() *UserStoreMock {
	return &UserStoreMock{
		str: make(map[string]models.UserDatas),
	}
}

func (as *UserStoreMock) UserIsEmailExist(email string) error {
	as.RLock()
	_, ok := as.str[email]
	as.RUnlock()
	if ok {
		return errors.New("email already exist")
	}
	return nil
}

func (as *UserStoreMock) UserCreate(k string, d models.UserDatas) error {
	as.Lock()
	as.str[k] = d
	as.Unlock()
	return nil
}

func (as *UserStoreMock) UserUpdate(k string, d models.UserDatas) error {
	as.Lock()
	as.str[k] = d
	as.Unlock()
	return nil
}

func (as *UserStoreMock) UserUpdateTokens(k string, refreshTK, jwtRefreshToken string) error {
	as.Lock()
	dt, ok := as.str[k]
	if !ok {
		as.Unlock()
		return errors.New("unfound email")
	}
	dt.RefreshTK = refreshTK
	dt.RefreshJWT = jwtRefreshToken
	as.str[k] = dt
	as.Unlock()
	return nil
}

func (as *UserStoreMock) UserGetByEmail(k string) (models.UserDatas, error) {
	as.RLock()
	dt, ok := as.str[k]
	as.RUnlock()
	if !ok {
		return models.UserDatas{}, errors.New("Not found")
	}
	return dt, nil
}

func (as *UserStoreMock) UserDelete(k string) error {
	as.Lock()
	delete(as.str, k)
	as.Unlock()
	return nil
}

func (as *UserStoreMock) UserCount() int {
	var c int
	as.RLock()
	for range as.str {
		c++
	}
	as.RUnlock()
	return c
}

func (as *UserStoreMock) UserReset() {
	as.Lock()
	as.str = make(map[string]models.UserDatas)
	as.Unlock()
}

/************
* APIserverMock mock the APIserver, implement the IAPIserverStore
*************/
type APIserverStoreMock struct {
	str map[string]models.APIserverDatas
	sync.RWMutex
}

func NewAPIserverStoreMock() *APIserverStoreMock {
	return &APIserverStoreMock{
		str: make(map[string]models.APIserverDatas),
	}
}

func (as *APIserverStoreMock) APIserverCreate(k string, d models.APIserverDatas) error {
	as.Lock()
	as.str[k] = d
	as.Unlock()
	return nil
}

func (as *APIserverStoreMock) APIserverUpdateTokens(svcID, refreshTK, jwtRefreshToken string) error {
	as.Lock()
	dt, ok := as.str[svcID]
	if !ok {
		as.Unlock()
		return errors.New("svcID not found")
	}
	dt.RefreshTK = refreshTK
	dt.RefreshJWT = jwtRefreshToken
	as.str[svcID] = dt
	as.Unlock()
	return nil
}

func (as *APIserverStoreMock) APIserverUpdate(svcID string, newData models.APIserverDatas) error {
	as.Lock()
	as.str[svcID] = newData
	as.Unlock()
	return nil
}

func (as *APIserverStoreMock) APIserverGetByID(k string) (models.APIserverDatas, error) {
	as.RLock()
	dt, ok := as.str[k]
	as.RUnlock()
	if !ok {
		return models.APIserverDatas{}, errors.New("Not found")
	}
	return dt, nil
}

func (as *APIserverStoreMock) APIserverDelete(k string) error {
	as.Lock()
	delete(as.str, k)
	as.Unlock()
	return nil
}

func (as *APIserverStoreMock) APIserverCount() int {
	var c int
	as.RLock()
	for range as.str {
		c++
	}
	as.RUnlock()
	return c
}

func (as *APIserverStoreMock) APIserverReset() {
	as.Lock()
	as.str = make(map[string]models.APIserverDatas)
	as.Unlock()
}

/****************
* RedisMock mock the RedisStore, implement the IRedisStore
****************/
type RedisMock struct {
	userStr map[string]models.UserRedisDatas
	apiStr  map[string]models.APIserverRedisDatas
	sync.RWMutex
}

func NewRedisMock() *RedisMock {
	return &RedisMock{
		userStr: make(map[string]models.UserRedisDatas),
		apiStr:  make(map[string]models.APIserverRedisDatas),
	}
}

// RedisAPIservice
func (as *RedisMock) RedisAPIserverSet(k string, d models.APIserverRedisDatas) error {
	as.Lock()
	as.apiStr[k] = d
	as.Unlock()
	return nil
}

func (as *RedisMock) RedisAPIserverGet(k string) (models.APIserverRedisDatas, error) {
	as.RLock()
	dt, ok := as.apiStr[k]
	as.RUnlock()
	if !ok {
		return models.APIserverRedisDatas{}, errors.New("Not found")
	}
	return dt, nil
}

func (as *RedisMock) RedisAPIserverDelete(k string) error {
	as.Lock()
	delete(as.apiStr, k)
	as.Unlock()
	return nil
}

func (as *RedisMock) RedisAPIserverCount() int {
	var c int
	as.RLock()
	for range as.apiStr {
		c++
	}
	as.RUnlock()
	return c
}

func (as *RedisMock) RedisAPIserverReset() {
	as.Lock()
	as.apiStr = make(map[string]models.APIserverRedisDatas)
	as.Unlock()
}

// user
func (as *RedisMock) RedisUserSet(k string, d models.UserRedisDatas) error {
	as.Lock()
	as.userStr[k] = d
	as.Unlock()
	return nil
}

func (as *RedisMock) RedisUserGet(k string) (models.UserRedisDatas, error) {
	as.RLock()
	dt, ok := as.userStr[k]
	as.RUnlock()
	if !ok {
		return models.UserRedisDatas{}, errors.New("Not found")
	}
	return dt, nil
}

func (as *RedisMock) RedisUserDelete(k string) error {
	as.Lock()
	delete(as.userStr, k)
	as.Unlock()
	return nil
}

func (as *RedisMock) RedisUserCount() int {
	var c int
	as.RLock()
	for range as.userStr {
		c++
	}
	as.RUnlock()
	return c
}

func (as *RedisMock) RedisUserReset() {
	as.Lock()
	as.userStr = make(map[string]models.UserRedisDatas)
	as.Unlock()
}
