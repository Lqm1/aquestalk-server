package koeconv

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/exp/slices"
	"golang.org/x/text/unicode/norm"

	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func TokensToTokenDatas(tokens []tokenizer.Token) []tokenizer.TokenData {
	var result []tokenizer.TokenData
	for _, token := range tokens {
		result = append(result, tokenizer.NewTokenData(token))
	}
	return result
}

func IsDigit(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func ApplyAquesTalkTags(tokens []tokenizer.TokenData) []tokenizer.TokenData {
	// 桁読みタグ
	// TODO: 助数詞の影響をうけて数詞・助数詞の読みやアクセントが変化 (<NUMK VAL=(数値) COUNTER=(助数詞)>)
	var numkResult []tokenizer.TokenData
	i := 0
	for i < len(tokens) {
		current := tokens[i]
		if IsDigit(current.Surface) {
			numberStr := current.Surface
			j := i + 1
			for j < len(tokens) && (IsDigit(tokens[j].Surface) || tokens[j].Surface == ".") {
				numberStr += tokens[j].Surface
				j++
			}
			numberStr = norm.NFKC.String(numberStr)
			newToken := current
			newToken.Start = current.Start
			newToken.End = tokens[j-1].End
			newToken.Surface = numberStr
			newToken.Pronunciation = fmt.Sprintf("<NUMK VAL=%s>", numberStr)
			numkResult = append(numkResult, newToken)
			i = j
		} else {
			numkResult = append(numkResult, current)
			i++
		}
	}

	tokens = numkResult

	// 英数読みタグ
	SYMBOL_TO_KANA := map[string]string{
		"!": "ビック'リ",
		"#": "シャ'ープ",
		"$": "ド'ル",
		"%": "パーセ'ント",
		"&": "アンド",
		"*": "ア'スタ",
		"+": "プラス",
		",": "カ'ンマ",
		"-": "ハ'イフン",
		".": "ドット",
		"/": "スラ'ッシュ",
		":": "コ'ロン",
		";": "セミコ'ロン",
		"<": "ショ'ーナリ",
		"=": "イコ'ール",
		">": "ダ'イナリ",
		"?": "ハ'テナ",
		"@": "ア'ット",
		"¥": "エ'ン",
		"^": "ハ'ット",
		"_": "ア'ンダー",
		" ": "、",
	}
	LETTER_TO_KANA := map[string]string{
		"a": "エー",
		"b": "ビー",
		"c": "シー",
		"d": "ディー",
		"e": "イー",
		"f": "エフ",
		"g": "ジー",
		"h": "エイチ",
		"i": "アイ",
		"j": "ジェイ",
		"k": "ケー",
		"l": "エル",
		"m": "エム",
		"n": "エヌ",
		"o": "オー",
		"p": "ピー",
		"q": "キュー",
		"r": "アール",
		"s": "エス",
		"t": "ティー",
		"u": "ユー",
		"v": "ブイ",
		"w": "ダブリュー",
		"x": "エックス",
		"y": "ワイ",
		"z": "ゼット",
	}
	var alphaResult []tokenizer.TokenData
	i = 0
	for i < len(tokens) {
		current := tokens[i]
		if IsLetter(current.Surface) && current.Pronunciation == "" {
			alphaStr := current.Surface
			j := i + 1
			for j < len(tokens) && (IsLetter(tokens[j].Surface)) {
				alphaStr += tokens[j].Surface
				j++
			}
			alphaStr = norm.NFKC.String(alphaStr)
			alphaStr = strings.ToLower(alphaStr)
			newAlphaStr := ""
			for _, r := range alphaStr {
				newAlphaStr += LETTER_TO_KANA[string(r)]
				newAlphaStr += SYMBOL_TO_KANA[string(r)]
			}
			newToken := current
			newToken.Start = current.Start
			newToken.End = tokens[j-1].End
			newToken.Surface = alphaStr
			newToken.Pronunciation = newAlphaStr
			alphaResult = append(alphaResult, newToken)
			i = j
		} else {
			alphaResult = append(alphaResult, current)
			i++
		}
	}

	tokens = alphaResult

	return tokens
}

type KoeConv struct {
	tokenizer *tokenizer.Tokenizer
}

func New() (*KoeConv, error) {
	t, err := tokenizer.New(uni.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, err
	}
	return &KoeConv{
		tokenizer: t,
	}, nil
}

func (k *KoeConv) Convert(input string) (string, error) {
	ALLOW_SYMBOL := []string{
		"。", "？", "、", ",", ";", "/", "+",
	}

	runes := []rune(input)
	tokens := k.tokenizer.Tokenize(input)
	tokenDatas := TokensToTokenDatas(tokens)
	tokenDatas = ApplyAquesTalkTags(tokenDatas)

	var result []rune
	lastIndex := 0

	for _, tokenData := range tokenDatas {
		result = append(result, runes[lastIndex:tokenData.Start]...)
		if tokenData.Pronunciation != "" {
			result = append(result, []rune(tokenData.Pronunciation)...)
		} else {
			if slices.Contains(ALLOW_SYMBOL, tokenData.Surface) {
				result = append(result, runes[tokenData.Start:tokenData.End]...)
			} else {
				result = append(result, []rune("、")...)
			}
		}
		lastIndex = tokenData.End
	}

	result = append(result, runes[lastIndex:]...)

	koe := string(result)
	fmt.Println(koe)
	return koe, nil
}
