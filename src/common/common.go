package common

import (
	"errors"
)

var ErrNoAvailableClients = errors.New("no available clients")

type KV struct {
	Key		string
	Weight	int
}
