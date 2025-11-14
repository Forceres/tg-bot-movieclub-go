package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
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
	case stateSaveSchedule:
		return
	case stateDate:
		return
	case statePrepareVotingTitle:
		fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
		title := cases.Title(language.Russian).String(update.Message.Text)
		h.f.Set(userID, "title", title)
		h.f.Transition(userID, statePrepareMovies, userID, ctx, b, update)
	case statePrepareVotingDuration:
		fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
		duration, errDuration := strconv.Atoi(update.Message.Text)
		if errDuration != nil || duration <= 0 {
			msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Введите корректное целое число > 0",
			})
			if err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
			fsmutils.AppendMessageID(h.f, userID, msg.ID)
			return
		}
		h.f.Set(userID, "duration", duration)
		h.f.Transition(userID, stateStartVoting, userID, ctx, b, update)
	case statePrepareMovies:
		indexes := update.Message.Text
		fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
		movieIndexes := []int64{}
		iter := strings.SplitSeq(indexes, ",")
		unique := make(map[int64]struct{})
		for idx := range iter {
			movieID, err := strconv.ParseInt(strings.TrimSpace(idx), 10, 64)
			if err == nil {
				if _, exists := unique[movieID]; exists {
					continue
				}
				unique[movieID] = struct{}{}
				movieIndexes = append(movieIndexes, movieID)
			}
		}
		if len(movieIndexes) == 0 {
			msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Введите корректные целые числа через запятую",
			})
			if err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
			fsmutils.AppendMessageID(h.f, userID, msg.ID)
			return
		}
		h.f.Set(userID, "movieIndexes", movieIndexes)
		h.f.Transition(userID, statePrepareVotingDuration, userID, ctx, b, update)
	case statePrepareCancelIDs:
		idxs := update.Message.Text
		fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
		cancelIndexes := []int64{}
		iter := strings.SplitSeq(idxs, ",")
		unique := make(map[int64]struct{})
		for id := range iter {
			cancelIdx, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
			if err == nil {
				if _, exists := unique[cancelIdx]; exists {
					continue
				}
				unique[cancelIdx] = struct{}{}
				cancelIndexes = append(cancelIndexes, cancelIdx)
			}
		}
		if len(cancelIndexes) == 0 {
			msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Введите корректные целые числа через запятую",
			})
			if err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
			fsmutils.AppendMessageID(h.f, userID, msg.ID)
			return
		}
		votings, ok := h.f.Get(userID, "votings")
		if !ok {
			h.f.Reset(userID)
			return
		}
		votingIDs := []int64{}
		for _, idx := range cancelIndexes {
			for votingIdx, voting := range votings.([]*model.Voting) {
				if int64(votingIdx+1) == idx {
					votingIDs = append(votingIDs, voting.ID)
				}
			}
		}
		paginatorMsgID, ok := h.f.Get(userID, "paginatorMsgID")
		if !ok {
			h.f.Reset(userID)
			return
		}
		ok, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    update.Message.Chat.ID,
			MessageID: paginatorMsgID.(int),
		})
		if err != nil || !ok {
			h.f.Reset(userID)
			return
		}
		h.f.Set(userID, "votingIDs", votingIDs)
		h.f.Transition(userID, stateCancel, userID, ctx, b, update)
	case stateTime:
		timeString := update.Message.Text
		var hour *int
		var minute *int
		fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
		iter := strings.SplitSeq(timeString, ":")
		for part := range iter {
			part = strings.TrimSpace(part)
			if hour == nil {
				parsedHour, err := strconv.Atoi(part)
				if err != nil || (parsedHour < 0 || parsedHour > 23) {
					break
				}
				hour = &parsedHour
			} else if minute == nil {
				parsedMinute, err := strconv.Atoi(part)
				if err != nil || (parsedMinute < 0 || parsedMinute > 59) {
					break
				}
				minute = &parsedMinute
			}
		}
		if hour == nil || minute == nil {
			msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Введите корректное время в формате ЧЧ:ММ",
			})
			if err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
			fsmutils.AppendMessageID(h.f, userID, msg.ID)
			return
		}
		h.f.Set(userID, "hour", *hour)
		h.f.Set(userID, "minute", *minute)
		h.f.Transition(userID, stateLocation, userID, ctx, b, update)
	case stateLocation:
		rawLocation := update.Message.Text
		fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
		location, err := time.LoadLocation(rawLocation)
		if err != nil {
			msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Введите корректную локацию (например, Europe/Moscow)",
			})
			if err != nil {
				log.Printf("Error sending message: %v", err)
				return
			}
			fsmutils.AppendMessageID(h.f, userID, msg.ID)
			return
		}
		h.f.Set(userID, "location", location.String())
		h.f.Transition(userID, stateSaveSchedule, userID, ctx, b, update)
	default:
		fmt.Printf("unexpected state %s\n", currentState)
	}
}
