package tasks

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

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
}

type CloseSelectionVotingPayload struct {
	PollID    string `json:"poll_id"`
	MessageID int    `json:"message_id"`
	ChatID    int64  `json:"chat_id"`
	VotingID  int64  `json:"voting_id"`
}

type ICloseSelectionVotingProcessor interface {
	Process(ctx context.Context, task *asynq.Task) error
}

func NewCloseSelectionVotingTaskProcessor(b *bot.Bot, votingService service.IVotingService, voteService service.IVoteService, movieService service.IMovieService) *CloseSelectionVotingTaskProcessor {
	return &CloseSelectionVotingTaskProcessor{
		b:             b,
		votingService: votingService,
		voteService:   voteService,
		movieService:  movieService,
	}
}

func NewCloseSelectionVotingTask(pollID string, messageID int, chatID int64, votingID int64) (*asynq.Task, error) {
	payload, err := json.Marshal(CloseSelectionVotingPayload{PollID: pollID, MessageID: messageID, ChatID: chatID, VotingID: votingID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(CloseSelectionVotingTaskType, payload), nil
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
		return err
	}
	log.Printf("Max movie count: %d for movie ID: %d", count, movieID)
	movie, err := t.movieService.GetMovieByID(movieID)
	if err != nil {
		return err
	}
	_, err = t.votingService.FinishSelectionVoting(&service.FinishSelectionVotingParams{
		VotingID:  p.VotingID,
		PollID:    p.PollID,
		MovieID:   movie.ID,
		CreatedBy: p.ChatID,
	})
	if err != nil {
		return err
	}
	_, err = t.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: p.ChatID,
		Text:   "Финальное решение принято! Победил фильм: " + movie.Title + "; с количеством голосов: " + strconv.FormatInt(count, 10),
	})
	if err != nil {
		return err
	}
	return nil
}
