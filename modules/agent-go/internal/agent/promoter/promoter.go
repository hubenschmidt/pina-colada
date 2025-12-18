package promoter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

var ErrFirstTokenTimeout = errors.New("first token timeout exceeded")

// ModelTier represents a model with its first-token timeout
type ModelTier struct {
	Model             string
	FirstTokenTimeout time.Duration
}

// RunFunc starts a streaming operation with the given model
type RunFunc func(ctx context.Context, model string) (<-chan agents.StreamEvent, <-chan error, error)

// ModelPromoter handles automatic model escalation on slow responses
type ModelPromoter struct {
	tiers []ModelTier
}

// NewModelPromoter creates a promoter with the given model tiers
func NewModelPromoter(tiers []ModelTier) *ModelPromoter {
	return &ModelPromoter{tiers: tiers}
}

// RunStreamWithPromotion runs streaming with automatic model promotion.
// Returns the model that was used and any error.
func (p *ModelPromoter) RunStreamWithPromotion(ctx context.Context, runFn RunFunc, eventCh chan<- agents.StreamEvent) (string, error) {
	for i, tier := range p.tiers {
		usedModel, err := p.tryTier(ctx, tier, runFn, eventCh)

		// Success - return model used (guard clause)
		if err == nil {
			return usedModel, nil
		}

		// Non-timeout error - don't promote (guard clause)
		if !errors.Is(err, ErrFirstTokenTimeout) {
			return "", err
		}

		// Last tier exhausted - return error (guard clause)
		if i == len(p.tiers)-1 {
			return "", fmt.Errorf("all model tiers exhausted: %w", err)
		}

		log.Printf("⏱️ Model %s timed out after %v, promoting to %s", tier.Model, tier.FirstTokenTimeout, p.tiers[i+1].Model)
	}
	return "", fmt.Errorf("no model tiers configured")
}

func (p *ModelPromoter) tryTier(ctx context.Context, tier ModelTier, runFn RunFunc, eventCh chan<- agents.StreamEvent) (string, error) {
	timeoutCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	streamCh, errCh, err := runFn(timeoutCtx, tier.Model)
	if err != nil {
		return "", err
	}

	timer := time.NewTimer(tier.FirstTokenTimeout)
	defer timer.Stop()

	return p.processStream(ctx, cancel, streamCh, errCh, eventCh, timer, tier.Model)
}

func (p *ModelPromoter) processStream(
	ctx context.Context,
	cancel context.CancelFunc,
	streamCh <-chan agents.StreamEvent,
	errCh <-chan error,
	eventCh chan<- agents.StreamEvent,
	timer *time.Timer,
	model string,
) (string, error) {
	firstTokenReceived := false

	for {
		select {
		case evt, ok := <-streamCh:
			if !ok {
				return model, <-errCh // Stream completed
			}
			firstTokenReceived = p.handleEvent(evt, firstTokenReceived, timer, eventCh)

		case <-timer.C:
			if firstTokenReceived {
				continue // Already receiving tokens, ignore timer
			}
			cancel()
			return "", ErrFirstTokenTimeout

		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func (p *ModelPromoter) handleEvent(evt agents.StreamEvent, firstTokenReceived bool, timer *time.Timer, eventCh chan<- agents.StreamEvent) bool {
	eventCh <- evt

	if firstTokenReceived {
		return true
	}

	// Check if this is a text delta event (first token)
	rawEvt, ok := evt.(agents.RawResponsesStreamEvent)
	if !ok {
		return false
	}
	if rawEvt.Data.Type != "response.output_text.delta" {
		return false
	}
	if rawEvt.Data.Delta == "" {
		return false
	}

	// First text token received - stop timeout
	timer.Stop()
	log.Printf("✅ First token received, cancelling timeout")
	return true
}
