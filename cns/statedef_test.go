package cns

import "testing"

func TestStateDef_ZeroValue_HasPredictableDefaults(t *testing.T) {
	var s StateDef

	if s.Number != 0 {
		t.Errorf("expected zero-value StateDef to have Number 0, got %d", s.Number)
	}
	if s.Type != "" {
		t.Errorf("expected zero-value StateDef to have empty Type, got %q", s.Type)
	}
	if s.MoveType != "" {
		t.Errorf("expected zero-value StateDef to have empty MoveType, got %q", s.MoveType)
	}
	if s.Physics != "" {
		t.Errorf("expected zero-value StateDef to have empty Physics, got %q", s.Physics)
	}
	if s.Anim != 0 {
		t.Errorf("expected zero-value StateDef to have Anim 0, got %d", s.Anim)
	}
	if s.Ctrl {
		t.Errorf("expected zero-value StateDef to have Ctrl false, got %v", s.Ctrl)
	}
	if s.PowerAdd != 0 {
		t.Errorf("expected zero-value StateDef to have PowerAdd 0, got %d", s.PowerAdd)
	}
	if s.Juggle != 0 {
		t.Errorf("expected zero-value StateDef to have Juggle 0, got %d", s.Juggle)
	}
	if s.FaceP2 {
		t.Errorf("expected zero-value StateDef to have FaceP2 false, got %v", s.FaceP2)
	}
	if s.HitDefPersist {
		t.Errorf("expected zero-value StateDef to have HitDefPersist false, got %v", s.HitDefPersist)
	}
	if s.MoveHitPersist {
		t.Errorf("expected zero-value StateDef to have MoveHitPersist false, got %v", s.MoveHitPersist)
	}
	if s.HitCountPersist {
		t.Errorf("expected zero-value StateDef to have HitCountPersist false, got %v", s.HitCountPersist)
	}
	if s.SprPriority != 0 {
		t.Errorf("expected zero-value StateDef to have SprPriority 0, got %d", s.SprPriority)
	}
	if s.Controllers != nil {
		t.Errorf("expected zero-value StateDef to have nil Controllers, got %v", s.Controllers)
	}
	if len(s.Controllers) != 0 {
		t.Errorf("expected zero-value StateDef to have 0 controllers, got %d", len(s.Controllers))
	}

	// Ranging over a nil Controllers slice must not panic.
	for range s.Controllers {
		t.Fatal("expected no iterations over a nil Controllers slice")
	}
}

func TestController_ZeroValue_HasPredictableDefaults(t *testing.T) {
	var c Controller

	if c.Type != "" {
		t.Errorf("expected zero-value Controller to have empty Type, got %q", c.Type)
	}
	if c.Triggers != nil {
		t.Errorf("expected zero-value Controller to have nil Triggers, got %v", c.Triggers)
	}
	if c.Parameters != nil {
		t.Errorf("expected zero-value Controller to have nil Parameters, got %v", c.Parameters)
	}

	// Ranging over a nil Triggers slice and a nil Parameters map must not panic.
	for range c.Triggers {
		t.Fatal("expected no iterations over a nil Triggers slice")
	}
	for range c.Parameters {
		t.Fatal("expected no iterations over a nil Parameters map")
	}
	if _, ok := c.Parameters["value"]; ok {
		t.Fatal("expected a lookup on a nil Parameters map to report not found, not panic")
	}
}

func TestStateDef_WithControllers_PreservesAssignedFieldValues(t *testing.T) {
	s := StateDef{
		Number:          200,
		Type:            StateTypeStanding,
		MoveType:        MoveTypeAttack,
		Physics:         PhysicsStanding,
		Anim:            200,
		Ctrl:            true,
		PowerAdd:        30,
		Juggle:          5,
		FaceP2:          true,
		HitDefPersist:   true,
		MoveHitPersist:  true,
		HitCountPersist: true,
		SprPriority:     2,
		Controllers: []Controller{
			{
				Type:     "ChangeState",
				Triggers: []string{"Time = 0", "AnimTime = 0"},
				Parameters: map[string]string{
					"value": "0",
					"ctrl":  "1",
				},
			},
		},
	}

	if s.Number != 200 {
		t.Errorf("expected Number 200, got %d", s.Number)
	}
	if s.Type != StateTypeStanding {
		t.Errorf("expected Type StateTypeStanding, got %q", s.Type)
	}
	if s.MoveType != MoveTypeAttack {
		t.Errorf("expected MoveType MoveTypeAttack, got %q", s.MoveType)
	}
	if s.Physics != PhysicsStanding {
		t.Errorf("expected Physics PhysicsStanding, got %q", s.Physics)
	}
	if s.Anim != 200 {
		t.Errorf("expected Anim 200, got %d", s.Anim)
	}
	if !s.Ctrl {
		t.Errorf("expected Ctrl true, got %v", s.Ctrl)
	}
	if s.PowerAdd != 30 {
		t.Errorf("expected PowerAdd 30, got %d", s.PowerAdd)
	}
	if s.Juggle != 5 {
		t.Errorf("expected Juggle 5, got %d", s.Juggle)
	}
	if !s.FaceP2 || !s.HitDefPersist || !s.MoveHitPersist || !s.HitCountPersist {
		t.Errorf("expected all persist/facing flags true, got %+v", s)
	}
	if s.SprPriority != 2 {
		t.Errorf("expected SprPriority 2, got %d", s.SprPriority)
	}
	if len(s.Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(s.Controllers))
	}

	ctrl := s.Controllers[0]
	if ctrl.Type != "ChangeState" {
		t.Errorf("expected controller Type %q, got %q", "ChangeState", ctrl.Type)
	}
	if len(ctrl.Triggers) != 2 || ctrl.Triggers[0] != "Time = 0" || ctrl.Triggers[1] != "AnimTime = 0" {
		t.Errorf("expected Triggers [\"Time = 0\", \"AnimTime = 0\"], got %v", ctrl.Triggers)
	}
	if ctrl.Parameters["value"] != "0" {
		t.Errorf("expected Parameters[\"value\"] = %q, got %q", "0", ctrl.Parameters["value"])
	}
	if ctrl.Parameters["ctrl"] != "1" {
		t.Errorf("expected Parameters[\"ctrl\"] = %q, got %q", "1", ctrl.Parameters["ctrl"])
	}
}

func TestController_WithoutTriggers_RepresentsAnUnconditionalController(t *testing.T) {
	// A controller with no Triggers models an .cns [State N] block with no
	// "trigger1 = ..." lines at all — it always fires when its state is
	// active, not "never fires" (nil is not treated as "always false").
	c := Controller{
		Type:       "VelSet",
		Parameters: map[string]string{"y": "0"},
	}

	if c.Triggers != nil {
		t.Errorf("expected Triggers to stay nil when not assigned, got %v", c.Triggers)
	}
	if c.Type != "VelSet" {
		t.Errorf("expected Type %q, got %q", "VelSet", c.Type)
	}
	if c.Parameters["y"] != "0" {
		t.Errorf("expected Parameters[\"y\"] = %q, got %q", "0", c.Parameters["y"])
	}
}
