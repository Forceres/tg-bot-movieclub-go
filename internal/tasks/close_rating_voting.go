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

const CloseRatingVotingTaskType = "close_rating_voting"

type CloseRatingVotingTaskProcessor struct {
	b             *bot.Bot
	votingService service.IVotingService
	voteService   service.IVoteService
	movieService  service.IMovieService
}

type CloseRatingVotingPayload struct {
	PollID    string `json:"poll_id"`
	MessageID int    `json:"message_id"`
	ChatID    int64  `json:"chat_id"`
	VotingID  int64  `json:"voting_id"`
	MovieID   int    `json:"movie_id"`
	UserID    int64  `json:"user_id"`
}

func NewCloseRatingVotingTask(pollID string, messageID int, chatID int64, votingID int64, movieID int) (*asynq.Task, error) {
	payload, err := json.Marshal(CloseRatingVotingPayload{PollID: pollID, MessageID: messageID, ChatID: chatID, VotingID: votingID, MovieID: movieID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(CloseRatingVotingTaskType, payload), nil
}

type ICloseRatingVotingTaskProcessor interface {
	Process() error
}

func NewCloseRatingVotingTaskProcessor(b *bot.Bot, votingService service.IVotingService, voteService service.IVoteService, movieService service.IMovieService) *CloseRatingVotingTaskProcessor {
	return &CloseRatingVotingTaskProcessor{
		b:             b,
		votingService: votingService,
		voteService:   voteService,
		movieService:  movieService,
	}
}

func (t *CloseRatingVotingTaskProcessor) Process(ctx context.Context, task *asynq.Task) error {
	var p CloseRatingVotingPayload
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
	mean, err := t.voteService.CalculateRatingMean(p.VotingID)
	if err != nil {
		return err
	}
	err = t.votingService.FinishRatingVoting(&service.FinishRatingVotingParams{
		VotingID:  p.VotingID,
		PollID:    p.PollID,
		MovieID:   p.MovieID,
		Mean:      mean,
		CreatedBy: p.UserID,
	})
	if err != nil {
		return err
	}
	movie, err := t.movieService.GetMovieByID(p.MovieID)
	if err != nil {
		return err
	}
	_, err = t.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: p.ChatID,
		Text: "Голосование завершено!\n" +
			"Фильм для просмотра: 🎬\n" +
			"<b>" + movie.Title + "</b>\n" +
			"Средний рейтинг: ⭐️ " + strconv.FormatFloat(mean, 'f', 2, 64),
		ParseMode: "HTML",
	})
	if err != nil {
		return err
	}
	return nil
}
