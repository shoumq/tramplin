package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"tramplin/internal/models"
)

var (
	ErrNotConfigured = errors.New("yandex ai is not configured")
	ErrProvider      = errors.New("yandex ai provider error")
)

type Config struct {
	FolderID        string
	APIKey          string
	Model           string
	Endpoint        string
	MaxOutputTokens int
	Temperature     float64
	Timeout         time.Duration
}

type Client struct {
	config     Config
	httpClient *http.Client
}

type responseRequest struct {
	Model           string  `json:"model"`
	Temperature     float64 `json:"temperature"`
	Instructions    string  `json:"instructions"`
	Input           string  `json:"input"`
	MaxOutputTokens int     `json:"max_output_tokens"`
}

type responseData struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func New(config Config) *Client {
	if config.Model == "" {
		config.Model = "deepseek-v4-flash/latest"
	}
	if config.Endpoint == "" {
		config.Endpoint = "https://ai.api.cloud.yandex.net/v1/responses"
	}
	if config.MaxOutputTokens == 0 {
		config.MaxOutputTokens = 2000
	}
	if config.Temperature == 0 {
		config.Temperature = 0.3
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &Client{
		config:     config,
		httpClient: &http.Client{Timeout: config.Timeout},
	}
}

func (c *Client) IsConfigured() bool {
	return c != nil && strings.TrimSpace(c.config.FolderID) != "" && strings.TrimSpace(c.config.APIKey) != ""
}

func (c *Client) Model() string {
	if c == nil {
		return ""
	}
	return c.config.Model
}

func (c *Client) AnalyzeVacancy(ctx context.Context, vacancy models.Opportunity) (string, error) {
	return c.analyzeVacancy(ctx, vacancy, employerVacancyAnalyticsInstructions())
}

func (c *Client) AnalyzeVacancyForApplicant(ctx context.Context, vacancy models.Opportunity) (string, error) {
	return c.analyzeVacancy(ctx, vacancy, applicantVacancyAnalyticsInstructions())
}

func (c *Client) analyzeVacancy(ctx context.Context, vacancy models.Opportunity, instructions string) (string, error) {
	if !c.IsConfigured() {
		return "", ErrNotConfigured
	}

	reqData := responseRequest{
		Model:           fmt.Sprintf("gpt://%s/%s", c.config.FolderID, c.config.Model),
		Temperature:     c.config.Temperature,
		Instructions:    instructions,
		Input:           vacancyAnalyticsInput(vacancy),
		MaxOutputTokens: c.config.MaxOutputTokens,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("marshal yandex ai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create yandex ai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+c.config.APIKey)
	req.Header.Set("OpenAI-Project", c.config.FolderID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrProvider, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read yandex ai response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: status %s: %s", ErrProvider, resp.Status, truncate(string(body), 500))
	}

	var response responseData
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("parse yandex ai response: %w", err)
	}
	text := strings.TrimSpace(response.outputText())
	if text == "" {
		return "", fmt.Errorf("%w: empty output", ErrProvider)
	}
	return text, nil
}

func (r responseData) outputText() string {
	if strings.TrimSpace(r.OutputText) != "" {
		return r.OutputText
	}
	for _, output := range r.Output {
		for _, content := range output.Content {
			if strings.TrimSpace(content.Text) != "" {
				return content.Text
			}
		}
	}
	return ""
}

func employerVacancyAnalyticsInstructions() string {
	return strings.TrimSpace(`
Ты HR-аналитик платформы для студентов и работодателей.
Проанализируй вакансию на русском языке: оцени понятность, привлекательность для кандидатов, полноту условий, риски и способы улучшить конверсию в отклик.
Пиши конкретно, прикладно и дружелюбно. Тон: профессиональный, спокойный, помогающий пользователю принять решение.
Не придумывай факты, которых нет в данных вакансии.

Верни только красивый HTML-фрагмент без markdown, без doctype, html, head, body, script, style и inline-стилей.
Используй только эти теги: <article>, <header>, <section>, <div>, <h2>, <h3>, <p>, <ul>, <li>, <strong>, <em>, <span>.
Можно использовать только class-атрибуты. Не используй id, onclick, href, src, data-атрибуты, таблицы, горизонтальные линии, эмодзи и markdown-маркеры вроде ###, **, -.

Сделай ответ визуально похожим на аккуратный аналитический отчёт, который приятно читать в веб-интерфейсе.
Не раздувай текст: hero — 2 предложения, каждый список — не больше 3 пунктов, один пункт = 1-2 предложения. Не повторяй одни и те же мысли в разных секциях.

Строго соблюдай структуру и классы:
<article class="ai-report">
  <header class="ai-report__hero">
    <span class="ai-report__eyebrow">ИИ-анализ вакансии</span>
    <h2 class="ai-report__title">Краткий вывод</h2>
    <p class="ai-report__summary">2-3 предложения с главной оценкой.</p>
    <div class="ai-report__badges">
      <span class="ai-report__badge ai-report__badge--positive">Главный плюс: ...</span>
      <span class="ai-report__badge ai-report__badge--warning">Главный риск: ...</span>
      <span class="ai-report__badge ai-report__badge--neutral">Что улучшить: ...</span>
    </div>
  </header>

  <section class="ai-report__grid">
    <div class="ai-report__card ai-report__card--positive">
      <h3 class="ai-report__card-title">Сильные стороны</h3>
      <ul class="ai-report__list">
        <li><strong>Короткий тезис.</strong> Пояснение в одном предложении.</li>
      </ul>
    </div>

    <div class="ai-report__card ai-report__card--warning">
      <h3 class="ai-report__card-title">Риски и пробелы</h3>
      <ul class="ai-report__list">
        <li><strong>Короткий тезис.</strong> Пояснение в одном предложении.</li>
      </ul>
    </div>
  </section>

  <section class="ai-report__card ai-report__card--action">
    <h3 class="ai-report__card-title">Что улучшить в первую очередь</h3>
    <ul class="ai-report__steps">
      <li><strong>1. Действие.</strong> Почему это повысит доверие или конверсию.</li>
      <li><strong>2. Действие.</strong> Почему это повысит доверие или конверсию.</li>
      <li><strong>3. Действие.</strong> Почему это повысит доверие или конверсию.</li>
    </ul>
  </section>

  <section class="ai-report__card ai-report__card--candidate">
    <h3 class="ai-report__card-title">Кому подойдёт вакансия</h3>
    <p class="ai-report__text">Портрет подходящего кандидата в 2-4 предложениях.</p>
  </section>
</article>
`)
}

func applicantVacancyAnalyticsInstructions() string {
	return strings.TrimSpace(`
Ты карьерный консультант для студентов и соискателей.
Проанализируй вакансию на русском языке глазами кандидата: стоит ли откликаться, насколько понятны условия, какие плюсы и риски есть для соискателя, что уточнить у работодателя и как подготовиться к отклику.
Пиши конкретно, прикладно и дружелюбно. Тон: профессиональный, спокойный, помогающий пользователю принять решение.
Не придумывай факты, которых нет в данных вакансии. Если данных мало, прямо говори, какую информацию нужно уточнить.

Верни только красивый HTML-фрагмент без markdown, без doctype, html, head, body, script, style и inline-стилей.
Используй только эти теги: <article>, <header>, <section>, <div>, <h2>, <h3>, <p>, <ul>, <li>, <strong>, <em>, <span>.
Можно использовать только class-атрибуты. Не используй id, onclick, href, src, data-атрибуты, таблицы, горизонтальные линии, эмодзи и markdown-маркеры вроде ###, **, -.

Сделай ответ визуально похожим на аккуратную карточку карьерного советника.
Не раздувай текст: hero — 2 предложения, каждый список — не больше 3 пунктов, один пункт = 1-2 предложения. Не повторяй одни и те же мысли в разных секциях.

Строго соблюдай структуру и классы:
<article class="ai-report ai-report--applicant">
  <header class="ai-report__hero">
    <h2 class="ai-report__title">Стоит ли откликаться?</h2>
    <p class="ai-report__summary">2-3 предложения с главным выводом для кандидата.</p>
    <div class="ai-report__badges">
      <span class="ai-report__badge ai-report__badge--positive">Подходит: ...</span>
      <span class="ai-report__badge ai-report__badge--warning">Риск: ...</span>
      <span class="ai-report__badge ai-report__badge--neutral">Уточнить: ...</span>
    </div>
  </header>

  <section class="ai-report__grid">
    <div class="ai-report__card ai-report__card--positive">
      <h3 class="ai-report__card-title">Почему вакансия может быть интересна</h3>
      <ul class="ai-report__list">
        <li><strong>Короткий тезис.</strong> Что это значит для соискателя.</li>
      </ul>
    </div>

    <div class="ai-report__card ai-report__card--warning">
      <h3 class="ai-report__card-title">На что обратить внимание</h3>
      <ul class="ai-report__list">
        <li><strong>Короткий тезис.</strong> Какой риск или неопределённость это создаёт для кандидата.</li>
      </ul>
    </div>
  </section>

  <section class="ai-report__card ai-report__card--questions">
    <h3 class="ai-report__card-title">Что спросить у работодателя</h3>
    <ul class="ai-report__list">
      <li><strong>Вопрос.</strong> Почему ответ важен перед откликом или собеседованием.</li>
    </ul>
  </section>

  <section class="ai-report__card ai-report__card--action">
    <h3 class="ai-report__card-title">Как подготовиться к отклику</h3>
    <ul class="ai-report__steps">
      <li><strong>1. Действие.</strong> Что подготовить в резюме, портфолио или сопроводительном письме.</li>
      <li><strong>2. Действие.</strong> Как связать свой опыт с задачами вакансии.</li>
      <li><strong>3. Действие.</strong> Что повторить или уточнить перед собеседованием.</li>
    </ul>
  </section>

  <section class="ai-report__card ai-report__card--candidate">
    <h3 class="ai-report__card-title">Кому вакансия подойдёт лучше всего</h3>
    <p class="ai-report__text">Портрет подходящего соискателя в 2-4 предложениях.</p>
  </section>
</article>
`)
}

func vacancyAnalyticsInput(vacancy models.Opportunity) string {
	var b strings.Builder
	writeLine(&b, "ID", vacancy.ID)
	writeLine(&b, "Название", vacancy.Title)
	writeLine(&b, "Краткое описание", vacancy.ShortDescription)
	writeLine(&b, "Полное описание", vacancy.FullDescription)
	writeLine(&b, "Тип", vacancy.OpportunityType)
	writeLine(&b, "Уровень", vacancy.VacancyLevel)
	writeLine(&b, "Занятость", vacancy.EmploymentType)
	writeLine(&b, "Формат работы", vacancy.WorkFormat)
	writeLine(&b, "Локация ID", vacancy.LocationID)
	writeLine(&b, "Зарплата", salaryText(vacancy))
	writeLine(&b, "Зарплата видима", boolText(vacancy.IsSalaryVisible))
	writeLine(&b, "Статус", vacancy.Status)
	writeLine(&b, "Дедлайн отклика", timeText(vacancy.ApplicationDeadline))
	writeLine(&b, "Истекает", timeText(vacancy.ExpiresAt))
	writeLine(&b, "Просмотры", fmt.Sprintf("%d", vacancy.ViewsCount))
	writeLine(&b, "В избранном", fmt.Sprintf("%d", vacancy.FavoritesCount))
	writeLine(&b, "Отклики", fmt.Sprintf("%d", vacancy.ApplicationsCount))
	if len(vacancy.TagIDs) > 0 {
		writeLine(&b, "Теги ID", strings.Join(vacancy.TagIDs, ", "))
	}
	return b.String()
}

func writeLine(b *strings.Builder, label, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "не указано"
	}
	fmt.Fprintf(b, "%s: %s\n", label, value)
}

func salaryText(vacancy models.Opportunity) string {
	currency := strings.TrimSpace(vacancy.SalaryCurrency)
	if currency == "" {
		currency = "валюта не указана"
	}
	switch {
	case vacancy.SalaryMin > 0 && vacancy.SalaryMax > 0:
		return fmt.Sprintf("%.0f-%.0f %s", vacancy.SalaryMin, vacancy.SalaryMax, currency)
	case vacancy.SalaryMin > 0:
		return fmt.Sprintf("от %.0f %s", vacancy.SalaryMin, currency)
	case vacancy.SalaryMax > 0:
		return fmt.Sprintf("до %.0f %s", vacancy.SalaryMax, currency)
	default:
		return "не указана"
	}
}

func timeText(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func boolText(value bool) string {
	if value {
		return "да"
	}
	return "нет"
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
