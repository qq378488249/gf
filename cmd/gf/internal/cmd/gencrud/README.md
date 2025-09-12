# GoFrame CRUD 生成器

GoFrame CLI 工具的 CRUD 生成器，用于快速生成完整的四层架构 CRUD 代码。

## 📋 目录

- [功能介绍](#功能介绍)
- [快速开始](#快速开始)
- [命令参数](#命令参数)
- [生成的文件结构](#生成的文件结构)
- [配置文件](#配置文件)
- [实现流程](#实现流程)
- [代码示例](#代码示例)
- [最佳实践](#最佳实践)

## 功能介绍

CRUD 生成器可以根据数据库表自动生成：

- **API 层**: 请求/响应结构体定义和接口
- **Controller 层**: HTTP 处理逻辑（分文件存储）
- **Service 层**: 业务接口定义
- **Logic 层**: 具体业务逻辑实现

### 🎯 核心特性

- ✅ 四层架构完整生成
- ✅ 结构体名称复用，避免重复定义
- ✅ 自动复用 DAO 生成的 entity 结构体
- ✅ 支持数据库配置复用
- ✅ 驼峰格式 JSON 字段名
- ✅ 按功能分文件组织
- ✅ 标准化代码模板

## 快速开始

### 1. 基本用法

```bash
# 生成单表 CRUD
gf gen crud -t user

# 生成多表 CRUD
gf gen crud -t user,product,order

# 指定数据库连接
gf gen crud -t user -l "mysql:root:password@tcp(127.0.0.1:3306)/database"
```

### 2. 前置条件

在使用 CRUD 生成器之前，通常需要：

```bash
# 1. 先生成 DAO 层（推荐）
gf gen dao

# 2. 然后生成 CRUD
gf gen crud -t user
```

## 命令参数

| 参数 | 简写 | 描述 | 默认值 | 必填 |
|------|------|------|--------|------|
| `--tables` | `-t` | 表名，多个用逗号分隔 | - | ✅ |
| `--link` | `-l` | 数据库连接字符串 | - | ❌ |
| `--group` | `-g` | 数据库配置组名 | `default` | ❌ |
| `--path` | `-p` | 生成文件根目录 | `./` | ❌ |
| `--ctrlPath` | `-c` | 控制器目录 | `internal/controller` | ❌ |
| `--servicePath` | `-s` | 服务接口目录 | `internal/service` | ❌ |
| `--logicPath` | `-o` | 业务逻辑目录 | `internal/logic` | ❌ |
| `--apiPath` | `-a` | API 定义目录 | `api` | ❌ |
| `--packageName` | `-n` | 包名 | 自动检测 | ❌ |
| `--overwrite` | `-w` | 覆盖现有文件 | `false` | ❌ |
| `--removePrefix` | `-r` | 移除表名前缀 | - | ❌ |

### 使用示例

```bash
# 基本生成
gf gen crud -t user

# 使用数据库连接字符串
gf gen crud -t user -l "mysql:root:123456@tcp(127.0.0.1:3306)/mydb"

# 自定义路径
gf gen crud -t user -c controllers -s services -o business -a api

# 移除表前缀
gf gen crud -t sys_user,sys_role -r sys_

# 强制覆盖
gf gen crud -t user -w
```

# 代码生成后，还需要注册到cmd和logic里（仅一次）
```go
// cmd.go
group.Bind(
    user.NewV1()
)

// logic.go
import (
	_ "goframe-test/internal/logic/user"
)
```

## 生成的文件结构

```
project/
├── api/                           # API 定义层
│   └── user/
│       ├── user.go               # IUserV1 接口定义
│       └── v1/
│           ├── create.go         # CreateReq/CreateRes
│           ├── delete.go         # DeleteReq/DeleteRes
│           ├── update.go         # UpdateReq/UpdateRes
│           ├── get_one.go        # GetOneReq/GetOneRes
│           └── get_list.go       # GetListReq/GetListRes
├── internal/
│   ├── controller/               # 控制器层
│   │   └── user/
│   │       ├── user_new.go       # 构造函数
│   │       ├── user_v1_create.go # 创建操作
│   │       ├── user_v1_delete.go # 删除操作
│   │       ├── user_v1_update.go # 更新操作
│   │       ├── user_v1_get_one.go# 获取单条
│   │       └── user_v1_get_list.go# 获取列表
│   ├── service/                  # 服务接口层
│   │   └── user.go              # IUser 接口定义和注册
│   └── logic/                    # 业务逻辑层
│       └── user/
│           └── user.go          # 业务逻辑实现
```

## 配置文件

### config.yaml

```yaml
# 数据库配置
database:
  default:
    link: "mysql:root:password@tcp(127.0.0.1:3306)/database"
  
  # 多数据库支持
  user_db:
    link: "mysql:root:password@tcp(127.0.0.1:3306)/user_db"

# CRUD 生成配置，复用dao配置
gfcli:
  gen:
    # CRUD 配置
    dao:
      # 使用数据库配置组
      - group: "default"
        tables: "user,product,order"
        path: "./"
        removePrefix: "sys_"
        overwrite: true
      
      # 使用直接连接
      - link: "mysql:root:password@tcp(127.0.0.1:3306)/another_db"
        tables: "admin_user,admin_role"
        path: "./admin"
        ctrlPath: "internal/controller"
        servicePath: "internal/service"
        logicPath: "internal/logic"
        apiPath: "api"
```

使用配置文件：

```bash
# 使用配置文件中的设置
gf gen crud
```

## 实现流程

### 1. 数据库表分析

```go
// 获取表信息
tableInfo := getTableInfo(db, tableName)
// 包含：表名、结构体名、字段信息、主键等
```

### 2. 四层文件生成

#### API 层生成
- 位置：`api/{table}/v1/{action}.go`
- 内容：请求/响应结构体
- 特点：复用 entity 结构体

#### Controller 层生成
- 位置：`internal/controller/{table}/{table}_v1_{action}.go`
- 内容：HTTP 处理方法
- 特点：按功能分文件

#### Service 层生成
- 位置：`internal/service/{table}.go`
- 内容：接口定义和注册机制
- 特点：使用依赖注入模式

#### Logic 层生成
- 位置：`internal/logic/{table}/{table}.go`
- 内容：具体业务逻辑实现
- 特点：实现 Service 接口

### 3. 类型系统设计

```
API Request/Response ←→ Controller ←→ Service Interface ←→ Logic Implementation
         ↑                                                       ↓
         └─────────────── 直接复用 ──────────────────────────────┘
```

## 代码示例

### API 层示例

```go
// api/user/v1/create.go
package v1

import (
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gtime"
)

type (
    CreateReq struct {
        g.Meta `path:"/user" tags:"User" method:"post" summary:"Create user"`
        Name   string     `json:"name" v:"required" dc:"User name"`
        Email  string     `json:"email" v:"required" dc:"User email"`
        Age    int        `json:"age" dc:"User age"`
    }
    CreateRes struct {
        g.Meta `mime:"application/json" example:"string"`
        Id     uint64 `json:"id" dc:"Created user ID"`
    }
)
```

### Controller 层示例

```go
// internal/controller/user/user_v1_create.go
package user

import (
    "context"
    "myproject/api/user/v1"
    "myproject/internal/service"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
    out, err := service.User().Create(ctx, req)
    if err != nil {
        return nil, err
    }
    return out, nil
}
```

### Service 层示例

```go
// internal/service/user.go
package service

import (
    "context"
    v1 "myproject/api/user/v1"
)

type (
    IUser interface {
        Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error)
        Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error)
        Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error)
        GetOne(ctx context.Context, req *v1.GetOneReq) (*v1.GetOneRes, error)
        GetList(ctx context.Context, req *v1.GetListReq) (*v1.GetListRes, error)
    }
)

var localUser IUser

func User() IUser {
    if localUser == nil {
        panic("implement not found for interface IUser, forgot register?")
    }
    return localUser
}

func RegisterUser(i IUser) {
    localUser = i
}
```

### Logic 层示例

```go
// internal/logic/user/user.go
package user

import (
    "context"
    "myproject/api/user/v1"
    "myproject/internal/dao"
    "myproject/internal/model/entity"
    "myproject/internal/service"
)

type sUser struct{}

func init() {
    service.RegisterUser(New())
}

func New() service.IUser {
    return &sUser{}
}

func (s *sUser) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
    // 业务逻辑处理
    result, err := dao.User.Ctx(ctx).Data(req).Insert()
    if err != nil {
        return nil, err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return nil, err
    }
    
    return &v1.CreateRes{Id: uint64(id)}, nil
}
```

## 最佳实践

### 1. 开发流程

```bash
# 推荐的开发流程
1. 设计数据库表结构
2. gf gen dao              # 生成 DAO 层
3. gf gen crud -t table    # 生成 CRUD 层
4. 完善业务逻辑           # 在 Logic 层添加业务代码
5. 添加路由绑定           # 在 router 中绑定控制器
6. 测试和调试             # API 测试
```

### 2. 命名规范

- **表名**: 使用下划线命名（如：`user_profile`）
- **结构体**: 使用大驼峰命名（如：`UserProfile`）
- **JSON字段**: 使用小驼峰命名（如：`firstName`）
- **接口**: 使用 `I` 前缀（如：`IUser`）

### 3. 目录组织

```
internal/
├── controller/     # 控制器按表分目录
├── service/        # 服务接口集中管理  
├── logic/          # 业务逻辑按表分目录
├── dao/            # DAO 层（gf gen dao 生成）
└── model/          # 模型层（gf gen dao 生成）
```

### 4. 配置管理

- 推荐使用配置文件方式
- 复用 DAO 命令的数据库配置
- 支持多数据库多环境配置

### 5. 代码维护

- Logic 层可以自由修改和扩展
- API 和 Controller 层谨慎修改
- Service 接口保持稳定
- 使用版本化 API（v1, v2...）

### 6. 错误处理

```go
// 在 Logic 层添加业务验证
func (s *sUser) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
    // 业务验证
    if req.Age < 0 {
        return nil, gerror.New("age cannot be negative")
    }
    
    // 数据库操作
    result, err := dao.User.Ctx(ctx).Data(req).Insert()
    if err != nil {
        return nil, err
    }
    
    // 返回结果
    id, _ := result.LastInsertId()
    return &v1.CreateRes{Id: uint64(id)}, nil
}
```

## 🎉 总结

GoFrame CRUD 生成器提供了：

- **高效开发**: 一键生成完整的四层架构
- **标准规范**: 遵循 GoFrame 最佳实践
- **类型安全**: 编译时类型检查
- **易于维护**: 清晰的代码结构和分层设计
- **灵活配置**: 支持多种配置方式和自定义选项

通过 CRUD 生成器，开发者可以专注于业务逻辑的实现，而无需关心基础架构代码的编写。🚀