package watcher

import (
	"github.com/tedcy/sheep/watcher/etcd"
	"github.com/tedcy/sheep/watcher/test"
	"fmt"
	"golang.org/x/net/context"
	"strings"
	"time"
)

type WatcherI interface {
	GetLocalIp() string
	Create(path string, data string) error
	Delete(path string) error

	Read(path string) (string, uint64, error)
	//return keys, index, error
	List(path string) ([]string, uint64, error)

	Update(path string, data []byte) error

	//cb should return afterIndex
	Watch(path string, cb func() (uint64, error)) error

	CreateEphemeral(path string, data []byte) error
	CreateEphemeralInOrder(path string, data []byte) error

	Close() error
}

type Config struct {
	//etcd://ip:port,ip:port
	Target  string
	Timeout time.Duration
}

func New(ctx context.Context, config *Config) (WatcherI, error) {
	ss := strings.SplitN(config.Target, "://", 2)
	if len(ss) != 2 {
		return nil, fmt.Errorf("invalid watcher target %s", config.Target)
	}
	switch ss[0] {
	case "etcd":
		etcdClient, err := etcd.New(ctx, ss[1], config.Timeout)
		if err != nil {
			return nil, err
		}
		return etcdClient, nil
	case "test":
		return test.New(), nil
	}
	return nil, fmt.Errorf("invalid watcherName %s", ss[0])
}
