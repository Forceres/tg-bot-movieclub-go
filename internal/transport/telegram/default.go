package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type DefaultHandler struct {
	f *fsm.FSM
}

func NewDefaultHandler(f *fsm.FSM) *DefaultHandler {
	return &DefaultHandler{f: f}
}

func (h *DefaultHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println("new update")
	if update.PollAnswer != nil {
		fmt.Println("Poll exists!")
	}

	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	currentState := h.f.Current(userID)

	switch currentState {
	case stateDefault:
		return

	case statePrepareVotingType:
		return
	case statePrepareVotingTitle:
		title := cases.Title(language.Russian).String(update.Message.Text)
		h.f.Set(userID, "title", title)

		h.f.Transition(userID, statePrepareMovies, userID, ctx, b, update)
	case statePrepareVotingDuration:
		duration, errDuration := strconv.Atoi(update.Message.Text)
		if errDuration != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Введите корректное целое число",
			})
			return
		}

		h.f.Set(userID, "duration", duration)

		h.f.Transition(userID, stateStartVoting, userID, ctx, b, update)
	case statePrepareMovies:
		ids := update.Message.Text

		movieIDs := []int64{}
		iter := strings.SplitSeq(ids, ",")
		for id := range iter {
			movieID, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
			if err == nil {
				movieIDs = append(movieIDs, movieID)
			}
		}

		if len(movieIDs) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Введите корректные целые числа через запятую",
			})
			return
		}

		h.f.Set(userID, "movieIDs", movieIDs)

		h.f.Transition(userID, statePrepareVotingDuration, userID, ctx, b, update)
	default:
		fmt.Printf("unexpected state %s\n", currentState)
	}
}
