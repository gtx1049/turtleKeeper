package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	mrand "math/rand"
	_ "modernc.org/sqlite"
	"net/http"
	"sync"
	"time"
)

// generateTurtleID 生成短随机字符串 ID,避免 UnixNano 大整数溢出 JS 安全范围。
func generateTurtleID() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = alphabet[mrand.Intn(len(alphabet))]
	}
	return "turtle_" + string(b)
}

// GameState 游戏状态
type GameState struct {
	PlayerID        string          `json:"player_id"`
	Coins           int             `json:"coins"`
	Day             int             `json:"day"`
	Season          string          `json:"season"`
	Turtles         []Turtle        `json:"turtles"`
	Tanks           []Tank          `json:"tanks"`
	Inventory       []InventoryItem `json:"inventory"`
	UnlockedSpecies []string        `json:"unlocked_species"`
	Achievements    []Achievement   `json:"achievements"`
	Eggs            []Egg           `json:"eggs"`
}

// Egg 龟蛋(M5 繁殖系统)
// 由同缸异性龟产下,孵化期满后变为新龟。
// 反映真实生物学:龟蛋需要一定天数才能孵化,且有成功率。
type Egg struct {
	ID           string `json:"id"`
	Species      string `json:"species"`
	TankID       string `json:"tank_id"`
	LaidDay      int    `json:"laid_day"`
	HatchDay     int    `json:"hatch_day"`
	ParentMomID  string `json:"parent_mom_id"`
	ParentDadID  string `json:"parent_dad_id"`
	Quality      int    `json:"quality"`       // 0-100,影响孵化成功率与子代初始属性
	DaysLeft     int    `json:"days_left"`     // 距离孵化还需多少天(不持久化,现算)
}

// Turtle 乌龟
type Turtle struct {
	ID             string                   `json:"id"`
	Species        string                   `json:"species"`
	Name           string                   `json:"name"`
	Gender         string                   `json:"gender"`
	BirthDay       int                      `json:"birth_day"`
	Weight         float64                  `json:"weight"`
	Personality    string                   `json:"personality"`
	Health         HealthStat               `json:"health"`
	Intimacy       int                      `json:"intimacy"`
	Melanism       int                      `json:"melanism"`
	TankID         string                   `json:"tank_id"`
	Hunger         int                      `json:"hunger"`
	Cleanliness    int                      `json:"cleanliness"`
	Mood           int                      `json:"mood"`
	Status         string                   `json:"status"`
	LastInteractDay int                     `json:"last_interact_day"`
	Suggestions    []map[string]interface{} `json:"suggestions,omitempty"`
}

// HealthStat 健康状态
type HealthStat struct {
	Vitality int `json:"vitality"`
	Appetite int `json:"appetite"`
	Skin     int `json:"skin"`
	Shell    int `json:"shell"`
}

// Tank 龟缸
type Tank struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"`
	Name         string                   `json:"name"`
	WaterLevel   string                   `json:"water_level"`
	Decor        []DecorItem              `json:"decor"`
	WaterQual    WaterStat                `json:"water_quality"`
	TempDay      float64                  `json:"temp_day"`
	TempNight    float64                  `json:"temp_night"`
	HasUVB       bool                     `json:"has_uvb"`
	HasFilter    bool                     `json:"has_filter"`
	WaterHistory []map[string]interface{} `json:"water_history,omitempty"`
}

// WaterStat 水质
type WaterStat struct {
	PH      float64 `json:"ph"`
	Ammonia float64 `json:"ammonia"`
	Nitrite float64 `json:"nitrite"`
	Clarity int     `json:"clarity"`
}

// DecorItem 造景元素
type DecorItem struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
	Scale    float64 `json:"scale"`
}

// DecorSpec 描述一种布景的元信息与游戏效果(M3 布景=机制)。
// effects 字段会被 advance-day 读取,把美观与玩法绑在一起。
type DecorSpec struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Icon       string  `json:"icon"`
	Desc       string  `json:"desc"`
	Cost       int     `json:"cost"`
	Category   string  `json:"category"` // shelter / equipment / plant / basking
	// 玩法效果:
	FilterBoost  float64 `json:"filter_boost,omitempty"`  // 降低水质衰减系数(0~1,越大越稳)
	ClarityBoost int     `json:"clarity_boost,omitempty"` // 每天清澈度回补
	Basking      bool    `json:"basking,omitempty"`       // 提供晒台(半水龟/水龟受益)
	Shelter      bool    `json:"shelter,omitempty"`       // 提供躲藏(提升心情)
}

// decorCatalog 是布景白名单(含 M3 新加的设备类)。
// handleAddDecor 用它校验 type、并在前端 /api/decor-catalog 暴露。
func decorCatalog() []DecorSpec {
	return []DecorSpec{
		{Type: "wood", Name: "沉木", Icon: "🪵", Desc: "麝香龟最爱钻洞,提升躲藏感。", Cost: 0, Category: "shelter", Shelter: true},
		{Type: "stone", Name: "晒台石", Icon: "🪨", Desc: "半水龟歇脚晒背的专属位。", Cost: 0, Category: "basking", Basking: true},
		{Type: "plant", Name: "水草丛", Icon: "🌿", Desc: "少量降氨吸硝,画面也更鲜活。", Cost: 0, Category: "plant"},
		{Type: "sponge", Name: "生化海绵", Icon: "🧽", Desc: "软过滤:没装过滤器的缸也能稳水。", Cost: 40, Category: "equipment", FilterBoost: 0.30, ClarityBoost: 1},
		{Type: "heater", Name: "加热棒", Icon: "🌡️", Desc: "冬天不掉温,半水龟体感舒适。", Cost: 60, Category: "equipment", FilterBoost: 0.10},
		{Type: "driftwood_basking", Name: "沉木晒台", Icon: "🪜", Desc: "高级晒台 + 躲藏二合一。", Cost: 80, Category: "basking", Basking: true, Shelter: true},
	}
}

// findDecorSpec O(n) 查表,类型有限不需要 map
func findDecorSpec(typ string) (DecorSpec, bool) {
	for _, d := range decorCatalog() {
		if d.Type == typ {
			return d, true
		}
	}
	return DecorSpec{}, false
}

// summarizeDecorEffects 把当前缸里的 decor 聚合成总效果,
// 供 advancePlayerTanks/Turtles 使用。
func summarizeDecorEffects(tankID string) (filterBoost float64, clarityBoost int, hasBasking, hasShelter bool) {
	rows, err := db.Query("SELECT type FROM decor WHERE tank_id = ?", tankID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var typ string
		if err := rows.Scan(&typ); err != nil {
			continue
		}
		spec, ok := findDecorSpec(typ)
		if !ok {
			continue
		}
		filterBoost += spec.FilterBoost
		clarityBoost += spec.ClarityBoost
		if spec.Basking {
			hasBasking = true
		}
		if spec.Shelter {
			hasShelter = true
		}
	}
	// 上限防爆
	if filterBoost > 0.45 {
		filterBoost = 0.45
	}
	if clarityBoost > 4 {
		clarityBoost = 4
	}
	return
}

// InventoryItem 背包物品
type InventoryItem struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Icon  string `json:"icon"`
}

// Achievement 成就
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Unlocked    bool   `json:"unlocked"`
	UnlockDay   int    `json:"unlock_day"`
}

// SpeciesInfo 龟种信息
type SpeciesInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	Difficulty      int    `json:"difficulty"`
	Description     string `json:"description"`
	UnlockCost      int    `json:"unlock_cost"`
	HabitatType     string `json:"habitat_type"`
	UnlockCondition string `json:"unlock_condition"`
	Trivia          string `json:"trivia"`           // 图鉴里展示的生物学小科普
	ScientificName  string `json:"scientific_name"`  // 学名
	NativeRegion    string `json:"native_region"`    // 原产地
	AdultSize       string `json:"adult_size"`       // 成年体型
}

// PokedexEntry 单条图鉴信息:基础龟种信息 + 玩家维度的解锁/拥有状态
type PokedexEntry struct {
	Species     SpeciesInfo `json:"species"`
	Unlocked    bool        `json:"unlocked"`
	UnlockDay   int         `json:"unlock_day"`
	OwnedCount  int         `json:"owned_count"`     // 当前在养的数量
	HatchedCount int        `json:"hatched_count"`   // 历史上孵化出来的数量(粗略:当前 birth_day>1 的)
}

// 全局数据库连接
var db *sql.DB

// advance-day 串行化锁：缓解并发写冲突（AI_TESTER P1）
var advanceDayMutex sync.Mutex

// 初始化数据库
func initDB(dataDir string) error {
	var err error
	db, err = sql.Open("sqlite", dataDir+"/turtlekeeper.db")
	if err != nil {
		return err
	}

	// WAL 模式 + 5 秒 busy timeout，缓解并发写冲突（AI_TESTER P1）
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Printf("警告: 无法设置 WAL 模式: %v", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=10000;"); err != nil {
		log.Printf("警告: 无法设置 busy_timeout: %v", err)
	}

	// 创建表
	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id TEXT PRIMARY KEY,
		coins INTEGER DEFAULT 500,
		day INTEGER DEFAULT 1,
		season TEXT DEFAULT 'spring',
		created_at INTEGER,
		last_played INTEGER
	);

	CREATE TABLE IF NOT EXISTS turtles (
		id TEXT PRIMARY KEY,
		player_id TEXT,
		species TEXT,
		name TEXT,
		gender TEXT,
		birth_day INTEGER DEFAULT 0,
		weight REAL DEFAULT 10.0,
		personality TEXT,
		vitality INTEGER DEFAULT 100,
		appetite INTEGER DEFAULT 100,
		skin INTEGER DEFAULT 100,
		shell INTEGER DEFAULT 100,
		intimacy INTEGER DEFAULT 0,
		melanism INTEGER DEFAULT 0,
		tank_id TEXT,
		hunger INTEGER DEFAULT 50,
		cleanliness INTEGER DEFAULT 80,
		mood INTEGER DEFAULT 70,
		status TEXT DEFAULT 'healthy',
		last_interact_day INTEGER DEFAULT 0,
		FOREIGN KEY (player_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS tanks (
		id TEXT PRIMARY KEY,
		player_id TEXT,
		type TEXT,
		name TEXT,
		water_level TEXT,
		temp_day REAL DEFAULT 26.0,
		temp_night REAL DEFAULT 24.0,
		has_uvb INTEGER DEFAULT 0,
		has_filter INTEGER DEFAULT 0,
		ph REAL DEFAULT 7.0,
		ammonia REAL DEFAULT 0.0,
		nitrite REAL DEFAULT 0.0,
		clarity INTEGER DEFAULT 100,
		FOREIGN KEY (player_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS decor (
		id TEXT PRIMARY KEY,
		tank_id TEXT,
		type TEXT,
		x REAL,
		y REAL,
		rotation REAL DEFAULT 0,
		scale REAL DEFAULT 1.0,
		FOREIGN KEY (tank_id) REFERENCES tanks(id)
	);

	CREATE TABLE IF NOT EXISTS inventory (
		id TEXT PRIMARY KEY,
		player_id TEXT,
		item_type TEXT,
		name TEXT,
		count INTEGER DEFAULT 0,
		icon TEXT,
		FOREIGN KEY (player_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS achievements (
		id TEXT PRIMARY KEY,
		player_id TEXT,
		name TEXT,
		description TEXT,
		unlocked INTEGER DEFAULT 0,
		unlock_day INTEGER DEFAULT 0,
		FOREIGN KEY (player_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS unlocked_species (
		player_id TEXT,
		species_id TEXT,
		unlock_day INTEGER DEFAULT 0,
		PRIMARY KEY (player_id, species_id)
	);

	CREATE TABLE IF NOT EXISTS water_history (
		tank_id TEXT,
		day INTEGER,
		ammonia REAL,
		nitrite REAL,
		clarity INTEGER,
		PRIMARY KEY (tank_id, day)
	);
	CREATE INDEX IF NOT EXISTS idx_water_history_tank ON water_history(tank_id, day);

	CREATE TABLE IF NOT EXISTS eggs (
		id TEXT PRIMARY KEY,
		player_id TEXT,
		species TEXT,
		tank_id TEXT,
		laid_day INTEGER,
		hatch_day INTEGER,
		parent_mom_id TEXT,
		parent_dad_id TEXT,
		quality INTEGER DEFAULT 60,
		FOREIGN KEY (player_id) REFERENCES players(id)
	);
	CREATE INDEX IF NOT EXISTS idx_eggs_player ON eggs(player_id);
	`

	if _, err = db.Exec(schema); err != nil {
		return err
	}

	// 在线迁移:给老库补 unlocked_species.unlock_day 列
	// ALTER TABLE ADD COLUMN 在 SQLite 里幂等性差,要先查 PRAGMA
	if !columnExists("unlocked_species", "unlock_day") {
		db.Exec("ALTER TABLE unlocked_species ADD COLUMN unlock_day INTEGER DEFAULT 0")
	}
	// 在线迁移:给老库补 turtles.status / last_interact_day 列 (AI_TESTER P2)
	if !columnExists("turtles", "status") {
		db.Exec("ALTER TABLE turtles ADD COLUMN status TEXT DEFAULT 'healthy'")
	}
	if !columnExists("turtles", "last_interact_day") {
		db.Exec("ALTER TABLE turtles ADD COLUMN last_interact_day INTEGER DEFAULT 0")
	}
	return nil
}

// columnExists 检查表里是否已有某列,用于幂等迁移
func columnExists(table, column string) bool {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if name == column {
			return true
		}
	}
	return false
}

// 获取或创建玩家
func getOrCreatePlayer(playerID string) (*GameState, error) {
	var player GameState
	player.PlayerID = playerID

	// 查询玩家
	row := db.QueryRow("SELECT coins, day, season FROM players WHERE id = ?", playerID)
	err := row.Scan(&player.Coins, &player.Day, &player.Season)

	if err == sql.ErrNoRows {
		// 创建新玩家
		now := time.Now().Unix()
		_, err = db.Exec(
			"INSERT INTO players (id, coins, day, season, created_at, last_played) VALUES (?, ?, ?, ?, ?, ?)",
			playerID, 500, 1, "spring", now, now,
		)
		if err != nil {
			return nil, err
		}
		player.Coins = 500
		player.Day = 1
		player.Season = "spring"

		// 创建初始乌龟
		initTurtles(playerID)
		// 创建初始龟缸
		initTanks(playerID)
		// 初始化背包
		initInventory(playerID)
		// 初始化成就
		initAchievements(playerID)
		// 解锁初始龟种
		initUnlockedSpecies(playerID)
	} else if err != nil {
		return nil, err
	}

	// 加载所有数据
	player.Turtles, _ = loadTurtles(playerID)
	player.Tanks, _ = loadTanks(playerID)
	player.Inventory, _ = loadInventory(playerID)
	player.Achievements, _ = loadAchievements(playerID)
	player.UnlockedSpecies, _ = loadUnlockedSpecies(playerID)
	player.Eggs, _ = loadEggs(playerID)

	// 迁移补齐:新增成就在老存档里可能不存在,补上去。
	ensureAchievementsExist(playerID, &player)

	return &player, nil
}

// ensureAchievementsExist 在老存档里补上后加的成就记录,以免玩家看不到。
func ensureAchievementsExist(playerID string, player *GameState) {
	defaults := []Achievement{
		{ID: "ach_1", Name: "初来乍到", Description: "获得第一只乌龟"},
		{ID: "ach_2", Name: "喂食新手", Description: "第一次喂食"},
		{ID: "ach_3", Name: "换水达人", Description: "第一次换水"},
		{ID: "ach_4", Name: "布景师", Description: "第一次布置造景"},
		{ID: "ach_5", Name: "破产边缘", Description: "第一次在商店购物"},
	}
	existing := map[string]bool{}
	for _, a := range player.Achievements {
		existing[a.ID] = true
	}
	changed := false
	for _, a := range defaults {
		if !existing[a.ID] {
			db.Exec("INSERT OR IGNORE INTO achievements (id, player_id, name, description) VALUES (?, ?, ?, ?)",
				a.ID, playerID, a.Name, a.Description)
			changed = true
		}
	}
	if changed {
		player.Achievements, _ = loadAchievements(playerID)
	}
}

// 初始化默认乌龟
func initTurtles(playerID string) {
	turtles := []Turtle{
		{
			ID: "turtle_1", Species: "muskTurtle", Name: "小麝香",
			Gender: "♀", Personality: "活泼",
			Health: HealthStat{Vitality: 100, Appetite: 100, Skin: 100, Shell: 100},
			TankID: "tank_1", Hunger: 50, Cleanliness: 80, Mood: 70,
		},
		{
			ID: "turtle_2", Species: "chinesePondTurtle", Name: "小草",
			Gender: "♂", Personality: "慵懒",
			Health: HealthStat{Vitality: 100, Appetite: 100, Skin: 100, Shell: 100},
			TankID: "tank_2", Hunger: 40, Cleanliness: 85, Mood: 75,
		},
	}

	for _, t := range turtles {
		db.Exec(`INSERT INTO turtles (id, player_id, species, name, gender, personality,
			vitality, appetite, skin, shell, intimacy, melanism, tank_id, hunger, cleanliness, mood)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.ID, playerID, t.Species, t.Name, t.Gender, t.Personality,
			t.Health.Vitality, t.Health.Appetite, t.Health.Skin, t.Health.Shell,
			t.Intimacy, t.Melanism, t.TankID, t.Hunger, t.Cleanliness, t.Mood,
		)
	}
}

// 初始化默认龟缸
func initTanks(playerID string) {
	tanks := []Tank{
		{
			ID: "tank_1", Type: "square", Name: "麝香的家",
			WaterLevel: "deep", TempDay: 26, TempNight: 24,
			HasUVB: false, HasFilter: true,
			WaterQual: WaterStat{PH: 7.0, Ammonia: 0, Nitrite: 0, Clarity: 100},
		},
		{
			ID: "tank_2", Type: "square", Name: "草龟的家",
			WaterLevel: "middle", TempDay: 25, TempNight: 22,
			HasUVB: true, HasFilter: false,
			WaterQual: WaterStat{PH: 7.2, Ammonia: 0, Nitrite: 0, Clarity: 95},
		},
	}

	for _, t := range tanks {
		db.Exec(`INSERT INTO tanks (id, player_id, type, name, water_level, temp_day, temp_night,
			has_uvb, has_filter, ph, ammonia, nitrite, clarity)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.ID, playerID, t.Type, t.Name, t.WaterLevel, t.TempDay, t.TempNight,
			boolToInt(t.HasUVB), boolToInt(t.HasFilter),
			t.WaterQual.PH, t.WaterQual.Ammonia, t.WaterQual.Nitrite, t.WaterQual.Clarity,
		)
	}
}

// 初始化背包
func initInventory(playerID string) {
	items := []InventoryItem{
		{ID: "food_1", Type: "food", Name: "龟粮", Count: 20, Icon: "🍖"},
		{ID: "food_2", Type: "food", Name: "红虫", Count: 10, Icon: "🪱"},
		{ID: "tool_1", Type: "tool", Name: "水质测试剂", Count: 5, Icon: "🧪"},
		{ID: "decor_1", Type: "decor", Name: "沉木", Count: 2, Icon: "🪵"},
	}

	for _, item := range items {
		db.Exec("INSERT INTO inventory (id, player_id, item_type, name, count, icon) VALUES (?, ?, ?, ?, ?, ?)",
			item.ID, playerID, item.Type, item.Name, item.Count, item.Icon)
	}
}

// 初始化成就
func initAchievements(playerID string) {
	achievements := []Achievement{
		{ID: "ach_1", Name: "初来乍到", Description: "获得第一只乌龟"},
		{ID: "ach_2", Name: "喂食新手", Description: "第一次喂食"},
		{ID: "ach_3", Name: "换水达人", Description: "第一次换水"},
		{ID: "ach_4", Name: "布景师", Description: "第一次布置造景"},
		{ID: "ach_5", Name: "破产边缘", Description: "第一次在商店购物"},
	}

	for _, a := range achievements {
		db.Exec("INSERT INTO achievements (id, player_id, name, description) VALUES (?, ?, ?, ?)",
			a.ID, playerID, a.Name, a.Description)
	}
}

// 初始化解锁龟种
func initUnlockedSpecies(playerID string) {
	species := []string{"muskTurtle", "chinesePondTurtle"}
	for _, s := range species {
		db.Exec("INSERT OR IGNORE INTO unlocked_species (player_id, species_id, unlock_day) VALUES (?, ?, ?)",
			playerID, s, 1)
	}
}

// unlockSpeciesForPlayer 给玩家解锁某龟种,记录解锁日。
// 已解锁则保留旧 unlock_day,不覆盖。
func unlockSpeciesForPlayer(playerID, speciesID string, day int) {
	db.Exec("INSERT OR IGNORE INTO unlocked_species (player_id, species_id, unlock_day) VALUES (?, ?, ?)",
		playerID, speciesID, day)
}

// 加载乌龟
func loadTurtles(playerID string) ([]Turtle, error) {
	rows, err := db.Query(`SELECT id, species, name, gender, birth_day, weight, personality,
		vitality, appetite, skin, shell, intimacy, melanism, tank_id, hunger, cleanliness, mood, status, last_interact_day
		FROM turtles WHERE player_id = ?`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turtles []Turtle
	for rows.Next() {
		var t Turtle
		rows.Scan(&t.ID, &t.Species, &t.Name, &t.Gender, &t.BirthDay, &t.Weight, &t.Personality,
			&t.Health.Vitality, &t.Health.Appetite, &t.Health.Skin, &t.Health.Shell,
			&t.Intimacy, &t.Melanism, &t.TankID, &t.Hunger, &t.Cleanliness, &t.Mood, &t.Status, &t.LastInteractDay)
		turtles = append(turtles, t)
	}
	return turtles, nil
}

// 加载龟缸
func loadTanks(playerID string) ([]Tank, error) {
	rows, err := db.Query(`SELECT id, type, name, water_level, temp_day, temp_night,
		has_uvb, has_filter, ph, ammonia, nitrite, clarity
		FROM tanks WHERE player_id = ?`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tanks []Tank
	for rows.Next() {
		var t Tank
		var hasUVB, hasFilter int
		rows.Scan(&t.ID, &t.Type, &t.Name, &t.WaterLevel, &t.TempDay, &t.TempNight,
			&hasUVB, &hasFilter, &t.WaterQual.PH, &t.WaterQual.Ammonia, &t.WaterQual.Nitrite, &t.WaterQual.Clarity)
		t.HasUVB = hasUVB == 1
		t.HasFilter = hasFilter == 1

		// 加载造景
		t.Decor, _ = loadDecor(t.ID)
		tanks = append(tanks, t)
	}
	return tanks, nil
}

// 加载造景
func loadDecor(tankID string) ([]DecorItem, error) {
	rows, err := db.Query("SELECT id, type, x, y, rotation, scale FROM decor WHERE tank_id = ?", tankID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DecorItem
	for rows.Next() {
		var d DecorItem
		rows.Scan(&d.ID, &d.Type, &d.X, &d.Y, &d.Rotation, &d.Scale)
		items = append(items, d)
	}
	return items, nil
}

// 加载背包
func loadInventory(playerID string) ([]InventoryItem, error) {
	rows, err := db.Query("SELECT id, item_type, name, count, icon FROM inventory WHERE player_id = ?", playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InventoryItem
	for rows.Next() {
		var i InventoryItem
		rows.Scan(&i.ID, &i.Type, &i.Name, &i.Count, &i.Icon)
		items = append(items, i)
	}
	return items, nil
}

// 加载成就
func loadAchievements(playerID string) ([]Achievement, error) {
	rows, err := db.Query("SELECT id, name, description, unlocked, unlock_day FROM achievements WHERE player_id = ?", playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []Achievement
	for rows.Next() {
		var a Achievement
		var unlocked int
		rows.Scan(&a.ID, &a.Name, &a.Description, &unlocked, &a.UnlockDay)
		a.Unlocked = unlocked == 1
		achievements = append(achievements, a)
	}
	return achievements, nil
}

// 加载已解锁龟种
func loadUnlockedSpecies(playerID string) ([]string, error) {
	rows, err := db.Query("SELECT species_id FROM unlocked_species WHERE player_id = ?", playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var species []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		species = append(species, s)
	}
	return species, nil
}

// loadEggs 加载玩家所有未孵化龟蛋
func loadEggs(playerID string) ([]Egg, error) {
	var day int
	_ = db.QueryRow("SELECT day FROM players WHERE id = ?", playerID).Scan(&day)
	rows, err := db.Query(`SELECT id, species, tank_id, laid_day, hatch_day,
		COALESCE(parent_mom_id,''), COALESCE(parent_dad_id,''), quality
		FROM eggs WHERE player_id = ? ORDER BY hatch_day ASC`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eggs []Egg
	for rows.Next() {
		var e Egg
		rows.Scan(&e.ID, &e.Species, &e.TankID, &e.LaidDay, &e.HatchDay,
			&e.ParentMomID, &e.ParentDadID, &e.Quality)
		e.DaysLeft = e.HatchDay - day
		if e.DaysLeft < 0 {
			e.DaysLeft = 0
		}
		eggs = append(eggs, e)
	}
	return eggs, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// API 处理函数

func handleGetState(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		playerID = "default"
	}

	state, err := getOrCreatePlayer(playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 为每只龟追加智能建议(最多 2 条高优先级),让玩家一打开游戏就知道该干什么
	for i := range state.Turtles {
		t := &state.Turtles[i]
		sp, _ := findSpecies(t.Species)
		var tankMap map[string]interface{}
		for _, tank := range state.Tanks {
			if tank.ID == t.TankID {
				tankMap = map[string]interface{}{
					"water_level": tank.WaterLevel,
					"ammonia":     tank.WaterQual.Ammonia,
					"nitrite":     tank.WaterQual.Nitrite,
					"clarity":     tank.WaterQual.Clarity,
					"has_filter":  tank.HasFilter,
					"has_uvb":     tank.HasUVB,
				}
				break
			}
		}
		t.Suggestions = buildTurtleSuggestions(*t, sp, tankMap)
		// 限制最多 2 条,避免 state 膨胀
		if len(t.Suggestions) > 2 {
			t.Suggestions = t.Suggestions[:2]
		}
	}

	// 为每个龟缸追加最近 14 天水质历史,方便前端趋势展示
	for i := range state.Tanks {
		state.Tanks[i].WaterHistory = loadWaterHistory(state.Tanks[i].ID, 14)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func handleFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string `json:"player_id"`
		TurtleID string `json:"turtle_id"`
		FoodID   string `json:"food_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}

	// 验证背包里还有该食物,避免刷接口刷负数
	var haveCount int
	err := db.QueryRow("SELECT count FROM inventory WHERE id = ? AND player_id = ?", req.FoodID, req.PlayerID).Scan(&haveCount)
	if err == sql.ErrNoRows || haveCount <= 0 {
		http.Error(w, "该食物已用完,请到商店补货", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 不同食物效果不同,记住买什么不仅仅是堆数量。
	hungerDelta, intimacyDelta, vitalityDelta, moodDelta := foodEffect(req.FoodID)
	db.Exec(`UPDATE turtles SET
		hunger     = MIN(100, hunger + ?),
		intimacy   = MIN(100, intimacy + ?),
		vitality   = MIN(100, vitality + ?),
		mood       = MIN(100, mood + ?)
		WHERE id = ?`, hungerDelta, intimacyDelta, vitalityDelta, moodDelta, req.TurtleID)

	// 减少食物数量
	db.Exec("UPDATE inventory SET count = count - 1 WHERE id = ? AND player_id = ?", req.FoodID, req.PlayerID)

	// 喂食成就
	db.Exec("UPDATE achievements SET unlocked = 1, unlock_day = (SELECT day FROM players WHERE id = ?) WHERE player_id = ? AND id = 'ach_2' AND unlocked = 0", req.PlayerID, req.PlayerID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ok",
		"hunger_delta":   hungerDelta,
		"intimacy_delta": intimacyDelta,
		"vitality_delta": vitalityDelta,
		"mood_delta":     moodDelta,
	})
}

// foodEffect 返回指定食物 id 的四项加成 (hunger, intimacy, vitality, mood)
// 跟 shopCatalog 里的描述保持一致,避免口说无凭。
func foodEffect(foodID string) (hunger, intimacy, vitality, mood int) {
	switch foodID {
	case "food_1": // 龟粮:主粮均衡
		return 30, 2, 0, 1
	case "food_2": // 红虫:亲密度极高
		return 22, 6, 1, 3
	case "food_3": // 虾干:饱腹强、提甲
		return 38, 1, 2, 0
	case "food_4": // 小鱼苗:野性填充 + 活力
		return 40, 4, 5, 2
	case "tool_2": // 维生素片:补充品,不填股
		return 0, 1, 20, 4
	default: // 未知食物走默认值
		return 25, 2, 0, 0
	}
}

func handleClean(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string `json:"player_id"`
		TankID   string `json:"tank_id"`
		TurtleID string `json:"turtle_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.PlayerID == "" {
		req.PlayerID = "default"
	}
	// 支持 turtle_id 自动解析所属 tank,提升 API 易用性
	if req.TankID == "" && req.TurtleID != "" {
		db.QueryRow("SELECT tank_id FROM turtles WHERE id = ? AND player_id = ?", req.TurtleID, req.PlayerID).Scan(&req.TankID)
	}

	result, err := maintainTank(req.PlayerID, req.TankID, "deep_clean")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleMaintainTank(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string `json:"player_id"`
		TankID   string `json:"tank_id"`
		Mode     string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}

	result, err := maintainTank(req.PlayerID, req.TankID, req.Mode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func maintainTank(playerID, tankID, mode string) (map[string]interface{}, error) {
	if tankID == "" {
		return nil, fmt.Errorf("tank_id required")
	}
	if mode == "" {
		mode = "partial_change"
	}

	var hasFilter int
	var ammonia, nitrite float64
	var clarity int
	if err := db.QueryRow(`SELECT has_filter, ammonia, nitrite, clarity
		FROM tanks WHERE id = ? AND player_id = ?`, tankID, playerID).Scan(&hasFilter, &ammonia, &nitrite, &clarity); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tank not found")
		}
		return nil, err
	}

	cost := 0
	message := ""
	switch mode {
	case "scoop_waste":
		// 普通清理:免费但效果有限,适合日常点一下。
		clarity = clampInt(clarity+18, 0, 100)
		ammonia = roundFloat(clampFloat(ammonia*0.78, 0, 5), 2)
		nitrite = roundFloat(clampFloat(nitrite*0.86, 0, 5), 2)
		message = "已捞走残饵和粪便,清澈度小幅回升"
	case "partial_change":
		// 部分换水:低成本,把水质拉回安全区但不清零。
		cost = 20
		clarity = clampInt(clarity+35, 0, 100)
		ammonia = roundFloat(clampFloat(ammonia*0.42, 0, 5), 2)
		nitrite = roundFloat(clampFloat(nitrite*0.50, 0, 5), 2)
		message = "完成 1/3 换水,水质明显好转"
	case "deep_clean":
		// 深度清洁:保留旧 /api/clean 的一键重置语义,但加入经营成本。
		// 若水质已良好,提示无需清洁并免扣费。
		if clarity >= 90 && ammonia < 0.3 {
			return map[string]interface{}{
				"status":        "ok",
				"mode":          mode,
				"cost":          0,
				"message":       "水质良好,无需深度清洁",
				"has_filter":    hasFilter == 1,
				"water_quality": WaterStat{PH: 7.0, Ammonia: ammonia, Nitrite: nitrite, Clarity: clarity},
			}, nil
		}
		cost = 60
		clarity = 100
		ammonia = 0
		nitrite = 0
		message = "深度清洁完成,龟缸恢复清爽"
	case "install_filter":
		if hasFilter == 1 {
			return nil, fmt.Errorf("这个龟缸已经有过滤器了")
		}
		cost = 180
		hasFilter = 1
		clarity = clampInt(clarity+25, 0, 100)
		ammonia = roundFloat(clampFloat(ammonia*0.70, 0, 5), 2)
		nitrite = roundFloat(clampFloat(nitrite*0.78, 0, 5), 2)
		message = "过滤器安装完成,之后水质恶化会变慢"
	default:
		return nil, fmt.Errorf("unknown maintenance mode")
	}

	var coins int
	if err := db.QueryRow("SELECT coins FROM players WHERE id = ?", playerID).Scan(&coins); err != nil {
		return nil, fmt.Errorf("player not found")
	}
	if coins < cost {
		return nil, fmt.Errorf("龟币不足,还差 %d", cost-coins)
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE players SET coins = coins - ? WHERE id = ?", cost, playerID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec("UPDATE tanks SET has_filter = ?, ph = 7.0, ammonia = ?, nitrite = ?, clarity = ? WHERE id = ? AND player_id = ?",
		hasFilter, ammonia, nitrite, clarity, tankID, playerID); err != nil {
		return nil, err
	}
	cleanBoost := map[string]int{"scoop_waste": 8, "partial_change": 18, "deep_clean": 35, "install_filter": 6}[mode]
	if _, err := tx.Exec("UPDATE turtles SET cleanliness = MIN(100, cleanliness + ?), mood = MIN(100, mood + 2) WHERE player_id = ? AND tank_id = ?",
		cleanBoost, playerID, tankID); err != nil {
		return nil, err
	}
	if mode == "partial_change" || mode == "deep_clean" {
		_, _ = tx.Exec("UPDATE achievements SET unlocked = 1, unlock_day = (SELECT day FROM players WHERE id = ?) WHERE player_id = ? AND id = 'ach_3' AND unlocked = 0", playerID, playerID)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":        "ok",
		"mode":          mode,
		"cost":          cost,
		"message":       message,
		"has_filter":    hasFilter == 1,
		"water_quality": WaterStat{PH: 7.0, Ammonia: ammonia, Nitrite: nitrite, Clarity: clarity},
	}, nil
}

func handleInteract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string `json:"player_id"`
		TurtleID string `json:"turtle_id"`
		Action   string `json:"action"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}

	switch req.Action {
	case "pet":
		// AI_TESTER P2: 每只龟每天亲密增长上限 +5（即每天只能 pet 涨一次亲密）
		var lastInteractDay, playerDay int
		db.QueryRow("SELECT last_interact_day FROM turtles WHERE id = ?", req.TurtleID).Scan(&lastInteractDay)
		db.QueryRow("SELECT day FROM players WHERE id = ?", req.PlayerID).Scan(&playerDay)
		if lastInteractDay == playerDay {
			// 今天已经互动过，只涨心情不涨亲密
			db.Exec("UPDATE turtles SET mood = MIN(100, mood + 3) WHERE id = ?", req.TurtleID)
		} else {
			db.Exec("UPDATE turtles SET intimacy = MIN(100, intimacy + 5), mood = MIN(100, mood + 3), last_interact_day = ? WHERE id = ?", playerDay, req.TurtleID)
		}
	case "play":
		db.Exec("UPDATE turtles SET mood = MIN(100, mood + 8), hunger = MAX(0, hunger - 5) WHERE id = ?", req.TurtleID)
	case "check":
		// 检查健康状态,返回详细建议
		var t Turtle
		if err := db.QueryRow(`SELECT id, species, name, gender, birth_day, weight, personality,
			hunger, cleanliness, mood, intimacy, vitality, appetite, skin, shell, tank_id, status, last_interact_day
			FROM turtles WHERE id = ? AND player_id = ?`, req.TurtleID, req.PlayerID).
			Scan(&t.ID, &t.Species, &t.Name, &t.Gender, &t.BirthDay, &t.Weight, &t.Personality,
				&t.Hunger, &t.Cleanliness, &t.Mood, &t.Intimacy, &t.Health.Vitality, &t.Health.Appetite, &t.Health.Skin, &t.Health.Shell, &t.TankID, &t.Status, &t.LastInteractDay); err != nil {
			http.Error(w, "turtle not found", http.StatusNotFound)
			return
		}
		sp, _ := findSpecies(t.Species)
		var tankData map[string]interface{}
		if t.TankID != "" {
			var ph, ammonia, nitrite float64
			var clarity int
			var hasFilter, hasUVB bool
			if err := db.QueryRow(`SELECT ph, ammonia, nitrite, clarity, has_filter, has_uvb, water_level FROM tanks WHERE id = ?`, t.TankID).
				Scan(&ph, &ammonia, &nitrite, &clarity, &hasFilter, &hasUVB, &tankData); err == nil {
				tankData = map[string]interface{}{
					"ph": ph, "ammonia": ammonia, "nitrite": nitrite, "clarity": clarity,
					"has_filter": hasFilter, "has_uvb": hasUVB, "water_level": tankData,
				}
			}
		}
		suggestions := buildTurtleSuggestions(t, sp, tankData)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok", "turtle": t, "suggestions": suggestions,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleAddDecor(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string    `json:"player_id"`
		TankID   string    `json:"tank_id"`
		Decor    DecorItem `json:"decor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.TankID == "" || req.Decor.Type == "" {
		http.Error(w, "tank_id and decor.type required", http.StatusBadRequest)
		return
	}
	spec, specOK := findDecorSpec(req.Decor.Type)
	if !specOK {
		http.Error(w, "unknown decor type: "+req.Decor.Type, http.StatusBadRequest)
		return
	}
	// 检查越限购买(cost > 0 的类需要扣龟币)
	if spec.Cost > 0 {
		var coins int
		if err := db.QueryRow("SELECT coins FROM players WHERE id = ?", req.PlayerID).Scan(&coins); err != nil {
			http.Error(w, "player not found", http.StatusBadRequest)
			return
		}
		if coins < spec.Cost {
			http.Error(w, fmt.Sprintf("龟币不足,需要 %d", spec.Cost), http.StatusBadRequest)
			return
		}
	}
	if req.Decor.ID == "" {
		req.Decor.ID = fmt.Sprintf("decor_%d", time.Now().UnixNano())
	}
	if req.Decor.Scale <= 0 {
		req.Decor.Scale = 1
	}
	if req.Decor.X < 0 {
		req.Decor.X = 0
	}
	if req.Decor.X > 1 {
		req.Decor.X = 1
	}
	if req.Decor.Y < 0 {
		req.Decor.Y = 0
	}
	if req.Decor.Y > 1 {
		req.Decor.Y = 1
	}

	if _, err := db.Exec("INSERT INTO decor (id, tank_id, type, x, y, rotation, scale) VALUES (?, ?, ?, ?, ?, ?, ?)",
		req.Decor.ID, req.TankID, req.Decor.Type, req.Decor.X, req.Decor.Y, req.Decor.Rotation, req.Decor.Scale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 扣除 cost(如果有)
	if spec.Cost > 0 {
		db.Exec("UPDATE players SET coins = coins - ? WHERE id = ?", spec.Cost, req.PlayerID)
	}
	// 装上海绵同步点亮 has_filter(软过滤也算过滤)。拳头产品。
	if req.Decor.Type == "sponge" {
		db.Exec("UPDATE tanks SET has_filter = 1 WHERE id = ? AND player_id = ?", req.TankID, req.PlayerID)
	}

	// 第一次布景奖励:给一点龟币和成就反馈,让系统有正向循环。
	db.Exec("UPDATE achievements SET unlocked = 1, unlock_day = (SELECT day FROM players WHERE id = ?) WHERE player_id = ? AND id = 'ach_4' AND unlocked = 0", req.PlayerID, req.PlayerID)
	// 免费装饰品才送龟币,避免装备送贺雙获利
	if spec.Cost == 0 {
		db.Exec("UPDATE players SET coins = coins + 20 WHERE id = ?", req.PlayerID)
	}

	// M3 生态评分:计算该缸当前布景得分并返回
	score := calculateTankDecorScore(req.TankID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"decor":        req.Decor,
		"ecoscore":     score.Score,
		"ecoscore_max": score.Max,
		"ecoscore_tips": score.Tips,
	})
}

// EcoScore 布景生态评分结果
type EcoScore struct {
	Score int    `json:"score"`
	Max   int    `json:"max"`
	Tips  string `json:"tips"`
}

func calculateTankDecorScore(tankID string) EcoScore {
	rows, err := db.Query("SELECT type FROM decor WHERE tank_id = ?", tankID)
	if err != nil {
		return EcoScore{Score: 0, Max: 100, Tips: "无法读取造景数据"}
	}
	defer rows.Close()

	counts := map[string]int{}
	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			counts[t]++
			if counts[t] == 1 {
				types = append(types, t)
			}
		}
	}

	// 基础分
	base := map[string]int{"wood": 10, "stone": 10, "plant": 15, "sponge": 10, "heater": 10, "driftwood_basking": 15}
	score := 0
	for t, c := range counts {
		b := base[t]
		if b == 0 {
			b = 8
		}
		// 同类型堆叠递减:第1个100%,第2个70%,第3个49%...
		for i := 0; i < c; i++ {
			mul := 1.0
			for j := 0; j < i; j++ {
				mul *= 0.7
			}
			score += int(float64(b) * mul)
		}
	}

	// 多样性 bonus
	if len(types) >= 3 {
		score += 15
	}
	if len(types) >= 5 {
		score += 15
	}

	// 查询 tank 类型匹配 bonus
	var wl string
	db.QueryRow("SELECT water_level FROM tanks WHERE id = ?", tankID).Scan(&wl)
	pref := map[string][]string{
		"deep":   {"wood", "plant", "sponge", "heater"},
		"middle": {"wood", "stone", "plant", "heater"},
		"land":   {"stone", "driftwood_basking", "plant"},
	}
	if prefs, ok := pref[wl]; ok {
		for _, p := range prefs {
			if counts[p] > 0 {
				score += 5
			}
		}
	}

	if score > 100 {
		score = 100
	}

	tips := ""
	switch {
	case score >= 80:
		tips = "生态完美！龟龟幸福感爆棚"
	case score >= 60:
		tips = "环境不错，继续丰富造景"
	case score >= 40:
		tips = "略显单调，多加点植物或躲避"
	default:
		tips = "龟缸太空了，快布置一下吧"
	}
	if wl == "deep" && counts["sponge"] == 0 && counts["heater"] == 0 {
		tips = "深水缸建议至少配备过滤或加热设备"
	}
	if wl == "land" && counts["driftwood_basking"] == 0 {
		tips = "半水陆缸缺少晒台，龟龟无法晒背"
	}

	return EcoScore{Score: score, Max: 100, Tips: tips}
}

func handleMoveDecor(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string  `json:"player_id"`
		DecorID  string  `json:"decor_id"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.DecorID == "" {
		http.Error(w, "decor_id required", http.StatusBadRequest)
		return
	}
	if req.X < 0 {
		req.X = 0
	}
	if req.X > 1 {
		req.X = 1
	}
	if req.Y < 0 {
		req.Y = 0
	}
	if req.Y > 1 {
		req.Y = 1
	}

	res, err := db.Exec(`UPDATE decor
		SET x = ?, y = ?
		WHERE id = ? AND tank_id IN (SELECT id FROM tanks WHERE player_id = ?)`, req.X, req.Y, req.DecorID, req.PlayerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	changed, _ := res.RowsAffected()
	if changed == 0 {
		http.Error(w, "decor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleAdvanceDay(w http.ResponseWriter, r *http.Request) {
	advanceDayMutex.Lock()
	defer advanceDayMutex.Unlock()

	var req struct {
		PlayerID string `json:"player_id"`
	}
	// 允许空 body(curl / 调试友好);只在 body 非空时严格解码。
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}

	writeErr := func(code int, msg string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": msg})
	}

	// 懒初始化:所有写 API 都走这一步,避免冷启动 + 直接调用就 player not found。
	if _, err := getOrCreatePlayer(req.PlayerID); err != nil {
		writeErr(http.StatusInternalServerError, err.Error())
		return
	}

	// 获取当前天数
	var day int
	if err := db.QueryRow("SELECT day FROM players WHERE id = ?", req.PlayerID).Scan(&day); err != nil {
		writeErr(http.StatusBadRequest, "player not found")
		return
	}
	day++

	// 更新天数和季节
	season := getSeason(day)
	if _, err := db.Exec("UPDATE players SET day = ?, season = ?, last_played = ? WHERE id = ?", day, season, time.Now().Unix(), req.PlayerID); err != nil {
		writeErr(http.StatusInternalServerError, err.Error())
		return
	}

	// M5 水质时间系统:所有衰减都限定在当前玩家,避免多玩家互相污染。
	// 龟越多、无过滤、水草少时水质恶化更快;清洁度低会继续拖累心情和健康。
	if err := advancePlayerTanks(req.PlayerID); err != nil {
		writeErr(http.StatusInternalServerError, err.Error())
		return
	}
	if err := advancePlayerTurtles(req.PlayerID); err != nil {
		writeErr(http.StatusInternalServerError, err.Error())
		return
	}

	// M4 经济正循环:根据健康龟数发放日常龟币(健康/亲密度加成),
	// 让玩家不会因为长时间挂机直接破产。
	income, incomeBreakdown := computeDailyIncome(req.PlayerID)
	if income > 0 {
		db.Exec("UPDATE players SET coins = coins + ? WHERE id = ?", income, req.PlayerID)
	}

	// M5 季节性事件提示(不写库,只回传前端用作 toast/弹幕)。
	seasonEvent := seasonalEvent(season, day, req.PlayerID)

	// M5 繁殖系统:同缸异性高亲密产蛋,到期孵化。
	neweggs, newhatches, breedMsgs := advanceBreeding(req.PlayerID, day, season)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"day":              day,
		"season":           season,
		"income":           income,
		"income_breakdown": incomeBreakdown,
		"season_event":     seasonEvent,
		"new_eggs":         neweggs,
		"new_hatches":      newhatches,
		"breed_messages":   breedMsgs,
	})
}

// computeDailyIncome 根据每只龟的健康/亲密度结算"萌宠收益"。
// 设计目标:3-5 只健康龟 / 天 ≈ 30-80 龟币,长期可负担基础食物+维护。
func computeDailyIncome(playerID string) (int, []map[string]interface{}) {
	// AI_TESTER P2: income 与 health 强挂钩，四维均值决定基础系数
	rows, err := db.Query(`SELECT id, name, vitality, appetite, skin, shell, mood, intimacy FROM turtles WHERE player_id = ?`, playerID)
	if err != nil {
		return 0, nil
	}
	defer rows.Close()

	total := 0
	var breakdown []map[string]interface{}
	for rows.Next() {
		var id, name string
		var vit, app, skin, shell, mood, intim int
		if err := rows.Scan(&id, &name, &vit, &app, &skin, &shell, &mood, &intim); err != nil {
			continue
		}
		// 基础 5;心情 70+ 加 3;亲密度每 20 加 1(封顶 5)。
		coins := 5
		if mood >= 70 {
			coins += 3
		}
		intimBonus := intim / 20
		if intimBonus > 5 {
			intimBonus = 5
		}
		coins += intimBonus

		// 健康四维均值决定收入系数: 健康养龟=赚更多
		healthAvg := (vit + app + skin + shell) / 4
		if healthAvg >= 80 {
			coins += 4
		} else if healthAvg >= 60 {
			coins += 2
		} else if healthAvg >= 40 {
			coins += 0
		} else {
			coins -= 3 // 病龟反而要花钱治疗(隐喻)
		}

		if coins < 0 {
			coins = 0
		}
		total += coins
		breakdown = append(breakdown, map[string]interface{}{
			"turtle_id": id,
			"name":      name,
			"coins":     coins,
			"health_avg": healthAvg,
		})
	}
	return total, breakdown
}

// seasonalEvent 给前端展示季节小事件,配合 M5 时间系统。
// 不真改龟属性(避免和水质系统打架),但让玩家感受到节律。
func seasonalEvent(season string, day int, playerID string) map[string]interface{} {
	dayInSeason := day % 30
	switch season {
	case "spring":
		if dayInSeason == 5 {
			return map[string]interface{}{"type": "breeding_hint", "icon": "💕", "text": "春暖,龟们开始追逐求偶。同缸异性高亲密龟有机会产蛋。"}
		}
	case "summer":
		if dayInSeason == 5 {
			return map[string]interface{}{"type": "heat_warning", "icon": "🔥", "text": "夏季高温,注意换水频率和遮阴"}
		}
		if dayInSeason == 18 {
			return map[string]interface{}{"type": "feast", "icon": "🍤", "text": "伏天龟食欲旺,多投喂可加速成长"}
		}
	case "autumn":
		if dayInSeason == 5 {
			return map[string]interface{}{"type": "fatten", "icon": "🍂", "text": "秋季囤膘期,多喂红虫/小鱼储备脂肪"}
		}
	case "winter":
		if dayInSeason == 2 {
			return map[string]interface{}{"type": "hibernate_hint", "icon": "❄️", "text": "水温下降,龟将进入半冬眠(暂未实装详细系统)"}
		}
		if dayInSeason == 15 {
			return map[string]interface{}{"type": "warm_check", "icon": "🛁", "text": "寒潮中,检查加热棒和过滤器是否在线"}
		}
	}
	return nil
}

func advancePlayerTanks(playerID string) error {
	type tankDecay struct {
		ID        string
		HasFilter int
		Ammonia   float64
		Nitrite   float64
		Clarity   int
	}

	rows, err := db.Query(`SELECT id, has_filter, ammonia, nitrite, clarity FROM tanks WHERE player_id = ?`, playerID)
	if err != nil {
		return err
	}

	var tanks []tankDecay
	for rows.Next() {
		var t tankDecay
		if err := rows.Scan(&t.ID, &t.HasFilter, &t.Ammonia, &t.Nitrite, &t.Clarity); err != nil {
			rows.Close()
			return err
		}
		tanks = append(tanks, t)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, tank := range tanks {
		var turtleCount, plantCount int
		db.QueryRow("SELECT COUNT(*) FROM turtles WHERE player_id = ? AND tank_id = ?", playerID, tank.ID).Scan(&turtleCount)
		db.QueryRow("SELECT COUNT(*) FROM decor WHERE tank_id = ? AND type = 'plant'", tank.ID).Scan(&plantCount)

		// M3 布景=机制:设备类布景对水质衰减的叠加补助。
		decorFilter, decorClarity, _, _ := summarizeDecorEffects(tank.ID)

		bioLoad := 1.0 + float64(maxInt(0, turtleCount-1))*0.35
		filterFactor := 1.0
		clarityDrop := 4
		if tank.HasFilter == 1 {
			filterFactor = 0.55
			clarityDrop = 2
		}
		// 软过滤/加热棒进一步压低衰减系数
		if decorFilter > 0 {
			filterFactor = clampFloat(filterFactor-decorFilter, 0.20, 1.0)
		}
		plantBonus := float64(minInt(3, plantCount)) * 0.08
		ammonia := roundFloat(clampFloat(tank.Ammonia+(0.12*bioLoad*filterFactor)-plantBonus, 0, 5), 2)
		nitrite := roundFloat(clampFloat(tank.Nitrite+(0.06*bioLoad*filterFactor)-plantBonus*0.5, 0, 5), 2)
		clarity := clampInt(tank.Clarity-clarityDrop-turtleCount+plantCount+decorClarity, 0, 100)

		if _, err := db.Exec("UPDATE tanks SET ammonia = ?, nitrite = ?, clarity = ? WHERE id = ?", ammonia, nitrite, clarity, tank.ID); err != nil {
			return err
		}
		// 记录水质历史,前端 sparkline 用;只保留最近 14 天
		var curDay int
		if err := db.QueryRow("SELECT day FROM players WHERE id = ?", playerID).Scan(&curDay); err == nil {
			db.Exec(`INSERT OR REPLACE INTO water_history (tank_id, day, ammonia, nitrite, clarity) VALUES (?, ?, ?, ?, ?)`,
				tank.ID, curDay, ammonia, nitrite, clarity)
			db.Exec(`DELETE FROM water_history WHERE tank_id = ? AND day < ?`, tank.ID, curDay-14)
		}
	}
	return nil
}

// loadWaterHistory 读取该龟缸最近 N 天水质历史
func loadWaterHistory(tankID string, lastN int) []map[string]interface{} {
	rows, err := db.Query(`SELECT day, ammonia, nitrite, clarity FROM water_history WHERE tank_id = ? ORDER BY day DESC LIMIT ?`, tankID, lastN)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var day, clarity int
		var ammonia, nitrite float64
		if err := rows.Scan(&day, &ammonia, &nitrite, &clarity); err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"day": day, "ammonia": ammonia, "nitrite": nitrite, "clarity": clarity,
		})
	}
	// 反转成时间升序,前端方便画
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list
}

func advancePlayerTurtles(playerID string) error {
	type turtleDecay struct {
		ID      string
		TankID  string
		Species string
		Hunger  int
		Ammonia float64
		Nitrite float64
		Clarity int
	}

	rows, err := db.Query(`SELECT t.id, t.tank_id, t.species, t.hunger, COALESCE(tk.ammonia, 0), COALESCE(tk.nitrite, 0), COALESCE(tk.clarity, 100)
		FROM turtles t LEFT JOIN tanks tk ON t.tank_id = tk.id
		WHERE t.player_id = ?`, playerID)
	if err != nil {
		return err
	}

	var turtles []turtleDecay
	for rows.Next() {
		var t turtleDecay
		if err := rows.Scan(&t.ID, &t.TankID, &t.Species, &t.Hunger, &t.Ammonia, &t.Nitrite, &t.Clarity); err != nil {
			rows.Close()
			return err
		}
		turtles = append(turtles, t)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	// 缓存每个缸的 decor 汇总,避免同一个缸多次查
	tankSummary := map[string]struct {
		hasBasking bool
		hasShelter bool
	}{}
	for _, turtle := range turtles {
		if _, ok := tankSummary[turtle.TankID]; ok {
			continue
		}
		_, _, basking, shelter := summarizeDecorEffects(turtle.TankID)
		tankSummary[turtle.TankID] = struct {
			hasBasking bool
			hasShelter bool
		}{basking, shelter}
	}

	for _, turtle := range turtles {
		moodDrop := 3
		cleanDrop := 5
		healthDrop := 0
		if turtle.Ammonia >= 0.5 || turtle.Nitrite >= 0.3 || turtle.Clarity < 70 {
			moodDrop += 3
			cleanDrop += 4
		}
		if turtle.Ammonia >= 1.0 || turtle.Nitrite >= 0.8 || turtle.Clarity < 45 {
			healthDrop = 2
		}

		// 饥饿衰减:连续饥饿(hunger 已为 0)每天额外损耗活力/食欲
		starveDrop := 0
		if turtle.Hunger <= 0 {
			starveDrop = 2
		}

		// M3 布景效果:适合中/陕水位龟种的晒台反馈
		summary := tankSummary[turtle.TankID]
		sp, hasSp := findSpecies(turtle.Species)
		if hasSp && summary.hasBasking && (sp.HabitatType == "middle" || sp.HabitatType == "land") {
			moodDrop = maxInt(0, moodDrop-2)
			cleanDrop = maxInt(0, cleanDrop-1)
		}
		if summary.hasShelter {
			moodDrop = maxInt(0, moodDrop-1)
		}

		// 长期饥饿还会损伤皮肤与壳况，让"饿"的后果更立体。
		starveSkinShell := 0
		if turtle.Hunger <= 0 {
			starveSkinShell = 1
		}

		_, err := db.Exec(`UPDATE turtles SET
			hunger = MAX(0, hunger - 10),
			cleanliness = MAX(0, cleanliness - ?),
			mood = MAX(0, mood - ?),
			vitality = MAX(0, vitality - ? - ?),
			appetite = MAX(0, appetite - ? - ?),
			skin = MAX(0, skin - ? - ?),
			shell = MAX(0, shell - ? - ?)
			WHERE id = ? AND player_id = ?`, cleanDrop, moodDrop, healthDrop, starveDrop, healthDrop, starveDrop, healthDrop, starveSkinShell, healthDrop, starveSkinShell, turtle.ID, playerID)
		if err != nil {
			return err
		}
	}
	// AI_TESTER P2: vitality<40 触发 sick 状态，否则恢复 healthy
	_, err2 := db.Exec(`UPDATE turtles SET status = CASE WHEN vitality < 40 THEN 'sick' ELSE 'healthy' END WHERE player_id = ?`, playerID)
	if err2 != nil {
		return err2
	}
	return nil
}

// advanceBreeding 繁殖系统(M5):
// 1. 检查同缸异性成熟龟(出生≥20天),亲密度均≥40,按概率产蛋
// 2. 推进已有蛋的孵化倒计时,到期则孵化为新龟并删除蛋记录
// 3. 季节加成:春季产蛋概率 ×1.5
// 返回 (newEggsLaid, newTurtlesHatched, messages)
func advanceBreeding(playerID string, day int, season string) (int, int, []string) {
	var msgs []string
	newEggs := 0
	newHatches := 0

	// ── 1. 产蛋检测 ──
	rows, err := db.Query(`SELECT id, species, name, tank_id, gender, birth_day, intimacy, mood
		FROM turtles WHERE player_id = ?`, playerID)
	if err != nil {
		return 0, 0, nil
	}
	type turtleInfo struct {
		ID, Species, Name, TankID, Gender string
		BirthDay, Intimacy, Mood          int
	}
	var all []turtleInfo
	for rows.Next() {
		var t turtleInfo
		rows.Scan(&t.ID, &t.Species, &t.Name, &t.TankID, &t.Gender, &t.BirthDay, &t.Intimacy, &t.Mood)
		all = append(all, t)
	}
	rows.Close()

	tankTurtles := map[string][]turtleInfo{}
	for _, t := range all {
		if t.BirthDay <= day-20 { // 成年龟
			tankTurtles[t.TankID] = append(tankTurtles[t.TankID], t)
		}
	}

	for tankID, ts := range tankTurtles {
		var males, females []turtleInfo
		for _, t := range ts {
			if t.Gender == "♂" {
				males = append(males, t)
			} else if t.Gender == "♀" {
				females = append(females, t)
			}
		}
		if len(males) == 0 || len(females) == 0 {
			continue
		}

		var mom, dad turtleInfo
		found := false
		for _, f := range females {
			for _, m := range males {
				if f.Intimacy >= 40 && m.Intimacy >= 40 {
					mom, dad = f, m
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			continue
		}

		// 产蛋概率:基础 12%,亲密度加成,春季 ×1.5
		prob := 0.12
		avgIntimacy := (mom.Intimacy + dad.Intimacy) / 2
		prob += float64(avgIntimacy-40) * 0.003
		if season == "spring" {
			prob *= 1.5
		}
		avgMood := (mom.Mood + dad.Mood) / 2
		prob += float64(avgMood) * 0.0005

		if mrand.Float64() > prob {
			continue
		}

		// 子代种类 = 母方种类(简化遗传)
		eggID := fmt.Sprintf("egg_%d_%d", time.Now().UnixNano(), mrand.Intn(1000))
		hatchAfter := 5 + mrand.Intn(4) // 5-8 天
		hatchDay := day + hatchAfter
		quality := 40 + mrand.Intn(41) // 40-80
		db.Exec(`INSERT INTO eggs (id, player_id, species, tank_id, laid_day, hatch_day,
			parent_mom_id, parent_dad_id, quality)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			eggID, playerID, mom.Species, tankID, day, hatchDay, mom.ID, dad.ID, quality)

		sp, _ := findSpecies(mom.Species)
		spName := mom.Species
		if sp.ID != "" {
			spName = sp.Name
		}
		msgs = append(msgs, fmt.Sprintf("🥚 %s 和 %s 产下了一枚%s蛋!预计 %d 天后孵化。",
			mom.Name, dad.Name, spName, hatchAfter))
		newEggs++
	}

	// ── 2. 孵化检测 ──
	eggRows, err := db.Query(`SELECT id, species, tank_id, laid_day, hatch_day,
		COALESCE(parent_mom_id,''), COALESCE(parent_dad_id,''), quality
		FROM eggs WHERE player_id = ? AND hatch_day <= ?`, playerID, day)
	if err != nil {
		return newEggs, newHatches, msgs
	}
	type eggInfo struct {
		ID, Species, TankID, MomID, DadID string
		LaidDay, HatchDay, Quality        int
	}
	var ready []eggInfo
	for eggRows.Next() {
		var e eggInfo
		eggRows.Scan(&e.ID, &e.Species, &e.TankID, &e.LaidDay, &e.HatchDay,
			&e.MomID, &e.DadID, &e.Quality)
		ready = append(ready, e)
	}
	eggRows.Close()

	for _, e := range ready {
		hatchProb := float64(e.Quality) / 100.0
		if season == "summer" {
			hatchProb += 0.1
		}
		if hatchProb > 0.95 {
			hatchProb = 0.95
		}

		db.Exec("DELETE FROM eggs WHERE id = ?", e.ID)

		if mrand.Float64() > hatchProb {
			msgs = append(msgs, "💔 一枚蛋没有孵化成功...")
			continue
		}

		tID := generateTurtleID()
		sp, _ := findSpecies(e.Species)
		name := "小宝宝"
		if sp.ID != "" {
			name = defaultTurtleName(sp.ID)
		}
		gender := "♂"
		if mrand.Float64() < 0.5 {
			gender = "♀"
		}
		personalities := []string{"胆小", "活泼", "凶猛", "慵懒", "吃货"}
		personality := personalities[mrand.Intn(len(personalities))]

		initVitality := clampInt(e.Quality, 40, 100)
		initAppetite := clampInt(e.Quality+5, 40, 100)
		initSkin := clampInt(e.Quality-5, 40, 100)
		initShell := clampInt(e.Quality, 40, 100)

		db.Exec(`INSERT INTO turtles (id, player_id, species, name, gender, birth_day, weight,
			personality, vitality, appetite, skin, shell, intimacy, melanism, tank_id,
			hunger, cleanliness, mood)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			tID, playerID, e.Species, name, gender, day, 5.0+float64(mrand.Intn(5)),
			personality, initVitality, initAppetite, initSkin, initShell,
			10, 0, e.TankID, 70, 90, 80)

		// 保险解锁图鉴:该龟种如果之前未解锁(理论上不会,但万一),补上解锁记录
		unlockSpeciesForPlayer(playerID, e.Species, day)

		spName := e.Species
		if sp.ID != "" {
			spName = sp.Name
		}
		msgs = append(msgs, fmt.Sprintf("🐣 一只%s破壳而出!是%s的%s,性格%s。",
			spName, gender, name, personality))
		newHatches++
	}

	return newEggs, newHatches, msgs
}

func roundFloat(v float64, prec int) float64 {
	p := math.Pow(10, float64(prec))
	return math.Round(v*p) / p
}

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func getSeason(day int) string {
	seasons := []string{"spring", "summer", "autumn", "winter"}
	return seasons[(day/30)%4]
}

func speciesCatalog() []SpeciesInfo {
	return []SpeciesInfo{
		{ID: "muskTurtle", Name: "麝香龟", Category: "蛋龟", Difficulty: 1, Description: "体小、皮实、爱钻洞", UnlockCost: 0, HabitatType: "deep", UnlockCondition: "初始赠送",
			ScientificName: "Sternotherus odoratus", NativeRegion: "北美东部", AdultSize: "10-13cm",
			Trivia: "受惊会从腋窝腺体喷出麝香味液体,故名「臭味龟」(Stinkpot)。深水蛋龟一族,幼体在水下爬行而非游泳。"},
		{ID: "razorbackTurtle", Name: "剃刀龟", Category: "蛋龟", Difficulty: 2, Description: "头大壳高、性格凶", UnlockCost: 500, HabitatType: "deep", UnlockCondition: "商店购买",
			ScientificName: "Sternotherus carinatus", NativeRegion: "美国南部", AdultSize: "10-15cm",
			Trivia: "背甲中央有明显高耸的脊棱,像剃刀刀刃。咬合力惊人,混养需谨慎。"},
		{ID: "loggerheadMuskTurtle", Name: "果核泥龟", Category: "蛋龟", Difficulty: 2, Description: "黄色三道纹、温顺", UnlockCost: 600, HabitatType: "deep", UnlockCondition: "商店购买",
			ScientificName: "Sternotherus minor", NativeRegion: "美国东南部", AdultSize: "7-12cm",
			Trivia: "头部宽厚像「果核」,下颌有须状突起。幼体头侧三道金黄条纹是最大辨识特征。"},
		{ID: "chinesePondTurtle", Name: "中华草龟", Category: "水龟", Difficulty: 1, Description: "国民龟、墨化老头乐", UnlockCost: 0, HabitatType: "middle", UnlockCondition: "初始赠送",
			ScientificName: "Mauremys reevesii", NativeRegion: "中国/朝鲜半岛/日本", AdultSize: "15-20cm",
			Trivia: "成年公龟会全身墨化(黑化),俗称「墨龟/老头乐」。中国传统文化里长寿吉祥的代表。"},
		{ID: "yellowPondTurtle", Name: "黄喉拟水龟", Category: "水龟", Difficulty: 2, Description: "大青/小青、活泼亲人", UnlockCost: 800, HabitatType: "middle", UnlockCondition: "商店购买",
			ScientificName: "Mauremys mutica", NativeRegion: "中国南部/越南/日本", AdultSize: "15-22cm",
			Trivia: "颈部黄色,两侧有黄白纵纹。按产地分大青(广东)和小青(越南)两个品系,宠物市场最热门。"},
		{ID: "chineseStripeTurtle", Name: "中华花龟", Category: "水龟", Difficulty: 2, Description: "颈纹密布、群居热闹", UnlockCost: 1000, HabitatType: "middle", UnlockCondition: "商店购买",
			ScientificName: "Mauremys sinensis", NativeRegion: "中国南部/越南/台湾", AdultSize: "20-25cm",
			Trivia: "头颈布满细密黄绿色纵纹,故又名「珍珠龟」。群居性强,多只混养更活跃。"},
		{ID: "redEaredSlider", Name: "巴西龟", Category: "水龟", Difficulty: 1, Description: "入侵物种警示,请勿野放", UnlockCost: 300, HabitatType: "middle", UnlockCondition: "商店购买(带科普)",
			ScientificName: "Trachemys scripta elegans", NativeRegion: "美国密西西比河流域", AdultSize: "20-30cm",
			Trivia: "⚠️ 全球百大入侵物种。在中国南方野外繁殖已严重威胁本土水龟。养就养一辈子,绝不可野放。"},
		{ID: "yellowMarginTurtle", Name: "黄缘闭壳龟", Category: "半水龟", Difficulty: 4, Description: "国宝级、能闭壳、贵", UnlockCost: 2000, HabitatType: "land", UnlockCondition: "成就解锁",
			ScientificName: "Cuora flavomarginata", NativeRegion: "中国大陆/台湾", AdultSize: "15-19cm",
			Trivia: "背甲与腹甲之间有铰链,受惊可完全闭壳保护头尾。CITES 附录II保护物种,繁殖代售合法但请认准来源。"},
	}
}

func findSpecies(speciesID string) (SpeciesInfo, bool) {
	for _, s := range speciesCatalog() {
		if s.ID == speciesID {
			return s, true
		}
	}
	return SpeciesInfo{}, false
}

func handleGetSpecies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(speciesCatalog())
}

// handleGetPokedex 返回玩家维度的完整图鉴:所有龟种 + 是否解锁 + 解锁日 + 拥有数 + 孵化数
// 未解锁仍会返回基本信息(名字/分类/难度/条件),但不返回 trivia、学名、产地、体型
// 让玩家有「发现」的期待感。
func handleGetPokedex(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		playerID = "default"
	}
	if _, err := getOrCreatePlayer(playerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	unlocked := map[string]int{}
	rows, err := db.Query("SELECT species_id, COALESCE(unlock_day,0) FROM unlocked_species WHERE player_id = ?", playerID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sid string
			var day int
			if err := rows.Scan(&sid, &day); err == nil {
				unlocked[sid] = day
			}
		}
	}

	owned := map[string]int{}
	hatched := map[string]int{}
	trows, err := db.Query("SELECT species, birth_day FROM turtles WHERE player_id = ?", playerID)
	if err == nil {
		defer trows.Close()
		for trows.Next() {
			var sp string
			var bd int
			if err := trows.Scan(&sp, &bd); err == nil {
				owned[sp]++
				if bd > 1 {
					hatched[sp]++
				}
			}
		}
	}

	entries := []PokedexEntry{}
	for _, sp := range speciesCatalog() {
		day, ok := unlocked[sp.ID]
		entry := PokedexEntry{
			Species:      sp,
			Unlocked:     ok,
			UnlockDay:    day,
			OwnedCount:   owned[sp.ID],
			HatchedCount: hatched[sp.ID],
		}
		if !ok {
			// 未解锁：隐藏 trivia / 学名 / 产地 / 体型，保留诱惑性
			entry.Species.Trivia = ""
			entry.Species.ScientificName = ""
			entry.Species.NativeRegion = ""
			entry.Species.AdultSize = ""
			entry.Species.Description = "这是一种神秘的龟类…继续探索以解锁详情。"
		}
		entries = append(entries, entry)
	}

	totalUnlocked := 0
	for _, e := range entries {
		if e.Unlocked {
			totalUnlocked++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries":         entries,
		"total":           len(entries),
		"unlocked":        totalUnlocked,
		"completion_pct":  int(float64(totalUnlocked) / float64(len(entries)) * 100),
	})
}

// handleGetDecorCatalog 返回 M3 布景白名单,前端不再写死。
func handleGetDecorCatalog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decorCatalog())
}

// ShopItemSpec 是商店里能买到的消耗品/工具。
// 食物缺货是当前最痛的循环 gap:饥饿值会涨但买不到食物。这里补上。
type ShopItemSpec struct {
	ID       string `json:"id"`       // 同时也是 inventory 表里的 id
	Type     string `json:"type"`     // food / tool
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Desc     string `json:"desc"`
	Cost     int    `json:"cost"`     // 单次购买价格(每次买 PackSize 个)
	PackSize int    `json:"pack_size"`// 每次购买进背包多少个
}

// shopCatalog 全部走代码定义,跟 initInventory 的 id 复用,方便堆叠。
func shopCatalog() []ShopItemSpec {
	return []ShopItemSpec{
		{ID: "food_1", Type: "food", Name: "龟粮", Icon: "🍖", Desc: "日常主粮,浮性颗粒。", Cost: 12, PackSize: 10},
		{ID: "food_2", Type: "food", Name: "红虫", Icon: "🪱", Desc: "高蛋白零食,亲密度 +3。", Cost: 20, PackSize: 5},
		{ID: "food_3", Type: "food", Name: "虾干", Icon: "🦐", Desc: "硬质零食,磨喙也磨爪。", Cost: 30, PackSize: 5},
		{ID: "food_4", Type: "food", Name: "小鱼苗", Icon: "🐟", Desc: "野性十足,半水龟最爱。", Cost: 45, PackSize: 4},
		{ID: "tool_1", Type: "tool", Name: "水质测试剂", Icon: "🧪", Desc: "立刻刷新水质显示。", Cost: 25, PackSize: 3},
		{ID: "tool_2", Type: "tool", Name: "维生素片", Icon: "💊", Desc: "活力 +20(喂食时随机生效)。", Cost: 35, PackSize: 2},
	}
}

func findShopItem(id string) (ShopItemSpec, bool) {
	for _, s := range shopCatalog() {
		if s.ID == id {
			return s, true
		}
	}
	return ShopItemSpec{}, false
}

// handleGetShopCatalog 返回商店白名单 + 该玩家当前龟币,方便前端一次取齐。
func handleGetShopCatalog(w http.ResponseWriter, r *http.Request) {
	pid := r.URL.Query().Get("player_id")
	if pid == "" {
		pid = "default"
	}
	var coins int
	db.QueryRow("SELECT coins FROM players WHERE id = ?", pid).Scan(&coins)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":   shopCatalog(),
		"species": speciesCatalog(),
		"coins":   coins,
	})
}

// handleBuyItem 用龟币买消耗品。已存在则 count += pack_size,否则插入新行。
// 一次只买一包;前端循环调用更直观。
func handleBuyItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string `json:"player_id"`
		ItemID   string `json:"item_id"`
		Quantity int    `json:"quantity"` // 买几包,默认 1
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	if req.Quantity > 20 {
		req.Quantity = 20
	}
	spec, ok := findShopItem(req.ItemID)
	if !ok {
		http.Error(w, "unknown item: "+req.ItemID, http.StatusBadRequest)
		return
	}
	totalCost := spec.Cost * req.Quantity
	totalGain := spec.PackSize * req.Quantity

	var coins int
	if err := db.QueryRow("SELECT coins FROM players WHERE id = ?", req.PlayerID).Scan(&coins); err != nil {
		http.Error(w, "player not found", http.StatusBadRequest)
		return
	}
	if coins < totalCost {
		http.Error(w, fmt.Sprintf("龟币不足,需要 %d,当前 %d", totalCost, coins), http.StatusBadRequest)
		return
	}

	// 扣钱
	if _, err := db.Exec("UPDATE players SET coins = coins - ? WHERE id = ?", totalCost, req.PlayerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 入库:先查有没有同 id,有就 +count,没有就 insert
	var existing int
	err := db.QueryRow("SELECT count FROM inventory WHERE id = ? AND player_id = ?", spec.ID, req.PlayerID).Scan(&existing)
	if err == sql.ErrNoRows {
		db.Exec("INSERT INTO inventory (id, player_id, item_type, name, count, icon) VALUES (?, ?, ?, ?, ?, ?)",
			spec.ID, req.PlayerID, spec.Type, spec.Name, totalGain, spec.Icon)
	} else if err == nil {
		db.Exec("UPDATE inventory SET count = count + ? WHERE id = ? AND player_id = ?", totalGain, spec.ID, req.PlayerID)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 成就:破产边缘(首次商店购物)
	db.Exec(`UPDATE achievements SET unlocked = 1, unlock_day = (SELECT day FROM players WHERE id = ?)
		WHERE player_id = ? AND id = 'ach_5' AND unlocked = 0`, req.PlayerID, req.PlayerID)

	var coinsAfter int
	db.QueryRow("SELECT coins FROM players WHERE id = ?", req.PlayerID).Scan(&coinsAfter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"item_id":   spec.ID,
		"gained":    totalGain,
		"cost":      totalCost,
		"coins":     coinsAfter,
	})
}

func handleBuySpecies(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID  string `json:"player_id"`
		SpeciesID string `json:"species_id"`
		Name      string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}
	sp, ok := findSpecies(req.SpeciesID)
	if !ok {
		http.Error(w, "unknown species", http.StatusBadRequest)
		return
	}
	if sp.UnlockCost == 0 {
		http.Error(w, "初始龟种已赠送,无需购买", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = defaultTurtleName(sp.ID)
	}

	// 检查金币
	var coins int
	if err := db.QueryRow("SELECT coins FROM players WHERE id = ?", req.PlayerID).Scan(&coins); err != nil {
		http.Error(w, "player not found", http.StatusBadRequest)
		return
	}
	if coins < sp.UnlockCost {
		http.Error(w, "金币不足", http.StatusBadRequest)
		return
	}

	tankID, err := findOrCreateTankForSpecies(req.PlayerID, sp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 扣除金币并解锁龟种
	if _, err := db.Exec("UPDATE players SET coins = coins - ? WHERE id = ?", sp.UnlockCost, req.PlayerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var curDay int
	db.QueryRow("SELECT day FROM players WHERE id = ?", req.PlayerID).Scan(&curDay)
	unlockSpeciesForPlayer(req.PlayerID, req.SpeciesID, curDay)

	// 创建新乌龟并自动入住匹配水位的龟缸,避免买完龟出现在"无家可归"状态。
	turtleID := generateTurtleID()
	gender := "♀"
	if mrand.Intn(2) == 0 {
		gender = "♂"
	}
	if _, err := db.Exec(`INSERT INTO turtles (id, player_id, species, name, gender, personality,
		vitality, appetite, skin, shell, intimacy, melanism, tank_id, hunger, cleanliness, mood)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		turtleID, req.PlayerID, req.SpeciesID, req.Name, gender, "好奇",
		100, 100, 100, 100, 0, 0, tankID, 70, 90, 80); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "turtle_id": turtleID, "tank_id": tankID})
}

func handleCreateTank(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID   string `json:"player_id"`
		Name       string `json:"name"`
		WaterLevel string `json:"water_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}
	if !isValidWaterLevel(req.WaterLevel) {
		http.Error(w, "unknown water_level", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = defaultTankName(req.WaterLevel)
	}

	const cost = 120
	var coins int
	if err := db.QueryRow("SELECT coins FROM players WHERE id = ?", req.PlayerID).Scan(&coins); err != nil {
		http.Error(w, "player not found", http.StatusBadRequest)
		return
	}
	if coins < cost {
		http.Error(w, fmt.Sprintf("龟币不足,还差 %d", cost-coins), http.StatusBadRequest)
		return
	}

	tankID := fmt.Sprintf("tank_%d", time.Now().UnixNano())
	hasUVB := req.WaterLevel == "middle" || req.WaterLevel == "land"
	hasFilter := req.WaterLevel != "land"
	tempDay := 26.0
	tempNight := 24.0
	if req.WaterLevel == "land" {
		tempDay = 27.5
		tempNight = 23.0
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	if _, err := tx.Exec("UPDATE players SET coins = coins - ? WHERE id = ?", cost, req.PlayerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := tx.Exec(`INSERT INTO tanks (id, player_id, type, name, water_level, temp_day, temp_night,
		has_uvb, has_filter, ph, ammonia, nitrite, clarity)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tankID, req.PlayerID, "square", req.Name, req.WaterLevel, tempDay, tempNight,
		boolToInt(hasUVB), boolToInt(hasFilter), 7.0, 0.0, 0.0, 100); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "tank_id": tankID, "cost": cost})
}

func handleMoveTurtle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID string `json:"player_id"`
		TurtleID string `json:"turtle_id"`
		TankID   string `json:"tank_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PlayerID == "" {
		req.PlayerID = "default"
	}
	if req.TurtleID == "" || req.TankID == "" {
		http.Error(w, "turtle_id and tank_id required", http.StatusBadRequest)
		return
	}

	var species, oldTankID string
	if err := db.QueryRow("SELECT species, tank_id FROM turtles WHERE id = ? AND player_id = ?", req.TurtleID, req.PlayerID).Scan(&species, &oldTankID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "turtle not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var waterLevel string
	if err := db.QueryRow("SELECT water_level FROM tanks WHERE id = ? AND player_id = ?", req.TankID, req.PlayerID).Scan(&waterLevel); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "tank not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sp, ok := findSpecies(species); ok && sp.HabitatType != waterLevel {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("%s更适合%s,不能搬到%s", sp.Name, waterLevelName(sp.HabitatType), waterLevelName(waterLevel)),
		})
		return
	}
	if oldTankID == req.TankID {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "已经在这个龟缸里"})
		return
	}

	if _, err := db.Exec("UPDATE turtles SET tank_id = ?, mood = MAX(0, mood - 2) WHERE id = ? AND player_id = ?", req.TankID, req.TurtleID, req.PlayerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "搬家完成", "tank_id": req.TankID})
}

func isValidWaterLevel(level string) bool {
	switch level {
	case "deep", "middle", "shallow", "land":
		return true
	default:
		return false
	}
}

func defaultTankName(level string) string {
	switch level {
	case "deep":
		return "新的深水缸"
	case "middle":
		return "新的中水位缸"
	case "shallow":
		return "新的浅水缸"
	case "land":
		return "新的半水陆缸"
	default:
		return "新的龟缸"
	}
}

func waterLevelName(level string) string {
	switch level {
	case "deep":
		return "深水缸"
	case "middle":
		return "中水位缸"
	case "shallow":
		return "浅水缸"
	case "land":
		return "半水陆缸"
	default:
		return level
	}
}

func defaultTurtleName(speciesID string) string {
	switch speciesID {
	case "razorbackTurtle":
		return "小剃刀"
	case "yellowPondTurtle":
		return "小黄喉"
	case "yellowMarginTurtle":
		return "小黄缘"
	default:
		return "新朋友"
	}
}

func findOrCreateTankForSpecies(playerID string, sp SpeciesInfo) (string, error) {
	var tankID string
	err := db.QueryRow("SELECT id FROM tanks WHERE player_id = ? AND water_level = ? ORDER BY id LIMIT 1", playerID, sp.HabitatType).Scan(&tankID)
	if err == nil {
		return tankID, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	tankID = fmt.Sprintf("tank_%d", time.Now().UnixNano())
	tankName := sp.Name + "的新家"
	hasUVB := sp.HabitatType == "middle" || sp.HabitatType == "land"
	hasFilter := sp.HabitatType != "land"
	if _, err := db.Exec(`INSERT INTO tanks (id, player_id, type, name, water_level, temp_day, temp_night,
		has_uvb, has_filter, ph, ammonia, nitrite, clarity)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tankID, playerID, "square", tankName, sp.HabitatType, 26.0, 24.0,
		boolToInt(hasUVB), boolToInt(hasFilter), 7.0, 0.0, 0.0, 100); err != nil {
		return "", err
	}
	return tankID, nil
}

// handleTurtleDetail GET /api/turtle?id=xxx&player_id=default
// 返回某只龟的详细信息:基础属性 + 健康四维 + 龟种习性 + 当前缸水质 + 14日水质历史 + 智能建议。
func handleTurtleDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	playerID := r.URL.Query().Get("player_id")
	if playerID == "" {
		playerID = "default"
	}

	var t Turtle
	err := db.QueryRow(`SELECT id, species, name, gender, birth_day, weight, personality,
		vitality, appetite, skin, shell, intimacy, melanism,
		COALESCE(tank_id, ''), hunger, cleanliness, mood
		FROM turtles WHERE id = ? AND player_id = ?`, id, playerID).Scan(
		&t.ID, &t.Species, &t.Name, &t.Gender, &t.BirthDay, &t.Weight, &t.Personality,
		&t.Health.Vitality, &t.Health.Appetite, &t.Health.Skin, &t.Health.Shell,
		&t.Intimacy, &t.Melanism, &t.TankID, &t.Hunger, &t.Cleanliness, &t.Mood,
	)
	if err != nil {
		http.Error(w, "turtle not found", http.StatusNotFound)
		return
	}

	var curDay int
	db.QueryRow("SELECT day FROM players WHERE id = ?", playerID).Scan(&curDay)
	ageDays := curDay - t.BirthDay
	if ageDays < 0 {
		ageDays = 0
	}

	species, _ := findSpecies(t.Species)

	// 当前缸 + 历史
	var tank map[string]interface{}
	var waterHistory []map[string]interface{}
	if t.TankID != "" {
		var tk Tank
		var hasFilterI, hasUVBI int
		err := db.QueryRow(`SELECT id, name, water_level, has_filter, has_uvb,
			ph, ammonia, nitrite, clarity, temp_day, temp_night
			FROM tanks WHERE id = ?`, t.TankID).Scan(
			&tk.ID, &tk.Name, &tk.WaterLevel, &hasFilterI, &hasUVBI,
			&tk.WaterQual.PH, &tk.WaterQual.Ammonia, &tk.WaterQual.Nitrite, &tk.WaterQual.Clarity,
			&tk.TempDay, &tk.TempNight,
		)
		if err == nil {
			tk.HasFilter = hasFilterI == 1
			tk.HasUVB = hasUVBI == 1
			tank = map[string]interface{}{
				"id":           tk.ID,
				"name":         tk.Name,
				"water_level":  tk.WaterLevel,
				"water_name":   waterLevelName(tk.WaterLevel),
				"has_filter":   tk.HasFilter,
				"has_uvb":      tk.HasUVB,
				"ph":           tk.WaterQual.PH,
				"ammonia":      tk.WaterQual.Ammonia,
				"nitrite":      tk.WaterQual.Nitrite,
				"clarity":      tk.WaterQual.Clarity,
				"temp_day":     tk.TempDay,
				"temp_night":   tk.TempNight,
			}
			waterHistory = loadWaterHistory(t.TankID, 14)
		}
	}

	// 智能建议(toast/UI 高亮的依据)
	suggestions := buildTurtleSuggestions(t, species, tank)

	resp := map[string]interface{}{
		"turtle":         t,
		"age_days":       ageDays,
		"species_info":   species,
		"tank":           tank,
		"water_history":  waterHistory,
		"suggestions":    suggestions,
	}
	json.NewEncoder(w).Encode(resp)
}

// buildTurtleSuggestions 根据当前龟+缸状态生成 0~N 条操作建议。
// 用于详情面板顶部彩色提示 + 后续主界面按钮智能高亮。
func buildTurtleSuggestions(t Turtle, sp SpeciesInfo, tank map[string]interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	// AI_TESTER P2: sick 状态置顶提示
	if t.Status == "sick" {
		out = append(out, map[string]interface{}{"level": "danger", "icon": "🩺", "text": fmt.Sprintf("%s 生病了(vitality=%d)，建议隔离治疗、改善水质、补充营养", t.Name, t.Health.Vitality)})
	}
	if t.Hunger <= 30 {
		out = append(out, map[string]interface{}{"level": "warn", "icon": "🍖", "text": fmt.Sprintf("%s 已经很饿了(饥饿度 %d),建议立刻喂食", t.Name, t.Hunger)})
	} else if t.Hunger <= 55 {
		out = append(out, map[string]interface{}{"level": "info", "icon": "🥗", "text": fmt.Sprintf("%s 有些饿了,可以补一顿", t.Name)})
	}
	if t.Cleanliness <= 40 {
		out = append(out, map[string]interface{}{"level": "warn", "icon": "🛁", "text": "身上脏了,建议清洁或换水"})
	} else if t.Cleanliness <= 60 {
		out = append(out, map[string]interface{}{"level": "info", "icon": "🧼", "text": "清洁度一般,可顺手清理一下"})
	}
	if t.Mood <= 40 {
		out = append(out, map[string]interface{}{"level": "warn", "icon": "👋", "text": "心情低落,多互动可提升亲密度"})
	} else if t.Mood <= 55 {
		out = append(out, map[string]interface{}{"level": "info", "icon": "🎈", "text": "心情一般,陪它玩玩吧"})
	}
	if t.Health.Vitality <= 50 || t.Health.Appetite <= 50 {
		out = append(out, map[string]interface{}{"level": "warn", "icon": "🩺", "text": "活力/食欲偏低,注意水质和环境温度"})
	}
	if tank != nil {
		if a, ok := tank["ammonia"].(float64); ok && a >= 1.0 {
			out = append(out, map[string]interface{}{"level": "danger", "icon": "☠️", "text": fmt.Sprintf("NH3 已达 %.2f mg/L,强烈建议换水", a)})
		}
		if c, ok := tank["clarity"].(int); ok && c < 50 {
			out = append(out, map[string]interface{}{"level": "warn", "icon": "💧", "text": "水体浑浊,建议深度清洁或开过滤"})
		}
		if hf, ok := tank["has_filter"].(bool); ok && !hf {
			out = append(out, map[string]interface{}{"level": "info", "icon": "⚙️", "text": "当前缸未安装过滤器,可在维护菜单中安装"})
		}
		// 龟种 vs 缸水位匹配性
		if wl, ok := tank["water_level"].(string); ok && sp.ID != "" && wl != sp.HabitatType && sp.HabitatType != "" {
			out = append(out, map[string]interface{}{"level": "warn", "icon": "🏠", "text": fmt.Sprintf("%s 偏好「%s」缸,当前为「%s」,可考虑搬家", sp.Name, waterLevelName(sp.HabitatType), waterLevelName(wl))})
		}
	}
	if len(out) == 0 {
		out = append(out, map[string]interface{}{"level": "ok", "icon": "✅", "text": "一切安好,享受佛系养龟时光吧"})
	}
	return out
}

// handleBreedingHints 返回繁殖条件进度，让玩家知道每缸离产蛋还差多少。
func handleBreedingHints(w http.ResponseWriter, r *http.Request) {
	pid := r.URL.Query().Get("player_id")
	if pid == "" {
		pid = "default"
	}
	if _, err := getOrCreatePlayer(pid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var day int
	db.QueryRow("SELECT day FROM players WHERE id = ?", pid).Scan(&day)
	season := getSeason(day)

	rows, err := db.Query(`SELECT id, species, name, tank_id, gender, birth_day, intimacy, mood
		FROM turtles WHERE player_id = ? ORDER BY tank_id`, pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type ti struct {
		ID, Species, Name, TankID, Gender string
		BirthDay, Intimacy, Mood          int
	}
	byTank := map[string][]ti{}
	for rows.Next() {
		var t ti
		rows.Scan(&t.ID, &t.Species, &t.Name, &t.TankID, &t.Gender, &t.BirthDay, &t.Intimacy, &t.Mood)
		byTank[t.TankID] = append(byTank[t.TankID], t)
	}
	rows.Close()

	var hints []map[string]interface{}
	for tankID, ts := range byTank {
		var adults []ti
		for _, t := range ts {
			if t.BirthDay <= day-20 {
				adults = append(adults, t)
			}
		}
		if len(adults) < 2 {
			continue
		}
		var males, females []ti
		for _, t := range adults {
			if t.Gender == "♂" {
				males = append(males, t)
			} else if t.Gender == "♀" {
				females = append(females, t)
			}
		}
		if len(males) == 0 || len(females) == 0 {
			continue
		}
		// 找最佳配对（平均亲密度最高且均≥0）
		bestAvg := -1
		var bestM, bestF ti
		for _, m := range males {
			for _, f := range females {
				if m.Species != f.Species {
					continue // 简化：同种才能繁殖
				}
				avg := (m.Intimacy + f.Intimacy) / 2
				if avg > bestAvg {
					bestAvg = avg
					bestM, bestF = m, f
				}
			}
		}
		if bestM.ID == "" {
			continue
		}
		progressM := bestM.Intimacy
		progressF := bestF.Intimacy
		ready := progressM >= 40 && progressF >= 40
		seasonBonus := ""
		if season == "spring" {
			seasonBonus = "，春季产蛋概率×1.5"
		}
		msg := fmt.Sprintf("%s缸：%s♂(亲密度%d) + %s♀(亲密度%d) ", bestM.Name, bestM.Name, progressM, bestF.Name, progressF)
		if ready {
			msg += fmt.Sprintf("已满足繁殖条件%s", seasonBonus)
		} else {
			msg += fmt.Sprintf("繁殖需双方亲密度≥40%s", seasonBonus)
		}
		hints = append(hints, map[string]interface{}{
			"tank_id":        tankID,
			"ready":          ready,
			"male":           map[string]interface{}{"id": bestM.ID, "name": bestM.Name, "intimacy": progressM},
			"female":         map[string]interface{}{"id": bestF.ID, "name": bestF.Name, "intimacy": progressF},
			"species":        bestM.Species,
			"season_bonus":   season == "spring",
			"message":        msg,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"day":    day,
		"season": season,
		"hints":  hints,
	})
}
