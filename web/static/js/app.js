// ===== Mayfly WebShell 前端应用 =====

// 全局状态
const state = {
    token: localStorage.getItem('mayfly_token'),
    user: localStorage.getItem('mayfly_user') || 'admin',
    terminals: [],   // { id, name, term, fitAddon, ws, active }
    activeId: null,
    termCounter: 0,
    settings: {
        fontSize: 14,
        theme: 'dark',
        cursorBlink: true,
    },
};

// 主题配置
const themes = {
    dark: {
        background: '#0d1117',
        foreground: '#e6edf3',
        cursor: '#e6edf3',
        selectionBackground: 'rgba(99, 102, 241, 0.3)',
        black: '#484f58',
        red: '#ff7b72',
        green: '#3fb950',
        yellow: '#d29922',
        blue: '#58a6ff',
        magenta: '#bc8cff',
        cyan: '#39c5cf',
        white: '#b1bac4',
        brightBlack: '#6e7681',
        brightRed: '#ffa198',
        brightGreen: '#56d364',
        brightYellow: '#e3b341',
        brightBlue: '#79c0ff',
        brightMagenta: '#d2a8ff',
        brightCyan: '#56d4dd',
        brightWhite: '#f0f6fc',
    },
    light: {
        background: '#ffffff',
        foreground: '#24292f',
        cursor: '#24292f',
        selectionBackground: 'rgba(99, 102, 241, 0.2)',
        black: '#24292f',
        red: '#cf222e',
        green: '#1a7f37',
        yellow: '#9a6700',
        blue: '#0969da',
        magenta: '#8250df',
        cyan: '#1b7c83',
        white: '#57606a',
        brightBlack: '#6e7681',
        brightRed: '#a40e26',
        brightGreen: '#2da44e',
        brightYellow: '#bf8700',
        brightBlue: '#218bff',
        brightMagenta: '#a475f9',
        brightCyan: '#3192aa',
        brightWhite: '#8c959f',
    },
    solarized: {
        background: '#002b36',
        foreground: '#839496',
        cursor: '#93a1a1',
        selectionBackground: 'rgba(131, 148, 150, 0.3)',
        black: '#073642',
        red: '#dc322f',
        green: '#859900',
        yellow: '#b58900',
        blue: '#268bd2',
        magenta: '#d33682',
        cyan: '#2aa198',
        white: '#eee8d5',
        brightBlack: '#002b36',
        brightRed: '#cb4b16',
        brightGreen: '#586e75',
        brightYellow: '#657b83',
        brightBlue: '#839496',
        brightMagenta: '#6c71c4',
        brightCyan: '#93a1a1',
        brightWhite: '#fdf6e3',
    },
};

// ===== 初始化 =====
async function init() {
    // 检查认证
    if (!state.token) {
        window.location.href = '/login';
        return;
    }

    // 校验 token 有效性，避免过期 token 导致无限重连
    const valid = await validateToken();
    if (!valid) {
        localStorage.removeItem('mayfly_token');
        localStorage.removeItem('mayfly_user');
        window.location.href = '/login';
        return;
    }

    // 显示用户名
    document.getElementById('currentUser').textContent = state.user;

    // 加载设置
    loadSettings();

    // 绑定事件
    bindEvents();

    // 自动创建第一个终端
    createTerminal();

    // 加载系统信息
    loadSysInfo();
    setInterval(loadSysInfo, 30000);
}

// 校验 token 是否有效（通过受保护的 API 探测）
async function validateToken() {
    try {
        const res = await fetch('/api/sessions', {
            headers: { 'Authorization': 'Bearer ' + state.token },
        });
        // 401 表示 token 无效或过期
        return res.status !== 401;
    } catch (e) {
        // 网络异常时返回 true，交给 WebSocket 自动重连机制兜底
        return true;
    }
}

// ===== 终端管理 =====
function createTerminal() {
    state.termCounter++;
    const termId = 'term-' + state.termCounter;
    const termName = '终端 ' + state.termCounter;

    // 创建终端 DOM
    const container = document.getElementById('terminalContainer');
    const div = document.createElement('div');
    div.className = 'terminal-instance';
    div.id = termId;
    container.appendChild(div);

    // 创建 xterm 实例
    const term = new Terminal({
        fontSize: state.settings.fontSize,
        fontFamily: 'Menlo, Monaco, "DejaVu Sans Mono", Consolas, monospace',
        theme: themes[state.settings.theme],
        cursorBlink: state.settings.cursorBlink,
        allowProposedApi: true,
        scrollback: 10000,
    });

    const fitAddon = new FitAddon.FitAddon();
    const webLinksAddon = new WebLinksAddon.WebLinksAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(webLinksAddon);
    term.open(div);

    // 延迟一帧后 fit
    requestAnimationFrame(() => {
        fitAddon.fit();
    });

    // 终端会话数据
    const termData = {
        id: termId,
        name: termName,
        term: term,
        fitAddon: fitAddon,
        ws: null,
        active: false,
        reconnect: true,       // 是否自动重连
        reconnectAttempts: 0,  // 重连次数
        reconnectTimer: null,  // 重连定时器
    };

    state.terminals.push(termData);

    // 终端输入 -> WebSocket
    term.onData((data) => {
        if (termData.ws && termData.ws.readyState === WebSocket.OPEN) {
            termData.ws.send(JSON.stringify({ type: 'input', data: data }));
        }
    });

    // 终端大小变化 -> 发送 resize
    term.onResize(({ cols, rows }) => {
        if (termData.ws && termData.ws.readyState === WebSocket.OPEN) {
            termData.ws.send(JSON.stringify({ type: 'resize', data: { cols, rows } }));
        }
    });

    // 建立 WebSocket 连接
    connectWS(termData);

    // 切换到新终端
    switchTerminal(termId);

    // 监听容器大小变化
    const resizeObserver = new ResizeObserver(() => {
        if (termData.active) {
            fitAddon.fit();
        }
    });
    resizeObserver.observe(div);
}

// 建立 WebSocket 连接（含自动重连）
function connectWS(termData) {
    const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') +
        location.host + '/api/terminal?token=' + encodeURIComponent(state.token);
    const ws = new WebSocket(wsUrl);
    termData.ws = ws;

    ws.onopen = () => {
        termData.reconnectAttempts = 0;
        if (termData.term) {
            termData.term.write('\x1b[32m[已连接]\x1b[0m\r\n');
        }
        sendResize(termData);
        updateTerminalList();
    };

    ws.onmessage = (event) => {
        if (!termData.term) return;
        try {
            const msg = JSON.parse(event.data);
            switch (msg.type) {
                case 'output':
                    termData.term.write(msg.data);
                    break;
                case 'closed':
                    termData.term.write('\r\n\x1b[33m[终端已关闭]\x1b[0m\r\n');
                    break;
                case 'error':
                    termData.term.write('\r\n\x1b[31m[错误] ' + msg.data + '\x1b[0m\r\n');
                    break;
            }
        } catch (e) {
            termData.term.write(event.data);
        }
    };

    ws.onerror = () => {
        // onerror 之后必然触发 onclose，统一在 onclose 中处理，避免重复提示
    };

    ws.onclose = () => {
        updateTerminalList();
        if (termData.reconnect && termData.term) {
            // 重连前校验 token，若已失效则跳转登录页，避免无限 401 循环
            validateToken().then((valid) => {
                if (!valid) {
                    localStorage.removeItem('mayfly_token');
                    localStorage.removeItem('mayfly_user');
                    window.location.href = '/login';
                    return;
                }
                // 指数退避重连，最大 30 秒
                const delay = Math.min(2000 * Math.pow(2, termData.reconnectAttempts), 30000);
                termData.reconnectAttempts++;
                if (termData.term) {
                    termData.term.write('\r\n\x1b[33m[连接已断开，' + Math.round(delay / 1000) + ' 秒后自动重连...]\x1b[0m\r\n');
                }
                termData.reconnectTimer = setTimeout(() => connectWS(termData), delay);
            });
        }
    };
}

function switchTerminal(termId) {
    state.terminals.forEach((t) => {
        const el = document.getElementById(t.id);
        if (t.id === termId) {
            t.active = true;
            el.classList.add('active');
            document.getElementById('toolbarTitle').textContent = t.name;
            requestAnimationFrame(() => {
                t.fitAddon.fit();
                t.term.focus();
            });
        } else {
            t.active = false;
            el.classList.remove('active');
        }
    });
    state.activeId = termId;
    updateTerminalList();
}

function closeTerminal(termId) {
    const idx = state.terminals.findIndex((t) => t.id === termId);
    if (idx === -1) return;

    const termData = state.terminals[idx];
    // 停止自动重连
    termData.reconnect = false;
    if (termData.reconnectTimer) {
        clearTimeout(termData.reconnectTimer);
        termData.reconnectTimer = null;
    }
    if (termData.ws) {
        termData.ws.close();
    }
    if (termData.term) {
        termData.term.dispose();
        termData.term = null;
    }

    document.getElementById(termId).remove();
    state.terminals.splice(idx, 1);

    if (state.activeId === termId) {
        if (state.terminals.length > 0) {
            switchTerminal(state.terminals[0].id);
        } else {
            state.activeId = null;
            document.getElementById('toolbarTitle').textContent = '';
        }
    }
    updateTerminalList();
}

function updateTerminalList() {
    const list = document.getElementById('terminalList');
    list.innerHTML = '';

    state.terminals.forEach((t) => {
        const tab = document.createElement('div');
        tab.className = 'terminal-tab' + (t.active ? ' active' : '');

        const dot = document.createElement('span');
        dot.className = 'status-dot' + (t.ws && t.ws.readyState === WebSocket.OPEN ? '' : ' inactive');

        const name = document.createElement('span');
        name.className = 'tab-name';
        name.textContent = t.name;

        const closeBtn = document.createElement('button');
        closeBtn.className = 'tab-close';
        closeBtn.innerHTML = '<i class="fas fa-times"></i>';
        closeBtn.onclick = (e) => {
            e.stopPropagation();
            closeTerminal(t.id);
        };

        tab.appendChild(dot);
        tab.appendChild(name);
        tab.appendChild(closeBtn);
        tab.onclick = () => switchTerminal(t.id);

        list.appendChild(tab);
    });
}

function sendResize(termData) {
    if (termData.ws && termData.ws.readyState === WebSocket.OPEN) {
        const cols = termData.term.cols;
        const rows = termData.term.rows;
        termData.ws.send(JSON.stringify({ type: 'resize', data: { cols, rows } }));
    }
}

// ===== 系统信息 =====
function loadSysInfo() {
    // 占位：后续可通过专用 API 获取系统信息
    const hostname = document.getElementById('hostname');
    const cpuInfo = document.getElementById('cpuInfo');
    const memInfo = document.getElementById('memInfo');
    const uptime = document.getElementById('uptime');

    hostname.textContent = 'Ubuntu Server';
    cpuInfo.textContent = '-';
    memInfo.textContent = '-';
    uptime.textContent = '-';
}

// ===== 设置 =====
function loadSettings() {
    const saved = localStorage.getItem('mayfly_settings');
    if (saved) {
        try {
            state.settings = { ...state.settings, ...JSON.parse(saved) };
        } catch (e) {}
    }
}

function saveSettings() {
    localStorage.setItem('mayfly_settings', JSON.stringify(state.settings));
}

function applySettings() {
    state.terminals.forEach((t) => {
        t.term.options.fontSize = state.settings.fontSize;
        t.term.options.theme = themes[state.settings.theme];
        t.term.options.cursorBlink = state.settings.cursorBlink;
        t.fitAddon.fit();
    });
}

// ===== 事件绑定 =====
function bindEvents() {
    // 新建终端
    document.getElementById('newTerminalBtn').onclick = createTerminal;

    // 侧边栏切换
    document.getElementById('sidebarToggle').onclick = () => {
        document.getElementById('sidebar').classList.toggle('collapsed');
        setTimeout(() => {
            const active = state.terminals.find((t) => t.active);
            if (active) active.fitAddon.fit();
        }, 250);
    };

    // 退出登录
    document.getElementById('logoutBtn').onclick = () => {
        if (confirm('确定退出登录？')) {
            localStorage.removeItem('mayfly_token');
            localStorage.removeItem('mayfly_user');
            window.location.href = '/login';
        }
    };

    // 清屏
    document.getElementById('clearBtn').onclick = () => {
        const active = state.terminals.find((t) => t.active);
        if (active) {
            active.term.clear();
        }
    };

    // 全屏
    document.getElementById('fullscreenBtn').onclick = () => {
        const main = document.querySelector('.main-content');
        main.classList.toggle('fullscreen');
        const icon = document.querySelector('#fullscreenBtn i');
        if (main.classList.contains('fullscreen')) {
            icon.className = 'fas fa-compress';
        } else {
            icon.className = 'fas fa-expand';
        }
        setTimeout(() => {
            const active = state.terminals.find((t) => t.active);
            if (active) active.fitAddon.fit();
        }, 100);
    };

    // 设置弹窗
    document.getElementById('settingsBtn').onclick = () => {
        document.getElementById('settingsModal').style.display = 'flex';
        document.getElementById('fontSizeSlider').value = state.settings.fontSize;
        document.getElementById('fontSizeValue').textContent = state.settings.fontSize + 'px';
        document.getElementById('themeSelect').value = state.settings.theme;
        document.getElementById('cursorBlink').checked = state.settings.cursorBlink;
    };

    // 字体大小滑块
    document.getElementById('fontSizeSlider').oninput = (e) => {
        document.getElementById('fontSizeValue').textContent = e.target.value + 'px';
    };

    // 保存设置
    document.getElementById('saveSettingsBtn').onclick = () => {
        state.settings.fontSize = parseInt(document.getElementById('fontSizeSlider').value);
        state.settings.theme = document.getElementById('themeSelect').value;
        state.settings.cursorBlink = document.getElementById('cursorBlink').checked;
        saveSettings();
        applySettings();
        document.getElementById('settingsModal').style.display = 'none';
    };

    // 窗口大小变化时重新 fit
    window.addEventListener('resize', () => {
        const active = state.terminals.find((t) => t.active);
        if (active) active.fitAddon.fit();
    });

    // 键盘快捷键
    document.addEventListener('keydown', (e) => {
        // Ctrl+Shift+T: 新建终端
        if (e.ctrlKey && e.shiftKey && e.key === 'T') {
            e.preventDefault();
            createTerminal();
        }
    });
}

// ===== 启动 =====
document.addEventListener('DOMContentLoaded', init);
