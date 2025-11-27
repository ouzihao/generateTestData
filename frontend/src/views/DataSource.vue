<template>
  <div class="datasource">
    <div class="page-container">
      <!-- 页面头部 -->
      <div class="page-header">
        <h2 class="page-title">数据源管理</h2>
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          添加数据源
        </el-button>
      </div>

      <!-- 数据源列表 -->
      <div class="table-container">
        <el-table :data="dataSourceList" v-loading="loading">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="type" label="类型">
            <template #default="{ row }">
              <el-tag :type="getDbTypeColor(row.type)">
                {{ getDbTypeName(row.type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="host" label="主机" />
          <el-table-column prop="port" label="端口" />
          <el-table-column prop="database" label="数据库" />
          <el-table-column prop="username" label="用户名" />
          <el-table-column label="操作" width="250">
            <template #default="{ row }">
              <div class="button-group">
                <el-button size="small" @click="testConnection(row)">
                  <el-icon><Connection /></el-icon>
                  测试连接
                </el-button>
                <el-button size="small" type="primary" @click="editDataSource(row)">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
                <el-button size="small" type="danger" @click="deleteDataSource(row)">
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 创建/编辑数据源对话框 -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
        class="form-container"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入数据源名称" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择数据库类型" class="form-item-full">
            <el-option label="MySQL" value="mysql" />
            <el-option label="PostgreSQL" value="postgresql" />
            <el-option label="SQLite" value="sqlite" />
          </el-select>
        </el-form-item>
        <el-form-item label="主机" prop="host" v-if="formData.type !== 'sqlite'">
          <el-input v-model="formData.host" placeholder="请输入主机地址" />
        </el-form-item>
        <el-form-item label="端口" prop="port" v-if="formData.type !== 'sqlite'">
          <el-input-number v-model="formData.port" :min="1" :max="65535" class="form-item-full" />
        </el-form-item>
        <el-form-item label="数据库" prop="database">
          <el-input 
            v-model="formData.database" 
            :placeholder="formData.type === 'sqlite' ? '请输入SQLite文件路径' : '请输入数据库名称'" 
          />
        </el-form-item>
        <el-form-item label="用户名" prop="username" v-if="formData.type !== 'sqlite'">
          <el-input v-model="formData.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="formData.type !== 'sqlite'">
          <el-input v-model="formData.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="button-group">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="testConnectionInDialog" :loading="testing">
            测试连接
          </el-button>
          <el-button type="success" @click="saveDataSource" :loading="saving">
            保存
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Connection, Edit, Delete } from '@element-plus/icons-vue'
import { datasourceApi } from '@/api/datasource'

const loading = ref(false)
const testing = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const dataSourceList = ref([])
const formRef = ref()

const formData = reactive({
  id: null,
  name: '',
  type: 'mysql',
  host: 'localhost',
  port: 3306,
  database: '',
  username: '',
  password: ''
})

const formRules = {
  name: [{ required: true, message: '请输入数据源名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择数据库类型', trigger: 'change' }],
  host: [{ required: true, message: '请输入主机地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口号', trigger: 'blur' }],
  database: [{ required: true, message: '请输入数据库名称', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
}

const dialogTitle = computed(() => isEdit.value ? '编辑数据源' : '添加数据源')

// 加载数据源列表
const loadDataSources = async () => {
  loading.value = true
  try {
    const res = await datasourceApi.getList()
    dataSourceList.value = res.data || []
  } catch (error) {
    console.error('加载数据源列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  isEdit.value = false
  dialogVisible.value = true
}

// 编辑数据源
const editDataSource = (row) => {
  isEdit.value = true
  Object.assign(formData, row)
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(formData, {
    id: null,
    name: '',
    type: 'mysql',
    host: 'localhost',
    port: 3306,
    database: '',
    username: '',
    password: ''
  })
}

// 测试连接
const testConnection = async (dataSource) => {
  testing.value = true
  try {
    await datasourceApi.testConnection(dataSource)
    ElMessage.success('连接成功')
  } catch (error) {
    console.error('测试连接失败:', error)
  } finally {
    testing.value = false
  }
}

// 在对话框中测试连接
const testConnectionInDialog = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    testing.value = true
    await datasourceApi.testConnection(formData)
    ElMessage.success('连接成功')
  } catch (error) {
    if (error.errors) {
      ElMessage.error('请先完善表单信息')
    } else {
      console.error('测试连接失败:', error)
    }
  } finally {
    testing.value = false
  }
}

// 保存数据源
const saveDataSource = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    if (isEdit.value) {
      await datasourceApi.update(formData.id, formData)
      ElMessage.success('更新成功')
    } else {
      await datasourceApi.create(formData)
      ElMessage.success('创建成功')
    }
    
    dialogVisible.value = false
    loadDataSources()
  } catch (error) {
    if (error.errors) {
      ElMessage.error('请完善表单信息')
    } else {
      console.error('保存数据源失败:', error)
    }
  } finally {
    saving.value = false
  }
}

// 删除数据源
const deleteDataSource = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除数据源 "${row.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await datasourceApi.delete(row.id)
    ElMessage.success('删除成功')
    loadDataSources()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除数据源失败:', error)
    }
  }
}

// 获取数据库类型颜色
const getDbTypeColor = (type) => {
  const colorMap = {
    mysql: 'primary',
    postgresql: 'success',
    sqlite: 'warning'
  }
  return colorMap[type] || 'info'
}

// 获取数据库类型名称
const getDbTypeName = (type) => {
  const nameMap = {
    mysql: 'MySQL',
    postgresql: 'PostgreSQL',
    sqlite: 'SQLite'
  }
  return nameMap[type] || type
}

onMounted(() => {
  loadDataSources()
})
</script>

<style scoped>
.datasource {
  max-width: 100%;
}

.button-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.button-group .el-button {
  transition: all 0.3s ease;
}

.button-group .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 表格样式优化 */
.table-container :deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}

.table-container :deep(.el-table th) {
  background: linear-gradient(135deg, #f5f7fa 0%, #ffffff 100%);
  font-weight: 600;
}

.table-container :deep(.el-table tr:hover) {
  background: #f0f9ff;
}

/* 对话框样式优化 */
:deep(.el-dialog) {
  border-radius: 12px;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.15);
}

:deep(.el-dialog__header) {
  padding: 20px 24px;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-bottom: 1px solid #e4e7ed;
}

:deep(.el-dialog__title) {
  font-size: 18px;
  font-weight: 600;
  background: linear-gradient(135deg, #409EFF, #66b1ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

:deep(.el-dialog__body) {
  padding: 24px;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px;
  background: #f8f9fa;
  border-top: 1px solid #e4e7ed;
}

/* 表单样式优化 */
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input__wrapper) {
  border-radius: 6px;
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #409EFF inset;
}

:deep(.el-select .el-input__wrapper) {
  border-radius: 6px;
}

@media (max-width: 768px) {
  .button-group {
    width: 100%;
  }
  
  .button-group .el-button {
    flex: 1;
    min-width: 100px;
  }
}
</style>