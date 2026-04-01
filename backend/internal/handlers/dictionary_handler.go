package handlers

import (
	"english-learning/internal/services"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type DictionaryHandler struct {
	dictionaryService *services.DictionaryService
}

func NewDictionaryHandler(dictionaryService *services.DictionaryService) *DictionaryHandler {
	return &DictionaryHandler{
		dictionaryService: dictionaryService,
	}
}

// GetWordInfoResponse contains the data the frontend uses for card creation and reader hints.
type GetWordInfoResponse struct {
	Word        string   `json:"word"`
	Phonetic    string   `json:"phonetic"`
	AudioURL    string   `json:"audio_url,omitempty"`
	ImageURL    string   `json:"image_url,omitempty"`
	Translation string   `json:"translation"`
	Definition  string   `json:"definition"`
	Example     string   `json:"example,omitempty"`
	Meanings    []string `json:"meanings,omitempty"`
}

func buildSuggestedImageURL(word string) string {
	cleanedWord := strings.TrimSpace(strings.ToLower(word))
	if cleanedWord == "" {
		return ""
	}

	return "https://source.unsplash.com/featured/640x480/?" + url.QueryEscape(cleanedWord)
}

// GetWordInfo returns dictionary data plus a best-effort image suggestion for the admin card form.
func (h *DictionaryHandler) GetWordInfo(c *gin.Context) {
	word := c.Param("word")
	if word == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "word is required"})
		return
	}

	response := GetWordInfoResponse{
		Word:     word,
		ImageURL: buildSuggestedImageURL(word),
	}

	translation, err := h.dictionaryService.TranslateToRussian(word)
	if err == nil {
		response.Translation = translation
	}

	entry, err := h.dictionaryService.GetWordInfo(word)
	if err == nil {
		response.Phonetic = entry.Phonetic
		if response.Phonetic == "" && len(entry.Phonetics) > 0 {
			response.Phonetic = entry.Phonetics[0].Text
		}

		for _, phonetic := range entry.Phonetics {
			if phonetic.Audio != "" {
				response.AudioURL = phonetic.Audio
				break
			}
		}

		if len(entry.Meanings) > 0 && len(entry.Meanings[0].Definitions) > 0 {
			response.Definition = entry.Meanings[0].Definitions[0].Definition
			if entry.Meanings[0].Definitions[0].Example != "" {
				response.Example = entry.Meanings[0].Definitions[0].Example
			}
		}

		for _, meaning := range entry.Meanings {
			for _, definition := range meaning.Definitions {
				if definition.Definition != "" {
					response.Meanings = append(response.Meanings, definition.Definition)
				}
			}
		}

		if response.Translation == "" && response.Definition != "" {
			response.Translation = "📖 " + response.Definition
		}
	} else if response.Translation == "" {
		response.Translation = "Перевод не найден"
	}

	c.JSON(http.StatusOK, response)
}
