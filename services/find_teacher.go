package services

import (
	"context"
	"fmt"
	"strings"
	"sync"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"github.com/Karielka/Hackaton_MAX/models"
)

// Внутренние payload для подменю поиска
const (
	FT_FindByFaculty   = "find_by_faculty"
	FT_FindByDepartment= "find_by_department"
	FT_FindByFIO       = "find_by_fio"
)

// --- состояние диалога (по peerId) ---
type ftState struct{ Mode string } // faculty | department | fio

var (
	ftMu   sync.RWMutex
	ftData = map[int64]ftState{}
)

func ftPeerFromCallback(upd *schemes.MessageCallbackUpdate) int64 {
	if upd.Message.Recipient.ChatId != 0 {
		return upd.Message.Recipient.ChatId
	}
	if upd.Message.Recipient.UserId != 0 {
		return upd.Message.Recipient.UserId
	}
	return 0
}
func ftPeerFromMessage(upd *schemes.MessageCreatedUpdate) int64 {
	if upd.Message.Recipient.ChatId != 0 {
		return upd.Message.Recipient.ChatId
	}
	return upd.Message.Sender.UserId
}
func ftSet(peer int64, st ftState) { ftMu.Lock(); ftData[peer] = st; ftMu.Unlock() }
func ftGet(peer int64) (ftState, bool) { ftMu.RLock(); s, ok := ftData[peer]; ftMu.RUnlock(); return s, ok }
func ftClear(peer int64) { ftMu.Lock(); delete(ftData, peer); ftMu.Unlock() }

// --- UI подменю выбора режима поиска ---
func FT_ShowModeMenu(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	kb := sc.API.Messages.NewKeyboardBuilder()
	kb.AddRow().
		AddCallback("По факультету", schemes.POSITIVE,  FT_FindByFaculty).
		AddCallback("По кафедре",    schemes.POSITIVE,  FT_FindByDepartment)
	kb.AddRow().
		AddCallback("По ФИО",        schemes.POSITIVE, FT_FindByFIO)

	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Recipient.UserId)
	}
	msg.SetText("Как будем искать?").AddKeyboard(kb)
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// --- выбор режима и запрос ввода ---
func FT_AskForQuery(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	peer := ftPeerFromCallback(upd)

	mode := ""
	switch upd.Callback.Payload {
	case FT_FindByFaculty:    mode = "faculty"
	case FT_FindByDepartment: mode = "department"
	case FT_FindByFIO:        mode = "fio"
	}
	ftSet(peer, ftState{Mode: mode})

	prompt := map[string]string{
		"faculty":   "Введите название факультета:",
		"department":"Введите название кафедры:",
		"fio":       "Введите часть ФИО (например, «иванов»):",
	}[mode]

	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Recipient.UserId)
	}
	msg.SetText(prompt)
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}

// --- обработка пользовательского ввода из чата ---
func FT_OnMessage(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate) (bool, error) {
	peer := ftPeerFromMessage(upd)
	st, ok := ftGet(peer)
	if !ok || st.Mode == "" {
		return false, nil // это не наш сценарий
	}

	// правильно
	query := strings.TrimSpace(upd.GetText())

	if query == "" {
		return true, ftReplyMsg(ctx, sc, upd, "Введите текст запроса.")
	}

	// Готовим запрос в БД с нужными JOIN по цепочке N-1
	var res []models.Teacher
	q := sc.DB.Model(&models.Teacher{}).
		Preload("Department").
		Preload("Department.Faculty").
		Preload("Department.Faculty.Institute")

	switch st.Mode {
	case "faculty":
		// Teacher -> Department -> Faculty (по имени факультета)
		q = q.Joins("JOIN departments d ON d.id = teachers.department_id").
			Joins("JOIN faculties f ON f.id = d.faculty_id").
			Where("f.name ILIKE ?", "%"+query+"%")

	case "department":
		// Teacher -> Department (по имени кафедры)
		q = q.Joins("JOIN departments d ON d.id = teachers.department_id").
			Where("d.name ILIKE ?", "%"+query+"%")

	case "fio":
		q = q.Where("full_name ILIKE ?", "%"+query+"%")
	}

	if err := q.Limit(10).Find(&res).Error; err != nil {
		return true, ftReplyMsg(ctx, sc, upd, fmt.Sprintf("Ошибка поиска: %v", err))
	}
	if len(res) == 0 {
		_ = ftReplyMsg(ctx, sc, upd, "Совпадений не найдено. Попробуйте иначе.")
		return true, nil
	}

	var b strings.Builder
	b.WriteString("Найдено:\n")
	for _, t := range res {
		b.WriteString(ftFormatTeacher(t))
		b.WriteString("\n")
	}
	_ = ftReplyMsg(ctx, sc, upd, b.String())
	ftClear(peer)
	return true, nil
}

// ---- утилиты ответа/форматирования ----

func ftReplyMsg(ctx context.Context, sc Ctx, upd *schemes.MessageCreatedUpdate, text string) error {
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

func ftFormatTeacher(t models.Teacher) string {
	fac := "—"
	inst := "—"
	dep := "—"
	if t.Department.ID != 0 {
		dep = t.Department.Name
		if t.Department.Faculty.ID != 0 {
			fac = t.Department.Faculty.Name
			if t.Department.Faculty.Institute.ID != 0 {
				inst = t.Department.Faculty.Institute.Name
			}
		}
	}
	email := t.Email
	if strings.TrimSpace(email) == "" {
		email = "—"
	}
	sch := t.Schedule
	if strings.TrimSpace(sch) == "" {
		sch = "расписание не добавлено"
	}

	return fmt.Sprintf(
		"• %s\n  Институт: %s\n  Факультет: %s\n  Кафедра: %s\n  Почта: %s\n  %s",
		t.FullName, inst, fac, dep, email, ftFormatSchedule(sch),
	)
}

// отдельная функция — чтобы унифицировать печать расписания по проекту
func ftFormatSchedule(raw string) string {
	// позже заменишь на парсер/табличный вывод
	return "Расписание: " + raw
}
