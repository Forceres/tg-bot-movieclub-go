package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/hibiken/asynq"
)

const OpenRatingVotingTaskType = "open_rating_voting"

var RATING_VOTING_OPTIONS = []models.InputPollOption{
	{Text: "1"},
	{Text: "2"},
	{Text: "3"},
	{Text: "4"},
	{Text: "5"},
	{Text: "6"},
	{Text: "7"},
	{Text: "8"},
	{Text: "9"},
	{Text: "10"},
}

type OpenRatingVotingPayload struct {
	ChatID    int64       `json:"chat_id"`
	SessionID int64       `json:"session_id"`
	Movie     model.Movie `json:"movie"`
	UserID    int64       `json:"user_id"`
}

func NewOpenRatingVotingTask(chatID int64, sessionID int64, movie model.Movie, userID int64) (*asynq.Task, error) {
	payload, err := json.Marshal(OpenRatingVotingPayload{ChatID: chatID, SessionID: sessionID, Movie: movie, UserID: userID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(OpenRatingVotingTaskType, payload), nil
}

type EnqueueOpenRatingVotingParams struct {
	SessionID int64       `json:"session_id"`
	ChatID    int64       `json:"chat_id"`
	Movie     model.Movie `json:"movie"`
	UserID    int64       `json:"user_id"`
	TaskID    string      `json:"task_id"`
	Duration  time.Duration
}

func EnqueueOpenRatingVotingTask(client *asynq.Client, params *EnqueueOpenRatingVotingParams) error {
	task, err := NewOpenRatingVotingTask(params.ChatID, params.SessionID, params.Movie, params.UserID)
	if err != nil {
		log.Printf("Error creating finish session task: %v", err)
		return err
	}
	scheduleOpts := []asynq.Option{asynq.MaxRetry(1), asynq.ProcessIn(params.Duration), asynq.TaskID(params.TaskID), asynq.Queue(QUEUE)}
	taskInfo, err := client.Enqueue(task, scheduleOpts...)
	if err != nil {
		log.Printf("Error scheduling session finish task: %v", err)
		return err
	}
	log.Printf("Scheduled session finish task: %s", taskInfo.ID)
	return nil
}

type OpenRatingVotingTaskProcessor struct {
	b             *bot.Bot
	votingService service.IVotingService
	movieService  service.IMovieService
	asynqClient   *asynq.Client
}

type IOpenRatingVotingTaskProcessor interface {
	Process() error
}

func NewOpenRatingVotingTaskProcessor(b *bot.Bot, votingService service.IVotingService, movieService service.IMovieService, asynqClient *asynq.Client) *OpenRatingVotingTaskProcessor {
	return &OpenRatingVotingTaskProcessor{
		b:             b,
		asynqClient:   asynqClient,
		votingService: votingService,
		movieService:  movieService,
	}
}

func (t *OpenRatingVotingTaskProcessor) Process(ctx context.Context, task *asynq.Task) error {
	var p OpenRatingVotingPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return err
	}
	duration := time.Duration(15) * time.Minute
	finishedAt := time.Now().Add(duration).Unix()
	title := fmt.Sprintf("Оцените фильм: %s", p.Movie.Title)
	poll, err := t.votingService.StartVoting(&service.StartRatingVotingParams{
		Bot:     t.b,
		Context: ctx,
		ChatID:  p.ChatID,
		Options: service.VotingOptions{
			Title:      title,
			Type:       model.VOTING_RATING_TYPE,
			CreatedBy:  p.UserID,
			FinishedAt: &finishedAt,
			MovieID:    &p.Movie.ID,
			SessionID:  &p.SessionID,
		},
		PollOptions: RATING_VOTING_OPTIONS,
		Question:    title,
	})
	if err != nil {
		log.Printf("Error while starting a voting: %v", err)
		return err
	}
	err = EnqueueCloseRatingVotingTask(t.asynqClient, duration, &CloseRatingVotingPayload{
		PollID:    poll.PollID,
		MessageID: poll.MessageID,
		ChatID:    p.ChatID,
		VotingID:  poll.VotingID,
		MovieID:   p.Movie.ID,
	})
	if err != nil {
		log.Printf("Error scheduling close rating voting task: %v", err)
	}
	return nil
}
