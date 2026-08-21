// Command mock_webshell 是一个本地授权的模拟 PHP WebShell 服务，
// 用于 Phase 3 的传输与协议层验证。仅用于本地测试，请勿用于生产。
//
// 模拟行为：
//   - 收到 cmd=echo WSM_OK 时返回 "WSM_OK"，用于探活。
//   - 收到其他 cmd 时原样回显命令输出。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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
		// 模拟 PHP webshell 执行命令：这里直接回显命令名作为"输出"。
		// 探活命令 echo WSM_OK 直接返回标记。
		if cmd == "echo WSM_OK" {
			fmt.Fprint(w, "WSM_OK")
			return
		}
		fmt.Fprintf(w, "executed: %s", cmd)
	})

	log.Printf("mock webshell listening on http://%s/shell.php", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
