package dto

// FileListRequest 列出目录请求。
type FileListRequest struct {
	Path string `json:"path"`
}

// FileReadRequest 读取文件请求。
type FileReadRequest struct {
	Path string `json:"path" binding:"required"`
}

// FileWriteRequest 写入文件请求。
type FileWriteRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content"`
}

// FileRenameRequest 重命名请求。
type FileRenameRequest struct {
	From string `json:"from" binding:"required"`
	To   string `json:"to" binding:"required"`
}

// FileMkdirRequest 创建目录请求。
type FileMkdirRequest struct {
	Path string `json:"path" binding:"required"`
}

// FileDeleteRequest 删除请求。
type FileDeleteRequest struct {
	Path string `json:"path" binding:"required"`
}

// FileResult 文件操作结果（统一封装）。
type FileResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}
