package main

import (
	"os"
	"testing"
)

func TestClampFloat(t *testing.T) {
	cases := []struct {
		v, lo, hi, want float64
	}{
		{0.5, 0, 1, 0.5},
		{-1, 0, 1, 0},
		{2, 0, 1, 1},
		{0, 0, 0, 0},
	}
	for _, c := range cases {
		got := clampFloat(c.v, c.lo, c.hi)
		if got != c.want {
			t.Errorf("clampFloat(%v, %v, %v) = %v, want %v", c.v, c.lo, c.hi, got, c.want)
		}
	}
}

func TestClampInt(t *testing.T) {
	if clampInt(150, 0, 100) != 100 {
		t.Fail()
	}
	if clampInt(-5, 0, 100) != 0 {
		t.Fail()
	}
	if clampInt(50, 0, 100) != 50 {
		t.Fail()
	}
}

func TestMinMaxInt(t *testing.T) {
	if minInt(3, 5) != 3 {
		t.Fail()
	}
	if maxInt(3, 5) != 5 {
		t.Fail()
	}
}

func TestGetSeason(t *testing.T) {
	// 0~29 spring, 30~59 summer, 60~89 autumn, 90~119 winter
	if getSeason(1) != "spring" {
		t.Errorf("day 1 should be spring, got %s", getSeason(1))
	}
	if getSeason(35) != "summer" {
		t.Errorf("day 35 should be summer, got %s", getSeason(35))
	}
	if getSeason(65) != "autumn" {
		t.Errorf("day 65 should be autumn, got %s", getSeason(65))
	}
	if getSeason(95) != "winter" {
		t.Errorf("day 95 should be winter, got %s", getSeason(95))
	}
	// 循环回 spring
	if getSeason(120) != "spring" {
		t.Errorf("day 120 should be spring, got %s", getSeason(120))
	}
}

func TestSpeciesCatalog(t *testing.T) {
	cat := speciesCatalog()
	if len(cat) < 8 {
		t.Fatalf("expected >=8 species, got %d", len(cat))
	}
	// 初始两只龟种应 cost=0
	freeCount := 0
	for _, s := range cat {
		if s.UnlockCost == 0 {
			freeCount++
		}
		if s.ID == "" || s.Name == "" {
			t.Errorf("species missing id/name: %+v", s)
		}
	}
	if freeCount < 1 {
		t.Errorf("expected at least 1 free starter species, got %d", freeCount)
	}
	// 关键龟种存在性
	for _, want := range []string{"muskTurtle", "chinesePondTurtle", "yellowMarginTurtle"} {
		if _, ok := findSpecies(want); !ok {
			t.Errorf("missing species %s", want)
		}
	}
}

func TestWaterLevelHelpers(t *testing.T) {
	if !isValidWaterLevel("deep") || !isValidWaterLevel("middle") || !isValidWaterLevel("land") {
		t.Fail()
	}
	if isValidWaterLevel("ocean") {
		t.Fail()
	}
	if waterLevelName("deep") == "" || waterLevelName("middle") == "" || waterLevelName("land") == "" {
		t.Fail()
	}
}

func TestBuildTurtleSuggestions(t *testing.T) {
	turtle := Turtle{Name: "小测试", Hunger: 20, Cleanliness: 30, Mood: 30, Health: HealthStat{Vitality: 40, Appetite: 60}}
	sp := SpeciesInfo{ID: "muskTurtle", Name: "麝香龟", HabitatType: "deep"}
	tank := map[string]interface{}{
		"water_level": "middle",
		"ammonia":     1.2,
		"clarity":     40,
		"has_filter":  false,
	}
	out := buildTurtleSuggestions(turtle, sp, tank)
	if len(out) == 0 {
		t.Fatal("expected suggestions")
	}
	// 确认包含饥饿警告
	found := false
	for _, s := range out {
		if txt, ok := s["text"].(string); ok && contains(txt, "饥饿") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected hunger warning, got %v", out)
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

// TestComputeDailyIncome 用临时 sqlite 数据库验证收益逻辑
func TestComputeDailyIncome(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tkeeper_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := initDB(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 插入一名玩家 + 三只状态各异的龟
	_, err = db.Exec("INSERT INTO players (id, coins, day, season) VALUES (?, 500, 1, 'spring')", "p1")
	if err != nil {
		t.Fatal(err)
	}
	insertTurtle := func(id string, vit, app, mood, intim int) {
		_, err := db.Exec(`INSERT INTO turtles (id, player_id, species, name, vitality, appetite, mood, intimacy) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, "p1", "muskTurtle", id, vit, app, mood, intim)
		if err != nil {
			t.Fatal(err)
		}
	}
	insertTurtle("t1", 90, 90, 80, 100) // 健康 + 心情好 + 亲密
	insertTurtle("t2", 30, 30, 20, 0)   // 病龟
	insertTurtle("t3", 80, 80, 50, 40)  // 中等

	total, breakdown := computeDailyIncome("p1")
	if total <= 0 {
		t.Errorf("expected positive total income, got %d", total)
	}
	if len(breakdown) != 3 {
		t.Errorf("expected 3 breakdown rows, got %d", len(breakdown))
	}
	t.Logf("daily income total=%d breakdown=%v", total, breakdown)
}
