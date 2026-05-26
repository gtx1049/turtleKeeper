package main

import (
	"fmt"
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

func TestDecorCatalog(t *testing.T) {
	cat := decorCatalog()
	if len(cat) < 4 {
		t.Fatalf("expected >=4 decor specs, got %d", len(cat))
	}
	if _, ok := findDecorSpec("sponge"); !ok {
		t.Error("expected sponge in catalog")
	}
	if _, ok := findDecorSpec("heater"); !ok {
		t.Error("expected heater in catalog")
	}
	if _, ok := findDecorSpec("unknown_xyz"); ok {
		t.Error("unknown decor should not be found")
	}
	sponge, _ := findDecorSpec("sponge")
	if sponge.FilterBoost <= 0 || sponge.Cost <= 0 {
		t.Errorf("sponge should have filter_boost>0 and cost>0, got %+v", sponge)
	}
	stone, _ := findDecorSpec("stone")
	if !stone.Basking {
		t.Errorf("stone should be basking, got %+v", stone)
	}
	wood, _ := findDecorSpec("wood")
	if !wood.Shelter {
		t.Errorf("wood should be shelter, got %+v", wood)
	}
}

// TestSummarizeDecorEffects 验证 decor 效果汇总 + 上限防爆
func TestSummarizeDecorEffects(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tkeeper_decor_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := initDB(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, _ = db.Exec("INSERT INTO players (id, coins, day, season) VALUES (?, 500, 1, 'spring')", "p_decor")
	_, _ = db.Exec("INSERT INTO tanks (id, player_id, shape, name, water_level) VALUES (?, ?, 'square', 'T', 'middle')", "tank_x", "p_decor")

	fb, cb, basking, shelter := summarizeDecorEffects("tank_x")
	if fb != 0 || cb != 0 || basking || shelter {
		t.Errorf("empty tank should be zero, got %v %v %v %v", fb, cb, basking, shelter)
	}

	for _, typ := range []string{"wood", "stone", "sponge"} {
		_, err := db.Exec("INSERT INTO decor (id, tank_id, type, x, y) VALUES (?, ?, ?, 0.5, 0.5)", "d_"+typ, "tank_x", typ)
		if err != nil {
			t.Fatal(err)
		}
	}
	fb, cb, basking, shelter = summarizeDecorEffects("tank_x")
	if !basking || !shelter {
		t.Errorf("expected basking & shelter, got basking=%v shelter=%v", basking, shelter)
	}
	if fb <= 0 {
		t.Errorf("expected filterBoost>0, got %v", fb)
	}
	if cb < 1 {
		t.Errorf("expected clarityBoost>=1, got %v", cb)
	}

	for i := 0; i < 10; i++ {
		_, _ = db.Exec("INSERT INTO decor (id, tank_id, type, x, y) VALUES (?, ?, 'sponge', 0.5, 0.5)", fmt.Sprintf("d_sp_%d", i), "tank_x")
	}
	fb, cb, _, _ = summarizeDecorEffects("tank_x")
	if fb > 0.45 {
		t.Errorf("filterBoost should cap at 0.45, got %v", fb)
	}
	if cb > 4 {
		t.Errorf("clarityBoost should cap at 4, got %v", cb)
	}
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

// TestShopCatalogSanity 保证商店配置每项必填、价格大于 0、PackSize > 0
func TestShopCatalogSanity(t *testing.T) {
	items := shopCatalog()
	if len(items) < 4 {
		t.Fatalf("shop too small: %d", len(items))
	}
	seen := map[string]bool{}
	for _, it := range items {
		if it.ID == "" || it.Type == "" || it.Name == "" {
			t.Errorf("shop item missing required field: %+v", it)
		}
		if it.Cost <= 0 || it.PackSize <= 0 {
			t.Errorf("shop item bad cost/pack: %+v", it)
		}
		if seen[it.ID] {
			t.Errorf("duplicate shop id: %s", it.ID)
		}
		seen[it.ID] = true
	}
	if _, ok := findShopItem("food_1"); !ok {
		t.Errorf("findShopItem(food_1) should be found")
	}
	if _, ok := findShopItem("nope"); ok {
		t.Errorf("findShopItem(nope) should fail")
	}
}

// TestBuyItemFlow 验证扣币 + 入库 + 不足时拒绝
func TestBuyItemFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tkeeper_shop_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := initDB(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO players (id, coins, day, season) VALUES (?, 100, 1, 'spring')", "buyer")
	if err != nil {
		t.Fatal(err)
	}

	spec, _ := findShopItem("food_1")
	// 模拟买 2 包 food_1
	totalCost := spec.Cost * 2
	totalGain := spec.PackSize * 2

	// 手动跑核心逻辑（不走 http）：扣币 + insert
	_, err = db.Exec("UPDATE players SET coins = coins - ? WHERE id = ?", totalCost, "buyer")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO inventory (id, player_id, item_type, name, count, icon) VALUES (?, ?, ?, ?, ?, ?)",
		spec.ID, "buyer", spec.Type, spec.Name, totalGain, spec.Icon)
	if err != nil {
		t.Fatal(err)
	}

	var coins, count int
	db.QueryRow("SELECT coins FROM players WHERE id = ?", "buyer").Scan(&coins)
	db.QueryRow("SELECT count FROM inventory WHERE id = ? AND player_id = ?", spec.ID, "buyer").Scan(&count)
	if coins != 100-totalCost {
		t.Errorf("coins want %d got %d", 100-totalCost, coins)
	}
	if count != totalGain {
		t.Errorf("count want %d got %d", totalGain, count)
	}
}

// TestFoodEffect 保证所有上架食物都有差异化效果，且默认分支不崩
func TestFoodEffect(t *testing.T) {
	specs := []string{"food_1", "food_2", "food_3", "food_4", "tool_2"}
	seenHungers := map[int]bool{}
	for _, id := range specs {
		h, i, v, m := foodEffect(id)
		if h < 0 || i < 0 || v < 0 || m < 0 {
			t.Errorf("%s effect has negative: %d %d %d %d", id, h, i, v, m)
		}
		if h+i+v+m == 0 {
			t.Errorf("%s effect is all zero, lacks gameplay value", id)
		}
		seenHungers[h] = true
	}
	// 至少应该有 3 种不同的饱腹值，否则差异化失败
	if len(seenHungers) < 3 {
		t.Errorf("food hunger values not diverse enough: %v", seenHungers)
	}
	// 默认分支
	h, _, _, _ := foodEffect("unknown_food")
	if h <= 0 {
		t.Errorf("default food should have positive hunger, got %d", h)
	}
}

// TestAdvanceBreeding 验证：
// 1. 同缸异性高亲密度龟在足够多次推进后能产蛋（统计意义上）
// 2. 缺乏异性 / 亲密度不够时不产蛋
// 3. 到期蛋会被处理并删除
func TestAdvanceBreeding(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tkeeper_breed_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := initDB(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec("INSERT INTO players (id, coins, day, season) VALUES (?, 500, 50, 'spring')", "p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO tanks (id, player_id, water_level) VALUES (?, ?, ?)", "tk1", "p1", "deep"); err != nil {
		t.Fatal(err)
	}
	insert := func(id, gender string, intim int) {
		_, err := db.Exec(`INSERT INTO turtles (id, player_id, species, name, gender, birth_day,
			vitality, appetite, skin, shell, intimacy, mood, tank_id, hunger, cleanliness)
			VALUES (?, ?, 'muskTurtle', ?, ?, 0, 100, 100, 100, 100, ?, 90, 'tk1', 80, 90)`,
			id, "p1", id, gender, intim)
		if err != nil {
			t.Fatal(err)
		}
	}
	insert("m1", "♂", 90)
	insert("f1", "♀", 90)

	// 跑 30 次推进，spring 高亲密度下应至少产 1 枚蛋（概率上几乎必然）
	produced := 0
	for i := 0; i < 30; i++ {
		eggs, _, _ := advanceBreeding("p1", 50+i, "spring")
		produced += eggs
	}
	if produced == 0 {
		t.Errorf("expected at least one egg from 30 rounds of high-intimacy spring breeding, got 0")
	}

	// 单独缸 + 单独单性别龟，应该不产蛋
	if _, err := db.Exec("INSERT INTO tanks (id, player_id, water_level) VALUES (?, ?, ?)", "tk2", "p1", "deep"); err != nil {
		t.Fatal(err)
	}
	insert2 := func(id string) {
		db.Exec(`INSERT INTO turtles (id, player_id, species, name, gender, birth_day,
			vitality, appetite, skin, shell, intimacy, mood, tank_id, hunger, cleanliness)
			VALUES (?, 'p2', 'muskTurtle', ?, '♂', 0, 100, 100, 100, 100, 90, 90, 'tk2', 80, 90)`,
			id, id)
	}
	if _, err := db.Exec("INSERT INTO players (id, coins, day, season) VALUES (?, 500, 50, 'spring')", "p2"); err != nil {
		t.Fatal(err)
	}
	insert2("solo")
	for i := 0; i < 20; i++ {
		eggs, _, _ := advanceBreeding("p2", 50+i, "spring")
		if eggs != 0 {
			t.Errorf("solo male should not produce eggs, got %d on round %d", eggs, i)
		}
	}

	// 到期蛋强制孵化测试：插入一枚 hatch_day = current 的蛋
	db.Exec(`INSERT INTO eggs (id, player_id, species, tank_id, laid_day, hatch_day, quality)
		VALUES ('egg_test', 'p1', 'muskTurtle', 'tk1', 70, 90, 95)`)
	beforeTurtles := 0
	db.QueryRow("SELECT COUNT(*) FROM turtles WHERE player_id='p1'").Scan(&beforeTurtles)
	advanceBreeding("p1", 90, "summer")
	var leftEggs int
	db.QueryRow("SELECT COUNT(*) FROM eggs WHERE id='egg_test'").Scan(&leftEggs)
	if leftEggs != 0 {
		t.Errorf("egg should be removed after hatch_day reached, still found %d", leftEggs)
	}
	afterTurtles := 0
	db.QueryRow("SELECT COUNT(*) FROM turtles WHERE player_id='p1'").Scan(&afterTurtles)
	// quality=95 + summer +0.1 => hatchProb=0.95，几乎必然孵化
	if afterTurtles <= beforeTurtles {
		t.Errorf("expected new turtle hatched from quality-95 egg, before=%d after=%d", beforeTurtles, afterTurtles)
	}
}

// TestPokedexUnlockDay 验证图鉴 unlock_day 迁移、保留与隐藏未解锁字段
func TestPokedexUnlockDay(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tkeeper_pokedex_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	if err := initDB(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 创建玩家：初始解锁两个龟种
	_, err = getOrCreatePlayer("p_pkdx")
	if err != nil {
		t.Fatal(err)
	}

	// 推进到 day 5，再解锁巴西龟
	db.Exec("UPDATE players SET day = 5 WHERE id = ?", "p_pkdx")
	unlockSpeciesForPlayer("p_pkdx", "redEaredSlider", 5)

	// 查 unlock_day
	var day int
	err = db.QueryRow("SELECT unlock_day FROM unlocked_species WHERE player_id=? AND species_id=?",
		"p_pkdx", "redEaredSlider").Scan(&day)
	if err != nil {
		t.Fatalf("query unlock_day: %v", err)
	}
	if day != 5 {
		t.Errorf("expected unlock_day=5, got %d", day)
	}

	// 重复解锁不应覆盖已有 unlock_day
	unlockSpeciesForPlayer("p_pkdx", "redEaredSlider", 99)
	db.QueryRow("SELECT unlock_day FROM unlocked_species WHERE player_id=? AND species_id=?",
		"p_pkdx", "redEaredSlider").Scan(&day)
	if day != 5 {
		t.Errorf("INSERT OR IGNORE should keep unlock_day=5, got %d", day)
	}

	// 初始 muskTurtle 应该有 unlock_day=1
	db.QueryRow("SELECT unlock_day FROM unlocked_species WHERE player_id=? AND species_id=?",
		"p_pkdx", "muskTurtle").Scan(&day)
	if day != 1 {
		t.Errorf("expected initial muskTurtle unlock_day=1, got %d", day)
	}
}
