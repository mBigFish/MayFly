// ===== Mayfly WebShell 管理器前端 =====

// 界面状态持久化：记住最后选中的节点、当前功能面板/标签页，刷新后自动恢复
const UI_KEY = 'mayfly_ui_state';
let savedUI = {};
try { savedUI = JSON.parse(localStorage.getItem(UI_KEY) || '{}'); } catch (e) { savedUI = {}; }

const state = {
    token: localStorage.getItem('mayfly_token'),
    user: localStorage.getItem('mayfly_user') || 'admin',
    nodes: [],
    currentNode: null,
    currentTab: savedUI.currentTab || 'file',
    currentView: savedUI.currentView || 'connections', // 'workspace' or 'connections'
    filePath: '',
    editingFile: null,
    terms: {},
    activeTermId: null,
    srvTerms: {},
    activeSrvTermId: null,
    connStatus: {}, // { nodeId: { status: 'ok'|'fail'|'testing'|'untested', message, info } }
    clearedNodes: {}, // 节点列表中已"清空"隐藏的节点 id（不影响连接列表数据）
    connFilter: 'all',
    listeners: [],
    activeListenerId: null,
    payloads: [],
    payloadFilter: 'all',
    listenerPollTimer: null,
    servers: [],
    serverGroups: [],
};

// ===== 工具函数 =====
function debounce(fn, ms) {
    let t;
    return function (...args) {
        clearTimeout(t);
        t = setTimeout(() => fn.apply(this, args), ms);
    };
}

function api(method, path, body) {
    const headers = { 'Authorization': 'Bearer ' + state.token };
    const opts = { method, headers };
    if (body !== undefined) {
        headers['Content-Type'] = 'application/json';
        opts.body = JSON.stringify(body);
    }
    return fetch('/api' + path, opts).then(async (res) => {
        if (res.status === 401) { logout(true); throw new Error('登录已过期'); }
        const data = await res.json().catch(() => ({}));
        return { ok: res.ok, status: res.status, data };
    });
}

function toast(msg, type) {
    let wrap = document.getElementById('toastWrap');
    if (!wrap) {
        wrap = document.createElement('div');
        wrap.className = 'toast-container';
        wrap.id = 'toastWrap';
        document.body.appendChild(wrap);
    }
    const iconMap = { error: 'fa-times-circle', success: 'fa-check-circle', warning: 'fa-exclamation-circle', info: 'fa-info-circle' };
    const t = document.createElement('div');
    t.className = 'toast ' + (type || 'info');
    t.innerHTML =
        '<div class="toast-icon"><i class="fas ' + (iconMap[type] || iconMap.info) + '"></i></div>' +
        '<div class="toast-body">' + escapeHtml(msg) + '</div>' +
        '<button class="toast-close" title="关闭"><i class="fas fa-times"></i></button>' +
        '<div class="toast-progress"></div>';
    wrap.appendChild(t);

    const removeToast = () => {
        if (t.classList.contains('toast-out')) return;
        t.classList.add('toast-out');
        setTimeout(() => t.remove(), 300);
    };
    t.querySelector('.toast-close').onclick = removeToast;
    const timer = setTimeout(removeToast, 2600);
    t.addEventListener('mouseenter', () => clearTimeout(timer));
}

function logout(expired) {
    localStorage.removeItem('mayfly_token');
    localStorage.removeItem('mayfly_user');
    localStorage.removeItem(UI_KEY);
    if (expired) alert('登录已过期，请重新登录');
    window.location.href = '/login';
}

function fmtSize(n) {
    if (n == null) return '-';
    if (n < 1024) return n + ' B';
    if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB';
    if (n < 1024 * 1024 * 1024) return (n / 1024 / 1024).toFixed(1) + ' MB';
    return (n / 1024 / 1024 / 1024).toFixed(2) + ' GB';
}

function fmtTime(ts) {
    if (!ts) return '-';
    const d = new Date(ts * 1000);
    const p = (n) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

function joinPath(dir, name) {
    if (!dir) return name;
    if (dir.includes('\\')) return dir.replace(/\\+$/, '') + '\\' + name;
    return dir.replace(/\/+$/, '') + '/' + name;
}

function dirOf(path) {
    if (!path) return path;
    if (path.includes('\\')) {
        const p = path.replace(/\\+$/, '');
        const i = p.lastIndexOf('\\');
        return i > 1 ? p.slice(0, i) : p.slice(0, 1) + '\\';
    }
    const p = path.replace(/\/+$/, '');
    const i = p.lastIndexOf('/');
    if (i <= 0) return '/';
    return p.slice(0, i);
}

function arrayBufferToBase64(buf) {
    let binary = '';
    const bytes = new Uint8Array(buf);
    const chunk = 0x8000;
    for (let i = 0; i < bytes.length; i += chunk) {
        binary += String.fromCharCode.apply(null, bytes.subarray(i, i + chunk));
    }
    return btoa(binary);
}

function base64ToBytes(b64) {
    const bin = atob(b64);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return bytes;
}

function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

function closeModal(id) { document.getElementById(id).style.display = 'none'; }
function openModal(id) { document.getElementById(id).style.display = 'flex'; }

function promptModal(title, placeholder, val) {
    return new Promise((resolve) => {
        document.getElementById('promptTitle').textContent = title;
        const input = document.getElementById('promptInput');
        input.placeholder = placeholder || '';
        input.value = val || '';
        openModal('promptModal');
        input.focus();
        const ok = () => {
            const v = input.value;
            closeModal('promptModal');
            document.getElementById('promptOkBtn').onclick = null;
            resolve(v);
        };
        document.getElementById('promptOkBtn').onclick = ok;
        input.onkeydown = (e) => { if (e.key === 'Enter') ok(); };
    });
}

// ===== 节点管理 =====
async function loadNodes(hideExisting) {
    const r = await api('GET', '/nodes');
    state.nodes = r.data.nodes || [];
    // 登录后首次加载时，默认隐藏历史节点（数据保留，连接列表里测试/进入后可恢复显示）
    if (hideExisting) {
        state.nodes.forEach((n) => { state.clearedNodes[n.id] = true; });
        state.currentNode = null;
    }
    renderNodeList();
    updateStats();
}

function updateStats() {
    const total = state.nodes.length;
    const connected = 0; // 后端暂无实时状态，保持为0或按需扩展
    const php = state.nodes.filter((n) => n.type === 'php').length;
    const other = total - php;
    document.getElementById('statTotal').textContent = total;
    document.getElementById('statConnected').textContent = connected;
    document.getElementById('statPhp').textContent = php;
    document.getElementById('statOther').textContent = other;
}

function renderNodeList() {
    const list = document.getElementById('nodeList');
    const q = document.getElementById('nodeSearch')?.value?.trim()?.toLowerCase() || '';
    list.innerHTML = '';
    // 只显示已连通且未被"清空"的节点
    const connectedNodes = state.nodes.filter((n) => getConnStatus(n.id).status === 'ok' && !state.clearedNodes[n.id]);
    const nodes = q ? connectedNodes.filter((n) => n.name.toLowerCase().includes(q) || n.url.toLowerCase().includes(q)) : connectedNodes;
    if (nodes.length === 0) {
        list.innerHTML = '<div class="empty-state"><i class="fas fa-inbox"></i><p>' + (connectedNodes.length ? '无匹配节点' : '暂无已连通节点，请先在连接列表中测试') + '</p></div>';
        return;
    }
    nodes.forEach((n) => {
        const item = document.createElement('div');
        item.className = 'node-item' + (state.currentNode && state.currentNode.id === n.id ? ' active' : '');
        const dot = document.createElement('span');
        dot.className = 'node-type-badge type-' + n.type;
        dot.textContent = n.type.toUpperCase();
        const nameDiv = document.createElement('div');
        nameDiv.className = 'node-item-name';
        nameDiv.textContent = n.name;
        const urlDiv = document.createElement('div');
        urlDiv.className = 'node-item-url';
        urlDiv.textContent = n.url;
        const textWrap = document.createElement('div');
        textWrap.className = 'node-item-text';
        textWrap.appendChild(nameDiv);
        textWrap.appendChild(urlDiv);
        item.appendChild(dot);
        item.appendChild(textWrap);
        item.onclick = () => selectNode(n);
        list.appendChild(item);
    });
}

// ===== 界面状态持久化 =====
function persistUI() {
    localStorage.setItem(UI_KEY, JSON.stringify({
        currentView: state.currentView,
        currentTab: state.currentTab,
        currentNodeId: state.currentNode ? state.currentNode.id : null,
        srvTermIds: Object.keys(state.srvTerms).map(Number),
    }));
}

function restoreUI() {
    const v = state.currentView;

    if (v === 'dashboard') { switchToDashboard(); return; }
    if (v === 'reverse-shell') { switchToReverseShell(); return; }
    if (v === 'servers') { switchToServers(); return; }
    if (v === 'server-term') { switchToServerTerm(); restoreSrvTerms(); return; }
    if (v === 'connections') { switchToConnections(); return; }
    if (v === 'system') { switchToSystem(); return; }

    // 工作区视图：切回对应标签页（登录后节点列表默认清空，不恢复历史选中节点）
    switchTab(state.currentTab || 'file');
}

// 刷新后恢复服务器终端列表：根据保存的服务器 ID 重新打开终端
async function restoreSrvTerms() {
    if (!state.servers || state.servers.length === 0) {
        await loadServers();
    }
    const ids = savedUI.srvTermIds || [];
    ids.forEach((id) => {
        const s = state.servers.find((x) => x.id === id);
        if (s) ensureSrvTerminal(s);
    });
}

// 渲染面包屑：已选中节点时显示节点名，否则显示「未选择节点」
function renderBreadcrumb() {
    const n = state.currentNode;
    document.querySelector('.breadcrumb').innerHTML = n
        ? '节点列表 / <span id="currentNodeTitle"><span class="type-badge type-' + n.type + '">' + n.type.toUpperCase() + '</span> ' + escapeHtml(n.name) + '</span>'
        : '节点列表 / <span id="currentNodeTitle"><span class="muted">未选择节点</span></span>';
}

function selectNode(node) {
    delete state.clearedNodes[node.id]; // 重新进入节点列表
    state.currentNode = node;
    state.filePath = '';
    state.editingFile = null;
    document.getElementById('fileEditor').classList.add('hidden');
    renderBreadcrumb();
    renderNodeList();
    updatePanelsForNode();
    persistUI();
}

function updatePanelsForNode() {
    if (state.currentTab === 'file') loadFiles('');
    else if (state.currentTab === 'cmd') loadCmdHistory();
    // terminal 由 switchTab 的 setTimeout 统一调用 ensureTerminal()，避免重复弹 toast
    else if (state.currentTab === 'db') {
        document.getElementById('dbResultTable').style.display = 'none';
        document.getElementById('dbResultMsg').textContent = '';
    }
}

let editingNodeId = null;
function showNodeModal(node) {
    editingNodeId = node ? node.id : null;
    document.getElementById('nodeModalTitle').textContent = node ? '编辑节点' : '添加节点';
    document.getElementById('nodeName').value = node ? node.name : '';
    document.getElementById('nodeURL').value = node ? node.url : '';
    document.getElementById('nodeType').value = node ? node.type : 'php';
    document.getElementById('nodeGroup').value = node ? (node.group || '') : '';
    document.getElementById('nodePass').value = node ? node.pass : '';
    document.getElementById('nodeRemark').value = node ? node.remark : '';
    openModal('nodeModal');
}

async function saveNode() {
    const body = {
        name: document.getElementById('nodeName').value.trim(),
        url: document.getElementById('nodeURL').value.trim(),
        type: document.getElementById('nodeType').value,
        group: document.getElementById('nodeGroup').value.trim(),
        pass: document.getElementById('nodePass').value.trim() || 'mayfly',
        remark: document.getElementById('nodeRemark').value.trim(),
    };
    if (!body.name || !body.url) { toast('名称和 URL 不能为空', 'error'); return; }
    let r;
    if (editingNodeId) r = await api('PUT', '/nodes/' + editingNodeId, body);
    else r = await api('POST', '/nodes', body);
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    closeModal('nodeModal');
    await loadNodes();
    toast(editingNodeId ? '已更新' : '已添加', 'success');
    if (editingNodeId && state.currentNode && state.currentNode.id === editingNodeId) {
        state.currentNode = r.data.node;
        document.getElementById('currentNodeTitle').innerHTML =
            '<span class="type-badge type-' + r.data.node.type + '">' + r.data.node.type.toUpperCase() + '</span> ' + escapeHtml(r.data.node.name);
    }
}

function clearNodeList() {
    const visible = state.nodes.filter((n) => getConnStatus(n.id).status === 'ok' && !state.clearedNodes[n.id]);
    if (visible.length === 0) { toast('节点列表已为空', 'warning'); return; }
    if (!confirm('确定清空节点列表？此操作不会删除连接列表中的数据')) return;
    visible.forEach((n) => { state.clearedNodes[n.id] = true; });
    if (state.currentNode && state.terms[state.currentNode.id]) destroyTerminal(state.currentNode.id);
    state.currentNode = null;
    document.getElementById('currentNodeTitle').innerHTML = '<span class="muted">未选择节点</span>';
    renderNodeList();
    persistUI();
    toast('已清空节点列表', 'success');
}

async function testNode() {
    const n = state.currentNode;
    if (!n) { toast('请先选择节点', 'error'); return; }
    toast('正在测试连接...', 'info');
    const r = await api('POST', '/nodes/' + n.id + '/test', {});
    if (r.data.ok) {
        toast('连接成功', 'success');
        document.getElementById('cmdOutput').textContent = r.data.info || '';
        switchTab('cmd');
    } else {
        toast('连接失败: ' + r.data.message, 'error');
    }
}

// ===== 连接列表 =====
function getConnGroups() {
    const q = document.getElementById('connSearch')?.value?.trim()?.toLowerCase() || '';
    const filtered = q
        ? state.nodes.filter((n) => n.name.toLowerCase().includes(q) || n.url.toLowerCase().includes(q))
        : state.nodes;
    const groups = {};
    filtered.forEach((n) => {
        const g = n.group || '默认';
        if (!groups[g]) groups[g] = [];
        groups[g].push(n);
    });
    // 默认分组排第一
    const sorted = Object.keys(groups).sort((a, b) => {
        if (a === '默认') return -1;
        if (b === '默认') return 1;
        return a.localeCompare(b);
    });
    return sorted.map((g) => ({ name: g, nodes: groups[g] }));
}

function getConnStatus(nodeId) {
    const rt = state.connStatus[nodeId];
    if (rt && rt.status === 'testing') {
        return rt;
    }
    const n = state.nodes.find((x) => x.id === nodeId);
    if (n && n.last_test_status) {
        return {
            status: n.last_test_status,
            message: n.last_test_message || '',
            info: rt ? rt.info : '',
        };
    }
    return rt || { status: 'untested', message: '', info: '' };
}

function renderConnList() {
    const container = document.getElementById('connGroups');
    const groups = getConnGroups();
    if (groups.length === 0) {
        container.innerHTML = '<div class="conn-empty"><i class="fas fa-server"></i><p>' + (state.nodes.length ? '无匹配连接' : '暂无连接，请先添加节点') + '</p></div>';
        return;
    }
    container.innerHTML = '';
    groups.forEach((grp) => {
        const visibleNodes = grp.nodes.filter((n) => {
            if (state.connFilter === 'all') return true;
            const s = getConnStatus(n.id).status;
            return s === state.connFilter;
        });
        if (visibleNodes.length === 0) return;

        const groupEl = document.createElement('div');
        groupEl.className = 'conn-group';

        // 分组头部
        const header = document.createElement('div');
        header.className = 'conn-group-header';
        header.innerHTML = `
            <div class="conn-group-title">
                <i class="fas fa-chevron-down"></i>
                <span>${escapeHtml(grp.name)}</span>
                <span class="conn-group-count">${visibleNodes.length}</span>
            </div>
            <div class="conn-group-actions">
                <button class="btn btn-primary btn-sm batch-group-btn" data-group="${escapeHtml(grp.name)}">
                    <i class="fas fa-bolt"></i> 批量测试
                </button>
            </div>
        `;
        header.onclick = (e) => {
            if (e.target.closest('.batch-group-btn')) return;
            groupEl.classList.toggle('collapsed');
        };
        groupEl.appendChild(header);

        // 分组内容
        const body = document.createElement('div');
        body.className = 'conn-group-body';
        visibleNodes.forEach((n) => {
            const st = getConnStatus(n.id);
            const testTime = n.last_test_time ? formatTestTime(n.last_test_time) : '';
            const item = document.createElement('div');
            item.className = 'conn-item ' + st.status;
            item.dataset.nodeId = n.id;
            item.innerHTML = `
                <div class="conn-item-header">
                    <span class="conn-item-name">${escapeHtml(n.name)}</span>
                    <span class="type-badge type-${n.type}">${n.type.toUpperCase()}</span>
                </div>
                <div class="conn-item-url">${escapeHtml(n.url)}</div>
                <div class="conn-item-footer">
                    <div class="conn-footer-left">
                        <span class="conn-status ${st.status}">${st.status === 'ok' ? '已连通' : st.status === 'fail' ? '连接失败' : st.status === 'testing' ? '测试中...' : '未测试'}</span>
                        ${testTime ? `<span class="conn-test-time"><i class="fas fa-clock"></i>${testTime}</span>` : ''}
                    </div>
                    <div class="conn-item-actions">
                        <button class="conn-mini-btn connect-btn" title="连接此节点（进入文件管理）"><i class="fas fa-link"></i></button>
                        <button class="conn-mini-btn test-single-btn" title="测试此连接"><i class="fas fa-plug"></i></button>
                        <button class="conn-mini-btn edit-node-btn" title="编辑"><i class="fas fa-edit"></i></button>
                        <button class="conn-mini-btn danger del-node-btn" title="删除"><i class="fas fa-trash"></i></button>
                    </div>
                </div>
                ${st.message && st.status === 'fail' ? `<div style="font-size:11px;color:var(--danger);word-break:break-all;">${escapeHtml(st.message)}</div>` : ''}
                ${st.info && st.status === 'ok' ? `<div style="font-size:11px;color:var(--text-muted);word-break:break-all;">${escapeHtml(st.info)}</div>` : ''}
            `;
            // 事件绑定
            item.querySelector('.connect-btn').onclick = (e) => { e.stopPropagation(); connectNode(n); };
            item.querySelector('.test-single-btn').onclick = (e) => { e.stopPropagation(); testSingleConn(n.id); };
            item.querySelector('.edit-node-btn').onclick = (e) => { e.stopPropagation(); showNodeModal(n); };
            item.querySelector('.del-node-btn').onclick = (e) => { e.stopPropagation(); deleteConnNode(n); };
            item.onclick = () => connectNode(n);
            body.appendChild(item);
        });
        groupEl.appendChild(body);
        container.appendChild(groupEl);
    });

    // 绑定分组批量测试
    container.querySelectorAll('.batch-group-btn').forEach((btn) => {
        btn.onclick = (e) => {
            e.stopPropagation();
            const groupName = btn.dataset.group;
            const group = groups.find((g) => g.name === groupName);
            if (group) batchTest(group.nodes.map((n) => n.id));
        };
    });
}

async function batchTest(ids) {
    if (!ids || ids.length === 0) { toast('没有可测试的连接', 'warning'); return; }
    // 标记为测试中
    ids.forEach((id) => {
        state.connStatus[id] = { status: 'testing', message: '', info: '' };
    });
    renderConnList();
    toast(`正在测试 ${ids.length} 个连接...`, 'info');

    const r = await api('POST', '/nodes/batch-test', { ids });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    const results = r.data.results || [];
    let okCount = 0, failCount = 0;
    results.forEach((res) => {
        state.connStatus[res.id] = {
            status: res.ok ? 'ok' : 'fail',
            message: res.message || '',
            info: res.info || '',
        };
        if (res.ok) okCount++; else failCount++;
    });
    renderConnList();
    toast(`测试完成: ${okCount} 成功, ${failCount} 失败`, okCount > 0 ? 'success' : 'error');
}

// 连接节点：未连通时先发起测试，成功后自动进入文件管理
async function connectNode(node) {
    if (getConnStatus(node.id).status !== 'ok') {
        toast(`正在连接「${node.name}」...`, 'info');
        await testSingleConn(node.id);
        if (getConnStatus(node.id).status !== 'ok') {
            toast('连接失败，请检查脚本地址与连接密码', 'error');
            return;
        }
    }
    selectNode(node);
    switchTab('file');
}

async function testSingleConn(nodeId) {
    state.connStatus[nodeId] = { status: 'testing', message: '', info: '' };
    renderConnList();
    const r = await api('POST', '/nodes/batch-test', { ids: [nodeId] });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    const res = (r.data.results || [])[0];
    if (res) {
        state.connStatus[nodeId] = {
            status: res.ok ? 'ok' : 'fail',
            message: res.message || '',
            info: res.info || '',
        };
    }
    renderConnList();
}

async function deleteConnNode(node) {
    if (!confirm('确定删除节点「' + node.name + '」？')) return;
    const r = await api('DELETE', '/nodes/' + node.id);
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    delete state.connStatus[node.id];
    await loadNodes();
    renderConnList();
    toast('已删除', 'success');
}

// ===== 标签页 / 视图切换 =====
function switchTab(tab) {
    state.currentTab = tab;
    state.currentView = 'workspace';
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.toggle('active', t.dataset.tab === tab));
    document.getElementById('workspaceView').classList.remove('hidden');
    document.getElementById('dashboardView').classList.add('hidden');
    document.getElementById('connectionsView').classList.add('hidden');
    document.getElementById('reverseShellView').classList.add('hidden');
    document.getElementById('serversView').classList.add('hidden');
    document.getElementById('serverTermView').classList.add('hidden');
    document.getElementById('systemView').classList.add('hidden');
    document.getElementById('headerSearchWrap').style.display = '';
    renderBreadcrumb();
    document.querySelectorAll('.tab-panel').forEach((p) => p.classList.remove('active'));
    document.getElementById('panel-' + tab).classList.add('active');
    renderNodeList();
    updatePanelsForNode();
    if (tab === 'terminal') {
        // 等浏览器完成 visibility 切换和一次 paint 后再初始化/刷新终端
        requestAnimationFrame(() => setTimeout(ensureTerminal, 30));
    }
    persistUI();
}

function switchToDashboard() {
    state.currentView = 'dashboard';
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.remove('active'));
    document.querySelector('.nav-item[data-view="dashboard"]').classList.add('active');
    document.getElementById('workspaceView').classList.add('hidden');
    document.getElementById('fileEditor').classList.add('hidden');
    document.getElementById('dashboardView').classList.remove('hidden');
    document.getElementById('connectionsView').classList.add('hidden');
    document.getElementById('reverseShellView').classList.add('hidden');
    document.getElementById('serversView').classList.add('hidden');
    document.getElementById('serverTermView').classList.add('hidden');
    document.getElementById('systemView').classList.add('hidden');
    document.getElementById('headerSearchWrap').style.display = 'none';
    document.querySelector('.breadcrumb').textContent = '仪表盘';
    loadDashboard();
    persistUI();
}

// ===== 仪表盘数据加载与渲染 =====
async function loadDashboard() {
    try {
        const r = await api('GET', '/dashboard');
        if (r.status === 401) return;
        renderDashboard(r.data);
    } catch (e) {
        console.error('loadDashboard error:', e);
    }
}

function renderDashboard(d) {
    // 统计卡片
    document.getElementById('dashNodeTotal').textContent = d.nodes.total;
    document.getElementById('dashNodeOnline').textContent = d.nodes.online;
    document.getElementById('dashNodeOffline').textContent = d.nodes.offline;
    document.getElementById('dashSrvTotal').textContent = d.servers.total;
    document.getElementById('dashSrvOnline').textContent = d.servers.online;
    document.getElementById('dashSrvOffline').textContent = d.servers.offline;
    document.getElementById('dashListenerTotal').textContent = d.listeners.total;
    document.getElementById('dashListenerActive').textContent = d.listeners.active;

    // 分组数 = 节点分组 + 服务器分组（去重）
    const allGroups = new Set([
        ...Object.keys(d.nodes.groups || {}),
        ...Object.keys(d.servers.groups || {}),
    ]);
    document.getElementById('dashGroupTotal').textContent = allGroups.size;
    document.getElementById('dashGroupSub').textContent = allGroups.size > 0
        ? '节点 ' + Object.keys(d.nodes.groups || {}).length + ' / 服务器 ' + Object.keys(d.servers.groups || {}).length
        : '-';

    // 节点类型分布条形图
    const typeEl = document.getElementById('dashTypeChart');
    const types = d.nodes.types || {};
    const typeKeys = Object.keys(types);
    if (typeKeys.length === 0) {
        typeEl.innerHTML = '<div class="empty-state"><i class="fas fa-inbox"></i><p>暂无节点数据</p></div>';
    } else {
        const maxVal = Math.max(...Object.values(types), 1);
        typeEl.innerHTML = '<div class="dash-type-bars">' + typeKeys.map((k) =>
            '<div class="dash-type-row">' +
                '<div class="dash-type-label">' + escapeHtml(k) + '</div>' +
                '<div class="dash-type-bar-wrap"><div class="dash-type-bar ' + esc(k) + '" style="width:' + Math.round(types[k] / maxVal * 100) + '%"></div></div>' +
                '<div class="dash-type-count">' + types[k] + '</div>' +
            '</div>'
        ).join('') + '</div>';
    }

    // 分组概览
    const groupEl = document.getElementById('dashGroupOverview');
    if (allGroups.size === 0) {
        groupEl.innerHTML = '<div class="empty-state"><i class="fas fa-inbox"></i><p>暂无分组</p></div>';
    } else {
        const merged = {};
        for (const [g, c] of Object.entries(d.nodes.groups || {})) {
            merged[g] = (merged[g] || '') + '节点 ' + c;
        }
        for (const [g, c] of Object.entries(d.servers.groups || {})) {
            merged[g] = (merged[g] ? merged[g] + ' / ' : '') + '服务器 ' + c;
        }
        groupEl.innerHTML = '<div class="dash-group-grid">' + Object.entries(merged).map(([g, desc]) =>
            '<div class="dash-group-chip">' +
                '<span class="dash-group-name">' + escapeHtml(g) + '</span>' +
                '<span class="dash-group-count">' + desc + '</span>' +
            '</div>'
        ).join('') + '</div>';
    }

    // 最近命令活动
    const cmdEl = document.getElementById('dashRecentCmds');
    const cmds = d.recent_commands || [];
    if (cmds.length === 0) {
        cmdEl.innerHTML = '<div class="empty-state"><i class="fas fa-inbox"></i><p>暂无活动记录</p></div>';
    } else {
        cmdEl.innerHTML = '<div class="dash-cmd-list">' + cmds.map((c) => {
            const t = new Date(c.time * 1000);
            const ts = t.getMonth() + 1 + '/' + t.getDate() + ' ' +
                String(t.getHours()).padStart(2, '0') + ':' + String(t.getMinutes()).padStart(2, '0');
            return '<div class="dash-cmd-item">' +
                '<div class="dash-cmd-icon"><i class="fas fa-terminal"></i></div>' +
                '<div class="dash-cmd-body">' +
                    '<div class="dash-cmd-header">' +
                        '<span class="dash-cmd-node">' + escapeHtml(c.node) + '</span>' +
                        '<span class="dash-cmd-time">' + ts + '</span>' +
                    '</div>' +
                    '<div class="dash-cmd-text">$ ' + escapeHtml(c.cmd) + '</div>' +
                    (c.output ? '<div class="dash-cmd-output">' + escapeHtml(c.output) + '</div>' : '') +
                '</div>' +
            '</div>';
        }).join('') + '</div>';
    }

    // 连接失败告警
    const failCard = document.getElementById('dashFailCard');
    const failList = d.recent_fails || [];
    if (failList.length === 0) {
        failCard.style.display = 'none';
    } else {
        failCard.style.display = '';
        document.getElementById('dashFailList').innerHTML = '<div class="dash-fail-list">' + failList.map((f) => {
            const t = new Date(f.time * 1000);
            const ts = t.getMonth() + 1 + '/' + t.getDate() + ' ' +
                String(t.getHours()).padStart(2, '0') + ':' + String(t.getMinutes()).padStart(2, '0');
            return '<div class="dash-fail-item">' +
                '<div class="dash-fail-icon"><i class="fas fa-times"></i></div>' +
                '<div class="dash-fail-info">' +
                    '<div class="dash-fail-name">' + escapeHtml(f.name) + '</div>' +
                    '<div class="dash-fail-msg">' + escapeHtml(f.message || '连接失败') + '</div>' +
                '</div>' +
                '<div class="dash-fail-time">' + ts + '</div>' +
            '</div>';
        }).join('') + '</div>';
    }
}

function switchToConnections() {
    state.currentView = 'connections';
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.remove('active'));
    document.querySelector('.nav-item[data-view="connections"]').classList.add('active');
    document.getElementById('workspaceView').classList.add('hidden');
    document.getElementById('fileEditor').classList.add('hidden');
    document.getElementById('dashboardView').classList.add('hidden');
    document.getElementById('connectionsView').classList.remove('hidden');
    document.getElementById('reverseShellView').classList.add('hidden');
    document.getElementById('serversView').classList.add('hidden');
    document.getElementById('serverTermView').classList.add('hidden');
    document.getElementById('systemView').classList.add('hidden');
    document.getElementById('headerSearchWrap').style.display = '';
    document.querySelector('.breadcrumb').innerHTML = '节点列表 / <span id="currentNodeTitle"><span class="muted">连接列表</span></span>';
    renderConnList();
    persistUI();
}

function switchToReverseShell() {
    state.currentView = 'reverse-shell';
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.remove('active'));
    document.querySelector('.nav-item[data-view="reverse-shell"]').classList.add('active');
    document.getElementById('workspaceView').classList.add('hidden');
    document.getElementById('fileEditor').classList.add('hidden');
    document.getElementById('dashboardView').classList.add('hidden');
    document.getElementById('connectionsView').classList.add('hidden');
    document.getElementById('reverseShellView').classList.remove('hidden');
    document.getElementById('serversView').classList.add('hidden');
    document.getElementById('serverTermView').classList.add('hidden');
    document.getElementById('systemView').classList.add('hidden');
    document.getElementById('headerSearchWrap').style.display = 'none';
    document.querySelector('.breadcrumb').textContent = '反弹Shell';
    loadListeners();
    genPayloads();
    persistUI();
}

function switchToServers() {
    state.currentView = 'servers';
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.remove('active'));
    document.querySelector('.nav-item[data-view="servers"]').classList.add('active');
    document.getElementById('workspaceView').classList.add('hidden');
    document.getElementById('fileEditor').classList.add('hidden');
    document.getElementById('dashboardView').classList.add('hidden');
    document.getElementById('connectionsView').classList.add('hidden');
    document.getElementById('reverseShellView').classList.add('hidden');
    document.getElementById('serversView').classList.remove('hidden');
    document.getElementById('serverTermView').classList.add('hidden');
    document.getElementById('systemView').classList.add('hidden');
    document.getElementById('headerSearchWrap').style.display = 'none';
    document.querySelector('.breadcrumb').textContent = '资源管理';
    loadServers();
    persistUI();
}

// ===== 服务器 SSH 交互终端 =====
function switchToServerTerm() {
    state.currentView = 'server-term';
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.remove('active'));
    document.querySelector('.nav-item[data-view="server-term"]').classList.add('active');
    document.getElementById('workspaceView').classList.add('hidden');
    document.getElementById('fileEditor').classList.add('hidden');
    document.getElementById('dashboardView').classList.add('hidden');
    document.getElementById('connectionsView').classList.add('hidden');
    document.getElementById('reverseShellView').classList.add('hidden');
    document.getElementById('serversView').classList.add('hidden');
    document.getElementById('serverTermView').classList.remove('hidden');
    document.getElementById('systemView').classList.add('hidden');
    document.getElementById('headerSearchWrap').style.display = 'none';
    document.querySelector('.breadcrumb').textContent = '终端操作';
    renderSrvTermList();
    // 切回时重新激活当前终端实例（等价于点击服务器列表项），确保 active 类和重绘都到位
    if (state.activeSrvTermId && state.srvTerms[state.activeSrvTermId]) {
        requestAnimationFrame(() => setTimeout(() => switchSrvTerm(state.activeSrvTermId), 30));
    }
    persistUI();
}

function switchToSystem() {
    state.currentView = 'system';
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.remove('active'));
    document.querySelector('.nav-item[data-view="system"]').classList.add('active');
    document.getElementById('workspaceView').classList.add('hidden');
    document.getElementById('fileEditor').classList.add('hidden');
    document.getElementById('dashboardView').classList.add('hidden');
    document.getElementById('connectionsView').classList.add('hidden');
    document.getElementById('reverseShellView').classList.add('hidden');
    document.getElementById('serversView').classList.add('hidden');
    document.getElementById('serverTermView').classList.add('hidden');
    document.getElementById('systemView').classList.remove('hidden');
    document.getElementById('headerSearchWrap').style.display = 'none';
    document.querySelector('.breadcrumb').textContent = '系统管理';
    loadSystemInfo();
    loadSystemSettings();
    loadSystemAuditLogs();
    persistUI();
}

// ===== 系统管理逻辑 =====
async function loadSystemInfo() {
    try {
        const r = await api('GET', '/system/info');
        if (r.status === 401) return;
        const d = r.data;

        // 系统概览统计卡片
        const stats = [
            { icon: 'fa-rocket', label: '版本', value: escapeHtml(d.version) },
            { icon: 'fa-code', label: 'Go 版本', value: escapeHtml(d.go_version) },
            { icon: 'fa-laptop', label: '操作系统', value: escapeHtml(d.os + '/' + d.arch) },
            { icon: 'fa-microchip', label: 'CPU 核心', value: d.cpu_num },
            { icon: 'fa-circle-nodes', label: 'Goroutine', value: d.goroutines },
            { icon: 'fa-memory', label: '内存分配', value: d.mem_alloc + ' MB' },
            { icon: 'fa-hdd', label: '系统内存', value: d.mem_sys + ' MB' },
            { icon: 'fa-database', label: '累计分配', value: d.mem_total + ' MB' },
            { icon: 'fa-recycle', label: 'GC 次数', value: d.gc_count },
            { icon: 'fa-clipboard-list', label: '审计日志', value: d.audit_count + ' 条' },
        ];
        document.getElementById('sysStatGrid').innerHTML = stats.map((s) =>
            '<div class="sys-stat-item">' +
                '<div class="sys-stat-icon"><i class="fas ' + s.icon + '"></i></div>' +
                '<div class="sys-stat-info">' +
                    '<div class="sys-stat-value">' + s.value + '</div>' +
                    '<div class="sys-stat-label">' + s.label + '</div>' +
                '</div>' +
            '</div>'
        ).join('');

        // 头部运行时长
        document.getElementById('sysUptimeMeta').innerHTML =
            '<i class="fas fa-clock"></i> 已运行 ' + escapeHtml(d.uptime) +
            ' · 启动于 ' + escapeHtml(d.start_time);

        // 数据文件
        const files = d.data_files || {};
        const fileNames = Object.keys(files);

        // 数据管理 - 数据文件列表
        const dataFilesEl = document.getElementById('sysDataFiles');
        if (dataFilesEl) {
            if (fileNames.length > 0) {
                dataFilesEl.innerHTML = '<div class="sys-data-files-list">' + fileNames.map((f) =>
                    '<div class="sys-data-file-item">' +
                        '<span class="sys-data-file-icon"><i class="fas fa-file-alt"></i></span>' +
                        '<span class="sys-data-file-name">' + escapeHtml(f) + '</span>' +
                        '<span class="sys-data-file-size">' + formatFileSize(files[f]) + '</span>' +
                    '</div>'
                ).join('') + '</div>';
            } else {
                dataFilesEl.innerHTML = '<div style="font-size:12px;color:var(--text-muted);padding:8px 0;">暂无数据文件</div>';
            }
        }
    } catch (e) {
        console.error('loadSystemInfo error:', e);
    }
}

async function loadSystemSettings() {
    try {
        const r = await api('GET', '/system/settings');
        if (r.status === 401) return;
        document.getElementById('sysSessionTimeout').value = r.data.session_timeout || 30;
        document.getElementById('sysShell').value = r.data.shell || '';
    } catch (e) {
        console.error('loadSystemSettings error:', e);
    }
}

async function loadSystemAuditLogs() {
    const el = document.getElementById('sysAuditBody');
    try {
        const r = await api('GET', '/system/audit-logs?limit=200');
        if (r.status === 401) return;
        const logs = r.data.logs || [];
        if (logs.length === 0) {
            el.innerHTML = '<div class="empty-state"><i class="fas fa-clipboard-check"></i><p>暂无审计记录</p></div>';
            return;
        }
        el.innerHTML =
            '<div style="overflow-x:auto;">' +
            '<table class="sys-audit-table">' +
            '<thead><tr>' +
                '<th>时间</th><th>用户</th><th>操作</th><th>目标</th><th>IP</th>' +
            '</tr></thead>' +
            '<tbody>' + logs.map((l) => {
                const t = new Date(l.time);
                const ts = t.getFullYear() + '-' +
                    String(t.getMonth() + 1).padStart(2, '0') + '-' +
                    String(t.getDate()).padStart(2, '0') + ' ' +
                    String(t.getHours()).padStart(2, '0') + ':' +
                    String(t.getMinutes()).padStart(2, '0') + ':' +
                    String(t.getSeconds()).padStart(2, '0');
                return '<tr>' +
                    '<td class="sys-audit-time">' + ts + '</td>' +
                    '<td>' + escapeHtml(l.user || '-') + '</td>' +
                    '<td><span class="sys-audit-action">' + escapeHtml(l.action) + '</span></td>' +
                    '<td>' + escapeHtml(l.detail || l.target || '-') + '</td>' +
                    '<td class="sys-audit-ip">' + escapeHtml(l.ip || '-') + '</td>' +
                '</tr>';
            }).join('') + '</tbody></table></div>' +
            '<div style="margin-top:8px;font-size:11px;color:var(--text-muted);">共 ' + r.data.total + ' 条记录</div>';
    } catch (e) {
        el.innerHTML = '<div class="empty-state"><i class="fas fa-exclamation-triangle"></i><p>加载失败</p></div>';
    }
}

function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// 从服务器卡片打开终端（入口）
function openServerTerminal(serverId) {
    const s = state.servers.find((x) => x.id === serverId);
    if (!s) { toast('服务器不存在', 'error'); return; }
    switchToServerTerm();
    ensureSrvTerminal(s);
}

function ensureSrvTerminal(s) {
    if (state.srvTerms[s.id]) { switchSrvTerm(s.id); return; }
    createSrvTerminal(s);
}

function createSrvTerminal(s) {
    const container = document.getElementById('srvTermWrap');
    if (!container._ready) { container.innerHTML = ''; container._ready = true; }
    const div = document.createElement('div');
    div.id = 'srvterm-' + s.id;
    div.className = 'terminal-instance active';
    container.appendChild(div);

    const term = new Terminal({
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, Consolas, monospace',
        cursorBlink: true,
        theme: getTerminalTheme(),
        scrollback: 10000,
    });
    let fitAddon;
    try { fitAddon = new FitAddon.FitAddon(); term.loadAddon(fitAddon); } catch (e) {}
    term.open(div);
    if (fitAddon) setTimeout(() => { fitAddon.fit(); sendSrvResize(s.id); }, 50);

    const termData = { term, fitAddon, ws: null, server: s, el: div };
    state.srvTerms[s.id] = termData;
    state.activeSrvTermId = s.id;

    term.onData((data) => handleSrvTermInput(s.id, data));
    term.onResize(() => sendSrvResize(s.id));
    connectSrvTermWS(s.id);
    watchTermResize(termData, div);
    renderSrvTermList();
    persistUI();
}

function connectSrvTermWS(serverId) {
    const td = state.srvTerms[serverId];
    if (!td) return;
    const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') +
        location.host + '/api/servers/' + serverId + '/terminal?token=' + encodeURIComponent(state.token);
    // 清理旧连接和心跳定时器
    if (td._pingTimer) { clearInterval(td._pingTimer); td._pingTimer = null; }
    if (td.ws) { try { td.ws.onclose = null; td.ws.close(); } catch (e) {} }
    const ws = new WebSocket(wsUrl);
    td.ws = ws;
    td._closed = false;
    ws.onopen = () => {
        // 心跳保活：每 25s 发一个 ping，防止空闲被中间层断开
        td._pingTimer = setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
                try { ws.send(JSON.stringify({ type: 'ping' })); } catch (e) {}
            }
        }, 25000);
    };
    ws.onmessage = (e) => {
        try {
            const msg = JSON.parse(e.data);
            if (msg.type === 'output' || msg.type === 'error') {
                td.term.write(String(msg.data));
            } else if (msg.type === 'ready') {
                td.term.write('\x1b[32m[SSH 已连接 ' + escapeHtml(td.server.host || '') + ']\x1b[0m\r\n');
                sendSrvResize(serverId);
            }
        } catch (err) {}
    };
    ws.onclose = () => {
        if (td._pingTimer) { clearInterval(td._pingTimer); td._pingTimer = null; }
        if (state.srvTerms[serverId] && !td._closed) {
            td.term.write('\x1b[33m\r\n[连接已断开，3 秒后重连…]\x1b[0m\r\n');
            setTimeout(() => {
                if (state.srvTerms[serverId] && state.currentView === 'server-term' && state.activeSrvTermId === serverId) {
                    connectSrvTermWS(serverId);
                }
            }, 3000);
        }
    };
    ws.onerror = () => { try { ws.close(); } catch (e) {} };
}

function handleSrvTermInput(serverId, data) {
    const td = state.srvTerms[serverId];
    if (!td || !td.ws || td.ws.readyState !== WebSocket.OPEN) return;
    td.ws.send(JSON.stringify({ type: 'input', data }));
}

function sendSrvResize(serverId) {
    const td = state.srvTerms[serverId];
    if (!td || !td.ws || td.ws.readyState !== WebSocket.OPEN || !td.term) return;
    td.ws.send(JSON.stringify({ type: 'resize', cols: td.term.cols, rows: td.term.rows }));
}

function switchSrvTerm(serverId) {
    const td = state.srvTerms[serverId];
    if (!td) return;
    document.querySelectorAll('#srvTermWrap .terminal-instance').forEach((el) => el.classList.remove('active'));
    const div = document.getElementById('srvterm-' + serverId);
    if (div) div.classList.add('active');
    state.activeSrvTermId = serverId;
    // 切回来时如果 WS 已断开，立即重连
    if (td.ws && td.ws.readyState !== WebSocket.OPEN && td.ws.readyState !== WebSocket.CONNECTING) {
        td.term.write('\x1b[33m\r\n[重新连接中…]\x1b[0m\r\n');
        connectSrvTermWS(serverId);
    }
    requestAnimationFrame(() => setTimeout(() => refitTerm(td), 30));
    renderSrvTermList();
}

function destroySrvTerminal(serverId) {
    const t = state.srvTerms[serverId];
    if (!t) return;
    if (t._ro) { try { t._ro.disconnect(); } catch (e) {} }
    if (t.ws) t.ws.close();
    if (t.term) t.term.dispose();
    delete state.srvTerms[serverId];
    const div = document.getElementById('srvterm-' + serverId);
    if (div) div.remove();
    if (state.activeSrvTermId === serverId) {
        const ids = Object.keys(state.srvTerms);
        state.activeSrvTermId = ids.length > 0 ? parseInt(ids[0]) : null;
        if (state.activeSrvTermId) {
            const d = document.getElementById('srvterm-' + state.activeSrvTermId);
            if (d) d.classList.add('active');
        }
    }
    renderSrvTermList();
    persistUI();
}

function renderSrvTermList() {
    const container = document.getElementById('srvTermList');
    if (!container) return;
    const ids = Object.keys(state.srvTerms);
    if (ids.length === 0) {
        container.innerHTML = '<div class="server-term-empty">暂无终端</div>';
        return;
    }
    container.innerHTML = ids.map((id) => {
        const td = state.srvTerms[id];
        const s = td.server;
        const active = state.activeSrvTermId == id ? 'active' : '';
        return `
            <div class="srv-term-item ${active}" data-srv-term="${id}">
                <div class="srv-term-info">
                    <span class="srv-term-name" title="${escapeHtml(s.name || s.host)}">${escapeHtml(s.name || s.host)}</span>
                    <span class="mono srv-term-addr">${escapeHtml(s.host)}:${s.port || 22}</span>
                </div>
                <button class="srv-term-close" data-srv-term-close="${id}" title="关闭终端"><i class="fas fa-times"></i></button>
            </div>`;
    }).join('');
    container.querySelectorAll('.srv-term-item').forEach((item) => {
        item.onclick = (e) => {
            if (e.target.closest('.srv-term-close')) return;
            switchSrvTerm(parseInt(item.dataset.srvTerm));
        };
    });
    container.querySelectorAll('.srv-term-close').forEach((btn) => {
        btn.onclick = (e) => {
            e.stopPropagation();
            destroySrvTerminal(parseInt(btn.dataset.srvTermClose));
        };
    });
}

// ===== 命令执行 =====
async function runCmd() {
    const n = state.currentNode;
    if (!n) { toast('请先选择节点', 'error'); return; }
    const input = document.getElementById('cmdInput');
    const cmd = input.value.trim();
    if (!cmd) return;
    const out = document.getElementById('cmdOutput');
    out.textContent += '\n> ' + cmd + '\n';
    input.value = '';
    const r = await api('POST', '/nodes/' + n.id + '/cmd', { cmd });
    if (r.data.error) {
        out.textContent += '[错误] ' + r.data.error + '\n';
    } else {
        out.textContent += r.data.output || '(无输出)';
        if (r.data.output && !r.data.output.endsWith('\n')) out.textContent += '\n';
    }
    out.scrollTop = out.scrollHeight;
    input.focus();
}

// ===== 命令快捷面板 =====
const quickCommands = [
    { group: '系统信息', cmds: ['id', 'whoami', 'uname -a', 'hostname', 'pwd'] },
    { group: '网络配置', cmds: ['ifconfig', 'ip addr', 'netstat -antp', 'ss -antp'] },
    { group: '进程与资源', cmds: ['ps aux', 'df -h', 'free -m', 'ls -la'] },
    { group: '用户管理', cmds: ['cat /etc/passwd', 'w', 'who', 'useradd test'] },
];

function renderQuickCmds() {
    const grid = document.getElementById('cmdQuickGrid');
    if (!grid) return;
    grid.innerHTML = '';
    quickCommands.forEach((g) => {
        const group = document.createElement('div');
        group.className = 'cmd-quick-group';
        const title = document.createElement('div');
        title.className = 'cmd-quick-group-title';
        title.textContent = g.group;
        group.appendChild(title);
        const wrap = document.createElement('div');
        wrap.className = 'cmd-quick-grid';
        g.cmds.forEach((cmd) => {
            const btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'cmd-quick-btn';
            btn.textContent = cmd;
            btn.title = '执行 ' + cmd;
            btn.onclick = () => runQuickCmd(cmd);
            wrap.appendChild(btn);
        });
        group.appendChild(wrap);
        grid.appendChild(group);
    });
}

function runQuickCmd(cmd) {
    const input = document.getElementById('cmdInput');
    if (!input) return;
    input.value = cmd;
    runCmd();
}

// 从服务端缓存文件加载当前节点的命令执行历史
async function loadCmdHistory() {
    const n = state.currentNode;
    const out = document.getElementById('cmdOutput');
    if (!n) { out.textContent = ''; toast('请先选择节点', 'error'); return; }
    const r = await api('GET', '/nodes/' + n.id + '/cmd/history');
    const records = (r.data && r.data.history) || [];
    let text = '';
    records.forEach((rec) => {
        text += '\n> ' + rec.cmd + '\n';
        if (rec.error) {
            text += '[错误] ' + rec.error + '\n';
        } else {
            text += (rec.output || '(无输出)') + (rec.output && !rec.output.endsWith('\n') ? '\n' : '');
        }
    });
    out.textContent = text;
    out.scrollTop = out.scrollHeight;
}

// 清空当前节点的命令执行历史
async function clearCmdHistory() {
    const n = state.currentNode;
    if (!n) { toast('请先选择节点', 'error'); return; }
    await api('DELETE', '/nodes/' + n.id + '/cmd/history');
    document.getElementById('cmdOutput').textContent = '';
    toast('已清空命令历史', 'success');
}

// ===== 文件管理 =====
async function loadFiles(path) {
    const n = state.currentNode;
    if (!n) { toast('请先选择节点', 'error'); return; }
    const r = await api('POST', '/nodes/' + n.id + '/file/list', { path: path || '' });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    state.filePath = r.data.path;
    document.getElementById('filePath').value = state.filePath;
    state.fileParent = r.data.parent || '';
    renderFileTable(r.data.entries || []);
}

function renderFileTable(entries) {
    const tbody = document.getElementById('fileTableBody');
    tbody.innerHTML = '';
    entries.forEach((e) => {
        const tr = document.createElement('tr');
        const icon = e.type === 'd' ? '<i class="fas fa-folder folder-icon"></i>' : '<i class="fas fa-file file-icon"></i>';
        const nameTd = document.createElement('td');
        nameTd.innerHTML = icon + ' ' + escapeHtml(e.name);
        nameTd.className = 'file-name';
        nameTd.onclick = () => {
            if (e.type === 'd') {
                const target = e.name === '..' ? (state.fileParent || dirOf(state.filePath)) : joinPath(state.filePath, e.name);
                loadFiles(target);
            } else {
                openFile(e.name);
            }
        };
        tr.appendChild(nameTd);
        const typeTd = document.createElement('td');
        typeTd.textContent = e.type === 'd' ? '目录' : '文件';
        tr.appendChild(typeTd);
        const sizeTd = document.createElement('td');
        sizeTd.textContent = e.type === 'd' ? '-' : fmtSize(e.size);
        tr.appendChild(sizeTd);
        const timeTd = document.createElement('td');
        timeTd.textContent = fmtTime(e.mtime);
        tr.appendChild(timeTd);
        const opTd = document.createElement('td');
        opTd.className = 'op-cell';
        if (e.type === 'f') {
            opTd.appendChild(miniBtn('编辑', () => openFile(e.name)));
            opTd.appendChild(miniBtn('下载', () => downloadFile(e.name)));
        }
        opTd.appendChild(miniBtn('重命名', () => renameEntry(e.name)));
        opTd.appendChild(miniBtn('删除', () => deleteEntry(e.name), true));
        tr.appendChild(opTd);
        tbody.appendChild(tr);
    });
}

function miniBtn(text, onClick, danger) {
    const b = document.createElement('button');
    b.className = 'mini-btn' + (danger ? ' danger' : '');
    b.textContent = text;
    b.onclick = (e) => { e.stopPropagation(); onClick(); };
    return b;
}

async function openFile(name) {
    const n = state.currentNode;
    const path = joinPath(state.filePath, name);
    const r = await api('POST', '/nodes/' + n.id + '/file/read', { path });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    state.editingFile = { path, base64: r.data.content };
    document.getElementById('fileEditor').classList.remove('hidden');
    document.getElementById('editorPath').textContent = path;
    try {
        const text = new TextDecoder('utf-8', { fatal: false }).decode(base64ToBytes(r.data.content));
        document.getElementById('editorContent').value = text;
    } catch (e) {
        document.getElementById('editorContent').value = '';
        toast('二进制文件，无法直接编辑，可下载', 'info');
    }
}

async function saveFile() {
    if (!state.editingFile) return;
    const text = document.getElementById('editorContent').value;
    const b64 = btoa(unescape(encodeURIComponent(text)));
    const r = await api('POST', '/nodes/' + state.currentNode.id + '/file/write', { path: state.editingFile.path, content: b64 });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    toast('保存成功', 'success');
}

async function downloadFile(name) {
    const n = state.currentNode;
    const path = joinPath(state.filePath, name);
    const r = await api('POST', '/nodes/' + n.id + '/file/read', { path });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    const blob = new Blob([base64ToBytes(r.data.content)]);
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = name;
    a.click();
    URL.revokeObjectURL(a.href);
}

async function renameEntry(name) {
    const newName = await promptModal('重命名', '输入新名称', name);
    if (!newName || newName === name) return;
    const path = joinPath(state.filePath, name);
    const newPath = joinPath(state.filePath, newName);
    const r = await api('POST', '/nodes/' + state.currentNode.id + '/file/rename', { path, newPath });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    loadFiles(state.filePath);
}

async function deleteEntry(name) {
    if (!confirm('确定删除「' + name + '」？')) return;
    const path = joinPath(state.filePath, name);
    const r = await api('POST', '/nodes/' + state.currentNode.id + '/file/delete', { path });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    loadFiles(state.filePath);
}

async function mkdirFile() {
    const name = await promptModal('新建目录', '输入目录名');
    if (!name) return;
    const path = joinPath(state.filePath, name);
    const r = await api('POST', '/nodes/' + state.currentNode.id + '/file/mkdir', { path });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    loadFiles(state.filePath);
}

async function uploadFiles() {
    const input = document.getElementById('fileUploadInput');
    const files = input.files;
    if (!files.length) return;
    for (const f of files) {
        const buf = await f.arrayBuffer();
        const b64 = arrayBufferToBase64(buf);
        const path = joinPath(state.filePath, f.name);
        const r = await api('POST', '/nodes/' + state.currentNode.id + '/file/write', { path, content: b64 });
        if (r.data.error) toast('上传 ' + f.name + ' 失败: ' + r.data.error, 'error');
        else toast('上传 ' + f.name + ' 成功', 'success');
    }
    input.value = '';
    loadFiles(state.filePath);
}

// ===== 数据库 =====
async function runDb() {
    const n = state.currentNode;
    if (!n) { toast('请先选择节点', 'error'); return; }
    const sql = document.getElementById('dbSql').value.trim();
    if (!sql) { toast('SQL 不能为空', 'error'); return; }
    const r = await api('POST', '/nodes/' + n.id + '/db', {
        dbType: 'mysql',
        host: document.getElementById('dbHost').value,
        port: document.getElementById('dbPort').value,
        user: document.getElementById('dbUser').value,
        pass: document.getElementById('dbPass').value,
        name: document.getElementById('dbName').value,
        sql,
    });
    const msgEl = document.getElementById('dbResultMsg');
    const table = document.getElementById('dbResultTable');
    if (r.data.error) {
        msgEl.textContent = r.data.error;
        msgEl.className = 'db-result-msg error';
        table.style.display = 'none';
        return;
    }
    renderDbResult(r.data.result || '');
}

function renderDbResult(text) {
    const msgEl = document.getElementById('dbResultMsg');
    const table = document.getElementById('dbResultTable');
    const lines = text.replace(/\r/g, '').split('\n').filter((l) => l.length > 0);
    if (lines.length === 0) {
        msgEl.textContent = '执行成功（无返回结果）';
        msgEl.className = 'db-result-msg success';
        table.style.display = 'none';
        return;
    }
    const cols = lines[0].split('\t');
    const rows = lines.slice(1).map((l) => l.split('\t'));
    let html = '<thead><tr>' + cols.map((c) => '<th>' + escapeHtml(c) + '</th>').join('') + '</tr></thead><tbody>';
    rows.forEach((row) => {
        html += '<tr>' + cols.map((_, i) => '<td>' + escapeHtml(row[i] !== undefined ? row[i] : 'NULL') + '</td>').join('') + '</tr>';
    });
    html += '</tbody>';
    table.innerHTML = html;
    table.style.display = 'table';
    msgEl.textContent = '共 ' + rows.length + ' 行';
    msgEl.className = 'db-result-msg success';
}

// ===== 虚拟终端 =====
// 等待 DOM 布局完成后重新计算终端尺寸并强制重绘（解决切换回终端时空白的问题）
function refitTerm(termData) {
    if (!termData || !termData.term || !termData.fitAddon || !termData.el) return;
    const term = termData.term;
    const fitAddon = termData.fitAddon;

    // 先 fit 唤醒 renderer，再用「resize 抖动」强制重建 canvas，最后多次 refresh 兜底
    const redraw = () => {
        try {
            fitAddon.fit();
        } catch (e) { /* ignore */ }
        // 强制触发一次真实的 resize（列数 +1 再还原），让 renderer 完整重绘，
        // 比 refresh 更可靠，能覆盖从 visibility:hidden 恢复后 canvas 空白的情况
        try {
            if (term.cols > 0 && term.rows > 0) {
                term.resize(term.cols + 1, term.rows);
                term.resize(term.cols, term.rows);
            }
        } catch (e) { /* ignore */ }
        requestAnimationFrame(() => {
            try {
                if (term.rows > 0) term.refresh(0, term.rows - 1);
            } catch (e) {}
            setTimeout(() => {
                try {
                    if (term.rows > 0) term.refresh(0, term.rows - 1);
                } catch (e) {}
            }, 100);
            setTimeout(() => {
                try {
                    if (term.rows > 0) term.refresh(0, term.rows - 1);
                    if (term.cols > 0) {
                        term.resize(term.cols + 1, term.rows);
                        term.resize(term.cols, term.rows);
                    }
                } catch (e) {}
            }, 300);
        });
    };

    // 持续尝试到容器有非零尺寸为止，避免 fit 在 0 尺寸时静默失败
    let tries = 0;
    const tryFit = () => {
        const rect = termData.el.getBoundingClientRect();
        if (rect && rect.width > 0 && rect.height > 0) {
            redraw();
        } else if (tries++ < 60) {
            setTimeout(tryFit, 30);
        } else {
            redraw();
        }
    };
    requestAnimationFrame(() => requestAnimationFrame(tryFit));
    setTimeout(tryFit, 50);
    setTimeout(tryFit, 250);
}

// 为终端挂载 ResizeObserver：容器尺寸变化（含隐藏后恢复可见）时自动重算，彻底避免空白
function watchTermResize(termData, containerEl) {
    if (!termData || !containerEl || typeof ResizeObserver === 'undefined') return;
    try {
        const ro = new ResizeObserver(() => {
            if (termData.fitAddon) {
                try {
                    termData.fitAddon.fit();
                    const t = termData.term;
                    if (t && t.rows > 0) t.refresh(0, t.rows - 1);
                } catch (e) {}
            }
        });
        ro.observe(containerEl);
        termData._ro = ro;
    } catch (e) { /* ignore */ }
}

function ensureTerminal() {
    const n = state.currentNode;
    if (!n) { toast('请先选择节点', 'error'); return; }
    const container = document.getElementById('terminalWrap');
    if (!container._ready) {
        container.innerHTML = '';
        container._ready = true;
    }
    const existing = state.terms[n.id];
    if (existing) {
        document.querySelectorAll('.terminal-instance').forEach((el) => el.classList.remove('active'));
        const div = document.getElementById('term-' + n.id);
        if (div) div.classList.add('active');
        state.activeTermId = n.id;
        // 切回来时如果 WS 已断开，立即重连，避免空白
        if (existing.ws && existing.ws.readyState !== WebSocket.OPEN && existing.ws.readyState !== WebSocket.CONNECTING) {
            existing.term.write('\x1b[33m\r\n[重新连接中…]\x1b[0m\r\n');
            connectTermWS(n.id);
        }
        refitTerm(existing);
        return;
    }
    createTerminal(n);
}

function createTerminal(node) {
    const container = document.getElementById('terminalWrap');
    const div = document.createElement('div');
    div.id = 'term-' + node.id;
    div.className = 'terminal-instance active';
    container.appendChild(div);

    const term = new Terminal({
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, Consolas, monospace',
        cursorBlink: true,
        theme: getTerminalTheme(),
        scrollback: 10000,
    });

    let fitAddon;
    try {
        fitAddon = new FitAddon.FitAddon();
        term.loadAddon(fitAddon);
    } catch (e) {
        // FitAddon 未加载时继续
    }
    term.open(div);
    if (fitAddon) setTimeout(() => fitAddon.fit(), 30);

    const termData = { term, fitAddon, ws: null, buf: '', el: div };
    state.terms[node.id] = termData;
    state.activeTermId = node.id;

    term.onData((data) => handleTermInput(node.id, data));
    connectTermWS(node.id);
    watchTermResize(termData, div);
    return termData;
}

function connectTermWS(nodeId) {
    const termData = state.terms[nodeId];
    if (!termData) return;
    const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') +
        location.host + '/api/nodes/' + nodeId + '/terminal?token=' + encodeURIComponent(state.token);

    // 清理旧连接和心跳定时器
    if (termData._pingTimer) { clearInterval(termData._pingTimer); termData._pingTimer = null; }
    if (termData.ws) { try { termData.ws.onclose = null; termData.ws.close(); } catch (e) {} }

    const ws = new WebSocket(wsUrl);
    termData.ws = ws;
    termData._closed = false;
    ws.onopen = () => {
        termData.term.write('\x1b[32m[已连接]\x1b[0m\r\n');
        // 心跳保活：每 25s 发一个 ping，防止空闲被中间层断开
        termData._pingTimer = setInterval(() => {
            if (ws.readyState === WebSocket.OPEN) {
                try { ws.send(JSON.stringify({ type: 'ping' })); } catch (e) {}
            }
        }, 25000);
    };
    ws.onmessage = (e) => {
        try {
            const msg = JSON.parse(e.data);
            if (msg.type === 'output') {
                termData.term.write(String(msg.data).replace(/\n/g, '\r\n'));
            }
        } catch (err) { /* ignore */ }
    };
    ws.onclose = () => {
        if (termData._pingTimer) { clearInterval(termData._pingTimer); termData._pingTimer = null; }
        if (state.terms[nodeId] && !termData._closed) {
            termData.term.write('\x1b[33m\r\n[连接已断开，3 秒后重连…]\x1b[0m\r\n');
            // 自动重连：3 秒后重连，避免切回来时空白
            setTimeout(() => {
                if (state.terms[nodeId] && state.currentTab === 'terminal' && state.currentNode && state.currentNode.id === nodeId) {
                    connectTermWS(nodeId);
                }
            }, 3000);
        }
    };
    ws.onerror = () => { try { ws.close(); } catch (e) {} };
}

function handleTermInput(nodeId, data) {
    const termData = state.terms[nodeId];
    if (!termData || !termData.ws || termData.ws.readyState !== WebSocket.OPEN) return;
    const term = termData.term;
    for (const ch of data) {
        if (ch === '\r') {
            term.write('\r\n');
            const line = termData.buf;
            termData.buf = '';
            termData.ws.send(JSON.stringify({ type: 'input', data: line }));
        } else if (ch === '\n') {
            // 已由 \r 处理，忽略
        } else if (ch === '\x7f' || ch === '\b') {
            if (termData.buf.length > 0) {
                termData.buf = termData.buf.slice(0, -1);
                term.write('\b \b');
            }
        } else {
            termData.buf += ch;
            term.write(ch);
        }
    }
}

function destroyTerminal(nodeId) {
    const t = state.terms[nodeId];
    if (!t) return;
    if (t._ro) { try { t._ro.disconnect(); } catch (e) {} }
    if (t.ws) t.ws.close();
    if (t.term) t.term.dispose();
    delete state.terms[nodeId];
    const div = document.getElementById('term-' + nodeId);
    if (div) div.remove();
    if (state.activeTermId === nodeId) state.activeTermId = null;
}

// ===== 事件绑定 =====
function bindEvents() {
    document.getElementById('saveNodeBtn').onclick = saveNode;
    document.getElementById('clearNodeListBtn').onclick = clearNodeList;
    document.getElementById('connectBtn').onclick = testNode;

    // 用户下拉菜单
    const userDropdown = document.getElementById('userDropdown');
    userDropdown.onclick = (e) => {
        e.stopPropagation();
        userDropdown.classList.toggle('open');
    };
    document.addEventListener('click', () => userDropdown.classList.remove('open'));

    document.getElementById('logoutBtn').onclick = () => {
        if (confirm('确定退出登录？')) logout(false);
    };
    document.getElementById('changePassMenuItem').onclick = (e) => {
        e.stopPropagation();
        userDropdown.classList.remove('open');
        openModal('changePassModal');
    };

    // 顶部搜索框联动节点搜索
    document.getElementById('headerSearch').oninput = debounce((e) => {
        const nodeSearch = document.getElementById('nodeSearch');
        nodeSearch.value = e.target.value;
        renderNodeList();
    }, 150);

    // 通知按钮
    document.getElementById('notifyBtn').onclick = () => toast('暂无新通知', 'info');

    // 外观设置下拉菜单
    const themeDropdown = document.getElementById('themeDropdown');
    document.getElementById('settingsBtn').onclick = (e) => {
        e.stopPropagation();
        themeDropdown.classList.toggle('open');
    };
    document.addEventListener('click', () => themeDropdown.classList.remove('open'));
    document.querySelectorAll('.theme-item').forEach((item) => {
        item.onclick = (e) => {
            e.stopPropagation();
            applyTheme(item.dataset.theme);
            persistTheme(item.dataset.theme);
            themeDropdown.classList.remove('open');
            toast('已切换为「' + (THEME_NAMES[item.dataset.theme] || item.dataset.theme) + '」', 'success');
        };
    });

    document.querySelectorAll('.nav-item').forEach((t) => {
        t.onclick = () => {
            if (t.dataset.view === 'dashboard') switchToDashboard();
            else if (t.dataset.view === 'connections') switchToConnections();
            else if (t.dataset.view === 'reverse-shell') switchToReverseShell();
            else if (t.dataset.view === 'servers') switchToServers();
            else if (t.dataset.view === 'server-term') switchToServerTerm();
            else if (t.dataset.view === 'system') switchToSystem();
            else if (t.dataset.tab) switchTab(t.dataset.tab);
        };
    });

    document.getElementById('nodeSearch').oninput = debounce(renderNodeList, 150);

    // 连接列表
    document.getElementById('connSearch').oninput = debounce(renderConnList, 150);
    document.getElementById('connAddNodeBtn').onclick = () => showNodeModal(null);
    document.getElementById('batchTestAllBtn').onclick = () => {
        const allIds = state.nodes.map((n) => n.id);
        batchTest(allIds);
    };

    // 资源管理 - 服务器
    document.getElementById('srvSearch').oninput = debounce(renderServerList, 150);
    document.getElementById('srvAddBtn').onclick = () => showServerModal(null);
    document.getElementById('saveSrvBtn').onclick = saveServer;
    document.getElementById('srvBatchTestBtn').onclick = batchTestServers;
    // 密码显示/隐藏切换
    document.querySelectorAll('.pwd-toggle').forEach((btn) => {
        btn.onclick = () => {
            const input = document.getElementById(btn.dataset.target);
            if (!input) return;
            const show = input.type === 'password';
            input.type = show ? 'text' : 'password';
            btn.querySelector('i').className = show ? 'fas fa-eye-slash' : 'fas fa-eye';
        };
    });
    document.querySelectorAll('.conn-filter-tab').forEach((tab) => {
        tab.onclick = () => {
            document.querySelectorAll('.conn-filter-tab').forEach((t) => t.classList.remove('active'));
            tab.classList.add('active');
            state.connFilter = tab.dataset.filter;
            renderConnList();
        };
    });

    // 反弹Shell
    document.getElementById('startListenerBtn').onclick = startListener;
    document.getElementById('genPayloadBtn').onclick = genPayloads;
    document.getElementById('clearOutputBtn').onclick = () => {
        if (state.activeListenerId) {
            document.getElementById('listenerOutput').textContent = '';
        }
    };
    document.querySelectorAll('.rs-type-filter').forEach((btn) => {
        btn.onclick = () => {
            document.querySelectorAll('.rs-type-filter').forEach((b) => b.classList.remove('active'));
            btn.classList.add('active');
            state.payloadFilter = btn.dataset.type;
            renderPayloads();
        };
    });

    // 命令执行
    document.getElementById('cmdInput').onkeydown = (e) => { if (e.key === 'Enter') runCmd(); };
    document.getElementById('clearCmdBtn').onclick = clearCmdHistory;

    // 文件管理
    document.getElementById('fileUpBtn').onclick = () => loadFiles(dirOf(state.filePath));
    document.getElementById('filePath').onkeydown = (e) => { if (e.key === 'Enter') loadFiles(e.target.value.trim()); };
    document.getElementById('fileRefreshBtn').onclick = () => loadFiles(state.filePath);
    document.getElementById('fileNewFolderBtn').onclick = mkdirFile;
    document.getElementById('fileUploadInput').onchange = uploadFiles;
    document.getElementById('saveFileBtn').onclick = saveFile;
    document.getElementById('closeEditorBtn').onclick = () => {
        document.getElementById('fileEditor').classList.add('hidden');
        state.editingFile = null;
        loadFiles(state.filePath);
    };

    // 数据库
    document.getElementById('dbRunBtn').onclick = runDb;
    document.getElementById('dbConnectBtn').onclick = runDb;
    document.getElementById('dbSql').onkeydown = (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') runDb();
    };

    // 终端自适应
    const refitActive = debounce(() => {
        const t = state.terms[state.activeTermId];
        if (t) refitTerm(t);
        const st = state.srvTerms[state.activeSrvTermId];
        if (st) refitTerm(st);
    }, 100);
    window.addEventListener('resize', refitActive);
    // 切回浏览器标签页 / 窗口重新聚焦时也重绘终端，避免空白
    document.addEventListener('visibilitychange', () => {
        if (!document.hidden) refitActive();
    });
    window.addEventListener('focus', refitActive);

    // 系统管理
    document.getElementById('sysSaveSettingsBtn').onclick = async () => {
        const timeout = parseInt(document.getElementById('sysSessionTimeout').value) || 30;
        const shell = document.getElementById('sysShell').value.trim();
        const r = await api('PUT', '/system/settings', { session_timeout: timeout, shell: shell });
        if (r.status === 401) return;
        if (r.status >= 400) { toast(r.data.error || '保存失败', 'error'); return; }
        toast('设置已保存', 'success');
    };

    document.getElementById('sysChangePassBtn').onclick = async () => {
        const oldP = document.getElementById('sysOldPass').value;
        const newP = document.getElementById('sysNewPass').value;
        const confirmP = document.getElementById('sysConfirmPass').value;
        if (!oldP || !newP) { toast('请填写完整', 'warning'); return; }
        if (newP.length < 6) { toast('新密码至少 6 位', 'warning'); return; }
        if (newP !== confirmP) { toast('两次输入的新密码不一致', 'warning'); return; }
        const r = await api('POST', '/system/password', { old_password: oldP, new_password: newP });
        if (r.status === 401) return;
        if (r.status >= 400) { toast(r.data.error || '修改失败', 'error'); return; }
        toast(r.data.message || '密码已修改', 'success');
        document.getElementById('sysOldPass').value = '';
        document.getElementById('sysNewPass').value = '';
        document.getElementById('sysConfirmPass').value = '';
        closeModal('changePassModal');
    };

    document.getElementById('sysExportBtn').onclick = async () => {
        const r = await api('GET', '/system/export');
        if (r.status === 401) return;
        const blob = new Blob([JSON.stringify(r.data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'mayfly_backup_' + new Date().toISOString().slice(0, 10) + '.json';
        a.click();
        URL.revokeObjectURL(url);
        toast('数据已导出', 'success');
    };

    document.getElementById('sysImportInput').onchange = async (e) => {
        const file = e.target.files[0];
        if (!file) return;
        try {
            const text = await file.text();
            const json = JSON.parse(text);
            const data = json.data || json;
            const r = await api('POST', '/system/import', { data: data });
            if (r.status === 401) return;
            if (r.status >= 400) { toast(r.data.error || '导入失败', 'error'); return; }
            toast(r.data.message || '导入成功', 'success');
            loadSystemInfo();
        } catch (err) {
            toast('文件解析失败: ' + err.message, 'error');
        }
        e.target.value = '';
    };

    document.getElementById('sysRefreshAuditBtn').onclick = loadSystemAuditLogs;

    document.getElementById('sysClearAuditBtn').onclick = async () => {
        if (!confirm('确定清空所有审计日志？')) return;
        const r = await api('DELETE', '/system/audit-logs');
        if (r.status === 401) return;
        toast('审计日志已清空', 'success');
        loadSystemAuditLogs();
        loadSystemInfo();
    };
}

// ===== 主题切换 =====
const THEME_KEY = 'mayfly_theme';
const THEME_NAMES = { glass: '系统默认效果', frosted: '白天玻璃效果', dark: '黑夜玻璃效果' };

function currentTheme() {
    const t = localStorage.getItem(THEME_KEY);
    return (t === 'frosted' || t === 'glass' || t === 'dark') ? t : 'glass';
}

function getTerminalTheme() {
    const theme = document.documentElement.getAttribute('data-theme') || 'glass';
    if (theme === 'dark') {
        return {
            background: '#0d1117', foreground: '#e6edf3', cursor: '#e6edf3',
            selectionBackground: 'rgba(86, 156, 214, 0.35)',
            black: '#0d1117', red: '#f48771', green: '#89d185', yellow: '#dcdcaa',
            blue: '#569cd6', magenta: '#c586c0', cyan: '#4ec9b0', white: '#e6edf3',
        };
    }
    return {
        background: '#fafbfd', foreground: '#1f2430', cursor: '#1f2430',
        selectionBackground: 'rgba(49, 109, 202, 0.3)',
        black: '#1f2430', red: '#d64550', green: '#1a7f37', yellow: '#b08800',
        blue: '#316dca', magenta: '#8250df', cyan: '#1b7c83', white: '#fafbfd',
    };
}

function updateTerminalsTheme() {
    const t = getTerminalTheme();
    Object.values(state.terms).forEach((item) => {
        if (item.term) { item.term.options.theme = t; item.term.refresh(0, item.term.rows - 1); }
    });
    Object.values(state.srvTerms).forEach((item) => {
        if (item.term) { item.term.options.theme = t; item.term.refresh(0, item.term.rows - 1); }
    });
}

function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    document.querySelectorAll('.theme-item').forEach((el) => {
        el.classList.toggle('active', el.dataset.theme === theme);
    });
    localStorage.setItem(THEME_KEY, theme);
    updateTerminalsTheme();
}

// 将主题保存到服务端（持久化到 data/settings.json）
function persistTheme(theme) {
    api('PUT', '/system/theme', { theme }).catch(() => {});
}

// 登录后从服务端同步主题，覆盖本地缓存
async function syncThemeFromServer() {
    try {
        const r = await api('GET', '/system/settings');
        if (r.status === 401) return;
        const theme = r.data.theme;
        if (theme === 'frosted' || theme === 'glass' || theme === 'dark') {
            applyTheme(theme);
        }
    } catch (e) { /* 服务端未返回主题时保留本地缓存 */ }
}

function initTheme() {
    applyTheme(currentTheme());
}

// 浏览器刷新/后退时会自动恢复输入框内未提交的历史内容，
// 这里清空搜索框、命令框等，避免出现"自动输入文字"的假象。
function clearAutoFilledInputs() {
    ['headerSearch', 'nodeSearch', 'connSearch', 'srvSearch', 'cmdInput'].forEach((id) => {
        const el = document.getElementById(id);
        if (el) el.value = '';
    });
}

// ===== 初始化 =====
async function init() {
    if (!state.token) { window.location.href = '/login'; return; }
    initTheme();
    syncThemeFromServer();
    document.getElementById('username').textContent = state.user;
    // 阻止浏览器刷新时自动恢复输入框历史内容（搜索框/命令框被自动填入）
    clearAutoFilledInputs();
    setTimeout(clearAutoFilledInputs, 200);
    bindEvents();
    renderQuickCmds();
    try {
        await loadNodes(true);
        restoreUI();
    } catch (e) { /* 401 会在 api() 中处理跳转 */ }
}

document.addEventListener('DOMContentLoaded', init);

// ===== 反弹Shell =====
async function loadListeners() {
    const r = await api('GET', '/listeners');
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    state.listeners = r.data.listeners || [];
    renderListeners();
}

function renderListeners() {
    const list = document.getElementById('listenerList');
    if (state.listeners.length === 0) {
        list.innerHTML = '<div style="text-align:center;padding:20px;color:var(--text-muted);font-size:13px;"><i class="fas fa-info-circle"></i> 暂无监听，输入端口开启监听</div>';
        return;
    }
    list.innerHTML = '';
    state.listeners.forEach((l) => {
        const item = document.createElement('div');
        item.className = 'rs-listener-item' + (l.id === state.activeListenerId ? ' active' : '');
        item.innerHTML = `
            <div class="rs-listener-info">
                <span class="rs-listener-port">:${l.port}</span>
                <span class="rs-listener-status ${l.status}">${l.status === 'listening' ? '监听中' : '已停止'}</span>
                <span style="font-size:11px;color:var(--text-muted);">${l.protocol.toUpperCase()}</span>
            </div>
            <div class="rs-listener-actions">
                ${l.status === 'listening' ? `<button class="conn-mini-btn stop-listener-btn" title="停止"><i class="fas fa-stop"></i></button>` : ''}
                <button class="conn-mini-btn danger del-listener-btn" title="删除"><i class="fas fa-trash"></i></button>
            </div>
        `;
        item.onclick = (e) => {
            if (e.target.closest('.stop-listener-btn') || e.target.closest('.del-listener-btn')) return;
            state.activeListenerId = l.id;
            renderListeners();
            pollListenerOutput();
        };
        const stopBtn = item.querySelector('.stop-listener-btn');
        if (stopBtn) stopBtn.onclick = async (e) => { e.stopPropagation(); await stopListener(l.id); };
        item.querySelector('.del-listener-btn').onclick = async (e) => { e.stopPropagation(); await deleteListener(l.id); };
        list.appendChild(item);
    });
}

async function startListener() {
    const port = parseInt(document.getElementById('listenerPort').value);
    if (!port || port < 1 || port > 65535) { toast('端口号无效（1-65535）', 'error'); return; }
    const r = await api('POST', '/listeners', { port, protocol: 'tcp' });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    toast(`监听端口 ${port} 已开启`, 'success');
    state.activeListenerId = r.data.id;
    await loadListeners();
    startPolling();
}

async function stopListener(id) {
    const r = await api('POST', `/listeners/${id}/stop`);
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    toast('监听已停止', 'success');
    await loadListeners();
}

async function deleteListener(id) {
    if (!confirm('确定删除此监听？')) return;
    const r = await api('DELETE', `/listeners/${id}`);
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    if (state.activeListenerId === id) state.activeListenerId = null;
    await loadListeners();
    document.getElementById('listenerOutput').innerHTML = '<div class="rs-output-placeholder"><i class="fas fa-satellite-dish"></i><p>选择一个监听端口查看输出</p></div>';
    toast('已删除', 'success');
}

function startPolling() {
    if (state.listenerPollTimer) clearInterval(state.listenerPollTimer);
    state.listenerPollTimer = setInterval(pollListenerOutput, 2000);
}

function stopPolling() {
    if (state.listenerPollTimer) { clearInterval(state.listenerPollTimer); state.listenerPollTimer = null; }
}

async function pollListenerOutput() {
    if (!state.activeListenerId) return;
    const r = await api('GET', `/listeners/${state.activeListenerId}/output`);
    if (r.data.error) { stopPolling(); return; }
    const output = r.data.output || '';
    const el = document.getElementById('listenerOutput');
    if (output) {
        el.textContent = output;
        el.scrollTop = el.scrollHeight;
    } else {
        el.innerHTML = '<div class="rs-output-placeholder"><i class="fas fa-clock"></i><p>等待连接...</p></div>';
    }
}

// 生成反弹Shell命令
async function genPayloads() {
    const ip = document.getElementById('payloadIP').value.trim();
    const port = document.getElementById('payloadPort').value.trim();
    if (!ip || !port) { return; }
    const r = await api('GET', `/reverse-shells?ip=${encodeURIComponent(ip)}&port=${encodeURIComponent(port)}`);
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    state.payloads = r.data.payloads || [];
    renderPayloads();
}

function renderPayloads() {
    const list = document.getElementById('payloadList');
    const filtered = state.payloads.filter((p) => {
        if (state.payloadFilter === 'all') return true;
        if (state.payloadFilter === 'other') return !['bash','nc','python','php','perl','powershell'].includes(p.type);
        return p.type === state.payloadFilter;
    });
    if (filtered.length === 0) {
        list.innerHTML = '<div style="text-align:center;padding:20px;color:var(--text-muted);font-size:13px;">输入IP和端口后点击生成</div>';
        return;
    }
    list.innerHTML = '';
    filtered.forEach((p, i) => {
        const item = document.createElement('div');
        item.className = 'rs-payload-item';
        item.innerHTML = `
            <div class="rs-payload-header">
                <span class="rs-payload-name">${escapeHtml(p.name)} <span class="rs-payload-type-badge">${p.type}</span></span>
                <div class="rs-payload-actions">
                    <button class="rs-payload-copy-btn" title="复制"><i class="fas fa-copy"></i> 复制</button>
                </div>
            </div>
            <div class="rs-payload-code">${escapeHtml(p.cmd)}</div>
        `;
        item.querySelector('.rs-payload-copy-btn').onclick = () => {
            navigator.clipboard.writeText(p.cmd).then(() => toast('已复制到剪贴板', 'success'));
        };
        list.appendChild(item);
    });
}

function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => toast('已复制', 'success'));
}

// ===== 资源管理 - 服务器 =====
async function loadServers() {
    const r = await api('GET', '/servers');
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    state.servers = r.data.servers || [];
    state.serverGroups = r.data.groups || [];
    renderServerList();
}

function renderServerList() {
    const container = document.getElementById('srvGroups');
    const keyword = (document.getElementById('srvSearch').value || '').toLowerCase();

    let filtered = state.servers;
    if (keyword) {
        filtered = filtered.filter((s) =>
            (s.name || '').toLowerCase().includes(keyword) ||
            (s.host || '').toLowerCase().includes(keyword) ||
            (s.group || '').toLowerCase().includes(keyword) ||
            (s.username || '').toLowerCase().includes(keyword)
        );
    }

    if (filtered.length === 0) {
        container.innerHTML = `
            <div class="conn-empty">
                <i class="fas fa-server"></i>
                <p>${keyword ? '没有匹配的服务器' : '暂无服务器，点击「添加服务器」开始管理'}</p>
            </div>`;
        return;
    }

    // 按分组组织
    const groupMap = {};
    filtered.forEach((s) => {
        const g = s.group || '默认';
        if (!groupMap[g]) groupMap[g] = [];
        groupMap[g].push(s);
    });

    let html = '';
    Object.keys(groupMap).sort().forEach((g) => {
        const items = groupMap[g];
        html += `
            <div class="conn-group">
                <div class="conn-group-header" onclick="this.parentElement.classList.toggle('collapsed')">
                    <div class="conn-group-info">
                        <i class="fas fa-chevron-right conn-group-arrow"></i>
                        <i class="fas fa-folder conn-group-icon"></i>
                        <span class="conn-group-name">${escapeHtml(g)}</span>
                        <span class="conn-group-count">${items.length}</span>
                    </div>
                </div>
                <div class="conn-group-body">
                    ${items.map((s) => renderServerItem(s)).join('')}
                </div>
            </div>`;
    });
    container.innerHTML = html;

    // 绑定事件
    container.querySelectorAll('.srv-action-btn[data-act]').forEach((btn) => {
        btn.onclick = (e) => {
            e.stopPropagation();
            const id = parseInt(btn.dataset.id);
            const act = btn.dataset.act;
            if (act === 'edit') showServerModal(id);
            else if (act === 'delete') deleteServer(id);
            else if (act === 'test') testServer(id);
            else if (act === 'term') openServerTerminal(id);
        };
    });
}

function formatTestTime(iso) {
    if (!iso) return '';
    const d = new Date(iso);
    if (isNaN(d.getTime())) return '';
    const p = (n) => String(n).padStart(2, '0');
    return `${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`;
}

function renderServerItem(s) {
    const testing = s._testStatus === 'testing';
    const status = testing ? 'testing' : (s.last_test_status || 'untested');
    const statusText = { ok: '已连通', fail: '连接失败', testing: '测试中...', untested: '未测试' }[status];
    const port = s.port || 22;
    const testTime = s.last_test_time ? formatTestTime(s.last_test_time) : '';
    const testTip = s.last_test_message ? `title="${escapeHtml(s.last_test_message)}"` : '';
    const desc = s.description ? `<div class="srv-desc" title="${escapeHtml(s.description)}">${escapeHtml(s.description)}</div>` : '';
    return `
        <div class="srv-item ${status}">
            <div class="srv-head">
                <div class="srv-icon"><i class="fas fa-server"></i></div>
                <div class="srv-title">
                    <div class="srv-name" title="${escapeHtml(s.name || s.host)}">${escapeHtml(s.name || s.host)}</div>
                    <div class="srv-addr"><span class="mono">${escapeHtml(s.host)}:${port}</span><span class="dot">·</span><span>${escapeHtml(s.username || '未设置用户')}</span></div>
                </div>
                <span class="srv-status ${status}" ${testTip}>${statusText}</span>
            </div>
            ${desc}
            <div class="srv-foot">
                <span class="srv-test-time"><i class="fas fa-clock"></i>${testTime ? '上次测试 ' + testTime : '尚未测试'}</span>
                <div class="srv-actions">
                    <button class="srv-action-btn" data-act="term" data-id="${s.id}"><i class="fas fa-terminal"></i>终端</button>
                    <button class="srv-action-btn success" data-act="test" data-id="${s.id}"><i class="fas fa-plug"></i>测试</button>
                    <button class="srv-action-btn" data-act="edit" data-id="${s.id}"><i class="fas fa-edit"></i>编辑</button>
                    <button class="srv-action-btn danger" data-act="delete" data-id="${s.id}"><i class="fas fa-trash"></i>删除</button>
                </div>
            </div>
        </div>`;
}

function showServerModal(id) {
    const modal = document.getElementById('serverModal');
    const title = document.getElementById('serverModalTitle');
    if (id) {
        const s = state.servers.find((x) => x.id === id);
        if (!s) return;
        title.textContent = '编辑服务器';
        document.getElementById('srvName').value = s.name || '';
        document.getElementById('srvGroup').value = s.group || '';
        document.getElementById('srvHost').value = s.host || '';
        document.getElementById('srvPort').value = s.port || 22;
        document.getElementById('srvUsername').value = s.username || '';
        document.getElementById('srvPassword').value = s.password || '';
        document.getElementById('srvPrivateKey').value = s.private_key || '';
        document.getElementById('srvDesc').value = s.description || '';
        modal.dataset.editId = id;
    } else {
        title.textContent = '添加服务器';
        document.getElementById('srvName').value = '';
        document.getElementById('srvGroup').value = '';
        document.getElementById('srvHost').value = '';
        document.getElementById('srvPort').value = '22';
        document.getElementById('srvUsername').value = '';
        document.getElementById('srvPassword').value = '';
        document.getElementById('srvPrivateKey').value = '';
        document.getElementById('srvDesc').value = '';
        delete modal.dataset.editId;
    }
    modal.style.display = 'flex';
}

async function saveServer() {
    const modal = document.getElementById('serverModal');
    const editId = modal.dataset.editId;
    const data = {
        name: document.getElementById('srvName').value.trim(),
        group: document.getElementById('srvGroup').value.trim(),
        host: document.getElementById('srvHost').value.trim(),
        port: parseInt(document.getElementById('srvPort').value) || 22,
        username: document.getElementById('srvUsername').value.trim(),
        password: document.getElementById('srvPassword').value,
        private_key: document.getElementById('srvPrivateKey').value.trim(),
        description: document.getElementById('srvDesc').value.trim(),
    };
    if (!data.host || !data.username) { toast('IP地址和用户名不能为空', 'error'); return; }

    if (editId) {
        data.id = parseInt(editId);
        const r = await api('PUT', '/servers', data);
        if (r.data.error) { toast(r.data.error, 'error'); return; }
        toast('更新成功', 'success');
    } else {
        const r = await api('POST', '/servers', data);
        if (r.data.error) { toast(r.data.error, 'error'); return; }
        toast('创建成功', 'success');
    }
    closeModal('serverModal');
    await loadServers();
}

async function deleteServer(id) {
    if (!confirm('确定删除此服务器？')) return;
    const r = await api('DELETE', '/servers', { id });
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    toast('已删除', 'success');
    await loadServers();
}

async function testServer(id) {
    const s = state.servers.find((x) => x.id === id);
    if (!s) return;
    s._testStatus = 'testing';
    renderServerList();
    const r = await api('POST', '/servers/test', {
        server_id: id,
        host: s.host, port: s.port, username: s.username,
        password: s.password, private_key: s.private_key,
    });
    if (r.data.success) {
        toast(`连接成功 — 主机名: ${r.data.hostname}`, 'success');
    } else {
        toast(r.data.message || '连接失败', 'error');
    }
    // 后端已持久化测试结果，重新加载以渲染最新状态与时间
    await loadServers();
}

async function batchTestServers() {
    if (state.servers.length === 0) { toast('暂无服务器', 'warning'); return; }
    for (const s of state.servers) {
        s._testStatus = 'testing';
    }
    renderServerList();
    for (const s of state.servers) {
        const r = await api('POST', '/servers/test', {
            server_id: s.id,
            host: s.host, port: s.port, username: s.username,
            password: s.password, private_key: s.private_key,
        });
        s._testStatus = r.data.success ? 'ok' : 'fail';
        renderServerList();
    }
    await loadServers();
    const ok = state.servers.filter((s) => s.last_test_status === 'ok').length;
    toast(`测试完成：${ok}/${state.servers.length} 连通`, ok === state.servers.length ? 'success' : 'warning');
}
