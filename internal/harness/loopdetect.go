package harness

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

type LoopAction int

const (
	LoopOK        LoopAction = iota
	LoopWarn                         // getting suspicious
	LoopForceText                    // tell model to stop using tools
	LoopAbort                        // kill the loop
)

type LoopDetector struct {
	mu         sync.Mutex
	history    []callSignature
	windowSize int
	warnAt     int
	forceAt    int
	abortAt    int
}

type callSignature struct {
	toolName string
	argsHash string
}

func NewLoopDetector() *LoopDetector {
	return &LoopDetector{
		windowSize: 20,
		warnAt:     3,
		forceAt:    5,
		abortAt:    8,
	}
}

func (ld *LoopDetector) Record(toolName string, argsJSON string) LoopAction {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	h := sha256.Sum256([]byte(argsJSON))
	sig := callSignature{
		toolName: toolName,
		argsHash: fmt.Sprintf("%x", h[:8]),
	}

	// immutable append: new slice, don't mutate the old one
	updated := make([]callSignature, len(ld.history), len(ld.history)+1)
	copy(updated, ld.history)
	updated = append(updated, sig)

	if len(updated) > ld.windowSize {
		updated = updated[len(updated)-ld.windowSize:]
	}
	ld.history = updated

	count := 0
	for _, s := range ld.history {
		if s.toolName == sig.toolName && s.argsHash == sig.argsHash {
			count++
		}
	}

	if count >= ld.abortAt {
		return LoopAbort
	}
	if count >= ld.forceAt {
		return LoopForceText
	}
	if count >= ld.warnAt {
		return LoopWarn
	}
	return LoopOK
}

func (ld *LoopDetector) Reset() {
	ld.mu.Lock()
	defer ld.mu.Unlock()
	ld.history = nil
}
