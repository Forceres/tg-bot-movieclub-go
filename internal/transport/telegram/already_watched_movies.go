package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegraph"
	telegraphv2 "github.com/celestix/telegraph-go/v2"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type AlreadyWatchedMoviesHandler struct {
	movieService service.IMovieService
	telegraph    *telegraph.Telegraph
	chatData     map[int64]*ChatData
}

type ChatData struct {
	LastPageURL string
	LastPage    string
	MessageID   int
	Links       []string
}

type IAlreadyWatchedMoviesHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewAlreadyWatchedMoviesHandler(movieService service.IMovieService,
	telegraph *telegraph.Telegraph) *AlreadyWatchedMoviesHandler {
	return &AlreadyWatchedMoviesHandler{movieService: movieService, telegraph: telegraph, chatData: make(map[int64]*ChatData)}
}

func (h *AlreadyWatchedMoviesHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	data := h.getChatData(chatID)

	formattedMovies, err := h.movieService.GetAlreadyWatchedMovies()
	if err != nil {
		log.Printf("Error retrieving watched movies: %v", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Ошибка при получении списка фильмов",
		})
		if err != nil {
			log.Printf("Error sending error message: %v", err)
		}
		return
	}
	if len(formattedMovies) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "📋 Список просмотренных фильмов пуст.",
		})
		if err != nil {
			log.Printf("Error sending empty list message: %v", err)
		}
		return
	}

	if data.LastPageURL == "" {
		h.handleNewPages(ctx, b, update, data, formattedMovies)
		return
	}

	h.handleUpdatePage(ctx, b, update, data)
}

func (h *AlreadyWatchedMoviesHandler) addNewPage(ctx context.Context, b *bot.Bot, update *models.Update, data *ChatData, pages []string) {
	newPage, err := h.telegraph.Client.CreatePage(
		h.telegraph.Account.AccessToken,
		"Список просмотренных фильмов",
		pages[len(pages)-1],
		&telegraphv2.PageOpts{
			AuthorName: "КиноКлассБот",
		},
	)
	if err != nil {
		log.Printf("Error creating telegraph page: %v", err)
		return
	}

	links := append(data.Links, fmt.Sprintf("%d. %s", len(data.Links)+1, newPage.Url))
	data.LastPageURL = newPage.Url
	data.Links = links

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.Message.Chat.ID,
		Text:      strings.Join(links, "\n"),
		MessageID: data.MessageID,
	})
	if err != nil {
		log.Printf("Error editing message text: %v", err)
	}
}

func (h *AlreadyWatchedMoviesHandler) handleUpdatePage(ctx context.Context, b *bot.Bot, update *models.Update, data *ChatData) {
	log.Println("Handle update page...")
	urlParts := strings.Split(data.LastPageURL, "/")
	path := urlParts[len(urlParts)-1]

	page, err := h.telegraph.Client.EditPage(h.telegraph.Account.AccessToken, path, "Список просмотренных фильмов", data.LastPage, &telegraphv2.PageOpts{
		AuthorName:    "КиноКлассБот",
		ReturnContent: true,
	})
	if err != nil {
		log.Printf("Error editing telegraph page: %v, retrying...", err)
		data.LastPageURL = ""
		h.Handle(ctx, b, update)
		return
	}

	links := data.Links
	if len(links) > 0 {
		links = links[:len(links)-1]
	}
	links = append(links, fmt.Sprintf("%d. %s", len(links)+1, page.Url))
	data.Links = links
	timestamp := time.Now().UTC().Format("01-02-2006 15:04:05")
	linksWithTime := append(links, timestamp)

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.Message.Chat.ID,
		Text:      strings.Join(linksWithTime, "\n"),
		MessageID: data.MessageID,
	})
	if err != nil {
		log.Printf("Error editing message text: %v", err)
	}
}

func (h *AlreadyWatchedMoviesHandler) handleNewPages(ctx context.Context, b *bot.Bot, update *models.Update, data *ChatData, pages []string) {
	log.Println("Handle new pages...")
	if data.MessageID == 0 {
		log.Println("Handle initial message...")
		h.createInitialMessage(ctx, b, update, data, pages)
		return
	}
	h.addNewPage(ctx, b, update, data, pages)
}

func (h *AlreadyWatchedMoviesHandler) createInitialMessage(ctx context.Context, b *bot.Bot, update *models.Update, data *ChatData, pages []string) {
	newPages := h.createTelegraphPages(pages)
	if len(newPages) == 0 {
		return
	}

	data.LastPage = pages[len(pages)-1]
	data.LastPageURL = newPages[len(newPages)-1]
	var links []string
	for idx, url := range newPages {
		links = append(links, fmt.Sprintf("%d. %s", idx+1, url))
	}

	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   strings.Join(links, "\n"),
	})
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}

	data.MessageID = msg.ID
	data.Links = links

	_, err = b.PinChatMessage(ctx, &bot.PinChatMessageParams{
		ChatID:              update.Message.Chat.ID,
		MessageID:           data.MessageID,
		DisableNotification: true,
	})
	if err != nil {
		log.Printf("Error pinning message: %v", err)
	}
}

func (h *AlreadyWatchedMoviesHandler) getChatData(chatID int64) *ChatData {
	if h.chatData[chatID] == nil {
		h.chatData[chatID] = &ChatData{}
	}
	return h.chatData[chatID]
}

func (h *AlreadyWatchedMoviesHandler) createTelegraphPages(pagesData []string) []string {
	var newPages []string
	for i, pageData := range pagesData {
		page, err := h.telegraph.Client.CreatePage(
			h.telegraph.Account.AccessToken,
			"Список просмотренных фильмов",
			pageData,
			&telegraphv2.PageOpts{
				AuthorName: "КиноКлассБот",
			},
		)
		if err != nil {
			log.Printf("Error creating telegraph page %d: %v", i, err)
			continue
		}
		newPages = append(newPages, page.Url)
	}
	return newPages
}
