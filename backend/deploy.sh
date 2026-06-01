#!/usr/bin/env bash
# TurtleKeeper 一键部署脚本 (Linux)
# 用法: bash deploy.sh [install|start|stop|status|restart|logs|update]
set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVICE_FILE="$PROJECT_DIR/turtlekeeper.service"
SYSTEMD_PATH="/etc/systemd/system/turtlekeeper.service"
BINARY="$PROJECT_DIR/turtlekeeper"
PORT=1517

case "${1:-status}" in
  install)
    echo "🔨 编译 Go 二进制..."
    cd "$PROJECT_DIR" && go build -o turtlekeeper .
    echo "📦 安装 systemd 服务..."
    sudo cp "$SERVICE_FILE" "$SYSTEMD_PATH"
    sudo systemctl daemon-reload
    sudo systemctl enable turtlekeeper
    sudo systemctl start turtlekeeper
    sleep 2
    sudo systemctl status turtlekeeper --no-pager -l | head -20
    echo ""
    echo "✅ 安装完成! 访问 http://localhost:$PORT/"
    ;;
  start)
    sudo systemctl start turtlekeeper
    sudo systemctl status turtlekeeper --no-pager | head -10
    ;;
  stop)
    sudo systemctl stop turtlekeeper
    ;;
  restart)
    sudo systemctl restart turtlekeeper
    sudo systemctl status turtlekeeper --no-pager | head -10
    ;;
  status)
    sudo systemctl status turtlekeeper --no-pager -l | head -20
    echo ""
    echo "🔍 端口检查:"
    ss -tlnp 2>/dev/null | grep ":$PORT" || echo "  端口 $PORT 未监听"
    echo ""
    echo "🌐 健康检查:"
    curl -sS -o /dev/null -w "  HTTP %{http_code} | %{time_total}s\n" "http://localhost:$PORT/" || echo "  本地访问失败"
    ;;
  logs)
    sudo journalctl -u turtlekeeper -n 50 --no-pager
    ;;
  update)
    echo "🔨 重新编译..."
    cd "$PROJECT_DIR" && go build -o turtlekeeper .
    echo "🔄 重启服务..."
    sudo systemctl restart turtlekeeper
    sleep 2
    sudo systemctl status turtlekeeper --no-pager | head -10
    ;;
  *)
    echo "用法: $0 [install|start|stop|restart|status|logs|update]"
    exit 1
    ;;
esac
