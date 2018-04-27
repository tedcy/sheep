package weighter_balancer
//加权选择器

import (
	"coding.net/tedcy/sheep/src/common"	
)

//default weight = allweight / count

type WeightBalancerI interface{
	//choose
	Get() (key string)
	
	//watcher
	UpdateAllWithoutWeight(keys []string)
	
	//weighter
	UpdateAll(kvs []*common.KV)
	
	//breaker
	//重新有效后变成初始值
	Enable(key string)
	//熔断后无效
	Disable(key string)
}

func New() WeightBalancerI {
	return &balancer{}
}

type balancer struct {

}

func (this *balancer) Get() string{
	return ""
}

func (this *balancer) UpdateAllWithoutWeight(keys []string) {
	return
}

func (this *balancer) UpdateAll(kvs []*common.KV) {

}

func (this *balancer) Enable(key string) {

}

func (this *balancer) Disable(key string) {

}
