package lib

import (
	"sync"

	"github.com/mudkipme/lilycove/lib/purge"
	"github.com/mudkipme/lilycove/lib/queue"
)

var defaultPurger *purge.Purger
var mup sync.Mutex

// DefaultPurger returns a default cache purger
func DefaultPurger() *purge.Purger {
	mup.Lock()
	defer mup.Unlock()

	if defaultPurger != nil {
		return defaultPurger
	}

	config := Config()
	queue, err := queue.NewQueue(config.Queue)
	if err != nil {
		panic(err)
	}
	purger, err := purge.NewPurger(config.Purge, queue)
	if err != nil {
		panic(err)
	}
	defaultPurger = purger
	return purger
}
