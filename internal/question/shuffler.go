package question

import (
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/utils"
)

// Shuffler shuffles question answer options and returns a client-safe view
// along with a mapping from shuffled index (0-3) → original option label (A-D).
type Shuffler struct{}

// NewShuffler creates a Shuffler.
func NewShuffler() *Shuffler { return &Shuffler{} }

// Shuffle randomises the options of q and returns:
//   - a QuestionForClient with shuffled options
//   - a mapping[shuffledIndex] = originalLabel (e.g., mapping[2] = "A")
func (s *Shuffler) Shuffle(q *models.Question) (*models.QuestionForClient, map[int]string) {
	origOptions := []string{q.OptionA, q.OptionB, q.OptionC, q.OptionD}
	origLabels := []string{"A", "B", "C", "D"}

	shuffled, perm := utils.Shuffle(origOptions)

	// Build inverse mapping: shuffledIndex → original label
	mapping := make(map[int]string, 4)
	for newIdx, origIdx := range perm {
		mapping[newIdx] = origLabels[origIdx]
	}

	client := &models.QuestionForClient{
		ID:           q.ID,
		QuizID:       q.QuizID,
		SubjectID:    q.SubjectID,
		Text:         q.Text,
		Options:      shuffled,
		Difficulty:   q.Difficulty,
		LatexEnabled: q.LatexEnabled,
		DiagramURL:   q.DiagramURL,
	}

	return client, mapping
}
