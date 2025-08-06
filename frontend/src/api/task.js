import request from './request'

// 任务API
export const taskApi = {
  // 获取任务列表
  getList(params = {}) {
    return request.get('/tasks', { params })
  },

  // 获取单个任务
  getById(id) {
    return request.get(`/tasks/${id}`)
  },

  // 创建任务
  create(data) {
    return request.post('/tasks', data)
  },

  // 更新任务
  update(id, data) {
    return request.put(`/tasks/${id}`, data)
  },

  // 删除任务
  delete(id) {
    return request.delete(`/tasks/${id}`)
  },

  // 执行任务
  execute(id) {
    return request.post(`/tasks/${id}/execute`)
  },

  // 获取任务状态
  getStatus(id) {
    return request.get(`/tasks/${id}/status`)
  },

  // 生成预览数据
  preview(data) {
    return request.post('/tasks/preview', data)
  },

  // 导出任务规则模板
  exportTemplate(id, data) {
    return request.post(`/tasks/${id}/export-template`, data)
  }
}

// 规则模板API
export const templateApi = {
  // 获取模板列表
  getList() {
    return request.get('/templates')
  },

  // 导入模板
  import(data) {
    return request.post('/templates/import', data)
  },

  // 删除模板
  delete(id) {
    return request.delete(`/templates/${id}`)
  }
}