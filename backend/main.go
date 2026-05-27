package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

	// 日志落盘：同时输出到 stderr 和文件（AI_TESTER P2）
	logPath := filepath.Join(dataDir, "tk.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	} else {
		log.Printf("警告: 无法打开日志文件 %s: %v", logPath, err)
	}

	// 静态文件服务
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/", fs)

	// API 路由：统一通过 ensurePlayerMiddleware，避免写 API 冷启动后遇到 player not found。
	// 统一套 loggingMiddleware，记录请求耗时与状态码。
	api := http.NewServeMux()
	api.HandleFunc("/api/state", handleGetState)
	api.HandleFunc("/api/feed", ensurePlayerMiddleware(handleFeed))
	api.HandleFunc("/api/clean", ensurePlayerMiddleware(handleClean))
	api.HandleFunc("/api/maintain-tank", ensurePlayerMiddleware(handleMaintainTank))
	api.HandleFunc("/api/interact", ensurePlayerMiddleware(handleInteract))
	api.HandleFunc("/api/add-decor", ensurePlayerMiddleware(handleAddDecor))
	api.HandleFunc("/api/move-decor", ensurePlayerMiddleware(handleMoveDecor))
	api.HandleFunc("/api/advance-day", ensurePlayerMiddleware(handleAdvanceDay))
	api.HandleFunc("/api/species", handleGetSpecies)
	api.HandleFunc("/api/pokedex", handleGetPokedex)
	api.HandleFunc("/api/decor-catalog", handleGetDecorCatalog)
	api.HandleFunc("/api/shop-catalog", handleGetShopCatalog)
	api.HandleFunc("/api/buy-item", ensurePlayerMiddleware(handleBuyItem))
	api.HandleFunc("/api/buy-species", ensurePlayerMiddleware(handleBuySpecies))
	api.HandleFunc("/api/create-tank", ensurePlayerMiddleware(handleCreateTank))
	api.HandleFunc("/api/move-turtle", ensurePlayerMiddleware(handleMoveTurtle))
	api.HandleFunc("/api/turtle", handleTurtleDetail)
	api.HandleFunc("/api/breeding-hints", handleBreedingHints)
	http.Handle("/api/", loggingMiddleware(api))

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

// loggingMiddleware 记录每个请求的 method、path、status、耗时。
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// 包装 ResponseWriter 以捕获状态码
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("[%s] %s → %d (%s)", r.Method, r.URL.Path, lw.statusCode, time.Since(start))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
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
