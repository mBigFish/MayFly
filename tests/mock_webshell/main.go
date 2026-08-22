// Command mock_webshell 是一个本地授权的模拟 PHP WebShell 服务，
// 用于 Phase 3/4 的传输、协议、文件管理与终端验证。仅用于本地测试。
//
// 模拟行为（基于 cmd 参数）：
//   - echo WSM_OK              → 返回 "WSM_OK"（探活）
//   - ls -la <path>            → 返回模拟目录列表
//   - cat <path>               → 返回模拟文件内容
//   - printf ... > <path>      → 返回 "written: <path>"
//   - mv / mkdir / rm          → 返回对应操作结果
//   - uname -a && id && pwd    → 返回系统信息
//   - 其他命令                 → 回显 "executed: <cmd>"
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9999", "监听地址")
	flag.Parse()

	http.HandleFunc("/shell.php", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "parse error", http.StatusBadRequest)
			return
		}
		cmd := r.FormValue("cmd")
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, simulate(cmd))
	})

	log.Printf("mock webshell listening on http://%s/shell.php", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// simulate 根据命令返回模拟输出。
func simulate(cmd string) string {
	switch {
	case cmd == "echo WSM_OK":
		return "WSM_OK"

	case strings.HasPrefix(cmd, "ls -la"):
		path := strings.TrimPrefix(cmd, "ls -la ")
		return fmt.Sprintf("total 4\ndrwxr-xr-x 2 root root 4096 Jan  1 00:00 .\ndrwxr-xr-x 3 root root 4096 Jan  1 00:00 ..\n-rw-r--r-- 1 root root  123 Jan  1 00:00 index.php\n(path=%s)", path)

	case strings.HasPrefix(cmd, "cat "):
		path := strings.TrimPrefix(cmd, "cat ")
		path = strings.Trim(path, "'")
		return fmt.Sprintf("<?php echo 'hello from %s'; ?>", path)

	case strings.HasPrefix(cmd, "printf "):
		// printf '%s' <content> > <path>
		return fmt.Sprintf("written: %s", strings.TrimSpace(cmd))

	case strings.HasPrefix(cmd, "mv "):
		return "renamed"

	case strings.HasPrefix(cmd, "mkdir "):
		return "created"

	case strings.HasPrefix(cmd, "rm -rf "):
		return "deleted"

	case cmd == "uname -a && id && pwd":
		return "Linux mock 5.10.0 x86_64\nuid=33(www-data) gid=33(www-data)\n/var/www/html"

	default:
		return "executed: " + cmd
	}
}
