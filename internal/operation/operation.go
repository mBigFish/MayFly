// Package operation 定义统一的操作抽象。
// Terminal、FileManager、SystemInfo 等能力都通过统一的 Operation 表达。
package operation

// OperationType 操作类型。
type OperationType string

// 支持的操作类型（对应 PROJECT_SPEC.md 第 11、12 节）。
const (
	OperationCommand    OperationType = "command"
	OperationReadFile   OperationType = "read_file"
	OperationListDir    OperationType = "list_dir"
	OperationWriteFile  OperationType = "write_file"
	OperationSystemInfo OperationType = "system_info"

	// 文件管理扩展操作（第 12 节）。
	OperationRename OperationType = "rename"
	OperationMkdir  OperationType = "mkdir"
	OperationDelete OperationType = "delete"
)

// Operation 统一操作请求。
type Operation struct {
	Type   OperationType  `json:"type"`
	Params map[string]any `json:"params"`
}

// Result 操作执行结果。
type Result struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}
