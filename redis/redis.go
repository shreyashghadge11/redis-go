package redis

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

type redisObject struct {
	datatype reflect.Type
	value    interface{}
	mu       sync.RWMutex
}

type multiCommand struct {
	commands []string
}

type Redis struct {
	dict             map[string]redisObject
	mu               sync.RWMutex
	multiCommand     multiCommand
	isMultiCommandOn bool
}

func NewRedis() *Redis {
	return &Redis{
		dict:             make(map[string]redisObject),
		isMultiCommandOn: false,
	}
}

func isNumber(s string) bool {
	_, errInt := strconv.Atoi(s)
	_, errFloat := strconv.ParseFloat(s, 64)

	return errInt == nil || errFloat == nil
}

func (r *Redis) Multi() {
	r.isMultiCommandOn = true
}

func (r *Redis) Status() bool {
	return r.isMultiCommandOn
}

func (r *Redis) AddToMultiCommand(command string) {
	r.multiCommand.commands = append(r.multiCommand.commands, command)
}

func (r *Redis) Exec() []string {
	return r.multiCommand.commands
}

func (r *Redis) Discard() {
	r.isMultiCommandOn = false
	r.multiCommand.commands = nil
}

func (r *Redis) Set(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if isNumber(value.(string)) {
		value, _ = strconv.ParseFloat(value.(string), 64)
	}

	r.dict[key] = redisObject{
		datatype: reflect.TypeOf(value),
		value:    value,
	}
}

func (r *Redis) Get(key string) interface{} {
	if _, ok := r.dict[key]; !ok {
		return nil
	}
	fmt.Println(r.dict[key].datatype)
	return r.dict[key].value
}

func (r *Redis) Del(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if obj, ok := r.dict[key]; ok {
		obj.mu.Lock()
		defer obj.mu.Unlock()
		delete(r.dict, key)
		return true
	}

	return false
}

func (r *Redis) Increment(key string) bool {

	if obj, ok := r.dict[key]; ok {
		obj.mu.Lock()
		defer obj.mu.Unlock()
		if obj.datatype == reflect.TypeOf(float64(0)) {
			obj.value = obj.value.(float64) + 1
			r.dict[key] = obj
			return true
		} else {
			fmt.Println("Error: value is not a number")
			return false
		}
	}
	return false
}

func (r *Redis) IncrementBy(key string, increment float64) bool {

	if obj, ok := r.dict[key]; ok {
		obj.mu.Lock()
		defer obj.mu.Unlock()
		if obj.datatype == reflect.TypeOf(float64(0)) {
			obj.value = obj.value.(float64) + increment
			r.dict[key] = obj
			return true
		} else {
			fmt.Println("Error: value is not a number")
			return false
		}
	}
	return false
}
