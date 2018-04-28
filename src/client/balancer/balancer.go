package balancer

import (
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"coding.net/tedcy/sheep/src/client/balancer/weighter_balancer"
	"coding.net/tedcy/sheep/src/client/balancer/watcher_notify"
	"coding.net/tedcy/sheep/src/watcher"
)

/*
type Balancer interface {
	Start(target string, config BalancerConfig) error
	Up(addr Address) (down func(error))
	Get(ctx context.Context, opts BalancerGetOptions) (addr Address, put func(), err error)
	Notify() <-chan []Address
	Close() error
}
*/

type Balancer struct {
	weighterBalancer	weighter_balancer.WeightBalancerI
	watcherNotify		watcher_notify.WatcherNotifyI
	addressChan			chan []grpc.Address
}

func New(config *watcher.Config) (balancer *Balancer, err error){
	balancer = &Balancer{}	
	balancer.weighterBalancer = weighter_balancer.New()
	balancer.watcherNotify, err = watcher_notify.New(config)
	balancer.addressChan = make(chan []grpc.Address)
	if err != nil {
		return
	}
	return
}

func (this *Balancer) Start(target string, config grpc.BalancerConfig) (err error){
	this.SetNotifyWatcher(this.watcherNotify.NotifyWatcherChange(target))
	return
}

func (this *Balancer) Up(addr grpc.Address) (down func(error)) {
	return nil
}
func (this *Balancer) Get(ctx context.Context, opts grpc.BalancerGetOptions) (addr grpc.Address, put func(), err error) {
	addr.Addr = this.weighterBalancer.Get()
	return
}
func (this *Balancer) Notify() <-chan []grpc.Address {
	return this.addressChan
}
func (this *Balancer) Close() error {
	return nil
}
