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
	b := &balancer{}
	b.data = make(map[string]*value)
	return b
}

type balancer struct {
	data		map[string]*value
}

type value struct {
	key			string
	weight		int
	enable		bool
}

//update的任何数据会生成weight池用于get
//生成weight池需要加写锁，get为读锁
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
