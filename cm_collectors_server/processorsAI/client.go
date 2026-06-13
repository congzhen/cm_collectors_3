package processorsAI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ImageInput struct {
	DataURL string
}

type TagSuggestion struct {
	TagID      string  `json:"tagId"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type TagAnalyzeResult struct {
	BatchIndex int             `json:"batchIndex"`
	Tags       []TagSuggestion `json:"tags"`
	Summary    string          `json:"summary"`
	Uncertain  bool            `json:"uncertain"`
}

type ModelTestMetrics struct {
	PromptTokens              int     `json:"promptTokens"`
	CompletionTokens          int     `json:"completionTokens"`
	TotalTokens               int     `json:"totalTokens"`
	UsageReturned             bool    `json:"usageReturned"`
	ElapsedMs                 int64   `json:"elapsedMs"`
	EstimatedTokensPerSecond  float64 `json:"estimatedTokensPerSecond"`
	ServiceTokensPerSecond    float64 `json:"serviceTokensPerSecond"`
	ServicePromptPerSecond    float64 `json:"servicePromptPerSecond"`
	ServiceGeneratedPerSecond float64 `json:"serviceGeneratedPerSecond"`
}

// ModelTestReport 是“测试模型输出”接口返回给前端的诊断报告。
// 它不会参与正式打标签，只用于判断模型是否能稳定返回 JSON、耗时和吞吐是否适合当前参数。
type ModelTestReport struct {
	Success        bool             `json:"success"`
	Model          string           `json:"model"`
	Endpoint       string           `json:"endpoint"`
	ResponseFormat string           `json:"responseFormat"`
	FallbackUsed   bool             `json:"fallbackUsed"`
	FinishReason   string           `json:"finishReason"`
	Summary        string           `json:"summary"`
	Content        string           `json:"content"`
	Error          string           `json:"error"`
	FirstError     string           `json:"firstError"`
	Metrics        ModelTestMetrics `json:"metrics"`
}

type Client struct {
	BaseURL               string
	APIKey                string
	Model                 string
	RequestTimeoutSeconds int
}

// AnalyzeTags 是正式 AI 打标签调用入口。
// 先使用 json_schema 约束输出；如果服务不支持该格式，再回退到 text 模式并由后端解析 JSON。
func (c Client) AnalyzeTags(prompt string, images []ImageInput, batchIndex int) (*TagAnalyzeResult, string, error) {
	if strings.TrimSpace(c.Model) == "" {
		return nil, "", fmt.Errorf("AI模型不能为空")
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		return nil, "", fmt.Errorf("AI调用地址不能为空")
	}
	content := []map[string]interface{}{
		{"type": "text", "text": prompt},
	}
	for _, img := range images {
		if img.DataURL == "" {
			continue
		}
		content = append(content, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]interface{}{
				"url": img.DataURL,
			},
		})
	}
	result, contentText, err := c.analyzeTagsWithFormat(content, batchIndex, tagAnalyzeResponseFormat())
	if err == nil {
		return result, contentText, nil
	}
	fallbackResult, fallbackContentText, fallbackErr := c.analyzeTagsWithFormat(content, batchIndex, map[string]string{"type": "text"})
	if fallbackErr == nil {
		return fallbackResult, fallbackContentText, nil
	}
	return nil, fallbackContentText, fmt.Errorf("%w；已回退 text 模式仍失败：%v", err, fallbackErr)
}

// analyzeTagsWithFormat 执行一次指定 response_format 的 Chat Completions 请求。
// 该函数只返回业务需要的 JSON 结果；测试页面需要更完整的指标时使用 analyzeTagsWithReport。
func (c Client) analyzeTagsWithFormat(content []map[string]interface{}, batchIndex int, responseFormat interface{}) (*TagAnalyzeResult, string, error) {
	payload := c.chatPayload(content, responseFormat)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}
	respBody, err := c.post(body)
	if err != nil {
		return nil, "", err
	}
	contentText, err := extractChatContent(respBody)
	if err != nil {
		return nil, string(respBody), err
	}
	var result TagAnalyzeResult
	if err := json.Unmarshal([]byte(extractJSONText(contentText)), &result); err != nil {
		return nil, contentText, fmt.Errorf("AI返回JSON解析失败: %w", err)
	}
	if result.BatchIndex == 0 {
		result.BatchIndex = batchIndex
	}
	return &result, contentText, nil
}

// chatPayload 构造 OpenAI 兼容的 Chat Completions 请求体。
// system prompt 强约束模型只返回 JSON，减少 markdown 代码块和解释性文本导致的解析失败。
func (c Client) chatPayload(content []map[string]interface{}, responseFormat interface{}) map[string]interface{} {
	return map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a strict JSON API. Return exactly one JSON object matching the requested schema. Do not use markdown, code fences, comments, explanations, or trailing text.",
			},
			{
				"role":    "user",
				"content": content,
			},
		},
		"response_format": responseFormat,
	}
}

// tagAnalyzeResponseFormat 返回严格 JSON schema。
// 有些本地服务不支持 json_schema，调用方会自动回退到 text，而不是把兼容分支散落在业务层。
func tagAnalyzeResponseFormat() map[string]interface{} {
	return map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   "tag_analyze_result",
			"strict": true,
			"schema": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"batchIndex", "tags", "summary", "uncertain"},
				"properties": map[string]interface{}{
					"batchIndex": map[string]interface{}{"type": "integer"},
					"tags": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type":                 "object",
							"additionalProperties": false,
							"required":             []string{"tagId", "confidence", "reason"},
							"properties": map[string]interface{}{
								"tagId":      map[string]interface{}{"type": "string"},
								"confidence": map[string]interface{}{"type": "number"},
								"reason":     map[string]interface{}{"type": "string"},
							},
						},
					},
					"summary":   map[string]interface{}{"type": "string"},
					"uncertain": map[string]interface{}{"type": "boolean"},
				},
			},
		},
	}
}

func (c Client) TestConnection() error {
	_, _, err := c.AnalyzeTags(`{"instruction":"return strict JSON: {\"tags\":[],\"summary\":\"ok\",\"uncertain\":false}"}`, nil, 1)
	return err
}

// TestModelOutput 执行一次轻量模型输出测试，并返回结构化诊断。
// 它保留 json_schema -> text fallback 的完整过程，方便前端说明“服务可用但需要兼容模式”。
func (c Client) TestModelOutput() *ModelTestReport {
	report := &ModelTestReport{
		Model:    c.Model,
		Endpoint: c.endpoint(),
	}
	if strings.TrimSpace(c.Model) == "" {
		report.Error = "AI模型不能为空"
		return report
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		report.Error = "AI调用地址不能为空"
		return report
	}
	content := []map[string]interface{}{
		{
			"type": "text",
			"text": `Return exactly this JSON object, and do not add markdown: {"batchIndex":1,"tags":[],"summary":"ok","uncertain":false}`,
		},
	}
	result, contentText, finishReason, metrics, err := c.analyzeTagsWithReport(content, 1, tagAnalyzeResponseFormat())
	if err == nil {
		report.Success = true
		report.ResponseFormat = "json_schema"
		report.FinishReason = finishReason
		report.Content = contentText
		report.Metrics = metrics
		if result != nil {
			report.Summary = result.Summary
		}
		return report
	}
	report.FirstError = err.Error()

	result, contentText, finishReason, metrics, fallbackErr := c.analyzeTagsWithReport(content, 1, map[string]string{"type": "text"})
	report.ResponseFormat = "text"
	report.FallbackUsed = true
	report.FinishReason = finishReason
	report.Content = contentText
	report.Metrics = metrics
	if fallbackErr != nil {
		report.Error = fallbackErr.Error()
		return report
	}
	report.Success = true
	if result != nil {
		report.Summary = result.Summary
	}
	return report
}

// TestService 只检查 /models 是否可访问以及模型名是否存在。
// 它不做推理请求，所以速度快；真实输出能力由 TestModelOutput 判断。
func (c Client) TestService() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("AI调用地址不能为空")
	}
	req, err := http.NewRequest(http.MethodGet, c.modelsEndpoint(), nil)
	if err != nil {
		return err
	}
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("AI服务请求失败: %s %s", resp.Status, string(respBody))
	}
	if strings.TrimSpace(c.Model) == "" {
		return nil
	}
	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil || len(data.Data) == 0 {
		return nil
	}
	for _, model := range data.Data {
		if model.ID == c.Model {
			return nil
		}
	}
	return fmt.Errorf("AI服务可用，但没有找到当前模型: %s", c.Model)
}

// analyzeTagsWithReport 与 analyzeTagsWithFormat 类似，但额外记录耗时、finish_reason、usage 和扩展性能字段。
// 测试接口失败时也需要这些信息，所以错误路径同样返回已采集到的 metrics。
func (c Client) analyzeTagsWithReport(content []map[string]interface{}, batchIndex int, responseFormat interface{}) (*TagAnalyzeResult, string, string, ModelTestMetrics, error) {
	payload := c.chatPayload(content, responseFormat)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", "", ModelTestMetrics{}, err
	}
	start := time.Now()
	respBody, err := c.post(body)
	elapsed := time.Since(start)
	metrics := ModelTestMetrics{ElapsedMs: elapsed.Milliseconds()}
	if err != nil {
		return nil, "", "", metrics, err
	}
	contentText, finishReason, parsedMetrics, err := extractChatDetails(respBody)
	parsedMetrics.ElapsedMs = metrics.ElapsedMs
	if parsedMetrics.CompletionTokens > 0 && elapsed.Seconds() > 0 {
		parsedMetrics.EstimatedTokensPerSecond = roundFloat(float64(parsedMetrics.CompletionTokens)/elapsed.Seconds(), 2)
	}
	if err != nil {
		return nil, string(respBody), finishReason, parsedMetrics, err
	}
	var result TagAnalyzeResult
	if err := json.Unmarshal([]byte(extractJSONText(contentText)), &result); err != nil {
		return nil, contentText, finishReason, parsedMetrics, fmt.Errorf("AI返回JSON解析失败: %w", err)
	}
	if result.BatchIndex == 0 {
		result.BatchIndex = batchIndex
	}
	return &result, contentText, finishReason, parsedMetrics, nil
}

// post 执行 HTTP 请求并统一处理超时和非 2xx 状态。
// 超时时间来自 AI 自动标签设置，本地大模型默认给较长时间，避免长推理被前端请求生命周期截断。
func (c Client) post(body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, c.endpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	timeoutSeconds := c.RequestTimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 1800
	}
	client := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AI请求失败: %s %s", resp.Status, string(respBody))
	}
	return respBody, nil
}

// endpoint 兼容用户输入的多种地址形式：
// 根地址、/v1、/models、/chat 或完整 /chat/completions 都会归一化到推理接口。
func (c Client) endpoint() string {
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if strings.Contains(base, "/chat/completions") {
		return base
	}
	if strings.HasSuffix(base, "/chat") {
		return strings.TrimSuffix(base, "/chat") + "/chat/completions"
	}
	if strings.HasSuffix(base, "/models") {
		return strings.TrimSuffix(base, "/models") + "/chat/completions"
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

// modelsEndpoint 与 endpoint 对应，用于“测试服务”快速检查模型列表。
func (c Client) modelsEndpoint() string {
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if strings.HasSuffix(base, "/models") {
		return base
	}
	if strings.Contains(base, "/chat/completions") {
		return strings.TrimSuffix(base, "/chat/completions") + "/models"
	}
	if strings.HasSuffix(base, "/chat") {
		return strings.TrimSuffix(base, "/chat") + "/models"
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/models"
	}
	return base + "/v1/models"
}

// extractJSONText 从模型回复中提取 JSON 对象。
// 兼容模型偶尔包一层 ```json 代码块或在 JSON 前后添加少量文本的情况。
func extractJSONText(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}

func extractChatContent(respBody []byte) (string, error) {
	content, _, _, err := extractChatDetails(respBody)
	return content, err
}

// extractChatDetails 解析 Chat Completions 响应。
// 正式路径只关心 content；测试路径还需要 finish_reason 和 usage 指标，因此统一在这里拆解。
func extractChatDetails(respBody []byte) (string, string, ModelTestMetrics, error) {
	var data struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage map[string]interface{} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return "", "", ModelTestMetrics{}, fmt.Errorf("AI响应解析失败: %w", err)
	}
	metrics := parseModelTestMetrics(respBody, data.Usage)
	if len(data.Choices) == 0 {
		return "", "", metrics, fmt.Errorf("AI响应为空 choices=0")
	}
	content := strings.TrimSpace(data.Choices[0].Message.Content)
	if content == "" {
		return "", data.Choices[0].FinishReason, metrics, fmt.Errorf("AI响应为空 finish_reason=%s", data.Choices[0].FinishReason)
	}
	return content, data.Choices[0].FinishReason, metrics, nil
}

// parseModelTestMetrics 兼容 OpenAI 标准 usage 和常见本地服务扩展性能字段。
// 没有 usage 不视为失败，因为部分兼容服务只返回文本，不返回 token 统计。
func parseModelTestMetrics(respBody []byte, usage map[string]interface{}) ModelTestMetrics {
	metrics := ModelTestMetrics{}
	if usage != nil {
		metrics.PromptTokens = intFromMap(usage, "prompt_tokens", "promptTokens", "input_tokens", "inputTokens")
		metrics.CompletionTokens = intFromMap(usage, "completion_tokens", "completionTokens", "output_tokens", "outputTokens", "generated_tokens", "generatedTokens")
		metrics.TotalTokens = intFromMap(usage, "total_tokens", "totalTokens")
		if metrics.TotalTokens == 0 && (metrics.PromptTokens > 0 || metrics.CompletionTokens > 0) {
			metrics.TotalTokens = metrics.PromptTokens + metrics.CompletionTokens
		}
		metrics.UsageReturned = metrics.PromptTokens > 0 || metrics.CompletionTokens > 0 || metrics.TotalTokens > 0
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return metrics
	}
	metrics.ServiceTokensPerSecond = roundFloat(floatFromNested(raw,
		[]string{"tokens_per_second"},
		[]string{"tokensPerSecond"},
		[]string{"eval_rate"},
		[]string{"timings", "tokens_per_second"},
		[]string{"timings", "predicted_per_second"},
	), 2)
	metrics.ServicePromptPerSecond = roundFloat(floatFromNested(raw,
		[]string{"prompt_per_second"},
		[]string{"promptPerSecond"},
		[]string{"prompt_eval_rate"},
		[]string{"timings", "prompt_per_second"},
	), 2)
	metrics.ServiceGeneratedPerSecond = roundFloat(floatFromNested(raw,
		[]string{"generated_per_second"},
		[]string{"generatedPerSecond"},
		[]string{"predicted_per_second"},
		[]string{"timings", "generated_per_second"},
		[]string{"timings", "predicted_per_second"},
	), 2)
	return metrics
}

func intFromMap(m map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			switch v := value.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case json.Number:
				i, _ := v.Int64()
				return int(i)
			}
		}
	}
	return 0
}

func floatFromNested(raw map[string]interface{}, paths ...[]string) float64 {
	for _, path := range paths {
		var current interface{} = raw
		for _, key := range path {
			currentMap, ok := current.(map[string]interface{})
			if !ok {
				current = nil
				break
			}
			current = currentMap[key]
		}
		switch v := current.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case json.Number:
			f, _ := v.Float64()
			return f
		}
	}
	return 0
}

func roundFloat(value float64, precision int) float64 {
	if value == 0 {
		return 0
	}
	scale := 1.0
	for i := 0; i < precision; i++ {
		scale *= 10
	}
	if value >= 0 {
		return float64(int(value*scale+0.5)) / scale
	}
	return float64(int(value*scale-0.5)) / scale
}
