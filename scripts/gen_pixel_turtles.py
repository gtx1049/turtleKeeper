#!/usr/bin/env python3
"""
龟乐园 M1 素材：程序化像素龟生成器
- 32×32 PNG，透明背景
- 俯视角（top-down），适配 H5 Canvas 游戏 sprite
- 每种龟独立调色板（基于真实物种）
"""
from PIL import Image, ImageDraw
import os, json

OUT = "/root/.openclaw/workspace/turtleKeeper/assets_preview/m1"
os.makedirs(OUT, exist_ok=True)

# 龟种调色板（壳深色、壳浅色、纹路色、头/腿色、特征点缀）
SPECIES = {
    # 后端 species id 对齐（见 GET /api/species）
    "muskTurtle":         {"name":"麝香龟",     "shell_dark":"#3a2d1e","shell_mid":"#5a4530","shell_pat":"#2a1f12","skin":"#3d2f20","accent":None},
    "razorbackTurtle":    {"name":"剃刀龟",     "shell_dark":"#4a3520","shell_mid":"#7a5a3a","shell_pat":"#2a1f12","skin":"#5a4530","accent":"keel_yellow"},
    "loggerheadMuskTurtle":{"name":"果核泥龟",   "shell_dark":"#5a4020","shell_mid":"#806035","shell_pat":"#3a2510","skin":"#705030","accent":"head_stripe_yellow"},
    "yellowMarginTurtle": {"name":"黄缘闭壳龟", "shell_dark":"#3a2515","shell_mid":"#5a3a20","shell_pat":"#f5c542","skin":"#a07030","accent":"head_yellow"},
    "chinesePondTurtle":  {"name":"中华草龟",   "shell_dark":"#2a3a1a","shell_mid":"#4a5a2a","shell_pat":"#1a2510","skin":"#5a5030","accent":"head_stripe_yellow"},
    "yellowPondTurtle":   {"name":"黄喉拟水龟", "shell_dark":"#3a4520","shell_mid":"#5a6530","shell_pat":"#2a3515","skin":"#a08040","accent":"throat_yellow"},
    "chineseStripeTurtle":{"name":"花龟",       "shell_dark":"#2a3520","shell_mid":"#4a5a35","shell_pat":"#3a2510","skin":"#3a4525","accent":"neck_stripes_white"},
    "redEaredSlider":     {"name":"巴西龟",     "shell_dark":"#2a4525","shell_mid":"#4a6535","shell_pat":"#1a2510","skin":"#3a5025","accent":"ear_red"},
}

def hex_to_rgba(h, a=255):
    h = h.lstrip("#")
    return (int(h[0:2],16), int(h[2:4],16), int(h[4:6],16), a)

def make_turtle(palette, size=32, scale=4):
    """生成单只龟（俯视角），返回放大版 PNG。"""
    W = size
    img = Image.new("RGBA", (W, W), (0,0,0,0))
    d = ImageDraw.Draw(img)

    cx, cy = W//2, W//2

    sd = hex_to_rgba(palette["shell_dark"])
    sm = hex_to_rgba(palette["shell_mid"])
    sp = hex_to_rgba(palette["shell_pat"])
    sk = hex_to_rgba(palette["skin"])
    ac = palette.get("accent")

    # 1) 四肢（先画底层）
    for px,py in [(-9,-7),(8,-7),(-9,7),(8,7)]:
        d.rectangle([cx+px, cy+py-1, cx+px+3, cy+py+2], fill=sk)
    # 2) 尾巴
    d.rectangle([cx-1, cy+10, cx+1, cy+12], fill=sk)
    # 3) 头
    d.ellipse([cx-3, cy-14, cx+3, cy-8], fill=sk)
    # 眼睛
    d.point((cx-2, cy-12), fill=(0,0,0,255))
    d.point((cx+1, cy-12), fill=(0,0,0,255))

    # 4) 龟壳 椭圆（外深内浅）
    d.ellipse([cx-10, cy-9, cx+10, cy+10], fill=sd)
    d.ellipse([cx-9,  cy-8, cx+9,  cy+9 ], fill=sm)

    # 5) 壳面甲片纹（5 块顶骨甲 + 周围肋骨甲）
    # 中线纵向
    d.line([cx, cy-8, cx, cy+8], fill=sp, width=1)
    # 横向 2 条
    d.line([cx-7, cy-3, cx+7, cy-3], fill=sp, width=1)
    d.line([cx-7, cy+3, cx+7, cy+3], fill=sp, width=1)
    # 边缘点缀
    d.line([cx-7, cy-7, cx+7, cy-7], fill=sp, width=1)
    d.line([cx-7, cy+7, cx+7, cy+7], fill=sp, width=1)

    # 6) 特征点缀
    if ac == "keel_yellow":  # 剃刀龟黄棱
        d.line([cx, cy-7, cx, cy+7], fill=(245,200,80,255), width=1)
    elif ac == "head_stripe_yellow":  # 头部黄纹
        d.point((cx-3, cy-11), fill=(245,200,80,255))
        d.point((cx+3, cy-11), fill=(245,200,80,255))
    elif ac == "head_yellow":  # 整头偏黄
        d.ellipse([cx-3, cy-14, cx+3, cy-8], fill=(220,180,80,255))
        d.point((cx-2, cy-12), fill=(0,0,0,255))
        d.point((cx+1, cy-12), fill=(0,0,0,255))
    elif ac == "throat_yellow":  # 黄喉
        d.rectangle([cx-1, cy-9, cx+1, cy-7], fill=(240,210,90,255))
    elif ac == "neck_stripes_white":  # 花龟颈纹
        d.point((cx-1, cy-9), fill=(230,230,210,255))
        d.point((cx+1, cy-9), fill=(230,230,210,255))
    elif ac == "ear_red":  # 巴西龟红耳
        d.point((cx-3, cy-12), fill=(220,40,40,255))
        d.point((cx+3, cy-12), fill=(220,40,40,255))

    # 放大（nearest 保持像素感）
    if scale > 1:
        img = img.resize((W*scale, W*scale), Image.NEAREST)
    return img

def make_contact_sheet():
    """生成总览图（4×2 网格）"""
    keys = list(SPECIES.keys())
    cols, rows = 4, 2
    cell = 160  # 32*5
    sheet = Image.new("RGBA", (cols*cell, rows*cell), (245,240,220,255))
    d = ImageDraw.Draw(sheet)
    for i, k in enumerate(keys):
        r, c = divmod(i, cols)
        x, y = c*cell, r*cell
        turtle = make_turtle(SPECIES[k], scale=4)  # 128
        sheet.paste(turtle, (x+16, y+16), turtle)
        d.text((x+8, y+cell-18), f"{SPECIES[k]['name']}", fill=(60,40,20,255))
    return sheet

if __name__ == "__main__":
    print("=== 生成 8 种像素龟（32×32, 透明背景）===")
    for key, pal in SPECIES.items():
        # 原始 32×32 留档
        raw = make_turtle(pal, scale=1)
        raw.save(f"{OUT}/{key}_32.png")
        # 放大 4× 用于展示和游戏直接使用
        big = make_turtle(pal, scale=4)
        big.save(f"{OUT}/{key}.png")
        print(f"  ✓ {pal['name']:8s} → {key}.png (128×128)")

    sheet = make_contact_sheet()
    sheet.save(f"{OUT}/_contact_sheet.png")
    print(f"\n✓ 总览图: _contact_sheet.png ({sheet.size[0]}×{sheet.size[1]})")

    # 同时写一份 manifest 给前端用
    manifest = {k: {"name":v["name"], "sprite":f"sprites/{k}.png"} for k,v in SPECIES.items()}
    with open(f"{OUT}/manifest.json", "w") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)
    print("✓ manifest.json")
