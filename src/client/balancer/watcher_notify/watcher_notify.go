package watcher_notify

import (
	"coding.net/tedcy/sheep/src/watcher"
)

type WatcherNotifyI interface{
	NotifyWatcherChange(path string) <-chan []string
}

func New(config *watcher.Config) (WatcherNotifyI, error) {
	w := &watcherNotify{}
	var err error
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

func (this *watcherNotify) NotifyWatcherChange(path string) <-chan []string {
	this.path = path
	for ;; {
		err := this.watcher.Watch(path, this.pushChan)
		if err != nil {
			println(err.Error())		
		}
	}
	return nil
}

func (this *watcherNotify) pushChan() (uint64, error) {
	nodes, afterIndex, err := this.watcher.List(this.path)
	if err != nil {
		return 0, err
	}
	this.nodes <- nodes
	return afterIndex, nil
}
