package database

import (
	"database/sql"
	"fmt"
	"time"
)

type PostHistory struct {
	ID                     int        `json:"id"`
	UserID                 int64      `json:"user_id"`
	VoiceText              string     `json:"voice_text"`
	VoiceFileID            string     `json:"voice_file_id"`
	VoiceDuration          int        `json:"voice_duration"`
	VoiceFileSize          int        `json:"voice_file_size"`
	VoiceSentAt            time.Time  `json:"voice_sent_at"`
	VoiceReceivedAt        *time.Time `json:"voice_received_at"`
	AISentAt               *time.Time `json:"ai_sent_at"`
	AIReceivedAt           *time.Time `json:"ai_received_at"`
	AIResponse             string     `json:"ai_response"`
	AIModel                string     `json:"ai_model"`
	AITokensUsed           *int       `json:"ai_tokens_used"`
	AICost                 *float64   `json:"ai_cost"`
	IsSaved                bool       `json:"is_saved"`
	SavedAt                *time.Time `json:"saved_at"`
	ProcessingDurationMs   *int       `json:"processing_duration_ms"`
	WhisperDurationMs      *int       `json:"whisper_duration_ms"`
	AIGenerationDurationMs *int       `json:"ai_generation_duration_ms"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type PostHistoryRepository struct {
	db *sql.DB
}

func NewPostHistoryRepository(db *sql.DB) *PostHistoryRepository {
	return &PostHistoryRepository{db: db}
}

// CreatePostHistory создает новую запись в истории
func (r *PostHistoryRepository) CreatePostHistory(history *PostHistory) error {
	query := `
		INSERT INTO post_history (
			user_id, voice_text, voice_file_id, voice_duration, voice_file_size,
			voice_sent_at, voice_received_at, ai_sent_at, ai_received_at,
			ai_response, ai_model, ai_tokens_used, ai_cost, is_saved,
			processing_duration_ms, whisper_duration_ms, ai_generation_duration_ms
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17
		) RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		history.UserID, history.VoiceText, history.VoiceFileID, history.VoiceDuration, history.VoiceFileSize,
		history.VoiceSentAt, history.VoiceReceivedAt, history.AISentAt, history.AIReceivedAt,
		history.AIResponse, history.AIModel, history.AITokensUsed, history.AICost, history.IsSaved,
		history.ProcessingDurationMs, history.WhisperDurationMs, history.AIGenerationDurationMs,
	).Scan(&history.ID, &history.CreatedAt, &history.UpdatedAt)
}

// UpdatePostHistory обновляет существующую запись
func (r *PostHistoryRepository) UpdatePostHistory(history *PostHistory) error {
	query := `
		UPDATE post_history SET
			voice_text = $1, voice_received_at = $2, ai_sent_at = $3, ai_received_at = $4,
			ai_response = $5, ai_tokens_used = $6, ai_cost = $7,
			processing_duration_ms = $8, whisper_duration_ms = $9, ai_generation_duration_ms = $10
		WHERE id = $11`

	_, err := r.db.Exec(
		query,
		history.VoiceText, history.VoiceReceivedAt, history.AISentAt, history.AIReceivedAt,
		history.AIResponse, history.AITokensUsed, history.AICost,
		history.ProcessingDurationMs, history.WhisperDurationMs, history.AIGenerationDurationMs,
		history.ID,
	)
	return err
}

// UpdateVoiceTranscription обновляет только результаты транскрипции
func (r *PostHistoryRepository) UpdateVoiceTranscription(id int, voiceText string, voiceReceivedAt *time.Time, whisperDurationMs *int) error {
	query := `
		UPDATE post_history SET
			voice_text = $1, voice_received_at = $2, whisper_duration_ms = $3
		WHERE id = $4`

	_, err := r.db.Exec(query, voiceText, voiceReceivedAt, whisperDurationMs, id)
	return err
}

// UpdateAISentAt обновляет только время отправки в AI
func (r *PostHistoryRepository) UpdateAISentAt(id int, aiSentAt *time.Time) error {
	query := `UPDATE post_history SET ai_sent_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, aiSentAt, id)
	return err
}

// UpdateAIResponse обновляет только ответ AI
func (r *PostHistoryRepository) UpdateAIResponse(id int, aiResponse string, aiReceivedAt *time.Time, aiGenerationDurationMs *int, processingDurationMs *int) error {
	query := `
		UPDATE post_history SET
			ai_response = $1, ai_received_at = $2, ai_generation_duration_ms = $3, processing_duration_ms = $4
		WHERE id = $5`

	_, err := r.db.Exec(query, aiResponse, aiReceivedAt, aiGenerationDurationMs, processingDurationMs, id)
	return err
}

// MarkAsSaved отмечает пост как сохраненный
func (r *PostHistoryRepository) MarkAsSaved(id int) error {
	query := `UPDATE post_history SET is_saved = true, saved_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// GetUserPostHistory возвращает историю постов пользователя
func (r *PostHistoryRepository) GetUserPostHistory(userID int64, limit, offset int) ([]*PostHistory, error) {
	query := `
		SELECT id, user_id, voice_text, voice_file_id, voice_duration, voice_file_size,
			   voice_sent_at, voice_received_at, ai_sent_at, ai_received_at,
			   ai_response, ai_model, ai_tokens_used, ai_cost, is_saved, saved_at,
			   processing_duration_ms, whisper_duration_ms, ai_generation_duration_ms,
			   created_at, updated_at
		FROM post_history 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*PostHistory
	for rows.Next() {
		h := &PostHistory{}
		err := rows.Scan(
			&h.ID, &h.UserID, &h.VoiceText, &h.VoiceFileID, &h.VoiceDuration, &h.VoiceFileSize,
			&h.VoiceSentAt, &h.VoiceReceivedAt, &h.AISentAt, &h.AIReceivedAt,
			&h.AIResponse, &h.AIModel, &h.AITokensUsed, &h.AICost, &h.IsSaved, &h.SavedAt,
			&h.ProcessingDurationMs, &h.WhisperDurationMs, &h.AIGenerationDurationMs,
			&h.CreatedAt, &h.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// GetPostHistoryStats возвращает статистику по истории постов
func (r *PostHistoryRepository) GetPostHistoryStats(userID int64) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_posts,
			COUNT(CASE WHEN is_saved = true THEN 1 END) as saved_posts,
			AVG(processing_duration_ms) as avg_processing_time,
			AVG(whisper_duration_ms) as avg_whisper_time,
			AVG(ai_generation_duration_ms) as avg_ai_time,
			SUM(ai_tokens_used) as total_tokens,
			SUM(ai_cost) as total_cost
		FROM post_history 
		WHERE user_id = $1`

	var stats struct {
		TotalPosts        int      `db:"total_posts"`
		SavedPosts        int      `db:"saved_posts"`
		AvgProcessingTime *float64 `db:"avg_processing_time"`
		AvgWhisperTime    *float64 `db:"avg_whisper_time"`
		AvgAITime         *float64 `db:"avg_ai_time"`
		TotalTokens       *int     `db:"total_tokens"`
		TotalCost         *float64 `db:"total_cost"`
	}

	err := r.db.QueryRow(query, userID).Scan(
		&stats.TotalPosts, &stats.SavedPosts, &stats.AvgProcessingTime,
		&stats.AvgWhisperTime, &stats.AvgAITime, &stats.TotalTokens, &stats.TotalCost,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_posts":         stats.TotalPosts,
		"saved_posts":         stats.SavedPosts,
		"avg_processing_time": stats.AvgProcessingTime,
		"avg_whisper_time":    stats.AvgWhisperTime,
		"avg_ai_time":         stats.AvgAITime,
		"total_tokens":        stats.TotalTokens,
		"total_cost":          stats.TotalCost,
	}

	return result, nil
}

// UpdateProcessingDuration обновляет общее время обработки
func (r *PostHistoryRepository) UpdateProcessingDuration(id int, processingDurationMs int) error {
	query := `UPDATE post_history SET processing_duration_ms = $1 WHERE id = $2`
	_, err := r.db.Exec(query, processingDurationMs, id)
	return err
}

// GetPostHistoryByID возвращает запись истории по ID
func (r *PostHistoryRepository) GetPostHistoryByID(id int) (*PostHistory, error) {
	query := `
		SELECT id, user_id, voice_text, voice_file_id, voice_duration, voice_file_size,
			   voice_sent_at, voice_received_at, ai_sent_at, ai_received_at,
			   ai_response, ai_model, ai_tokens_used, ai_cost, is_saved, saved_at,
			   processing_duration_ms, whisper_duration_ms, ai_generation_duration_ms,
			   created_at, updated_at
		FROM post_history 
		WHERE id = $1`

	history := &PostHistory{}
	err := r.db.QueryRow(query, id).Scan(
		&history.ID, &history.UserID, &history.VoiceText, &history.VoiceFileID, &history.VoiceDuration, &history.VoiceFileSize,
		&history.VoiceSentAt, &history.VoiceReceivedAt, &history.AISentAt, &history.AIReceivedAt,
		&history.AIResponse, &history.AIModel, &history.AITokensUsed, &history.AICost, &history.IsSaved, &history.SavedAt,
		&history.ProcessingDurationMs, &history.WhisperDurationMs, &history.AIGenerationDurationMs,
		&history.CreatedAt, &history.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// AddVoiceToHistory добавляет голосовое сообщение к существующей записи истории
func (r *PostHistoryRepository) AddVoiceToHistory(id int, voiceText string, voiceDuration int, voiceFileSize int) error {
	// Получаем текущую запись
	history, err := r.GetPostHistoryByID(id)
	if err != nil {
		return fmt.Errorf("ошибка получения записи истории: %v", err)
	}

	// Конкатенируем текст
	newVoiceText := history.VoiceText
	if newVoiceText != "" {
		newVoiceText += "\n\n" + voiceText
	} else {
		newVoiceText = voiceText
	}

	// Суммируем длительность и размер
	newVoiceDuration := history.VoiceDuration + voiceDuration
	newVoiceFileSize := history.VoiceFileSize + voiceFileSize

	// Обновляем запись
	query := `
		UPDATE post_history SET
			voice_text = $1, voice_duration = $2, voice_file_size = $3
		WHERE id = $4`

	_, err = r.db.Exec(query, newVoiceText, newVoiceDuration, newVoiceFileSize, id)
	return err
}

// UpdateVoiceHistoryComplete обновляет запись истории с полной информацией о всех голосовых сообщениях
func (r *PostHistoryRepository) UpdateVoiceHistoryComplete(id int, combinedVoiceText string, totalDuration int, totalFileSize int) error {
	query := `
		UPDATE post_history SET
			voice_text = $1, voice_duration = $2, voice_file_size = $3
		WHERE id = $4`

	_, err := r.db.Exec(query, combinedVoiceText, totalDuration, totalFileSize, id)
	return err
}
