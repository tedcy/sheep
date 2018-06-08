package etcd

import (
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"strings"
	"time"
	"sync"
	"errors"
	"net"
)

var (
	ErrClosedClient = errors.New("use of closed etcd client")
	ErrNotDir  = errors.New("etcd: not a dir")
)

type EtcdClient struct {
	kapi			client.KeysAPI
	timeout			time.Duration
	refreshTimeout	time.Duration
	ctx				context.Context
	cancel			context.CancelFunc
	closed			bool
	rwlock			sync.RWMutex
	//just for get localip
	addrList		[]string
}

func New(ctx context.Context, addrStr string, timeout time.Duration) (c *EtcdClient, err error) {
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
	if c.timeout == 0 {
		c.refreshTimeout = time.Second * 12
	}else {
		c.refreshTimeout = c.timeout
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.addrList = addrList
	return
}

func (this *EtcdClient) GetLocalIp() string {
	addr := strings.TrimLeft(this.addrList[0], "http://")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return ""
	}
	defer conn.Close()
	return strings.SplitN(conn.LocalAddr().String(),":",2)[0]
}

func (this *EtcdClient) Create(path string, data []byte) (err error) {
	return
}
func (this *EtcdClient) Delete(path string) (err error) {
	return
}
func (this *EtcdClient) Read(path string) (data []byte, err error) {
	return
}
func (this *EtcdClient) List(path string) (paths []string, index uint64, err error) {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	if this.closed {
		err = ErrClosedClient
		return
	}
	ctx := this.ctx
	if this.timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, this.timeout)
	}
	resp, err := this.kapi.Get(ctx, path, nil)
	if err != nil {
		return
	}
	if !resp.Node.Dir {
		err = ErrNotDir
		return
	}
	for _, node := range resp.Node.Nodes {
		paths = append(paths, strings.TrimPrefix(node.Key, path+"/"))
	}
	index = resp.Index
	return
}
func (this *EtcdClient) Update(path string, data []byte) (err error) {
	return
}
func (this *EtcdClient) Watch(path string, cb func() (uint64, error)) (err error) {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	if this.closed {
		return ErrClosedClient
	}
	var afterIndex uint64
	afterIndex, err = cb()
	if err != nil {
		return
	}
	w := this.kapi.Watcher(path, &client.WatcherOptions{AfterIndex: afterIndex})
	for {
		var resp *client.Response
		resp, err = w.Next(this.ctx)
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
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	if this.closed {
		err = ErrClosedClient
		return
	}
	ctx := this.ctx
	if this.timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, this.timeout)
	}
	_, err = this.kapi.Set(ctx, path, string(data),
		&client.SetOptions{
			PrevExist: client.PrevIgnore,
			TTL:       this.refreshTimeout})
	if err != nil {
		return
	}
	this.runRefresh(path)
	return
}
func (this *EtcdClient) CreateEphemeralInOrder(path string, data []byte) (err error) {
	return
}

func (this *EtcdClient) runRefresh(path string) {
	go func() {
		for {
			if err := this.refresh(path); err != nil {
				return
			}
			select {
			case <-time.After(this.refreshTimeout / 2):
			case <-this.ctx.Done():
				return
			}
		}
	}()
}

func (this *EtcdClient) refresh(path string) (err error) {
	this.rwlock.RLock()
	defer this.rwlock.RUnlock()
	if this.closed {
		return ErrClosedClient
	}
	ctx := this.ctx
	if this.timeout != 0 {
		ctx, _ = context.WithTimeout(ctx, this.timeout)
	}
	_, err = this.kapi.Set(ctx, path, "",
		&client.SetOptions{
			PrevExist: client.PrevExist,
			Refresh:   true,
			TTL:       this.refreshTimeout})
	if err != nil {
		return
	}
	return
}

//TODO need test
func (this *EtcdClient) Close() (err error) {
	this.rwlock.Lock()
	defer this.rwlock.Unlock()
	if this.closed {
		return nil
	}
	this.closed = true
	this.cancel()
	return nil
}
