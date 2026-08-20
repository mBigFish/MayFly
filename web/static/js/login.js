// 登录页逻辑
(function() {
    // 如果已有 token，跳转到主页
    const token = localStorage.getItem('mayfly_token');
    if (token) {
        window.location.href = '/';
        return;
    }

    const form = document.getElementById('loginForm');
    const errorMsg = document.getElementById('errorMsg');
    const loginBtn = document.getElementById('loginBtn');
    const btnText = loginBtn.querySelector('.btn-text');
    const btnLoading = loginBtn.querySelector('.btn-loading');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;

        // 显示加载状态
        btnText.style.display = 'none';
        btnLoading.style.display = 'inline-flex';
        loginBtn.disabled = true;
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
            btnText.style.display = '';
            btnLoading.style.display = 'none';
            loginBtn.disabled = false;
        }
    });
})();
