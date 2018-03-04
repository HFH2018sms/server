package lib

import (
	"github.com/go-redis/redis"
)

const LastMessageKey = "sms:last_message"
const DefaultProgram = "none"
const DefaultProgramConfig = ""
const DefaultGlobalConfig = ""

type Redis struct {
	*redis.Client
}

func RedisConnect(socket string) *Redis {
	return &Redis{
		redis.NewClient(&redis.Options{
			Network:  "unix",
			Addr:     "/var/run/redis/redis.sock",
			Password: "",
			DB:       0,
		}),
	}
}

func (r *Redis) UserExists(number PhoneNumber) (bool, error) {
	res := r.Exists(getUserKey(number))
	if res.Err() != nil {
		return false, res.Err()
	}
	if res.Val() == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (r *Redis) SetData(number PhoneNumber, f, data string) error {
	key := getFunctionKey(number, f)
	res := r.Set(key, data, 0)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (r *Redis) GetPrevData(number PhoneNumber, f string) (string, error) {
	key := getFunctionKey(number, f)
	exist := r.Exists(key)
	if exist.Err() != nil {
		err := r.SetData(number, f, DefaultProgramConfig)
		if err != nil {
			return "", err
		}
		return DefaultProgramConfig, nil
	}
	if exist.Val() == 0 {
		return "", nil
	}
	res := r.Get(key)
	if res.Err() != nil {
		return "", res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) SetGlobalData(f, data string) error {
	key := getGlobalDataKey(f)
	res := r.Set(key, data, 0)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (r *Redis) GetGlobalData(f string) (string, error) {
	key := getGlobalDataKey(f)
	exist := r.Exists(key)
	if exist.Err() != nil {
		err := r.SetGlobalData(f, DefaultGlobalConfig)
		if err != nil {
			return "", err
		}
		return DefaultGlobalConfig, nil
	}
	if exist.Val() == 0 {
		return "", nil
	}
	res := r.Get(key)
	if res.Err() != nil {
		return "", res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) SetCurrentProgram(number PhoneNumber, prog string) error {
	key := getCurrentProgramKey(number)
	return r.Set(key, prog, 0).Err()
}

func (r *Redis) GetCurrentProgram(number PhoneNumber) (string, error) {
	key := getCurrentProgramKey(number)
	res := r.Get(key)
	if res.Err() != nil {
		setRes := r.SetCurrentProgram(number, DefaultProgram)
		if setRes != nil {
			return "", setRes
		}
		return DefaultProgram, nil
	}
	str := res.Val()
	return str, nil
}

func (r *Redis) GetLastMessage() (string, error) {
	res := r.Get(LastMessageKey)
	if res.Err() != nil {
		return "", res.Err()
	}
	str := res.Val()
	return str, nil
}

func (r *Redis) SetLastMessage(sid string) error {
	return r.Set(LastMessageKey, sid, 0).Err()
}

func (r *Redis) LastMessageExists() (bool, error) {
	res := r.Exists(LastMessageKey)
	if res.Err() != nil {
		return false, res.Err()
	}
	if res.Val() == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func getUserKey(number PhoneNumber) string {
	return "user:" + string(number)
}

func getFunctionKey(number PhoneNumber, funcName string) string {
	return getUserKey(number) + ":" + funcName
}

func getCurrentProgramKey(number PhoneNumber) string {
	return getUserKey(number) + ":current_func"
}

func getGlobalDataKey(funcName string) string {
	return "func_dat:" + funcName
}
