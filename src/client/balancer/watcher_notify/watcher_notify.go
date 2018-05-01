package watcher_notify

import (
	"coding.net/tedcy/sheep/src/watcher"
	"golang.org/x/net/context"
	"time"
)

type WatcherNotifyI interface {
	NotifyWatcherChange() <-chan []string
	Close() error
}

func New(ctx context.Context, target, path string, timeout time.Duration) (WatcherNotifyI, error) {
	w := &watcherNotify{}
	var err error
	config := &watcher.Config{}
	config.Target = target
	config.Timeout = timeout
	w.path = path
	w.watcher, err = watcher.New(ctx, config)
	if err != nil {
		return nil, err
	}
	w.nodes = make(chan []string)
	w.ctx, w.cancel = context.WithCancel(ctx)
	return w, nil
}

type watcherNotify struct {
	watcher watcher.WatcherI
	nodes   chan []string
	path    string
	ctx     context.Context
	cancel  context.CancelFunc
}

func (this *watcherNotify) NotifyWatcherChange() <-chan []string {
	go func() {
		for {
			err := this.watcher.Watch(this.path, this.pushChan)
			if err != nil {
				println(err.Error())
			}
			select {
			case <-this.ctx.Done():
				return
			case <-time.NewTimer(time.Second).C:
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

func (this *watcherNotify) Close() error {
	this.cancel()
	close(this.nodes)
	this.nodes = nil
	err := this.watcher.Close()
	if err != nil {
		println(err)
	}
	return nil
}
