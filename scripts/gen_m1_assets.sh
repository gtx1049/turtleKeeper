#!/usr/bin/env bash
# 批量生成龟乐园 M1 像素素材（Pollinations.ai 免费文生图）
set -e
OUT=/root/.openclaw/workspace/turtleKeeper/assets_preview/m1
mkdir -p "$OUT"

declare -A turtles=(
  ["muskTurtle"]="cute musk turtle, small dark brown shell, top-down view"
  ["razorbackTurtle"]="razor-back musk turtle, tall keeled brown shell, top-down view"
  ["loggerheadMusk"]="loggerhead musk turtle with yellow stripes on head, top-down view"
  ["yellowMargined"]="yellow-margined box turtle, golden rim shell, dome shaped, top-down view"
  ["chinesePond"]="Chinese pond turtle, dark olive shell with three keels, top-down view"
  ["yellowPond"]="yellow pond turtle, bright yellow throat, smooth olive shell, top-down view"
  ["chineseStripe"]="Chinese stripe-necked turtle, yellow stripes on neck, top-down view"
  ["redEared"]="red-eared slider turtle with red mark on side of head, top-down view"
)

for key in "${!turtles[@]}"; do
  desc="${turtles[$key]}"
  prompt="pixel art, 32x32 sprite, ${desc}, transparent background, simple clean pixels, no shadow, game asset, cute kawaii style"
  enc=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1]))" "$prompt")
  url="https://image.pollinations.ai/prompt/${enc}?width=128&height=128&nologo=true&model=flux&seed=42"
  echo "→ $key"
  curl -s --max-time 60 -o "$OUT/${key}.png" "$url" && \
    echo "  $(file $OUT/${key}.png | cut -d: -f2-)" || echo "  ❌ 失败"
  sleep 1
done

echo ""
echo "=== 生成完成 ==="
ls -la "$OUT"
