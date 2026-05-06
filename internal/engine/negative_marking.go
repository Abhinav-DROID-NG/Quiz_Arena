package engine

// Mark computes the mark awarded for a single answer using GATE-style negative marking.
// correct=true  → +1
// correct=false, skipped=true  → 0
// correct=false, skipped=false → -1
func Mark(correct, skipped bool) int {
	if correct {
		return 1
	}
	if skipped {
		return 0
	}
	return -1
}
