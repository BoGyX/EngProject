package services

import (
	"english-learning/internal/models"
	"fmt"
	"testing"
)

func TestSelectTrainingCardPlansForActiveBatchKeepsCurrentBatchOnly(t *testing.T) {
	plans := make([]trainingSessionCardPlan, 0, 20)
	for cardID := int64(1); cardID <= 20; cardID++ {
		modes := []string{"choice"}
		if cardID == 1 {
			modes = nil
		}

		plans = append(plans, trainingSessionCardPlan{
			Card:  trainingCardPlanStub{ID: cardID}.toModel(),
			Modes: modes,
		})
	}

	selected := selectTrainingCardPlansForActiveBatch(plans, 10)
	if len(selected) != 9 {
		t.Fatalf("expected 9 cards from the current batch, got %d", len(selected))
	}

	for index, plan := range selected {
		expectedID := int64(index + 2)
		if plan.Card == nil || plan.Card.ID != expectedID {
			t.Fatalf("expected card %d at index %d, got %+v", expectedID, index, plan.Card)
		}
	}
}

func TestSelectTrainingCardPlansForActiveBatchMovesToNextBatchAfterFirstIsCompleted(t *testing.T) {
	plans := make([]trainingSessionCardPlan, 0, 20)
	for cardID := int64(1); cardID <= 20; cardID++ {
		modes := []string{"choice"}
		if cardID <= 11 {
			modes = nil
		}

		plans = append(plans, trainingSessionCardPlan{
			Card:  trainingCardPlanStub{ID: cardID}.toModel(),
			Modes: modes,
		})
	}

	selected := selectTrainingCardPlansForActiveBatch(plans, 10)
	if len(selected) != 9 {
		t.Fatalf("expected 9 cards from the second batch, got %d", len(selected))
	}

	for index, plan := range selected {
		expectedID := int64(index + 12)
		if plan.Card == nil || plan.Card.ID != expectedID {
			t.Fatalf("expected card %d at index %d, got %+v", expectedID, index, plan.Card)
		}
	}
}

func TestBuildTrainingQueueForBatchKeepsChaosWithinSelectedWords(t *testing.T) {
	plans := []trainingSessionCardPlan{
		{
			Card:     trainingCardPlanStub{ID: 1}.toModel(),
			Modes:    []string{"view", "choice", "constructor"},
			Progress: 0,
		},
		{
			Card:     trainingCardPlanStub{ID: 2}.toModel(),
			Modes:    []string{"view", "choice", "constructor"},
			Progress: 0,
		},
	}

	queue := buildTrainingQueueForBatch(plans)
	if len(queue) != 6 {
		t.Fatalf("expected 6 queue entries, got %d", len(queue))
	}

	expectedPairs := map[string]int{
		"1:view":        1,
		"1:choice":      1,
		"1:constructor": 1,
		"2:view":        1,
		"2:choice":      1,
		"2:constructor": 1,
	}

	for _, entry := range queue {
		key := trainingQueuePairKey(entry.CardID, entry.Mode)
		if expectedPairs[key] == 0 {
			t.Fatalf("unexpected queue entry %s", key)
		}
		expectedPairs[key]--
	}

	for key, remaining := range expectedPairs {
		if remaining != 0 {
			t.Fatalf("expected queue entry %s to appear exactly once, remaining=%d", key, remaining)
		}
	}
}

type trainingCardPlanStub struct {
	ID int64
}

func (s trainingCardPlanStub) toModel() *models.Card {
	return &models.Card{ID: s.ID}
}

func trainingQueuePairKey(cardID int64, mode string) string {
	return fmt.Sprintf("%d:%s", cardID, mode)
}
