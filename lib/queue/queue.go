package queue

import (
	"bytes"
	"encoding/gob"
	"sync"
	"time"

	"github.com/google/uuid"

	rate "github.com/beefsack/go-rate"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Config defines the configurations of a message queue
type Config struct {
	Topic        string `toml:"topic"`
	Broker       string `toml:"broker"`
	GroupID      string `toml:"group_id"`
	RateLimit    int    `toml:"rate_limit"`
	RateInterval int64  `toml:"rate_interval"`
}

type handlerEntry struct {
	name    string
	handler func(*Job)
}

// Queue is a Kafka-based message queue
type Queue struct {
	topic         string
	producer      *kafka.Producer
	consumer      *kafka.Consumer
	handlers      []handlerEntry
	mu            sync.Mutex
	consuming     bool
	closeNotifyCh chan bool
	limiter       *rate.RateLimiter
}

// NewQueue creates a new Kafka-based message queue
func NewQueue(config *Config) (*Queue, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": config.Broker})
	if err != nil {
		return nil, err
	}
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.Broker,
		"group.id":          config.GroupID,
	})
	if err != nil {
		return nil, err
	}
	var limiter *rate.RateLimiter
	if config.RateLimit > 0 {
		if config.RateInterval == 0 {
			config.RateInterval = 1000
		}
		limiter = rate.New(config.RateLimit, time.Duration(config.RateInterval)*time.Millisecond)
	}
	return &Queue{
		producer: p,
		consumer: c,
		topic:    config.Topic,
		limiter:  limiter,
	}, nil
}

// Add adds a job to the queue
func (queue *Queue) Add(name string, data interface{}, options *JobOptions) error {
	var value bytes.Buffer
	e := gob.NewEncoder(&value)
	if err := e.Encode(data); err != nil {
		return err
	}
	headers := make([]kafka.Header, 0, 1)
	if options.JobID == "" {
		options.JobID = uuid.New().String()
	}
	headers = append(headers, kafka.Header{Key: "job_id", Value: []byte(options.JobID)})
	headers = append(headers, kafka.Header{Key: "name", Value: []byte(name)})
	return queue.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &queue.topic, Partition: kafka.PartitionAny},
		Headers:        headers,
		Value:          value.Bytes(),
	}, nil)
}

// Process adds a job handler function to the message queue for a certain job name
func (queue *Queue) Process(name string, handler func(*Job)) {
	queue.mu.Lock()
	queue.handlers = append(queue.handlers, handlerEntry{name: name, handler: handler})
	queue.mu.Unlock()
}

// Start starts processing jobs
func (queue *Queue) Start() error {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.consuming {
		return nil
	}
	err := queue.consumer.Subscribe(queue.topic, nil)
	if err != nil {
		return err
	}
	go queue.startConsume()
	return nil
}

func (queue *Queue) startConsume() {
	queue.mu.Lock()
	queue.consuming = true
	queue.mu.Unlock()
	queue.closeNotifyCh = make(chan bool, 1)
	for queue.consuming {
		select {
		case <-queue.closeNotifyCh:
			queue.mu.Lock()
			queue.consuming = false
			queue.mu.Unlock()
		default:
			ev := queue.consumer.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				if queue.limiter != nil {
					queue.limiter.Wait()
				}
				queue.mu.Lock()
				queue.handleMessage(e)
				queue.mu.Unlock()
			case *kafka.Error:
				if e.Code() == kafka.ErrAllBrokersDown {
					queue.mu.Lock()
					queue.consuming = false
					queue.mu.Unlock()
				}
			}
		}
	}
}

func (queue *Queue) handleMessage(km *kafka.Message) {
	var name string
	jobOptions := new(JobOptions)
	for _, header := range km.Headers {
		if header.Key == "name" {
			name = string(header.Value)
		}
		if header.Key == "job_id" {
			jobOptions.JobID = string(header.Value)
		}
	}

	for _, he := range queue.handlers {
		if he.name == name || he.name == "" {
			he.handler(&Job{
				Name:    name,
				Data:    km.Value,
				Options: jobOptions,
			})
		}
	}
}

// Stop stops processing messages
func (queue *Queue) Stop() {
	queue.closeNotifyCh <- true
}
