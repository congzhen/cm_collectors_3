package processorsAI

import (
	"encoding/json"
	"fmt"
)

type PromptTagClass struct {
	TagClassID   string      `json:"tagClassId"`
	TagClassName string      `json:"tagClassName"`
	Tags         []PromptTag `json:"tags"`
}

type PromptTag struct {
	TagID         string `json:"tagId"`
	Name          string `json:"name"`
	AIDescription string `json:"aiDescription"`
}

type PromptResource struct {
	ResourceID  string   `json:"resourceId"`
	Title       string   `json:"title"`
	IssueNumber string   `json:"issueNumber"`
	Abstract    string   `json:"abstract"`
	Country     string   `json:"country"`
	Definition  string   `json:"definition"`
	Performers  []string `json:"performers"`
	Directors   []string `json:"directors"`
	Files       []string `json:"files"`
}

// BuildTagPrompt 构造单批截图的 AI 提示词。
// 每批都会携带完整资源信息和完整标签池，但只附带当前批图片；这样可以降低单次上下文压力，
// 又能让后端在所有批次完成后统一汇总，最终仍然是“资源级”标签。
func BuildTagPrompt(resource PromptResource, tagClasses []PromptTagClass, batchIndex, batchCount int) (string, error) {
	payload := map[string]interface{}{
		"resource":    resource,
		"tagClasses":  tagClasses,
		"batchIndex":  batchIndex,
		"batchCount":  batchCount,
		"instruction": "You are a resource tag classifier. Choose tagId values only from tagClasses. Never create new tags. Tag names may be abbreviations, so prefer aiDescription when judging meaning. If evidence is insufficient, return an empty tags array and uncertain=true. Return exactly one valid JSON object only. Do not return markdown, code fences, comments, explanations, or any text before or after the JSON. The JSON shape must be {\"batchIndex\": number, \"tags\": [{\"tagId\": string, \"confidence\": number, \"reason\": string}], \"summary\": string, \"uncertain\": boolean}.",
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Analyze the resource information, tag descriptions, and the current batch of images. Select existing tags for the whole resource. Input JSON:\n%s", string(b)), nil
}
