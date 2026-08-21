package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"mayfly/internal/model"
)

// Store 节点持久化存储（JSON 文件）
type Store struct {
	mu    sync.RWMutex
	path  string
	nodes map[string]*model.Node
}

// New 创建存储实例并加载已有数据
func New(path string) (*Store, error) {
	s := &Store{path: path, nodes: make(map[string]*model.Node)}
	if err := s.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var list []*model.Node
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	for _, n := range list {
		if n.ID != "" {
			s.nodes[n.ID] = n
		}
	}
	return nil
}

func (s *Store) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	list := s.List()
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// List 返回所有节点，按创建时间排序
func (s *Store) List() []*model.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		list = append(list, n)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})
	return list
}

// Get 获取单个节点
func (s *Store) Get(id string) (*model.Node, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.nodes[id]
	return n, ok
}

// SetTestResult 记录节点的最近一次连接测试结果并持久化
func (s *Store) SetTestResult(id string, status, message string) {
	s.mu.Lock()
	n, ok := s.nodes[id]
	if ok {
		now := time.Now()
		n.LastTestStatus = status
		n.LastTestTime = &now
		n.LastTestMessage = message
	}
	s.mu.Unlock()
	if ok {
		_ = s.save()
	}
}

// Add 新增节点
func (s *Store) Add(n *model.Node) error {
	s.mu.Lock()
	s.nodes[n.ID] = n
	s.mu.Unlock()
	return s.save()
}

// Update 更新节点
func (s *Store) Update(n *model.Node) error {
	s.mu.Lock()
	old, ok := s.nodes[n.ID]
	if !ok {
		s.mu.Unlock()
		return os.ErrNotExist
	}
	n.CreatedAt = old.CreatedAt
	s.nodes[n.ID] = n
	s.mu.Unlock()
	return s.save()
}

// Delete 删除节点
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	delete(s.nodes, id)
	s.mu.Unlock()
	return s.save()
}