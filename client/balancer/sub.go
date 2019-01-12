package balancer

import (
	"github.com/tedcy/sheep/common"
	"google.golang.org/grpc"
)

//watcher
func (this *Balancer) SetNotifyWatcher(notify <-chan []string) {
	go func() {
		for nodes := range notify {
			this.weighterBalancer.UpdateAllWithoutWeight(nodes)
			//通知grpc
			var addrs []grpc.Address
			for _, key := range nodes {
				addr := grpc.Address{}
				addr.Addr = key
				addrs = append(addrs, addr)
			}
			this.addressChan <- addrs
		}
	}()
}

//weighter
func (this *Balancer) SetNotifyWeighterChange(notify <-chan []*common.KV) {
	go func() {
		for nodes := range notify {
			this.weighterBalancer.UpdateAll(nodes)
		}
	}()
}

//breaker
func (this *Balancer) SetNotifyOpen(notify <-chan string) {
	go func() {
		for node := range notify {
			this.weighterBalancer.Disable(node)
		}
	}()
}

func (this *Balancer) SetNotifyClose(notify <-chan string) {
	go func() {
		for node := range notify {
			this.weighterBalancer.Enable(node)
		}
	}()
}

func (this *Balancer) SetNotifyHalfOpen(notify <-chan string) {
	go func() {
		for node := range notify {
			this.weighterBalancer.Enable(node)
		}
	}()
}
