package models

import (
	"github.com/google/uuid"
)

// Question represents a single quiz question with answer options.
type Question struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	QuizID     uuid.UUID  `json:"quiz_id" db:"quiz_id"`
	Text       string     `json:"text" db:"text"`
	OptionA    string     `json:"option_a" db:"option_a"`
	OptionB    string     `json:"option_b" db:"option_b"`
	OptionC    string     `json:"option_c" db:"option_c"`
	OptionD    string     `json:"option_d" db:"option_d"`
	Answer     string     `json:"answer,omitempty" db:"answer"` // omitted from client responses
	Difficulty Difficulty `json:"difficulty" db:"difficulty"`
	EloWeight  int        `json:"elo_weight" db:"elo_weight"`
}

// QuestionForClient strips the correct answer before sending to clients.
type QuestionForClient struct {
	ID         uuid.UUID  `json:"id"`
	QuizID     uuid.UUID  `json:"quiz_id"`
	Text       string     `json:"text"`
	OptionA    string     `json:"option_a"`
	OptionB    string     `json:"option_b"`
	OptionC    string     `json:"option_c"`
	OptionD    string     `json:"option_d"`
	Difficulty Difficulty `json:"difficulty"`
}

// ToClient converts a Question to a client-safe representation.
func (q *Question) ToClient() *QuestionForClient {
	return &QuestionForClient{
		ID:         q.ID,
		QuizID:     q.QuizID,
		Text:       q.Text,
		OptionA:    q.OptionA,
		OptionB:    q.OptionB,
		OptionC:    q.OptionC,
		OptionD:    q.OptionD,
		Difficulty: q.Difficulty,
	}
}
