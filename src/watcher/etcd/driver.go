package etcd

import (
	"time"
	"golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"strings"
)

type EtcdClient struct {
	kapi			client.KeysAPI
	timeout			time.Duration
}

func New(addrStr string, timeout time.Duration) (c *EtcdClient, err error) {
	c = &EtcdClient{}
	
	var addrList []string
	addrListTemp := strings.Split(addrStr, ",")
	for _, addr := range addrListTemp {
		addr = "http://" + addr
		addrList = append(addrList, addr)	
	}
	config := client.Config{}
	config.Endpoints = addrList
	config.Transport = client.DefaultTransport
	config.HeaderTimeoutPerRequest = timeout

	eC, err := client.New(config)	
	if err != nil {
		return
	}
	c.kapi = client.NewKeysAPI(eC)
	c.timeout = timeout

	return
}

func (this *EtcdClient) Create(path string, data []byte) (err error) {
	return
}
func (this *EtcdClient) Delete(path string) (err error) {
	return
}
func (this *EtcdClient) Read(path string) (data []byte,err error) {
	return
}
func (this *EtcdClient) List(path string) (paths []string, index uint64, err error) {
	ctx := context.Background()
	if this.timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, this.timeout)
	}
	resp, err := this.kapi.Get(ctx, path, nil)
	if err != nil {
		return
	}
	for _, node := range resp.Node.Nodes {
		paths = append(paths, strings.TrimPrefix(node.Key, path + "/"))
	}
	index = resp.Index
	return
}
func (this *EtcdClient) Update(path string, data []byte) (err error) {
	return
}
func (this *EtcdClient) Watch(path string, cb func() (uint64, error)) (err error) {
	var afterIndex uint64
	afterIndex, err = cb()
	if err != nil {
		return
	}
	w := this.kapi.Watcher(path, &client.WatcherOptions{AfterIndex: afterIndex})
	for ;; {
		var resp *client.Response
		resp, err = w.Next(context.Background())
		if err != nil {
			return
		}
		if resp.Action == "expire" || 
			resp.Action == "delete" ||
			resp.Action == "create" {
			afterIndex, err = cb()
			if err != nil {
				println(err)
			}
		}
	}
	return
}
func (this *EtcdClient) CreateEphemeral(path string, data []byte) (err error) {
	return
}
func (this *EtcdClient) CreateEphemeralInOrder(path string, data []byte) (err error) {
	return
}
