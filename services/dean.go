package services

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func Dean_Handle(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Recipient.UserId)
	}
	msg.SetText("📅 Привет из сервиса «Расписание деканата».")
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}
