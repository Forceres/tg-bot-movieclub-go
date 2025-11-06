package permission

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
    } else {
        return false
    }
    chatMember, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
        ChatID: groupID,
        UserID: userID,
    })
    if err != nil {
        log.Printf("Authentication error -> %d is not a group member: %v", userID, err)
        b.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: chatID,
            Text:   "Вы не являетесь членом кинокласса!",
        })
        return false
    }
    // Check if user is member
    if chatMember.Left != nil || chatMember.Banned != nil {
        log.Printf("Authentication error -> %d is not a group member", userID)
        b.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: chatID,
            Text:   "Вы не являетесь членом кинокласса!",
        })
        return false
    }

    return true
}

// CheckIfAdmin checks if user is an administrator of the group
func CheckIfAdmin(ctx context.Context, b *bot.Bot, update *models.Update, groupID int64) bool {
    var userID int64
    var chatID int64

    if update.Message != nil {
        userID = update.Message.From.ID
        chatID = update.Message.Chat.ID
    } else if update.CallbackQuery != nil {
        userID = update.CallbackQuery.From.ID
        chatID = update.CallbackQuery.Message.Message.MigrateToChatID
    } else {
        return false
    }

    chatAdmins, err := b.GetChatAdministrators(ctx, &bot.GetChatAdministratorsParams{
        ChatID: groupID,
    })

    if err != nil {
        log.Printf("Destination error -> there is no such a chat -> %d: %v", groupID, err)
        b.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: chatID,
            Text:   "Нет такого чата!",
        })
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

// Authentication middleware decorator
func Authentication(groupID int64, handler bot.HandlerFunc) bot.HandlerFunc {
    return func(ctx context.Context, b *bot.Bot, update *models.Update) {
        if !CheckIfInGroup(ctx, b, update, groupID) {
            return
        }
        handler(ctx, b, update)
    }
}

// AdminOnly middleware decorator
func AdminOnly(groupID int64, handler bot.HandlerFunc) bot.HandlerFunc {
    return func(ctx context.Context, b *bot.Bot, update *models.Update) {
        if !CheckIfInGroup(ctx, b, update, groupID) {
            return
        }

        if !CheckIfAdmin(ctx, b, update, groupID) {
            return
        }

        handler(ctx, b, update)
    }
}