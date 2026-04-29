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

func TestBuildTrainingQueueForBatchPrioritizesStudyBeforePractice(t *testing.T) {
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
	if len(queue) != 2 {
		t.Fatalf("expected 2 queue entries, got %d", len(queue))
	}

	seenCards := map[int64]bool{}
	for _, entry := range queue {
		if seenCards[entry.CardID] {
			t.Fatalf("card %d should appear only once per session", entry.CardID)
		}
		seenCards[entry.CardID] = true

		switch entry.CardID {
		case 1, 2:
			if entry.Mode != "view" {
				t.Fatalf("expected study mode view for card %d, got %q", entry.CardID, entry.Mode)
			}
		default:
			t.Fatalf("unexpected card %d in queue", entry.CardID)
		}
	}

	if !seenCards[1] || !seenCards[2] {
		t.Fatalf("expected both selected cards to appear exactly once, got %+v", seenCards)
	}
}

func TestBuildTrainingQueueForBatchKeepsStudyingUntilAllSelectedWordsAreViewed(t *testing.T) {
	plans := []trainingSessionCardPlan{
		{
			Card:     trainingCardPlanStub{ID: 1}.toModel(),
			Modes:    []string{"view", "choice"},
			Progress: 0,
		},
		{
			Card:     trainingCardPlanStub{ID: 2}.toModel(),
			Modes:    []string{"choice", "constructor"},
			Progress: 50,
		},
	}

	queue := buildTrainingQueueForBatch(plans)
	if len(queue) != 1 {
		t.Fatalf("expected only 1 study entry while some words are still being introduced, got %d", len(queue))
	}
	if queue[0].CardID != 1 || queue[0].Mode != "view" {
		t.Fatalf("expected only card 1 in view mode, got %+v", queue[0])
	}
}

func TestBuildTrainingQueueForBatchUsesSingleRandomPracticeModePerWordAfterStudy(t *testing.T) {
	plans := []trainingSessionCardPlan{
		{
			Card:     trainingCardPlanStub{ID: 1}.toModel(),
			Modes:    []string{"choice", "constructor"},
			Progress: 25,
		},
		{
			Card:     trainingCardPlanStub{ID: 2}.toModel(),
			Modes:    []string{"choice", "russian", "constructor"},
			Progress: 50,
		},
	}

	queue := buildTrainingQueueForBatch(plans)
	if len(queue) != 2 {
		t.Fatalf("expected 2 practice entries, got %d", len(queue))
	}

	seenCards := map[int64]bool{}
	for _, entry := range queue {
		if seenCards[entry.CardID] {
			t.Fatalf("card %d should appear only once per practice session", entry.CardID)
		}
		seenCards[entry.CardID] = true

		switch entry.CardID {
		case 1:
			if entry.Mode != "choice" && entry.Mode != "constructor" {
				t.Fatalf("unexpected practice mode %q for card 1", entry.Mode)
			}
		case 2:
			if entry.Mode != "choice" && entry.Mode != "russian" && entry.Mode != "constructor" {
				t.Fatalf("unexpected practice mode %q for card 2", entry.Mode)
			}
		default:
			t.Fatalf("unexpected card %d in practice queue", entry.CardID)
		}
	}
}

func TestBuildTrainingQueueForBatchUsesOnlyRemainingModes(t *testing.T) {
	plans := []trainingSessionCardPlan{
		{
			Card:     trainingCardPlanStub{ID: 7}.toModel(),
			Modes:    []string{"russian"},
			Progress: 75,
		},
	}

	queue := buildTrainingQueueForBatch(plans)
	if len(queue) != 1 {
		t.Fatalf("expected 1 queue entry, got %d", len(queue))
	}
	if queue[0].CardID != 7 || queue[0].Mode != "russian" {
		t.Fatalf("expected only remaining mode russian for card 7, got %+v", queue[0])
	}
}

type trainingCardPlanStub struct {
	ID int64
}

func (s trainingCardPlanStub) toModel() *models.Card {
	return &models.Card{ID: s.ID}
}
