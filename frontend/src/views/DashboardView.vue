<script setup lang="ts">
// 概览页。完整 Dashboard 统计将在后续阶段实现。
</script>

<template>
  <div class="dashboard">
    <h1>WebShell Manager</h1>
    <p>面向授权安全测试、CTF/靶场及内部安全验证环境的 WebShell 管理与分析平台。</p>
    <el-alert
      title="提示"
      description="请前往「目标」页面管理授权测试目标，并进行探活检测。"
      type="info"
      show-icon
      :closable="false"
    />
  </div>
</template>

<style scoped>
.dashboard {
  padding: 8px;
}
</style>
