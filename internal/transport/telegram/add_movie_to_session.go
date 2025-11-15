package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/kinopoisk"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type AddMovieToSessionHandler struct {
	movieService     service.IMovieService
	kinopoiskService service.IKinopoiskService
	sessionService   service.ISessionService
	pollService      service.IPollService
	asynqClient      *asynq.Client
	inspector        *asynq.Inspector
}

type IAddMovieToSessionHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

func NewAddMovieToSessionHandler(
	movieService service.IMovieService,
	kinopoiskService service.IKinopoiskService,
	sessionService service.ISessionService,
	pollService service.IPollService,
	asynqClient *asynq.Client,
	inspector *asynq.Inspector,
) IAddMovieToSessionHandler {
	return &AddMovieToSessionHandler{
		movieService:     movieService,
		kinopoiskService: kinopoiskService,
		sessionService:   sessionService,
		pollService:      pollService,
		asynqClient:      asynqClient,
		inspector:        inspector,
	}
}

func (h *AddMovieToSessionHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	rawPayload := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/adds"))
	if rawPayload == "" {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "📝 Пожалуйста, отправьте ссылки или идентификаторы фильмов после команды /adds.",
		})
		if err != nil {
			log.Printf("failed to send error message: %v", err)
		}
		return
	}

	movieIDs, invalidTokens := parseMovieIDs(rawPayload)
	if len(movieIDs) == 0 {
		text := "❌ Не удалось найти ID фильмов. Убедитесь, что вы отправили корректные ссылки на Кинопоиск или числовые идентификаторы."
		if len(invalidTokens) > 0 {
			text = fmt.Sprintf("%s\n⚠️ Невалидные значения: %s", text, strings.Join(invalidTokens, ", "))
		}
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   text,
		})
		if err != nil {
			log.Printf("failed to send error message: %v", err)
		}
		return
	}

	var existingIDs []int64
	var lookupIDs []int64
	for _, id := range movieIDs {
		_, err := h.movieService.GetMovieByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				lookupIDs = append(lookupIDs, id)
				continue
			}
			log.Printf("failed to get movie %d: %v", id, err)
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "❌ Ошибка при проверке фильмов в базе данных.",
			})
			if err != nil {
				log.Printf("failed to send error message: %v", err)
			}
			return
		}
		existingIDs = append(existingIDs, id)
	}

	var createdIDs []int64
	if len(lookupIDs) > 0 {
		moviesDTO, err := h.kinopoiskService.SearchMovies(lookupIDs, update.Message.From.FirstName)
		if err != nil {
			log.Printf("kinopoisk search failed: %v", err)
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "❌ Ошибка при запросе фильмов в Кинопоиске.",
			})
			if err != nil {
				log.Printf("failed to send error message: %v", err)
			}
			return
		}
		if len(moviesDTO) == 0 {
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "🔍 Не удалось найти фильмы по указанным ссылкам.",
			})
			if err != nil {
				log.Printf("failed to send error message: %v", err)
			}
			return
		}
		for _, movieDTO := range moviesDTO {
			if err := h.movieService.Create(&movieDTO, update.Message.From.ID); err != nil {
				log.Printf("failed to create movie %d: %v", movieDTO.KinopoiskID, err)
				continue
			}
			createdIDs = append(createdIDs, movieDTO.KinopoiskID)
		}
		if len(createdIDs) == 0 {
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "❌ Не удалось сохранить новые фильмы. Попробуйте снова позже.",
			})
			if err != nil {
				log.Printf("failed to send error message: %v", err)
			}
			return
		}
	}

	targetIDs := append(existingIDs, createdIDs...)
	if len(targetIDs) == 0 {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "ℹ️ Фильмы уже находятся в сессии или не удалось их обработать.",
		})
		if err != nil {
			log.Printf("failed to send error message: %v", err)
		}
		return
	}

	session, newSessionMovieIDs, sessionCreated, err := h.sessionService.AddMoviesToSession(update.Message.From.ID, targetIDs)
	if err != nil {
		log.Printf("failed to add movies to session: %v", err)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Не удалось добавить фильмы в сессию.",
		})
		if err != nil {
			log.Printf("failed to send error message: %v", err)
		}
		return
	}

	if sessionCreated && session.FinishedAt > 0 {
		finishTime := time.Unix(session.FinishedAt, 0)
		duration := time.Until(finishTime)

		if duration > 0 {
			err := tasks.EnqueueFinishSessionTask(h.asynqClient, &tasks.EnqueueFinishSessionParams{
				SessionID: session.ID,
				Duration:  duration,
			})
			if err != nil {
				log.Printf("failed to enqueue finish session task: %v", err)
			} else {
				log.Printf("Scheduled finish session task for session %d at %s", session.ID, finishTime.Format(time.RFC3339))
			}
		}
	}

	if len(newSessionMovieIDs) > 0 && session.FinishedAt > 0 {
		finishTime := time.Unix(session.FinishedAt, 0)
		duration := time.Until(finishTime)

		if duration > 0 {
			for _, movieID := range newSessionMovieIDs {
				movie, err := h.movieService.GetMovieByID(movieID)
				if err != nil {
					log.Printf("failed to get movie %d for rating task: %v", movieID, err)
					continue
				}

				taskID := fmt.Sprintf("%s-%d-%d", tasks.OpenRatingVotingTaskType, session.ID, movieID)

				taskExists := false
				if h.inspector != nil {
					_, err := h.inspector.GetTaskInfo(tasks.QUEUE, taskID)
					if err == nil {
						taskExists = true
						log.Printf("Rating voting task already exists for movie %d in session %d", movieID, session.ID)
					}
				}

				if !taskExists {
					err = tasks.EnqueueOpenRatingVotingTask(h.asynqClient, &tasks.EnqueueOpenRatingVotingParams{
						SessionID: session.ID,
						ChatID:    update.Message.Chat.ID,
						Movie:     *movie,
						UserID:    update.Message.From.ID,
						TaskID:    taskID,
						Duration:  duration,
					})
					if err != nil {
						log.Printf("failed to enqueue open rating voting task for movie %d: %v", movieID, err)
					} else {
						log.Printf("Scheduled open rating voting task for movie %d in session %d at %s",
							movieID, session.ID, finishTime.Format(time.RFC3339))
					}
				}
			}
		}
	}

	var responseText string
	if sessionCreated {
		responseText = fmt.Sprintf("✅ Создана новая сессия с %d фильмами.\n", len(targetIDs))
	} else {
		responseText = fmt.Sprintf("✅ Добавлено %d новых фильмов в текущую сессию.\n", len(newSessionMovieIDs))
	}

	if len(existingIDs) > 0 {
		responseText += fmt.Sprintf("ℹ️ %d фильмов уже были в базе.\n", len(existingIDs))
	}
	if len(createdIDs) > 0 {
		responseText += fmt.Sprintf("🆕 %d новых фильмов добавлено в базу.\n", len(createdIDs))
	}

	if session.FinishedAt > 0 {
		finishTime := time.Unix(session.FinishedAt, 0)
		responseText += fmt.Sprintf("\n📅 Дата окончания просмотра: %s", finishTime.Format("02.01.2006 15:04"))
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   responseText,
	})
	if err != nil {
		log.Printf("failed to send adds movie response: %v", err)
	}
}

func parseMovieIDs(raw string) ([]int64, []string) {
	candidates := kinopoisk.ParseIDsOrRefs(raw)
	if len(candidates) == 0 {
		for _, token := range strings.Fields(raw) {
			token = strings.Trim(token, ",;\"'")
			if token != "" {
				candidates = append(candidates, token)
			}
		}
	}
	seen := make(map[int64]struct{})
	var ids []int64
	var invalid []string
	for _, candidate := range candidates {
		id, err := strconv.ParseInt(candidate, 10, 64)
		if err != nil {
			invalid = append(invalid, candidate)
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, invalid
}
