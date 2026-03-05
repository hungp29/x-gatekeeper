package word

import (
	"net/http"

	wordv1 "github.com/hungp29/x-proto/gen/go/word/v1"
	"github.com/gin-gonic/gin"
)

// Handler exposes x-word gRPC methods as HTTP endpoints.
type Handler struct {
	client wordv1.WordServiceClient
}

// NewHandler constructs a Handler backed by the given WordServiceClient.
func NewHandler(client wordv1.WordServiceClient) *Handler {
	return &Handler{client: client}
}

// GetWord handles GET /word/:word?dict=english|english-vietnamese
func (h *Handler) GetWord(c *gin.Context) {
	word := c.Param("word")
	if word == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "word is required"})
		return
	}

	dict, ok := parseDictQuery(c)
	if !ok {
		return
	}

	resp, err := h.client.GetWord(c.Request.Context(), &wordv1.GetWordRequest{
		Word: word,
		Dict: dict,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, protoWordToJSON(resp.Word))
}

// GetWords handles POST /words
// Request body: {"words": ["hello", "world"], "dict": "english"}
func (h *Handler) GetWords(c *gin.Context) {
	var req struct {
		Words []string `json:"words" binding:"required,min=1"`
		Dict  string   `json:"dict"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dict := dictFromString(req.Dict)

	resp, err := h.client.GetWords(c.Request.Context(), &wordv1.GetWordsRequest{
		Words: req.Words,
		Dict:  dict,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	out := make([]any, len(resp.Words))
	for i, w := range resp.Words {
		out[i] = protoWordToJSON(w)
	}
	c.JSON(http.StatusOK, out)
}

// parseDictQuery reads ?dict= and returns the corresponding proto enum.
// Returns false and writes a 400 if the value is unrecognised.
func parseDictQuery(c *gin.Context) (wordv1.Dictionary, bool) {
	raw := c.DefaultQuery("dict", "english")
	d := dictFromString(raw)
	if d == wordv1.Dictionary_DICTIONARY_UNSPECIFIED {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown dict: " + raw + "; valid values: english, english-vietnamese"})
		return d, false
	}
	return d, true
}

// dictFromString converts the user-supplied string to a proto Dictionary enum.
func dictFromString(s string) wordv1.Dictionary {
	switch s {
	case "english", "":
		return wordv1.Dictionary_DICTIONARY_ENGLISH
	case "english-vietnamese":
		return wordv1.Dictionary_DICTIONARY_ENGLISH_VIETNAMESE
	default:
		return wordv1.Dictionary_DICTIONARY_UNSPECIFIED
	}
}

// wordJSON is the JSON shape returned for a single word entry.
type wordJSON struct {
	Text         string      `json:"text"`
	Phonetic     string      `json:"phonetic,omitempty"`
	PhoneticUK   string      `json:"phonetic_uk,omitempty"`
	PhoneticUS   string      `json:"phonetic_us,omitempty"`
	AudioUK      string      `json:"audio_uk,omitempty"`
	AudioUS      string      `json:"audio_us,omitempty"`
	PartOfSpeech []string    `json:"part_of_speech,omitempty"`
	Meanings     []meaningJSON `json:"meanings,omitempty"`
}

type meaningJSON struct {
	Definition string   `json:"definition"`
	Examples   []string `json:"examples,omitempty"`
}

func protoWordToJSON(w *wordv1.Word) wordJSON {
	if w == nil {
		return wordJSON{}
	}
	meanings := make([]meaningJSON, len(w.Meanings))
	for i, m := range w.Meanings {
		meanings[i] = meaningJSON{Definition: m.Definition, Examples: m.Examples}
	}
	return wordJSON{
		Text:         w.Text,
		Phonetic:     w.Phonetic,
		PhoneticUK:   w.PhoneticUk,
		PhoneticUS:   w.PhoneticUs,
		AudioUK:      w.AudioUk,
		AudioUS:      w.AudioUs,
		PartOfSpeech: w.PartOfSpeech,
		Meanings:     meanings,
	}
}
