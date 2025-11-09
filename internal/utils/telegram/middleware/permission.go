package middleware

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func CheckIfInGroup(ctx context.Context, b *bot.Bot, update *models.Update, groupID int64) bool {
	var userID int64
	var chatID int64

	if update.Message != nil {
		userID = update.Message.From.ID
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
		chatID = update.CallbackQuery.Message.Message.MigrateToChatID
	} else if update.PollAnswer != nil {
		userID = update.PollAnswer.User.ID
		chatID = 0
	} else {
		log.Println("Update type is not supported for group check")
		return false
	}
	chatMember, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: groupID,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Authentication error -> %d is not a group member: %v", userID, err)
		if chatID != 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Вы не являетесь членом кинокласса!",
			})
		}
		log.Println("Error getting chat member")
		return false
	}
	if chatMember.Left != nil || chatMember.Banned != nil {
		log.Printf("Authentication error -> %d is not a group member", userID)
		if chatID != 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Вы не являетесь членом кинокласса!",
			})
		}
		log.Println("Chat member is not active")
		return false
	}

	return true
}

func CheckIfAdmin(ctx context.Context, b *bot.Bot, update *models.Update, groupID int64) bool {
	var userID int64
	var chatID int64

	if update.Message != nil {
		userID = update.Message.From.ID
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
		chatID = update.CallbackQuery.Message.Message.MigrateToChatID
	} else if update.PollAnswer != nil {
		userID = update.PollAnswer.User.ID
		chatID = 0
	} else {
		log.Println("Update type is not supported for admin check")
		return false
	}

	chatAdmins, err := b.GetChatAdministrators(ctx, &bot.GetChatAdministratorsParams{
		ChatID: groupID,
	})

	if err != nil {
		log.Printf("Destination error -> there is no such a chat -> %d: %v", groupID, err)
		if chatID != 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Нет такого чата!",
			})
		}
		return false
	}
	// Check if user is in admins list
	for _, chatMember := range chatAdmins {
		if chatMember.Owner != nil && chatMember.Owner.User.ID == userID {
			return true
		}
		if chatMember.Administrator != nil && chatMember.Administrator.User.ID == userID {
			return true
		}
	}

	log.Printf("Authentication error -> %d is not an administrator", userID)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Вы не являетесь администратором!",
	})
	return false
}

func Authentication(groupID int64) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if !CheckIfInGroup(ctx, b, update, groupID) {
				return
			}
			next(ctx, b, update)
		}
	}
}

func AdminOnly(groupID int64) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if !CheckIfInGroup(ctx, b, update, groupID) {
				return
			}

			if !CheckIfAdmin(ctx, b, update, groupID) {
				return
			}

			next(ctx, b, update)
		}
	}
}
