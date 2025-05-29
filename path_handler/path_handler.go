package path_handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Pathing struct {
	NotesDir      string `json:"notesDir"`
	ContentMd     string `json:"contentMd"`
	KanjiMd       string `json:"kanjiMd"`
	SentencesMd   string `json:"sentencesMd"`
	WordsMd       string `json:"wordsMd"`
	ContentPath   string `json:"contentPath"`
	KanjiPath     string `json:"kanjiPath"`
	SentencesPath string `json:"sentencesPath"`
	WordsPath     string `json:"wordsPath"`
	CsvPath       string `json:"csvPath"`
	NewContent    string `json:"newContent"`
}

func DefaultPathing() Pathing {
	notesDir := filepath.Join("Notes", "Japanese Notes")
	return Pathing{
		NotesDir:      notesDir,
		ContentMd:     filepath.Join(notesDir, "Content.md"),
		KanjiMd:       filepath.Join(notesDir, "Kanji.md"),
		SentencesMd:   filepath.Join(notesDir, "Sentences.md"),
		WordsMd:       filepath.Join(notesDir, "Words.md"),
		ContentPath:   filepath.Join(notesDir, "Content"),
		KanjiPath:     filepath.Join(notesDir, "Kanji"),
		SentencesPath: filepath.Join(notesDir, "Sentences"),
		WordsPath:     filepath.Join(notesDir, "Words"),
		CsvPath:       filepath.Join(notesDir, "CSV"),
		NewContent:    filepath.Join(notesDir, "New"),
	}
}

var TestPathing = Pathing{
	NotesDir:      filepath.Join("Test", "Japanese Notes"),
	ContentMd:     filepath.Join(filepath.Join("Test", "Japanese Notes"), "Content.md"),
	KanjiMd:       filepath.Join(filepath.Join("Test", "Japanese Notes"), "Kanji.md"),
	SentencesMd:   filepath.Join(filepath.Join("Test", "Japanese Notes"), "Sentences.md"),
	WordsMd:       filepath.Join(filepath.Join("Test", "Japanese Notes"), "Words.md"),
	ContentPath:   filepath.Join(filepath.Join("Test", "Japanese Notes"), "Content"),
	KanjiPath:     filepath.Join(filepath.Join("Test", "Japanese Notes"), "Kanji"),
	SentencesPath: filepath.Join(filepath.Join("Test", "Japanese Notes"), "Sentences"),
	WordsPath:     filepath.Join(filepath.Join("Test", "Japanese Notes"), "Words"),
	CsvPath:       filepath.Join(filepath.Join("Test", "Japanese Notes"), "CSV"),
	NewContent:    "./New Content",
}

func LoadPathing(filePath string) Pathing {
	// Load default values
	defaults := DefaultPathing()
	pathing := defaults

	// Try opening the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Could not open pathing.json, using defaults and creating pathing.json:", err)
		data := map[string]string{
			"notesDir":      "Notes/Japanese Notes",
			"contentMd":     "Notes/Japanese Notes/Content.md",
			"kanjiMd":       "Notes/Japanese Notes/Kanji.md",
			"sentencesMd":   "Notes/Japanese Notes/Sentences.md",
			"wordsMd":       "Notes/Japanese Notes/Words.md",
			"contentPath":   "Notes/Japanese Notes/Content",
			"kanjiPath":     "Notes/Japanese Notes/Kanji",
			"sentencesPath": "Notes/Japanese Notes/Sentences",
			"wordsPath":     "Notes/Japanese Notes/Words",
			"csvPath":       "Notes/Japanese Notes/CSV",
			"newContent":    "Notes/Japanese Notes/New",
		}

		file, err := os.Create("pathing.json")
		if err != nil {
			fmt.Println("Could not create pathing.json", err)
			return pathing
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ") // Pretty-print the JSON
		if err := encoder.Encode(data); err != nil {
			fmt.Println("Could not create pathing.json", err)
		}
		return pathing
	}
	defer file.Close()

	// Decode the JSON file
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&pathing); err != nil {
		fmt.Println("Error decoding JSON, using defaults:", err)
		return pathing
	}

	// Validate fields and fallback to defaults for empty strings
	if pathing.NotesDir == "" {
		pathing.NotesDir = defaults.NotesDir
	}
	if pathing.ContentMd == "" {
		pathing.ContentMd = defaults.ContentMd
	}
	if pathing.KanjiMd == "" {
		pathing.KanjiMd = defaults.KanjiMd
	}
	if pathing.SentencesMd == "" {
		pathing.SentencesMd = defaults.SentencesMd
	}
	if pathing.WordsMd == "" {
		pathing.WordsMd = defaults.WordsMd
	}
	if pathing.ContentPath == "" {
		pathing.ContentPath = defaults.ContentPath
	}
	if pathing.KanjiPath == "" {
		pathing.KanjiPath = defaults.KanjiPath
	}
	if pathing.SentencesPath == "" {
		pathing.SentencesPath = defaults.SentencesPath
	}
	if pathing.WordsPath == "" {
		pathing.WordsPath = defaults.WordsPath
	}
	if pathing.CsvPath == "" {
		pathing.CsvPath = defaults.CsvPath
	}
	if pathing.NewContent == "" {
		pathing.NewContent = defaults.NewContent
	}
	return pathing
}
