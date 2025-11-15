package middleware

import (
	"context"
	"log"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func CheckIfInGroup(ctx context.Context, b *bot.Bot, update *models.Update, groupID int64, userService service.IUserService) bool {
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

	if b.ID() == userID {
		return true
	}

	chatMember, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: groupID,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Authentication error -> %d is not a group member: %v", userID, err)
		if chatID != 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Вы не являетесь членом кинокласса!",
			})
			if err != nil {
				log.Printf("Error sending not a group member message: %v", err)
			}
		}
		log.Println("Error getting chat member")
		return false
	}
	if chatMember.Left != nil || chatMember.Banned != nil {
		log.Printf("Authentication error -> %d is not a group member", userID)
		if chatID != 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Вы не являетесь членом кинокласса!",
			})
			if err != nil {
				log.Printf("Error sending not a group member message: %v", err)
			}
		}
		log.Println("Chat member is not active")
		return false
	}

	if update.Message != nil && (update.Message.Text == "/register" || update.Message.Text == "/start") {
		return true
	}

	_, err = userService.FindByID(userID)
	if err != nil {
		log.Printf("User %d not found in database, attempting auto-registration", userID)
		var userName, firstName, lastName string
		if update.Message != nil && update.Message.From != nil {
			userName = update.Message.From.Username
			firstName = update.Message.From.FirstName
			lastName = update.Message.From.LastName
		} else if update.CallbackQuery != nil {
			userName = update.CallbackQuery.From.Username
			firstName = update.CallbackQuery.From.FirstName
			lastName = update.CallbackQuery.From.LastName
		} else if update.PollAnswer != nil && update.PollAnswer.User != nil {
			userName = update.PollAnswer.User.Username
			firstName = update.PollAnswer.User.FirstName
			lastName = update.PollAnswer.User.LastName
		}
		role := model.ROLE_USER
		chatAdmins, adminErr := b.GetChatAdministrators(ctx, &bot.GetChatAdministratorsParams{
			ChatID: groupID,
		})
		if adminErr == nil {
			for _, admin := range chatAdmins {
				if (admin.Owner != nil && admin.Owner.User.ID == userID) ||
					(admin.Administrator != nil && admin.Administrator.User.ID == userID) {
					role = model.ROLE_ADMIN
					break
				}
			}
		}
		createErr := userService.Create(&model.User{
			ID:        userID,
			FirstName: firstName,
			LastName:  lastName,
			Username:  userName,
		}, role)
		if createErr != nil {
			log.Printf("Failed to auto-register user %d: %v", userID, createErr)
			if chatID != 0 {
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "❌ Ошибка автоматической регистрации. Пожалуйста, используйте команду /register.",
				})
				if err != nil {
					log.Printf("Error sending auto-registration failure message: %v", err)
				}
			}
			return false
		}
		log.Printf("Successfully auto-registered user %d with role %s", userID, role)
		if chatID != 0 {
			welcomeText := "✅ Вы успешно зарегистрированы в системе кинокласса!"
			if role == model.ROLE_ADMIN {
				welcomeText += "\n👑 Вам присвоена роль администратора."
			}
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   welcomeText,
			})
			if err != nil {
				log.Printf("Error sending welcome message: %v", err)
			}
		}
	}

	return true
}

func CheckIfAdmin(ctx context.Context, b *bot.Bot, update *models.Update, groupID int64, userService service.IUserService) bool {
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
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ Нет такого чата!",
			})
			if err != nil {
				log.Printf("Error sending no such chat message: %v", err)
			}
		}
		return false
	}

	for _, chatMember := range chatAdmins {
		if chatMember.Owner != nil && chatMember.Owner.User.ID == userID {
			return true
		}
		if chatMember.Administrator != nil && chatMember.Administrator.User.ID == userID {
			return true
		}
	}

	user, err := userService.FindByID(userID)
	if err != nil {
		log.Printf("User %d not found in database for admin check", userID)
		return false
	}

	if user.Role.Name == model.ROLE_ADMIN {
		return true
	}

	log.Printf("Authentication error -> %d is not an administrator", userID)
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "❌ Вы не являетесь администратором!",
	})
	if err != nil {
		log.Printf("Error sending admin check failure message: %v", err)
	}
	return false
}

func Authentication(groupID int64, userService service.IUserService) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if !CheckIfInGroup(ctx, b, update, groupID, userService) {
				return
			}
			next(ctx, b, update)
		}
	}
}

func AdminOnly(groupID int64, userService service.IUserService) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if !CheckIfAdmin(ctx, b, update, groupID, userService) {
				return
			}

			next(ctx, b, update)
		}
	}
}
