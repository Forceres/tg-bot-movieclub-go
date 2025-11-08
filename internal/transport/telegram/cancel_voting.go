package telegram

import (
	"context"
	"fmt"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	fsmutils "github.com/Forceres/tg-bot-movieclub-go/internal/utils/fsm"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
	"github.com/go-telegram/ui/paginator"
)

const statePrepareCancelIDs fsm.StateID = "prepare_cancel_ids"
const stateCancel fsm.StateID = "cancel"

type CancelVotingHandler struct {
	f             *fsm.FSM
	votingService service.IVotingService
}

type ICancelVotingHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	PrepareCancelIDs(f *fsm.FSM, args ...any)
	Cancel(f *fsm.FSM, args ...any)
}

func NewCancelVotingHandler(f *fsm.FSM, votingService service.IVotingService) ICancelVotingHandler {
	return &CancelVotingHandler{
		f:             f,
		votingService: votingService,
	}
}

func (h *CancelVotingHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	currentState := h.f.Current(userID)
	if currentState != stateDefault {
		return
	}
	votings, err := h.votingService.FindVotingByStatus("active")
	if err != nil || len(votings) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Нет активных голосований.",
		})
		return
	}
	opts := []paginator.Option{
		paginator.PerPage(5),
	}
	var formattedVotings []string
	for idx, voting := range votings {
		formattedVotings = append(formattedVotings, bot.EscapeMarkdown(fmt.Sprintf("%d. %s", idx+1, voting.Title)))
	}
	h.f.Set(userID, "votings", votings)
	p := paginator.New(b, formattedVotings, opts...)
	showOpts := []paginator.ShowOption{}
	paginatorMsg, err := p.Show(ctx, b, update.Message.Chat.ID, showOpts...)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при отображении голосований.",
		})
		return
	}
	fsmutils.AppendMessageID(h.f, userID, update.Message.ID)
	fsmutils.AppendMessageID(h.f, userID, paginatorMsg.ID)
	h.f.Transition(userID, statePrepareCancelIDs, userID, ctx, b, update, paginatorMsg.ID)
}

func (h *CancelVotingHandler) PrepareCancelIDs(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	paginatorMsgID := args[4].(int)
	f.Set(userID, "paginatorMsgID", paginatorMsgID)
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Перечислите номера голосований, которые хотите отменить, через запятую.",
	})
	if err != nil {
		return
	}
	fsmutils.AppendMessageID(f, userID, msg.ID)
}

func (h *CancelVotingHandler) Cancel(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := f.Current(userID)
	if currentState == stateDefault {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	ids, ok := f.Get(userID, "votingIDs")
	if !ok {
		f.Reset(userID)
		return
	}
	for _, id := range ids.([]int64) {
		voting, err := h.votingService.UpdateVotingStatus(&model.Voting{ID: id, Status: "cancelled"})
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Ошибка при отмене голосования с ID %d.", voting.ID),
			})
			continue
		}
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Выбранные голосования были отменены.",
	})
	h.f.Reset(userID)
}
