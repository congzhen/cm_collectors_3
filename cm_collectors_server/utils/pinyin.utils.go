package utils

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// 获取字符串的拼音首字母
func PinyinInitials(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder

	for _, r := range s {
		// 对于英文字母和数字，直接保留
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if unicode.Is(unicode.Scripts["Han"], r) {
			// 对于汉字，使用原来的pinyin库处理
			initial := getPinyinInitial(r)
			if initial != 0 {
				result.WriteRune(initial)
			}
		} else {
			// 处理日文假名
			initial := getJapaneseKanaInitial(r)
			if initial != 0 {
				result.WriteRune(initial)
			}
			// 其他字符（包括空格和特殊字符）会被忽略，不添加到结果中
		}
	}

	return result.String()
}

// 获取汉字的拼音首字母
func getPinyinInitial(r rune) rune {
	// 使用go-pinyin库获取汉字拼音首字母
	args := &pinyin.Args{}
	results := pinyin.LazyConvert(string(r), args)
	if len(results) > 0 && len(results[0]) > 0 {
		return rune(results[0][0])
	}
	return 0
}

// 获取日文假名的首字母
func getJapaneseKanaInitial(r rune) rune {
	// 平假名处理
	switch {
	case r >= 'ぁ' && r <= 'ゔ':
		return getHiraganaInitial(r)
	case r >= 'ァ' && r <= 'ヴ':
		return getKatakanaInitial(r)
	}

	// 非假名字符返回0表示不处理
	return 0
}

// 获取平假名首字母
func getHiraganaInitial(r rune) rune {
	switch r {
	case 'あ', 'い', 'う', 'え', 'お':
		return 'a'
	case 'か', 'き', 'く', 'け', 'こ':
		return 'k'
	case 'さ', 'し', 'す', 'せ', 'そ':
		return 's'
	case 'た', 'ち', 'つ', 'て', 'と':
		return 't'
	case 'な', 'に', 'ぬ', 'ね', 'の':
		return 'n'
	case 'は', 'ひ', 'ふ', 'へ', 'ほ':
		return 'h'
	case 'ま', 'み', 'む', 'め', 'も':
		return 'm'
	case 'や', 'ゆ', 'よ':
		return 'y'
	case 'ら', 'り', 'る', 'れ', 'ろ':
		return 'r'
	case 'わ', 'を':
		return 'w'
	case 'が', 'ぎ', 'ぐ', 'げ', 'ご':
		return 'g'
	case 'ざ', 'じ', 'ず', 'ぜ', 'ぞ':
		return 'z'
	case 'だ', 'ぢ', 'づ', 'で', 'ど':
		return 'd'
	case 'ば', 'び', 'ぶ', 'べ', 'ぼ':
		return 'b'
	case 'ぱ', 'ぴ', 'ぷ', 'ぺ', 'ぽ':
		return 'p'
	}
	return 0
}

// 获取片假名首字母
func getKatakanaInitial(r rune) rune {
	switch r {
	case 'ア', 'イ', 'ウ', 'エ', 'オ':
		return 'a'
	case 'カ', 'キ', 'ク', 'ケ', 'コ':
		return 'k'
	case 'サ', 'シ', 'ス', 'セ', 'ソ':
		return 's'
	case 'タ', 'チ', 'ツ', 'テ', 'ト':
		return 't'
	case 'ナ', 'ニ', 'ヌ', 'ネ', 'ノ':
		return 'n'
	case 'ハ', 'ヒ', 'フ', 'ヘ', 'ホ':
		return 'h'
	case 'マ', 'ミ', 'ム', 'メ', 'モ':
		return 'm'
	case 'ヤ', 'ユ', 'ヨ':
		return 'y'
	case 'ラ', 'リ', 'ル', 'レ', 'ロ':
		return 'r'
	case 'ワ', 'ヲ':
		return 'w'
	case 'ガ', 'ギ', 'グ', 'ゲ', 'ゴ':
		return 'g'
	case 'ザ', 'ジ', 'ズ', 'ゼ', 'ゾ':
		return 'z'
	case 'ダ', 'ヂ', 'ヅ', 'デ', 'ド':
		return 'd'
	case 'バ', 'ビ', 'ブ', 'ベ', 'ボ':
		return 'b'
	case 'パ', 'ピ', 'プ', 'ペ', 'ポ':
		return 'p'
	}
	return 0
}
