package harness

import "sync"

// TokenTracker approximates token usage per session.
// Uses a simple chars/4 heuristic (no real tokenizer needed).
type TokenTracker struct {
	inputTokens  int
	outputTokens int
	mu           sync.Mutex
}

func NewTokenTracker() *TokenTracker {
	return &TokenTracker{}
}

// AddInput records approximate input tokens from text.
func (t *TokenTracker) AddInput(text string) {
	t.mu.Lock()
	t.inputTokens += len(text) / 4
	t.mu.Unlock()
}

// AddOutput records approximate output tokens from text.
func (t *TokenTracker) AddOutput(text string) {
	t.mu.Lock()
	t.outputTokens += len(text) / 4
	t.mu.Unlock()
}

// Stats returns the current approximate token counts.
func (t *TokenTracker) Stats() (input, output int) {
	t.mu.Lock()
	input = t.inputTokens
	output = t.outputTokens
	t.mu.Unlock()
	return input, output
}

// Reset clears all token counts.
func (t *TokenTracker) Reset() {
	t.mu.Lock()
	t.inputTokens = 0
	t.outputTokens = 0
	t.mu.Unlock()
}
