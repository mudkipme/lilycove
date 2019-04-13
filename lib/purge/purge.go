package purge

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mudkipme/lilycove/lib/queue"
)

// EntryConfig defines a purge configuration entry
type EntryConfig struct {
	Host     string   `toml:"host"`
	Method   string   `toml:"method"`
	Variants []string `toml:"variants"`
	URIs     []string `toml:"uris"`
}

// Config defines the configuration of a purger
type Config struct {
	Expiry  int64          `toml:"expiry"`
	Entries []*EntryConfig `toml:"entries"`
}

// Purger manages a cache purging queue
type Purger struct {
	config *Config
	queue  *queue.Queue
	client *http.Client
}

type purgeItem struct {
	Host          string
	URL           string
	ScheduledTime time.Time
}

// NewPurger creates a new purger
func NewPurger(c *Config, q *queue.Queue) (*Purger, error) {
	purger := &Purger{
		config: c,
		queue:  q,
		client: &http.Client{
			Timeout: time.Duration(1) * time.Second,
		},
	}
	q.Process("purge", func(job *queue.Job) {
		var item purgeItem
		if err := job.Decode(&item); err != nil {
			fmt.Printf("[Purger] Error decoding: %v\n", err)
			return
		}
		if time.Now().After(item.ScheduledTime.Add(time.Duration(purger.config.Expiry) * time.Millisecond)) {
			fmt.Printf("[Purger] Skip purging %v\n", item.Host+item.URL)
			return
		}
		fmt.Printf("[Purger] Started purging %v\n", item.Host+item.URL)
		purger.handlePurge(item)
		fmt.Printf("[Purger] Finished purging %v\n", item.Host+item.URL)
	})
	if err := q.Start(); err != nil {
		return nil, err
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		q.Stop()
		fmt.Println("[Purger] Purger stopped.\n")
		os.Exit(0)
	}()
	return purger, nil
}

// Add a purge job
func (p *Purger) Add(host string, url string) {
	jobID := host + url
	err := p.queue.Add("purge", purgeItem{
		Host:          host,
		URL:           url,
		ScheduledTime: time.Now(),
	}, &queue.JobOptions{
		JobID: jobID,
	})
	if err != nil {
		fmt.Printf("[Purger] Error adding purge job: %v\n", err)
	}
}
