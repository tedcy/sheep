package weighter_balancer

//加权选择器

import (
	"coding.net/tedcy/sheep/src/common"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

//default weight = allweight / count
const (
	unsetWeight   = -1
	defaultWeight = 1
)

type WeightBalancerI interface {
	//choose
	Get() (key string, ok bool)

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
	b.data = make(map[string]*weightNode)
	b.rand = rand.New(rand.NewSource(time.Now().Unix()))
	return b
}

type balancer struct {
	rand          *rand.Rand
	data          map[string]*weightNode
	weightEndPool []*weightEndPoolNode
	weightSum     int
	rwlock        sync.RWMutex
}

type weightNode struct {
	key    string
	weight int
	enable bool
}

type weightEndPoolNode struct {
	weightEnd int
	node      *weightNode
}

func (this *balancer) updateWeightEndPool() {
	var weightEndPool []*weightEndPoolNode
	var weightSum int
	var weightSoFar int
	for _, node := range this.data {
		if node.enable {
			weightSum += node.weight
			weightSoFar += node.weight
			wepn := &weightEndPoolNode{}
			wepn.node = node
			wepn.weightEnd = weightSoFar
			weightEndPool = append(weightEndPool, wepn)
		}
	}
	this.weightEndPool = weightEndPool
	this.weightSum = weightSoFar
	fmt.Println("updateWeightEnd")
	for _, node := range this.weightEndPool {
		fmt.Println(node.node.key, "-", node.weightEnd)
	}
}

//update的任何数据会生成weight池用于get
//生成weight池需要加写锁，get为读锁
func (this *balancer) Get() (string, bool) {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	if this.weightSum == 0 {
		return "", false
	}
	r := int(this.rand.Uint32()) % this.weightSum
	for _, node := range this.weightEndPool {
		if node.weightEnd > r {
			return node.node.key, true
		}
	}
	panic("")
	return "", false
}

func (this *balancer) getEnableAvg(data map[string]*weightNode) int {
	//计算平均值
	var sum int
	var count int
	var avg int
	for _, wn := range data {
		if wn.enable && wn.weight != unsetWeight {
			sum += wn.weight
			count++
		}
	}
	if count != 0 {
		avg = sum / count
	} else {
		avg = defaultWeight
	}
	return avg
}

//watcher触发
//以新keys为基准，多出来的不要，少的补全
func (this *balancer) UpdateAllWithoutWeight(keys []string) {
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	var wn *weightNode
	var ok bool
	//创建新表复制老的weight
	data := make(map[string]*weightNode)
	for _, key := range keys {
		wn, ok = this.data[key]
		if !ok {
			wn = &weightNode{}
			wn.enable = true
			wn.key = key
			wn.weight = unsetWeight
		}
		data[key] = wn
	}
	avg := this.getEnableAvg(data)
	//新节点设置为平均值
	for _, wn = range data {
		if wn.weight == unsetWeight {
			wn.weight = avg
		}
	}
	this.data = data
	this.updateWeightEndPool()
}

//weighter触发，不新增删除节点
func (this *balancer) UpdateAll(kvs []*common.KV) {
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	var wn *weightNode
	var ok bool
	for _, kv := range kvs {
		wn, ok = this.data[kv.Key]
		if ok {
			wn.weight = kv.Weight
		}
	}
	this.updateWeightEndPool()
}

//TODO open和halfOpen触发，weight值设置多少还要考虑
func (this *balancer) Enable(key string) {
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	wn, ok := this.data[key]
	//如果没这个节点就不用开关了
	if ok {
		wn.enable = true
		wn.weight = unsetWeight
		wn.weight = this.getEnableAvg(this.data)
		this.updateWeightEndPool()
	}
}

func (this *balancer) Disable(key string) {
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	wn, ok := this.data[key]
	if ok {
		wn.enable = false
		this.updateWeightEndPool()
	}
}
