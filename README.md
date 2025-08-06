# 数据生成平台

一个功能强大的数据生成平台，支持基于数据库表结构生成测试数据，以及生成复杂JSON数据用于压力测试。

## 功能特性

### 核心功能
- 🗄️ **多数据库支持**: 支持MySQL、PostgreSQL、SQLite
- 📊 **智能数据生成**: 基于数据库表结构自动生成测试数据
- 📝 **JSON数据生成**: 支持复杂JSON结构的数据生成
- ⚡ **高性能处理**: 支持大批量数据生成和导出
- 🎯 **灵活规则配置**: 支持多种数据生成规则

### 数据生成规则
- **固定值**: 生成固定的数据值
- **序列**: 生成递增序列数据
- **随机**: 生成随机数据
- **范围**: 在指定范围内生成数据
- **正则表达式**: 基于正则表达式生成数据
- **枚举**: 从预定义列表中随机选择
- **UUID**: 生成唯一标识符
- **自定义**: 支持自定义生成逻辑

### 输出格式
- **数据库插入**: 直接插入到目标数据库
- **SQL文件**: 导出为SQL插入语句
- **JSON文件**: 导出为JSON格式文件

## 技术架构

### 后端技术栈
- **Go 1.19+**: 主要开发语言
- **Gin**: Web框架
- **GORM**: ORM框架
- **SQLite**: 默认数据库（支持MySQL、PostgreSQL）

### 前端技术栈
- **Vue 3**: 前端框架
- **Element Plus**: UI组件库
- **Vite**: 构建工具
- **Pinia**: 状态管理
- **Axios**: HTTP客户端

## 项目结构

```
generateTestData/
├── backend/                 # 后端代码
│   ├── config/             # 配置管理
│   ├── controllers/        # 控制器层
│   ├── models/             # 数据模型
│   ├── services/           # 业务逻辑层
│   ├── utils/              # 工具函数
│   └── main.go             # 程序入口
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── api/            # API接口
│   │   ├── components/     # 组件
│   │   ├── views/          # 页面
│   │   ├── router/         # 路由配置
│   │   └── utils/          # 工具函数
│   ├── package.json
│   └── vite.config.js
├── uploads/                # 文件上传目录
├── go.mod                  # Go模块文件
└── README.md               # 项目说明
```

## 快速开始

### 环境要求
- Go 1.19+
- Node.js 16+
- npm 或 yarn

### 安装依赖

#### 后端依赖
```bash
cd backend
go mod tidy
```

#### 前端依赖
```bash
cd frontend
npm install
```

### 运行项目

#### 启动后端服务
```bash
cd backend
go run main.go
```
后端服务将在 `http://localhost:8080` 启动

#### 启动前端服务
```bash
cd frontend
npm run dev
```
前端服务将在 `http://localhost:3000` 启动

### 构建部署

#### 构建后端
```bash
cd backend
go build -o generateTestData main.go
```

#### 构建前端
```bash
cd frontend
npm run build
```

## 使用指南

### 1. 配置数据源
1. 进入「数据源管理」页面
2. 点击「添加数据源」
3. 填写数据库连接信息
4. 测试连接成功后保存

### 2. 创建数据生成任务
1. 进入「任务管理」页面
2. 点击「创建任务」
3. 选择任务类型（数据库任务或JSON任务）
4. 配置生成规则和参数
5. 保存并执行任务

### 3. 监控任务执行
- 在任务列表中查看任务状态和进度
- 支持实时进度更新
- 查看任务执行结果和错误信息

## API文档

### 数据源管理
- `GET /datasources` - 获取数据源列表
- `POST /datasources` - 创建数据源
- `PUT /datasources/:id` - 更新数据源
- `DELETE /datasources/:id` - 删除数据源
- `POST /datasources/test` - 测试数据源连接
- `GET /datasources/:id/tables` - 获取表列表
- `GET /datasources/:id/tables/:table` - 获取表结构

### 任务管理
- `GET /tasks` - 获取任务列表
- `POST /tasks` - 创建任务
- `GET /tasks/:id` - 获取任务详情
- `DELETE /tasks/:id` - 删除任务
- `POST /tasks/:id/execute` - 执行任务
- `GET /tasks/:id/status` - 获取任务状态

### 文件下载
- `GET /download/:filename` - 下载生成的文件

## 配置说明

### 环境变量
- `PORT`: 服务端口（默认: 8080）
- `DB_PATH`: SQLite数据库路径（默认: ./data.db）
- `UPLOAD_DIR`: 文件上传目录（默认: ./uploads）

### 数据库配置
项目默认使用SQLite作为元数据存储，支持配置MySQL或PostgreSQL作为元数据库。

## 开发指南

### 添加新的数据生成规则
1. 在 `backend/services/generator.go` 中添加新的生成逻辑
2. 在前端 `Task.vue` 中添加对应的配置选项
3. 更新相关的类型定义和验证规则

### 添加新的数据库支持
1. 在 `backend/services/database.go` 中添加新的数据库驱动
2. 实现对应的连接和查询逻辑
3. 更新前端的数据库类型选项

## 常见问题

### Q: 如何处理大量数据生成？
A: 系统采用分批处理机制，可以通过调整批次大小来优化性能。建议单次生成不超过100万条记录。

### Q: 支持哪些数据类型？
A: 支持常见的数据类型包括字符串、数字、日期、布尔值等，具体支持情况取决于目标数据库。

### Q: 如何自定义数据生成规则？
A: 可以使用自定义规则类型，支持JavaScript表达式来定义复杂的生成逻辑。

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

如有问题或建议，请通过以下方式联系：
- 提交 Issue
- 发送邮件至项目维护者

---

**注意**: 本项目仅用于测试数据生成，请勿在生产环境中使用生成的数据进行实际业务操作。