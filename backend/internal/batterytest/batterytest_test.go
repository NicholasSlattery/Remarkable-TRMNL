package batterytest

import (
	"testing"
	"time"
)

func TestLifecycleAndEstimate(t *testing.T) {
	start := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	var state State
	if err := state.Start(start, 100); err != nil {
		t.Fatal(err)
	}
	state.Refreshes = 3
	state.Wakes = 2
	state.AddSample(start.Add(24*time.Hour), 80, false)
	s := state.Snapshot(start.Add(24*time.Hour), "Discharging")
	if !s.EstimateAvailable || s.PercentUsed != 20 || s.ProjectedTotalHours != 120 || s.ProjectedRemainHours != 96 {
		t.Fatalf("unexpected snapshot: %#v", s)
	}
	state.Stop(start.Add(30*time.Hour), 75)
	if state.Active || state.StoppedAt == nil || len(state.Samples) != 3 {
		t.Fatalf("unexpected stopped state: %#v", state)
	}
}

func TestEstimateWaitsForTenPercent(t *testing.T) {
	start := time.Now().Add(-time.Hour)
	var state State
	_ = state.Start(start, 100)
	state.AddSample(time.Now(), 91, true)
	if state.Snapshot(time.Now(), "Discharging").EstimateAvailable {
		t.Fatal("a nine percent drop is too noisy for an estimate")
	}
	state.AddSample(time.Now().Add(time.Minute), 90, true)
	if !state.Snapshot(time.Now().Add(time.Minute), "Discharging").EstimateAvailable {
		t.Fatal("a ten percent drop should produce an estimate")
	}
}

func TestChargingInvalidatesEstimate(t *testing.T) {
	start := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	var state State
	if err := state.Start(start, 100); err != nil {
		t.Fatal(err)
	}
	state.AddSampleWithStatus(start.Add(24*time.Hour), 85, false, "Discharging")
	if !state.Snapshot(start.Add(24*time.Hour), "Discharging").EstimateAvailable {
		t.Fatal("clean discharge should produce an estimate")
	}
	state.AddSampleWithStatus(start.Add(25*time.Hour), 90, false, "Charging")
	snapshot := state.Snapshot(start.Add(25*time.Hour), "Charging")
	if !snapshot.ChargingObserved || snapshot.EstimateAvailable {
		t.Fatalf("charging should invalidate estimate: %+v", snapshot)
	}
}
