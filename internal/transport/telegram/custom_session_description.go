package telegram

import (
	"context"
	"fmt"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
)

const (
	stateDescription     fsm.StateID = "description"
	stateSaveDescription fsm.StateID = "save_description"
)

type CustomSessionDescriptionHandler struct {
	sessionService service.ISessionService
	fsm            *fsm.FSM
}

func NewCustomSessionDescriptionHandler(sessionService service.ISessionService, f *fsm.FSM) *CustomSessionDescriptionHandler {
	return &CustomSessionDescriptionHandler{
		sessionService: sessionService,
		fsm:            f,
	}
}

func (h *CustomSessionDescriptionHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID

	currentState := h.fsm.Current(userID)
	if currentState != stateDefault {
		return
	}

	session, err := h.sessionService.FindOngoingSession()
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Нет активной сессии для установки описания.",
		})
		if err != nil {
			fmt.Printf("failed to send message: %v\n", err)
		}
		return
	}
	h.fsm.Set(userID, "session_id", session.ID)
	h.fsm.Set(userID, "chat_id", update.Message.Chat.ID)
	h.fsm.Transition(userID, stateDescription, userID, ctx, b, update, session)
}

func (h *CustomSessionDescriptionHandler) HandleDescriptionInput(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)
	currentState := h.fsm.Current(userID)
	if currentState != stateDescription {
		return
	}
	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)
	session := args[4].(*model.Session)
	currentDesc := session.Description
	promptText := "📝 Отправьте описание для текущей сессии."
	if currentDesc != "" {
		promptText += fmt.Sprintf("\n\n💡 Текущее описание:\n%s", currentDesc)
	}
	promptText += "\n\nℹ️ Отправьте /cancel для отмены."
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   promptText,
	})
	if err != nil {
		fmt.Printf("failed to send message: %v\n", err)
	}
}

func (h *CustomSessionDescriptionHandler) SaveDescription(f *fsm.FSM, args ...any) {
	userID := args[0].(int64)

	currentState := h.fsm.Current(userID)
	if currentState != stateSaveDescription {
		return
	}

	ctx := args[1].(context.Context)
	b := args[2].(*bot.Bot)
	update := args[3].(*models.Update)

	sessionIDVal, exists := h.fsm.Get(userID, "session_id")
	if !exists {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка: сессия не найдена. Попробуйте снова с /custom.",
		})
		if err != nil {
			fmt.Printf("failed to send message: %v\n", err)
		}
		h.fsm.Reset(userID)
		return
	}
	sessionID := sessionIDVal.(int64)
	description, _ := h.fsm.Get(userID, "description")

	err := h.sessionService.UpdateSessionDescription(sessionID, description.(string))
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Ошибка при обновлении описания сессии. Попробуйте позже.",
		})
		if err != nil {
			fmt.Printf("failed to send message: %v\n", err)
		}
		h.fsm.Transition(userID, stateDefault)
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("✅ Описание сессии обновлено:\n\n%s", description),
	})
	if err != nil {
		fmt.Printf("failed to send message: %v\n", err)
	}

	h.fsm.Reset(userID)
}
