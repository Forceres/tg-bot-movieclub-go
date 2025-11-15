package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/model"
	"github.com/Forceres/tg-bot-movieclub-go/internal/service"
	"github.com/Forceres/tg-bot-movieclub-go/internal/tasks"
	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/kinopoisk"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type AddsMovieHandler struct {
	movieService     service.IMovieService
	kinopoiskService service.IKinopoiskService
	sessionService   service.ISessionService
	pollService      service.IPollService
	asynqClient      *asynq.Client
	inspector        *asynq.Inspector
	redisUnavailable bool
}

func NewAddsMovieHandler(
	movieService service.IMovieService,
	kinopoiskService service.IKinopoiskService,
	sessionService service.ISessionService,
	pollService service.IPollService,
	asynqClient *asynq.Client,
	inspector *asynq.Inspector,
) *AddsMovieHandler {
	return &AddsMovieHandler{
		movieService:     movieService,
		kinopoiskService: kinopoiskService,
		sessionService:   sessionService,
		pollService:      pollService,
		asynqClient:      asynqClient,
		inspector:        inspector,
	}
}

func (h *AddsMovieHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	rawPayload := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/adds"))
	if rawPayload == "" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Пожалуйста, отправьте ссылки или идентификаторы фильмов после команды /adds.",
		})
		return
	}

	movieIDs, invalidTokens := parseMovieIDs(rawPayload)
	if len(movieIDs) == 0 {
		text := "Не удалось найти ID фильмов. Убедитесь, что вы отправили корректные ссылки на Кинопоиск или числовые идентификаторы."
		if len(invalidTokens) > 0 {
			text = fmt.Sprintf("%s Невалидные значения: %s", text, strings.Join(invalidTokens, ", "))
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   text,
		})
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
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ошибка при проверке фильмов в базе данных.",
			})
			return
		}
		existingIDs = append(existingIDs, id)
	}

	var createdIDs []int64
	if len(lookupIDs) > 0 {
		moviesDTO, err := h.kinopoiskService.SearchMovies(lookupIDs, update.Message.From.FirstName)
		if err != nil {
			log.Printf("kinopoisk search failed: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ошибка при запросе фильмов в Кинопоиске.",
			})
			return
		}
		if len(moviesDTO) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось найти фильмы по указанным ссылкам.",
			})
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
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось сохранить новые фильмы. Попробуйте снова позже.",
			})
			return
		}
	}

	targetIDs := append(existingIDs, createdIDs...)
	if len(targetIDs) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Фильмы уже находятся в сессии или не удалось их обработать.",
		})
		return
	}

	finishSet, finishOk := h.activeTaskSet("before /adds", tasks.FinishSessionTaskType)
	closeSet, closeOk := h.activeTaskSet("before /adds", tasks.CloseRatingVotingTaskType)

	session, newSessionMovieIDs, sessionCreated, err := h.sessionService.AddMoviesToSession(update.Message.From.ID, targetIDs)
	if err != nil {
		log.Printf("failed to add movies to session: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не удалось добавить фильмы в сессию.",
		})
		return
	}

	h.scheduleFinishSessionIfNeeded(session, finishSet, sessionCreated, finishOk)
	h.scheduleCloseVotingTasks(update, newSessionMovieIDs, closeSet, closeOk)

	if !h.redisUnavailable {
		h.activeTaskSet("after /adds", tasks.FinishSessionTaskType)
		h.activeTaskSet("after /adds", tasks.CloseRatingVotingTaskType)
	} else {
		log.Printf("after /adds redis dependent tasks skipped because redis is unavailable")
	}

	response := buildSuccessMessage(session.ID, existingIDs, createdIDs)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   response,
	})
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

func buildSuccessMessage(sessionID int64, existing, created []int64) string {
	var parts []string
	if len(created) > 0 {
		parts = append(parts, fmt.Sprintf("Добавлены новые фильмы: %v", created))
	}
	if len(existing) > 0 {
		parts = append(parts, fmt.Sprintf("Добавлены ранее существовавшие фильмы: %v", existing))
	}
	parts = append(parts, fmt.Sprintf("Сессия ID: %d", sessionID))
	return strings.Join(parts, "\n")
}

func (h *AddsMovieHandler) activeTaskSet(prefix, taskType string) (map[string]struct{}, bool) {
	if h.redisUnavailable {
		log.Printf("%s %s tasks: skipped because redis is unavailable", prefix, taskType)
		return nil, false
	}
	if h.inspector == nil {
		log.Printf("%s %s tasks: inspector is not configured", prefix, taskType)
		return nil, false
	}
	activeTasks, err := h.inspector.ListActiveTasks(taskType)
	if err != nil {
		log.Printf("%s failed to list %s tasks: %v", prefix, taskType, err)
		h.redisUnavailable = true
		return nil, false
	}
	ids := make([]string, 0, len(activeTasks))
	set := make(map[string]struct{}, len(activeTasks))
	for _, task := range activeTasks {
		id := strings.TrimSpace(task.ID)
		if id == "" {
			continue
		}
		ids = append(ids, id)
		set[id] = struct{}{}
	}
	log.Printf("%s %s tasks: %v", prefix, taskType, ids)
	return set, true
}

func (h *AddsMovieHandler) scheduleFinishSessionIfNeeded(session *model.Session, active map[string]struct{}, created bool, ok bool) {
	if h.asynqClient == nil || session == nil || h.redisUnavailable || !ok {
		return
	}
	if session.FinishedAt == 0 {
		log.Printf("session %d has no finished_at, skipping finish task scheduling", session.ID)
		return
	}
	taskID := fmt.Sprint(session.ID)
	if _, exists := active[taskID]; exists && !created {
		return
	}
	duration := time.Until(time.Unix(session.FinishedAt, 0))
	if duration <= 0 {
		duration = time.Second
	}
	if err := tasks.EnqueueFinishSessionTask(h.asynqClient, &tasks.EnqueueFinishSessionParams{
		SessionID: session.ID,
		Duration:  duration,
	}); err != nil {
		log.Printf("failed to enqueue finish_session task for session %d: %v", session.ID, err)
		h.redisUnavailable = true
	}
}

func (h *AddsMovieHandler) scheduleCloseVotingTasks(update *models.Update, movieIDs []int64, active map[string]struct{}, ok bool) {
	if h.asynqClient == nil || h.pollService == nil || update.Message == nil || len(movieIDs) == 0 || h.redisUnavailable || !ok {
		return
	}
	for _, movieID := range movieIDs {
		poll, err := h.pollService.GetOpenedPollByMovieID(movieID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("failed to fetch opened poll for movie %d: %v", movieID, err)
			}
			continue
		}
		if poll.Voting.ID == 0 || poll.Voting.FinishedAt == nil {
			log.Printf("poll %s for movie %d lacks voting finish time, skipping", poll.PollID, movieID)
			continue
		}
		taskID := fmt.Sprint(poll.VotingID)
		if _, exists := active[taskID]; exists {
			continue
		}
		duration := time.Until(time.Unix(*poll.Voting.FinishedAt, 0))
		if duration <= 0 {
			duration = time.Second
		}
		payload := &tasks.CloseRatingVotingPayload{
			PollID:    poll.PollID,
			MessageID: poll.MessageID,
			ChatID:    update.Message.Chat.ID,
			VotingID:  poll.VotingID,
			MovieID:   movieID,
			UserID:    update.Message.From.ID,
		}
		if err := tasks.EnqueueCloseRatingVotingTask(h.asynqClient, duration, payload); err != nil {
			log.Printf("failed to enqueue close_rating_voting task for movie %d: %v", movieID, err)
			h.redisUnavailable = true
		}
	}
}
