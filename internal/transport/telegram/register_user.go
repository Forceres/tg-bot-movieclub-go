package telegram

import (
	"context"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type RegisterUserHandler struct {
	userService service.IUserService
}

type IRegisterUserHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewRegisterUserHandler(userService service.IUserService) *RegisterUserHandler {
	return &RegisterUserHandler{userService: userService}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	var user *models.User
	var role string = "USER"
	if update.Message != nil && update.Message.From != nil {
		user = update.Message.From
	}
	if update.ChatMember != nil && (update.ChatMember.NewChatMember.Administrator != nil || update.ChatMember.OldChatMember.Administrator != nil || update.ChatMember.NewChatMember.Owner != nil || update.ChatMember.OldChatMember.Owner != nil) {
		role = "ADMIN"
	}
	err := h.userService.CreateIfNotExist(&model.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	}, role)
	if err != nil {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Произошла ошибка при регистрации пользователя.",
		})
		if err != nil {
			log.Printf("Error sending error message: %v", err)
		}
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Пользователь успешно зарегистрирован.",
	})
	if err != nil {
		return
	}
}
