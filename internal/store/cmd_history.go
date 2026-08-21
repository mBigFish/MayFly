package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CmdRecord 一条命令执行历史
type CmdRecord struct {
	Cmd    string    `json:"cmd"`
	Output string    `json:"output"`
	Error  string    `json:"error"`
	Time   time.Time `json:"time"`
}

// CmdHistoryStore 命令执行历史存储（按节点缓存，JSON 文件持久化）
type CmdHistoryStore struct {
	mu          sync.RWMutex
	path        string
	history     map[string][]CmdRecord // nodeID -> records
	maxPerNode  int
	maxOutputLen int
}

// NewCmdHistory 创建命令历史存储并加载已有数据
func NewCmdHistory(path string) (*CmdHistoryStore, error) {
	s := &CmdHistoryStore{
		path:         path,
		history:      make(map[string][]CmdRecord),
		maxPerNode:   300,
		maxOutputLen: 256 * 1024,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *CmdHistoryStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.history)
}

func (s *CmdHistoryStore) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// Append 追加一条命令历史（超出长度截断输出，超出条数裁剪最旧的）
func (s *CmdHistoryStore) Append(nodeID, cmd, output, errMsg string) {
	if len(output) > s.maxOutputLen {
		output = output[:s.maxOutputLen] + "\n...[输出过长已截断]"
	}
	rec := CmdRecord{Cmd: cmd, Output: output, Error: errMsg, Time: time.Now()}

	s.mu.Lock()
	recs := s.history[nodeID]
	recs = append(recs, rec)
	if len(recs) > s.maxPerNode {
		recs = recs[len(recs)-s.maxPerNode:]
	}
	s.history[nodeID] = recs
	s.mu.Unlock()
	_ = s.save()
}

// History 返回某节点的历史（按时间升序）
func (s *CmdHistoryStore) History(nodeID string) []CmdRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	recs := s.history[nodeID]
	out := make([]CmdRecord, len(recs))
	copy(out, recs)
	return out
}

// Clear 清空某节点的历史
func (s *CmdHistoryStore) Clear(nodeID string) {
	s.mu.Lock()
	delete(s.history, nodeID)
	s.mu.Unlock()
	_ = s.save()
}