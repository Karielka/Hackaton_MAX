package services

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"gorm.io/gorm"
)

// Пэйлоады главного меню
const (
	ServiceFindTeacher  = "svc_find_teacher"
	ServiceDeanSchedule = "svc_dean_schedule"
	ServiceCampusInfo   = "svc_campus_info"
	ServiceFoodAndCopy  = "svc_food_copy"
	ServiceFAQ          = "svc_faq"
)

// Контекст сервисов
type Ctx struct {
	API *maxbot.Api
	DB  *gorm.DB
}

// ГЛАВНЫЙ РОУТЕР КНОПОК (из main.go для MessageCallbackUpdate)
func Route(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	switch upd.Callback.Payload {

	case ServiceFindTeacher:
		return FT_ShowModeMenu(ctx, sc, upd) // открываем подменю поиска

	// Под-обработчики поиска (выбор режима)
	case FT_FindByFaculty, FT_FindByDepartment, FT_FindByFIO:
		return FT_AskForQuery(ctx, sc, upd)

	// Остальные разделы — заглушки
	case ServiceDeanSchedule:
		return Dean_Handle(ctx, sc, upd)
	case ServiceCampusInfo:
		return Campus_Handle(ctx, sc, upd)
	case ServiceFoodAndCopy:
		return Places_Handle(ctx, sc, upd)
	case ServiceFAQ:
		return FAQ_Handle(ctx, sc, upd)

	default:
		// неизвестный payload
		msg := maxbot.NewMessage()
		if upd.Message.Recipient.ChatId != 0 {
			msg.SetChat(upd.Message.Recipient.ChatId)
		} else if upd.Message.Recipient.UserId != 0 {
			msg.SetUser(upd.Message.Recipient.UserId)
		}
		msg.SetText("Неизвестная команда. Нажмите кнопку меню.")
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}
}

// ОБРАБОТКА ТЕКСТОВЫХ СООБЩЕНИЙ (делегируем в сервис поиска)
func OnMessage(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate) (bool, error) {
	return FT_OnMessage(ctx, sc, upd) // true если сообщение обработано сценарием поиска
}

// Главное меню (клавиатура)
func MenuKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	kb := api.Messages.NewKeyboardBuilder()
	kb.AddRow().
		AddCallback("1) Поиск препода", schemes.POSITIVE, ServiceFindTeacher).
		AddCallback("2) Деканат",       schemes.POSITIVE, ServiceDeanSchedule)
	kb.AddRow().
		AddCallback("3) Корпуса",       schemes.POSITIVE, ServiceCampusInfo).
		AddCallback("4) Столовые/копирки", schemes.POSITIVE, ServiceFoodAndCopy)
	kb.AddRow().
		AddCallback("5) Частые вопросы", schemes.NEGATIVE, ServiceFAQ)
	return kb // AddKeyboard ждёт *maxbot.Keyboard (билдер), не вызываем Build()
}

func WelcomeText() string { return "Выбери раздел 👇" }
