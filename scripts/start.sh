#!/bin/bash
# TurtleKeeper 守护启动脚本
# 用法: ./scripts/start.sh [端口]
# 若进程崩溃会自动重启

set -e

PORT="${1:-1517}"
APP_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BINARY="$APP_DIR/backend/turtlekeeper"
LOG_FILE="$APP_DIR/backend/turtlekeeper.log"
PID_FILE="$APP_DIR/backend/turtlekeeper.pid"

cd "$APP_DIR/backend"

# 先 build
echo "[TurtleKeeper] 正在编译..."
go build -o turtlekeeper .

# 停止旧进程
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE" 2>/dev/null || true)
    if [ -n "$OLD_PID" ] && kill -0 "$OLD_PID" 2>/dev/null; then
        echo "[TurtleKeeper] 停止旧进程 PID=$OLD_PID"
        kill "$OLD_PID" 2>/dev/null || true
        sleep 1
    fi
fi

# 清理旧日志（保留最近 3 天）
find "$APP_DIR/backend" -name "turtlekeeper.log.*" -mtime +3 -delete 2>/dev/null || true

# 启动守护循环
(
    while true; do
        echo "[TurtleKeeper] $(date '+%F %T') 启动服务，端口 $PORT"
        # 轮转日志
        if [ -f "$LOG_FILE" ] && [ $(stat -f%z "$LOG_FILE" 2>/dev/null || stat -c%s "$LOG_FILE" 2>/dev/null || echo 0) -gt 10485760 ]; then
            mv "$LOG_FILE" "$LOG_FILE.$(date +%s)"
        fi
        PORT="$PORT" nohup "$BINARY" >> "$LOG_FILE" 2>&1 &
        PID=$!
        echo $PID > "$PID_FILE"
        wait $PID
        EXIT_CODE=$?
        echo "[TurtleKeeper] $(date '+%F %T') 进程退出，码=$EXIT_CODE，5秒后重启..."
        sleep 5
    done
) &

echo "[TurtleKeeper] 守护进程已启动，日志: $LOG_FILE"
echo "[TurtleKeeper] 查看状态: tail -f $LOG_FILE"
echo "[TurtleKeeper] 停止服务: kill \$(cat $PID_FILE)"
