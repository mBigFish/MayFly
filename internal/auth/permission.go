package auth

// 权限码定义（对应 PROJECT_SPEC.md 第 22 节）。
const (
	PermTargetRead    = "target:read"
	PermTargetCreate  = "target:create"
	PermTargetUpdate  = "target:update"
	PermTargetDelete  = "target:delete"
	PermTerminalExec  = "terminal:execute"
	PermFileRead      = "file:read"
	PermFileWrite     = "file:write"
	PermFileDelete    = "file:delete"
	PermPluginExecute = "plugin:execute"
	PermAuditRead     = "audit:read"
	PermUserManage    = "user:manage"
)

// 内置角色名。
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleAuditor  = "auditor"
)

// AllPermissions 返回系统支持的全部权限码。
func AllPermissions() []string {
	return []string{
		PermTargetRead,
		PermTargetCreate,
		PermTargetUpdate,
		PermTargetDelete,
		PermTerminalExec,
		PermFileRead,
		PermFileWrite,
		PermFileDelete,
		PermPluginExecute,
		PermAuditRead,
		PermUserManage,
	}
}

// RolePermissionMap 返回内置角色到权限码的映射。
func RolePermissionMap() map[string][]string {
	return rolePermissions
}

// rolePermissions 定义内置角色的权限映射。
var rolePermissions = map[string][]string{
	RoleAdmin: AllPermissions(),
	RoleOperator: {
		PermTargetRead,
		PermTargetCreate,
		PermTargetUpdate,
		PermTerminalExec,
		PermFileRead,
		PermFileWrite,
		PermFileDelete,
		PermPluginExecute,
	},
	RoleAuditor: {
		PermTargetRead,
		PermAuditRead,
	},
}
