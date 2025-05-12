package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/TheShadowblast123/Japanese-Content-2-Md-And-Anki/path_handler"
	"github.com/ikawaha/kagome/tokenizer"
)

// Type definitions for data structures
type KanjiData struct {
	Kanji    string
	Keyword  string
	Readings string
	Strokes  int
	Radicals string
}

type WordData struct {
	Word        string
	Definitions string
	Reading     string
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

// Parser uses kagome for tokenization
func parser(item string) []string {
	t := tokenizer.New()
	tokens := t.Tokenize(item)
	var words []string
	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		words = append(words, token.Surface)
	}
	return words
}

// CheckTitle checks if a line matches the pattern of a wiki link title
func checkTitle(title, test string) bool {
	return title == fmt.Sprintf("[[%s]]\n", test)
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
	punctRe := regexp.MustCompile(`[[:punct:]]|！|＂|"|"|＃|＄|％|＆|＇|（|）|＊|＋|，|－|．|／|：|；|＜|＝|＞|？|＠|［|＼|］|＾|＿|｀|｛|｜|｝|～|、|。|〃|〄|々|〆|〇|〈|〉|《|》|「|」|『|』|【|】|〒|〓|〔|〕|〖|〗|〘|〙|〚|〛|〜|〝|〞|〟|〠|〡|〢|〣|〤|〥|〦|〧|〨|〩|〪|〭|〮|〯|〫|〬|〰|〱|〲|〳|〴|〵|〶|〷|〸|〹|〺|〻|〼|〽|〾|｟|｠|｡|｢|｣|､|･|〿`)

	wordsString := punctRe.ReplaceAllString(strings.Join(parser(sentence), " "), "")

	var tempArray []string
	for _, word := range strings.Fields(wordsString) {
		tempArray = append(tempArray, fmt.Sprintf("[%s](%s\\%s.md)", word, wordsPath, word))
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
	for _, item := range slice {
		if item == r {
			return true
		}
	}
	return false
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

	_, err = f.WriteString(fmt.Sprintf("[[%s]]\n", name))
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
		if checkTitle(line+"\n", sentence) {
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
		if checkTitle(line+"\n", kanji) {
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
		if checkTitle(line+"\n", word) {
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
			lines[i] = fmt.Sprintf("%s [[%s]]", strings.TrimSpace(line), currentName)
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
			lines[i] = fmt.Sprintf("%s [[%s]]", strings.TrimSpace(line), currentName)
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
			lines[i] = fmt.Sprintf("%s [[%s]]", strings.TrimSpace(line), currentName)
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
		fmt.Sprintf("Tags: [[%s]]", currentName),
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
		fmt.Sprintf("Tags: [[%s]]", currentName),
		"",
		"END",
	}

	writeCard(strings.Join(content, "\n"), filepath.Join(sentencesPath, sentence+".md"))
}

// WordCard generates word flashcard markdown file
func wordCard(data WordData) {
	content := []string{
		"TARGET DECK: Words",
		"START",
		"Basic",
		wordToKanjiString(data.Word),
		fmt.Sprintf("Back: %s", data.Definitions),
		data.Reading,
		fmt.Sprintf("Tags: [[%s]]", currentName),
		"",
		"END",
	}

	writeCard(strings.Join(content, "\n"), filepath.Join(wordsPath, data.Word+".md"))
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
		fmt.Sprintf("Tags: [[%s]]", currentName),
		"",
		"END",
	}

	writeCard(strings.Join(content, "\n"), filepath.Join(kanjiPath, data.Kanji+".md"))
}

// === Data Fetching Functions (Dummy versions) ===

// KanjiData fetches kanji data (dummy function replacing Jisho API)
func fetchKanjiData(kanji string) KanjiData {
	// Dummy implementation - would be replaced with actual implementation
	return KanjiData{
		Kanji:    kanji,
		Keyword:  "meaning of " + kanji,
		Readings: "on reading, kun reading",
		Strokes:  10,
		Radicals: "radical1, radical2",
	}
}

// WordData fetches word data (dummy function replacing Jisho API)
func fetchWordData(word string) WordData {
	// Dummy implementation - would be replaced with actual implementation
	return WordData{
		Word:        word,
		Definitions: "(meaning1, meaning2)",
		Reading:     word,
	}
}

// SentenceData generates sentence data with translation
func fetchSentenceData(sentence string) SentenceData {
	// Dummy implementation - would be replaced with actual implementation
	return SentenceData{
		Sentence:    sentence,
		Translation: "Translation of: " + sentence,
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
func writeSentencesToContentMd(sentences []string) {
	f, err := os.OpenFile(filepath.Join(contentPath, currentName+".md"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", filepath.Join(contentPath, currentName+".md"), err)
		return
	}
	defer f.Close()

	f.WriteString("\n")
	for _, s := range sentences {
		f.WriteString(fmt.Sprintf("[[%s]]\n", s))
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
		var wordList []string

		// Process sentences
		for _, sentence := range sentences {
			fmt.Println(sentence)
			words := parser(sentence)

			// Extract kanji
			for _, word := range words {
				for _, c := range word {
					if containsRune(kanjiSet, string(c)) && !containsRune(kanjiList, string(c)) && !containsRune(oldKanji, string(c)) {
						kanjiList = append(kanjiList, string(c))
					}
				}
			}

			// Extract words
			for _, word := range words {
				hasKanji := false
				for _, c := range word {
					if containsRune(kanjiSet, string(c)) {
						hasKanji = true
						break
					}
				}

				if hasKanji && !containsRune(wordList, word) && !containsRune(oldWords, word) {
					wordList = append(wordList, word)
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
			go func(w string) {
				defer wg.Done()
				wData := fetchWordData(w)
				wordCard(wData)
				editWordsTags(w)
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
		oldWords = append(oldWords, wordList...)

		// Add new entries to index files
		var kanjiEntries []string
		var wordEntries []string
		var sentenceEntries []string

		for _, k := range kanjiList {
			kanjiEntries = append(kanjiEntries, fmt.Sprintf("[[%s]]\n", k))
		}

		for _, w := range wordList {
			wordEntries = append(wordEntries, fmt.Sprintf("[[%s]]\n", w))
		}

		for _, s := range sentences {
			sentenceEntries = append(sentenceEntries, fmt.Sprintf("[[%s]]\n", s))
		}

		addNewStuff(kanjiEntries, wordEntries, sentenceEntries)
		writeSentencesToContentMd(sentences)
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

// Main is the entry point of the application
func main() {
	askForCSVs()
}
