package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
)

const CloseSelectionVotingTaskType = "close_selection_voting"

type CloseSelectionVotingTaskProcessor struct {
	b             *bot.Bot
	movieService  service.IMovieService
	votingService service.IVotingService
	voteService   service.IVoteService
	client        *asynq.Client
	inspector     *asynq.Inspector
}

type CloseSelectionVotingPayload struct {
	PollID    string `json:"poll_id"`
	MessageID int    `json:"message_id"`
	ChatID    int64  `json:"chat_id"`
	VotingID  int64  `json:"voting_id"`
	UserID    int64  `json:"user_id"`
}

type ICloseSelectionVotingProcessor interface {
	Process(ctx context.Context, task *asynq.Task) error
}

func NewCloseSelectionVotingTaskProcessor(b *bot.Bot, votingService service.IVotingService, voteService service.IVoteService, movieService service.IMovieService, inspector *asynq.Inspector, client *asynq.Client) *CloseSelectionVotingTaskProcessor {
	return &CloseSelectionVotingTaskProcessor{
		b:             b,
		votingService: votingService,
		voteService:   voteService,
		movieService:  movieService,
		inspector:     inspector,
		client:        client,
	}
}

func NewCloseSelectionVotingTask(pollID string, messageID int, chatID int64, votingID int64) (*asynq.Task, error) {
	payload, err := json.Marshal(CloseSelectionVotingPayload{PollID: pollID, MessageID: messageID, ChatID: chatID, VotingID: votingID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(CloseSelectionVotingTaskType, payload), nil
}

func EnqueueCloseSelectionVotingTask(client *asynq.Client, duration time.Duration, params *CloseSelectionVotingPayload) error {
	task, err := NewCloseSelectionVotingTask(params.PollID, params.MessageID, params.ChatID, params.VotingID)
	if err != nil {
		log.Printf("Error creating close selection voting task: %v", err)
		return err
	}
	scheduleOpts := []asynq.Option{asynq.MaxRetry(1), asynq.ProcessIn(duration), asynq.TaskID(fmt.Sprintf("%s-%d", CloseSelectionVotingTaskType, params.VotingID)), asynq.Queue(QUEUE)}
	taskInfo, err := client.Enqueue(task, scheduleOpts...)
	if err != nil {
		log.Printf("Error scheduling voting end task: %v", err)
		return err
	}
	log.Printf("Scheduled voting end task: %s", taskInfo.ID)
	return nil
}

func (t *CloseSelectionVotingTaskProcessor) Process(ctx context.Context, task *asynq.Task) error {
	var p CloseSelectionVotingPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return err
	}
	ok, err := t.b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    p.ChatID,
		MessageID: p.MessageID,
	})
	if err != nil || !ok {
		log.Println("Message doesn't exist or couldn't be deleted")
	}
	count, movieID, err := t.voteService.CalculateMaxMovieCount(p.VotingID)
	if err != nil {
		log.Printf("Error calculating max movie count: %v", err)
		return err
	}
	if count == 0 || movieID == 0 {
		log.Println("No votes were cast or no movie selected")
		return nil
	}
	log.Printf("Max movie count: %d for movie ID: %d", count, movieID)
	movie, err := t.movieService.GetMovieByID(movieID)
	if err != nil {
		log.Printf("Error getting movie by ID: %v", err)
		return err
	}
	session, created, err := t.votingService.FinishSelectionVoting(&service.FinishSelectionVotingParams{
		VotingID:  p.VotingID,
		PollID:    p.PollID,
		MovieID:   movie.ID,
		CreatedBy: p.UserID,
	})
	if err != nil {
		log.Printf("Error finishing selection voting: %v", err)
		return err
	}
	_, err = t.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: p.ChatID,
		Text:   "Финальное решение принято! Победил фильм: " + movie.Title + "; с количеством голосов: " + strconv.FormatInt(count, 10),
	})
	if err != nil {
		log.Printf("Error sending final decision message: %v", err)
		return err
	}
	duration := time.Until(time.Unix(session.FinishedAt, 0))
	if created {
		err = EnqueueFinishSessionTask(t.client, &EnqueueFinishSessionParams{
			SessionID: session.ID,
			Duration:  duration,
		})
		if err != nil {
			log.Printf("Error scheduling new finish session task: %v", err)
		} else {
			log.Printf("Scheduled new finish session task for session: %d", session.ID)
		}
	}
	taskId := fmt.Sprintf("%s-%d-%d", OpenRatingVotingTaskType, session.ID, movie.ID)
	taskInfo, err := t.inspector.GetTaskInfo(QUEUE, taskId)
	if err != nil {
		log.Printf("Error getting task info: %v", err)
	}
	if taskInfo == nil {
		err = EnqueueOpenRatingVotingTask(t.client, &EnqueueOpenRatingVotingParams{
			ChatID:    p.ChatID,
			SessionID: session.ID,
			Movie:     *movie,
			UserID:    p.UserID,
			TaskID:    taskId,
			Duration:  duration,
		})
		if err != nil {
			log.Printf("Error scheduling new open rating voting task: %v", err)
		} else {
			log.Printf("Scheduled new open rating voting task for session: %d", session.ID)
		}
	}
	return nil
}
