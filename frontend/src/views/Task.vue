<template>
  <div class="task">
    <div class="page-container">
      <!-- 页面头部 -->
      <div class="page-header">
        <h2 class="page-title">任务管理</h2>
        <div class="header-buttons">
          <el-button type="info" @click="showTemplateDialog">
            <el-icon><Document /></el-icon>
            规则模板
          </el-button>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            创建任务
          </el-button>
        </div>
      </div>

      <!-- 任务列表 -->
      <div class="table-container">
        <el-table :data="taskList" v-loading="loading">
          <el-table-column prop="name" label="任务名称" />
          <el-table-column prop="type" label="类型">
            <template #default="{ row }">
              <el-tag :type="row.type === 'database' ? 'primary' : 'success'">
                {{ row.type === 'database' ? '数据库' : 'JSON' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="count" label="生成数量" />
          <el-table-column prop="status" label="状态">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" class="status-tag">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="progress" label="进度" width="200">
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
          <el-table-column label="操作" width="300">
            <template #default="{ row }">
              <div class="button-group">
                <el-button 
                  size="small" 
                  type="success" 
                  @click="executeTask(row)"
                  :disabled="row.status === 'running'"
                >
                  执行
                </el-button>
                <el-button size="small" @click="editTask(row)">
                  编辑
                </el-button>
                <el-button size="small" type="warning" @click="exportTemplate(row)">
                  导出模板
                </el-button>
                <el-button size="small" @click="viewTask(row)">
                  查看
                </el-button>
                <el-button size="small" type="danger" @click="deleteTask(row)">
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <div style="margin-top: 20px; text-align: center;">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="pagination.total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadTasks"
            @current-change="loadTasks"
          />
        </div>
      </div>
    </div>

    <!-- 创建任务对话框 -->
    <el-dialog
      v-model="dialogVisible"
      width="1200px"
      top="5vh"
      @close="resetForm"
      class="task-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <span class="dialog-title">创建任务</span>
          <el-button type="primary" size="small" @click="showTemplateDialog">导入模板</el-button>
        </div>
      </template>
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
        class="form-container"
      >
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入任务名称" class="form-item-full" />
        </el-form-item>
        <el-form-item label="任务类型" prop="type">
          <el-radio-group v-model="formData.type">
            <el-radio label="database">数据库任务</el-radio>
            <el-radio label="json">JSON任务</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <!-- 数据库任务配置 -->
        <template v-if="formData.type === 'database'">
          <el-form-item label="数据源" prop="dataSourceId">
            <el-select v-model="formData.dataSourceId" placeholder="请选择数据源" class="form-item-full">
              <el-option
                v-for="ds in dataSourceList"
                :key="ds.id"
                :label="ds.name + ' --> ' +  ds.database"
                :value="ds.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="表名" prop="tableName">
            <el-select 
              v-model="formData.tableName" 
              placeholder="请选择表" 
              class="form-item-full"
              @change="onTableNameChange"
            >
              <el-option
                v-for="table in tableList"
                :key="table"
                :label="table"
                :value="table"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="输出类型" prop="outputType">
            <el-radio-group v-model="formData.outputType">
              <el-radio label="database">插入数据库</el-radio>
              <el-radio label="sql">导出SQL文件</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="输出文件名" prop="outputPath" v-if="formData.outputType === 'sql'">
            <el-input v-model="formData.outputPath" placeholder="请输入SQL文件名，如：data.sql" class="form-item-full" />
          </el-form-item>
        </template>
        
        <!-- JSON任务配置 -->
        <template v-if="formData.type === 'json'">
          <el-form-item label="JSON结构" prop="jsonSchema">
            <el-input
              v-model="formData.jsonSchema"
              type="textarea"
              :rows="6"
              placeholder="请输入JSON结构，例如：{'name': 'string', 'age': 'number'}"
              class="form-item-full"
            />
            <div v-if="jsonParseError" class="json-error-tip">
              <el-alert
                :title="jsonParseError"
                type="error"
                :closable="false"
                show-icon
              />
            </div>
          </el-form-item>
          <el-form-item label="输出类型" prop="outputType">
            <el-radio-group v-model="formData.outputType">
              <el-radio label="json">JSON文件</el-radio>
              <el-radio label="txt">TXT文件（每行一个JSON）</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="输出文件名" prop="outputPath">
            <el-input v-model="formData.outputPath" :placeholder="formData.outputType === 'json' ? '请输入JSON文件名，如：data.json' : '请输入TXT文件名，如：data.txt'" class="form-item-full" />
          </el-form-item>
        </template>
        
        <el-form-item label="生成数量" prop="count">
          <el-input-number v-model="formData.count" :min="1" :max="1000000" class="form-item-full" />
        </el-form-item>
        
        <!-- 字段规则配置 -->
        <el-form-item label="字段规则" v-if="tableStructure.length > 0 || formData.type === 'json'">
          <div class="field-rules">
            <div v-for="(field, index) in getFields" :key="index" class="field-rule-item" :class="{
              'nested-field': field.name.includes('.'),
              'array-field': field.type === 'array',
              'object-field': field.type === 'object'
            }" :style="{ marginLeft: getFieldIndent(field.name) + 'px' }">
              <div class="field-info">
                <span class="field-path" v-if="field.name.includes('.') || field.name.includes('[]')">
                  {{ getFieldPath(field.name) }}
                </span>
                <span class="field-name" :title="field.name">{{ getDisplayFieldName(field.name) }}</span>
                <span class="field-type" :class="`type-${field.type}`">{{ field.type }}</span>
              </div>
              <div class="field-config">
                <div class="rule-config-row">
                  <el-select v-model="fieldRules[field.name]" placeholder="选择生成规则" class="rule-select" @change="onRuleTypeChange(field.name)">
                    <el-option label="固定值" value="fixed" />
                    <el-option label="序列" value="sequence" />
                    <el-option label="日期序列" value="date_sequence" v-if="isDateField(field)" />
                    <el-option label="随机" value="random" />
                    <el-option label="范围" value="range" />
                    <el-option label="正则" value="regex" />
                    <el-option label="枚举" value="enum" />
                    <el-option label="UUID" value="uuid" />
                    <el-option label="自定义" value="custom" />
                  </el-select>
                  <!-- 数组长度配置 -->
                  <el-input-number 
                    v-if="field.type === 'array'"
                    v-model="fieldArrayLengths[field.name]"
                    :min="1" 
                    :max="100"
                    placeholder="数组长度"
                    class="array-length-input"
                    size="small"
                  />
                </div>
                
                <!-- 规则参数配置 -->
                <div v-if="fieldRules[field.name] && fieldRules[field.name] !== 'uuid'" class="rule-params">
                  <!-- 固定值配置 -->
                  <el-input 
                    v-if="fieldRules[field.name] === 'fixed'"
                    v-model="fieldRuleParams[field.name].value"
                    placeholder="请输入固定值"
                    size="small"
                    class="param-input"
                  />
                  
                  <!-- 序列配置 -->
                  <div v-if="fieldRules[field.name] === 'sequence'" class="sequence-config">
                    <el-input 
                      v-model="fieldRuleParams[field.name].start"
                      placeholder="起始值（支持大整数）"
                      size="small"
                      class="param-input-small"
                    />
                    <el-input 
                      v-model="fieldRuleParams[field.name].step"
                      placeholder="步长"
                      size="small"
                      class="param-input-small"
                    />
                  </div>
                  
                  <!-- 随机配置 -->
                  <div v-if="fieldRules[field.name] === 'random' && isDateField(field)" class="random-date-config">
                    <el-input 
                      v-model="fieldRuleParams[field.name].start"
                      placeholder="开始日期 (YYYY-MM-DD，可选)"
                      size="small"
                      class="param-input-small"
                    />
                    <el-input 
                      v-model="fieldRuleParams[field.name].end"
                      placeholder="结束日期 (YYYY-MM-DD，可选)"
                      size="small"
                      class="param-input-small"
                    />
                    <el-input 
                      v-model="fieldRuleParams[field.name].format"
                      placeholder="日期格式 (可选)"
                      size="small"
                      class="param-input-small"
                    />
                  </div>
                  
                  <!-- 范围配置 -->
                  <div v-if="fieldRules[field.name] === 'range'" class="range-config">
                    <el-input 
                      v-model="fieldRuleParams[field.name].min"
                      placeholder="最小值"
                      size="small"
                      class="param-input-small"
                    />
                    <el-input 
                      v-model="fieldRuleParams[field.name].max"
                      placeholder="最大值"
                      size="small"
                      class="param-input-small"
                    />
                  </div>
                  
                  <!-- 正则表达式配置 -->
                  <el-autocomplete 
                    v-if="fieldRules[field.name] === 'regex'"
                    v-model="fieldRuleParams[field.name].pattern"
                    placeholder="输入关键词(如mail、phone)或正则表达式"
                    size="small"
                    class="param-input"
                    :fetch-suggestions="getRegexSuggestions"
                    @input="handleRegexInput"
                    @select="handleRegexSelect"
                  >
                    <template #default="{ item }">
                      <div class="regex-suggestion-item">
                        <div class="suggestion-keyword">{{ item.keyword }}</div>
                        <div class="suggestion-description">{{ item.description }}</div>
                        <div class="suggestion-pattern">{{ item.pattern }}</div>
                      </div>
                    </template>
                  </el-autocomplete>
                  
                  <!-- 枚举配置 -->
                  <el-input 
                    v-if="fieldRules[field.name] === 'enum'"
                    v-model="fieldRuleParams[field.name].values"
                    placeholder="请输入枚举值，用逗号分隔"
                    size="small"
                    class="param-input"
                  />
                  
                  <!-- 日期序列配置 -->
                  <div v-if="fieldRules[field.name] === 'date_sequence'" class="date-sequence-config">
                    <el-input 
                      v-model="fieldRuleParams[field.name].start"
                      placeholder="起始日期 (YYYY-MM-DD)"
                      size="small"
                      class="param-input-small"
                    />
                    <el-input 
                      v-model="fieldRuleParams[field.name].step"
                      placeholder="步长（天数）"
                      size="small"
                      class="param-input-small"
                    />
                    <el-tooltip 
                      content="日期格式说明：2006-01-02 (年-月-日), 2006-01-02 15:04:05 (年-月-日 时:分:秒), 2006/01/02 (年/月/日), 01-02-2006 (月-日-年)。留空则使用默认格式 2006-01-02"
                      placement="top"
                      effect="dark"
                    >
                      <el-input 
                        v-model="fieldRuleParams[field.name].format"
                        placeholder="日期格式 (如: 2006-01-02, 可选)"
                        size="small"
                        class="param-input-small"
                      />
                    </el-tooltip>
                  </div>
                  
                  <!-- 自定义配置 -->
                  <el-input 
                    v-if="fieldRules[field.name] === 'custom'"
                    v-model="fieldRuleParams[field.name].script"
                    placeholder="请输入自定义脚本"
                    size="small"
                    class="param-input"
                  />
                </div>
              </div>
            </div>
          </div>
        </el-form-item>
        
        <!-- 预览数据展示区域 -->
        <el-form-item label="预览数据" v-if="previewData">
          <div class="preview-data-container">
            <el-card class="preview-card">
              <template #header>
                <div class="preview-header">
                  <span>生成的预览数据</span>
                  <el-button size="small" type="text" @click="copyPreviewData">
                    <el-icon><CopyDocument /></el-icon>
                    复制
                  </el-button>
                </div>
              </template>
              <pre class="preview-content">{{ formatPreviewData(previewData) }}</pre>
            </el-card>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="button-group">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="info" @click="generatePreviewData" :loading="generatingPreview">
            生成预览数据
          </el-button>
          <el-button type="primary" @click="createTask" :loading="creating">
            创建
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑任务对话框 -->
    <el-dialog 
      v-model="editDialogVisible" 
      title="编辑任务" 
      width="80%" 
      :close-on-click-modal="false"
    >
      <el-form 
        ref="formRef" 
        :model="editingTask" 
        :rules="formRules" 
        label-width="120px"
        v-if="editingTask"
      >
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="editingTask.name" placeholder="请输入任务名称" />
        </el-form-item>
        
        <el-form-item label="任务类型" prop="type">
          <el-radio-group v-model="editingTask.type">
            <el-radio value="database">数据库任务</el-radio>
            <el-radio value="json">JSON任务</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="数据源" prop="dataSourceId" v-if="editingTask.type === 'database'">
          <el-select 
            v-model="editingTask.dataSourceId" 
            placeholder="请选择数据源"
            @change="loadTables"
            style="width: 100%"
          >
            <el-option 
              v-for="ds in dataSourceList" 
              :key="ds.id" 
              :label="ds.name" 
              :value="ds.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="表名" prop="tableName" v-if="editingTask.type === 'database'">
          <el-select 
            v-model="editingTask.tableName" 
            placeholder="请选择表名"
            @change="loadTableStructure"
            style="width: 100%"
          >
            <el-option 
              v-for="table in tableList" 
              :key="table" 
              :label="table" 
              :value="table"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="JSON结构" prop="jsonSchema" v-if="editingTask.type === 'json'">
          <el-input 
            v-model="editingTask.jsonSchema" 
            type="textarea" 
            :rows="8" 
            placeholder="请输入JSON结构"
          />
        </el-form-item>
        
        <el-form-item label="输出类型" prop="outputType">
          <el-radio-group v-model="editingTask.outputType">
            <template v-if="editingTask.type === 'database'">
              <el-radio value="database">插入数据库</el-radio>
              <el-radio value="sql">导出SQL文件</el-radio>
            </template>
            <template v-else-if="editingTask.type === 'json'">
              <el-radio value="json">JSON文件</el-radio>
              <el-radio value="txt">TXT文件（每行一个JSON）</el-radio>
            </template>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="输出文件名" prop="outputPath" v-if="editingTask.outputType === 'sql' || editingTask.outputType === 'json' || editingTask.outputType === 'txt'">
          <el-input v-model="editingTask.outputPath" :placeholder="editingTask.outputType === 'sql' ? '请输入SQL文件名，如：data.sql' : editingTask.outputType === 'json' ? '请输入JSON文件名，如：data.json' : '请输入TXT文件名，如：data.txt'" />
        </el-form-item>
        
        <el-form-item label="生成数量" prop="count">
          <el-input-number 
            v-model="editingTask.count" 
            :min="1" 
            :max="1000000" 
            style="width: 100%"
          />
        </el-form-item>
        
        <!-- 字段规则配置 -->
        <el-form-item label="字段规则">
          <div class="field-rules">
            <div 
              v-for="field in getFields" 
              :key="field.name" 
              class="field-rule-item"
              :class="getFieldLevelClass(field.name)"
              :style="{ marginLeft: (getFieldDepth(field.name) * 20) + 'px' }"
            >
              <div class="field-info">
                <span class="field-name">{{ field.name }}</span>
                <span class="field-type">({{ field.type }})</span>
              </div>
              
              <div class="rule-config">
                <el-select 
                  :model-value="fieldRules[field.name]?.type" 
                  placeholder="选择规则类型"
                  style="width: 200px; margin-right: 10px"
                  @change="(value) => updateFieldRule(field.name, 'type', value)"
                >
                  <el-option label="随机值" value="random" />
                  <el-option label="固定值" value="fixed" />
                  <el-option label="序列" value="sequence" />
                  <el-option label="日期序列" value="date_sequence" v-if="isDateField(field)" />
                  <el-option label="正则表达式" value="regex" />
                  <el-option label="枚举" value="enum" />
                  <el-option label="引用" value="reference" />
                </el-select>
                
                <!-- 根据规则类型显示不同的参数配置 -->
                <template v-if="fieldRules[field.name]?.type === 'fixed'">
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.value" 
                    placeholder="固定值"
                    style="width: 200px"
                    @input="(value) => updateFieldRuleParam(field.name, 'value', value)"
                  />
                </template>
                
                <template v-else-if="fieldRules[field.name]?.type === 'sequence'">
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.start" 
                    placeholder="起始值（支持大整数）"
                    style="width: 150px; margin-right: 10px"
                    @input="(value) => updateFieldRuleParam(field.name, 'start', value)"
                  />
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.step" 
                    placeholder="步长"
                    style="width: 100px"
                    @input="(value) => updateFieldRuleParam(field.name, 'step', value)"
                  />
                </template>
                
                <template v-else-if="fieldRules[field.name]?.type === 'random' && isDateField(field)">
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.start" 
                    placeholder="开始日期 (YYYY-MM-DD，可选)"
                    size="small"
                    class="param-input-small"
                    @input="(value) => updateFieldRuleParam(field.name, 'start', value)"
                  />
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.end" 
                    placeholder="结束日期 (YYYY-MM-DD，可选)"
                    size="small"
                    class="param-input-small"
                    @input="(value) => updateFieldRuleParam(field.name, 'end', value)"
                  />
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.format" 
                    placeholder="日期格式 (可选)"
                    size="small"
                    class="param-input-small"
                    @input="(value) => updateFieldRuleParam(field.name, 'format', value)"
                  />
                </template>
                
                <template v-else-if="fieldRules[field.name]?.type === 'date_sequence'">
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.start" 
                    placeholder="起始日期 (YYYY-MM-DD)"
                    size="small"
                    class="param-input-small"
                    @input="(value) => updateFieldRuleParam(field.name, 'start', value)"
                  />
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.step" 
                    placeholder="步长（天数）"
                    size="small"
                    class="param-input-small"
                    @input="(value) => updateFieldRuleParam(field.name, 'step', value)"
                  />
                  <el-tooltip 
                    content="日期格式说明：2006-01-02 (年-月-日), 2006-01-02 15:04:05 (年-月-日 时:分:秒), 2006/01/02 (年/月/日), 01-02-2006 (月-日-年)。留空则使用默认格式 2006-01-02"
                    placement="top"
                    effect="dark"
                  >
                    <el-input 
                      :model-value="fieldRules[field.name]?.parameters?.format" 
                      placeholder="日期格式 (如: 2006-01-02, 可选)"
                      size="small"
                      class="param-input-small"
                      @input="(value) => updateFieldRuleParam(field.name, 'format', value)"
                    />
                  </el-tooltip>
                </template>
                
                <template v-else-if="fieldRules[field.name]?.type === 'regex'">
                  <el-autocomplete
                    :model-value="fieldRules[field.name]?.parameters?.pattern" 
                    placeholder="输入关键词或正则表达式 (如: mail, phone, name)"
                    style="width: 300px"
                    :fetch-suggestions="getRegexSuggestions"
                    @input="(value) => handleRegexInput(field.name, value)"
                    @select="(item) => handleRegexSelect(field.name, item)"
                    clearable
                  >
                    <template #default="{ item }">
                      <div class="regex-suggestion">
                        <div class="keyword">{{ item.keyword }}</div>
                        <div class="description">{{ item.description }}</div>
                        <div class="pattern">{{ item.pattern }}</div>
                      </div>
                    </template>
                  </el-autocomplete>
                </template>
                
                <template v-else-if="fieldRules[field.name]?.type === 'enum'">
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.values" 
                    placeholder="枚举值，用逗号分隔"
                    style="width: 200px"
                    @input="(value) => updateFieldRuleParam(field.name, 'values', value)"
                  />
                </template>
                
                <template v-else-if="fieldRules[field.name]?.type === 'reference'">
                  <el-input 
                    :model-value="fieldRules[field.name]?.parameters?.field" 
                    placeholder="引用字段名"
                    style="width: 200px"
                    @input="(value) => updateFieldRuleParam(field.name, 'field', value)"
                  />
                </template>
                
                <!-- 数组长度配置 -->
                <template v-if="field.name.includes('[]')">
                  <el-input-number 
                    v-model="fieldArrayLengths[field.name]" 
                    placeholder="数组长度"
                    :min="1" 
                    :max="100" 
                    style="width: 120px; margin-left: 10px"
                  />
                </template>
              </div>
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="button-group">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="updateTask" :loading="updating">
            更新
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 模板管理对话框 -->
    <el-dialog 
      v-model="templateDialogVisible" 
      title="规则模板管理" 
      width="60%"
    >
      <el-table :data="templateList" style="width: 100%">
        <el-table-column prop="name" label="模板名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="type" label="类型" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="380">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="applyTemplateToForm(row)">
              <el-icon><DocumentCopy /></el-icon>
              应用模板
            </el-button>
            <el-button size="small" type="success" @click="downloadTemplate(row)">
              <el-icon><Download /></el-icon>
              下载模板
            </el-button>
            <el-button size="small" type="danger" @click="deleteTemplate(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <template #footer>
        <div class="button-group">
          <el-button type="primary" @click="showImportTemplateFileDialog">
            <el-icon><Upload /></el-icon>
            导入模板
          </el-button>
          <el-button @click="templateDialogVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, CopyDocument, Document, Download, Delete, Upload, DocumentCopy } from '@element-plus/icons-vue'
import { taskApi, templateApi } from '@/api/task'
import { datasourceApi } from '@/api/datasource'

const loading = ref(false)
const creating = ref(false)
const updating = ref(false)
const generatingPreview = ref(false)
const dialogVisible = ref(false)
const editDialogVisible = ref(false)
const templateDialogVisible = ref(false)
const previewData = ref(null)
const editingTask = ref(null)
const templateList = ref([])
const progressTimer = ref(null)
const taskList = ref([])
const dataSourceList = ref([])
const tableList = ref([])
const tableStructure = ref([])
const jsonParseError = ref('')
const fieldRules = reactive({})
const fieldArrayLengths = reactive({})
const fieldRuleParams = reactive({})
const formRef = ref()

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  name: '',
  type: 'database',
  dataSourceId: null,
  tableName: '',
  outputType: 'database',
  outputPath: '',
  jsonSchema: '',
  count: 1000
})

const formRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择任务类型', trigger: 'change' }],
  dataSourceId: [{ required: true, message: '请选择数据源', trigger: 'change' }],
  tableName: [{ required: true, message: '请选择表名', trigger: 'change' }],
  outputType: [{ required: true, message: '请选择输出类型', trigger: 'change' }],
  outputPath: [{ required: true, message: '请输入输出文件名', trigger: 'blur' }],
  jsonSchema: [{ required: true, message: '请输入JSON结构', trigger: 'blur' }],
  count: [{ required: true, message: '请输入生成数量', trigger: 'blur' }]
}

// 加载任务列表
const loadTasks = async () => {
  loading.value = true
  try {
    const res = await taskApi.getList({
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    taskList.value = res.data?.list || []
    pagination.total = res.data?.total || 0
  } catch (error) {
    console.error('加载任务列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 加载数据源列表
const loadDataSources = async () => {
  try {
    const res = await datasourceApi.getList()
    dataSourceList.value = res.data || []
  } catch (error) {
    console.error('加载数据源列表失败:', error)
  }
}

// 加载表列表
const loadTables = async () => {
  // 在编辑模式下使用editingTask，否则使用formData
  const currentData = editingTask.value || formData
  if (!currentData.dataSourceId) return
  
  try {
    const res = await datasourceApi.getTables(currentData.dataSourceId)
    tableList.value = res.data || []
  } catch (error) {
    console.error('加载表列表失败:', error)
  }
}

// 处理表名变化事件
const onTableNameChange = (tableName) => {
  console.log('表名发生变化:', tableName)
  console.log('当前formData.tableName:', formData.tableName)
  console.log('当前formData.dataSourceId:', formData.dataSourceId)
  loadTableStructure()
}

// 加载表结构
const loadTableStructure = async () => {
  // 在编辑模式下使用editingTask，否则使用formData
  const currentData = editingTask.value || formData
  console.log('loadTableStructure 调用 - 数据源ID:', currentData.dataSourceId, '表名:', currentData.tableName)
  
  if (!currentData.dataSourceId || !currentData.tableName) {
    console.log('数据源ID或表名为空，跳过加载表结构')
    return
  }
  
  try {
    console.log('正在请求表结构:', `/api/datasource/table/${currentData.dataSourceId}/${currentData.tableName}`)
    const res = await datasourceApi.getTableStructure(currentData.dataSourceId, currentData.tableName)
    tableStructure.value = res.data?.columns || []
    console.log('表结构加载成功，列数:', tableStructure.value.length)
  } catch (error) {
    console.error('加载表结构失败:', error)
  }
}

// 递归解析JSON结构，生成扁平化的字段路径
const parseJSONFields = (obj, prefix = '', depth = 0) => {
  const fields = []
  
  if (typeof obj !== 'object' || obj === null) {
    return fields
  }
  
  // 防止无限递归，设置最大深度限制
  if (depth > 10) {
    return fields
  }
  
  for (const [key, value] of Object.entries(obj)) {
    const fieldPath = prefix ? `${prefix}.${key}` : key
    
    if (Array.isArray(value)) {
      // 数组类型
      fields.push({
        name: fieldPath,
        type: 'array',
        originalType: 'array'
      })
      
      // 如果数组有元素，递归解析数组元素的结构
      if (value.length > 0) {
        const arrayElementPath = `${fieldPath}[]`
        if (typeof value[0] === 'object' && value[0] !== null) {
          const arrayFields = parseJSONFields(value[0], arrayElementPath, depth + 1)
          fields.push(...arrayFields)
        } else {
          // 数组元素是基本类型
          fields.push({
            name: arrayElementPath,
            type: typeof value[0] === 'string' ? value[0] : typeof value[0],
            originalType: typeof value[0]
          })
        }
      }
    } else if (typeof value === 'object' && value !== null) {
      // 嵌套对象
      fields.push({
        name: fieldPath,
        type: 'object',
        originalType: 'object'
      })
      
      const nestedFields = parseJSONFields(value, fieldPath, depth + 1)
      fields.push(...nestedFields)
    } else {
      // 基本类型
      fields.push({
        name: fieldPath,
        type: typeof value === 'string' ? value : typeof value,
        originalType: typeof value
      })
    }
  }
  
  return fields
}

// 获取字段列表（计算属性）
const getFields = computed(() => {
  // 在编辑模式下使用editingTask，否则使用formData
  const currentData = editingTask.value || formData
  
  console.log('getFields 计算属性触发 - 任务类型:', currentData.type, 'JSON结构:', currentData.jsonSchema)
  
  if (currentData.type === 'database') {
    console.log('数据库任务，返回表结构字段数:', tableStructure.value.length)
    return tableStructure.value
  } else {
    // JSON任务，从JSON结构中解析字段
    try {
      if (!currentData.jsonSchema || currentData.jsonSchema.trim() === '') {
        console.log('JSON结构为空')
        jsonParseError.value = ''
        return []
      }
      
      console.log('尝试解析JSON:', currentData.jsonSchema)
      const schema = JSON.parse(currentData.jsonSchema)
      console.log('JSON解析成功:', schema)
      jsonParseError.value = '' // 清除错误信息
      const fields = parseJSONFields(schema)
      console.log('解析出的字段数:', fields.length, '字段列表:', fields)
      return fields
    } catch (error) {
      console.error('JSON解析失败:', error.message)
      console.error('输入的JSON内容:', currentData.jsonSchema)
      console.error('提示: JSON格式要求属性名必须用双引号包围，例如: {"name": "rami"}')
      jsonParseError.value = `JSON格式错误: ${error.message}。请确保属性名用双引号包围，例如: {"name": "rami"}`
      return []
    }
  }
})

// 计算字段缩进
const getFieldIndent = (fieldName) => {
  // 计算嵌套深度
  const dotCount = (fieldName.match(/\./g) || []).length
  const bracketCount = (fieldName.match(/\[\]/g) || []).length
  const depth = dotCount + bracketCount
  return depth * 20 // 每层缩进20px
}

// 获取字段显示名称（简化显示）
const getDisplayFieldName = (fieldName) => {
  // 如果字段名包含路径，只显示最后一部分
  const parts = fieldName.split('.')
  const lastPart = parts[parts.length - 1]
  
  // 如果是数组元素，显示更友好的格式
  if (lastPart.includes('[]')) {
    return lastPart.replace('[]', '[*]')
  }
  
  return lastPart
}

// 计算字段层级深度
const getFieldDepth = (fieldName) => {
  const dotCount = (fieldName.match(/\./g) || []).length
  const arrayCount = (fieldName.match(/\[\]/g) || []).length
  return dotCount + arrayCount
}

// 获取字段层级类名
const getFieldLevelClass = (fieldName) => {
  const depth = getFieldDepth(fieldName)
  const hasArray = fieldName.includes('[]')
  const hasNested = fieldName.includes('.')
  
  if (hasNested && hasArray) {
    return 'nested-array-field'
  } else if (hasArray) {
    return 'array-field'
  } else if (hasNested) {
    return 'nested-field'
  }
  return ''
}

// 获取字段路径（用于显示层级关系）
const getFieldPath = (fieldName) => {
  const parts = fieldName.split('.')
  if (parts.length <= 1) return ''
  
  // 显示路径，但不包含最后一部分
  const pathParts = parts.slice(0, -1)
  return pathParts.join('.').replace(/\[\]/g, '[*]') + ' →'
}

// 判断字段是否为日期类型
const isDateField = (field) => {
  if (!field || !field.type) {
    return false
  }
  const fieldType = field.type.toLowerCase()
  
  // 支持常见的日期时间类型，包括各种数据库的日期时间字段
  if (fieldType.includes('date') || fieldType.includes('time') || 
      fieldType === 'datetime' || fieldType === 'timestamp' ||
      fieldType === 'year' || fieldType.includes('datetime') ||
      // 支持更多数据库特定的日期时间类型
      fieldType === 'timestamptz' || fieldType === 'timetz' ||
      fieldType.startsWith('datetime') || fieldType.startsWith('timestamp')) {
    return true
  }
  
  // 对于JSON任务，检查字段值是否为日期格式
  // 匹配常见的日期格式：YYYY-MM-DD, YYYY-MM-DD HH:mm:ss 等
  const datePattern = /^\d{4}-\d{2}-\d{2}(\s\d{2}:\d{2}(:\d{2})?)?$/
  if (datePattern.test(fieldType)) {
    return true
  }
  
  return false
}

// 规则类型变化处理
const onRuleTypeChange = (fieldName) => {
  const ruleType = fieldRules[fieldName]
  
  // 初始化规则参数
  if (!fieldRuleParams[fieldName]) {
    fieldRuleParams[fieldName] = {}
  }
  
  // 根据规则类型初始化默认参数
  switch (ruleType) {
    case 'fixed':
      fieldRuleParams[fieldName] = { value: '' }
      break
    case 'sequence':
      fieldRuleParams[fieldName] = { start: 1, step: 1 }
      break
    case 'date_sequence':
      fieldRuleParams[fieldName] = { start: '', step: 1, format: '' }
      break
    case 'range':
      fieldRuleParams[fieldName] = { min: '', max: '' }
      break
    case 'regex':
      fieldRuleParams[fieldName] = { pattern: '' }
      break
    case 'enum':
      fieldRuleParams[fieldName] = { values: '' }
      break
    case 'custom':
      fieldRuleParams[fieldName] = { script: '' }
      break
    default:
      fieldRuleParams[fieldName] = {}
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  resetForm() // 重置表单数据
  dialogVisible.value = true
  loadDataSources()
}

// 重置表单
const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(formData, {
    name: '',
    type: 'database',
    dataSourceId: null,
    tableName: '',
    outputType: 'database',
    outputPath: '',
    jsonSchema: '',
    count: 1000
  })
  
  // 根据任务类型设置默认输出类型
  if (formData.type === 'json') {
    formData.outputType = 'json'
  }
  tableList.value = []
  tableStructure.value = []
  previewData.value = null
  Object.keys(fieldRules).forEach(key => {
    delete fieldRules[key]
  })
  Object.keys(fieldArrayLengths).forEach(key => {
    delete fieldArrayLengths[key]
  })
  Object.keys(fieldRuleParams).forEach(key => {
    delete fieldRuleParams[key]
  })
}

// 创建任务
const createTask = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    creating.value = true
    
    // 合并字段规则、规则参数和数组长度配置
    const mergedFieldRules = {}
    
    // 处理基本规则和参数
    Object.keys(fieldRules).forEach(fieldName => {
      const rule = fieldRules[fieldName]
      const ruleType = typeof rule === 'string' ? rule : (rule?.type || 'random')
      const ruleParams = typeof rule === 'object' && rule?.parameters ? rule.parameters : (fieldRuleParams[fieldName] || {})
      
      mergedFieldRules[fieldName] = {
        type: ruleType,
        parameters: { ...ruleParams }
      }
    })
    
    // 添加数组长度配置
    Object.keys(fieldArrayLengths).forEach(fieldName => {
      if (fieldArrayLengths[fieldName] && fieldArrayLengths[fieldName] > 0) {
        if (!mergedFieldRules[fieldName]) {
          mergedFieldRules[fieldName] = { type: 'random', parameters: {} }
        }
        mergedFieldRules[fieldName].parameters.length = fieldArrayLengths[fieldName]
      }
    })
    
    const taskData = {
      ...formData,
      fieldRules: JSON.stringify(mergedFieldRules)
    }
    
    await taskApi.create(taskData)
    ElMessage.success('任务创建成功')
    dialogVisible.value = false
    loadTasks()
  } catch (error) {
    if (error.errors) {
      ElMessage.error('请完善表单信息')
    } else {
      console.error('创建任务失败:', error)
    }
  } finally {
    creating.value = false
  }
}

// 执行任务
const executeTask = async (row) => {
  try {
    await taskApi.execute(row.id)
    ElMessage.success('任务已开始执行')
    loadTasks()
  } catch (error) {
    console.error('执行任务失败:', error)
  }
}

// 查看任务
const viewTask = async (row) => {
  try {
    const res = await taskApi.getById(row.id)
    ElMessageBox.alert(
      `<div>
        <p><strong>任务名称:</strong> ${res.data.name}</p>
        <p><strong>类型:</strong> ${res.data.type === 'database' ? '数据库' : 'JSON'}</p>
        <p><strong>状态:</strong> ${getStatusText(res.data.status)}</p>
        <p><strong>进度:</strong> ${res.data.progress}%</p>
        <p><strong>生成数量:</strong> ${res.data.count}</p>
        <p><strong>创建时间:</strong> ${formatTime(res.data.created_at)}</p>
        ${res.data.error_msg ? `<p><strong>错误信息:</strong> ${res.data.error_msg}</p>` : ''}
      </div>`,
      '任务详情',
      {
        dangerouslyUseHTMLString: true
      }
    )
  } catch (error) {
    console.error('获取任务详情失败:', error)
  }
}

// 删除任务
const deleteTask = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除任务 "${row.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await taskApi.delete(row.id)
    ElMessage.success('删除成功')
    loadTasks()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除任务失败:', error)
    }
  }
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
    pending: '等待中',
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

// 监听数据源变化
watch(() => formData.dataSourceId, () => {
  formData.tableName = ''
  tableList.value = []
  tableStructure.value = []
  if (formData.dataSourceId) {
    loadTables()
  }
})

// 监听任务类型变化，自动设置默认输出类型
watch(() => formData.type, (newType) => {
  if (newType === 'database') {
    formData.outputType = 'database'
  } else if (newType === 'json') {
    formData.outputType = 'json'
  }
})

// 监听JSON结构变化
watch(() => formData.jsonSchema, (newValue, oldValue) => {
  // 只有当JSON结构真正改变且不为空时才清理
  if (newValue && newValue !== oldValue && newValue.trim() !== '') {
    console.log('JSON结构发生变化，清理字段规则数据')
    // 当JSON结构改变时，清理旧的字段规则数据
    Object.keys(fieldRules).forEach(key => {
      delete fieldRules[key]
    })
    Object.keys(fieldArrayLengths).forEach(key => {
      delete fieldArrayLengths[key]
    })
    Object.keys(fieldRuleParams).forEach(key => {
      delete fieldRuleParams[key]
    })
    // 清理预览数据
    previewData.value = null
  }
}, { flush: 'post' })

// 生成预览数据
const generatePreviewData = async () => {
  try {
    generatingPreview.value = true
    
    // 构建预览数据请求
    const fields = getFields.value
    if (fields.length === 0) {
      ElMessage.warning('请先配置字段信息')
      return
    }
    
    // 合并字段规则、规则参数和数组长度配置
    const mergedFieldRules = {}
    
    // 处理基本规则和参数
    Object.keys(fieldRules).forEach(fieldName => {
      const rule = fieldRules[fieldName]
      const ruleType = typeof rule === 'string' ? rule : (rule?.type || 'random')
      const ruleParams = typeof rule === 'object' && rule?.parameters ? rule.parameters : (fieldRuleParams[fieldName] || {})
      
      mergedFieldRules[fieldName] = {
        type: ruleType,
        parameters: { ...ruleParams }
      }
    })
    
    // 添加数组长度配置
    Object.keys(fieldArrayLengths).forEach(fieldName => {
      if (fieldArrayLengths[fieldName] && fieldArrayLengths[fieldName] > 0) {
        if (!mergedFieldRules[fieldName]) {
          mergedFieldRules[fieldName] = { type: 'random', parameters: {} }
        }
        mergedFieldRules[fieldName].parameters.length = fieldArrayLengths[fieldName]
      }
    })
    
    // 为没有配置规则的字段设置默认规则
    fields.forEach(field => {
      if (!mergedFieldRules[field.name]) {
        mergedFieldRules[field.name] = {
          type: 'random',
          parameters: {}
        }
      }
    })
    
    const previewRequest = {
      ...formData,
      count: 1, // 预览只生成一条数据
      fieldRules: JSON.stringify(mergedFieldRules)
    }
    
    // 调用API生成预览数据
    const res = await taskApi.preview(previewRequest)
    previewData.value = res.data
    ElMessage.success('预览数据生成成功')
    
  } catch (error) {
    console.error('生成预览数据失败:', error)
    ElMessage.error('生成预览数据失败')
  } finally {
    generatingPreview.value = false
  }
}

// 格式化预览数据
const formatPreviewData = (data) => {
  if (!data) return ''
  return JSON.stringify(data, null, 2)
}

// 复制预览数据
const copyPreviewData = async () => {
  try {
    const text = formatPreviewData(previewData.value)
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败')
  }
}

// 更新字段规则
const updateFieldRule = (fieldName, property, value) => {
  if (!fieldRules[fieldName]) {
    fieldRules[fieldName] = { type: 'random', parameters: {} }
  }
  if (property === 'type') {
    fieldRules[fieldName].type = value
    // 重置参数
    fieldRules[fieldName].parameters = {}
  } else {
    fieldRules[fieldName][property] = value
  }
}

// 更新字段规则参数
const updateFieldRuleParam = (fieldName, paramName, value) => {
  if (!fieldRules[fieldName]) {
    fieldRules[fieldName] = { type: 'random', parameters: {} }
  }
  if (!fieldRules[fieldName].parameters) {
    fieldRules[fieldName].parameters = {}
  }
  fieldRules[fieldName].parameters[paramName] = value
}

// 关键词到正则表达式的映射
const regexKeywords = {
  'mail': {
    keyword: 'mail',
    description: '邮箱地址',
    pattern: '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}'
  },
  'email': {
    keyword: 'email',
    description: '邮箱地址',
    pattern: '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}'
  },
  'phone': {
    keyword: 'phone',
    description: '手机号码',
    pattern: '1[3-9]\\d{9}'
  },
  'mobile': {
    keyword: 'mobile',
    description: '手机号码',
    pattern: '1[3-9]\\d{9}'
  },
  'name': {
    keyword: 'name',
    description: '中文姓名',
    pattern: '[\\u4e00-\\u9fa5]{2,4}'
  },
  'chinese': {
    keyword: 'chinese',
    description: '中文字符',
    pattern: '[\\u4e00-\\u9fa5]+'
  },
  'english': {
    keyword: 'english',
    description: '英文字母',
    pattern: '[a-zA-Z]+'
  },
  'number': {
    keyword: 'number',
    description: '数字',
    pattern: '\\d+'
  },
  'id': {
    keyword: 'id',
    description: '身份证号',
    pattern: '[1-9]\\d{5}(18|19|20)\\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\\d{3}[0-9Xx]'
  },
  'url': {
    keyword: 'url',
    description: '网址链接',
    pattern: 'https?://[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}(/[a-zA-Z0-9._~:/?#[\\]@!$&\'()*+,;=-]*)?'
  },
  'ip': {
    keyword: 'ip',
    description: 'IP地址',
    pattern: '((25[0-5]|2[0-4]\\d|[01]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[01]?\\d\\d?)'
  },
  'date': {
    keyword: 'date',
    description: '日期格式 YYYY-MM-DD',
    pattern: '\\d{4}-\\d{2}-\\d{2}'
  },
  'time': {
    keyword: 'time',
    description: '时间格式 HH:MM:SS',
    pattern: '\\d{2}:\\d{2}:\\d{2}'
  },
  'uuid': {
    keyword: 'uuid',
    description: 'UUID格式',
    pattern: '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
  }
}

// 获取正则表达式建议
const getRegexSuggestions = (queryString, cb) => {
  const suggestions = Object.values(regexKeywords).filter(item => {
    return item.keyword.toLowerCase().includes(queryString.toLowerCase()) ||
           item.description.includes(queryString)
  })
  cb(suggestions)
}

// 处理正则表达式输入
const handleRegexInput = (fieldName, value) => {
  // 检查是否是关键词
  const keyword = regexKeywords[value.toLowerCase()]
  if (keyword) {
    // 如果是关键词，自动转换为对应的正则表达式
    updateFieldRuleParam(fieldName, 'pattern', keyword.pattern)
  } else {
    // 否则直接使用输入的值
    updateFieldRuleParam(fieldName, 'pattern', value)
  }
}

// 处理正则表达式选择
const handleRegexSelect = (fieldName, item) => {
  updateFieldRuleParam(fieldName, 'pattern', item.pattern)
}

// 编辑任务
const editTask = async (task) => {
  editingTask.value = { ...task }
  
  // 输出文件名由用户输入，不需要设置默认值
  
  // 清空当前规则
  Object.keys(fieldRules).forEach(key => {
    delete fieldRules[key]
  })
  
  // 清空当前规则参数
  Object.keys(fieldRuleParams).forEach(key => {
    delete fieldRuleParams[key]
  })
  
  // 解析字段规则
  if (task.fieldRules) {
    try {
      const rules = JSON.parse(task.fieldRules)
      // 确保规则格式正确
      Object.keys(rules).forEach(fieldName => {
        const rule = rules[fieldName]
        if (typeof rule === 'string') {
          // 兼容旧格式：字符串类型
          fieldRules[fieldName] = {
            type: rule,
            parameters: {}
          }
          fieldRuleParams[fieldName] = {}
        } else if (rule && typeof rule === 'object') {
          // 新格式：对象类型
          fieldRules[fieldName] = {
            type: rule.type || 'random',
            parameters: rule.parameters || {}
          }
          // 同步更新UI参数
          fieldRuleParams[fieldName] = rule.parameters || {}
        }
      })
    } catch (error) {
      console.error('解析字段规则失败:', error)
    }
  }
  
  editDialogVisible.value = true
  loadDataSources()
  if (task.dataSourceId) {
    await loadTables()
    if (task.tableName) {
      await loadTableStructure()
    }
  }
}

// 更新任务
const updateTask = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    updating.value = true
    
    const fields = getFields.value
    const mergedFieldRules = {}
    
    // 处理基本规则和参数
    Object.keys(fieldRules).forEach(fieldName => {
      const rule = fieldRules[fieldName]
      const ruleType = typeof rule === 'string' ? rule : (rule?.type || 'random')
      const ruleParams = typeof rule === 'object' && rule?.parameters ? rule.parameters : (fieldRuleParams[fieldName] || {})
      
      mergedFieldRules[fieldName] = {
        type: ruleType,
        parameters: { ...ruleParams }
      }
    })
    
    // 为没有配置规则的字段设置默认规则
    fields.forEach(field => {
      if (!mergedFieldRules[field.name]) {
        mergedFieldRules[field.name] = {
          type: 'random',
          parameters: {}
        }
      }
    })
    
    const updateData = {
      ...editingTask.value,
      fieldRules: JSON.stringify(mergedFieldRules)
    }
    
    await taskApi.update(editingTask.value.id, updateData)
    ElMessage.success('任务更新成功')
    editDialogVisible.value = false
    resetForm()
    loadTasks()
  } catch (error) {
    console.error('更新任务失败:', error)
    ElMessage.error('更新任务失败')
  } finally {
    updating.value = false
  }
}

// 导出模板
const exportTemplate = async (task) => {
  try {
    const templateName = await ElMessageBox.prompt('请输入模板名称', '导出规则模板', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /\S+/,
      inputErrorMessage: '模板名称不能为空'
    })
    
    await taskApi.exportTemplate(task.id, {
      name: templateName.value,
      description: `从任务 "${task.name}" 导出的规则模板`
    })
    
    ElMessage.success('模板导出成功')
    loadTemplates()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('导出模板失败:', error)
      ElMessage.error('导出模板失败')
    }
  }
}

// 显示模板管理对话框
const showTemplateDialog = () => {
  templateDialogVisible.value = true
  loadTemplates()
}

// 加载模板列表
const loadTemplates = async () => {
  try {
    const res = await templateApi.getList()
    templateList.value = res.data || []
    console.log(templateList.value)
  } catch (error) {
    console.error('加载模板列表失败:', error)
  }
}

// 应用模板到创建任务表单（在创建任务对话框中使用）
const applyTemplateToForm = async (template) => {
  try {
    // 如果创建任务对话框未打开，先打开它
    if (!dialogVisible.value) {
      showCreateDialog()
    }
    
    // 解析模板的字段规则
    const rules = JSON.parse(template.fieldRules || '{}')
    
    // 清空当前规则
    Object.keys(fieldRules).forEach(key => {
      delete fieldRules[key]
    })
    Object.keys(fieldRuleParams).forEach(key => {
      delete fieldRuleParams[key]
    })
    Object.keys(fieldArrayLengths).forEach(key => {
      delete fieldArrayLengths[key]
    })
    
    // 应用模板规则
    Object.keys(rules).forEach(fieldName => {
      const rule = rules[fieldName]
      if (typeof rule === 'string') {
        // 兼容旧格式：字符串类型
        fieldRules[fieldName] = rule
        fieldRuleParams[fieldName] = {}
      } else if (rule && typeof rule === 'object') {
        // 新格式：对象类型
        fieldRules[fieldName] = rule.type || 'random'
        fieldRuleParams[fieldName] = rule.parameters || {}
        
        // 处理数组长度
        if (rule.parameters && rule.parameters.length) {
          fieldArrayLengths[fieldName] = rule.parameters.length
        }
      }
    })
    
    // 设置任务类型和JSON结构
    formData.type = template.type
    if (template.jsonSchema) {
      formData.jsonSchema = template.jsonSchema
    }
    
    ElMessage.success(`模板"${template.name}"已应用到创建任务表单`)
    templateDialogVisible.value = false
  } catch (error) {
    console.error('应用模板失败:', error)
    ElMessage.error('应用模板失败')
  }
}

// 下载模板为JSON文件
const downloadTemplate = (template) => {
  try {
    const templateData = {
      name: template.name,
      description: template.description,
      type: template.type,
      jsonSchema: template.jsonSchema,
      fieldRules: template.fieldRules
    }
    
    const dataStr = JSON.stringify(templateData, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json;charset=utf-8' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${template.name || 'template'}_${Date.now()}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success('模板下载成功')
  } catch (error) {
    console.error('下载模板失败:', error)
    ElMessage.error('下载模板失败')
  }
}

// 显示导入模板文件对话框
const showImportTemplateFileDialog = () => {
  // 创建隐藏的文件输入元素
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.style.display = 'none'
  
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    
    try {
      const text = await file.text()
      const templateData = JSON.parse(text)
      
      // 验证模板格式
      if (!templateData.name || !templateData.type) {
        ElMessage.error('模板格式错误：缺少必要字段（name、type）')
        return
      }
      
      // 调用API导入模板到数据库
      await templateApi.import(templateData)
      ElMessage.success('模板导入成功')
      loadTemplates()
    } catch (error) {
      console.error('导入模板失败:', error)
      if (error.response && error.response.data && error.response.data.error) {
        ElMessage.error('导入模板失败：' + error.response.data.error)
      } else {
        ElMessage.error('导入模板失败：' + (error.message || '文件格式错误'))
      }
    }
  }
  
  input.oncancel = () => {
    // 用户取消选择文件
  }
  
  document.body.appendChild(input)
  input.click()
  // 延迟删除，确保文件选择对话框已关闭
  setTimeout(() => {
    document.body.removeChild(input)
  }, 100)
}

// 删除模板
const deleteTemplate = async (template) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${template.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await templateApi.delete(template.id)
    ElMessage.success('模板删除成功')
    loadTemplates()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除模板失败:', error)
      ElMessage.error('删除模板失败')
    }
  }
}

// 检查运行中的任务并更新进度
const checkRunningTasks = async () => {
  const runningTasks = taskList.value.filter(task => task.status === 'running')
  if (runningTasks.length === 0) return
  
  try {
    for (const task of runningTasks) {
      const res = await taskApi.getStatus(task.id)
      if (res.data) {
        // 更新任务列表中对应任务的状态和进度
        const index = taskList.value.findIndex(t => t.id === task.id)
        if (index !== -1) {
          taskList.value[index].status = res.data.status
          taskList.value[index].progress = res.data.progress
          taskList.value[index].error_msg = res.data.error_msg
        }
      }
    }
  } catch (error) {
    console.error('检查任务状态失败:', error)
  }
}

onMounted(() => {
  loadTasks()
  // 启动定时器，每3秒检查一次运行中的任务进度
  progressTimer.value = setInterval(checkRunningTasks, 3000)
})

onUnmounted(() => {
  // 清理定时器
  if (progressTimer.value) {
    clearInterval(progressTimer.value)
    progressTimer.value = null
  }
})
</script>

<style scoped>
/* ============================================
   任务管理页面样式
   ============================================ */
.task {
  max-width: 100%;
}

.field-rules {
  border: 1px solid var(--border-light, #e4e7ed);
  border-radius: var(--radius-md, 8px);
  padding: 20px;
  max-height: 500px;
  overflow-y: auto;
  min-height: 200px;
  background: var(--bg-secondary, #f5f7fa);
  transition: var(--transition-base, all 0.3s ease);
}

.field-rules:hover {
  border-color: var(--primary-color, #409EFF);
}

.field-rule-item {
  display: flex;
  flex-direction: column;
  margin-bottom: 16px;
  padding: 16px;
  background: #ffffff;
  border-radius: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  gap: 12px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.field-rule-item:last-child {
  margin-bottom: 0;
}

.field-rule-item:hover {
  background: #f0f9ff;
  border-color: #409EFF;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
  transform: translateY(-2px);
}

/* 嵌套字段样式 */
.nested-field {
  border-left: 4px solid #409eff;
  background: linear-gradient(90deg, #ecf5ff 0%, #ffffff 100%);
}

.nested-field .field-name {
  color: #409eff;
  font-weight: 600;
}

/* 数组字段样式 */
.array-field {
  border-left: 4px solid #67c23a;
  background: linear-gradient(90deg, #f0f9ff 0%, #ffffff 100%);
}

.array-field .field-name {
  color: #67c23a;
  font-weight: 600;
}

/* 深层嵌套字段样式 */
.field-rule-item.nested-array-field {
  border-left: 4px solid #f56c6c;
  background: linear-gradient(90deg, #fef0f0 0%, #ffffff 100%);
}

.field-rule-item.nested-array-field .field-name {
  color: #f56c6c;
  font-weight: 600;
}

.header-buttons {
  display: flex;
  gap: 12px;
}

.header-buttons .el-button {
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
  transition: all 0.3s ease;
}

.header-buttons .el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}

.button-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.field-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.field-name {
  font-weight: 600;
  color: #303133;
  font-size: 14px;
}

.field-type {
  color: #909399;
  font-size: 12px;
}

.rule-config {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

/* 对象字段样式 */
.object-field {
  border-left: 3px solid #e6a23c;
  background: #fdf6ec;
}

.object-field .field-name {
  color: #e6a23c;
}

/* 预览数据样式 */
.preview-data-container {
  width: 100%;
}

.preview-card {
  margin-top: 12px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #ffffff 100%);
  border-bottom: 1px solid #e4e7ed;
}

.preview-content {
  background: #1e1e1e;
  border: none;
  border-radius: 0;
  padding: 20px;
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #d4d4d4;
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
  box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.2);
}

.preview-content::-webkit-scrollbar {
  width: 8px;
}

.preview-content::-webkit-scrollbar-track {
  background: #252526;
}

.preview-content::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 4px;
}

.preview-content::-webkit-scrollbar-thumb:hover {
  background: #4e4e4e;
}
.object-field {
  border-left: 3px solid #e6a23c;
  background: #fdf6ec;
}

.object-field .field-name {
  color: #e6a23c;
}

.field-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  margin-right: 15px;
}

.field-path {
  font-size: 12px;
  color: #999;
  font-style: italic;
  margin-bottom: 2px;
}

.field-name {
  font-weight: 500;
  color: #303133;
  margin-bottom: 2px;
  font-size: 14px;
  word-break: break-all;
}

.field-config {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.rule-select {
  min-width: 150px;
  flex: 1;
}

.array-length-input {
  width: 80px;
}

.rule-config-row {
  display: flex;
  align-items: flex-start;
  gap: 15px;
  flex-wrap: wrap;
}

.rule-params {
  margin-top: 8px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.param-input {
  width: 100%;
}

/* 正则表达式建议样式 */
.regex-suggestion {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.regex-suggestion:last-child {
  border-bottom: none;
}

.regex-suggestion .keyword {
  font-weight: bold;
  color: #409eff;
  font-size: 14px;
  margin-bottom: 4px;
}

.regex-suggestion .description {
  color: #666;
  font-size: 12px;
  margin-bottom: 4px;
}

.regex-suggestion .pattern {
  font-family: 'Courier New', monospace;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  color: #333;
  word-break: break-all;
}

.param-input-small {
  width: 120px;
}

.sequence-config,
.range-config,
.date-sequence-config,
.random-date-config {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.date-sequence-config .param-input-small,
.random-date-config .param-input-small {
  min-width: 140px;
}

.sequence-config::before {
  content: '序列:';
  font-size: 12px;
  color: #666;
  margin-right: 5px;
}

.range-config::before {
  content: '范围:';
  font-size: 12px;
  color: #666;
  margin-right: 5px;
}

.field-type {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 3px;
  color: #fff;
  font-weight: 500;
}

.type-string {
  background: #909399;
}

.type-number {
  background: #f56c6c;
}

.type-array {
  background: #67c23a;
}

.type-object {
  background: #e6a23c;
}

.type-boolean {
  background: #409eff;
}

.field-type {
  font-size: 12px;
  color: #909399;
}

.field-rule-item .el-select {
  width: 150px;
  margin-left: 15px;
}

.json-error-tip {
  margin-top: 8px;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 4px 0;
}

.dialog-title {
  font-size: 20px;
  font-weight: 600;
  background: linear-gradient(135deg, #409EFF, #66b1ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 对话框样式 */
.task-dialog :deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.15);
}

.task-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-bottom: 1px solid #e4e7ed;
}

.task-dialog :deep(.el-dialog__body) {
  padding: 24px;
  background: #ffffff;
}

.task-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  background: #f8f9fa;
  border-top: 1px solid #e4e7ed;
}

/* 表单容器样式 */
.form-container {
  width: 100%;
}

/* 表单项全宽样式 */
.form-item-full {
  width: 100% !important;
}

/* 确保表单项内容区域能够充分利用空间 */
.form-container :deep(.el-form-item) {
  width: 100%;
  margin-bottom: 22px;
}

.form-container :deep(.el-form-item__content) {
  flex: 1;
  width: 100%;
  max-width: 100%;
  display: flex;
}

/* 确保输入框全宽显示 */
.form-container :deep(.el-input) {
  width: 100% !important;
  flex: 1;
}

.form-container :deep(.el-input__wrapper) {
  width: 100% !important;
}

/* 确保选择器全宽显示 */
.form-container :deep(.el-select) {
  width: 100% !important;
  flex: 1;
}

.form-container :deep(.el-select .el-input) {
  width: 100% !important;
}

.form-container :deep(.el-select .el-input__wrapper) {
  width: 100% !important;
}

/* 确保数字输入框全宽 */
.form-container :deep(.el-input-number) {
  width: 100% !important;
  flex: 1;
}

.form-container :deep(.el-input-number .el-input) {
  width: 100% !important;
}

.form-container :deep(.el-input-number .el-input__wrapper) {
  width: 100% !important;
}

/* 确保文本域全宽 */
.form-container :deep(.el-textarea) {
  width: 100% !important;
  flex: 1;
}

.form-container :deep(.el-textarea__inner) {
  width: 100% !important;
  min-width: 100%;
}

@media (max-width: 768px) {
  .field-rule-item {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .field-rule-item .el-select {
    width: 100%;
    margin-left: 0;
    margin-top: 10px;
  }
}
</style>