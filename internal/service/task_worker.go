package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mayfly/internal/database"
	"mayfly/internal/model"
	"mayfly/internal/protocol"
)

// TaskPayload 任务参数
type TaskPayload struct {
	TargetIDs []uint `json:"target_ids"`
	Command   string `json:"command"`
}

// TaskResultEntry 单个目标执行结果
type TaskResultEntry struct {
	TargetID   uint   `json:"target_id"`
	TargetName string `json:"target_name"`
	Status     string `json:"status"`
	Output     string `json:"output"`
	Error      string `json:"error"`
	Duration   string `json:"duration"`
}

// TaskResult 任务结果
type TaskResult struct {
	Results []TaskResultEntry `json:"results"`
}

// taskWorker 任务工作池
type taskWorker struct {
	taskChan chan uint // task ID channel
	workers int
	wg      sync.WaitGroup
	stopCh  chan struct{}
}

var (
	workerOnce sync.Once
	worker     *taskWorker
)

// GetTaskWorker 获取全局任务工作池
func GetTaskWorker() *taskWorker {
	workerOnce.Do(func() {
		worker = &taskWorker{
			taskChan: make(chan uint, 100),
			workers: 5,
			stopCh:  make(chan struct{}),
		}
		go worker.start()
	})
	return worker
}

// start 启动工作池
func (w *taskWorker) start() {
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.work(i)
	}
}

// work 单个 worker 循环
func (w *taskWorker) work(id int) {
	defer w.wg.Done()
	for {
		select {
		case taskID := <-w.taskChan:
			w.executeTask(taskID)
		case <-w.stopCh:
			return
		}
	}
}

// Submit 提交任务到工作池
func (w *taskWorker) Submit(taskID uint) {
	go func() {
		w.taskChan <- taskID
	}()
}

// executeTask 执行单个任务
func (w *taskWorker) executeTask(taskID uint) {
	db := database.Get()
	task, err := GetTaskByID(db, taskID)
	if err != nil {
		return
	}

	// 检查是否已取消
	if task.Status == "cancelled" {
		return
	}

	// 更新状态为 running
	now := time.Now()
	_ = db.Model(task).Updates(map[string]interface{}{
		"status":     "running",
		"started_at": now,
	})

	// 解析 payload
	var payload TaskPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		_ = db.Model(task).Updates(map[string]interface{}{
			"status": "failed",
			"result": fmt.Sprintf(`{"error":"%s"}`, err.Error()),
			"done_at": time.Now(),
		})
		return
	}

	task.Total = len(payload.TargetIDs)
	_ = db.Model(task).Update("total", task.Total)

	// 并发执行
	var results []TaskResultEntry
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, tid := range payload.TargetIDs {
		// 检查是否已取消
		var t model.Task
		db.First(&t, taskID)
		if t.Status == "cancelled" {
			break
		}

		wg.Add(1)
		go func(targetID uint) {
			defer wg.Done()
			entry := w.executeOnTarget(targetID, task.Type, payload.Command)
			mu.Lock()
			results = append(results, entry)
			task.Done++
			_ = db.Model(task).Update("done", task.Done)
			mu.Unlock()
		}(tid)
	}
	wg.Wait()

	// 保存结果
	resultJSON, _ := json.Marshal(TaskResult{Results: results})
	finalStatus := "completed"
	if task.Status == "cancelled" {
		finalStatus = "cancelled"
	}

	_ = db.Model(task).Updates(map[string]interface{}{
		"status":  finalStatus,
		"result":  string(resultJSON),
		"done_at": time.Now(),
	})
}

// executeOnTarget 在单个目标上执行任务
func (w *taskWorker) executeOnTarget(targetID uint, taskType, command string) TaskResultEntry {
	db := database.Get()
	target, err := GetTargetByID(db, targetID)
	if err != nil {
		return TaskResultEntry{
			TargetID: targetID,
			Status:   "error",
			Error:    "目标不存在",
		}
	}

	entry := TaskResultEntry{
		TargetID:   targetID,
		TargetName: target.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	switch taskType {
	case "batch_check":
		if err := protocol.CheckTarget(ctx, target); err != nil {
			entry.Status = "error"
			entry.Error = err.Error()
			_ = UpdateTargetStatus(db, targetID, "offline")
		} else {
			entry.Status = "ok"
			entry.Output = "连接成功"
			_ = UpdateTargetStatus(db, targetID, "online")
		}

	case "batch_command":
		result, err := protocol.ExecuteForTarget(ctx, target, &protocol.Operation{
			Type:   protocol.OpCommand,
			Params: map[string]string{"cmd": command},
		})
		if err != nil {
			entry.Status = "error"
			entry.Error = err.Error()
		} else {
			entry.Status = result.Status
			entry.Output = string(result.Data)
			entry.Duration = result.Message
		}

	default:
		entry.Status = "error"
		entry.Error = "不支持的任务类型: " + taskType
	}

	return entry
}
