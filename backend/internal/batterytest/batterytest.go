package batterytest

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

const MaxSamples = 4032 // 42 days at one sample every 15 minutes.

type Sample struct {
	At      time.Time `json:"at"`
	Percent int       `json:"percent"`
}

type State struct {
	Active       bool       `json:"active"`
	StartedAt    time.Time  `json:"started_at,omitempty"`
	StoppedAt    *time.Time `json:"stopped_at,omitempty"`
	StartPercent int        `json:"start_percent"`
	Current      int        `json:"current_percent"`
	Refreshes    int        `json:"refresh_count"`
	Wakes        int        `json:"wake_count"`
	Samples      []Sample   `json:"samples"`
}

type Snapshot struct {
	State
	ElapsedHours          float64 `json:"elapsed_hours"`
	PercentUsed           int     `json:"percent_used"`
	EstimateAvailable     bool    `json:"estimate_available"`
	ProjectedTotalHours   float64 `json:"projected_total_hours"`
	ProjectedRemainHours  float64 `json:"projected_remaining_hours"`
	MinimumEstimateNeeded int     `json:"minimum_percent_for_estimate"`
	BatteryStatus         string  `json:"battery_status,omitempty"`
}

func Load(path string) (State, error) {
	var s State
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	if len(s.Samples) > MaxSamples {
		s.Samples = append([]Sample(nil), s.Samples[len(s.Samples)-MaxSamples:]...)
	}
	return s, nil
}

func (s *State) Start(at time.Time, percent int) error {
	if percent < 0 || percent > 100 {
		return errors.New("battery percentage is unavailable")
	}
	*s = State{
		Active: true, StartedAt: at, StartPercent: percent, Current: percent,
		Samples: []Sample{{At: at, Percent: percent}},
	}
	return nil
}

func (s *State) Stop(at time.Time, percent int) {
	if !s.Active {
		return
	}
	s.AddSample(at, percent, true)
	s.Active = false
	s.StoppedAt = &at
}

func (s *State) Reset() { *s = State{} }

func (s *State) AddSample(at time.Time, percent int, force bool) {
	if !s.Active || percent < 0 || percent > 100 {
		return
	}
	s.Current = percent
	if len(s.Samples) > 0 && !force {
		last := s.Samples[len(s.Samples)-1]
		if at.Sub(last.At) < 5*time.Minute {
			return
		}
	}
	s.Samples = append(s.Samples, Sample{At: at, Percent: percent})
	if len(s.Samples) > MaxSamples {
		s.Samples = append([]Sample(nil), s.Samples[len(s.Samples)-MaxSamples:]...)
	}
}

func (s State) Clone() State {
	s.Samples = append([]Sample(nil), s.Samples...)
	if s.StoppedAt != nil {
		stopped := *s.StoppedAt
		s.StoppedAt = &stopped
	}
	return s
}

func (s State) Snapshot(now time.Time, status string) Snapshot {
	result := Snapshot{State: s.Clone(), MinimumEstimateNeeded: 2, BatteryStatus: status}
	if s.StartedAt.IsZero() {
		return result
	}
	end := now
	if !s.Active && s.StoppedAt != nil {
		end = *s.StoppedAt
	}
	result.ElapsedHours = end.Sub(s.StartedAt).Hours()
	if result.ElapsedHours < 0 {
		result.ElapsedHours = 0
	}
	result.PercentUsed = s.StartPercent - s.Current
	if result.PercentUsed >= result.MinimumEstimateNeeded && result.ElapsedHours > 0 {
		result.EstimateAvailable = true
		result.ProjectedTotalHours = result.ElapsedHours * 100 / float64(result.PercentUsed)
		result.ProjectedRemainHours = result.ElapsedHours * float64(s.Current) / float64(result.PercentUsed)
	}
	return result
}
