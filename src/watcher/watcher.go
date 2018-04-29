package watcher

import (
	"fmt"
	"time"
	"strings"
	"coding.net/tedcy/sheep/src/watcher/etcd"
	"coding.net/tedcy/sheep/src/watcher/test"
)

type WatcherI interface {
	Create(path string, data []byte) error
	Delete(path string) error

	Read(path string) ([]byte, error)
	//return keys, index, error
	List(path string) ([]string, uint64, error)
	
	Update(path string, data []byte) error

	//cb should return afterIndex
	Watch(path string, cb func() (uint64, error)) error

	CreateEphemeral(path string, data []byte) error
	CreateEphemeralInOrder(path string, data []byte) error
}

type Config struct {
	//etcd://ip:port,ip:port
	Target			string
	Timeout			time.Duration
}

func New(config *Config) (WatcherI, error){
	ss := strings.SplitN(config.Target, "://", 2)
	if len(ss) != 2 {
		return nil, fmt.Errorf("invalid watcher target %s", config.Target)
	}
	switch ss[0] {
	case "etcd":
		etcdClient,err := etcd.New(ss[1], config.Timeout)
		if err != nil {
			return nil, err
		}
		return etcdClient, nil
	case "test":
		return test.New(), nil
	}
	return nil, fmt.Errorf("invalid watcherName %s", ss[0])
}
