package services

import (
	"context"
	"strings"

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

// Пэйлоады для корпусов
const (
	CampusShowMap = "campus_show_map"
)

// Контекст сервисов
type Ctx struct {
	API *maxbot.Api
	DB  *gorm.DB
}

// ГЛАВНЫЙ РОУТЕР КНОПОК (из main.go для MessageCallbackUpdate)
func Route(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	// Обработка возврата в главное меню
	if upd.Callback.Payload == "back_to_menu" {
		return showMainMenu(ctx, sc, upd.Message.Recipient)
	}

	// Обработка выбора корпуса (формат: "campus_1", "campus_2")
	if strings.HasPrefix(upd.Callback.Payload, "campus_") {
		// Проверяем, это не кнопка "показать на карте"
		if !strings.HasPrefix(upd.Callback.Payload, CampusShowMap) {
			return handleCampusSelection(ctx, sc, upd)
		}
	}

	// Обработка кнопки "Показать на карте"
	if strings.HasPrefix(upd.Callback.Payload, CampusShowMap) {
		return handleCampusMap(ctx, sc, upd)
	}

	switch upd.Callback.Payload {
	case ServiceFindTeacher:
		return FT_ShowModeMenu(ctx, sc, upd)

	// Под-обработчики поиска (выбор режима)
	case FT_FindByFaculty, FT_FindByDepartment, FT_FindByFIO:
		return FT_AskForQuery(ctx, sc, upd)

	// Обработчики корпусов
	case ServiceCampusInfo:
		return Campus_Handle(ctx, sc, upd)

	// Остальные разделы — заглушки
	case ServiceDeanSchedule:
		return Dean_Handle(ctx, sc, upd)
	case ServiceFoodAndCopy:
		return Places_Handle(ctx, sc, upd)
	case ServiceFAQ:
		return FAQ_Handle(ctx, sc, upd)

	default:
		// неизвестный payload
		msg := maxbot.NewMessage()
		setRecipient(msg, upd.Message.Recipient)
		msg.SetText("Неизвестная команда. Нажмите кнопку меню.")
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}
}

// ОБРАБОТКА ТЕКСТОВЫХ СООБЩЕНИЙ (делегируем в сервис поиска)
func OnMessage(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate) (bool, error) {
	// Сначала пробуем обработать как запрос преподавателя
	if handled, err := FT_OnMessage(ctx, sc, upd); handled || err != nil {
		return handled, err
	}

	// Затем пробуем обработать как запрос корпуса
	if handled, err := Campus_OnMessage(ctx, sc, upd); handled || err != nil {
		return handled, err
	}

	// Не обработано
	return false, nil
}

// showMainMenu - показывает главное меню
func showMainMenu(ctx context.Context, sc Ctx, recipient schemes.Recipient) error {
	msg := maxbot.NewMessage()
	setRecipient(msg, recipient)
	msg.SetText(WelcomeText()).AddKeyboard(MenuKeyboard(sc.API))
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// setRecipient - вспомогательная функция для установки получателя
func setRecipient(msg *maxbot.Message, recipient schemes.Recipient) {
	if recipient.ChatId != 0 {
		msg.SetChat(recipient.ChatId)
	} else {
		msg.SetUser(recipient.UserId)
	}
}

// Главное меню (клавиатура)
func MenuKeyboard(api *maxbot.Api) *maxbot.Keyboard {
	kb := api.Messages.NewKeyboardBuilder()
	kb.AddRow().
		AddCallback("1) Поиск препода", schemes.POSITIVE, ServiceFindTeacher).
		AddCallback("2) Деканат", schemes.POSITIVE, ServiceDeanSchedule)
	kb.AddRow().
		AddCallback("3) Корпуса", schemes.POSITIVE, ServiceCampusInfo).
		AddCallback("4) Столовые/копирки", schemes.POSITIVE, ServiceFoodAndCopy)
	kb.AddRow().
		AddCallback("5) Частые вопросы", schemes.NEGATIVE, ServiceFAQ)
	return kb
}

func WelcomeText() string { return "Выбери раздел 👇" }
