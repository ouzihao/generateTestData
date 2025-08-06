import request from './request'

// 数据源API
export const datasourceApi = {
  // 获取数据源列表
  getList() {
    return request.get('/datasource')
  },

  // 获取单个数据源
  getById(id) {
    return request.get(`/datasource/${id}`)
  },

  // 创建数据源
  create(data) {
    return request.post('/datasource', data)
  },

  // 更新数据源
  update(id, data) {
    return request.put(`/datasource/${id}`, data)
  },

  // 删除数据源
  delete(id) {
    return request.delete(`/datasource/${id}`)
  },

  // 测试连接
  testConnection(data) {
    return request.post('/datasource/test', data)
  },

  // 获取表列表
  getTables(id) {
    return request.get(`/datasource/tables/${id}`)
  },

  // 获取表结构
  getTableStructure(id, tableName) {
    return request.get(`/datasource/table/${id}/${tableName}`)
  }
}