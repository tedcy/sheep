package common

import (
	"sync"
	//"sort"
	//"fmt"
	"time"
)

const max_count = 1200

type node struct {
	t		time.Time
	data	int64
}

type SimpleQueue struct {
	historyDatas		[]*node
	rwlock				sync.RWMutex
}

func NewSimpleQueue() *SimpleQueue {
	return &SimpleQueue{}
}

func (this *SimpleQueue) Insert(data int64) {
	//log.Infof("SimpleQueue.insert data=%.2f", data)
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	if len (this.historyDatas) >= max_count {
		this.historyDatas = this.historyDatas[1:]
    }
	n := &node{}
	n.data = data
	n.t = time.Now()
	this.historyDatas = append(this.historyDatas, n)
}

func (this *SimpleQueue) getDatas(t time.Time) []int64 {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	var tempDatas []int64
	var ok bool
	for _, n := range this.historyDatas {
		if !ok {
			if n.t.After(t) {
				ok = true
			}
		}else {
			tempDatas = append(tempDatas, n.data)
		}
	}
	return tempDatas
}

func (this *SimpleQueue) GetAverage(t time.Time) int64{
	datas := this.getDatas(t)
	if len(datas) == 0 {
		return 0
	}
	var sum int64
	for _, data := range datas {
		sum += data
		//log.Infof("GetAverage count=%d, data=%.2f, sum=%.2f", count, data, sum)
    }
	return sum / int64(len(datas))
}

func (this *SimpleQueue) GetMax(t time.Time) int64{
	datas := this.getDatas(t)
	if len(datas) == 0 {
		return 0
    }
	var max int64
	for _, data := range datas {
		if max < data {
			max = data
        }
    }
	return max
}

func (this *SimpleQueue) GetMost(t time.Time) int64 {
	datas := this.getDatas(t)
	if len(datas) < 10 {
		return 0
    }
	//sort.Sort(int64S(datas))
	base := getMostBase(datas)
	if base == 0 {
		return 0
	}
	var sum, c int64
	for _, data := range datas {
		if data / base > 10 || data / base == 0 {
			continue
		}
		sum += (data / base) * base
		c++
    }
	if c == 0 {
		return 0
	}
	//取平均数或者众数
	return sum / c
}

func getMostBase(datas []int64) int64{
	var m map[int64]int64 = make(map[int64]int64)
	for _, data := range datas {
		m[getBase(data)]++
	}
	var max, baseMax int64
	for base, count := range m {
		if max < count {
			max = count
			baseMax = base
        }
	}
	return baseMax
}

func getBase(data int64) int64 {
	var base int64 = 1
	/*defer func(data int64,base *int64) {
		fmt.Printf("%d --- %d\n", data, *base)	
	}(data, &base)*/
	for ;data > 10; {
		data /= 10
		base *= 10
	}
	return base
}

type int64S []int64

func (s int64S) Len() int {return len(s)}
func (s int64S) Less(i, j int) bool {return s[i] < s[j]}
func (s int64S) Swap(i, j int) {
	temp := s[i]
	s[i] = s[j]
	s[j] = temp
}
