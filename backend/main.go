package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// 获取数据目录
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	os.MkdirAll(dataDir, 0755)

	// 初始化数据库
	if err := initDB(dataDir); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 静态文件服务
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/", fs)

	// API 路由：统一通过 ensurePlayerMiddleware，避免写 API 冷启动后遇到 player not found。
	http.HandleFunc("/api/state", handleGetState)
	http.HandleFunc("/api/feed", ensurePlayerMiddleware(handleFeed))
	http.HandleFunc("/api/clean", ensurePlayerMiddleware(handleClean))
	http.HandleFunc("/api/maintain-tank", ensurePlayerMiddleware(handleMaintainTank))
	http.HandleFunc("/api/interact", ensurePlayerMiddleware(handleInteract))
	http.HandleFunc("/api/add-decor", ensurePlayerMiddleware(handleAddDecor))
	http.HandleFunc("/api/move-decor", ensurePlayerMiddleware(handleMoveDecor))
	http.HandleFunc("/api/advance-day", ensurePlayerMiddleware(handleAdvanceDay))
	http.HandleFunc("/api/species", handleGetSpecies)
	http.HandleFunc("/api/decor-catalog", handleGetDecorCatalog)
	http.HandleFunc("/api/shop-catalog", handleGetShopCatalog)
	http.HandleFunc("/api/buy-item", ensurePlayerMiddleware(handleBuyItem))
	http.HandleFunc("/api/buy-species", ensurePlayerMiddleware(handleBuySpecies))
	http.HandleFunc("/api/create-tank", ensurePlayerMiddleware(handleCreateTank))
	http.HandleFunc("/api/move-turtle", ensurePlayerMiddleware(handleMoveTurtle))
	http.HandleFunc("/api/turtle", handleTurtleDetail)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "1517"
	}

	// 启动服务器
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("🐢 龟乐园 · TurtleKeeper 启动中...")
	log.Printf("📍 访问地址: http://localhost:%s", port)
	log.Printf("📂 数据目录: %s", dataDir)
	log.Printf("📂 静态文件: %s", staticDir)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

// ensurePlayerMiddleware 在任何写 API 调用前先懒初始化 default 玩家，
// 避免冷启动 + 直接访问写接口出现 player not found。
// 该中间件不读 body，只从 query 里取 player_id（其他在 handler 里再取）。
func ensurePlayerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pid := r.URL.Query().Get("player_id")
		if pid == "" {
			pid = "default"
		}
		if _, err := getOrCreatePlayer(pid); err != nil {
			http.Error(w, "ensure player failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		next(w, r)
	}
}

// 获取静态文件路径（用于内嵌资源）
func getStaticPath(relativePath string) string {
	// 先检查相对路径
	if _, err := os.Stat(relativePath); err == nil {
		return relativePath
	}

	// 再检查可执行文件同目录
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	altPath := filepath.Join(exeDir, relativePath)
	if _, err := os.Stat(altPath); err == nil {
		return altPath
	}

	return relativePath
}
