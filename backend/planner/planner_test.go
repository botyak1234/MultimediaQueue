package planner

import (
	"testing"

	"media-planner/backend/models"
)

func TestCalculatePlan(t *testing.T) {
	items := []models.MediaItem{
		{
			ID:                 1,
			Type:               models.MediaTypeMovie,
			Title:              "Short Film",
			DurationMinutes:    30,
		},
		{
			ID:                 2,
			Type:               models.MediaTypeSeries,
			Title:              "Sitcom",
			TotalEpisodes:      10,
			WatchedEpisodes:    0,
			AvgEpisodeDuration: 20,
		},
		{
			ID:              3,
			Type:            models.MediaTypeBook,
			Title:           "Sci-Fi Novel",
			TotalPages:      100,
			ReadPages:       0,
			AvgPageReadTime: 2, // 2 mins per page
		},
	}

	// 1. Available 60 minutes
	// Movie takes 30 mins -> 30 left.
	// Sitcom takes 20 mins per episode -> 1 episode = 20 mins, 10 left.
	// Novel takes 2 mins per page -> 5 pages = 10 mins, 0 left.
	plan := CalculatePlan(items, 60)

	if plan.UsedMinutes != 60 {
		t.Fatalf("Expected used 60 minutes, got %d", plan.UsedMinutes)
	}
	if plan.RemainingMinutes != 0 {
		t.Fatalf("Expected remaining 0 minutes, got %d", plan.RemainingMinutes)
	}
	if len(plan.Actions) != 3 {
		t.Fatalf("Expected 3 actions, got %d", len(plan.Actions))
	}

	// Verify actions details
	if plan.Actions[0].UnitsToConsume != 1 || plan.Actions[0].DurationMinutes != 30 {
		t.Errorf("Movie action error: %+v", plan.Actions[0])
	}
	if plan.Actions[1].UnitsToConsume != 1 || plan.Actions[1].DurationMinutes != 20 {
		t.Errorf("Series action error: %+v", plan.Actions[1])
	}
	if plan.Actions[2].UnitsToConsume != 5 || plan.Actions[2].DurationMinutes != 10 {
		t.Errorf("Book action error: %+v", plan.Actions[2])
	}
}

func TestCalculatePlanEdgeCases(t *testing.T) {
	items := []models.MediaItem{
		{
			ID:              1,
			Type:            models.MediaTypeMovie,
			Title:           "Long Movie",
			DurationMinutes: 180,
		},
	}

	// 100 available minutes for 180 min movie -> no action fit
	plan := CalculatePlan(items, 100)
	if len(plan.Actions) != 0 {
		t.Fatalf("Expected 0 actions, got %d", len(plan.Actions))
	}
	if plan.RemainingMinutes != 100 {
		t.Fatalf("Expected 100 remaining minutes, got %d", plan.RemainingMinutes)
	}
}
