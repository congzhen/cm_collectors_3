package utils

import (
	"encoding/xml"
	"io"
	"strings"
)

// 将XML解析为map[string]interface{}结构
func XML_parseXMLToMap(decoder *xml.Decoder) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch se := token.(type) {
		case xml.StartElement:
			name := se.Name.Local
			value, err := XML_parseXMLElement(decoder, se)
			if err != nil {
				return nil, err
			}

			// 处理重复元素名，存储为数组
			if existing, exists := result[name]; exists {
				switch existing := existing.(type) {
				case []interface{}:
					result[name] = append(existing, value)
				default:
					result[name] = []interface{}{existing, value}
				}
			} else {
				result[name] = value
			}
		}
	}

	return result, nil
}

// 解析单个XML元素
func XML_parseXMLElement(decoder *xml.Decoder, start xml.StartElement) (interface{}, error) {
	elementData := make(map[string]interface{})

	// 解析元素的属性
	for _, attr := range start.Attr {
		elementData["@"+attr.Name.Local] = attr.Value
	}

	// 处理元素内容和子元素
	content := ""
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// 递归解析子元素
			childValue, err := XML_parseXMLElement(decoder, t)
			if err != nil {
				return nil, err
			}

			name := t.Name.Local
			if existing, exists := elementData[name]; exists {
				switch existing := existing.(type) {
				case []interface{}:
					elementData[name] = append(existing, childValue)
				default:
					elementData[name] = []interface{}{existing, childValue}
				}
			} else {
				elementData[name] = childValue
			}

		case xml.CharData:
			// 处理文本内容
			content += string(t)

		case xml.EndElement:
			// 元素结束
			if t.Name.Local == start.Name.Local {
				// 如果只有文本内容且没有子元素，则返回文本内容
				if len(elementData) == len(start.Attr) { // 只有属性
					trimmedContent := strings.TrimSpace(content)
					if trimmedContent != "" {
						if len(start.Attr) == 0 { // 没有属性
							return trimmedContent, nil
						}
						// 有属性的情况
						elementData["#text"] = trimmedContent
					} else if len(start.Attr) == 0 {
						// 没有内容也没有属性
						return "", nil
					}
				}
				return elementData, nil
			}
		}
	}
}

// 根据路径获取XML中的值，支持嵌套节点（如：movie.title）
func XML_getXMLValueByPath(data map[string]interface{}, path string) interface{} {
	// 分割路径
	paths := strings.Split(path, ".")
	current := data

	// 遍历路径
	for i, p := range paths {
		// 如果是最后一个元素，直接获取值
		if i == len(paths)-1 {
			if val, ok := current[p]; ok {
				return val
			}
			return ""
		}

		// 否则继续深入
		if val, ok := current[p]; ok {
			switch v := val.(type) {
			case map[string]interface{}:
				current = v
			case []interface{}:
				// 如果是数组，取第一个元素
				if len(v) > 0 {
					if next, ok := v[0].(map[string]interface{}); ok {
						current = next
					} else {
						return ""
					}
				} else {
					return ""
				}
			default:
				return ""
			}
		} else {
			return ""
		}
	}

	return ""
}

// 根据路径获取XML中的值数组，支持嵌套节点（如：movie.actors.actor）
func XML_getXMLValuesByPath(data map[string]interface{}, path string) []string {
	paths := strings.Split(path, ".")
	current := data

	// 遍历到倒数第二个路径
	for i := 0; i < len(paths)-1; i++ {
		p := paths[i]
		if val, ok := current[p]; ok {
			switch v := val.(type) {
			case map[string]interface{}:
				current = v
			case []interface{}:
				// 如果是数组，取第一个元素
				if len(v) > 0 {
					if next, ok := v[0].(map[string]interface{}); ok {
						current = next
					} else {
						return []string{}
					}
				} else {
					return []string{}
				}
			default:
				return []string{}
			}
		} else {
			return []string{}
		}
	}

	// 获取最后一个路径的值
	lastPath := paths[len(paths)-1]
	if val, ok := current[lastPath]; ok {
		// 如果是数组
		if arr, ok := val.([]interface{}); ok {
			var result []string
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				} else if mapped, ok := item.(map[string]interface{}); ok {
					// 如果是map，尝试获取#text字段
					if text, exists := mapped["#text"]; exists {
						if str, ok := text.(string); ok {
							result = append(result, str)
						}
					}
				}
			}
			return result
		}
		// 如果是单个值
		if str, ok := val.(string); ok {
			return []string{str}
		} else if mapped, ok := val.(map[string]interface{}); ok {
			// 如果是map，尝试获取#text字段
			if text, exists := mapped["#text"]; exists {
				if str, ok := text.(string); ok {
					return []string{str}
				}
			}
		}
	}

	return []string{}
}
