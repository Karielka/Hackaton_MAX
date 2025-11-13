package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Karielka/Hackaton_MAX/models"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

// Campus_Handle - обработчик меню "Корпуса"
func Campus_Handle(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	return showCampusSelection(ctx, sc, upd.Message.Recipient)
}

// showCampusSelection - показывает список корпусов для выбора
func showCampusSelection(ctx context.Context, sc Ctx, recipient schemes.Recipient) error {
	var campuses []models.Campus
	if err := sc.DB.Find(&campuses).Error; err != nil {
		return fmt.Errorf("failed to fetch campuses: %w", err)
	}

	if len(campuses) == 0 {
		msg := maxbot.NewMessage()
		setRecipient(msg, recipient)
		msg.SetText("Информация о корпусах временно недоступна.")
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}

	kb := sc.API.Messages.NewKeyboardBuilder()

	for i := 0; i < len(campuses); i += 2 {
		row := kb.AddRow()
		row.AddCallback(campuses[i].ShortName, schemes.POSITIVE, fmt.Sprintf("campus_%d", campuses[i].ID))

		if i+1 < len(campuses) {
			row.AddCallback(campuses[i+1].ShortName, schemes.POSITIVE, fmt.Sprintf("campus_%d", campuses[i+1].ID))
		}
	}

	kb.AddRow().AddCallback("◀️ Назад", schemes.NEGATIVE, "back_to_menu")

	msg := maxbot.NewMessage()
	setRecipient(msg, recipient)
	msg.SetText("🏫 Выберите корпус:").AddKeyboard(kb)

	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// handleCampusSelection - обработчик выбора конкретного корпуса
func handleCampusSelection(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	campusID := strings.TrimPrefix(upd.Callback.Payload, "campus_")

	var campus models.Campus
	if err := sc.DB.First(&campus, campusID).Error; err != nil {
		msg := maxbot.NewMessage()
		setRecipient(msg, upd.Message.Recipient)
		msg.SetText("Корпус не найден.")
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}

	return sendCampusInfo(ctx, sc, campus, upd.Message.Recipient)
}

// sendCampusInfo - отправляет информацию о корпусе
func sendCampusInfo(ctx context.Context, sc Ctx, campus models.Campus, recipient schemes.Recipient) error {
	text := fmt.Sprintf(
		"🏫 %s (%s)\n\n📍 Адрес: %s\n🚇 Метро: %s\n\nЧто находится внутри:\n%s",
		campus.FullName,
		campus.ShortName,
		campus.Address,
		campus.Metro,
		campus.Description,
	)

	kb := sc.API.Messages.NewKeyboardBuilder()
	kb.AddRow().
		AddCallback("🗺️ Показать на карте", schemes.POSITIVE, fmt.Sprintf("%s_%d", CampusShowMap, campus.ID))
	kb.AddRow().
		AddCallback("◀️ К списку корпусов", schemes.NEGATIVE, ServiceCampusInfo).
		AddCallback("🏠 Главное меню", schemes.NEGATIVE, "back_to_menu")

	//TODO сделать чтоб фотка была в виде токена - как - хз

	//if campus.ImageURL != "" {
	//	photoMsg := maxbot.NewMessage().AddPhoto()
	//	setRecipient(photoMsg, recipient)
	//	if _, err := sc.API.Messages.Send(ctx, photoMsg); err != nil {
	//		// Логируем ошибку, но продолжаем отправлять текст
	//		fmt.Printf("Failed to send image: %v\n", err)
	//	}
	//}

	msg := maxbot.NewMessage()
	setRecipient(msg, recipient)
	msg.SetText(text).AddKeyboard(kb)

	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// handleCampusMap - обработчик кнопки "Показать на карте"
func handleCampusMap(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	payload := strings.TrimPrefix(upd.Callback.Payload, CampusShowMap+"_")

	var campus models.Campus
	if err := sc.DB.First(&campus, payload).Error; err != nil {
		msg := maxbot.NewMessage()
		setRecipient(msg, upd.Message.Recipient)
		msg.SetText("Информация о расположении корпуса временно недоступна.")
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}

	//TODO - фотка карты пока не сделана

	//if campus.MapImageURL != "" {
	//	msg := maxbot.NewMessage()
	//	setRecipient(msg, upd.Message.Recipient)
	//	msg.SetImage(campus.MapImageURL)
	//	msg.SetText(fmt.Sprintf("🗺️ %s на карте", campus.FullName))
	//
	//	_, err := sc.API.Messages.Send(ctx, msg)
	//	return err
	//}

	msg := maxbot.NewMessage()
	setRecipient(msg, upd.Message.Recipient)
	msg.SetText(fmt.Sprintf("🗺️ %s\n📍 %s\n🚇 %s", campus.FullName, campus.Address, campus.Metro))

	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// Campus_OnMessage - обработка текстовых запросов по корпусам
func Campus_OnMessage(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate) (bool, error) {
	text := strings.TrimSpace(upd.Message.Body.Text)
	if text == "" {
		return false, nil
	}

	var campus models.Campus
	query := sc.DB.Where("LOWER(short_name) = LOWER(?) OR LOWER(full_name) LIKE LOWER(?)",
		text, "%"+text+"%")

	if err := query.First(&campus).Error; err != nil {
		// Корпус не найден
		return false, nil
	}

	recipient := schemes.Recipient{}
	if upd.Message.Recipient.ChatId != 0 {
		recipient.ChatId = upd.Message.Recipient.ChatId
	} else {
		recipient.UserId = upd.Message.Sender.UserId
	}

	err := sendCampusInfo(ctx, sc, campus, recipient)
	return true, err
}
