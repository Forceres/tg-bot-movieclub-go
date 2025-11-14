package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/hibiken/asynq"
)

const FinishSessionTaskType = "finish_session"

type FinishSessionTaskPayload struct {
	SessionID int64
}

func NewFinishSessionTask(sessionID int64) (*asynq.Task, error) {
	payload, err := json.Marshal(FinishSessionTaskPayload{SessionID: sessionID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(FinishSessionTaskType, payload), nil
}

type EnqueueFinishSessionParams struct {
	SessionID int64
	Duration  time.Duration
}

func EnqueueFinishSessionTask(client *asynq.Client, params *EnqueueFinishSessionParams) error {
	task, err := NewFinishSessionTask(params.SessionID)
	if err != nil {
		log.Printf("Error creating finish session task: %v", err)
		return err
	}
	scheduleOpts := []asynq.Option{asynq.MaxRetry(1), asynq.ProcessIn(params.Duration), asynq.TaskID(fmt.Sprintf("%s-%d", FinishSessionTaskType, params.SessionID)), asynq.Queue(QUEUE)}
	taskInfo, err := client.Enqueue(task, scheduleOpts...)
	if err != nil {
		log.Printf("Error scheduling session finish task: %v", err)
		return err
	}
	log.Printf("Scheduled session finish task: %s", taskInfo.ID)
	return nil
}

type FinishSessionTaskProcessor struct {
	sessionService service.ISessionService
}

type IFinishSessionTaskProcessor interface {
	Process() error
}

func NewFinishSessionTaskProcessor(sessionService service.ISessionService) *FinishSessionTaskProcessor {
	return &FinishSessionTaskProcessor{
		sessionService: sessionService,
	}
}

func (t *FinishSessionTaskProcessor) Process(ctx context.Context, task *asynq.Task) error {
	var payload FinishSessionTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("Error unmarshaling finish session task payload: %v", err)
		return err
	}
	err := t.sessionService.FinishSession(payload.SessionID)
	if err != nil {
		log.Printf("Error finishing session %d: %v", payload.SessionID, err)
		return err
	}
	return nil
}
