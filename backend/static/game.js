// 龟乐园 · TurtleKeeper 前端主逻辑

const API_BASE = '';
let gameState = null;
let currentTurtleId = null;
let currentTankId = null;
let canvas, ctx;
let turtles = []; // 渲染用的乌龟对象
let decor = [];   // 渲染用的造景
let animationId;
let lastTime = 0;
let draggedDecorId = null;
let dragDirty = false;

// ============ 初始化 ============
document.addEventListener('DOMContentLoaded', async () => {
    canvas = document.getElementById('tank-canvas');
    ctx = canvas.getContext('2d');
    
    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);
    
    // 加载游戏状态
    await loadGameState();
    
    // 绑定事件
    bindEvents();
    
    // 启动渲染循环
    requestAnimationFrame(gameLoop);
});

function resizeCanvas() {
    const container = document.getElementById('tank-view');
    canvas.width = container.clientWidth;
    canvas.height = container.clientHeight;
}

async function loadGameState() {
    try {
        const res = await fetch(`${API_BASE}/api/state?player_id=default`);
        gameState = await res.json();
        
        updateUI();
        initTurtleRenderers();
    } catch (e) {
        console.error('加载游戏状态失败:', e);
        showToast('加载失败，请刷新重试');
    }
}

function updateUI() {
    if (!gameState) return;
    
    // 更新顶部状态
    document.getElementById('coins').textContent = gameState.coins;
    const seasonNames = {spring: '春季', summer: '夏季', autumn: '秋季', winter: '冬季'};
    document.getElementById('day-info').textContent = `第 ${gameState.day} 天 · ${seasonNames[gameState.season] || gameState.season}`;
    
    // 更新当前选中的龟缸
    if (gameState.tanks.length > 0) {
        const tank = getCurrentTank() || gameState.tanks[0];
        currentTankId = tank.id;
        updateTankHeader(tank);
    }

    // 更新龟缸/乌龟列表
    updateTankTabs();
    updateTurtleList();
    
    // 更新装饰
    if (gameState.tanks.length > 0) {
        const tank = getCurrentTank() || gameState.tanks[0];
        decor = tank.decor || [];
    }
    refreshActionHighlights();
}

// 根据当前选中龟的状态给底部按钮加 .urgent / .nudge 高亮
function refreshActionHighlights() {
    const t = getCurrentTurtle();
    const tank = getCurrentTank();
    const map = { feed: false, clean: false, decor: false, interact: false };
    const urgent = { feed: false, clean: false };
    if (t) {
        if (t.hunger <= 30) urgent.feed = true; else if (t.hunger <= 55) map.feed = true;
        if (t.cleanliness <= 40) urgent.clean = true; else if (t.cleanliness <= 65) map.clean = true;
        if (t.mood <= 40) map.interact = true;
    }
    if (tank) {
        const wq = tank.water_quality || {};
        if ((wq.ammonia||0) >= 1.0 || (wq.clarity||100) < 50) urgent.clean = true;
        const decorCount = (tank.decor||[]).length;
        if (decorCount === 0) map.decor = true;
    }
    document.querySelectorAll('.action-btn').forEach(btn => {
        const k = btn.dataset.action;
        btn.classList.remove('nudge', 'urgent');
        if (urgent[k]) btn.classList.add('urgent');
        else if (map[k]) btn.classList.add('nudge');
    });
}

function getCurrentTank() {
    if (!gameState || !gameState.tanks) return null;
    return gameState.tanks.find(t => t.id === currentTankId) || null;
}

function updateTankHeader(tank) {
    if (!tank) return;
    document.getElementById('tank-name').textContent = tank.name;
    const waterNames = {deep: '深水缸', middle: '中水位', shallow: '浅水缸', land: '半水陆缸'};
    const water = tank.water_quality || { ph: 7, ammonia: 0, nitrite: 0, clarity: 100 };
    const status = getWaterStatus(water);
    document.getElementById('tank-water').textContent = `${waterNames[tank.water_level] || tank.water_level} · ${status.label}`;

    const panel = document.getElementById('water-panel');
    if (panel) {
        panel.classList.remove('warning', 'danger');
        if (status.level !== 'good') panel.classList.add(status.level);
        document.getElementById('water-ammonia').textContent = Number(water.ammonia || 0).toFixed(2);
        document.getElementById('water-nitrite').textContent = Number(water.nitrite || 0).toFixed(2);
        document.getElementById('water-clarity').textContent = `${water.clarity ?? 100}%`;
        const filterEl = document.getElementById('water-filter');
        if (filterEl) filterEl.textContent = tank.has_filter ? 'ON' : 'OFF';
    }
}

function getWaterStatus(water) {
    const ammonia = Number(water.ammonia || 0);
    const nitrite = Number(water.nitrite || 0);
    const clarity = Number(water.clarity ?? 100);
    if (ammonia >= 1 || nitrite >= 0.8 || clarity < 45) {
        return { level: 'danger', label: '水质危险，建议立刻清洁' };
    }
    if (ammonia >= 0.5 || nitrite >= 0.3 || clarity < 70) {
        return { level: 'warning', label: '水质预警，尽快换水' };
    }
    return { level: 'good', label: `水质良好 · 清澈 ${clarity}%` };
}

function getCurrentTurtle() {
    if (!gameState || !gameState.turtles) return null;
    return gameState.turtles.find(t => t.id === currentTurtleId) || null;
}

function updateTankTabs() {
    const tabs = document.getElementById('tank-tabs');
    if (!tabs || !gameState || !gameState.tanks) return;
    tabs.innerHTML = '';

    const waterNames = {deep: '深水', middle: '中水', shallow: '浅水', land: '半水陆'};
    gameState.tanks.forEach(tank => {
        const inhabitants = gameState.turtles.filter(t => t.tank_id === tank.id);
        const btn = document.createElement('button');
        btn.className = 'tank-tab' + (tank.id === currentTankId ? ' active' : '');
        btn.dataset.tankId = tank.id;
        const water = tank.water_quality || { ammonia: 0, nitrite: 0, clarity: 100 };
        const status = getWaterStatus(water);
        btn.innerHTML = `
            <span class="tank-tab-name">${tank.name}</span>
            <span class="tank-tab-meta">${waterNames[tank.water_level] || tank.water_level} · ${inhabitants.length}龟 · ${tank.has_filter ? '过滤' : '无滤'} · ${status.level === 'good' ? '良好' : status.level === 'warning' ? '预警' : '危险'}</span>
        `;
        btn.addEventListener('click', () => selectTank(tank.id));
        tabs.appendChild(btn);
    });
}

function selectTank(id) {
    currentTankId = id;
    const tank = getCurrentTank();
    if (!tank) return;
    const activeInTank = gameState.turtles.find(t => t.id === currentTurtleId && t.tank_id === id);
    const firstTurtle = gameState.turtles.find(t => t.tank_id === id);
    if (!activeInTank && firstTurtle) currentTurtleId = firstTurtle.id;
    decor = tank.decor || [];
    updateTankHeader(tank);
    updateTankTabs();
    updateTurtleList();
    initTurtleRenderers();
}


function updateTurtleList() {
    const list = document.getElementById('turtle-list');
    list.innerHTML = '';
    if (gameState.turtles.length > 0 && !currentTurtleId) {
        currentTurtleId = gameState.turtles[0].id;
    }
    
    gameState.turtles.forEach((turtle) => {
        const card = document.createElement('div');
        card.className = 'turtle-card' + (turtle.id === currentTurtleId ? ' active' : '');
        card.dataset.turtleId = turtle.id;
        
        const hungerPercent = turtle.hunger;
        const tank = gameState.tanks.find(t => t.id === turtle.tank_id);
        let hungerClass = '';
        if (hungerPercent < 30) hungerClass = 'low';
        else if (hungerPercent < 60) hungerClass = 'medium';
        
        card.innerHTML = `
            <div class="turtle-avatar">${getTurtleAvatarMarkup(turtle.species)}</div>
            <div class="turtle-name">${turtle.name}</div>
            <div class="turtle-tank">${tank ? tank.name : '待分缸'}</div>
            <div class="turtle-hunger">
                <div class="bar"><div class="fill ${hungerClass}" style="width: ${hungerPercent}%"></div></div>
            </div>
        `;
        
        card.addEventListener('click', () => selectTurtle(turtle.id));
        list.appendChild(card);
    });
}

function getTurtleEmoji(species) {
    return '🐢';
}

function getTurtleVisual(species) {
    const visuals = {
        muskTurtle: {
            slug: 'musk', label: '麝香', shell: '#26381f', shellHi: '#40592e', skin: '#486438', stripe: '#f2d16b', rim: '#172415', belly: '#d6c58f',
            scale: 0.88, shape: 'flat', pattern: 'musk', rows: [4, 8, 10, 12, 12, 10, 8, 4], keel: false, spots: false, tail: 'short'
        },
        razorbackTurtle: {
            slug: 'razor', label: '剃刀', shell: '#42311f', shellHi: '#6b5730', skin: '#57462d', stripe: '#d7bd74', rim: '#21170f', belly: '#b99455',
            scale: 0.98, shape: 'tall', pattern: 'keel', rows: [4, 8, 10, 10, 10, 8, 6], keel: true, spots: false, tail: 'short'
        },
        chinesePondTurtle: {
            slug: 'pond', label: '草龟', shell: '#344f26', shellHi: '#5d7f36', skin: '#536f38', stripe: '#c6d27a', rim: '#24351a', belly: '#d8c36a',
            scale: 1.02, shape: 'oval', pattern: 'scutes', rows: [4, 8, 12, 12, 12, 10, 6], keel: false, spots: true, tail: 'long'
        },
        yellowPondTurtle: {
            slug: 'yellowpond', label: '黄喉', shell: '#6a5a28', shellHi: '#9a8138', skin: '#707b3b', stripe: '#ffd75f', rim: '#3f351a', belly: '#f5d56a',
            scale: 1.0, shape: 'oval', pattern: 'gold', rows: [4, 8, 12, 12, 12, 10, 6], keel: false, spots: false, tail: 'long'
        },
        yellowMarginTurtle: {
            slug: 'margin', label: '黄缘', shell: '#5b3420', shellHi: '#8a4d28', skin: '#715033', stripe: '#ffbf4b', rim: '#f0a531', belly: '#e5b65e',
            scale: 1.04, shape: 'box', pattern: 'margin', rows: [2, 6, 10, 12, 12, 10, 6, 2], keel: true, spots: false, tail: 'short'
        },
    };
    return visuals[species] || visuals.chinesePondTurtle;
}

// M1 像素龟 sprite 列表（程序化生成，详见 scripts/gen_pixel_turtles.py）
const SPRITE_SPECIES = new Set([
    'muskTurtle','razorbackTurtle','loggerheadMuskTurtle','yellowMarginTurtle',
    'chinesePondTurtle','yellowPondTurtle','chineseStripeTurtle','redEaredSlider'
]);

function getTurtleAvatarMarkup(species) {
    if (SPRITE_SPECIES.has(species)) {
        return `<img class="mini-turtle-sprite" src="sprites/${species}.png" alt="${species}" loading="lazy">`;
    }
    const v = getTurtleVisual(species);
    return `<span class="mini-turtle mini-turtle-${v.slug}" title="${v.label}"><span class="mini-shell"></span><span class="mini-head"></span></span>`;
}

function selectTurtle(id) {
    const sameTurtle = currentTurtleId === id;
    currentTurtleId = id;
    const turtle = getCurrentTurtle();
    if (turtle && turtle.tank_id) {
        currentTankId = turtle.tank_id;
        const tank = getCurrentTank();
        if (tank) {
            decor = tank.decor || [];
            updateTankHeader(tank);
        }
    }
    document.querySelectorAll('.turtle-card').forEach(c => c.classList.remove('active'));
    const card = document.querySelector(`[data-turtle-id="${id}"]`);
    if (card) card.classList.add('active');
    initTurtleRenderers();
    refreshActionHighlights();
    // 重复点击同一只龟 → 开详情（移动端不用双击）
    if (sameTurtle) {
        showTurtleDetail(id);
    }
}

// ============ 渲染系统 ============
function initTurtleRenderers() {
    turtles = gameState.turtles
        .filter(t => !currentTankId || !t.tank_id || t.tank_id === currentTankId)
        .map(t => ({
        ...t,
        x: Math.random() * 0.6 + 0.2,
        y: 0.7,
        vx: (Math.random() - 0.5) * 0.3,
        vy: 0,
        direction: Math.random() > 0.5 ? 1 : -1,
        animFrame: 0,
        swimPhase: Math.random() * Math.PI * 2,
        blinkTimer: Math.random() * 3,
    }));
}

function gameLoop(timestamp) {
    const dt = (timestamp - lastTime) / 1000;
    lastTime = timestamp;
    
    updateTurtles(dt);
    render();
    
    animationId = requestAnimationFrame(gameLoop);
}

function updateTurtles(dt) {
    turtles.forEach(t => {
        // 游泳动画
        t.swimPhase += dt * 2;
        t.animFrame += dt * 4;
        
        // 移动
        t.x += t.vx * dt * t.direction;
        
        // 边界反弹
        if (t.x < 0.1) { t.x = 0.1; t.direction = 1; }
        if (t.x > 0.9) { t.x = 0.9; t.direction = -1; }
        
        // 随机转向
        if (Math.random() < 0.005) {
            t.direction *= -1;
        }
        
        // 眨眼
        t.blinkTimer -= dt;
        if (t.blinkTimer < 0) t.blinkTimer = 2 + Math.random() * 4;
    });
}

function render() {
    const w = canvas.width;
    const h = canvas.height;
    
    ctx.clearRect(0, 0, w, h);
    
    // 绘制水面背景
    drawWater(w, h);
    drawWaterQualityOverlay(w, h);
    
    // 绘制造景
    drawDecor(w, h);
    
    // 绘制乌龟
    turtles.forEach(t => drawTurtle(t, w, h));
    
    // 绘制水面波纹
    drawRipples(w, h);
}

function drawWater(w, h) {
    // 水面渐变已在CSS中设置，这里可以加一些水下效果
    // 绘制水底沙地
    ctx.fillStyle = '#c4a35a';
    ctx.fillRect(0, h * 0.85, w, h * 0.15);
    
    // 沙地纹理：用确定性散点，避免每帧随机导致画面闪烁。
    ctx.fillStyle = '#b8934a';
    for (let i = 0; i < 28; i++) {
        const x = ((i * 73) % 101) / 101 * w;
        const y = h * 0.85 + (((i * 37) % 89) / 89) * h * 0.15;
        const r = 1 + (i % 3) * 0.8;
        ctx.beginPath();
        ctx.arc(x, y, r, 0, Math.PI * 2);
        ctx.fill();
    }
    
    // 水面光效
    const gradient = ctx.createLinearGradient(0, 0, 0, h * 0.3);
    gradient.addColorStop(0, 'rgba(255,255,255,0.1)');
    gradient.addColorStop(1, 'rgba(255,255,255,0)');
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, w, h * 0.3);
}

function drawWaterQualityOverlay(w, h) {
    const tank = getCurrentTank();
    if (!tank || !tank.water_quality) return;
    const water = tank.water_quality;
    const ammonia = Number(water.ammonia || 0);
    const nitrite = Number(water.nitrite || 0);
    const clarity = Number(water.clarity ?? 100);
    const murky = Math.max(0, Math.min(0.38, (100 - clarity) / 180 + ammonia * 0.045 + nitrite * 0.035));
    if (murky <= 0.02) return;

    ctx.fillStyle = `rgba(96, 74, 30, ${murky})`;
    ctx.fillRect(0, 0, w, h);
    ctx.fillStyle = `rgba(244, 222, 150, ${murky * 0.65})`;
    for (let i = 0; i < 18; i++) {
        const x = ((i * 47 + Math.floor(Date.now() / 160)) % 101) / 101 * w;
        const y = h * (0.18 + (((i * 29) % 79) / 79) * 0.68);
        ctx.beginPath();
        ctx.arc(x, y, 1.2 + (i % 4) * 0.7, 0, Math.PI * 2);
        ctx.fill();
    }
}

function drawDecor(w, h) {
    decor.forEach(d => {
        const x = d.x * w;
        const y = d.y * h;
        
        ctx.save();
        ctx.translate(x, y);
        ctx.rotate(d.rotation || 0);
        ctx.scale(d.scale || 1, d.scale || 1);
        
        switch (d.type) {
            case 'wood':
                drawWood(ctx);
                break;
            case 'stone':
                drawStone(ctx);
                break;
            case 'plant':
                drawPlant(ctx);
                break;
        }
        
        ctx.restore();
    });
}

function drawWood(ctx) {
    ctx.fillStyle = '#8B4513';
    ctx.beginPath();
    ctx.ellipse(0, 0, 30, 15, 0, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = '#654321';
    ctx.beginPath();
    ctx.ellipse(0, -5, 25, 10, 0, 0, Math.PI * 2);
    ctx.fill();
}

function drawStone(ctx) {
    ctx.fillStyle = '#808080';
    ctx.beginPath();
    ctx.ellipse(0, 0, 20, 15, 0, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = '#696969';
    ctx.beginPath();
    ctx.ellipse(-5, -3, 12, 10, 0, 0, Math.PI * 2);
    ctx.fill();
}

function drawPlant(ctx) {
    ctx.strokeStyle = '#228B22';
    ctx.lineWidth = 2;
    for (let i = 0; i < 5; i++) {
        const angle = -Math.PI / 2 + (i - 2) * 0.3;
        ctx.beginPath();
        ctx.moveTo(0, 0);
        ctx.quadraticCurveTo(
            Math.cos(angle) * 15,
            Math.sin(angle) * 15 - 20,
            Math.cos(angle) * 25,
            Math.sin(angle) * 25 - 40
        );
        ctx.stroke();
    }
}

function drawTurtle(t, w, h) {
    const x = t.x * w;
    const y = h * 0.6 + Math.sin(t.swimPhase) * 10;
    const dir = t.direction;
    const v = getTurtleVisual(t.species);
    const px = Math.max(3, Math.min(5, Math.floor(Math.min(w, h) / 105))) * v.scale;

    ctx.save();
    ctx.translate(x, y);
    ctx.scale(dir, 1);

    // 身体上下浮动动画
    const bob = Math.sin(t.swimPhase * 2) * 3;
    ctx.translate(0, bob);
    ctx.imageSmoothingEnabled = false;

    drawPixelTurtleSprite(ctx, v, px, t.animFrame, t.blinkTimer);

    // 名字小标签：多龟同缸时能快速分辨是哪只。
    ctx.scale(dir, 1);
    ctx.font = '10px system-ui, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillStyle = 'rgba(15, 23, 42, 0.45)';
    const label = t.name || v.label;
    const labelW = Math.min(58, ctx.measureText(label).width + 12);
    ctx.fillRect(-labelW / 2, 24 * px / 4, labelW, 14);
    ctx.fillStyle = '#f8fafc';
    ctx.fillText(label, 0, 24 * px / 4 + 10);

    ctx.restore();
}

function drawPixelTurtleSprite(ctx, v, px, animFrame, blinkTimer) {
    const p = (x, y, w, h, color) => {
        ctx.fillStyle = color;
        ctx.fillRect(Math.round(x * px), Math.round(y * px), Math.ceil(w * px), Math.ceil(h * px));
    };
    const legOffset = Math.sin(animFrame) > 0 ? 1 : 0;
    const rowStart = -Math.floor(v.rows.length / 2);

    // 阴影
    p(-10, 7, 21, 3, 'rgba(0,0,0,0.18)');

    // 尾巴：草龟/黄喉更长，蛋龟更短。
    const tailLen = v.tail === 'long' ? 4 : 2;
    p(-11 - tailLen, -1, tailLen + 2, 2, v.skin);
    p(-12 - tailLen, 0, 2, 1, v.rim);

    // 四肢，做成一格摆动的像素动画。
    p(-8, -6 - legOffset, 4, 3, v.skin);
    p(3, -6 + legOffset, 4, 3, v.skin);
    p(-8, 5 + legOffset, 4, 3, v.skin);
    p(3, 5 - legOffset, 4, 3, v.skin);
    p(-7, -5 - legOffset, 2, 1, v.stripe);
    p(4, 6 - legOffset, 2, 1, v.stripe);

    // 头像：黄喉/黄缘会有明显黄喉，麝香有头侧线。
    p(8, -4, 5, 7, v.skin);
    p(12, -2, 3, 3, v.skin);
    p(10, 2, 4, 2, v.belly);
    if (v.pattern === 'musk' || v.pattern === 'gold' || v.pattern === 'margin') {
        p(10, -3, 5, 1, v.stripe);
    }
    if (blinkTimer > 0.12) {
        p(13, -3, 1, 1, '#020617');
        p(14, -3, 1, 1, '#f8fafc');
    } else {
        p(13, -2, 2, 1, '#020617');
    }

    // 腹甲底色，边缘露出一点像素色块。
    p(-7, -5, 15, 11, v.belly);

    // 龟壳主体：用行宽数组拼出来，保证是像素块而不是通用圆形。
    v.rows.forEach((width, i) => {
        const y = rowStart + i;
        const x = -Math.floor(width / 2);
        p(x, y, width, 1, v.shell);
        if (width > 6) p(x + 1, y, width - 2, 1, v.shellHi);
    });

    // 外缘/物种特征。
    v.rows.forEach((width, i) => {
        const y = rowStart + i;
        const x = -Math.floor(width / 2);
        p(x, y, 1, 1, v.rim);
        p(x + width - 1, y, 1, 1, v.rim);
    });

    // 壳片纹路：不同龟种差异化。
    if (v.pattern === 'keel' || v.keel) {
        for (let y = rowStart + 1; y < rowStart + v.rows.length - 1; y++) {
            p(-1, y, 2, 1, v.stripe);
        }
        p(-3, rowStart + 3, 6, 1, v.rim);
    }
    if (v.pattern === 'scutes' || v.pattern === 'gold') {
        for (let y = rowStart + 2; y < rowStart + v.rows.length - 1; y += 2) {
            p(-4, y, 8, 1, v.rim);
        }
        p(-1, rowStart + 1, 1, v.rows.length - 2, v.rim);
    }
    if (v.pattern === 'musk') {
        p(-4, rowStart + 2, 2, 1, v.stripe);
        p(2, rowStart + 3, 2, 1, v.stripe);
        p(-2, rowStart + 5, 4, 1, v.rim);
    }
    if (v.pattern === 'margin') {
        v.rows.forEach((width, i) => {
            const y = rowStart + i;
            const x = -Math.floor(width / 2);
            if (width >= 6) p(x + 1, y, width - 2, 1, i === 0 || i === v.rows.length - 1 ? v.stripe : v.shellHi);
            p(x, y, width, 1, i === 0 || i === v.rows.length - 1 ? v.rim : 'rgba(0,0,0,0)');
        });
        p(-1, rowStart + 1, 2, v.rows.length - 2, v.stripe);
    }
    if (v.spots) {
        p(-5, rowStart + 2, 1, 1, v.stripe);
        p(4, rowStart + 4, 1, 1, v.stripe);
        p(0, rowStart + 5, 1, 1, v.stripe);
    }
}
function drawRipples(w, h) {
    // 简单的气泡效果
    ctx.fillStyle = 'rgba(255,255,255,0.15)';
    for (let i = 0; i < 5; i++) {
        const bx = (Math.sin(Date.now() / 2000 + i * 1.5) * 0.5 + 0.5) * w;
        const by = h * 0.3 + Math.sin(Date.now() / 1500 + i) * h * 0.2;
        const br = 2 + Math.sin(Date.now() / 1000 + i) * 1;
        ctx.beginPath();
        ctx.arc(bx, by, br, 0, Math.PI * 2);
        ctx.fill();
    }
}

// ============ 事件绑定 ============
function bindEvents() {
    // 底部操作按钮
    document.querySelectorAll('.action-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const action = btn.dataset.action;
            handleAction(action);
        });
    });
    
    // 下一天按钮
    document.getElementById('advance-day-btn').addEventListener('click', advanceDay);
    const manageBtn = document.getElementById('tank-manage-btn');
    if (manageBtn) manageBtn.addEventListener('click', showTankManageModal);
    
    // 关闭弹窗
    document.querySelectorAll('.close-btn').forEach(btn => {
        btn.addEventListener('click', closeAllModals);
    });
    
    // 点击弹窗背景关闭
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) closeAllModals();
        });
    });

    // 龟缸布景拖拽：鼠标/触屏都支持。
    canvas.addEventListener('mousedown', startDecorDrag);
    canvas.addEventListener('mousemove', moveDecorDrag);
    window.addEventListener('mouseup', endDecorDrag);
    canvas.addEventListener('touchstart', startDecorDrag, { passive: false });
    canvas.addEventListener('touchmove', moveDecorDrag, { passive: false });
    window.addEventListener('touchend', endDecorDrag);
}

function handleAction(action) {
    if (!currentTurtleId) {
        showToast('请先选择一只乌龟');
        return;
    }
    
    switch (action) {
        case 'feed':
            showFeedModal();
            break;
        case 'clean':
            showMaintenanceModal();
            break;
        case 'decor':
            showDecorModal();
            break;
        case 'collection':
            showCollection();
            break;
        case 'interact':
            interactWithTurtle();
            break;
    }
}

async function showFeedModal() {
    const modal = document.getElementById('feed-modal');
    const list = document.getElementById('food-list');
    list.innerHTML = '';
    
    const foods = gameState.inventory.filter(i => i.type === 'food' && i.count > 0);
    
    if (foods.length === 0) {
        list.innerHTML = '<p style="text-align:center;color:#aaa;">背包里没有食物了</p>';
    } else {
        foods.forEach(food => {
            const item = document.createElement('div');
            item.className = 'food-item';
            item.innerHTML = `
                <span class="food-icon">${food.icon}</span>
                <div class="food-info">
                    <div class="food-name">${food.name}</div>
                    <div class="food-count">剩余 ${food.count} 个</div>
                </div>
            `;
            item.addEventListener('click', () => feedTurtle(food.id));
            list.appendChild(item);
        });
    }
    
    modal.classList.remove('hidden');
}

async function feedTurtle(foodId) {
    try {
        const res = await fetch(`${API_BASE}/api/feed`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: 'default',
                turtle_id: currentTurtleId,
                food_id: foodId
            })
        });
        
        if (res.ok) {
            showToast('喂食成功！');
            closeAllModals();
            await loadGameState();
        } else {
            showToast('喂食失败');
        }
    } catch (e) {
        showToast('网络错误');
    }
}

async function showDecorModal() {
    if (!currentTankId) {
        showToast('没有可布置的龟缸');
        return;
    }

    const modal = document.getElementById('decor-modal');
    const list = document.getElementById('decor-list');
    list.innerHTML = '';

    const decorOptions = [
        { type: 'wood', name: '沉木', icon: '🪵', desc: '适合麝香龟躲藏，缸里更有层次。' },
        { type: 'stone', name: '晒台石', icon: '🪨', desc: '给草龟歇脚，也能打破空缸感。' },
        { type: 'plant', name: '水草丛', icon: '🌿', desc: '增加安全感，画面更鲜活。' },
    ];

    decorOptions.forEach(option => {
        const item = document.createElement('div');
        item.className = 'decor-item';
        item.innerHTML = `
            <span class="decor-icon">${option.icon}</span>
            <div class="decor-info">
                <div class="decor-name">${option.name}</div>
                <div class="decor-desc">${option.desc}</div>
            </div>
        `;
        item.addEventListener('click', () => addDecor(option.type));
        list.appendChild(item);
    });

    modal.classList.remove('hidden');
}

async function addDecor(type) {
    const newDecor = {
        id: `decor_${Date.now()}`,
        type,
        x: 0.25 + Math.random() * 0.5,
        y: type === 'plant' ? 0.82 : 0.78,
        rotation: (Math.random() - 0.5) * 0.25,
        scale: type === 'plant' ? 0.9 : 1,
    };

    try {
        const res = await fetch(`${API_BASE}/api/add-decor`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: 'default',
                tank_id: currentTankId,
                decor: newDecor,
            })
        });

        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        decor.push(data.decor || newDecor);
        showToast('布景已放入龟缸，拖动可调整位置 +20龟币');
        closeAllModals();
        await loadGameState();
    } catch (e) {
        console.error(e);
        showToast('布景失败');
    }
}

async function saveDecorPosition(decorItem) {
    if (!decorItem || !decorItem.id) return;
    try {
        await fetch(`${API_BASE}/api/move-decor`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: 'default',
                decor_id: decorItem.id,
                x: decorItem.x,
                y: decorItem.y,
            })
        });
    } catch (e) {
        console.warn('保存布景位置失败', e);
    }
}

function getCanvasPoint(e) {
    const rect = canvas.getBoundingClientRect();
    const point = e.touches ? e.touches[0] : e;
    return {
        x: (point.clientX - rect.left) / rect.width,
        y: (point.clientY - rect.top) / rect.height,
    };
}

function findDecorAt(point) {
    for (let i = decor.length - 1; i >= 0; i--) {
        const d = decor[i];
        const hitW = d.type === 'plant' ? 0.10 : 0.14;
        const hitH = d.type === 'plant' ? 0.16 : 0.10;
        if (Math.abs(point.x - d.x) <= hitW && Math.abs(point.y - d.y) <= hitH) {
            return d;
        }
    }
    return null;
}

function startDecorDrag(e) {
    const point = getCanvasPoint(e);
    const target = findDecorAt(point);
    if (!target) return;
    draggedDecorId = target.id;
    dragDirty = false;
    canvas.classList.add('dragging');
    if (e.cancelable) e.preventDefault();
}

function moveDecorDrag(e) {
    if (!draggedDecorId) return;
    const target = decor.find(d => d.id === draggedDecorId);
    if (!target) return;
    const point = getCanvasPoint(e);
    target.x = Math.max(0.05, Math.min(0.95, point.x));
    target.y = Math.max(0.10, Math.min(0.90, point.y));
    dragDirty = true;
    if (e.cancelable) e.preventDefault();
}

async function endDecorDrag() {
    if (!draggedDecorId) return;
    const target = decor.find(d => d.id === draggedDecorId);
    draggedDecorId = null;
    canvas.classList.remove('dragging');
    if (dragDirty && target) {
        await saveDecorPosition(target);
        showToast('布景位置已保存');
    }
}

function showMaintenanceModal() {
    if (!currentTankId) {
        showToast('没有可维护的龟缸');
        return;
    }
    const tank = getCurrentTank();
    const modal = document.getElementById('maintenance-modal');
    const list = document.getElementById('maintenance-list');
    list.innerHTML = '';

    const options = [
        { mode: 'scoop_waste', icon: '🪣', name: '捞残饵', cost: 0, desc: '日常小维护：清澈度 +18，轻微降低 NH₃/NO₂。' },
        { mode: 'partial_change', icon: '💧', name: '部分换水', cost: 20, desc: '推荐操作：花少量龟币，明显改善水质但保留稳定菌群。' },
        { mode: 'deep_clean', icon: '🧽', name: '深度清洁', cost: 60, desc: '重度污染急救：水质直接回满，龟龟清洁度大幅恢复。' },
        { mode: 'install_filter', icon: '⚙️', name: '安装过滤器', cost: 180, desc: tank && tank.has_filter ? '这个龟缸已经有过滤器了。' : '一次性投资：以后每天水质恶化速度显著降低。', disabled: tank && tank.has_filter },
    ];

    options.forEach(option => {
        const item = document.createElement('div');
        const disabled = option.disabled || gameState.coins < option.cost;
        item.className = 'maintenance-item' + (disabled ? ' disabled' : '');
        item.innerHTML = `
            <span class="maintenance-icon">${option.icon}</span>
            <div class="maintenance-info">
                <div class="maintenance-name">${option.name} <span class="maintenance-cost">${option.cost ? `${option.cost}币` : '免费'}</span></div>
                <div class="maintenance-desc">${option.desc}${gameState.coins < option.cost ? ` 龟币不足，还差 ${option.cost - gameState.coins}` : ''}</div>
            </div>
        `;
        if (!disabled) item.addEventListener('click', () => maintainTank(option.mode));
        list.appendChild(item);
    });

    modal.classList.remove('hidden');
}

async function maintainTank(mode) {
    try {
        const res = await fetch(`${API_BASE}/api/maintain-tank`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: 'default',
                tank_id: currentTankId,
                mode,
            })
        });

        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        closeAllModals();
        await loadGameState();
        showToast(data.message || '维护完成，水质变好了');
    } catch (e) {
        showToast((e.message || '维护失败').trim());
    }
}

async function cleanTank() {
    return maintainTank('deep_clean');
}

async function interactWithTurtle() {
    try {
        const res = await fetch(`${API_BASE}/api/interact`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: 'default',
                turtle_id: currentTurtleId,
                action: 'pet'
            })
        });
        
        if (res.ok) {
            showToast('你摸了摸乌龟，它看起来很开心');
            await loadGameState();
        }
    } catch (e) {
        showToast('互动失败');
    }
}


function showTankManageModal() {
    const modal = document.getElementById('tank-manage-modal');
    const createList = document.getElementById('create-tank-list');
    const moveList = document.getElementById('move-turtle-list');;
    if (!modal || !createList || !moveList) return;

    const levels = [
        { level: 'deep', icon: '🌊', name: '深水缸', desc: '蛋龟/麝香/剃刀更适合，默认带过滤器。' },
        { level: 'middle', icon: '💧', name: '中水位缸', desc: '草龟/黄喉等水龟适合，带晒背灯和过滤器。' },
        { level: 'shallow', icon: '🪣', name: '浅水缸', desc: '备用隔离/幼龟观察缸，后续可扩展医疗玩法。' },
        { level: 'land', icon: '🏝️', name: '半水陆缸', desc: '黄缘等半水龟适合，默认无过滤器但有 UVB。' },
    ];

    createList.innerHTML = '<h4 class="section-title">➕ 新建龟缸</h4>';
    levels.forEach(opt => {
        const item = document.createElement('div');
        const disabled = gameState.coins < 120;
        item.className = 'tank-manage-item' + (disabled ? ' disabled' : '');
        item.innerHTML = `
            <span class="tank-manage-icon">${opt.icon}</span>
            <div class="tank-manage-info">
                <div class="tank-manage-name">${opt.name} <span class="maintenance-cost">120币</span></div>
                <div class="tank-manage-desc">${opt.desc}${disabled ? ` 龟币不足，还差 ${120 - gameState.coins}` : ''}</div>
            </div>
        `;
        if (!disabled) item.addEventListener('click', () => createTank(opt));
        createList.appendChild(item);
    });

    moveList.innerHTML = '';
    if (!gameState.turtles.length) {
        moveList.innerHTML = '<p class="empty-tip">还没有可搬家的乌龟</p>';
    } else {
        gameState.turtles.forEach(turtle => {
            const sp = speciesMeta(turtle.species);
            const currentTank = gameState.tanks.find(t => t.id === turtle.tank_id);
            const row = document.createElement('div');
            row.className = 'move-row';
            row.innerHTML = `
                <div class="move-row-title">${getTurtleEmoji(turtle.species)} ${turtle.name} <span>${sp.name} · 适合${habitatName(sp.habitat_type)}</span></div>
                <div class="move-targets"></div>
                <div class="move-current">当前：${currentTank ? currentTank.name : '未入住'}</div>
            `;
            const targets = row.querySelector('.move-targets');
            gameState.tanks.forEach(tank => {
                const ok = tank.water_level === sp.habitat_type;
                const same = tank.id === turtle.tank_id;
                const btn = document.createElement('button');
                btn.className = 'move-target-btn' + (same ? ' active' : '') + (!ok ? ' disabled' : '');
                btn.textContent = same ? `${tank.name} ✓` : tank.name;
                btn.disabled = same || !ok;
                btn.title = ok ? '可搬入' : `不适合，建议${habitatName(sp.habitat_type)}`;
                if (!same && ok) btn.addEventListener('click', () => moveTurtleToTank(turtle.id, tank.id));
                targets.appendChild(btn);
            });
            moveList.appendChild(row);
        });
    }

    modal.classList.remove('hidden');
}

async function createTank(opt) {
    const name = prompt(`给新的${opt.name}取个名字`, `新的${opt.name}`);
    if (name === null) return;
    try {
        const res = await fetch(`${API_BASE}/api/create-tank`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: 'default',
                water_level: opt.level,
                name: name.trim() || `新的${opt.name}`,
            })
        });
        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        closeAllModals();
        await loadGameState();
        currentTankId = data.tank_id || currentTankId;
        selectTank(currentTankId);
        showToast(`${opt.name}已建好，花费 ${data.cost || 120} 龟币`);
    } catch (e) {
        showToast((e.message || '新建龟缸失败').trim());
    }
}

async function moveTurtleToTank(turtleId, tankId) {
    try {
        const res = await fetch(`${API_BASE}/api/move-turtle`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ player_id: 'default', turtle_id: turtleId, tank_id: tankId })
        });
        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        closeAllModals();
        await loadGameState();
        currentTurtleId = turtleId;
        currentTankId = data.tank_id || tankId;
        selectTank(currentTankId);
        showToast(data.message || '搬家完成');
    } catch (e) {
        showToast((e.message || '搬家失败').trim());
    }
}

function speciesMeta(speciesId) {
    const meta = {
        muskTurtle: { name: '麝香龟', habitat_type: 'deep' },
        razorbackTurtle: { name: '剃刀龟', habitat_type: 'deep' },
        chinesePondTurtle: { name: '中华草龟', habitat_type: 'middle' },
        yellowPondTurtle: { name: '黄喉拟水龟', habitat_type: 'middle' },
        yellowMarginTurtle: { name: '黄缘闭壳龟', habitat_type: 'land' },
    };
    return meta[speciesId] || { name: '未知龟种', habitat_type: 'middle' };
}

async function advanceDay() {
    try {
        const res = await fetch(`${API_BASE}/api/advance-day`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ player_id: 'default' })
        });
        
        if (res.ok) {
            const data = await res.json();
            await loadGameState();
            const tank = getCurrentTank();
            const status = tank ? getWaterStatus(tank.water_quality || {}) : { label: '已推进' };
            const seasonNames = {spring: '春季', summer: '夏季', autumn: '秋季', winter: '冬季'};
            // M4 日常龟币收入提示
            const incomeText = data.income > 0 ? ` 💰+${data.income}` : '';
            showToast(`第 ${data.day} 天 · ${seasonNames[data.season] || data.season} · ${status.label}${incomeText}`);
            // M5 季节事件提示（延迟 1.4s 避免 toast 重叠）
            if (data.season_event && data.season_event.text) {
                setTimeout(() => {
                    showToast(`${data.season_event.icon || '🌿'} ${data.season_event.text}`);
                }, 1400);
            }
        }
    } catch (e) {
        showToast('推进天数失败');
    }
}

async function showCollection() {
    const modal = document.getElementById('collection-modal');
    const list = document.getElementById('species-list');
    list.innerHTML = '';
    
    try {
        const res = await fetch(`${API_BASE}/api/species`);
        const species = await res.json();
        
        species.forEach(s => {
            const isUnlocked = gameState.unlocked_species.includes(s.id);
            const card = document.createElement('div');
            card.className = 'species-card' + (isUnlocked ? '' : ' locked');
            const ownedCount = gameState.turtles.filter(t => t.species === s.id).length;
            const canBuy = s.unlock_cost > 0;
            card.innerHTML = `
                <span class="species-emoji">${isUnlocked ? getTurtleAvatarMarkup(s.id) : '🔒'}</span>
                <div class="species-info">
                    <div class="species-name">${s.name}${ownedCount ? ` ×${ownedCount}` : ''}</div>
                    <div class="species-category">${s.category} · ${'⭐'.repeat(s.difficulty)} · ${habitatName(s.habitat_type)}</div>
                    <div class="species-desc">${s.description}</div>
                    ${!isUnlocked ? `<div class="species-unlock">🔓 ${s.unlock_condition}${s.unlock_cost > 0 ? ` · ${s.unlock_cost} 龟币` : ''}</div>` : ''}
                    ${canBuy ? `<button class="buy-species-btn" data-species-id="${s.id}">${isUnlocked ? '再领养一只' : `购买领养 · ${s.unlock_cost}币`}</button>` : '<div class="species-owned">已入住初始龟缸</div>'}
                </div>
            `;
            const buyBtn = card.querySelector('.buy-species-btn');
            if (buyBtn) buyBtn.addEventListener('click', () => buySpecies(s));
            list.appendChild(card);
        });
        
        modal.classList.remove('hidden');
    } catch (e) {
        showToast('加载图鉴失败');
    }
}

function habitatName(habitatType) {
    const names = { deep: '深水缸', middle: '中水位', shallow: '浅水缸', land: '半水陆缸' };
    return names[habitatType] || habitatType;
}

async function buySpecies(species) {
    if (!species || !species.id) return;
    if (gameState.coins < species.unlock_cost) {
        showToast(`龟币不足，还差 ${species.unlock_cost - gameState.coins}`);
        return;
    }
    const name = prompt(`给新来的${species.name}取个名字`, defaultNameForSpecies(species.id));
    if (name === null) return;
    try {
        const res = await fetch(`${API_BASE}/api/buy-species`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: 'default',
                species_id: species.id,
                name: name.trim() || defaultNameForSpecies(species.id),
            })
        });
        if (!res.ok) throw new Error(await res.text());
        const data = await res.json();
        closeAllModals();
        await loadGameState();
        currentTankId = data.tank_id || currentTankId;
        currentTurtleId = data.turtle_id || currentTurtleId;
        selectTank(currentTankId);
        showToast(`${species.name}已入住新龟缸！`);
    } catch (e) {
        console.error(e);
        showToast(e.message || '领养失败');
    }
}

function defaultNameForSpecies(speciesId) {
    const names = {
        razorbackTurtle: '小剃刀',
        yellowPondTurtle: '小黄喉',
        yellowMarginTurtle: '小黄缘',
    };
    return names[speciesId] || '新朋友';
}

function closeAllModals() {
    document.querySelectorAll('.modal').forEach(m => m.classList.add('hidden'));
}

function showToast(msg) {
    const toast = document.getElementById('toast');
    toast.textContent = msg;
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 2000);
}

// ============ 龟龟详情面板 + 水质 sparkline ============
async function showTurtleDetail(turtleId) {
    try {
        const res = await fetch(`/api/turtle?id=${encodeURIComponent(turtleId)}&player_id=${PLAYER_ID || 'default'}`);
        if (!res.ok) { showToast('加载详情失败'); return; }
        const data = await res.json();
        renderTurtleDetail(data);
        document.getElementById('turtle-detail-modal').classList.remove('hidden');
    } catch (e) {
        showToast('详情请求失败');
    }
}

function renderTurtleDetail(data) {
    const t = data.turtle || {};
    const sp = data.species_info || {};
    const tank = data.tank || null;
    const sugg = data.suggestions || [];
    const age = data.age_days || 0;

    document.getElementById('detail-name').textContent = `${t.name} · ${sp.name || t.species}`;

    // 建议
    const sBox = document.getElementById('detail-suggestions');
    sBox.innerHTML = sugg.map(s => `<div class="suggestion ${s.level}"><span class="s-icon">${s.icon}</span>${escapeHtml(s.text)}</div>`).join('');

    // 状态条
    const statsBox = document.getElementById('detail-stats');
    const bar = (label, val, max = 100, cls = '') => `
        <div class="stat-row ${cls}">
            <span class="stat-label">${label}</span>
            <div class="stat-bar"><div class="stat-fill" style="width:${Math.max(0, Math.min(100, val * 100 / max))}%"></div></div>
            <span class="stat-num">${val}</span>
        </div>`;
    const h = t.health || {};
    statsBox.innerHTML = `
        <div class="detail-meta">
            <div><b>性别</b> ${t.gender || '?'}</div>
            <div><b>性格</b> ${t.personality || '未知'}</div>
            <div><b>体重</b> ${(t.weight || 0).toFixed(1)} g</div>
            <div><b>年龄</b> ${age} 天</div>
            <div><b>所在</b> ${tank ? tank.name + ' / ' + tank.water_name : '待分缸'}</div>
            <div><b>偏好</b> ${sp.habitat_type ? habitatName(sp.habitat_type) : '-'}</div>
        </div>
        <div class="detail-bars">
            ${bar('饥饿度', t.hunger)}
            ${bar('清洁度', t.cleanliness)}
            ${bar('心情',   t.mood)}
            ${bar('活力',   h.vitality)}
            ${bar('食欲',   h.appetite)}
            ${bar('皮肤',   h.skin)}
            ${bar('壳',     h.shell)}
            ${bar('亲密度', t.intimacy)}
        </div>
        <p class="species-desc">${sp.description || ''}</p>
    `;

    drawWaterSparkline(data.water_history || []);
}

function drawWaterSparkline(history) {
    const cvs = document.getElementById('water-sparkline');
    const legend = document.getElementById('water-legend');
    if (!cvs) return;
    const ctx = cvs.getContext('2d');
    const W = cvs.width, H = cvs.height;
    ctx.clearRect(0, 0, W, H);
    // 背景
    ctx.fillStyle = '#101a20';
    ctx.fillRect(0, 0, W, H);
    if (!history.length) {
        ctx.fillStyle = '#aaa';
        ctx.font = '13px sans-serif';
        ctx.fillText('暂无水质历史（推进几天后会出现）', 16, H / 2);
        legend.innerHTML = '';
        return;
    }
    const padL = 28, padR = 12, padT = 12, padB = 22;
    const innerW = W - padL - padR;
    const innerH = H - padT - padB;

    // 轴线
    ctx.strokeStyle = '#26424c';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(padL, padT); ctx.lineTo(padL, H - padB); ctx.lineTo(W - padR, H - padB); ctx.stroke();

    const n = history.length;
    const xAt = i => padL + (n === 1 ? innerW / 2 : (i * innerW) / (n - 1));

    const drawLine = (vals, color, max) => {
        ctx.strokeStyle = color;
        ctx.lineWidth = 2;
        ctx.beginPath();
        vals.forEach((v, i) => {
            const x = xAt(i);
            const y = padT + innerH * (1 - Math.max(0, Math.min(1, v / max)));
            if (i === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y);
        });
        ctx.stroke();
        ctx.fillStyle = color;
        vals.forEach((v, i) => {
            const x = xAt(i);
            const y = padT + innerH * (1 - Math.max(0, Math.min(1, v / max)));
            ctx.beginPath(); ctx.arc(x, y, 2.5, 0, Math.PI * 2); ctx.fill();
        });
    };

    drawLine(history.map(h => h.ammonia), '#ff7676', 2.0);   // NH3 上限参考 2.0
    drawLine(history.map(h => h.nitrite), '#ffd166', 1.0);   // NO2 上限 1.0
    drawLine(history.map(h => (h.clarity || 0) / 100), '#76e0ff', 1.0); // 清澈 0~1

    // x 轴标签：起始/结束日
    ctx.fillStyle = '#7aa';
    ctx.font = '11px sans-serif';
    ctx.fillText('D' + history[0].day, padL - 8, H - 6);
    ctx.fillText('D' + history[history.length - 1].day, W - padR - 18, H - 6);

    legend.innerHTML = `
        <span class="lg"><i style="background:#ff7676"></i>NH₃</span>
        <span class="lg"><i style="background:#ffd166"></i>NO₂</span>
        <span class="lg"><i style="background:#76e0ff"></i>清澈%</span>
    `;
}

function escapeHtml(s) {
    return String(s||'').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}

const PLAYER_ID = 'default';
