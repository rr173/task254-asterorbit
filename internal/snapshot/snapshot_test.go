package snapshot

import (
	"testing"

	"task254-asterorbit/internal/model"
)

func TestComputeHashIsDeterministicAndContentSensitive(t *testing.T) {
	a := ComputeHash("payload-a")
	if a != ComputeHash("payload-a") {
		t.Fatal("same payload produced different hashes")
	}
	if a == ComputeHash("payload-b") {
		t.Fatal("different payloads produced the same hash")
	}
}

func TestComputeHashIgnoresPublicationTimestamp(t *testing.T) {
	a, err := BuildPayload(model.Orbit{ID: "orb-1", A: 1.1}, "cat-1", nil, "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("build first payload: %v", err)
	}
	b, err := BuildPayload(model.Orbit{ID: "orb-1", A: 1.1}, "cat-1", nil, "2026-01-01T00:00:01Z")
	if err != nil {
		t.Fatalf("build second payload: %v", err)
	}
	if ComputeHash(a) != ComputeHash(b) {
		t.Fatal("publication timestamp changed the scientific content hash")
	}
}

func TestBuildPayloadIncludesCalibrationSnapshotInputs(t *testing.T) {
	content, err := BuildPayload(model.Orbit{ID: "orb-1", A: 1.1}, "cat-1", []model.Station{{ID: "sta-1", RAZeroArcsec: 12}}, "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}
	for _, want := range []string{"orb-1", "cat-1", "sta-1", "12", "2026-01-01T00:00:00Z"} {
		if !contains(content, want) {
			t.Fatalf("payload missing %q: %s", want, content)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
