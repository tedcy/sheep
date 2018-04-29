package test

var g_test test

func New() *test{
	return &g_test
}

func DefaultWatch(notify <-chan struct{}) {
	g_test.watch = func(path string, cb func() (uint64, error)) (err error) {
		for _ = range notify {
			cb()
		}
		return
	}
}

func DefaultList(notify <-chan []string) {
	g_test.list = func(path string) (paths []string, index uint64, err error) {
		paths = <-notify
		return
	}
}

func SetList(list func(path string) (paths []string, index uint64, err error)) {
	g_test.list = list
}

func SetWatch(watch func(path string, cb func() (uint64, error)) (err error)){
	g_test.watch = watch
}

type test struct {
	list func(path string) (paths []string, index uint64, err error)
	watch func(path string, cb func() (uint64, error)) (err error)
}

func (this *test) Create(path string, data []byte) (err error) {
	return
}
func (this *test) Delete(path string) (err error) {
	return
}
func (this *test) Read(path string) (data []byte,err error) {
	return
}
func (this *test) List(path string) (paths []string, index uint64, err error) {
	return this.list(path)
}
func (this *test) Update(path string, data []byte) (err error) {
	return
}
func (this *test) Watch(path string, cb func() (uint64, error)) (err error) {
	return this.watch(path, cb)
}
func (this *test) CreateEphemeral(path string, data []byte) (err error) {
	return
}
func (this *test) CreateEphemeralInOrder(path string, data []byte) (err error) {
	return
}
