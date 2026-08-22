package handler

import (
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"mayfly/internal/store"
)

// DashboardHandler 仪表盘聚合统计
type DashboardHandler struct {
	nodeStore  *store.Store
	cmdHistory *store.CmdHistoryStore
}

// NewDashboardHandler 创建 DashboardHandler
func NewDashboardHandler(ns *store.Store, ch *store.CmdHistoryStore) *DashboardHandler {
	return &DashboardHandler{nodeStore: ns, cmdHistory: ch}
}

// Dashboard 返回仪表盘聚合统计数据
// GET /api/dashboard
func (h *DashboardHandler) Dashboard(c *gin.Context) {
	// 节点统计
	nodes := h.nodeStore.List()
	nodeTotal := len(nodes)
	nodeOnline := 0
	typeCount := map[string]int{}
	groupCount := map[string]int{}
	for _, n := range nodes {
		typeCount[n.Type]++
		g := n.Group
		if g == "" {
			g = "默认"
		}
		groupCount[g]++
		if n.LastTestStatus == "ok" {
			nodeOnline++
		}
	}

	// 服务器统计
	ensureServerStore()
	serverStoreMu.Lock()
	servers := make([]serverInfo, 0, len(serverStore.Servers))
	for _, s := range serverStore.Servers {
		servers = append(servers, serverInfo{
			ID:             s.ID,
			Name:           s.Name,
			Group:          s.Group,
			Host:           s.Host,
			LastTestStatus: s.LastTestStatus,
			LastTestTime:   s.LastTestTime,
			LastTestMessage: s.LastTestMessage,
		})
	}
	serverStoreMu.Unlock()

	serverTotal := len(servers)
	serverOnline := 0
	serverGroupCount := map[string]int{}
	for _, s := range servers {
		g := s.Group
		if g == "" {
			g = "默认"
		}
		serverGroupCount[g]++
		if s.LastTestStatus == "ok" {
			serverOnline++
		}
	}

	// 监听器统计
	listeners := getListenerManager().ListListeners()
	listenerTotal := len(listeners)
	listenerActive := 0
	for _, l := range listeners {
		if l.Status == "listening" {
			listenerActive++
		}
	}

	// 最近活动：取最近 20 条命令历史（跨所有节点）
	type cmdEntry struct {
		Cmd    string
		Output string
		Node   string
		NodeID string
		Time   time.Time
	}
	var allCmds []cmdEntry
	for _, n := range nodes {
		recs := h.cmdHistory.History(n.ID)
		for _, r := range recs {
			allCmds = append(allCmds, cmdEntry{Cmd: r.Cmd, Output: r.Output, Node: n.Name, NodeID: n.ID, Time: r.Time})
		}
	}
	sort.Slice(allCmds, func(i, j int) bool {
		return allCmds[i].Time.After(allCmds[j].Time)
	})
	if len(allCmds) > 20 {
		allCmds = allCmds[:20]
	}
	recentCmds := make([]gin.H, 0, len(allCmds))
	for _, e := range allCmds {
		output := e.Output
		if len(output) > 80 {
			output = output[:80] + "..."
		}
		recentCmds = append(recentCmds, gin.H{
			"cmd":     e.Cmd,
			"output":  output,
			"node":    e.Node,
			"node_id": e.NodeID,
			"time":    e.Time.Unix(),
		})
	}

	// 最近连接失败的节点
	type failInfo struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		URL     string `json:"url"`
		Message string `json:"message"`
		Time    int64  `json:"time"`
	}
	var recentFails []failInfo
	for _, n := range nodes {
		if n.LastTestStatus == "fail" && n.LastTestTime != nil {
			recentFails = append(recentFails, failInfo{
				ID:      n.ID,
				Name:    n.Name,
				URL:     n.URL,
				Message: n.LastTestMessage,
				Time:    n.LastTestTime.Unix(),
			})
		}
	}
	sort.Slice(recentFails, func(i, j int) bool {
		return recentFails[i].Time > recentFails[j].Time
	})
	if len(recentFails) > 10 {
		recentFails = recentFails[:10]
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": gin.H{
			"total":   nodeTotal,
			"online":  nodeOnline,
			"offline": nodeTotal - nodeOnline,
			"types":   typeCount,
			"groups":  groupCount,
		},
		"servers": gin.H{
			"total":   serverTotal,
			"online":  serverOnline,
			"offline": serverTotal - serverOnline,
			"groups":  serverGroupCount,
		},
		"listeners": gin.H{
			"total":  listenerTotal,
			"active": listenerActive,
		},
		"recent_commands": recentCmds,
		"recent_fails":    recentFails,
	})
}

// serverInfo 仪表盘用的服务器摘要信息
type serverInfo struct {
	ID              int
	Name            string
	Group           string
	Host            string
	LastTestStatus  string
	LastTestTime    *time.Time
	LastTestMessage string
}
