package services

import (
	"context"
	"fmt"
	"strings"
	"sync"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
	"gorm.io/gorm"

	"github.com/Karielka/Hackaton_MAX/models"
)

// Можно оставить для будущего разветвления (например, поиск по институту)
const (
	Dean_FindByFaculty     = "dean_find_by_faculty"
	Dean_BackToFacultyMenu = "dean_back_to_faculty_menu"
)

type deanState struct {
	WaitFacultyName bool // ждём ввод названия факультета
}

var (
	deanMu   sync.RWMutex
	deanData = map[int64]deanState{}
)

func deanPeerFromCallback(upd *schemes.MessageCallbackUpdate) int64 {
	if upd.Message.Recipient.ChatId != 0 {
		return upd.Message.Recipient.ChatId
	}
	if upd.Message.Recipient.UserId != 0 {
		return upd.Message.Recipient.UserId
	}
	return 0
}
func deanPeerFromMessage(upd *schemes.MessageCreatedUpdate) int64 {
	if upd.Message.Recipient.ChatId != 0 {
		return upd.Message.Recipient.ChatId
	}
	return upd.Message.Sender.UserId
}
func deanSet(peer int64, st deanState) { deanMu.Lock(); deanData[peer] = st; deanMu.Unlock() }
func deanGet(peer int64) (deanState, bool) {
	deanMu.RLock()
	s, ok := deanData[peer]
	deanMu.RUnlock()
	return s, ok
}
func deanClear(peer int64) { deanMu.Lock(); delete(deanData, peer); deanMu.Unlock() }

// --- шаг 1: показать подсказку и включить ожидание ввода
// ТВОЮ Dean_ShowModeMenu переиспользуем как "попросить ввести факультет"
func Dean_ShowModeMenu(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	peer := deanPeerFromCallback(upd)
	deanSet(peer, deanState{WaitFacultyName: true})

	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Recipient.UserId)
	}
	msg.SetText("Введите название факультета (например, «ИУ»):")
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// --- шаг 2: обработать текст и вернуть расписание деканата факультета
func Dean_OnMessage(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate) (bool, error) {
	peer := deanPeerFromMessage(upd)
	st, ok := deanGet(peer)
	if !ok || !st.WaitFacultyName {
		return false, nil // не наш сценарий — пусть разбирают другие
	}

	query := strings.TrimSpace(upd.GetText())
	if query == "" {
		return true, deanReplyMsg(ctx, sc, upd, "Введите название факультета.")
	}

	// Находим факультеты по ILIKE
	var facs []models.Faculty
	if err := sc.DB.Where("name ILIKE ?", "%"+query+"%").
		Order("name").Limit(10).Find(&facs).Error; err != nil {
		return true, deanReplyMsg(ctx, sc, upd, fmt.Sprintf("Ошибка запроса: %v", err))
	}

	if len(facs) == 0 {
		return true, deanReplyMsg(ctx, sc, upd, "Факультеты не найдены. Попробуйте иначе.")
	}

	// Если найден один — показываем расписание
	if len(facs) == 1 {
		if err := deanShowSchedule(ctx, sc, upd, facs[0]); err != nil {
			return true, err
		}
		deanClear(peer)
		return true, nil
	}

	// Попробуем точное совпадение без регистра
	lq := strings.ToLower(query)
	for _, f := range facs {
		if strings.ToLower(f.Name) == lq {
			if err := deanShowSchedule(ctx, sc, upd, f); err != nil {
				return true, err
			}
			deanClear(peer)
			return true, nil
		}
	}

	// Слишком много совпадений
	var b strings.Builder
	b.WriteString("Нашлось несколько факультетов:\n")
	for i, f := range facs {
		fmt.Fprintf(&b, "%d) %s\n", i+1, f.Name)
	}
	b.WriteString("\nУточните название (например, полное наименование).")

	// остаёмся в состоянии WaitFacultyName
	return true, deanReplyMsg(ctx, sc, upd, b.String())
}

// показать расписание по факультету
func deanShowSchedule(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate, fac models.Faculty) error {
	var office models.DeanOffice
	err := sc.DB.Where("faculty_id = ?", fac.ID).First(&office).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return deanReplyMsg(ctx, sc, upd, fmt.Sprintf("Для факультета %q расписание не заполнено.", fac.Name))
		}
		return deanReplyMsg(ctx, sc, upd, fmt.Sprintf("Ошибка запроса расписания: %v", err))
	}

	text := deanFormat(fac, office)

	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Sender.UserId)
	}
	msg.SetText(text).AddKeyboard(deanScheduleKB(sc)) // ← добавили кнопки тут
	_, err = sc.API.Messages.Send(ctx, msg)
	return err
}

// формат ответа
func deanFormat(f models.Faculty, d models.DeanOffice) string {
	var b strings.Builder
	fmt.Fprintf(&b, "📅 Расписание деканата факультета %s\n\n", f.Name)
	if strings.TrimSpace(d.Schedule) != "" {
		fmt.Fprintf(&b, "%s\n\n", d.Schedule)
	} else {
		b.WriteString("Расписание не указано.\n\n")
	}
	if strings.TrimSpace(d.Contacts) != "" {
		fmt.Fprintf(&b, "Контакты: %s\n", d.Contacts)
	}
	if strings.TrimSpace(d.DocsLink) != "" {
		fmt.Fprintf(&b, "Документы/ссылки: %s\n", d.DocsLink)
	}
	return b.String()
}

// отправка сообщения из деканата
func deanReplyMsg(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate, text string) error {
	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Sender.UserId)
	}
	msg.SetText(text)
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

func deanScheduleKB(sc Ctx) *maxbot.Keyboard {
	kb := sc.API.Messages.NewKeyboardBuilder()
	kb.AddRow().
		AddCallback("◀️ К выбору факультета", schemes.POSITIVE, Dean_BackToFacultyMenu).
		AddCallback("🏠 Главное меню", schemes.NEGATIVE, "back_to_menu")
	return kb
}
