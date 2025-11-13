package services

import (
	"context"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func FAQ_Handle(ctx context.Context, sc Ctx, upd *schemes.MessageCallbackUpdate) error {
	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Recipient.UserId)
	}
	str := `**Часто задаваемые вопросы**  

	1) Как найти преподавателя? 
	Нажмите кнопку «Поиск преподавателя» в главном меню.  
	Затем выберите способ поиска: *по кафедре, факультету или ФИО*.  
	После ввода данных бот покажет контактную информацию и расписание преподавателя.  

	2) Как узнать расписание занятий?  
	Для преподавателей расписание выводится автоматически при поиске.  
	Для студентов — функция появится в ближайших обновлениях.  

	3) Как написать преподавателю?
	После поиска преподавателя вы получите его e-mail.  
	Можно кликнуть по нему или скопировать адрес и отправить письмо напрямую.  

	4) Что делать, если не нашёл нужного преподавателя? 
	Попробуйте ввести фамилию без сокращений и уточните кафедру.  
	Если результата нет — возможно, данные ещё не внесены в базу.  

	5) К кому обратиться при технических проблемах с ботом?
	Если бот не отвечает или выдаёт ошибку — напишите администратору:  
	karielka@yandex.ru

	Спасибо, что используете нашего бота! `

	msg.SetText(str)
	msg.SetFormat("markdown")
	_, err := sc.API.Messages.Send(ctx, msg)
	return err
}
