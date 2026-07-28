package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const messageVersion = 1

type Message struct {
	Version int    `json:"version"`
	JobID   string `json:"job_id"`
}

type ImportQueue struct {
	client    *redis.Client
	queueName string
}

func NewImportQueue(redisURL, queueName string) (*ImportQueue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &ImportQueue{
		client:    client,
		queueName: queueName,
	}, nil
}

func (q *ImportQueue) Close() error {
	return q.client.Close()
}

func (q *ImportQueue) Enqueue(ctx context.Context, jobID string) error {
	payload, err := json.Marshal(Message{
		Version: messageVersion,
		JobID:   jobID,
	})
	if err != nil {
		return fmt.Errorf("marshal queue message: %w", err)
	}
	if err := q.client.LPush(ctx, q.queueName, payload).Err(); err != nil {
		return fmt.Errorf("enqueue import job: %w", err)
	}
	return nil
}

// Dequeue blocks until a job is available or the context is cancelled.
func (q *ImportQueue) Dequeue(ctx context.Context) (string, error) {
	result, err := q.client.BRPop(ctx, 0, q.queueName).Result()
	if err != nil {
		return "", err
	}
	if len(result) != 2 {
		return "", fmt.Errorf("unexpected redis brpop payload")
	}

	var message Message
	if err := json.Unmarshal([]byte(result[1]), &message); err != nil {
		return "", fmt.Errorf("malformed queue message: %w", err)
	}
	if message.Version != messageVersion || message.JobID == "" {
		return "", fmt.Errorf("unsupported queue message")
	}
	return message.JobID, nil
}
