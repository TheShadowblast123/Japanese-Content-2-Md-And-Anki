package main

import (
	"bufio"
	_ "embed"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
	"sync"

	"github.com/TheShadowblast123/Japanese-Content-2-Md-And-Anki/path_handler"
	"github.com/ikawaha/kagome/tokenizer"
)

type JMdict struct {
	XMLName xml.Name `xml:"JMdict"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	EntSeq string  `xml:"ent_seq"`
	KEle   []KEle  `xml:"k_ele"`
	REle   []REle  `xml:"r_ele"`
	Sense  []Sense `xml:"sense"`
}

type KEle struct {
	Keb   string   `xml:"keb"`
	KeInf []string `xml:"ke_inf"`
	KePri []string `xml:"ke_pri"`
}

type REle struct {
	Reb       string   `xml:"reb"`
	ReNoKanji string   `xml:"re_nokanji"`
	ReRestr   []string `xml:"re_restr"`
	ReInf     []string `xml:"re_inf"`
	RePri     []string `xml:"re_pri"`
}

type Sense struct {
	StagK   []string  `xml:"stagk"`
	StagR   []string  `xml:"stagr"`
	Pos     []string  `xml:"pos"`
	Xref    []string  `xml:"xref"`
	Ant     []string  `xml:"ant"`
	Field   []string  `xml:"field"`
	Misc    []string  `xml:"misc"`
	SInf    []string  `xml:"s_inf"`
	LSource []LSource `xml:"lsource"`
	Dial    []string  `xml:"dial"`
	Gloss   []Gloss   `xml:"gloss"`
	Example []Example `xml:"example"`
}

type LSource struct {
	Lang    string `xml:"xml:lang,attr"`
	Type    string `xml:"ls_type,attr"`
	Wasei   string `xml:"ls_wasei,attr"`
	Content string `xml:",chardata"`
}

type Gloss struct {
	Lang  string   `xml:"xml:lang,attr"`
	GGend string   `xml:"g_gend,attr"`
	GType string   `xml:"g_type,attr"`
	Text  string   `xml:",chardata"`
	Pri   []string `xml:"pri"`
}

type Example struct {
	ExSrce ExSrce   `xml:"ex_srce"`
	ExText string   `xml:"ex_text"`
	ExSent []ExSent `xml:"ex_sent"`
}

type ExSrce struct {
	Type string `xml:"exsrc_type,attr"`
	Text string `xml:",chardata"`
}

type ExSent struct {
	Lang string `xml:"xml:lang,attr"`
	Text string `xml:",chardata"`
}

// Top-level struct for kanjidic2
type Kanjidic2 struct {
	XMLName    xml.Name    `xml:"kanjidic2"`
	Characters []Character `xml:"character"`
}

// Character represents each kanji entry
type Character struct {
	Literal        string         `xml:"literal"`
	Misc           Misc           `xml:"misc"`
	ReadingMeaning ReadingMeaning `xml:"reading_meaning"`
}

// Misc contains miscellaneous information
type Misc struct {
	StrokeCount []int `xml:"stroke_count"`
	Freq        int   `xml:"freq"`
	JLPT        int   `xml:"jlpt"`
}

// ReadingMeaning contains readings and meanings
type ReadingMeaning struct {
	Groups   []RmGroup `xml:"rmgroup"` // Grouped readings/meanings
	Readings []Reading `xml:"reading"` // Direct readings
	Meanings []Meaning `xml:"meaning"` // Direct meanings
}

// RmGroup contains grouped readings and meanings
type RmGroup struct {
	Readings []Reading `xml:"reading"`
	Meanings []Meaning `xml:"meaning"`
}

// Reading represents pronunciation reading
type Reading struct {
	Type  string `xml:"r_type,attr"`
	Value string `xml:",chardata"`
}

// Meaning represents English meaning
type Meaning struct {
	Value string `xml:",chardata"`
}

// ProcessedKanji is the final structured data
type ProcessedKanji struct {
	Literal     string   `json:"literal"`
	StrokeCount int      `json:"stroke_count,omitempty"`
	Freq        int      `json:"freq,omitempty"`
	JLPT        int      `json:"jlpt,omitempty"`
	OnReadings  []string `json:"on_readings,omitempty"`
	KunReadings []string `json:"kun_readings,omitempty"`
	Meanings    []string `json:"meanings,omitempty"`
}

// Type definitions for data structures
type KanjiData struct {
	Kanji    string
	Keyword  string
	Readings string
	Strokes  int
	Radicals string
}
type Word struct {
	Pos      string
	DictForm string
	Form     string
	Word     string
}
type Augmentation struct {
	Description string
	PhraseStart bool
}
type Verb struct {
	Word          Word
	Augmentations []Augmentation
}
type WordData struct {
	Word        string
	Definitions string
	Reading     string
}

//go:embed kanjidic2.xml
var kanjiDic []byte

//go:embed JMdict_e.xml
var jmDictData []byte
var verbFormMap = map[string]string{
	"一段": "ichidan",
	"五段": "godan",
	"サ変": "suru",
	"カ変": "kuru",
}
var posMap = map[string]string{
	// Nouns
	"名詞":   "noun",
	"一般":   "general",
	"非自立":  "dependent",
	"固有名詞": "proper_noun",
	"代名詞":  "pronoun",

	// Verbs
	"動詞":  "verb",
	"自立":  "independent",
	"一段":  "ichidan",
	"五段":  "godan",
	"サ変":  "suru",
	"カ変":  "kuru",
	"連用形": "continuative",
	"基本形": "basic",
	"未然形": "imperfective",
	"命令形": "imperative",

	// Adjectives
	"形容詞":    "adjective",
	"形容動詞語幹": "na_adjective_stem",

	// Adverbs
	"副詞": "adverb",

	// Particles
	"助詞":   "particle",
	"格助詞":  "case_particle",
	"接続助詞": "conjunctive_particle",
	"係助詞":  "binding_particle",
	"副助詞":  "adverbial_particle",
	"連体化":  "adnominalizer",
	"終助詞":  "sentence_ending_particle",

	// Auxiliary verbs
	"助動詞": "auxiliary_verb",
	"特殊":  "special",

	// Others
	"接頭詞":  "prefix",
	"接尾":   "suffix",
	"記号":   "symbol",
	"空白":   "whitespace",
	"その他":  "other",
	"フィラー": "filler",
	"感動詞":  "interjection",
}

type SentenceData struct {
	Sentence    string
	Translation string
}

type FlashcardDict struct {
	Front string
	Back  string
	Cloze string
}

// Global variables
var (
	pathing        = path_handler.TestPathing
	oldKanji       []string
	oldWords       []string
	contentMd      = pathing.ContentMd
	kanjiMd        = pathing.KanjiMd
	sentencesMd    = pathing.SentencesMd
	wordsMd        = pathing.WordsMd
	contentPath    = pathing.ContentPath
	kanjiPath      = pathing.KanjiPath
	sentencesPath  = pathing.SentencesPath
	wordsPath      = pathing.WordsPath
	csvPath        = pathing.CsvPath
	newContentPath = pathing.NewContent
	currentName    = ""
	skipSentences  = false
)

// Unicode ranges for kanji detection
var (
	kanjiRange1 = [2]int{0x3400, 0x4DBF}
	kanjiRange2 = [2]int{0x4E00, 0x9FCB}
	kanjiRange3 = [2]int{0xF900, 0xFA6A}
	kanjiSet    []string
)

func init() {
	clearDir("Test")
	// Initialize kanji character sets
	for i := kanjiRange1[0]; i <= kanjiRange1[1]; i++ {
		kanjiSet = append(kanjiSet, string(rune(i)))
	}
	for i := kanjiRange2[0]; i <= kanjiRange2[1]; i++ {
		kanjiSet = append(kanjiSet, string(rune(i)))
	}
	for i := kanjiRange3[0]; i <= kanjiRange3[1]; i++ {
		kanjiSet = append(kanjiSet, string(rune(i)))
	}

	// Create directories if they don't exist
	dirs := []string{contentPath, kanjiPath, sentencesPath, wordsPath, csvPath, "./New Content"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}
	}

	// Create markdown index files if they don't exist
	files := []string{contentMd, kanjiMd, sentencesMd, wordsMd}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			f, _ := os.Create(file)
			f.Close()
		}
	}
}
func handleVerbs(dictForm, verbType, form, base string) Verb {
	iTells := []string{"ち", "り", "に", "み", "び", "き", "ぎ"}
	teTells := []string{"て", "で"}
	taTells := []string{"た", "だ"}
	tTells := []string{"ん", "っ"}
	currentVerb := Verb{
		Word: Word{
			Pos:      "verb",
			DictForm: dictForm,
			Form:     "",
			Word:     base,
		},
		Augmentations: []Augmentation{},
	}
	if strings.Contains(verbType, "五段") {
		//godan
		if getEnglishPOS(form) != "continuative" {
			switch form {
			case "基本形":
				currentVerb.Word.Form = "U"
			case "未然形":
				currentVerb.Word.Form = "A"
			case "未然ウ接続":
				currentVerb.Word.Form = "O"
			case "接続テ接続":
			case "連用タ接続":
				currentVerb.Word.Form = "T"
			default:
				currentVerb.Word.Form = "E Godan"

			}
		} else {

			if strings.Contains(verbType, "サ") {
				currentVerb.Word.Form = "I + T"
			} else {
				if form == "連用形" {
					currentVerb.Word.Form = "I"
				} else {

					input := base
					for _, suffix := range taTells {
						if strings.HasSuffix(input, suffix) {
							currentVerb.Word.Form = "Ta"
							return currentVerb
						}
					}
					for _, suffix := range teTells {
						if strings.HasSuffix(input, suffix) {
							currentVerb.Word.Form = "Te"
							return currentVerb
						}
					}
					currentVerb.Word.Form = "T"
				}
			}
		}
		return currentVerb
	} else {
		//ichidan, suru, kuru
		if getEnglishPOS(form) == "continuative" {
			for _, suffix := range taTells {
				if strings.HasSuffix(base, suffix) {

					currentVerb.Word.Form = "Ta"
					return currentVerb
				}
			}
			for _, suffix := range teTells {
				if strings.HasSuffix(base, suffix) {
					currentVerb.Word.Form = "Te"
					return currentVerb
				}
			}
			for _, suffix := range tTells {
				if strings.HasSuffix(base, suffix) {

					currentVerb.Word.Form = "T"
					return currentVerb
				}
			}

			for _, suffix := range iTells {
				if strings.HasSuffix(base, suffix) {
					currentVerb.Word.Form = "I"
					return currentVerb
				}
			}
			currentVerb.Word.Form = "I + T"
		} else {
			switch {
			case form == "基本形":
				currentVerb.Word.Form = "U"
			case form == "未然形":
				currentVerb.Word.Form = "A"
			case form == "未然ウ接続":
				currentVerb.Word.Form = "O"
			default:
				currentVerb.Word.Form = "E Ichidan"
			}
		}
	}
	return currentVerb
}
func skipVerbs(tokens []tokenizer.Token) []Word {
	var output []Word
	for _, token := range tokens {
		features := token.Features()
		if token.Class != tokenizer.KNOWN || getEnglishPOS(features[0]) == "symbol" {
			continue
		}
		word := Word{
			Pos:      features[0],
			DictForm: features[6],
			Form:     "",
			Word:     token.Surface,
		}
		if len(WordLookup(token.Surface)) > 0 {
			output = append(output, word)

			continue
		}
		if len(WordLookup(features[6])) > 0 {
			output = append(output, word)
			continue
		}

		if strings.HasSuffix(features[6], "せる") {
			test := strings.Split(features[6], "せ")[0] + "す"
			if len(WordLookup(test)) == 0 {
				output = append(output, word)
				continue
			}
			word.DictForm = test
			output = append(output, word)
		} else {
			continue
		}

		output = append(output, word)
	}
	return output

}

// Parser uses kagome for tokenization
func parser(item string) []any {
	// currently not functioning things
	// No multi verbs
	// No comprehensive map from augmentations to their respective definitions* particularly for verbs
	// the verb type sucks but it probably won't end up changing :c
	//iTells := []string{"ち", "り", "に", "み", "び", "き", "ぎ"}
	//teTells := []string{"て", "で"}
	//taTells := []string{"た", "だ"}
	//tTells := []string{"ん", "っ"}
	t := tokenizer.New()
	tokens := t.Tokenize(item)
	var output []any
	var currentVerb Verb
	for _, token := range tokens {
		if token.Class != tokenizer.KNOWN {
			continue
		}
		features := token.Features()
		pos := getEnglishPOS(features[0])

		if pos == "symbol" {
			if currentVerb.Word.DictForm != "" {
				switch currentVerb.Word.Form {
				case "U":
					currentVerb.Word.Form = "dictionary"
					break
				case "I":
					currentVerb.Word.Form = "conjunctive i"
					break
				case "A":
					currentVerb.Word.Form = "imperfective"
					break
				case "E Godan":
				case "E Ichidan":
					currentVerb.Word.Form = "imperative"
					break
				case "O":
					currentVerb.Word.Form = "volitional"
					break
				case "T":
					currentVerb.Word.Form = "Te or Ta"
					break
				default:
					//I forgot what the heck is actually going on here
					form := currentVerb.Word.Form
					if form != "Te" && form != "Ta" {
						currentVerb.Word.Form = ""
					}
					break
				}
				output = append(output, currentVerb)
			}
			currentVerb = Verb{}
			continue
		}

		if currentVerb.Word.DictForm == "" {
			if pos != "verb" {

				output = append(
					output,
					Word{
						Pos:      pos,
						DictForm: features[6],
						Form:     "",
						Word:     token.Surface,
					},
				)
				continue
			} else {

				dictform := ""
				if len(WordLookup(features[6])) > 0 {
					dictform = features[6]
				}

				if dictform == "" && strings.HasSuffix(features[6], "せる") {
					test := strings.Split(features[6], "せ")[0] + "す"
					if len(WordLookup(test)) > 0 {
						dictform = test
					}
				}
				if dictform == "" {
					dictform = features[6]
				}

				currentVerb = handleVerbs(features[6], features[4], features[5], token.Surface)
				continue
			}

		} else {
			switch currentVerb.Word.Form {
			case "U":
				result := handleDictionaryFormVerbs(currentVerb.Word, token.Surface)
				if result.Word.Form == "U" {
					currentVerb.Word.Form = "dictionary"
					output = append(output, Word{Word: token.Surface, Form: "", DictForm: features[6], Pos: pos})
				}
				output = append(output, currentVerb)
				break
			case "I":
				result := handleConjunctiveFormVerbs(currentVerb.Word, token.Surface, pos)
				if pos != "verb" {

					output = append(output, result)
					output = append(
						output,
						Word{
							Pos:      pos,
							DictForm: features[6],
							Form:     "",
							Word:     token.Surface,
						},
					)
					continue
				} else {
					currentVerb = handleVerbs(features[6], features[4], features[5], token.Surface)
					continue
				}
			case "A":
				result := handleAFormVerbs(currentVerb.Word, token.Surface)
				output = append(output, result)
				break
			case "E Godan":
				currentVerb.Word.Form = "Imperative"
				switch token.Surface {
				case "ば":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "conditional", PhraseStart: false})
					output = append(output, currentVerb)
					break
				case "いい":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "_ should ~", PhraseStart: false})
					output = append(output, currentVerb)
					break
				case "よかった":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "_ should have ~", PhraseStart: false})
					output = append(output, currentVerb)
					break
				default:
					output = append(output, Word{Word: token.Surface, Form: "", DictForm: features[6], Pos: pos})

				}
				break
			case "E Ichidan":
				// +ru doesn't need consideration since it'll detect that it's an ichidan verb, same with +rareru
				// therfore there's only single word set options
				currentVerb.Word.Form = "Imperative"
				switch token.Surface {
				case "れば":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "conditional", PhraseStart: false})
					output = append(output, currentVerb)
					break
				case "いい":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "_ should ~", PhraseStart: true})
					output = append(output, currentVerb)
					break
				case "よかった":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "_ should have ~", PhraseStart: true})
					output = append(output, currentVerb)
					break
				case "ろ":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "", PhraseStart: false})
					output = append(output, currentVerb)
					break
				case "よ":
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "", PhraseStart: false})
					output = append(output, currentVerb)
					break
				default:
					output = append(output, Word{Word: token.Surface, Form: "", DictForm: features[6], Pos: pos})

					break

				}
				break
			case "O":
				if token.Surface == "う" {
					currentVerb.Word.Word += "う"
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "lengthener", PhraseStart: false})
					output = append(output, currentVerb)
					currentVerb = Verb{}
				} else {
					output = append(output, currentVerb)
					currentVerb = Verb{}
				}
				break
			case "T":
				if strings.HasSuffix(token.Surface, "た") {
					currentVerb.Word.Form = "Ta"
					currentVerb.Word.Word += "た"
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "Past Tense Form", PhraseStart: false})
					continue
				} else if strings.HasSuffix(token.Surface, "だ") {
					currentVerb.Word.Form = "Ta"
					currentVerb.Word.Word += "だ"
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "Past Tense Form", PhraseStart: false})
					continue
				} else if strings.HasSuffix(token.Surface, "て") {
					currentVerb.Word.Form = "Te"
					currentVerb.Word.Word += "て"
					continue
				} else if strings.HasSuffix(token.Surface, "で") {
					currentVerb.Word.Form = "Te"
					currentVerb.Word.Word += "で"
					continue
				} else if token.Surface == "てる" {
					currentVerb.Word.Form = "Te"
					currentVerb.Word.Word += "てる"
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "Habitual", PhraseStart: false})
					currentVerb = Verb{}
					continue
				} else if token.Surface == "たら" {
					currentVerb.Word.Form = "Ta"
					currentVerb.Word.Word += "たら"
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "if/when", PhraseStart: false})
					currentVerb = Verb{}
					continue
				}
				fmt.Printf("AYO YOU DONE FUCKED IT UP GO TO case 'T' LINE 441: %s, %s\n", currentVerb.Word.Word, token.Surface)
				break

			case "Te":
				result := handleTeFormVerbs(currentVerb.Word, token.Surface, pos)
				output = append(output, result)
				break
			case "Ta":
				if currentVerb.Word.Word == "曲がりくねっ" {
					fmt.Println("bruh")
				}
				result := handleTaFormVerbs(currentVerb.Word, token.Surface)
				output = append(output, result)
				break
			case "I + T":
				if strings.HasSuffix(token.Surface, "た") {
					currentVerb.Word.Form = "Ta"
					currentVerb.Word.Word += "た"
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "Past Tense Form", PhraseStart: false})
				} else if strings.HasSuffix(token.Surface, "だ") {
					currentVerb.Word.Form = "Ta"
					currentVerb.Word.Word += "だ"
					currentVerb.Augmentations = append(currentVerb.Augmentations, Augmentation{Description: "Past Tense Form", PhraseStart: false})
				} else if strings.HasSuffix(token.Surface, "て") {
					currentVerb.Word.Form = "Te"
					currentVerb.Word.Word += "て"
				} else if strings.HasSuffix(token.Surface, "で") {
					currentVerb.Word.Form = "Te"
					currentVerb.Word.Word += "で"
				}
				result := handleConjunctiveFormVerbs(currentVerb.Word, token.Surface, pos)
				output = append(output, result)

			default:
				fmt.Println("AYO WTF IS THIS SHIT")
				break

			}
			currentVerb = Verb{}
		}

	}
	return output
}
func getEnglishPOS(s string) string {
	if result, exists := posMap[s]; exists {
		return result
	}
	return ""
}
func handleDictionaryFormVerbs(verb Word, nextWord string) Verb {
	var verbAugmentations = map[string]Augmentation{
		"な":   {"Negative imperative (don't ~)", false},
		"の":   {"Emphatic nominalization", false},
		"こと":  {"Abstract nominalization", true},
		"べき":  {"Idealistic 'should'", false},
		"まい":  {"Formal negative volitional", false},
		"はず":  {"Expected outcome", false},
		"なら":  {"Contextual 'if'", false},
		"つもり": {"Planned action", false},
		"と":   {"Definite conditional or quotation starter", true},
		"前":   {"Before ~ing", true}, //needs ni
		"みたい": {"Seems like ~", true},
		"そう":  {"Hearsay reporting", true},
		"らしい": {"Appearance-based inference", true},
		// will need to eventually include tomonaku
	}
	aug, exists := verbAugmentations[nextWord]
	if !exists {
		return Verb{Word: verb, Augmentations: []Augmentation{}}

	}

	// Create modified word
	modified := Word{
		Pos:      verb.Pos,
		DictForm: verb.DictForm,
		Form:     "dictionary",
		Word:     verb.Word + nextWord,
	}

	return Verb{Word: modified, Augmentations: []Augmentation{aug}}
}
func handleConjunctiveFormVerbs(verb Word, nextWord string, pos string) Verb {
	// Map of single-word conjunctive form augmentations
	conjunctiveAugmentations := map[string]Augmentation{
		"たい": {"Desire (い-adjective)", false},
		//	"たがる":  {"Desire (五段 verb)", false},
		"はしない": {"Strong negative desire", false},
		"ながら":  {"While ~ing", false},
		"がち":   {"Tends to ~ (often with だ)", false},
		"かた":   {"Way of ~ing", false},
		"方":    {"Way of ~ing (kanji)", false},      // Alternate form
		"そう":   {"Appearance (looks like ~)", true}, // Often with だ
		"つつ":   {"Continuing to ~", false},
		//	"やがる":  {"Rude/hostile nuance", false},
		//	"すぎる":  {"Excess (五段 verb)", false},
		"やすい": {"Easy to ~ (い-adjective)", false},
		"にくい": {"Difficult to ~ (い-adjective)", false},
		"もの":  {"Noun for verb target", false},
		// Polite ます forms
		"ます":     {"Polite non-past", false},
		"ません":    {"Polite negative", false},
		"ました":    {"Polite past", false},
		"ませんでした": {"Polite past negative", false},
		"ましょう":   {"Polite volitional", false},
		"まして":    {"Polite て-form", false},
		"ますれば":   {"Polite conditional (archaic)", false},
		"なさい":    {"Polite imperative", false},
		"な":      {"Casual imperative", false},
	}

	// Check for matching augmentations first
	if aug, exists := conjunctiveAugmentations[nextWord]; exists {
		modified := Word{
			Pos:      verb.Pos,
			DictForm: verb.DictForm,
			Form:     "conjunctive",
			Word:     verb.Word + nextWord,
		}
		return Verb{Word: modified, Augmentations: []Augmentation{aug}}
	}

	if pos == "verb" {
		modified := Word{
			Pos:      verb.Pos,
			DictForm: verb.DictForm,
			Form:     "conjunctive",
			Word:     verb.Word,
		}
		return Verb{
			Word: modified,
			Augmentations: []Augmentation{
				{"Compounding verb", false},
			},
		}
	}

	// No match found
	return Verb{Word: verb, Augmentations: []Augmentation{}}
}

func handleTeFormVerbs(verb Word, nextWord string, pos string) Verb {
	teAugmentations := map[string]Augmentation{
		//"いく":     {"Changing state (e.g., ~ていく)", false},
		//"いる":     {"Continuous/habitual action", false},
		"る":    {"Continuous (colloquial short form)", false},
		"おく":   {"Preparatory action", false},
		"く":    {"Preparatory (colloquial short form)", false},
		"しまう":  {"Completed action", false},
		"よかった": {"I'm glad that ~", false},
		//"みる":     {"Try ~ and see", false},
		"ほしい":    {"Favour request (e.g., ~てほしい)", false},
		"ある":     {"Changed state (resultative)", false},
		"から":     {"After ~ing", false},
		"くる":     {"State change (e.g., ~てくる)", false},
		"は":      {"Suggestive (must not ~)", false},
		"もかまわたい": {"Permissive (colloquial)", false},
		"いい":     {"Permission (e.g., ~ていい)", false},
		"すみません":  {"Apologetic (e.g., ~ですみません)", false},
		"も":      {"Even though ~", false},
		"ください":   {"Polite request", false},
		"あげる":    {"Benefit (giving)", false},
		//"くれる":    {"Benefit (receiving)", false},
		"もらう": {"Benefit (receiving)", false},
	}

	// Check for te-form augmentations first
	if aug, exists := teAugmentations[nextWord]; exists {
		modified := Word{
			Pos:      verb.Pos,
			DictForm: verb.DictForm,
			Form:     "Te",
			Word:     verb.Word + nextWord,
		}
		return Verb{Word: modified, Augmentations: []Augmentation{aug}}
	}

	// Compound verb detection (simplified)
	// In real usage, verify nextWord is a main verb via POS tag
	if pos == "verb" {
		modified := Word{
			Pos:      verb.Pos,
			DictForm: verb.DictForm,
			Form:     "Te",
			Word:     verb.Word + nextWord,
		}
		return Verb{
			Word: modified,
			Augmentations: []Augmentation{
				{"Compound verb", false},
			},
		}
	}

	return Verb{Word: verb, Augmentations: []Augmentation{}}
}
func handleTaFormVerbs(verb Word, nextWord string) Verb {
	taAugmentations := map[string]Augmentation{
		"から":    {"Reason for next clause", false},
		"り":     {"~ etc. (often paired with する)", false},
		"ら":     {"Conditional (if/when ~, colloquial)", false},
		"ばかり":   {"Just happened", false},
		"ほうがいい": {"Suggestive advice", false},
		"ことが":   {"Past experience (ことがある)", false},
		"だろう":   {"Past presumptive", false},
		"ろう":    {"Past volitional (rare)", false},
		// Special cases with implied continuations
		"ことがある": {"Past experience", false}, // If your tokenizer allows 2-word lookups
	}

	// Standard single-word check
	if aug, exists := taAugmentations[nextWord]; exists {
		modified := Word{
			Pos:      verb.Pos,
			DictForm: verb.DictForm,
			Form:     "Ta",
			Word:     verb.Word + nextWord,
		}
		return Verb{Word: modified, Augmentations: []Augmentation{aug}}
	}

	// No matching augmentation
	return Verb{Word: verb, Augmentations: []Augmentation{}}
}
func handleAmbiguousTeTaIFormVerbs() {

}
func handleAFormVerbs(verb Word, nextWord string) Verb {
	// Negative A-form augmentations (common to all verbs)
	negativeAugmentations := map[string]Augmentation{
		"ない":      {"Negative (い-adjective form)", false},
		"ないで":     {"Without ~ing", false},
		"なくて":     {"Negative て-form", false},
		"なかった":    {"Past negative", false},
		"なけれ":     {"Negative conditional stem", false},
		"なかろ":     {"Negative volitional stem", false},
		"ないでください": {"Negative request", false},
		"ないと":     {"Must (if not...)", false},
		"なくては":    {"Must (formal)", false},
		"なくちゃ":    {"Must (casual)", false},
		"なければ":    {"Must (standard)", false},
		"なきゃ":     {"Must (colloquial)", false},
		"ず":       {"Classical negative", false},
		"ずに":      {"Without ~ing (formal)", false},
	}

	// Check negative forms first
	if aug, exists := negativeAugmentations[nextWord]; exists {
		modified := Word{
			Pos:      verb.Pos,
			DictForm: verb.DictForm,
			Form:     "negative",
			Word:     verb.Word + nextWord,
		}
		return Verb{Word: modified, Augmentations: []Augmentation{aug}}
	}

	// Determine verb type from POS tag (example logic)
	isGodan := strings.HasPrefix(verb.Pos, "v5") // e.g., v5u = godan u-verb
	isIchidan := verb.Pos == "v1"                // ichidan verb

	// Causative/passive augmentations (verb-type specific)
	var causPassAug map[string]Augmentation
	switch {
	case isGodan:
		causPassAug = map[string]Augmentation{
			"れる":   {"Passive/Honorific (五段)", false},
			"せる":   {"Causative (五段)", false},
			"す":    {"Causative alternative (五段)", false},
			"せられる": {"Causative-passive (五段)", false},
			"される":  {"Causative-passive alternative (五段)", false},
		}
	case isIchidan:
		causPassAug = map[string]Augmentation{
			"られる":   {"Passive/Honorific (一段)", false},
			"させる":   {"Causative (一段)", false},
			"さす":    {"Causative alternative (一段)", false},
			"させられる": {"Causative-passive (一段)", false},
		}
	default:
		// Not a verb type we handle for caus/pass
		return Verb{Word: verb, Augmentations: []Augmentation{}}
	}

	// Check causative/passive forms
	if aug, exists := causPassAug[nextWord]; exists {
		modified := Word{
			Pos:      verb.Pos,
			DictForm: verb.DictForm,
			Form:     "causative-passive",
			Word:     verb.Word + nextWord,
		}
		return Verb{Word: modified, Augmentations: []Augmentation{aug}}
	}

	// No matching augmentation
	return Verb{Word: verb, Augmentations: []Augmentation{}}
}

// CheckTitle checks if a line matches the pattern of a wiki link title
func checkTitle(path, title, test string) bool {
	return title == fmt.Sprintf("[%s](%s\\%s.md)\n", test, path, test)
}

// IntakeContent loads and processes text files from the 'New Content' directory
func intakeContent() map[string]string {
	output := make(map[string]string)

	files, err := filepath.Glob(filepath.Join(newContentPath, "*.txt"))
	if err != nil || len(files) == 0 {
		fmt.Printf("No new sources found. Place .txt files in %s to begin\n", newContentPath)
		return nil
	}

	for _, txtFile := range files {
		name := replaceSpaces(strings.TrimSuffix(filepath.Base(txtFile), ".txt"))

		lines, err := readLines(txtFile)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", txtFile, err)
			continue
		}
		blob := ""
		for _, line := range lines {
			// Remove Latin characters and spaces
			re := regexp.MustCompile(`[A-Za-z0-9]`)
			temp := re.ReplaceAllString(line, "")
			temp = strings.ReplaceAll(temp, " ", "")
			if temp != "" {
				blob += temp + "\n"
			}
		}

		output[name] = blob

		thisContentMd := filepath.Join(contentPath, name+".md")
		err = os.WriteFile(thisContentMd, []byte(blob), 0644)
		if err != nil {
			fmt.Printf("Error writing to %s: %v\n", thisContentMd, err)
		}
	}

	return output
}

// GetSentences extracts sentences from processed content
func getSentences() map[string][]string {
	punctuation := []string{"\n", ".", "?", "!", "〪", "。", "〭", "！", "．", "？"}
	sources := intakeContent()
	output := make(map[string][]string)

	if sources == nil {
		return output
	}

	for name := range sources {
		output[name] = []string{}
	}

	for name, content := range sources {
		sentence := ""
		for _, char := range content {
			isPunctuation := false
			for _, p := range punctuation {
				if string(char) == p {
					isPunctuation = true
					break
				}
			}

			if !isPunctuation {
				sentence += string(char)
			} else {
				if sentence != "" {
					output[name] = append(output[name], sentence)
					sentence = ""
				}
			}
		}
		if sentence != "" { // Add remaining content
			output[name] = append(output[name], sentence)
		}
	}

	return output
}

// SentenceToWordString converts a sentence to a string of linked words
func sentenceToWordString(sentence string) string {
	// Word punctuation to remove
	words := parser(sentence)
	var tempArray []string
	for _, word := range words {
		switch v := word.(type) {
		case Word:
			tempArray = append(tempArray, fmt.Sprintf("[%s](%s\\%s.md)", v.Word, wordsPath, v.DictForm))
		case Verb:
			tempArray = append(tempArray, fmt.Sprintf("[%s](%s\\%s.md)", v.Word.Word, wordsPath, v.Word.DictForm))
		}
	}

	return strings.Join(tempArray, " ")
}

// WordToKanjiString converts a word to a string with linked kanji
func wordToKanjiString(word string) string {
	var result strings.Builder

	for _, c := range word {
		charStr := string(c)
		if containsRune(kanjiSet, charStr) || containsRune(oldKanji, charStr) {
			result.WriteString(fmt.Sprintf("[%s](%s\\%s.md)", charStr, kanjiPath, charStr))
		} else {
			result.WriteString(charStr)
		}
	}

	return result.String()
}

// Helper function to check if a rune is in a slice
func containsRune(slice []string, r string) bool {
	return slices.Contains(slice, r)
}

// ReplaceSpaces replaces spaces in tags with underscores
func replaceSpaces(tag string) string {
	return strings.ReplaceAll(tag, " ", "_")
}

// === File Management Functions ===

// AppendContent appends a new content entry to the main content markdown file
func appendContent(name string) {
	f, err := os.OpenFile(contentMd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", contentMd, err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("[%s](%s\\%s.md)\n", name, contentPath, name))
	if err != nil {
		fmt.Printf("Error writing to %s: %v\n", contentMd, err)
	}
}
func readLines(s string) ([]string, error) {
	file, err := os.Open(s)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var output []string
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}
	return output, nil
}

// AppendSentence checks if a sentence exists in the sentences index
func appendSentence(sentence string) bool {
	lines, err := readLines(sentencesMd)
	if err != nil {
		return false
	}

	for _, line := range lines {
		if checkTitle(sentencesPath, line+"\n", sentence) {
			return true
		}
	}
	return false
}

// AppendKanji checks if a kanji exists in the kanji index
func appendKanji(kanji string) bool {
	lines, err := readLines(kanjiMd)
	if err != nil {
		return false
	}
	for _, line := range lines {
		if checkTitle(kanjiPath, line+"\n", kanji) {
			return true
		}
	}
	return false
}

// AddNewStuff adds new entries to respective index files
func addNewStuff(kl, wl, sl []string) {
	// Update kanji index
	if len(kl) > 0 {
		f, err := os.OpenFile(kanjiMd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", kanjiMd, err)
		} else {
			for _, k := range kl {
				f.WriteString(k)
			}
			f.Close()
		}
	}

	// Update words index
	if len(wl) > 0 {
		f, err := os.OpenFile(wordsMd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", wordsMd, err)
		} else {
			for _, w := range wl {
				f.WriteString(w)
			}
			f.Close()
		}
	}

	// Update sentences index
	if len(sl) > 0 {
		f, err := os.OpenFile(sentencesMd, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", sentencesMd, err)
		} else {
			for _, s := range sl {
				f.WriteString(s)
			}
			f.Close()
		}
	}
}

// AppendWord checks if a word exists in the words index
func appendWord(word string) bool {
	lines, err := readLines(wordsMd)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", wordsMd, err)
		return false
	}
	for _, line := range lines {
		if checkTitle(wordsPath, line+"\n", word) {
			return true
		}
	}
	return false
}

// === Tag Editing Functions ===

// EditKanjiTags adds current content tag to a kanji's metadata
func editKanjiTags(item string) {
	path := filepath.Join(kanjiPath, item+".md")
	lines, err := readLines(path)
	if err != nil {
		return
	}
	for i, line := range lines {
		if strings.Contains(line, "Tags: ") {
			lines[i] = fmt.Sprintf("[%s](%s\\%s.md) ", currentName, contentPath, currentName)
			break
		}
	}

	err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Printf("Error writing to %s: %v\n", path, err)
	}
}

// EditSentenceTags adds current content tag to a sentence's metadata
func editSentenceTags(item string) {
	path := filepath.Join(sentencesPath, item+".md")
	lines, err := readLines(path)
	if err != nil {
		return
	}
	for i, line := range lines {
		if strings.Contains(line, "Tags: ") {
			lines[i] = fmt.Sprintf("[%s](%s\\%s.md) ", currentName, contentPath, currentName)
			break
		}
	}

	err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Printf("Error writing to %s: %v\n", path, err)
	}
}

// EditWordsTags adds current content tag to a word's metadata
func editWordsTags(item string) {
	path := filepath.Join(wordsPath, item+".md")
	lines, err := readLines(path)
	if err != nil {
		return
	}
	for i, line := range lines {
		if strings.Contains(line, "Tags: ") {
			lines[i] = fmt.Sprintf("[%s](%s\\%s.md) ", strings.TrimSpace(line), currentName, contentPath, currentName)
			break
		}
	}

	err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		fmt.Printf("Error writing to %s: %v\n", path, err)
	}
}

// === Data Processing Functions ===

// WriteToKanji writes kanji to index and returns list of existing items
func writeToKanji(lst []string) []string {
	var existing []string
	for _, k := range lst {
		if appendKanji(k) {
			existing = append(existing, k)
		}
	}
	return existing
}

// WriteToWords writes words to index and returns list of existing items
func writeToWords(lst []string) []string {
	var existing []string
	for _, w := range lst {
		if appendWord(w) {
			existing = append(existing, w)
		}
	}
	return existing
}

// WriteToSentences writes sentences to index and returns list of existing items
func writeToSentences(lst []string) []string {
	var existing []string
	for _, s := range lst {
		if appendSentence(s) {
			existing = append(existing, s)
		}
	}
	return existing
}

// === Flashcard Creation Functions ===

// SentenceCard generates sentence flashcard markdown file
func sentenceCard(data SentenceData) {
	content := []string{
		"TARGET DECK: Sentences",
		"START",
		"Basic",
		sentenceToWordString(data.Sentence),
		fmt.Sprintf("Back: %s", data.Translation),
		fmt.Sprintf("Tags: [%s](%s\\%s.md)", currentName, pathing.ContentPath, currentName),
		"",
		"END",
	}
	writeCard(strings.Join(content, "\n"), filepath.Join(sentencesPath, data.Sentence+".md"))
}
func debugger(a any) {
	debug.PrintStack() // similar to a breakpoint: prints the stack trace
	fmt.Println(a)
}

// SentenceCardSkipped generates sentence flashcard without translation
func sentenceCardSkipped(sentence string) {
	content := []string{
		"TARGET DECK: Sentences",
		"START",
		"Basic",
		sentenceToWordString(sentence),
		"Back: ",
		fmt.Sprintf("Tags: [%s](%s\\%s.md)", currentName, pathing.ContentPath, currentName),
		"",
		"END",
	}

	writeCard(strings.Join(content, "\n"), filepath.Join(sentencesPath, sentence+".md"))
}

// WordCard generates word flashcard markdown file
func wordCard(data []WordData) {
	definitions := ""
	readings := ""
	for _, d := range data {
		definitions += fmt.Sprintf("%s, ", d.Definitions)
	}
	for _, r := range data {
		if !strings.Contains(readings, r.Reading) {
			readings += fmt.Sprintf("%s, ", r.Reading)
		}
	}
	content := []string{
		"TARGET DECK: Words",
		"START",
		"Basic",
		wordToKanjiString(data[0].Word),
		fmt.Sprintf("Back: %s", definitions),
		readings,
		fmt.Sprintf("Tags: [%s](%s\\%s.md)", currentName, pathing.ContentPath, currentName),
		"",
		"END",
	}

	writeCard(strings.Join(content, "\n"), filepath.Join(wordsPath, data[0].Word+".md"))
}

func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}

	return nil
}

// KanjiCard generates kanji flashcard markdown file
func kanjiCard(data KanjiData) {
	content := []string{
		"TARGET DECK: Kanji",
		"START",
		"Basic",
		fmt.Sprintf("%s, %d", data.Kanji, data.Strokes),
		fmt.Sprintf("Back: %s", data.Keyword),
		data.Readings,
		data.Radicals,
		fmt.Sprintf("Tags: [%s](%s\\%s.md)", currentName, pathing.ContentPath, currentName),
		"",
		"END",
	}

	writeCard(strings.Join(content, "\n"), filepath.Join(kanjiPath, data.Kanji+".md"))
}

// === Data Fetching Functions (Dummy versions) ===

// KanjiData fetches kanji data (dummy function replacing Jisho API)
func fetchKanjiData(kanji string) KanjiData {
	// Dummy implementation - would be replaced with actual implementation
	result := KanjiLookup(kanji)
	if result.Kanji == "" && result.Strokes == 0 {
		return KanjiData{
			Kanji:    kanji,
			Keyword:  "",
			Readings: "",
			Strokes:  0,
			Radicals: "",
		}
	}
	return result
}

// WordData fetches word data (dummy function replacing Jisho API)
func fetchWordData(word string) []WordData {
	result := WordLookup(word)
	dummy := WordData{
		Word:        word,
		Definitions: "(meaning1, meaning2)",
		Reading:     word,
	}
	if len(result) == 0 {
		return []WordData{dummy}
	}
	return result
}

// SentenceData generates sentence data with translation
func fetchSentenceData(sentence string) SentenceData {
	return SentenceData{
		Sentence:    sentence,
		Translation: "", // currently not going to work until a better source for translations is found
	}
}

// === Utility Functions ===

// WriteCard writes formatted content to a markdown file
func writeCard(content, path string) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing to %s: %v\n", path, err)
	}
}

// WriteSentencesToContentMd appends sentence links to content markdown file
func writeSentencesToContentMd(sentences []string, path string) {
	f, err := os.OpenFile(filepath.Join(contentPath, currentName+".md"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", filepath.Join(contentPath, currentName+".md"), err)
		return
	}
	defer f.Close()

	f.WriteString("\n")
	for _, s := range sentences {
		f.WriteString(fmt.Sprintf("[%s](%s\\%s.md)\n", s, path, s))
	}
}

// === CSV Export Functions ===

// FilesToFlashcardClass converts markdown flashcard files to Flashcard dictionaries
func filesToFlashcardClass(filePaths []string) []FlashcardDict {
	var output []FlashcardDict

	for _, path := range filePaths {
		lines, _ := readLines(path)
		for i := range lines {
			lines[i] = strings.TrimSpace(lines[i])
		}

		// Find Basic section
		frontIndex := -1
		for i, line := range lines {
			if line == "Basic" {
				frontIndex = i + 1
				break
			}
		}
		if frontIndex == -1 {
			continue
		}

		// Find Back section
		backIndex := -1
		for i, line := range lines {
			if strings.HasPrefix(line, "Back:") {
				backIndex = i
				break
			}
		}
		if backIndex == -1 {
			continue
		}

		// Find Tags section
		tagIndex := -1
		for i, line := range lines {
			if strings.HasPrefix(line, "Tags:") {
				tagIndex = i
				break
			}
		}
		if tagIndex == -1 {
			continue
		}

		// Extract content
		front := strings.Join(lines[frontIndex:backIndex], "\n")
		back := strings.Join(lines[backIndex:tagIndex], "\n")
		back = strings.Replace(back, "Back: ", "", 1)

		output = append(output, FlashcardDict{
			Front: front,
			Back:  back,
			Cloze: "", // We're not using cloze in this implementation
		})
	}

	return output
}

// FlashcardsToCSV exports flashcards to CSV files for Anki import
func flashcardsToCSV(flashcards []FlashcardDict, csvFilePath, clozePath string) {
	// Create regular CSV file
	regularFile, err := os.Create(csvFilePath)
	if err != nil {
		fmt.Printf("Error creating %s: %v\n", csvFilePath, err)
		return
	}
	defer regularFile.Close()

	regularWriter := csv.NewWriter(regularFile)
	defer regularWriter.Flush()

	// Write header
	err = regularWriter.Write([]string{"Front", "Back"})
	if err != nil {
		fmt.Printf("Error writing to %s: %v\n", csvFilePath, err)
		return
	}

	// Create cloze CSV file
	clozeFile, err := os.Create(clozePath)
	if err != nil {
		fmt.Printf("Error creating %s: %v\n", clozePath, err)
		return
	}
	defer clozeFile.Close()

	clozeWriter := csv.NewWriter(clozeFile)
	defer clozeWriter.Flush()

	// Write header
	err = clozeWriter.Write([]string{"Cloze", "Back"})
	if err != nil {
		fmt.Printf("Error writing to %s: %v\n", clozePath, err)
		return
	}

	// Write data
	for _, card := range flashcards {
		err = regularWriter.Write([]string{card.Front, card.Back})
		if err != nil {
			fmt.Printf("Error writing to %s: %v\n", csvFilePath, err)
		}

		if card.Cloze != "" {
			err = clozeWriter.Write([]string{card.Cloze, card.Back})
			if err != nil {
				fmt.Printf("Error writing to %s: %v\n", clozePath, err)
			}
		}
	}
}

// MakeCSVs generates CSV files from all markdown flashcards
func makeCSVs() {
	// Get all markdown files
	inputSentences, _ := filepath.Glob(filepath.Join(sentencesPath, "*.md"))
	inputWords, _ := filepath.Glob(filepath.Join(wordsPath, "*.md"))
	inputKanji, _ := filepath.Glob(filepath.Join(kanjiPath, "*.md"))

	// Process sentences
	flashcardsToCSV(
		filesToFlashcardClass(inputSentences),
		filepath.Join(csvPath, "Sentences.csv"),
		filepath.Join(csvPath, "Sentences_cloze.csv"),
	)

	// Process words
	flashcardsToCSV(
		filesToFlashcardClass(inputWords),
		filepath.Join(csvPath, "Words.csv"),
		filepath.Join(csvPath, "Words_cloze.csv"),
	)

	// Process kanji
	flashcardsToCSV(
		filesToFlashcardClass(inputKanji),
		filepath.Join(csvPath, "Kanji.csv"),
		filepath.Join(csvPath, "Kanji_cloze.csv"),
	)
}

// MakeNotes generates notes from source content
func makeNotes() {
	// Get all sentences
	sentencesBySource := getSentences()
	if len(sentencesBySource) == 0 {
		fmt.Println("No content to process")
		return
	}
	for source, sentences := range sentencesBySource {
		currentName = source
		appendContent(source)

		var kanjiList []string
		var wordList []Word
		var wordListString []string

		// Process sentences
		for _, sentence := range sentences {
			fmt.Println(sentence)
			words := parser(sentence)
			fmt.Println(words)
			// Extract kanji
			for _, word := range words {
				switch v := word.(type) {
				case Word:
					for _, c := range v.DictForm {

						if containsRune(kanjiSet, string(c)) && !containsRune(kanjiList, string(c)) && !containsRune(oldKanji, string(c)) {
							kanjiList = append(kanjiList, string(c))
						}
					}
				case Verb:
					for _, c := range v.Word.DictForm {

						if containsRune(kanjiSet, string(c)) && !containsRune(kanjiList, string(c)) && !containsRune(oldKanji, string(c)) {
							kanjiList = append(kanjiList, string(c))
						}
					}
				}
			}

			// Extract words
			for _, word := range words {
				switch v := word.(type) {
				case Word:
					if !containsRune(wordListString, v.DictForm) && !containsRune(oldWords, v.DictForm) {
						wordList = append(wordList, v)
						wordListString = append(wordListString, v.DictForm)
					}
				case Verb:
					if !containsRune(wordListString, v.Word.DictForm) && !containsRune(oldWords, v.Word.DictForm) {
						wordList = append(wordList, v.Word)
						wordListString = append(wordListString, v.Word.DictForm)
					}
				}
			}
		}

		// Create flashcards
		var wg sync.WaitGroup

		// Process kanji
		for _, k := range kanjiList {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				kData := fetchKanjiData(k)
				kanjiCard(kData)
				editKanjiTags(k)
			}(k)
		}

		// Process words
		for _, w := range wordList {
			wg.Add(1)
			go func(w Word) {
				defer wg.Done()

				wData := fetchWordData(w.DictForm)
				wordCard(wData)
				//editWordsTags(w)
			}(w)
		}

		// Process sentences
		for _, s := range sentences {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				if skipSentences {
					sentenceCardSkipped(s)
				} else {
					sData := fetchSentenceData(s)
					sentenceCard(sData)
				}
				//editSentenceTags(s)
			}(s)
		}

		wg.Wait()

		// Update old lists
		oldKanji = append(oldKanji, kanjiList...)
		oldWords = append(oldWords, wordListString...)

		// Add new entries to index files
		var kanjiEntries []string
		var wordEntries []string
		var sentenceEntries []string

		for _, k := range kanjiList {
			kanjiEntries = append(kanjiEntries, fmt.Sprintf("[%s](%s\\%s.md)\n", k, kanjiPath, k))
		}

		for _, w := range wordList {
			wordEntries = append(wordEntries, fmt.Sprintf("[%s](%s\\%s.md)\n", w, wordsPath, w))
		}

		for _, s := range sentences {
			sentenceEntries = append(sentenceEntries, fmt.Sprintf("[%s](%s\\%s.md)\n", s, sentencesPath, s))
		}

		addNewStuff(kanjiEntries, wordEntries, sentenceEntries)
		writeSentencesToContentMd(sentences, pathing.SentencesPath)
	}
}

// === User Interaction Functions ===

// AskForTranslations prompts user about including sentence translations
func askForTranslations(n int) {
	fmt.Println("This program uses a translation service for sentence translations.")
	fmt.Println("Include sentence translations? (Y/N)")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		answer := strings.ToLower(scanner.Text())
		if answer == "y" {
			fmt.Println("Including sentence translations.")
			ready(n)
			return
		} else if answer == "n" {
			fmt.Println("Skipping sentence translations.")
			skipSentences = true
			ready(n)
			return
		}
		fmt.Println("Please enter Y or N.")
	}
}

// AskForCSVs prompts user about CSV generation
func askForCSVs() {
	fmt.Println("Generate CSV files for Anki? (Y/N)")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		answer := strings.ToLower(scanner.Text())
		if answer == "y" {
			justCSVs()
			return
		} else if answer == "n" {
			askForTranslations(0)
			return
		}
		fmt.Println("Please enter Y or N.")
	}
}

// JustCSVs handles CSV-only generation option
func justCSVs() {
	fmt.Println("Generate only CSV files (no new notes)? (Y/N)")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		answer := strings.ToLower(scanner.Text())
		if answer == "y" {
			ready(2)
			return
		} else if answer == "n" {
			askForTranslations(1)
			return
		}
		fmt.Println("Please enter Y or N.")
	}
}

// Ready is final confirmation before processing
func ready(n int) {
	messages := map[int]string{
		0: "Generating only notes",
		1: "Generating notes and CSV files",
		2: "Generating only CSV files",
	}
	fmt.Println(messages[n])
	fmt.Print("Confirm (Y/N): ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		confirm := strings.ToLower(scanner.Text())
		if confirm == "y" {
			if n == 0 || n == 1 {
				makeNotes()
			}
			if n == 1 || n == 2 {
				makeCSVs()
			}
			return
		} else if confirm == "n" {
			fmt.Println("Restarting...")
			askForCSVs()
			return
		}
		fmt.Println("Please enter Y or N.")
	}
}
func buildKanjiIndex(kd Kanjidic2) map[string]KanjiData {
	idx := make(map[string]KanjiData, len(kd.Characters))
	for _, char := range kd.Characters {
		var on, kun, meanings []string
		// collect readings & meanings
		for _, group := range char.ReadingMeaning.Groups {
			for _, r := range group.Readings {
				switch r.Type {
				case "ja_on":
					on = append(on, r.Value)
				case "ja_kun":
					kun = append(kun, r.Value)
				}
			}
			for _, m := range group.Meanings {
				meanings = append(meanings, m.Value)
			}

		}
		// use the first stroke count if multiple provided
		strokes := 0
		if len(char.Misc.StrokeCount) > 0 {
			strokes = char.Misc.StrokeCount[0]
		}
		if len(meanings) == 0 {
			meanings = append(meanings, "")
		}

		idx[char.Literal] = KanjiData{
			Kanji:    char.Literal,
			Keyword:  meanings[0], // pick the first meaning as “keyword”
			Readings: fmt.Sprintf("%v|%v", on, kun),
			Strokes:  strokes,
			Radicals: "", // Kanjidic2 doesn’t include radicals by default
		}
	}
	return idx
}
func buildWordIndex(jd JMdict) map[string][]WordData {
	idx := make(map[string][]WordData)
	for _, entry := range jd.Entries {
		// collect all kanji-forms and readings on this entry
		var readings []string
		for _, r := range entry.REle {
			readings = append(readings, r.Reb)
		}
		// collect senses
		var glosses []string
		for _, sense := range entry.Sense {
			for _, g := range sense.Gloss {
				glosses = append(glosses, g.Text)
			}
		}
		temp := ""
		for _, r := range readings {
			if r == "" {
				continue
			}
			temp = fmt.Sprintf("%s, ", r)
		}
		// for each distinct “word” form (kanji or kana only), store a WordData
		for _, k := range entry.KEle {
			idx[k.Keb] = append(idx[k.Keb], WordData{
				Word:        k.Keb,
				Definitions: fmt.Sprintf("%v", glosses),
				Reading:     temp, // pick the first reading
			})
		}
		for _, r := range entry.REle {
			// if it has no kanji restriction, also index it under purely-kana form
			if len(r.ReRestr) == 0 {
				idx[r.Reb] = append(idx[r.Reb], WordData{
					Word:        r.Reb,
					Definitions: fmt.Sprintf("%v", glosses),
					Reading:     r.Reb,
				})
			}
		}
	}
	return idx
}

var kanjiIdx map[string]KanjiData
var wordIdx map[string][]WordData

func KanjiLookup(s string) KanjiData {
	if kd, ok := kanjiIdx[s]; ok {
		return kd
	}

	return KanjiData{}
}
func WordLookup(s string) []WordData {
	if wd, ok := wordIdx[s]; ok {
		return wd
	}
	return []WordData{}
}

// Main is the entry point of the application
func main() {

	var kanjidic Kanjidic2
	if err := xml.Unmarshal(kanjiDic, &kanjidic); err != nil {
		panic(err)
	}
	var jmDict JMdict
	if err := xml.Unmarshal(jmDictData, &jmDict); err != nil {
		panic(err)
	}
	kanjiIdx = buildKanjiIndex(kanjidic)
	wordIdx = buildWordIndex(jmDict)
	/* test := "書かせられなければ、食べさせてやり、来させようとしたが、できずに来いと言われ、してしまった。"
		again := `夢のつづき追いかけていたはずなのに
	曲がりくねった細い道 人につまずく
	あの頃みたいにって戻りたい訳じゃないの
	無くしてきた空を探してる
	わかってくれますように犠牲になったような
	悲しい顔はやめてよ
	罪の最後は淚じゃないよ ずっと苦しく背負ってくんだ
	出口見えない感情迷路に誰を待ってるの?
	白いノ一トに綴ったようにもっと素直に吐き出したいよ
	何から 逃れたいんだ 現実ってやつ?
	叶えるために 生きてるんだって
	忘れちゃいそうな 夜の真ん中
	無難になんて やってられないから
	帰る場所もないの
	この想いを 消してしまうには
	まだ人生長いでしょ? (I'm on the way)
	懐かしくなる こんな痛みも歓迎じゃん
	謝らなくちゃいけないよね ah ごめんね
	うまく言えなくて心配かけたままだったね
	あの日かかえた全部 あしたかかえる全部
	順番つけたりはしないから
	わかってくれますようにそっと目を閉じたんだ
	見たくないものまで見えんだもん
	いらないウワサにちょっと初めて聞く発言どっち?
	2回会ったら友達だって? ウソはやめてね
	赤いハ一トが苛立つように身体ん中燃えているんだ
	ホントは 期待してんの 現実ってやつ?
	叶えるために 生きてるんだって
	叫びたくなるよ 聞こえていますか?
	無難になんて やってられないから
	帰る場所もないの
	優しさには いつも感謝してる
	だから強くなりたい (I'm on the way)
	進むために 敵も味方も歓迎じゃん
	どうやって次のドア開けるんだっけ? 考えてる?
	もう引き返せない 物語 始まってるんだ
	目を覚ませ 目を覚ませ
	この想いを 消してしまうには
	まだ人生長いでしょ?
	やり残してるコト やり直してみたいから
	もう一度ゆこう
	叶えるために 生きてるんだって
	叫びたくなるよ 聞こえていますか?
	無難になんて やってられないから
	帰る場所もないの
	優しさには いつも感謝してる
	だから強くなりたい (I'm on the way)
	懐かしくなる こんな痛みも歓迎じゃん`
		parser(test)
		parser("「コーヒーを飲まされたが、待たなくて話そうとしたら、払ったお金が足りず、歩けなくなり、家に行かせたのに、友達に笑われた。」")
		parser("買うの買い 買わない 買え 買おう 買って 買った 待つ 待ち 待たない 待て 待とう 待って 待った 取る 取り 取らない 取れ 取ろう 取って 取った 飲む 飲み 飲まない 飲め 飲もう 飲んで 飲んだ 聞く 聞き 聞かない 聞け 聞こう 聞いて 聞いた 泳ぐ 泳ぎ 泳がない 泳げ 泳ごう 泳いで 泳いだ 話す 話し 話さない 話せ 話そう 話して 話した 見る 見 見ない 見ろ 見よう 見て 見た 来る 来 来ない 来い 来よう 来て 来た する し しない しろ しよう して した")
		fmt.Println("\nTesting Dictionary forms")
		parser("行くらしい.彼は「行くな」と言ったが、行くのをやめることなく、行くと絶対成功するはずだ。行く前に準備すべきで、行くが早いか帰るつもりみたいだ。行くともなく駅へ向かい、行くらしい噂も聞いた。行くなら早く決めろ、行くまいと思っても無理だろう。")
		fmt.Println("\nTesting Volitional form")
		parser("行こう！彼と一緒に行くべきだと思うが、自分は行くまいと決めた。行こうとしたら、雨が降り始めた。")
		fmt.Println("\nTesting Imperative form")
		parser("行け！行けばいいのに、なぜ行かない？行けよ！行けばよかった…。行けるなら今すぐ行け！")
		fmt.Println("\nTesting A (nai) form")
		fmt.Print("\nTesting verbs")
		parser("母が子供に野菜を食べさせる。でも子供は食べさせられるのを嫌がり、無理やり食べさす（＝食べさせる）。時々食べせられる（＝食べさせられる）こともあるが、最近は食べされる（＝食べさせられる）と言う人もいる。先生が学生に作文を書かせる。学生は書かせられる（＝書かされる）のが苦手で、時々書かす（＝書かせる）代わりに絵を描く。昔は書かせられるより書かされるが使われた。")
		fmt.Println("\n Testing nonverbs")
		parser("彼は野菜を食べない。食べないで寝て、食べなくて元気がない。昨日も食べなかった。食べなければ痩せるが、食べなかろうか…？食べないでください！食べないといけない、食べなくてはいけない、食べなくちゃいけない、食べなければいけない、食べなきゃいけない！昔は食べず、食べずに生きていた。")
		fmt.Println("\n Full song Again by yui\n")
		fmt.Println(again)
		parser(again)
		fmt.Println("\nIntransitive Test")
		test = `ドアが開く前に、子供が起きてくる。窓が閉まるのを見たが、壊れる音がした。彼女は泣きながら走り去った。雨が降りそうで、電車が遅れがちだ。鍵がかかってしまい、中に入れなかった。疲れて寝てばかりいる。花が咲いてよかった。火が消えずにいる。時代が変わりつつある。風が止んだから出かけるつもりだ。彼の声が聞こえてほしい。ここに座ってもいい？ あの木が倒れそうだ。道が凍っていく。事件が解決したら知らせて。温度が下がりやすい。機械が動かなくなった。鳥が飛んでいく。霧が晴れてきた。夢が覚めるまいとした。波が静まるはずがない。息が続くかたを知りたい。光が増していく。涙が止まらなかった。時間が経つのは早い。 `
		parser(test) */
	askForCSVs()
}
