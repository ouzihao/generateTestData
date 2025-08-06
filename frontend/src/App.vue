<template>
  <div id="app">
    <el-container class="layout-container">
      <!-- 侧边栏 -->
      <el-aside width="250px" class="sidebar">
        <div class="logo">
          <h2>测试数据生成平台</h2>
        </div>
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <el-menu-item index="/">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/datasource">
            <el-icon><Platform /></el-icon>
            <span>数据源管理</span>
          </el-menu-item>
          <el-menu-item index="/task">
            <el-icon><List /></el-icon>
            <span>任务管理</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <!-- 头部 -->
        <el-header class="header">
          <div class="header-content">
            <h3>{{ getPageTitle() }}</h3>
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
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const getPageTitle = () => {
  const titleMap = {
    '/': '首页',
    '/datasource': '数据源管理',
    '/task': '任务管理'
  }
  return titleMap[route.path] || '数据生成平台'
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  color: #bfcbd9;
}

.logo {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #434a50;
}

.logo h2 {
  color: #fff;
  margin: 0;
  font-size: 18px;
}

.sidebar-menu {
  border: none;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  padding: 0 20px;
}

.header-content h3 {
  margin: 0;
  color: #303133;
}

.main-content {
  background-color: #f5f5f5;
  padding: 20px;
}
</style>