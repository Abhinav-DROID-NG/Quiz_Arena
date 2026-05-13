package models

import (
	"github.com/google/uuid"
)

// Question represents a single quiz question with answer options.
type Question struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	QuizID       uuid.UUID  `json:"quiz_id" db:"quiz_id"`
	SubjectID    *uuid.UUID `json:"subject_id,omitempty" db:"subject_id"`
	Text         string     `json:"text" db:"text"`
	OptionA      string     `json:"option_a" db:"option_a"`
	OptionB      string     `json:"option_b" db:"option_b"`
	OptionC      string     `json:"option_c" db:"option_c"`
	OptionD      string     `json:"option_d" db:"option_d"`
	Answer       string     `json:"answer,omitempty" db:"answer"` // omitted from client responses
	Difficulty   Difficulty `json:"difficulty" db:"difficulty"`
	EloWeight    int        `json:"elo_weight" db:"elo_weight"`
	LatexEnabled bool       `json:"latex_enabled" db:"latex_enabled"`
	DiagramURL   string     `json:"diagram_url,omitempty" db:"diagram_url"`
	Year         int        `json:"year,omitempty" db:"year"` // exam year (e.g., GATE year)
	Source       string     `json:"source,omitempty" db:"source"` // e.g., "GATE 2022"
}

// QuestionForClient strips the correct answer and sensitive fields before sending to clients.
type QuestionForClient struct {
	ID           uuid.UUID  `json:"id"`
	QuizID       uuid.UUID  `json:"quiz_id"`
	SubjectID    *uuid.UUID `json:"subject_id,omitempty"`
	Text         string     `json:"text"`
	Options      []string   `json:"options"` // shuffled options without labels
	Difficulty   Difficulty `json:"difficulty"`
	LatexEnabled bool       `json:"latex_enabled"`
	DiagramURL   string     `json:"diagram_url,omitempty"`
}

// OptionMap stores shuffled option → original label mapping (server-side only).
type OptionMap struct {
	QuestionID uuid.UUID         `json:"question_id"`
	Mapping    map[string]string `json:"mapping"` // shuffled index → "A"/"B"/"C"/"D"
}

// ToClient converts a Question to a client-safe representation with shuffled options.
// shuffle indicates whether options should be randomized.
func (q *Question) ToClient() *QuestionForClient {
	return &QuestionForClient{
		ID:           q.ID,
		QuizID:       q.QuizID,
		SubjectID:    q.SubjectID,
		Text:         q.Text,
		Options:      []string{q.OptionA, q.OptionB, q.OptionC, q.OptionD},
		Difficulty:   q.Difficulty,
		LatexEnabled: q.LatexEnabled,
		DiagramURL:   q.DiagramURL,
	}
}
