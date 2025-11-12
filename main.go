package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/configservice"
	"github.com/max-messenger/max-bot-api-client-go/schemes"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gorm.io/gorm"

	intdb "github.com/Karielka/Hackaton_MAX/internal/db"
	"github.com/Karielka/Hackaton_MAX/models"
	"github.com/Karielka/Hackaton_MAX/services"
)

func main() {
	// Логи
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).With().Timestamp().Caller().Logger()

	// 1) MAX: читаем конфиг и создаём клиента
	configPath := "./config/config.yaml"
	cfg := configservice.NewConfigInterface(configPath)
	if cfg == nil {
		log.Fatal().Str("configPath", configPath).Msg("NewConfigInterface failed. Stop.")
	}
	api, err := maxbot.NewWithConfig(cfg) // тип клиента из SDK: *maxbot.Api
	if err != nil {
		log.Fatal().Err(err).Msg("NewWithConfig failed. Stop.")
	}

	// 2) БД (GORM + Postgres)
	db := intdb.Connect()
	runMigrations(db)

	// 3) Контекст с graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, os.Interrupt)
		<-ch
		cancel()
	}()

	// 4) Команды бота (опционально)
	_, _ = api.Bots.PatchBot(ctx, &schemes.BotPatch{
		Commands: []schemes.BotCommand{
			{Name: "start", Description: "Показать меню"},
			{Name: "menu", Description: "Показать меню"},
		},
	})

	log.Info().Msg("Bot is up. Waiting for updates...")

	// 5) Главный цикл апдейтов
	for upd := range api.GetUpdates(ctx) {
		// полезно в отладке, можно выключить
		api.Debugs.Send(ctx, upd)

		switch upd := upd.(type) {
		case *schemes.MessageCreatedUpdate:
			handleMessage(ctx, api, db, upd)

		case *schemes.MessageCallbackUpdate:
			// маршрутизация в сервисы
			sc := services.Ctx{API: api, DB: db}
			if err := services.Route(ctx, sc, upd); err != nil {
				log.Err(err).Msg("services.Route")
			}

		default:
			log.Debug().Msgf("Skip update type: %T", upd)
		}
	}
}

// /start, /menu, любое сообщение — показываем меню
func handleMessage(ctx context.Context, api *maxbot.Api, db *gorm.DB, upd *schemes.MessageCreatedUpdate) {
	msg := maxbot.NewMessage()
	if upd.Message.Recipient.ChatId != 0 {
		msg.SetChat(upd.Message.Recipient.ChatId)
	} else {
		msg.SetUser(upd.Message.Sender.UserId)
	}

	// Одинаково для /start, /menu и прочих текстов — показываем меню
	msg.SetText(services.WelcomeText())
	// ВАЖНО: services.MenuKeyboard(api) должен возвращать schemes.Keyboard (НЕ указатель)
	kb := services.MenuKeyboard(api)
	msg.AddKeyboard(kb)

	if _, err := api.Messages.Send(ctx, msg); err != nil {
		log.Err(err).Msg("send menu")
	}
}

func runMigrations(db *gorm.DB) {
	if err := models.AutoMigrate(db); err != nil {
		log.Fatal().Err(err).Msg("AutoMigrate failed")
	}
	// пример сидирования
	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = seed(db)
	}()
}

func seed(db *gorm.DB) error {
	// добавь сюда первичные данные при необходимости
	return nil
}
