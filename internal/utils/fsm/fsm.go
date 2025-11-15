package fsmutils

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/fsm"
)

func GetMessageIDs(f *fsm.FSM, userID int64) ([]int, bool) {
	messageIDs, ok := f.Get(userID, "messageIDs")
	if ok {
		return messageIDs.([]int), true
	}
	return nil, false
}

func AppendMessageID(f *fsm.FSM, userID int64, messageID int) {
	messageIDs, ok := f.Get(userID, "messageIDs")
	if ok {
		f.Set(userID, "messageIDs", append(messageIDs.([]int), messageID))
	} else {
		f.Set(userID, "messageIDs", []int{messageID})
	}
}

func DeleteMessages(ctx context.Context, b *bot.Bot, f *fsm.FSM, userID int64, chatID int64) {
	msgIDs, ok := GetMessageIDs(f, userID)
	if ok {
		_, err := b.DeleteMessages(ctx, &bot.DeleteMessagesParams{
			ChatID:     chatID,
			MessageIDs: msgIDs,
		})
		if err != nil {
			log.Printf("Error deleting messages: %v", err)
		}
	}
}
