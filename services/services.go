package services

import (
	"context"
	"fmt"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"gorm.io/gorm"
)

// Payload для кнопок
const (
	ServiceFindTeacher  = "svc_find_teacher"
	ServiceDeanSchedule = "svc_dean_schedule"
	ServiceCampusInfo   = "svc_campus_info"
	ServiceFoodAndCopy  = "svc_food_copy"
	ServiceFAQ          = "svc_faq"
)

// Контекст сервиса: API + DB
type Ctx struct {
	API *maxbot.Api
	DB  *gorm.DB
}

// Единая точка входа из main.go по payload
func Route(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	switch upd.Callback.Payload {
	case ServiceFindTeacher:
		return FindTeacher(ctx, sc, upd)
	case ServiceDeanSchedule:
		return DeanSchedule(ctx, sc, upd)
	case ServiceCampusInfo:
		return CampusInfo(ctx, sc, upd)
	case ServiceFoodAndCopy:
		return FoodAndCopy(ctx, sc, upd)
	case ServiceFAQ:
		return FAQ(ctx, sc, upd)
	default:
		msg := maxbot.NewMessage()
		if upd.Message.Recipient.ChatId != 0 {
			msg.SetChat(upd.Message.Recipient.ChatId)
		} else if upd.Message.Recipient.UserId != 0 {
			msg.SetUser(upd.Message.Recipient.UserId)
		}
		msg.SetText("Неизвестная команда. Нажми одну из кнопок меню.")
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}
}

// ---- Заглушки сервисов ----

func FindTeacher(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	return reply(ctx, sc, upd, "👩‍🏫 Привет из сервиса «Поиск препода».")
}

func DeanSchedule(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	return reply(ctx, sc, upd, "📅 Привет из сервиса «Расписание деканата».")
}

func CampusInfo(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	return reply(ctx, sc, upd, "🏫 Привет из сервиса «Информация по корпусам».")
}

func FoodAndCopy(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	return reply(ctx, sc, upd, "🍽️ Привет из сервиса «Столовые/буфеты/копирки».")
}

func FAQ(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	return reply(ctx, sc, upd, "❓ Привет из сервиса «Частые вопросы».")
}

// ---- Вспомогалки ----

func reply(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate, text string) error {
	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else if upd.Message.Recipient.UserId != 0 {
		msg.SetUser(upd.Message.Recipient.UserId)
	}
	msg.SetText(text)
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// Клавиатура из 5 кнопок
// ВАЖНО: возвращаем ЗНАЧЕНИЕ, т.к. kb.Build() -> schemes.Keyboard (не *schemes.Keyboard)
func MenuKeyboard(api *maxbot.Api) schemes.Keyboard {
	kb := api.Messages.NewKeyboardBuilder()
	kb.AddRow().
		AddCallback("1) Поиск препода", schemes.POSITIVE, ServiceFindTeacher).
		AddCallback("2) Деканат", schemes.POSITIVE, ServiceDeanSchedule)
	kb.AddRow().
		// В твоей версии SDK нет SECONDARY, используем PRIMARY
		AddCallback("3) Корпуса", schemes.POSITIVE, ServiceCampusInfo).
		AddCallback("4) Столовые/копирки", schemes.POSITIVE, ServiceFoodAndCopy)
	kb.AddRow().
		AddCallback("5) Частые вопросы", schemes.NEGATIVE, ServiceFAQ)
	return kb.Build()
}

func WelcomeText() string {
	return "Выбери раздел 👇"
}

func UnknownText(cmd string) string {
	return fmt.Sprintf("Команда не поддерживается: %s. Нажми кнопку меню.", cmd)
}
