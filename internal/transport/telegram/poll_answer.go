package telegram

import (
	"context"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type PollAnswerHandler struct {
	pollService service.IPollService
	voteService service.IVoteService
}

type IPollAnswerHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewPollAnswerHandler(pollService service.IPollService, voteService service.IVoteService) *PollAnswerHandler {
	return &PollAnswerHandler{pollService: pollService, voteService: voteService}
}

func (h *PollAnswerHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	poll, err := h.pollService.GetPollByPollID(update.PollAnswer.PollID)
	if err != nil {
		log.Printf("Poll not found: %s", update.PollAnswer.PollID)
		return
	}
	if poll != nil && poll.Status == model.POLL_CLOSED_STATUS {
		log.Printf("Poll is closed: %s", update.PollAnswer.PollID)
		return
	}

	log.Printf("User %d voted in poll %s (type: %s)\n",
		update.PollAnswer.User.ID,
		update.PollAnswer.PollID,
		poll.Type)

	for _, optionID := range update.PollAnswer.OptionIDs {
		vote := &model.Vote{
			VotingID: poll.VotingID,
			UserID:   update.PollAnswer.User.ID,
		}

		if poll.Type == RATING_TYPE {
			rating := optionID + 1
			vote.Rating = &rating
			vote.MovieID = poll.MovieID
		} else if poll.Type == SELECTION_TYPE {
			options, err := h.pollService.GetPollOptionsByPollID(poll.ID)
			if err != nil {
				log.Printf("Error getting poll options: %v", err)
				continue
			}

			if optionID < len(options) {
				vote.MovieID = &options[optionID].MovieID
			} else {
				log.Printf("Invalid option ID: %d", optionID)
				continue
			}
		}

		if err := h.voteService.Create(vote); err != nil {
			log.Printf("Error saving vote: %v", err)
		}
	}
}
