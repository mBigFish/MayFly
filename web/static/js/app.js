// ===== Mayfly WebShell 管理器前端 =====

const state = {
    token: localStorage.getItem('mayfly_token'),
    user: localStorage.getItem('mayfly_user') || 'admin',
    nodes: [],
    currentNode: null,
    currentTab: 'file',
    filePath: '',
    editingFile: null,
    terms: {},
    activeTermId: null,
};

// ===== 工具函数 =====
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
    const t = document.createElement('div');
    t.className = 'toast ' + (type || 'info');
    t.innerHTML = '<i class="fas ' + (type === 'error' ? 'fa-times-circle' : type === 'success' ? 'fa-check-circle' : type === 'warning' ? 'fa-exclamation-circle' : 'fa-info-circle') + '"></i><span>' + escapeHtml(msg) + '</span>';
    wrap.appendChild(t);
    setTimeout(() => t.remove(), 2600);
}

function logout(expired) {
    localStorage.removeItem('mayfly_token');
    localStorage.removeItem('mayfly_user');
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
async function loadNodes() {
    const r = await api('GET', '/nodes');
    state.nodes = r.data.nodes || [];
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
    const nodes = q ? state.nodes.filter((n) => n.name.toLowerCase().includes(q) || n.url.toLowerCase().includes(q)) : state.nodes;
    if (nodes.length === 0) {
        list.innerHTML = '<div class="empty-state"><i class="fas fa-inbox"></i><p>' + (state.nodes.length ? '无匹配节点' : '暂无节点，点击右上角添加') + '</p></div>';
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

function selectNode(node) {
    state.currentNode = node;
    state.filePath = '';
    state.editingFile = null;
    document.getElementById('fileEditor').classList.add('hidden');
    document.getElementById('currentNodeTitle').innerHTML =
        '<span class="type-badge type-' + node.type + '">' + node.type.toUpperCase() + '</span> ' + escapeHtml(node.name);
    renderNodeList();
    updatePanelsForNode();
}

function updatePanelsForNode() {
    if (state.currentTab === 'file') loadFiles('');
    else if (state.currentTab === 'cmd') document.getElementById('cmdOutput').textContent = '';
    else if (state.currentTab === 'terminal') ensureTerminal();
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
    document.getElementById('nodePass').value = node ? node.pass : '';
    document.getElementById('nodeRemark').value = node ? node.remark : '';
    openModal('nodeModal');
}

async function saveNode() {
    const body = {
        name: document.getElementById('nodeName').value.trim(),
        url: document.getElementById('nodeURL').value.trim(),
        type: document.getElementById('nodeType').value,
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

async function deleteNode() {
    const n = state.currentNode;
    if (!n) { toast('请先选择节点', 'error'); return; }
    if (!confirm('确定删除节点「' + n.name + '」？')) return;
    const r = await api('DELETE', '/nodes/' + n.id);
    if (r.data.error) { toast(r.data.error, 'error'); return; }
    if (state.terms[n.id]) destroyTerminal(n.id);
    state.currentNode = null;
    document.getElementById('currentNodeTitle').innerHTML = '<span class="muted">未选择节点</span>';
    await loadNodes();
    toast('已删除', 'success');
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

// ===== 标签页 =====
function switchTab(tab) {
    state.currentTab = tab;
    document.querySelectorAll('.nav-item').forEach((t) => t.classList.toggle('active', t.dataset.tab === tab));
    document.querySelectorAll('.tab-panel').forEach((p) => p.classList.remove('active'));
    document.getElementById('panel-' + tab).classList.add('active');
    updatePanelsForNode();
    if (tab === 'terminal') ensureTerminal();
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
    const r = await api('POST', '/nodes/' + n.id + '/cmd', { cmd });
    if (r.data.error) {
        out.textContent += '[错误] ' + r.data.error + '\n';
    } else {
        out.textContent += r.data.output || '(无输出)';
        if (r.data.output && !r.data.output.endsWith('\n')) out.textContent += '\n';
    }
    out.scrollTop = out.scrollHeight;
    input.value = '';
    input.focus();
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
        setTimeout(() => existing.fitAddon.fit(), 30);
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
        theme: {
            background: '#1e1e1e', foreground: '#d4d4d4', cursor: '#d4d4d4',
            black: '#1e1e1e', red: '#f48771', green: '#89d185', yellow: '#dcdcaa',
            blue: '#569cd6', magenta: '#c586c0', cyan: '#4ec9b0', white: '#d4d4d4',
        },
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

    const termData = { term, fitAddon, ws: null, buf: '' };
    state.terms[node.id] = termData;
    state.activeTermId = node.id;

    term.onData((data) => handleTermInput(node.id, data));
    connectTermWS(node.id);
    return termData;
}

function connectTermWS(nodeId) {
    const termData = state.terms[nodeId];
    if (!termData) return;
    const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') +
        location.host + '/api/nodes/' + nodeId + '/terminal?token=' + encodeURIComponent(state.token);
    const ws = new WebSocket(wsUrl);
    termData.ws = ws;
    ws.onopen = () => { termData.term.write('\x1b[32m[已连接]\x1b[0m\r\n'); };
    ws.onmessage = (e) => {
        try {
            const msg = JSON.parse(e.data);
            if (msg.type === 'output') {
                termData.term.write(String(msg.data).replace(/\n/g, '\r\n'));
            }
        } catch (err) { /* ignore */ }
    };
    ws.onclose = () => {
        if (state.terms[nodeId]) termData.term.write('\x1b[33m\r\n[连接已断开]\x1b[0m\r\n');
    };
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
    if (t.ws) t.ws.close();
    if (t.term) t.term.dispose();
    delete state.terms[nodeId];
    const div = document.getElementById('term-' + nodeId);
    if (div) div.remove();
    if (state.activeTermId === nodeId) state.activeTermId = null;
}

// ===== 事件绑定 =====
function bindEvents() {
    document.getElementById('addNodeBtn').onclick = () => showNodeModal(null);
    document.getElementById('saveNodeBtn').onclick = saveNode;
    document.getElementById('deleteNodeBtn').onclick = deleteNode;
    document.getElementById('connectBtn').onclick = testNode;
    document.getElementById('logoutBtn').onclick = () => {
        if (confirm('确定退出登录？')) logout(false);
    };

    document.querySelectorAll('.nav-item').forEach((t) => {
        t.onclick = () => switchTab(t.dataset.tab);
    });

    document.getElementById('nodeSearch').oninput = renderNodeList;

    // 命令执行
    document.getElementById('cmdInput').onkeydown = (e) => { if (e.key === 'Enter') runCmd(); };

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
    window.addEventListener('resize', () => {
        const t = state.terms[state.activeTermId];
        if (t && t.fitAddon) t.fitAddon.fit();
    });
}

// ===== 初始化 =====
async function init() {
    if (!state.token) { window.location.href = '/login'; return; }
    document.getElementById('username').textContent = state.user;
    bindEvents();
    try {
        await loadNodes();
    } catch (e) { /* 401 会在 api() 中处理跳转 */ }
}

document.addEventListener('DOMContentLoaded', init);
