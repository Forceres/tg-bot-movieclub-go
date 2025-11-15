package telegram

import (
	"context"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func UpdateChatMemberMatchFunc(groupID int64) bot.MatchFunc {
	return func(update *models.Update) bool {
		return update != nil && (update.MyChatMember != nil || update.ChatMember != nil) && update.MyChatMember.Chat.ID == groupID
	}
}

type UpdateChatMemberHandler struct {
	userService service.IUserService
}

type IUpdateChatMemberHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewUpdateChatMemberHandler(userService service.IUserService) *UpdateChatMemberHandler {
	return &UpdateChatMemberHandler{userService: userService}
}

func (h *UpdateChatMemberHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	var user *model.User
	var role string = model.ROLE_USER
	if update.ChatMember != nil && (update.ChatMember.NewChatMember.Member != nil || update.ChatMember.OldChatMember.Member != nil) {
		user = &model.User{
			ID:        update.ChatMember.NewChatMember.Member.User.ID,
			FirstName: update.ChatMember.NewChatMember.Member.User.FirstName,
			LastName:  update.ChatMember.NewChatMember.Member.User.LastName,
			Username:  update.ChatMember.NewChatMember.Member.User.Username,
		}
	}
	if update.ChatMember != nil && (update.ChatMember.NewChatMember.Administrator != nil || update.ChatMember.OldChatMember.Administrator != nil || update.ChatMember.NewChatMember.Owner != nil || update.ChatMember.OldChatMember.Owner != nil) {
		role = model.ROLE_ADMIN
	}
	err := h.userService.Create(user, role)
	if err != nil {
		return
	}
}
