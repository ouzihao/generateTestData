<template>
  <div class="home">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stats-card">
        <div class="stats-number">{{ stats.totalDataSources }}</div>
        <div class="stats-label">数据源总数</div>
      </div>
      <div class="stats-card">
        <div class="stats-number">{{ stats.totalTasks }}</div>
        <div class="stats-label">任务总数</div>
      </div>
      <div class="stats-card">
        <div class="stats-number">{{ stats.runningTasks }}</div>
        <div class="stats-label">运行中任务</div>
      </div>
      <div class="stats-card">
        <div class="stats-number">{{ stats.completedTasks }}</div>
        <div class="stats-label">已完成任务</div>
      </div>
    </div>

    <!-- 快速操作 -->
    <div class="info-card">
      <div class="card-title">快速操作</div>
      <div class="quick-actions">
        <el-button type="primary" @click="$router.push('/datasource')">
          <el-icon><Plus /></el-icon>
          添加数据源
        </el-button>
        <el-button type="success" @click="$router.push('/task')">
          <el-icon><DocumentAdd /></el-icon>
          创建任务
        </el-button>
      </div>
    </div>

    <!-- 最近任务 -->
    <div class="info-card">
      <div class="card-title">最近任务</div>
      <el-table :data="recentTasks" v-loading="loading">
        <el-table-column prop="name" label="任务名称" />
        <el-table-column prop="type" label="类型">
          <template #default="{ row }">
            <el-tag :type="row.type === 'database' ? 'primary' : 'success'">
              {{ row.type === 'database' ? '数据库' : 'JSON' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag
              :type="getStatusType(row.status)"
              class="status-tag"
            >
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress" label="进度">
          <template #default="{ row }">
            <div class="progress-container">
              <el-progress
                :percentage="row.progress"
                :status="row.status === 'failed' ? 'exception' : undefined"
                :stroke-width="6"
                style="flex: 1"
              />
              <span class="progress-text">{{ row.progress }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 系统信息 -->
    <div class="info-card">
      <div class="card-title">系统信息</div>
      <div class="system-info">
        <div class="info-item">
          <span class="info-label">版本:</span>
          <span class="info-value">v1.0.0</span>
        </div>
        <div class="info-item">
          <span class="info-label">当前时间:</span>
          <span class="info-value">{{ currentTime }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">支持数据库:</span>
          <span class="info-value">MySQL, PostgreSQL, SQLite</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { datasourceApi } from '@/api/datasource'
import { taskApi } from '@/api/task'

const loading = ref(false)
const stats = ref({
  totalDataSources: 0,
  totalTasks: 0,
  runningTasks: 0,
  completedTasks: 0
})
const recentTasks = ref([])
const currentTime = ref('')

let timeTimer = null

// 获取统计数据
const loadStats = async () => {
  try {
    // 获取数据源统计
    const dataSourceRes = await datasourceApi.getList()
    stats.value.totalDataSources = dataSourceRes.data?.length || 0

    // 获取任务统计
    const taskRes = await taskApi.getList({ page: 1, pageSize: 100 })
    const tasks = taskRes.data?.list || []
    stats.value.totalTasks = taskRes.data?.total || 0
    stats.value.runningTasks = tasks.filter(t => t.status === 'running').length
    stats.value.completedTasks = tasks.filter(t => t.status === 'completed').length

    // 获取最近任务（前5个）
    recentTasks.value = tasks.slice(0, 5)
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 更新当前时间
const updateCurrentTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 获取状态类型
const getStatusType = (status) => {
  const typeMap = {
    pending: 'info',
    running: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return typeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const textMap = {
    pending: '待执行',
    running: '运行中',
    completed: '已完成',
    failed: '失败'
  }
  return textMap[status] || '未知'
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  loading.value = true
  loadStats().finally(() => {
    loading.value = false
  })
  
  // 启动时间计时器
  updateCurrentTime()
  timeTimer = setInterval(updateCurrentTime, 1000) // 每秒更新一次
})

onUnmounted(() => {
  if (timeTimer) {
    clearInterval(timeTimer)
  }
})
</script>

<style scoped>
.home {
  max-width: 1200px;
  margin: 0 auto;
}

.quick-actions {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
}

.system-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 15px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-weight: 500;
  color: #606266;
}

.info-value {
  color: #303133;
}

@media (max-width: 768px) {
  .quick-actions {
    flex-direction: column;
  }
  
  .system-info {
    grid-template-columns: 1fr;
  }
}
</style>