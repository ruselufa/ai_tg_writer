package bot

// Post представляет собой созданный пост
type Post struct {
	ContentType string   // тип контента (telegram_post, reels_script, youtube_script, instagram_post)
	Content     string   // текст поста
	Messages    []string // голосовые сообщения, использованные для создания поста
}

// UserState хранит состояние пользователя
type UserState struct {
	CurrentStep       string                         // текущий шаг (idle, selecting_content_type, waiting_for_voice, editing)
	ContentType       string                         // выбранный тип контента
	WaitingForVoice   bool                           // ожидание голосового сообщения
	VoiceMessages     []string                       // список транскрибированных голосовых сообщений
	EditMessages      []string                       // список транскрибированных голосовых сообщений для правок
	PostHistory       []Post                         // история постов
	CurrentPost       *Post                          // текущий пост (для редактирования)
	PendingVoices     map[string]*VoiceTranscription // голосовые сообщения в процессе транскрипции
	PendingEdits      map[string]*VoiceTranscription // голосовые сообщения для правок в процессе транскрипции
	ApprovalStatus    string                         // статус согласования (pending, approved, editing)
	LastGeneratedText string                         // последний сгенерированный текст для правок
}

type VoiceTranscription struct {
	MessageID int    // ID сообщения в Telegram
	FileID    string // ID файла голосового сообщения
	FilePath  string // Путь к скачанному файлу
	Status    string // статус транскрипции (pending, completed, error)
	Text      string // результат транскрипции
	Error     error  // ошибка, если есть
}

// StateManager управляет состояниями пользователей
type StateManager struct {
	states map[int64]*UserState
}

// NewStateManager создает новый менеджер состояний
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[int64]*UserState),
	}
}

// GetState возвращает состояние пользователя
func (sm *StateManager) GetState(userID int64) *UserState {
	state, exists := sm.states[userID]
	if !exists {
		state = &UserState{
			CurrentStep:       "idle",
			WaitingForVoice:   false,
			VoiceMessages:     make([]string, 0),
			EditMessages:      make([]string, 0),
			PostHistory:       make([]Post, 0),
			PendingVoices:     make(map[string]*VoiceTranscription),
			PendingEdits:      make(map[string]*VoiceTranscription),
			ApprovalStatus:    "pending",
			LastGeneratedText: "",
		}
		sm.states[userID] = state
	} else {
		if state.PendingVoices == nil {
			state.PendingVoices = make(map[string]*VoiceTranscription)
		}
		if state.PendingEdits == nil {
			state.PendingEdits = make(map[string]*VoiceTranscription)
		}
		if state.EditMessages == nil {
			state.EditMessages = make([]string, 0)
		}
	}
	return state
}

// UpdateStep обновляет текущий шаг пользователя
func (sm *StateManager) UpdateStep(userID int64, step string) {
	state := sm.GetState(userID)
	state.CurrentStep = step
}

// SetContentType устанавливает тип контента
func (sm *StateManager) SetContentType(userID int64, contentType string) {
	state := sm.GetState(userID)
	state.ContentType = contentType
}

// SetWaitingForVoice устанавливает флаг ожидания голосового сообщения
func (sm *StateManager) SetWaitingForVoice(userID int64, waiting bool) {
	state := sm.GetState(userID)
	state.WaitingForVoice = waiting
}

// AddVoiceMessage добавляет транскрибированное голосовое сообщение
func (sm *StateManager) AddVoiceMessage(userID int64, message string) {
	state := sm.GetState(userID)
	state.VoiceMessages = append(state.VoiceMessages, message)
}

// ClearVoiceMessages очищает список голосовых сообщений
func (sm *StateManager) ClearVoiceMessages(userID int64) {
	state := sm.GetState(userID)
	state.VoiceMessages = make([]string, 0)
}

// SavePost сохраняет пост в историю
func (sm *StateManager) SavePost(userID int64, post Post) {
	state := sm.GetState(userID)
	state.PostHistory = append(state.PostHistory, post)
}

// SetCurrentPost устанавливает текущий пост для редактирования
func (sm *StateManager) SetCurrentPost(userID int64, post *Post) {
	state := sm.GetState(userID)
	state.CurrentPost = post
}

// GetCurrentPost возвращает текущий пост для редактирования
func (sm *StateManager) GetCurrentPost(userID int64) *Post {
	state := sm.GetState(userID)
	return state.CurrentPost
}

// GetLastPost возвращает последний созданный пост
func (sm *StateManager) GetLastPost(userID int64) *Post {
	state := sm.GetState(userID)
	if len(state.PostHistory) == 0 {
		return nil
	}
	return &state.PostHistory[len(state.PostHistory)-1]
}

// AddPendingVoice добавляет голосовое сообщение в очередь на транскрипцию
func (sm *StateManager) AddPendingVoice(userID int64, messageID int, fileID string) {
	state := sm.GetState(userID)
	state.PendingVoices[fileID] = &VoiceTranscription{
		MessageID: messageID,
		FileID:    fileID,
		Status:    "pending",
	}
}

// UpdateVoiceTranscription обновляет статус транскрипции
func (sm *StateManager) UpdateVoiceTranscription(userID int64, fileID string, text string, err error) {
	state := sm.GetState(userID)
	if voice, ok := state.PendingVoices[fileID]; ok {
		if err != nil {
			voice.Status = "error"
			voice.Error = err
		} else {
			voice.Status = "completed"
			voice.Text = text
		}
	}
}

// GetPendingVoices возвращает все голосовые сообщения в обработке
func (sm *StateManager) GetPendingVoices(userID int64) map[string]*VoiceTranscription {
	return sm.GetState(userID).PendingVoices
}

// ClearPendingVoices очищает список обрабатываемых голосовых сообщений
func (sm *StateManager) ClearPendingVoices(userID int64) {
	state := sm.GetState(userID)
	state.PendingVoices = make(map[string]*VoiceTranscription)
}

// IsAllVoicesProcessed проверяет, все ли голосовые сообщения обработаны
func (sm *StateManager) IsAllVoicesProcessed(userID int64) bool {
	state := sm.GetState(userID)
	for _, voice := range state.PendingVoices {
		if voice.Status == "pending" {
			return false
		}
	}
	return true
}

// CollectVoiceResults собирает результаты всех транскрипций
func (sm *StateManager) CollectVoiceResults(userID int64) []string {
	state := sm.GetState(userID)
	results := make([]string, 0)
	for _, voice := range state.PendingVoices {
		if voice.Status == "completed" {
			results = append(results, voice.Text)
		}
	}
	return results
}

// AddToHistory добавляет пост в историю
func (sm *StateManager) AddToHistory(userID int64, post Post) {
	state := sm.GetState(userID)
	state.PostHistory = append(state.PostHistory, post)
}

// AddEditMessage добавляет транскрибированное голосовое сообщение для правок
func (sm *StateManager) AddEditMessage(userID int64, message string) {
	state := sm.GetState(userID)
	state.EditMessages = append(state.EditMessages, message)
}

// ClearEditMessages очищает список голосовых сообщений для правок
func (sm *StateManager) ClearEditMessages(userID int64) {
	state := sm.GetState(userID)
	state.EditMessages = make([]string, 0)
}

// AddPendingEdit добавляет голосовое сообщение для правок в очередь на транскрипцию
func (sm *StateManager) AddPendingEdit(userID int64, messageID int, fileID string) {
	state := sm.GetState(userID)
	state.PendingEdits[fileID] = &VoiceTranscription{
		MessageID: messageID,
		FileID:    fileID,
		Status:    "pending",
	}
}

// ClearPendingEdits очищает список обрабатываемых голосовых сообщений для правок
func (sm *StateManager) ClearPendingEdits(userID int64) {
	state := sm.GetState(userID)
	state.PendingEdits = make(map[string]*VoiceTranscription)
}

// SetApprovalStatus устанавливает статус согласования
func (sm *StateManager) SetApprovalStatus(userID int64, status string) {
	state := sm.GetState(userID)
	state.ApprovalStatus = status
}

// SetLastGeneratedText устанавливает последний сгенерированный текст
func (sm *StateManager) SetLastGeneratedText(userID int64, text string) {
	state := sm.GetState(userID)
	state.LastGeneratedText = text
}

// GetLastGeneratedText возвращает последний сгенерированный текст
func (sm *StateManager) GetLastGeneratedText(userID int64) string {
	state := sm.GetState(userID)
	return state.LastGeneratedText
}
