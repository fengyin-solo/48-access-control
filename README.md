# 门禁管理系统（accesscontrol）

纯 Go 标准库实现的门禁管理系统后端服务，零第三方依赖（仅 `net/http` + 标准库）。

## 运行

```bash
# 启动（默认监听 :8080）
go run ./cmd/server

# 环境变量
# PORT=8080         监听端口
# ADDR=:9090        完整监听地址（优先于 PORT）
# API_KEY=xxx       鉴权密钥（默认 access-control-secret）
# RATE_LIMIT=100    每 IP 每秒请求数上限
# MAX_PAGE_SIZE=100 分页最大条数
# LOG_LEVEL=info    debug/info/warn/error
```

启动后浏览器访问 `http://localhost:8080/` 查看看板页面。

## 鉴权

所有 `/api/*` 接口需携带请求头 `X-API-Key: access-control-secret`；`/healthz` 与静态页面无需鉴权。

## API 一览

统一响应格式：`{"code":0,"message":"ok","data":...}`；错误码：400 参数校验、401 鉴权失败、404 不存在、409 冲突、429 限流、500 内部错误。

### 区域 Zone

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/zones` | 创建区域 |
| GET | `/api/zones` | 列表（`name`/`status`/`level`/`parent_id`/`page`/`size`） |
| GET | `/api/zones/{id}` | 详情 |
| PUT | `/api/zones/{id}` | 更新 |
| DELETE | `/api/zones/{id}` | 删除 |

### 门禁点 AccessPoint

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/access-points` | 创建门禁点 |
| GET | `/api/access-points` | 列表（`zone_id`/`type`/`status`/`keyword`） |
| GET | `/api/access-points/{id}` | 详情 |
| PUT | `/api/access-points/{id}` | 更新 |
| DELETE | `/api/access-points/{id}` | 删除 |
| POST | `/api/access-points/{id}/transition` | 状态流转（`{"status":"offline"}`） |

### 读卡器 Reader

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/readers` | 创建读卡器 |
| GET | `/api/readers` | 列表（`access_point_id`/`status`/`keyword`） |
| GET | `/api/readers/{id}` | 详情 |
| PUT | `/api/readers/{id}` | 更新 |
| DELETE | `/api/readers/{id}` | 删除 |

### 人员 Person

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/persons` | 创建人员 |
| GET | `/api/persons` | 列表（`department`/`status`/`keyword`） |
| GET | `/api/persons/{id}` | 详情 |
| PUT | `/api/persons/{id}` | 更新 |
| DELETE | `/api/persons/{id}` | 删除 |

### 凭证 Credential

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/credentials` | 发放凭证 |
| GET | `/api/credentials` | 列表（`person_id`/`type`/`status`/`keyword`） |
| GET | `/api/credentials/{id}` | 详情 |
| PUT | `/api/credentials/{id}` | 更新 |
| DELETE | `/api/credentials/{id}` | 删除 |
| POST | `/api/credentials/{id}/transition` | 状态流转 |
| POST | `/api/credentials/{id}/check` | 有效性校验（过期自动置 expired） |
| POST | `/api/credentials/batch-issue` | 批量发放 |
| POST | `/api/credentials/batch-disable` | 批量停用 |

### 通行记录 AccessLog

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/access-logs` | 手工补录 |
| GET | `/api/access-logs` | 列表（`credential_id`/`access_point_id`/`result`/`direction`/`from`/`to`） |
| GET | `/api/access-logs/{id}` | 详情 |
| DELETE | `/api/access-logs/{id}` | 删除 |
| POST | `/api/access-logs/batch-delete` | 批量删除（`{"ids":[...]}`） |

### 时段规则 TimeRule

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/time-rules` | 创建规则 |
| GET | `/api/time-rules` | 列表（`access_point_id`/`effect`/`status`/`keyword`） |
| GET | `/api/time-rules/{id}` | 详情 |
| PUT | `/api/time-rules/{id}` | 更新 |
| DELETE | `/api/time-rules/{id}` | 删除 |

### 排班 Schedule

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/schedules` | 创建排班 |
| GET | `/api/schedules` | 列表（`person_id`/`zone_id`/`status`） |
| GET | `/api/schedules/{id}` | 详情 |
| PUT | `/api/schedules/{id}` | 更新 |
| DELETE | `/api/schedules/{id}` | 删除 |
| POST | `/api/schedules/{id}/transition` | 状态流转 |

### 告警 Alert

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/alerts` | 创建告警 |
| GET | `/api/alerts` | 列表（`type`/`level`/`status`/`handler`/`keyword`） |
| GET | `/api/alerts/{id}` | 详情 |
| PUT | `/api/alerts/{id}` | 更新 |
| DELETE | `/api/alerts/{id}` | 删除 |
| POST | `/api/alerts/{id}/transition` | 状态流转（`{"status":"acknowledged","handler":"张三"}`） |
| POST | `/api/alerts/batch-update` | 批量更新状态 |

### 通行鉴权 / 统计 / 导出

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/access` | 通行授权判定（综合凭证状态/有效期/时段规则） |
| GET | `/api/stats/overview` | 看板汇总 |
| GET | `/api/stats/zone-traffic` | 各区域通行量 |
| GET | `/api/stats/point-traffic/top?n=5` | 门禁点通行 TOP N |
| GET | `/api/stats/credential-status` | 凭证状态分布 |
| GET | `/api/stats/alert-level` | 告警按等级统计 |
| GET | `/api/stats/alert-status` | 告警按状态统计 |
| GET | `/api/export` | 全量数据导出快照 |

## 状态机

- Credential：`active → lost/expired/disabled`
- Alert：`open → acknowledged → resolved`
- AccessPoint：`online → offline → disabled`；`offline → online`
- Schedule：`active → ended`

## 项目结构

```
cmd/server/main.go        入口
internal/config           配置加载
internal/app              依赖装配
internal/model            实体模型 + 校验 + 状态机
internal/store            内存存储（接口 + 实现）
internal/service          业务逻辑 + 统计 + 导出
internal/handler          HTTP 处理器 + 中间件
pkg/httpx                 统一响应/分页
pkg/idgen                 ID 生成
pkg/logger                分级日志
web/                      前端看板（原生 JS，零 CDN）
```
