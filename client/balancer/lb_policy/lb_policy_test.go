package lb_policy

import (
	"testing"
	"fmt"
	"github.com/tedcy/sheep/common"	
)

func Test_All(t *testing.T) {
	wb := New()
	keys := []string{"A","B","C"}
	kvs := []*common.KV{
		&common.KV{
			Key: "A",
			Weight: 1,
		},
		&common.KV{
			Key: "B",
			Weight: 4,
		},
		&common.KV{
			Key: "C",
			Weight: 10,
		},
	}
	wb.UpdateAllWithoutWeight(keys)
	fmt.Println(wb.Get())
	wb.UpdateAll(kvs)
	fmt.Println(wb.Get())
	wb.Disable("C")
	fmt.Println(wb.Get())
	wb.Enable("C")
	fmt.Println(wb.Get())
}
