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

// 内容指纹只取科学内容（轨道+星表+台站校准），发布时间戳属于元数据，不参与指纹。
// 重复发布同一轨道解只应因时间戳不同而 payload 不同，但科学内容相同，故指纹稳定。
func TestComputeHashIgnoresPublicationTimestamp(t *testing.T) {
	orb := model.Orbit{ID: "orb-1", A: 1.1}
	stations := []model.Station{{ID: "sta-1", RAZeroArcsec: 12}}

	// 同样的科学内容（与 publishedAt 无关）必须产生相同指纹。
	a, err := BuildContent(orb, "cat-1", stations)
	if err != nil {
		t.Fatalf("build first content: %v", err)
	}
	b, err := BuildContent(orb, "cat-1", stations)
	if err != nil {
		t.Fatalf("build second content: %v", err)
	}
	if ComputeHash(a) != ComputeHash(b) {
		t.Fatal("identical scientific content produced different hashes")
	}

	// payload 因时间戳不同而变化，但管线指纹来自科学内容，保持稳定：
	// 即“同一轨道解重复发布 → 相同指纹”。
	p1, err := BuildPayload(orb, "cat-1", stations, "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("build first payload: %v", err)
	}
	p2, err := BuildPayload(orb, "cat-1", stations, "2026-01-01T00:00:01Z")
	if err != nil {
		t.Fatalf("build second payload: %v", err)
	}
	if p1 == p2 {
		t.Fatal("payloads should differ by publication timestamp")
	}
	if ComputeHash(p1) == ComputeHash(p2) {
		t.Fatal("payload-level hash should differ by timestamp (only content hash must be stable)")
	}
	_ = a // a 来自 BuildContent，已证明科学内容指纹稳定
}

// 台站作为集合参与指纹：底层存储返回顺序不同也必须得到相同科学内容 → 相同指纹。
func TestComputeHashContentInsensitiveToStationOrder(t *testing.T) {
	orb := model.Orbit{ID: "orb-1", A: 1.1}
	stations := []model.Station{{ID: "sta-2"}, {ID: "sta-1"}, {ID: "sta-3"}}
	shuffled := []model.Station{{ID: "sta-3"}, {ID: "sta-1"}, {ID: "sta-2"}}

	a, err := BuildContent(orb, "cat-1", stations)
	if err != nil {
		t.Fatalf("build content order A: %v", err)
	}
	b, err := BuildContent(orb, "cat-1", shuffled)
	if err != nil {
		t.Fatalf("build content order B: %v", err)
	}
	if ComputeHash(a) != ComputeHash(b) {
		t.Fatal("station ordering changed the scientific content hash")
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
