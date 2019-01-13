package resolver_notify

import (
	"github.com/tedcy/sheep/watcher"
	"golang.org/x/net/context"
	"time"
)

type ResolverNotifyI interface {
	NotifyResolverChange() <-chan []string
	Close() error
}

func New(ctx context.Context, target, path string, timeout time.Duration) (ResolverNotifyI, error) {
	w := &resolverNotify{}
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

type resolverNotify struct {
	watcher watcher.WatcherI
	nodes   chan []string
	path    string
	ctx     context.Context
	cancel  context.CancelFunc
}

//TODO add log interface to print err msg
func (this *resolverNotify) NotifyResolverChange() <-chan []string {
	go func() {
		for {
			select {
			case <-this.ctx.Done():
				return
			default:
			}
			err := this.watcher.Watch(this.path, this.pushChan)
			if err != nil {
				//println(err.Error())
			}
			select {
			case <-time.After(time.Second):
			case <-this.ctx.Done():
				return
			}
		}
	}()
	return this.nodes
}

func (this *resolverNotify) pushChan() (uint64, error) {
	nodes, afterIndex, err := this.watcher.List(this.path)
	if err != nil {
		return 0, err
	}
	this.nodes <- nodes
	return afterIndex, nil
}

//TODO add log interface to print err msg
//TODO return err
func (this *resolverNotify) Close() error {
	this.cancel()
	close(this.nodes)
	this.nodes = nil
	err := this.watcher.Close()
	if err != nil {
		//println(err)
	}
	return nil
}
