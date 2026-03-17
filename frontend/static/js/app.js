const API_BASE = '/api/v1';

function showResult(elementId, message, isError = false) {
    const element = document.getElementById(elementId);
    element.textContent = message;
    element.className = 'result-box show';
    if (isError) {
        element.classList.add('error');
    } else {
        element.classList.add('success');
    }
}

function clearResult(elementId) {
    const element = document.getElementById(elementId);
    element.className = 'result-box';
}

// 加载系统信息
async function loadSystemInfo() {
    try {
        const response = await fetch('/healthz');
        const data = await response.json();
        document.getElementById('systemInfo').innerHTML = `
            <p><strong>状态:</strong> ${data.status || 'ok'}</p>
            <p><strong>时间:</strong> ${new Date().toLocaleString()}</p>
        `;
    } catch (error) {
        document.getElementById('systemInfo').innerHTML = `
            <p style="color: red;">系统连接失败</p>
        `;
    }
}

// 用户注册
async function register() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!username || !password) {
        showResult('userResult', '请输入用户名和密码', true);
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/users/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password }),
        });
        const data = await response.json();
        if (response.ok) {
            showResult('userResult', `注册成功!\n${JSON.stringify(data, null, 2)}`);
        } else {
            showResult('userResult', `注册失败: ${data.message || data.error}`, true);
        }
    } catch (error) {
        showResult('userResult', `请求失败: ${error.message}`, true);
    }
}

// 用户登录
async function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!username || !password) {
        showResult('userResult', '请输入用户名和密码', true);
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/users/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password }),
        });
        const data = await response.json();
        if (response.ok) {
            showResult('userResult', `登录成功!\nToken: ${data.token?.substring(0, 20)}...\n${JSON.stringify(data, null, 2)}`);
            localStorage.setItem('token', data.token);
        } else {
            showResult('userResult', `登录失败: ${data.message || data.error}`, true);
        }
    } catch (error) {
        showResult('userResult', `请求失败: ${error.message}`, true);
    }
}

// 获取商品详情
async function getProduct() {
    const productId = document.getElementById('productId').value;

    if (!productId) {
        showResult('productResult', '请输入商品ID', true);
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/products/${productId}`);
        const data = await response.json();
        if (response.ok) {
            showResult('productResult', JSON.stringify(data, null, 2));
        } else {
            showResult('productResult', `获取失败: ${data.message || data.error}`, true);
        }
    } catch (error) {
        showResult('productResult', `请求失败: ${error.message}`, true);
    }
}

// 负载均衡测试
async function testLoadBalance() {
    const results = [];
    const count = 10;

    for (let i = 0; i < count; i++) {
        try {
            const response = await fetch(`${API_BASE}/users/healthz`);
            const data = await response.json();
            results.push({
                index: i + 1,
                status: data.status,
                instance: data.instance || 'unknown'
            });
        } catch (error) {
            results.push({
                index: i + 1,
                error: error.message
            });
        }
    }

    const output = results.map(r =>
        `请求 ${r.index}: ${r.instance || r.error || r.status}`
    ).join('\n');

    // 统计各实例处理请求数
    const instanceCount = {};
    results.forEach(r => {
        const key = r.instance || 'unknown';
        instanceCount[key] = (instanceCount[key] || 0) + 1;
    });

    const summary = '\n\n=== 请求分布统计 ===\n' +
        Object.entries(instanceCount).map(([k, v]) => `${k}: ${v}次`).join('\n');

    showResult('lbResult', output + summary);
}

// 页面加载时获取系统信息
document.addEventListener('DOMContentLoaded', () => {
    loadSystemInfo();

    // 每5秒刷新系统信息
    setInterval(loadSystemInfo, 5000);
});
