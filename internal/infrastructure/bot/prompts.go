package bot

// PromptEdit содержит промпты для редактирования
type PromptEdit struct {
	System string `json:"system"`
	User   string `json:"user"`
}

// Prompt содержит промпты для создания контента
type Prompt struct {
	System string     `json:"system"`
	User   string     `json:"user"`
	Edit   PromptEdit `json:"edit"`
}
