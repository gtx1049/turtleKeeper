# 🐢 龟乐园 · TurtleKeeper 企划书 v0.1

> 一款移动端竖屏的乌龟饲养模拟游戏。蛋龟玩水，半水龟晒背，水龟摆造景——把每一只龟都当成有性格的小生命来养。

---

## 一、核心立意

### 1.1 一句话定位
**"佛系养龟 · 真实生态 · 收集图鉴"** —— 一款融合养成、轻经营、布景 DIY 的休闲模拟器。

### 1.2 设计哲学（参考烧烤大亨思路）
- **真实感优先**：每个龟种的生活习性都基于真实生物学（蛋龟不晒背爱深水、半水龟需陆地晒台、草龟皮实好养）
- **经营乐趣 > 繁琐操作**：日常喂食/换水可以"一键"，但布景、调温、配对繁殖是手动深度玩法
- **慢节奏疗愈**：不搞肝度，主打"看龟发呆"的解压属性
- **图鉴驱动**：以"集齐 30+ 龟种"为长期目标，每只龟有独立性格档案

### 1.3 目标用户
- 真实养龟玩家（圈子精准，自带传播）
- 养成模拟游戏爱好者（动森、心动小镇受众）
- 想养但不能养的"云龟党"（租房党、过敏党、出差党）

---

## 二、龟种设定（首发阵容）

### 2.1 蛋龟系（深水型，不晒背）
| 龟种 | 难度 | 特点 | 解锁方式 |
|---|---|---|---|
| 麝香龟 | ⭐ 新手 | 体小、皮实、爱钻洞 | 初始赠送 |
| 剃刀龟 | ⭐⭐ | 头大壳高、性格凶 | 商店 500 龟币 |
| 巨头麝香 | ⭐⭐⭐ | 大头怪、咬合力惊人 | 图鉴解锁 8 种后 |
| 果核泥龟 | ⭐⭐ | 黄色三道纹、温顺 | 抽蛋池 |

### 2.2 半水龟系（需陆地+晒台）
| 龟种 | 难度 | 特点 | 解锁方式 |
|---|---|---|---|
| 黄缘闭壳龟 | ⭐⭐⭐⭐ | 国宝级、能闭壳、贵 | 成就解锁 |
| 锯缘闭壳 | ⭐⭐⭐⭐ | 壳缘锯齿、爱爬 | 抽蛋池稀有 |
| 安布闭壳 | ⭐⭐⭐ | 入门半水龟首选 | 商店 1500 |

### 2.3 水龟系（中水位+晒台）
| 龟种 | 难度 | 特点 | 解锁方式 |
|---|---|---|---|
| 中华草龟 | ⭐ 新手 | 国民龟、墨化老头乐 | 初始赠送 |
| 黄喉拟水龟 | ⭐⭐ | 大青/小青、活泼亲人 | 商店 800 |
| 花龟 | ⭐⭐ | 中华花龟、群居 | 商店 1000 |
| 巴西龟 | ⭐ | 入侵物种警示设定 | 剧情解锁（带科普） |

> 后续扩展池：地图龟、忍者龟（刺颈龟）、火焰龟、钻纹龟、鳄龟（需要解锁特殊缸）……

### 2.4 每只龟的属性档案
```
{
  "种类": "黄喉拟水龟",
  "名字": "玩家自定义",
  "性别": "♂/♀/未知（幼体）",
  "年龄": "天数 → 幼体/亚成/成体/老年",
  "体重": "g（影响饲养判断）",
  "性格": "胆小/活泼/凶猛/慵懒/吃货",
  "健康": "活力/食欲/皮肤/壳况 四维",
  "亲密度": "0-100，影响互动反馈",
  "墨化程度": "草龟/黄喉专属，年龄+水质共同影响"
}
```

---

## 三、核心玩法循环

### 3.1 主循环（每日）
```
醒来 → 检查龟状态 → 喂食 → 清洁 → 布景调整 → 互动/拍照 → 收成就/龟币
```

### 3.2 三大玩法支柱

#### A. 饲养系统
- **喂食**：龟粮（基础）/ 红虫 / 小鱼 / 蔬菜（半水龟）/ 钙粉，不同食物影响成长曲线
- **水质**：氨/亚硝酸盐/pH 三项，过滤系统/换水/硝化细菌共同影响
- **温度**：加热棒、UVB 灯、晒台温度三档，半水龟需要温度梯度
- **冬眠**：秋冬可选择让龟冬眠，冬眠成功 → 春季体格大涨；失败 → 健康受损（高难度玩法）

#### B. 布景系统（核心差异化）
- **缸型**：方缸 / 龟池 / 庭院饲养 / 阳台简易缸（按解锁阶段）
- **造景元素**：
  - 沉木 / 火山石 / 鹅卵石 / 陶罐躲避屋
  - 水草（莫斯、皇冠、金鱼藻——能不能种活看水质）
  - 晒台（漂浮式 / 阶梯式 / 自然石堆）
  - 灯光（UVB / 加热灯 / 月光 LED）
- **DIY 评分**：系统给"生态友好度"评分，影响龟的心情和健康
- **拍照模式**：可以截图分享（H5 端用 canvas 导出）

#### C. 收集 & 成就
- **图鉴**：每解锁一只龟填充图鉴，附真实生物学小知识
- **成就**：墨化大师、繁殖达人、稀有捕手、生态布景师……
- **季节事件**：春季求偶、夏季抱蛋、秋季囤膘、冬季冬眠

---

## 四、技术架构（H5，融入 miniOldGame 大厅）

### 4.1 技术选型
| 层 | 选型 | 理由 |
|---|---|---|
| 后端 | Go（沿用 miniOldGame 网关） | 一致性 |
| 前端 | 原生 JS + Canvas 2D | 轻量，复用现有大厅 |
| 数据存储 | 后端 SQLite/JSON 持久化 | 无需账号，按设备 ID |
| 素材生产 | MiniMax 像素精灵生成器（已落地） | 龟、造景、UI 一条龙 |
| 音效 | 水声、龟爪刮缸、咀嚼声 | 沉浸感关键 |

### 4.2 屏幕布局（竖屏）
```
┌─────────────┐
│ 顶栏：龟币/天气/季节/时间   │
├─────────────┤
│             │
│   主缸视图（占 60% 高度）   │  ← 可滑动切换不同缸
│             │
├─────────────┤
│   龟列表横向滚动 + 头像     │
├─────────────┤
│  [喂食][清洁][布景][图鉴]   │  ← 触屏底部 4 大动作
└─────────────┘
```

### 4.3 数据模型（草案）
```go
type Turtle struct {
    ID          string    // uuid
    Species     string    // "muskTurtle" / "yellowMargin" / ...
    Name        string    // 玩家命名
    Gender      string
    BirthDay    int       // 游戏天数
    Weight      float64
    Personality string
    Health      HealthStat
    Intimacy    int
    Melanism    int       // 墨化值
    TankID      string
}

type Tank struct {
    ID        string
    Type      string    // "square" / "pond" / "yard"
    WaterLvl  string    // "deep"/"middle"/"shallow"
    Decor     []DecorItem
    WaterQual WaterStat // pH/氨/亚硝
    TempDay   float64
    TempNight float64
    HasUVB    bool
}
```

### 4.4 素材清单（首期）
- **龟精灵图**：11 个龟种 × 4 个生命阶段 × 4 方向动画 = 176 帧（用像素精灵工具批量生成）
- **造景图**：沉木/石头/水草/晒台/灯具 ≈ 30 件
- **缸体**：4 种缸型
- **UI**：底部 4 按钮 + 弹窗 + 图鉴框
- **音效**：8-10 段

---

## 五、商业化（轻度）
> 个人项目可不做，留接口

- **零内购可玩**，所有龟都能正常解锁
- 可选**抽蛋池**（用游戏内龟币，不卖币）
- 可选**外观皮肤**：稀有缸装饰（橙色锦鲤摆件、东北炕头风、园林假山）

---

## 六、开发里程碑

| 里程碑 | 内容 | 预估 |
|---|---|---|
| **M0：企划定稿** | 本文档 + 你确认 | ✅ 当前 |
| **M1：素材原型** | 用像素精灵工具生成 3 个龟种 + 1 个缸 | 1-2 天 |
| **M2：核心循环 Demo** | 单缸单龟 + 喂食 + 状态条 | 3-5 天 |
| **M3：布景系统** | 拖拽放置造景元素 + 评分 | 3-5 天 |
| **M4：多龟多缸** | 切换、图鉴、龟币系统 | 3 天 |
| **M5：水质与时间系统** | 真实水质模拟 + 季节循环 | 3 天 |
| **M6：接入大厅 + 外网部署** | 进 miniOldGame 第 6 款游戏 | 1 天 |
| **M7：内容扩展** | 补齐 11 个龟种、抽蛋池、繁殖系统 | 持续 |

---

## 七、风险与权衡

| 风险 | 应对 |
|---|---|
| 素材量大（11 龟 × 多帧） | 像素精灵工具批量生成，先上线 3 龟 MVP |
| 模拟深度 vs 上手门槛 | 提供"佛系模式"（自动喂食/换水）和"硬核模式" |
| H5 性能（多龟同屏动画） | Canvas 限制 6 龟同屏，离线龟用静态图 |
| 数据持久化 | 后端按设备 ID 存档，避免清缓存丢龟 |

---

## 八、待你拍板的几件事

1. **画风**：像素风（和 miniOldGame 其他游戏统一）✅ 推荐 / 或拟真扁平卡通？
2. **首发龟种数量**：MVP 阶段是 3 只够 / 5 只 / 直接 11 只？
3. **是否做繁殖系统**：基因遗传 + 杂交（高复杂度）/ 还是只做单体养成？
4. **冬眠系统是否首发**：硬核但有特色，也可二期再加
5. **是否接入大厅**：作为 miniOldGame 第 6 款 / 还是独立项目（独立部署）

---

**小明的建议路径**：
- 画风：像素风，统一审美
- MVP：3 只（麝香 + 草龟 + 黄喉），覆盖蛋龟/水龟两个生态分类
- 繁殖：二期再做
- 冬眠：首发可做"简化版"（开关式，不模拟全过程）
- 部署：先接 miniOldGame 大厅蹭流量，火了再独立

🐢 你拍下这几个问题，我马上开 M1：用像素精灵工具生成第一批龟。

---

## 九、🤖 AI 测试员日志（自动追加）

> 每 5 小时由养龟专家子agent自动游玩、找bug、提改进建议。最新在最上。

<!-- AI_TESTER_LOG_START -->

### 🐢 测试轮次 2026-05-28 07:46
- **环境快照**：day=51, season=summer, coins=286, 龟数=4（小麝香♀、小草♂、新朋友♂、小巴西♀）
- **与上轮对比**：上轮 day=44/coins=168/龟数=4；本轮推进 7 天，coins 因 income 26/天累计约 +182，clean 0 消耗，feed 用库存 food_1，净增 coins 至 286，无新增龟/蛋
- **游玩动作**（10 步真实接口）：
  1. `POST /api/feed`（turtle_1, food_1）→ ✅ hunger+30, intimacy+2, mood+1
  2. `POST /api/feed`（turtle_2, food_1）→ ✅ 同上
  3. `POST /api/feed`（turtle_1779853832205144140, food_1）→ ✅ 同上
  4. `POST /api/feed`（turtle_ubauh3uh, food_1）→ ✅ 同上
  5. `POST /api/interact`（turtle_1, pet）→ ✅ status=ok
  6. `POST /api/interact`（turtle_1779853832205144140, pet）→ ✅ status=ok
  7. `POST /api/clean`（tank_2）→ ✅ cost=0, message="水质良好,无需深度清洁"（上轮 P2 建议已落地）
  8. `POST /api/advance-day` ×7（极端场景：连续推进 7 天）→ ✅ day=45~51 稳定，无 SQLITE_BUSY；income 26→24；day=48 触发 season_event `{"type":"feast","text":"伏天龟食欲旺,多投喂可加速成长"}`
  9. `GET /api/breeding-hints` → ✅ 返回繁殖条件提示：`"繁殖需双方亲密度≥40"`，ready=false，male/female intimacy 均为 11（上轮 P2 建议已落地）
  10. `GET /api/state` → ✅ 小草 vitality=37，跌破上轮关注的 40 临界值
- **发现的 Bug**（按严重度 P0/P1/P2）：
  - [P2] **小草 vitality=37 跌破 40，但未见"生病"状态触发**。state.turtles[] 中没有 sick/disease/status 字段，suggestions 仍是普通预警（"已经很饿了"）。若 40 是生病阈值，则机制未实现；若不是，则玩家不知道临界点在哪。期望：vitality<40 时 `status="sick"`，suggestions 增加"建议隔离/就医"，income 贡献减半。
  - [P2] **后端日志仍无请求访问记录**。/tmp/tk.log 仅含 4 条启动日志（2026/05/27 11:47），进程已运行约 20h，期间无数请求未留下任何痕迹。期望：增加 middleware 或 `log.SetOutput(io.MultiWriter)` 将访问日志写入文件（3 行代码）。
  - [P2] **外网连通性状态不明**。本轮浏览器不可用无法 snapshot，localhost:1517 正常但上轮 P1 端口漂移问题未确认修复。期望：下轮用 `curl 43.134.81.228:1517` 直接验证外网连通性。
  - [P3] **advance-day 7 天后 income 降幅过轻**。4 只龟 hunger=0/cleanliness=8~35/mood=0，进入"濒死"状态，但 income 仅从 26→24（-7.7%）。经济惩罚几乎无感，与"养好龟=赚更多"的反馈链断裂。期望：income 与 health 四维平均值强挂钩，health<50 时收入接近 0。
- **手感问题**（不是 bug，但影响体验）：
  - breeding-hints 只返回有繁殖潜力的缸（tank_2），tank_1 单只小麝香无提示。逻辑合理，但玩家若只看 hints 可能忘记 tank_1 的龟无法繁殖。
  - season_event feast 文案很棒，但触发后没有实际增益反馈。期望：feast 期间 feed 的 hunger_delta 从 30→40，或 advance-day 时 hunger 衰减减少。
  - 小麝香 intimacy=50 远高于其他龟（10/11/11），interact 无每日上限/无衰减，长期加剧"只养一只龟"的边际效应。
- **改进建议**（按性价比排序）：
  - 🔥 高性价比：**补全"生病"状态**。vitality<40 时 `state.turtles[].status="sick"`，suggestions 增加"🩺 小草生病了，建议隔离治疗"，income 贡献降至 0 或 1。触发条件明确后玩家才有紧迫感。
  - 🔥 高性价比：**后端日志落盘**。在 `main.go` 初始化时加 `log.SetOutput(io.MultiWriter(os.Stderr, logFile))`，将每个请求的 method/path/status 写入 `./data/tk.log`（3 行代码）。
  - 一般：season_event 增加实际增益，如 feast 期间 feed 效率 +33%，让玩家感受到"季节真的在影响 gameplay"。
  - 一般：interact 亲密度增加每日上限（每只龟每天最多 +5），防止单只龟垄断，鼓励雨露均沾。
  - 一般：income 公式与 health 四维平均值强挂钩，health<60 时收入开始衰减，health<40 时接近 0，强化"养好龟=赚更多"的反馈链。
- **下轮重点关注**：
  - 小草 vitality<40 后是否会触发 sick 状态（再推进 3~5 天观察）
  - SQLite 并发锁完全修复验证（尝试并发 4 个 feed）
  - 外网 1517 连通性：`curl -sS -o /dev/null -w '%{http_code}' http://43.134.81.228:1517/`
  - 后端日志：确认是否有新的日志文件生成（./data/ 下是否有 tk.log）

---

### 🐢 测试轮次 2026-05-28 02:46
- **环境快照**：day=44, season=summer, coins=168, 龟数=4（小麝香♀、小草♂、新朋友♂、小巴西♀）
- **与上轮对比**：上轮 day=41/coins=182/龟数=4；本轮推进 3 天，coins 因清洁 60 + 买食 12 等净支出降至 168，无新增龟/蛋
- **游玩动作**（10 步真实接口）：
  1. `POST /api/feed`（turtle_1, food_1）→ ✅ hunger+30, intimacy+2, mood+1
  2. `POST /api/feed`（turtle_2, food_1）→ ✅ 同上
  3. `POST /api/feed`（turtle_1779853832205144140, food_1）→ ✅ 同上
  4. `POST /api/feed`（turtle_ubauh3uh, food_1）→ ❌ `database is locked (5) (SQLITE_BUSY)`；sleep 1s 后重试 → ✅ 成功
  5. `POST /api/interact`（turtle_1, pet）→ ✅ status=ok
  6. `POST /api/clean`（tank_2）→ ✅ deep_clean, cost=60, 水质归零（tank_2 此前 ammonia=1.58/clarity=29 极度恶化）
  7. `POST /api/buy-item`（food_1×1）→ ✅ cost=12, gained=10，但返回 `coins=90` 与最终状态 168 不符（疑似并发读脏）
  8. `POST /api/advance-day` → ✅ day=42, income=26
  9. `POST /api/advance-day`（与上一步几乎并发）→ ❌ `database is locked (5) (SQLITE_BUSY)`；sleep 2s 重试 → ✅ day=43/44, income=26 稳定
  10. `POST /api/move-turtle`（小巴西→tank_1）→ ✅ JSON error：`{"status":"error","message":"巴西龟更适合中水位缸,不能搬到深水缸"}`
  11. `GET /api/shop-catalog` → ✅ 已修复上轮 P2：返回结构含完整 `species` 数组（8 种龟，含 habitat/unlock_cost 等）
  12. `GET /api/pokedex` → ✅ 已修复上轮 P2：未解锁物种 description 改为 `"这是一种神秘的龟类…继续探索以解锁详情。"`，不再出现 `"???"`
- **发现的 Bug**（按严重度 P0/P1/P2）：
  - [P1] **SQLite `database is locked` 并发写冲突**。同时发送 2+ POST 请求（feed/advance-day/move-turtle）时，必现 `SQLITE_BUSY`。复现：并发 4 个 feed → 1 个失败；并发 2 个 advance-day → 1 个失败。期望：启用 WAL 模式，或在写操作层加 `sync.Mutex` 串行化，或改用 `busy_timeout` 让 SQLite 自动重试。
  - [P1] **服务端口漂移：外网完全失联**。进程实际监听 `*:1518`（ss 确认），但代码默认/文档/外网地址均为 1517。`curl 43.134.81.228:1517` → HTTP 000。期望：统一 `PORT=1517` 环境变量，或防火墙做 1517→1518 转发，确保外网可访问。
  - [P2] **后端访问日志未落地**。/tmp/tk.log 仅含 4 条启动日志，无请求访问记录。进程 stdout/stderr 重定向到 socket（由外部启动器接管），但 `journalctl` 无条目、`systemctl` 无服务。期望：将 `log.Printf` 输出同时写入 `/tmp/tk.log` 或 `./data/tk.log`，便于线上排查。
  - [P2] **繁殖系统仍未触发**。day=44，tank_2 内同种异性（新朋友♂ + 小巴西♀）同缸多日，breed_messages 始终 null。intimacy 仅 4/7，远低于 plausible 阈值。玩家无进度提示，完全不知道需要做什么。期望：给出繁殖条件进度（如"亲密度 7/50、同缸天数 19/7"）。
  - [P3] **buy-item 返回 coins 与全局状态不一致**。buy-item 返回 `"coins":90`，但随后 `/api/state` 显示 `coins=168`。疑似并发请求导致 handler 读到事务隔离级别下的脏数据。期望：确保 buy-item 在事务内完成扣费后再返回最终余额。
- **手感问题**（不是 bug，但影响体验）：
  - 4 只龟再次全部 hunger=0，cleanliness 和 mood 普遍偏低（小草 mood=0/cleanliness=20）。玩家断签 10h（上轮 16:46 → 本轮 02:46）后，所有龟进入"濒死"状态，对佛系定位略残酷。
  - 小草 vitality=45，若再推 5~10 天无干预，可能跌破 40 进入"生病"状态。但目前未见生病机制触发。
  - season 连续 44 天均为 spring/summer，无秋冬切换，季节系统仍像静态标签。
  - income 在龟状态极差时仍有下限 26（4 龟合计），经济惩罚不够尖锐。
- **改进建议**（按性价比排序）：
  - 🔥 高性价比：**SQLite 并发写保护**。在 `main.go` 初始化时执行 `PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;`，可大幅缓解并发锁（2 行代码）。
  - 🔥 高性价比：**统一端口为 1517**。检查启动脚本/环境变量 `PORT`，确保外网可访问；若需保留 1518，则加 `iptables -t nat -A PREROUTING -p tcp --dport 1517 -j REDIRECT --to-port 1518`。
  - 🔥 高性价比：**后端日志落盘**。在 `main.go` 加 `log.SetOutput(io.MultiWriter(os.Stderr, logFile))`，将访问日志写入 `./data/tk.log`（3 行代码）。
  - 一般：繁殖系统给出条件提示，让玩家知道 intimacy/同缸天数门槛。
  - 一般：长期 hunger=0 时加速 health 衰减（ vitality/appetite 每 2 天 -2，skin 每 5 天 -1），让"断签惩罚"更立体但不过分。
  - 一般：季节切换增加倒计时或概率触发（如每 15 天强制换季），强化季节存在感。
- **下轮重点关注**：
  - SQLite 锁修复验证：并发 4 个 feed 是否不再报错
  - 外网连通性：1517/1518 端口统一后 curl 外网是否通
  - 小草生存线：再推进 5~10 天看 vitality 是否跌破 40，以及是否有"生病"状态触发
  - 后端日志：确认是否有新的日志文件生成

---

### 🐢 测试轮次 2026-05-27 16:46
- **环境快照**：day=41, season=summer, coins=182, 龟数=4（小麝香♀、小草♂、新朋友♂、小巴西♀）
- **与上轮对比**：上轮 day=34/coins=365/龟数=3；本轮新增小巴西（turtle_ubauh3uh，字符串ID），coins 下降主要因清洁花费 120 + advance-day 收入波动
- **游玩动作**（15 步真实接口）：
  1. `POST /api/feed`（turtle_1, food_1）→ ✅ hunger+30, intimacy+2, mood+1
  2. `POST /api/feed`（turtle_2, food_1）→ ✅ 同上
  3. `POST /api/feed`（turtle_1779853832205144140, food_1）→ ✅ 大数字ID仍可正常操作
  4. `POST /api/interact`（turtle_1, pet）→ ✅ status=ok
  5. `POST /api/interact`（turtle_ubauh3uh, pet）→ ✅ status=ok
  6. `POST /api/clean`（tank_1）→ ✅ deep_clean, cost=60, 水质归零
  7. `POST /api/advance-day` → ✅ day=36, income=36, income_breakdown 明细清晰（4 龟分别 10/5/9/12）
  8. `GET /api/turtle?id=turtle_ubauh3uh` → ✅ 字符串ID正常解析，返回完整档案+suggestions+water_history
  9. `POST /api/move-turtle`（小巴西→tank_1）→ ✅ 返回 JSON error：`{"status":"error","message":"巴西龟更适合中水位缸,不能搬到深水缸"}` — **上轮 P1 bug 已修复**
  10. `POST /api/add-decor`（tank_2, wood）→ ✅ 成功放置沉木
  11. `POST /api/advance-day` ×5（极端场景：连续推进 5 天）→ day=37~41, income 33→33→29→25→25，无繁殖/季节事件
  12. `POST /api/move-turtle`（小草→tank_1）→ ✅ JSON error（草龟不能进深水）
  13. `POST /api/move-turtle`（小麝香→tank_2）→ ✅ JSON error（麝香不能进中水）
  14. `POST /api/clean`（tank_1，连续清洁）→ ✅ cost=60，水质已是 100 仍扣费且无提示
  15. `GET /api/shop-catalog` → ✅ 返回 6 种商品，但无 species 购买列表
- **发现的 Bug**（按严重度 P0/P1/P2）：
  - [P0] **上轮服务崩溃问题 → 已解决**。服务自 11:47 重启后稳定运行至今约 5h，HTTP 200 正常。
  - [P1] **上轮 move-turtle 纯文本返回 → 已修复**。habitat 不匹配时现在返回标准 JSON error，前端可正常解析。
  - [P1] **上轮大数字ID问题 → 部分修复**。新购龟 turtle_ubauh3uh 使用字符串ID，旧大数字ID turtle_1779853832205144140 仍可正常操作，后端兼容。
  - [P2] **连续清洁无提示/无折扣**。clarity=100、ammonia=0 时再次清洁仍扣 60 龟币，无"已经很干净"提示。期望：水质良好时（clarity≥90 且 ammonia<0.3）返回 `{"status":"ok","message":"水质良好，无需深度清洁","cost":0}` 或类似提示。
  - [P2] **shop-catalog 缺少 species 列表**。返回结构仅有 `items`（食物/工具），无龟种购买入口。前端若依赖 shop-catalog 展示全部购买项，会漏掉 `/api/buy-species` 功能。期望：追加 `species` 数组，含各龟种 id/name/price/unlock_condition。
  - [P2] **繁殖系统仍无触发**。day=41，tank_2 内同种异性（新朋友♂ + 小巴西♀）同缸多日，但 breed_messages 始终 null。intimacy 过低（2 vs 5）可能是主因，但玩家完全不知道需要多少亲密度。期望：给出繁殖条件进度提示。
  - [P2] **hunger 衰减偏快**。feed 后 hunger=50，连续 advance-day 5 天后全部归零（每天约 -10），意味着玩家必须每天至少登录喂一次，否则 5 天后所有龟饥饿=0、收入降至 25。对"佛系养龟"定位略苛刻。
  - [P3] **后端日志缺失**。/tmp/tk.log 仅含启动日志，无请求访问日志，无法排查线上问题。
- **手感问题**（不是 bug，但影响体验）：
  - `income_breakdown` 明细很棒，玩家能直观看到每只龟的产币贡献。
  - 小麝香 intimacy=34（远高于其他龟的 2~6），说明互动可无限累积且无衰减/上限，长期会形成"只养一只龟"的边际效应。
  - 连续 41 天均为 spring/summer，无秋冬切换，季节系统存在感弱。
  - 新龟小巴西初始 health=100，但 5 天未喂食后 hunger=0，没有"新龟保护期"缓冲。
- **改进建议**（按性价比排序）：
  - 🔥 高性价比：连续清洁时若 clarity≥90 且 ammonia<0.3，提示无需清洁并跳过扣费（1 行阈值判断）。
  - 🔥 高性价比：`shop-catalog` 追加 `species` 数组，与 `buy-species` 接口对齐（数据结构已存在，只需聚合返回）。
  - 🔥 高性价比：后端增加请求访问日志 middleware（每次请求打印 method/path/status/time，便于线上排查）。
  - 一般：繁殖系统给出条件提示（如"亲密度 5/50、同缸天数 6/7"），让玩家知道还需做什么。
  - 一般：调整 hunger 衰减速度——幼体/成体区分，或饥饿<30 时 advance-day 只减 5 而非 10，降低"断签"惩罚。
  - 一般：增加季节事件触发频率或倒计时提示，强化季节系统存在感。
- **下轮重点关注**：
  - 繁殖系统：把 tank_2 的小巴西和新朋友亲密度刷到 20+ 后再推进，验证是否触发产蛋
  - shop-catalog 的 species 列表修复
  - 长期稳定性：验证服务能否持续 10h+ 不崩溃（当前已稳定 5h）
  - 后端日志路径：确认是否有其他日志文件或需要开启 debug 模式

---

### 🐢 测试轮次 2026-05-27 11:46
- **环境快照**：day=34, season=summer, coins=365, 龟数=3（小麝香♀、小草♂、巴西龟♂）
- **游玩动作**：（列 13 个真实跑过的接口调用，含结果）
  1. `curl localhost:1517/` → ❌ HTTP 000，服务已崩溃；手动 `./turtlekeeper` 重启后 → ✅ HTTP 200
  2. `GET /api/state` → ✅ day=25, coins=483, 2 龟 2 缸，inventory 有 food_1×34 等
  3. `POST /api/feed`（turtle_1, food_1）→ ✅ hunger+30, intimacy+2, mood+1, vitality_delta=0
  4. `POST /api/feed`（turtle_2, food_1）→ ✅ 同上
  5. `POST /api/interact`（turtle_1, pet）→ ✅ status=ok
  6. `POST /api/clean`（tank_2）→ ✅ deep_clean, cost=60, 水质归零
  7. `POST /api/buy-species`（redEaredSlider, 300 龟币）→ ✅ 新龟 turtle_1779853832205144140 入住 tank_2
  8. `POST /api/advance-day` → ✅ day=26, income=27（新龟贡献 12），season=spring
  9. `GET /api/turtle?id=turtle_1779853832205144140` → ⚠️ 返回合法 JSON 但 `jq` parse error（大数字 ID 或 UTF-8 截断）
  10. `POST /api/advance-day` ×5 → day=27~31, season spring→summer(day=30), income 27→24 稳定
  11. `POST /api/add-decor`（tank_1, wood）→ ✅ 成功放置沉木，decor id=decor_1779853869538246567
  12. `POST /api/move-turtle`（巴西龟→tank_1）→ ⚠️ 返回纯文本 `"巴西龟更适合中水位缸，不能搬到深水缸"`（非 JSON）
  13. `POST /api/advance-day` ×3（繁殖测试）→ day=32~34, income=24, breed_messages=null, new_eggs=0
  14. `GET /api/pokedex` → ✅ 37% 完成度，3/8 解锁；未解锁物种 description="???"、trivia=""
- **发现的 Bug**（按严重度 P0/P1/P2）：
  - [P0] **服务在 cron 唤醒前已崩溃**。本轮开始时 curl 返回 HTTP 000，进程不存在。日志仅含启动信息无 panic，疑似被系统终止或 OOM。需加守护进程（systemd/nohup+循环）或查 `/var/log/syslog` 确认死因。
  - [P1] `/api/move-turtle`  habitat 不匹配时返回**纯文本而非 JSON**。前端调用 `response.json()` 会直接抛异常，整条交互链路中断。复现：巴西龟（middle）→ tank_1（deep）→ 返回 `"巴西龟更适合中水位缸，不能搬到深水缸"`。期望：`{"status":"error","message":"巴西龟更适合中水位缸，不能搬到深水缸"}`。
  - [P1] `/api/turtle?id=<大数字ID>` 导致 jq parse error。新买巴西龟的 turtle_id=`1779853832205144140`，该值超出 JavaScript 安全整数范围（2^53-1≈9e15），前端/工具解析时会失真或报错。期望：ID 统一用字符串 UUID（如 `"turtle_abc123"`）或缩短为 64 位字符串。
  - [P2] 图鉴未解锁物种显示 `"???"` 和空字段。剃刀龟、果核泥龟等 description="???", trivia="", scientific_name=""，前端展示极其粗糙。期望：未解锁项给出模糊剪影文案，如 `"这是一种神秘的蛋龟…继续探索以解锁详情"`，而非三个问号。
  - [P2] 新购龟初始 hunger=0。巴西龟刚买到手就显示"饥饿度 0 / 很饿了"，与"新生命活力满满"的直觉冲突。期望：新龟初始化 hunger=50~70，给玩家一个缓冲期。
  - [P2] 繁殖系统未触发。day=34，同缸异性（小麝香♀ + 巴西龟♂ 曾同缸），intimacy 有差异（27 vs 0），但 breed_messages 始终 null。需确认繁殖条件：是否要求 intimacy≥50？是否要求同缸≥7 天？建议给玩家进度提示（如"亲密度不足，还需互动 X 次"）。
- **手感问题**（不是 bug，但影响体验）：
  - 小草的 health.vitality 从 77→65（10 天降 12），但 skin/shell 始终 100。长期饥饿只衰减活力/食欲，不影响皮肤/壳况，与真实养龟逻辑不符（饿久了会腐皮、软壳）。
  - income 在 mood=0/cleanliness=0 时仍有下限 24（3 龟合计），缺乏"濒死惩罚"。小麝香 vitality=83 仍在产币 10，经济系统对负向状态的反馈不够尖锐。
  - `add-decor` 成功后没有视觉/评分反馈。造景系统目前只是"放了东西"，缺少"生态友好度"评分或龟的互动反馈（如麝香龟钻沉木）。
- **改进建议**（按性价比排序）：
  - 🔥 高性价比：`move-turtle` 错误返回改为 JSON 格式（1 行 `json.NewEncoder(w).Encode` 替换 `fmt.Fprint`）。
  - 🔥 高性价比：新龟初始化 hunger=60、cleanliness=80、mood=60，避免"刚买就濒死"。
  - 🔥 高性价比：turtle ID 生成改用字符串 UUID（如 `uuid.New().String()` 或短 8 位随机），消除大整数溢出问题。
  - 一般：图鉴未解锁项增加占位文案，避免 `"???"` 出现。
  - 一般：长期 hunger=0 时，每 5 天 skin/shell -1，让"饿"的后果更立体。
  - 一般：繁殖系统给出条件进度条（亲密度 X/100、同缸天数 Y/7），让玩家知道还需做什么。
- **下轮重点关注**：
  - 服务稳定性：加 systemd 守护或 nohup 循环，防止再次崩溃
  - 大数字 ID 问题：验证前端是否能正常展示/操作新龟
  - 繁殖系统：把两只龟 intimacy 都刷到 50+ 后再推进看是否触发产蛋
  - 造景评分：add-decor 后是否有生态评分变化（目前 state 中 decor 仅返回 null 或列表，无评分字段）

---

### 🐢 测试轮次 2026-05-27 06:46
- **环境快照**：day=21, season=spring, coins=591, 龟数=2（小麝香、小草）
- **游玩动作**：
  1. `POST /api/feed`（turtle_1, 无 player_id）→ ✅ 已修复上轮 bug，正常喂食，hunger+30, intimacy+2, mood+1
  2. `POST /api/interact`（turtle_1, pet）→ ✅ status=ok
  3. `POST /api/clean`（turtle_id）→ ❌ "tank_id required"；改传 tank_id 后 → ✅ deep_clean, cost=60, 水质归零
  4. `POST /api/maintain-tank`（tank_1, partial_change）→ ✅ cost=20, ammonia 0.94→0.39
  5. `POST /api/feed`（turtle_2）→ ✅ 同上
  6. `POST /api/maintain-tank`（tank_2, partial_change）→ ✅ cost=20, ammonia 1.03→0.43
  7. `POST /api/advance-day` → ✅ day=16, income=18（两只龟各 9）
  8. `GET /api/turtle?id=turtle_1` → ✅ 返回完整档案 + suggestions（3 条预警）+ water_history（14 天记录）
  9. `POST /api/buy-item`（food_1×1）→ ✅ cost=12, gained=10, coins=501
  10. `POST /api/add-decor`（tank_1, decor_1）→ ❌ "tank_id and decor.type required"，参数命名与 catalog 不一致
  11. `POST /api/advance-day` ×5 → day=17~21, income 恒为 18，无繁殖、无季节事件
  12. `POST /api/clean`（tank_1, tank_id）→ ✅ deep_clean, cost=60
- **发现的 Bug**（按严重度 P0/P1/P2）：
  - [P1] `/api/feed` 返回字段 `vitality: 0` 语义不明。health.vitality 保持 100 未变，但接口返回了 0，前端若直接展示会造成玩家困惑（" vitality 归零？"）。期望：返回增量或移除该字段。
  - [P2] `/api/clean` 参数设计反直觉。传入 turtle_id 被拒，要求 tank_id；但玩家思维是"给某只龟清洁"而非"给某个缸清洁"。建议：支持 turtle_id（自动找所属 tank）或改进错误提示为 "请传入 tank_id 或 turtle_id"。
  - [P2] `/api/add-decor` 参数与 catalog 脱节。decor-catalog 返回 `type` 字段，但接口要求 `decor.type`（而非 `decor_id`），错误提示未说明有效值范围。建议：统一用 `decor_id` 或 `decor_type`，并在错误中给出可选值。
  - [P2] 饥饿为 0 长期化后 health 衰减过慢。day=0→21，hunger 几乎恒为 0，但 vitality/appetite 仅从 100→97/89（约 3 天降 1 点），skin/shell 纹丝不动。真实逻辑：长期饥饿应更快损耗 vitality，并影响 skin（脱水/腐皮）。
  - [P2] income 在 mood=0/cleanliness=0 时仍有下限 18。day=9 时降至 18 后便不再下降，龟处于"濒死"状态却稳定产币，经济系统缺乏惩罚深度。
  - [P3] 部分 POST 响应在并发测试时偶发 jq parse error，但原始输出为合法 JSON，疑似响应头未正确设置 Content-Type 或尾部含不可见字符。
- **手感问题**（不是 bug，但影响体验）：
  - `water_history` 只在 `/api/turtle` 中返回，主状态 `/api/state` 没有，玩家换水后无法直观看到水质趋势。
  - `suggestions` 仅存在于单龟详情页，state 中没有汇总预警，玩家需要逐只点进去才能看到"该喂食了"。
  - decor catalog 中沉木/晒台石 cost=0，但 add-decor 接口门槛高（需 decor.type），容易让玩家误以为造景系统不可用。
- **改进建议**（按性价比排序）：
  - 🔥 高性价比：明确 `/api/feed` 返回值语义——若 vitality 是增量则 hunger=0 时应为 0（未实现衰减），直接移除更干净；若为当前值则与 health.vitality 不一致，属于数据 bug。
  - 🔥 高性价比：`/api/clean` 支持 `turtle_id` 参数，自动解析到所属 tank，提升 API 易用性。
  - 🔥 高性价比：在 `/api/state` 中给每只龟追加 `suggestions` 数组（最多 2 条高优先级），让玩家一打开游戏就知道该干什么。
  - 一般：加速饥饿衰减逻辑——hunger=0 连续 3 天后，vitality/appetite 每天 -2，skin 每 5 天 -1，让"忘记喂食"有真实后果。
  - 一般：income 下限改为与 health 四维平均值挂钩，health<80 时收入开始衰减，health<50 时接近 0，强化"养好龟=赚更多"的反馈。
  - 一般：`/api/add-decor` 参数与 decor-catalog 字段对齐（`decor_type` 或 `type`），并在 400 错误中返回可用类型列表。
- **下轮重点关注**：
  - feed API 的 vitality 字段到底是 bug 还是设计如此
  - 繁殖系统：把两只龟移到同缸后推进到 day≥30 看是否触发产蛋
  - 长期挂机：再推进 10~15 天看 hunger=0 的龟是否会进入"生病"状态

---

### 🐢 测试轮次 2026-05-27 01:46
- **环境快照**：day=9, season=spring, coins=607, 龟数=2（小麝香、小草）
- **游玩动作**：
  1. `GET /api/state` → 正常返回，default 玩家有 2 龟 2 缸，inventory 有 food_1×20 等
  2. `POST /api/feed`（body 无 player_id）→ ❌ "该食物已用完，请到商店补货"（inventory 明明有 20 个）
  3. `POST /api/feed`（body 带 player_id="default"）→ ✅ 正常，hunger+30, intimacy+2, mood+1
  4. `POST /api/interact`（pet）→ ✅ status=ok
  5. `POST /api/maintain-tank`（tank_1, partial_change）→ ✅ cost=20, 水质重置
  6. `POST /api/advance-day` → ✅ day=2, income=24, season=spring, 无繁殖事件
  7. `POST /api/buy-item`（food_1×1）→ ✅ cost=12, gained=10, coins=492
  8. `POST /api/maintain-tank`（tank_2, scoop_waste）→ ✅ cost=0, clarity=100
  9. `POST /api/maintain-tank`（tank_2, partial_change）→ ✅ cost=20, 水质好转
  10. `POST /api/advance-day` ×7 → day=3~9, income 从 24→21→18（因 mood/cleanliness 下降）
  11. `GET /api/turtle?id=turtle_1` → ✅ 返回详情+suggestions，但 suggestions 仅一条（hunger=0）
  12. `POST /api/feed`（turtle_1 hunger=0）→ ✅ 允许喂食，hunger+30（未溢出）
  13. `POST /api/maintain-tank`（tank_2, install_filter）→ ✅ cost=180, has_filter=true
- **发现的 Bug**（按严重度 P0/P1/P2）：
  - [P1] `/api/feed` player_id 必须从 JSON body 传入，否则查询空字符串玩家导致食物"永远用完"。复现：`curl -d '{"turtle_id":"turtle_1","food_id":"food_1"}' /api/feed` → "该食物已用完"。期望：body 无 player_id 时默认 "default"（与 maintain-tank/interact/advance-day 等保持一致）。
  - [P2] `buildTurtleSuggestions` mood 阈值 `<=40` 过严。turtle_1 mood=49 时无心情提示，但已明显低于健康线。建议调到 `<=50` 或 `<=45`。
  - [P2] 饥饿为 0 长期不影响 health 四维。advancePlayerTurtles 中 hunger 只影响 income（间接 via mood），但 vitality/appetite/skin/shell 不因饥饿下降。真实养龟逻辑：连续饥饿应缓慢损耗活力/食欲。
  - [P3] `handleInteract` 的 `case "check"` 为空实现，前端若调用 check 无任何数据返回，造成困惑。建议补实现或移除该 case。
  - [P3] 浮点数精度溢出：`ammonia=0.6155183999999999` 等，建议 round 到 2 位小数后再入库/返回。
- **手感问题**（不是 bug，但影响体验）：
  - tank_2（草龟的家）默认无过滤器，第 9 天氨 0.88、清澈度 65，水质恶化速度快于新手学习曲线。建议初始赠送一个过滤器或开局引导安装。
  - income 7 天内从 24 降到 18，降幅 25%，对新手期略有挫败感。尤其是 feed API 有 bug 时，玩家更难及时发现需要喂食。
  - suggestions 仅提示 hunger，cleanliness=54 和 mood=49 均未触发。阈值整体偏保守，容易让玩家错过最佳干预窗口。
- **改进建议**（按性价比排序）：
  - 🔥 高性价比：在 `handleFeed` 顶部加 `if req.PlayerID == "" { req.PlayerID = "default" }`（与其他 handler 一致，1 行代码修复核心功能）。
  - 🔥 高性价比：饥饿为 0 时增加 health 衰减逻辑，如连续 hunger≤0 的天数每 2 天 vitality/appetite -1，让"饿"有真实后果。
  - 一般：水质数值 round(float, 2) 后入库，避免 `0.6155183999999999` 这种脏数据。
  - 一般：调整 suggestions 阈值 mood≤50、cleanliness≤60，提升预警敏感度。
  - 一般：tank_2 默认 has_filter=true，或开局弹窗引导安装过滤器。
- **下轮重点关注**：
  - feed API 修复后验证前端喂食链路
  - 繁殖系统：推进到 day≥20 看同缸异性高亲密龟是否产蛋
  - 长期挂机测试：推进 30+ 天看龟是否会饿死/水质崩溃

---
<!-- AI_TESTER_LOG_END -->
