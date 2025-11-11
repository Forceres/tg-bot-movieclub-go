package tasks

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
)

const FinishSessionTaskType = "finish_session"

type FinishSessionTaskProcessor struct {
	b *bot.Bot
}

type FinishSessionTaskPayload struct {
	SessionID int
}

func NewFinishSessionTask(sessionID int) (*asynq.Task, error) {
	payload, err := json.Marshal(FinishSessionTaskPayload{SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(FinishSessionTaskType, payload), nil
}

type IFinishSessionTaskProcessor interface {
	Process() error
}

func NewFinishSessionTaskProcessor(b *bot.Bot) *FinishSessionTaskProcessor {
	return &FinishSessionTaskProcessor{
		b: b,
	}
}

type EnqueueFinishSessionParams struct {
	ChatID   int64
	Duration int
}

func EnqueueFinishSessionTask(client *asynq.Client, duration time.Duration, params *FinishSessionTaskPayload) error {
	task, err := NewFinishSessionTask(params.SessionID)
	if err != nil {
		log.Printf("Error creating finish session task: %v", err)
		return err
	}
	scheduleOpts := []asynq.Option{asynq.MaxRetry(1), asynq.ProcessIn(duration)}
	taskInfo, err := client.Enqueue(task, scheduleOpts...)
	if err != nil {
		log.Printf("Error scheduling session finish task: %v", err)
		return err
	}
	log.Printf("Scheduled session finish task: %s", taskInfo.ID)
	return nil
}

func (t *FinishSessionTaskProcessor) Process(ctx context.Context, task *asynq.Task) error {
	return nil
}
