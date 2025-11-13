package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Karielka/Hackaton_MAX/models"
	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

// Places_Handle - обработчик меню "Столовые/копирки"
func Places_Handle(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	return showCampusSelectionForPlaces(ctx, sc, upd.Message.Recipient)
}

// showCampusSelectionForPlaces - показывает выбор корпуса для мест
func showCampusSelectionForPlaces(ctx context.Context, sc Ctx, recipient schemes.Recipient) error {
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
		row.AddCallback(campuses[i].ShortName, schemes.POSITIVE,
			fmt.Sprintf("places_campus_%d", campuses[i].ID))

		if i+1 < len(campuses) {
			row.AddCallback(campuses[i+1].ShortName, schemes.POSITIVE,
				fmt.Sprintf("places_campus_%d", campuses[i+1].ID))
		}
	}

	kb.AddRow().AddCallback("◀️ Назад", schemes.NEGATIVE, "back_to_menu")

	msg := maxbot.NewMessage()
	setRecipient(msg, recipient)
	msg.SetText("🏢 О каком корпусе идет речь?").AddKeyboard(kb)

	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// handleCampusSelectionForPlaces - обработчик выбора корпуса для мест
func handleCampusSelectionForPlaces(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	campusID := strings.TrimPrefix(upd.Callback.Payload, "places_campus_")

	var campus models.Campus
	if err := sc.DB.First(&campus, campusID).Error; err != nil {
		msg := maxbot.NewMessage()
		setRecipient(msg, upd.Message.Recipient)
		msg.SetText("Корпус не найден.")
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}

	return showPlaceTypesMenu(ctx, sc, campus, upd.Message.Recipient)
}

// showPlaceTypesMenu - показывает меню типов мест в корпусе
func showPlaceTypesMenu(ctx context.Context, sc Ctx, campus models.Campus, recipient schemes.Recipient) error {
	var placeTypes []string
	if err := sc.DB.Model(&models.Place{}).
		Where("campus_id = ?", campus.ID).
		Distinct("type").
		Pluck("type", &placeTypes).Error; err != nil {
		return fmt.Errorf("failed to fetch place types: %w", err)
	}

	kb := sc.API.Messages.NewKeyboardBuilder()

	hasCanteen := false
	hasBuffet := false
	hasCopy := false

	for _, placeType := range placeTypes {
		switch placeType {
		case "canteen":
			hasCanteen = true
			kb.AddRow().AddCallback("🍽️ Столовая", schemes.POSITIVE,
				fmt.Sprintf("places_canteen_%d", campus.ID))
		case "buffet":
			hasBuffet = true
		case "copy":
			hasCopy = true
			kb.AddRow().AddCallback("📄 Копирки", schemes.POSITIVE,
				fmt.Sprintf("places_copy_%d", campus.ID))
		}
	}

	if hasBuffet {
		kb.AddRow().AddCallback("☕ Буфеты", schemes.POSITIVE,
			fmt.Sprintf("places_buffet_%d", campus.ID))
	}

	kb.AddRow().
		AddCallback("◀️ К выбору корпуса", schemes.NEGATIVE, ServiceFoodAndCopy).
		AddCallback("🏠 Главное меню", schemes.NEGATIVE, "back_to_menu")

	msg := maxbot.NewMessage()
	setRecipient(msg, recipient)

	text := fmt.Sprintf("🏢 %s\n\nЧто вас интересует?", campus.FullName)
	if !hasCanteen && !hasBuffet && !hasCopy {
		text = fmt.Sprintf("🏢 %s\n\nВ этом корпусе пока нет информации о столовых, буфетах или копирках.", campus.FullName)
	}

	msg.SetText(text).AddKeyboard(kb)
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// handlePlaceTypeSelection - обработчик выбора типа места
func handlePlaceTypeSelection(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	payload := upd.Callback.Payload

	var placeType string
	var campusID string

	if strings.HasPrefix(payload, "places_canteen_") {
		placeType = "canteen"
		campusID = strings.TrimPrefix(payload, "places_canteen_")
	} else if strings.HasPrefix(payload, "places_buffet_") {
		placeType = "buffet"
		campusID = strings.TrimPrefix(payload, "places_buffet_")
	} else if strings.HasPrefix(payload, "places_copy_") {
		placeType = "copy"
		campusID = strings.TrimPrefix(payload, "places_copy_")
	} else {
		return fmt.Errorf("unknown place type payload: %s", payload)
	}

	return showPlacesByType(ctx, sc, placeType, campusID, upd.Message.Recipient)
}

// showPlacesByType - показывает места определенного типа в корпусе
func showPlacesByType(ctx context.Context, sc Ctx, placeType, campusID string, recipient schemes.Recipient) error {
	var places []models.Place
	if err := sc.DB.Where("campus_id = ? AND type = ?", campusID, placeType).Find(&places).Error; err != nil {
		return fmt.Errorf("failed to fetch places: %w", err)
	}

	if len(places) == 0 {
		msg := maxbot.NewMessage()
		setRecipient(msg, recipient)

		typeName := map[string]string{
			"canteen": "столовых",
			"buffet":  "буфетов",
			"copy":    "копировальных центров",
		}[placeType]

		msg.SetText(fmt.Sprintf("В этом корпусе нет %s.", typeName))
		_, err := sc.API.Messages.Send(ctx, msg)
		return err
	}

	if placeType == "canteen" && len(places) > 0 {
		return showCanteenDetails(ctx, sc, places[0], campusID, recipient)
	}

	return showPlacesList(ctx, sc, places, placeType, campusID, recipient)
}

// showCanteenDetails - показывает детальную информацию о столовой
func showCanteenDetails(ctx context.Context, sc Ctx, place models.Place, campusID string, recipient schemes.Recipient) error {
	var campus models.Campus
	if err := sc.DB.First(&campus, campusID).Error; err != nil {
		return fmt.Errorf("failed to fetch campus: %w", err)
	}

	text := fmt.Sprintf(
		"🍽️ %s (%s)\n📍 Расположение: %s\n🕐 Режим работы: %s\n\n📋 Меню на сегодня:\n%s",
		place.Name,
		campus.ShortName,
		place.Location,
		place.Schedule,
		place.MenuToday,
	)

	kb := sc.API.Messages.NewKeyboardBuilder()

	var otherTypes []string
	sc.DB.Model(&models.Place{}).
		Where("campus_id = ? AND type != ?", campusID, "canteen").
		Distinct("type").
		Pluck("type", &otherTypes)

	if len(otherTypes) > 0 {
		kb.AddRow().AddCallback("📋 Буфеты и копирки в этом корпусе", schemes.POSITIVE,
			fmt.Sprintf("places_back_to_campus_%s", campusID))
	}

	kb.AddRow().
		AddCallback("◀️ К выбору типа", schemes.NEGATIVE, fmt.Sprintf("places_campus_%s", campusID)).
		AddCallback("🏠 Главное меню", schemes.NEGATIVE, "back_to_menu")

	msg := maxbot.NewMessage()
	setRecipient(msg, recipient)
	msg.SetText(text).AddKeyboard(kb)

	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// showPlacesList - показывает список мест (буфетов или копирок)
func showPlacesList(ctx context.Context, sc Ctx, places []models.Place, placeType, campusID string, recipient schemes.Recipient) error {
	var campus models.Campus
	if err := sc.DB.First(&campus, campusID).Error; err != nil {
		return fmt.Errorf("failed to fetch campus: %w", err)
	}

	typeName := map[string]string{
		"buffet": "Буфеты",
		"copy":   "Копировальные центры",
	}[placeType]

	var b strings.Builder
	b.WriteString(fmt.Sprintf("📋 %s в %s:\n\n", typeName, campus.ShortName))

	for i, place := range places {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, place.Name))
		b.WriteString(fmt.Sprintf("   📍 %s\n", place.Location))
		b.WriteString(fmt.Sprintf("   🕐 %s\n", place.Schedule))

		if placeType == "buffet" {
			// Для буфетов показываем примеры из меню
			lines := strings.Split(place.MenuToday, "\n")
			if len(lines) > 0 {
				b.WriteString(fmt.Sprintf("   🍽️ %s\n", lines[0]))
			}
		} else if placeType == "copy" {
			// Для копирок показываем услуги
			lines := strings.Split(place.MenuToday, "\n")
			if len(lines) > 0 {
				b.WriteString(fmt.Sprintf("   📄 %s\n", lines[0]))
			}
		}
		b.WriteString("\n")
	}

	kb := sc.API.Messages.NewKeyboardBuilder()
	kb.AddRow().
		AddCallback("◀️ К выбору типа", schemes.NEGATIVE, fmt.Sprintf("places_campus_%s", campusID)).
		AddCallback("🏠 Главное меню", schemes.NEGATIVE, "back_to_menu")

	msg := maxbot.NewMessage()
	setRecipient(msg, recipient)
	msg.SetText(b.String()).AddKeyboard(kb)

	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// Places_OnMessage - обработка текстовых запросов по местам
func Places_OnMessage(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate) (bool, error) {
	text := strings.TrimSpace(strings.ToLower(upd.Message.Body.Text))
	if text == "" {
		return false, nil
	}

	keywords := []string{"столовая", "буфет", "копирка", "копир", "еда", "печать", "распечатать"}
	hasKeyword := false
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			hasKeyword = true
			break
		}
	}

	if !hasKeyword {
		return false, nil
	}

	recipient := schemes.Recipient{}
	if upd.Message.Recipient.ChatId != 0 {
		recipient.ChatId = upd.Message.Recipient.ChatId
	} else {
		recipient.UserId = upd.Message.Sender.UserId
	}

	err := showCampusSelectionForPlaces(ctx, sc, recipient)
	return true, err
}
