package fsmutils

import "github.com/go-telegram/fsm"

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
