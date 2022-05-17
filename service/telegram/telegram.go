package telegram

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const defaultParseMode = tgbotapi.ModeHTML

// Telegram struct holds necessary data to communicate with the Telegram API.
type Telegram struct {
	client  *tgbotapi.BotAPI
	chatIDs []int64
}

// New returns a new instance of a Telegram notification service.
// For more information about telegram api token:
//    -> https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api#NewBotAPI
func New(apiToken string) (*Telegram, error) {
	client, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, err
	}

	t := &Telegram{
		client:  client,
		chatIDs: []int64{},
	}

	return t, nil
}

// AddReceivers takes Telegram chat IDs and adds them to the internal chat ID list. The Send method will send
// a given message to all those chats.
func (t *Telegram) AddReceivers(chatIDs ...int64) {
	t.chatIDs = append(t.chatIDs, chatIDs...)
}

func (t *Telegram) Client() *tgbotapi.BotAPI {
	return t.client
}

// GetChatId returns the chat id of the given username if the user sends a message to the bot from call until timeout. Else it returns (-1,error)
func (t *Telegram) GetChatId(username string, timeout time.Duration) (int, error) {

	interruptSignal := time.After(timeout)
	ticker := time.NewTicker(time.Second * 2)
	// Keep trying until we're timed out or get a result/error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-interruptSignal:
			return -1, errors.New("timed out")
		// Got a tick, we should check on checkSomething()
		case <-ticker.C:
			updates, err := t.client.GetUpdates(tgbotapi.UpdateConfig{
				Timeout: int(time.Millisecond * 500),
			})
			if err != nil {
				continue
			}
			for _, update := range updates {
				if update.Message != nil && update.Message.From != nil {
					if update.Message.From.UserName == username {
						return update.Message.From.ID, nil
					}
				}
			}
		}
	}
}

// Send takes a message subject and a message body and sends them to all previously set chats. Message body supports
// html as markup language.
func (t Telegram) Send(ctx context.Context, subject, message string) error {
	fullMessage := subject + "\n" + message // Treating subject as message title

	msg := tgbotapi.NewMessage(0, fullMessage)
	msg.ParseMode = defaultParseMode

	for _, chatID := range t.chatIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg.ChatID = chatID
			_, err := t.client.Send(msg)
			if err != nil {
				return errors.Wrapf(err, "failed to send message to Telegram chat '%d'", chatID)
			}
		}
	}

	return nil
}
