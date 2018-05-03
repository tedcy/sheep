package common

import (
	"sync"
)

const max_count = 120

type SimpleQueue struct {
	historyDatas		[]int64
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
	this.historyDatas = append(this.historyDatas, data)
}

func (this *SimpleQueue) getDatas(count uint32) []int64 {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	var tempDatas []int64
	length := uint32(len(this.historyDatas))
	if length >= count {
		tempDatas = append(tempDatas, this.historyDatas[length-count:]...)
    }else {
		tempDatas = append(tempDatas, this.historyDatas[:]...)
    }
	return tempDatas
}

func (this *SimpleQueue) GetAverage(count uint32) int64{
	datas := this.getDatas(count)
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

func (this *SimpleQueue) GetMax(count uint32) int64{
	datas := this.getDatas(count)
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
