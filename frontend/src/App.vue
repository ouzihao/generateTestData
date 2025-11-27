<template>
  <div id="app">
    <el-container class="layout-container">
      <!-- 侧边栏 -->
      <el-aside width="260px" class="sidebar">
        <div class="logo">
          <div class="logo-icon">
            <el-icon :size="28"><Box /></el-icon>
          </div>
          <h2>数据生成平台</h2>
          <p class="logo-subtitle">Test Data Generator</p>
        </div>
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          background-color="transparent"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <el-menu-item index="/" class="menu-item">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/datasource" class="menu-item">
            <el-icon><Platform /></el-icon>
            <span>数据源管理</span>
          </el-menu-item>
          <el-menu-item index="/task" class="menu-item">
            <el-icon><List /></el-icon>
            <span>任务管理</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-container class="main-wrapper">
        <!-- 头部 -->
        <el-header class="header">
          <div class="header-content">
            <div class="header-left">
              <h3 class="header-title">{{ getPageTitle() }}</h3>
              <p class="header-subtitle">{{ getPageSubtitle() }}</p>
            </div>
            <div class="header-right">
              <div class="header-time">
                <el-icon><Clock /></el-icon>
                <span>{{ currentTime }}</span>
              </div>
            </div>
          </div>
        </el-header>

        <!-- 内容区 -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { House, Platform, List, Clock, Box } from '@element-plus/icons-vue'

const route = useRoute()
const currentTime = ref('')

const getPageTitle = () => {
  const titleMap = {
    '/': '首页',
    '/datasource': '数据源管理',
    '/task': '任务管理'
  }
  return titleMap[route.path] || '数据生成平台'
}

const getPageSubtitle = () => {
  const subtitleMap = {
    '/': '欢迎使用测试数据生成平台',
    '/datasource': '管理数据库连接配置',
    '/task': '创建和管理数据生成任务'
  }
  return subtitleMap[route.path] || ''
}

const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

let timeTimer = null

onMounted(() => {
  updateTime()
  timeTimer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (timeTimer) {
    clearInterval(timeTimer)
  }
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
  overflow: hidden;
}

/* 侧边栏样式 */
.sidebar {
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
  color: #bfcbd9;
  box-shadow: 4px 0 12px rgba(0, 0, 0, 0.1);
  position: relative;
  overflow: hidden;
}

.sidebar::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('data:image/svg+xml,<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg"><defs><pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse"><path d="M 20 0 L 0 0 0 20" fill="none" stroke="rgba(255,255,255,0.03)" stroke-width="1"/></pattern></defs><rect width="100" height="100" fill="url(%23grid)"/></svg>');
  opacity: 0.5;
  pointer-events: none;
}

.logo {
  padding: 24px 20px;
  text-align: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: relative;
  z-index: 1;
}

.logo-icon {
  margin-bottom: 12px;
  color: #409EFF;
}

.logo h2 {
  color: #fff;
  margin: 0 0 4px 0;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.logo-subtitle {
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
  margin: 0;
  font-weight: 400;
  letter-spacing: 1px;
}

.sidebar-menu {
  border: none;
  background: transparent;
  padding: 16px 0;
  position: relative;
  z-index: 1;
}

.sidebar-menu :deep(.el-menu-item) {
  margin: 4px 12px;
  border-radius: 8px;
  height: 48px;
  line-height: 48px;
  transition: all 0.3s ease;
}

.sidebar-menu :deep(.el-menu-item:hover) {
  background: rgba(64, 158, 255, 0.1) !important;
  color: #409EFF !important;
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.2), rgba(64, 158, 255, 0.1)) !important;
  color: #409EFF !important;
  font-weight: 500;
  border-left: 3px solid #409EFF;
}

.sidebar-menu :deep(.el-menu-item i) {
  margin-right: 8px;
  font-size: 18px;
}

/* 主内容区 */
.main-wrapper {
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

/* 头部样式 */
.header {
  background: linear-gradient(135deg, #ffffff 0%, #f8f9fa 100%);
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  height: 70px !important;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.header-title {
  margin: 0;
  color: #303133;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.2;
}

.header-subtitle {
  margin: 0;
  color: #909399;
  font-size: 13px;
  font-weight: 400;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-time {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #f5f7fa;
  border-radius: 20px;
  color: #606266;
  font-size: 14px;
  font-weight: 500;
  font-family: 'Monaco', 'Menlo', monospace;
}

.header-time .el-icon {
  color: #409EFF;
}

/* 主内容区样式 */
.main-content {
  background: transparent;
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sidebar {
    width: 200px !important;
  }
  
  .logo h2 {
    font-size: 16px;
  }
  
  .header {
    padding: 0 16px;
    height: 60px !important;
  }
  
  .header-title {
    font-size: 18px;
  }
  
  .header-subtitle {
    display: none;
  }
  
  .header-time {
    font-size: 12px;
    padding: 6px 12px;
  }
  
  .main-content {
    padding: 16px;
  }
}
</style>

