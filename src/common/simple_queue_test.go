package common

import (
	"testing"
	"fmt"
)

func TestGetMostBase(t *testing.T) {
	var datas = []int64{1,110,120,130,130,140,201,232,242,200,220,230,240,330,340,1000,10000,120000,300000}
	base := getMostBase(datas)
	var sum, count int64
	for _, data := range datas {
		if data / base > 10 || data / base == 0 {
			continue
		}
		sum += (data / base) * base
		count++
    }
	fmt.Println(sum / count)
}
