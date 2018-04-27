package balancer

import (
	"google.golang.org/grpc"
	"coding.net/tedcy/sheep/src/common"
)

//watcher
func (this *Balancer) SetNotifyWatcher(notify <-chan []string) {
	for nodes := range notify {
		this.weighterBalancer.UpdateAllWithoutWeight(nodes)
		//通知grpc
		var addrs []grpc.Address
		for _, key := range nodes {
			addr := &grpc.Address{}	
			addr.Addr = key
		}
		this.addressChan <- addrs
	}
}

//weighter
func (this *Balancer) SetNotifyWeighterChange(notify <-chan []*common.KV) {
	for nodes := range notify {
		this.weighterBalancer.UpdateAll(nodes)
	}
}

//breaker
func (this *Balancer) SetNotifyOpen(notify <-chan []string) {
	for nodes := range notify {
		for _, node := range nodes {
			this.weighterBalancer.Enable(node)
		}
	}
}

func (this *Balancer) SetNotifyClose(notify <-chan []string) {
	for nodes := range notify {
		for _, node := range nodes {
			this.weighterBalancer.Disable(node)
		}
	}
}

func (this *Balancer) SetNotifyHalfOpen(notify <-chan []string) {
	//todo add test list
}
