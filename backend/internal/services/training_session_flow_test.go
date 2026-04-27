package services

import (
	"english-learning/internal/models"
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

type trainingCardPlanStub struct {
	ID int64
}

func (s trainingCardPlanStub) toModel() *models.Card {
	return &models.Card{ID: s.ID}
}
