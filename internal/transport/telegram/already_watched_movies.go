package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/celestix/telegraph-go/v2"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type AlreadyWatchedMoviesHandler struct {
	movieService service.IMovieService
	telegraph *telegraph.TelegraphClient
	chatData map[int64]*ChatData
}

type ChatData struct {
	LastPageURL  string
	LastPage string
	MessageID int
	Links []string
}

type IAlreadyWatchedMoviesHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewAlreadyWatchedMoviesHandler(movieService service.IMovieService,
	telegraph *telegraph.TelegraphClient) *AlreadyWatchedMoviesHandler {
	return &AlreadyWatchedMoviesHandler{movieService: movieService, telegraph: telegraph, chatData: make(map[int64]*ChatData)}
}

func (h *AlreadyWatchedMoviesHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	// chat, err := b.GetChat(ctx, &bot.GetChatParams{ChatID: update.Message.Chat.ID})
	
	// movies, err := h.movieService.GetAlreadyWatchedMovies()
	// if err != nil {
	// 	b.SendMessage(ctx, &bot.SendMessageParams{})
	// 	return
	// }
}

func (h *AlreadyWatchedMoviesHandler) addNewPage(movies []model.Movie) {}

func (h *AlreadyWatchedMoviesHandler) handleUpdatePage(movies []model.Movie) {}

func (h *AlreadyWatchedMoviesHandler) handleNewPages(movies []model.Movie) {}

func (h *AlreadyWatchedMoviesHandler) createInitialMessage(movies []model.Movie) {}

func (h *AlreadyWatchedMoviesHandler) getChatData(chatID int64) *ChatData {
    if h.chatData[chatID] == nil {
        h.chatData[chatID] = &ChatData{}
    }
    return h.chatData[chatID]
}

func (h *AlreadyWatchedMoviesHandler) createTelegraphPages(pagesData []string) []string {
    var newPages []string
    for i, pageData := range pagesData {
        page, err := h.telegraph.CreatePage(
					  "", // access token
            "Список просмотренных фильмов",
            pageData,
            &telegraph.PageOpts{
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

func (h *AlreadyWatchedMoviesHandler) formatLinks(urls []string) []string {
    var links []string
    for idx, url := range urls {
        links = append(links, fmt.Sprintf("%d. %s", idx+1, url))
    }
    return links
}

func (h *AlreadyWatchedMoviesHandler) updateLinks(oldLinks []string, newURL string) []string {
    links := oldLinks
    if len(links) > 0 {
        links = links[:len(links)-1]
    }
    return append(links, fmt.Sprintf("%d. %s", len(links)+1, newURL))
}

func (h *AlreadyWatchedMoviesHandler) addTimestamp(links []string) []string {
    timestamp := time.Now().UTC().Format("01-02-2006 15:04:05")
    return append(links, timestamp)
}

func (h *AlreadyWatchedMoviesHandler) generateHTML(movies []model.Movie) ([]string, error) {
    var pages []string
    var html strings.Builder
    
    for i, movie := range movies {
        html.WriteString(fmt.Sprintf("<p>%d. %s (%d)</p>\n", i+1, movie.Title, movie.Year))
        
        if (i+1)%50 == 0 || i == len(movies)-1 {
            pages = append(pages, html.String())
            html.Reset()
        }
    }
    
    return pages, nil
}