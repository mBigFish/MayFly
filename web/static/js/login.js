// 登录页逻辑
(function() {
    // 如果已有 token，跳转到主页
    const token = localStorage.getItem('mayfly_token');
    if (token) {
        window.location.href = '/';
        return;
    }

    const form = document.getElementById('login-form');
    const errorMsg = document.getElementById('login-error');
    const loginBtn = document.getElementById('login-btn');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;

        // 显示加载状态
        loginBtn.disabled = true;
        loginBtn.textContent = '登录中...';
        errorMsg.style.display = 'none';

        try {
            const res = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
            });

            const data = await res.json();

            if (res.ok && data.token) {
                localStorage.setItem('mayfly_token', data.token);
                localStorage.setItem('mayfly_user', username);
                window.location.href = '/';
            } else {
                errorMsg.textContent = data.error || '登录失败';
                errorMsg.style.display = 'block';
            }
        } catch (err) {
            errorMsg.textContent = '网络错误，请检查连接';
            errorMsg.style.display = 'block';
        } finally {
            loginBtn.disabled = false;
            loginBtn.textContent = '登录';
        }
    });
})();
