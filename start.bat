@echo off
chcp 65001 >nul
cd /d "%~dp0"

echo.
echo  ============================================
echo    Mayfly WebShell 启动中...
echo  ============================================
echo.
echo    访问地址 : http://localhost:8080
echo    默认账号 : admin
echo    默认密码 : mayfly123
echo.
echo    修改配置（可选，在启动前设置）:
echo    set MAYFLY_PORT=9090
echo    set MAYFLY_USER=myuser
echo    set MAYFLY_PASS=mypassword
echo.
echo    按 Ctrl+C 或关闭本窗口即可停止服务
echo  ============================================
echo.

mayfly.exe

pause
