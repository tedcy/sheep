package watcher_notify

import (
	"coding.net/tedcy/sheep/src/watcher"
	"time"
)

type WatcherNotifyI interface{
	NotifyWatcherChange() <-chan []string
}

func New(target, path string, timeout time.Duration) (WatcherNotifyI, error) {
	w := &watcherNotify{}
	var err error
	config := &watcher.Config{}
	config.Target = target
	config.Timeout = timeout
	w.path = path
	w.watcher, err = watcher.New(config)
	if err != nil {
		return nil, err
	}
	w.nodes = make(chan []string)
	return w, nil
}

type watcherNotify struct {
	watcher			watcher.WatcherI
	nodes			chan []string
	path			string
}

func (this *watcherNotify) NotifyWatcherChange() <-chan []string {
	go func() {
		for ;; {
			err := this.watcher.Watch(this.path, this.pushChan)
			if err != nil {
				println(err.Error())		
			}
		}
	}()
	return this.nodes
}

func (this *watcherNotify) pushChan() (uint64, error) {
	nodes, afterIndex, err := this.watcher.List(this.path)
	if err != nil {
		return 0, err
	}
	this.nodes <- nodes
	return afterIndex, nil
}
