package balancer

import (
	"github.com/tedcy/sheep/common"
	"google.golang.org/grpc"
)

//resolverNotify
func (this *Balancer) SetNotifyResolver(notify <-chan []string) {
	go func() {
		for nodes := range notify {
			this.lbPolicy.UpdateAllWithoutWeight(nodes)
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

//lbPolicy
func (this *Balancer) SetNotifyLbPolicyChange(notify <-chan []*common.KV) {
	go func() {
		for nodes := range notify {
			this.lbPolicy.UpdateAll(nodes)
		}
	}()
}

//breaker
func (this *Balancer) SetNotifyOpen(notify <-chan string) {
	go func() {
		for node := range notify {
			this.lbPolicy.Disable(node)
		}
	}()
}

func (this *Balancer) SetNotifyClose(notify <-chan string) {
	go func() {
		for node := range notify {
			this.lbPolicy.Enable(node)
		}
	}()
}

func (this *Balancer) SetNotifyHalfOpen(notify <-chan string) {
	go func() {
		for node := range notify {
			this.lbPolicy.Enable(node)
		}
	}()
}
