# 内部技术博客系统

一个基于 Vue 3 + Go Gin 的内部博客系统，支持团队成员分享技术文章、踩坑记录等。

## 技术栈

### 前端
- Vue 3 (Composition API)
- Vite
- Vue Router
- Axios
- Marked (Markdown 渲染)
- jsPDF + html2canvas (PDF 导出)

### 后端
- Go 1.21+
- Gin 框架
- sync.Map (内存存储)
- goldmark (Markdown 转 HTML)
- gorilla/feeds (RSS 生成)

## 功能特性

### 文章管理
- ✅ 文章列表（分页、分类、标签、作者、搜索筛选）
- ✅ 文章详情页（目录导航、阅读进度条）
- ✅ Markdown 编辑器（实时预览、字数统计、阅读时间估算）
- ✅ 文章创建、编辑、删除
- ✅ 置顶功能（按置顶时间倒序）
- ✅ 草稿自动保存（localStorage）
- ✅ PDF 导出

### 评论系统
- ✅ 发表评论
- ✅ 楼中楼回复（一层嵌套）
- ✅ 评论点赞（cookie 去重）
- ✅ 评论举报

### 分类管理
- ✅ 分类 CRUD
- ✅ 删除分类前检查文章关联
- ✅ 写文章时可临时新建分类

### 标签与搜索
- ✅ 热门标签云（字体大小区分热度）
- ✅ 按标签筛选文章
- ✅ 全文搜索（标题、标签、内容）
- ✅ 搜索结果关键词高亮

### 附加功能
- ✅ 暗黑模式切换
- ✅ 近期活跃作者榜单
- ✅ 按月归档
- ✅ 文章相关推荐（基于标签）
- ✅ RSS 订阅
- ✅ 文章点赞（cookie 去重）
- ✅ 访问量统计（atomic.Int64）

## 项目结构

```
5202-vue-go-blog/
├── backend/                 # Go 后端
│   ├── main.go             # 入口文件
│   ├── go.mod              # Go 依赖
│   ├── config/             # 配置
│   ├── models/             # 数据模型
│   └── handlers/           # API 处理器
└── frontend/               # Vue 前端
    ├── package.json        # npm 依赖
    ├── vite.config.js      # Vite 配置
    ├── index.html
    └── src/
        ├── main.js         # 入口
        ├── App.vue         # 根组件
        ├── style.css       # 全局样式
        ├── api/            # API 封装
        ├── router/         # 路由配置
        └── views/          # 页面组件
```

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+

### 启动后端

```bash
cd backend
go mod tidy
go run main.go
```

后端服务将在 `http://localhost:8080` 启动

### 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端服务将在 `http://localhost:5173` 启动

## API 文档

### 文章接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/articles | 获取文章列表 |
| POST | /api/articles | 创建文章 |
| GET | /api/articles/:id | 获取文章详情 |
| PUT | /api/articles/:id | 更新文章 |
| DELETE | /api/articles/:id | 删除文章 |
| POST | /api/articles/:id/comments | 发表评论 |
| POST | /api/articles/:id/like | 点赞文章 |

**列表参数:**
- `page`: 页码（默认 1）
- `size`: 每页数量（默认 10）
- `tag`: 标签筛选
- `category`: 分类 ID 筛选
- `author`: 作者筛选
- `search`: 搜索关键词
- `sortBy`: 排序方式（time/views）

### 分类接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/categories | 获取分类列表 |
| POST | /api/categories | 创建分类 |
| PATCH | /api/categories/:id | 更新分类 |
| DELETE | /api/categories/:id | 删除分类 |

### 其他接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/tags | 获取热门标签云 |
| GET | /api/authors/active | 获取活跃作者 |
| GET | /api/archive | 获取归档列表 |
| GET | /api/reports | 获取举报列表 |
| POST | /api/comments/:id/like | 点赞评论 |
| POST | /api/comments/:id/report | 举报评论 |
| GET | /rss.xml | RSS 订阅 |

## 数据说明

- 所有数据存储在内存中（sync.Map）
- 重启服务会丢失数据，建议每周手动备份
- client_id 存在 cookie 中用于点赞去重
- 草稿存在 localStorage 中（key: `draft_{clientId}_{articleId}`）

## 置顶规则

置顶文章按以下规则排序：
1. 按置顶时间倒序排在最前
2. 同置顶时间内的按发布时间倒序
