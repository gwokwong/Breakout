#!/bin/bash
# 破局 Breakout 一键启动脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
echo "========================================="
echo "  破局 Breakout 一键启动脚本"
echo "========================================="

# ============ 1. 检查并启动 MySQL ============
echo ""
echo "[1/4] 检查 MySQL..."
if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "teampanel-mysql"; then
  echo "  ✓ MySQL 容器已在运行"
else
  echo "  启动 MySQL 容器..."
  docker run -d --name teampanel-mysql \
    -p 3306:3306 \
    -e MYSQL_ROOT_PASSWORD=123456 \
    -e MYSQL_DATABASE=teampanel \
    -v "$ROOT_DIR/backend/sql/init.sql:/docker-entrypoint-initdb.d/init.sql" \
    mysql:8.0 2>/dev/null || docker start teampanel-mysql
  echo "  等待 MySQL 就绪..."
  sleep 10
  echo "  ✓ MySQL 已启动"
fi

# ============ 2. 编译并启动后端 ============
echo ""
echo "[2/4] 编译并启动后端..."
cd "$ROOT_DIR/backend"
if [ ! -f breakout ]; then
  go build -o breakout .
fi
# 如果后端已在运行，先停止
if lsof -i :8080 -t >/dev/null 2>&1; then
  echo "  后端已在运行，跳过启动"
else
  ./breakout &
  BACKEND_PID=$!
  echo "  ✓ 后端已启动 (PID: $BACKEND_PID, 端口: 8080)"
  sleep 2
fi

# ============ 3. 安装前端依赖 ============
echo ""
echo "[3/4] 检查前端依赖..."
cd "$ROOT_DIR/frontend"
if [ ! -d node_modules ]; then
  echo "  安装依赖中..."
  npm install
fi
echo "  ✓ 前端依赖已就绪"

# ============ 4. 启动前端开发服务器 ============
echo ""
echo "[4/4] 启动前端开发服务器..."
if lsof -i :5173 -t >/dev/null 2>&1; then
  echo "  前端已在运行，跳过启动"
else
  npm run dev &
  FRONTEND_PID=$!
  echo "  ✓ 前端已启动 (PID: $FRONTEND_PID, 端口: 5173)"
fi

echo ""
echo "========================================="
echo "  启动完成！"
echo "  前端地址: http://localhost:5173"
echo "  后端地址: http://localhost:8080"
echo "  测试账号: 13800000001 / 123456"
echo "========================================="
echo ""
echo "按 Ctrl+C 停止所有服务"
wait