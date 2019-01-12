package common

import (
	"errors"
	"golang.org/x/net/context"
	"time"
)

var ErrNoAvailableClients = errors.New("no available clients")

type KV struct {
	Key		string
	Weight	int
}

func Assert(err error) {
	if err != nil {
		panic(err)
	}
}

func Hung() {
	for {
		time.Sleep(time.Hour)
	}
}

func WithError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, "key", err)
}

func GetError(ctx context.Context) error {
	err, ok := ctx.Value("key").(error)
	if !ok {
		return nil
	}
	return err
}
