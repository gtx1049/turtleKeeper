# 🐢 龟乐园 · TurtleKeeper

> 一款移动端竖屏的乌龟饲养模拟游戏。蛋龟玩水，半水龟晒背，水龟摆造景——把每一只龟都当成有性格的小生命来养。

**"佛系养龟 · 真实生态 · 收集图鉴"** —— 融合养成、轻经营、布景 DIY 的休闲模拟器。

---

## ✨ 设计哲学

- **真实感优先**：每个龟种的生活习性基于真实生物学
- **经营乐趣 > 繁琐操作**：喂食/换水一键化，布景、调温、繁殖才是深度玩法
- **慢节奏疗愈**：不搞肝度，主打"看龟发呆"的解压属性
- **图鉴驱动**：以"集齐 30+ 龟种"为长期目标，每只龟有独立性格档案

## 🎯 目标用户

真实养龟玩家 / 养成模拟爱好者 / 想养不能养的"云龟党"

---

## 🛠 技术架构

| 层 | 选型 |
|---|---|
| 后端 | Go + Gin + SQLite，单可执行文件 |
| 前端 | H5 竖屏 Canvas 2D（max 480px 宽） |
| 部署 | 独立部署 / 远程访问 / 单机版打包 |
| 端口 | `1517` |

### 已实现接口（M2）
- `GET /api/state`、`POST /api/feed`、`POST /api/clean`、`POST /api/interact`
- `POST /api/advance_day`、`POST /api/decorate`
- `GET /api/turtles`、`POST /api/buy_turtle`

### 数据表（7 张）
玩家 / 龟 / 缸 / 造景 / 背包 / 成就 / 解锁龟种

---

## 🚀 快速启动

```bash
cd backend
go build -o turtlekeeper
./turtlekeeper
# 浏览器打开 http://localhost:1517
```

---

## 📍 路线图

| 里程碑 | 内容 | 状态 |
|---|---|---|
| M0 | 企划立项 | ✅ |
| M1 | 像素素材（含文生图方案） | 🚧 |
| M2 | 核心循环 Demo（喂食/清洁/推进日） | ✅ |
| M3 | 布景拖拽系统 | ⏳ |
| M4 | 多缸多龟管理 | ⏳ |
| M5 | 繁殖与图鉴 | ⚡ 繁殖闭环已启用 |
| M6 | 单机版打包 + 商业化 | ⏳ |

完整企划见 [PLAN.md](./PLAN.md)。

---

## 📂 目录结构

```
turtleKeeper/
├── backend/          # Go 后端 + 静态资源
│   ├── main.go
│   ├── handlers.go
│   └── static/       # 前端 H5
├── assets_preview/   # 素材样张
├── PLAN.md           # 完整企划书
└── README.md
```

---

🐢 *Built with 🐢 by 高天星 & 小明 · 2026*
