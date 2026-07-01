我来帮你实现远程问诊会话模块。首先让我了解一下项目结构，然后创建详细的任务计划。
        
            
toolName: view_folder
            
status: success
          
c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go
          

        
            
toolName: view_folder
            
status: success
          
c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\go.mod
          
让我参考一下现有模块的代码风格：

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\types.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\models.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          
好的，我已经了解了项目结构和代码风格。现在让我创建一个详细的任务计划：


        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在开始实现。首先创建 types.go 定义核心数据结构：


        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: Write
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\types.go
          
changes: undefined
          
现在创建 service.go 实现核心业务逻辑：




        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: Write
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service.go
          
changes: undefined
          
现在创建完整的单元测试：

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: Write
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          
changes: undefined
          
现在让我们运行测试来验证代码是否正确：


        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/teleconsult/ -v undefined
          
有一个编译错误，让我修复：

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          
changes: undefined
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/teleconsult/ -v undefined
          
有三个测试失败了，让我逐个修复：



        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service.go
          
我需要修复三个问题。首先调整 AcceptSession 的验证顺序，然后修复测试：





        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service.go
          
changes: undefined
          
现在修复另外两个测试问题：



        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          
changes: undefined
          
现在修复第三个测试：




        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          
changes: undefined
          
再次运行测试：





        
            
toolName: run_command
            
status: success
          
command: go test ./internal/teleconsult/ -v undefined
          
太好了！所有测试都通过了。现在编写需求文档：





        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: Write
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\docs\teleconsult.md
          
changes: undefined
          
最后，让我更新 todo 列表并做最终的诊断检查：







        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
让我做最后的语法和编译检查：




        
            
toolName: GetDiagnostics
            
status: success
          
          

        
            
toolName: GetDiagnostics
            
status: success
          
          

        
            
toolName: GetDiagnostics
            
status: success
          
          
所有任务完成！以下是实现总结：

## 📦 模块文件结构

```
internal/teleconsult/
├── types.go          # 核心数据结构定义
├── service.go        # 业务逻辑实现
└── service_test.go   # 完整单元测试 (50+ 用例)

docs/
└── teleconsult.md    # 模块需求文档
```

## ✅ 完成的功能

### 1. 图文咨询
- 患者可通过 `CreateSession` 发起问诊，附带初始文字消息
- 通过 `SendMessage` 支持 `TEXT` 和 `IMAGE` 两种消息类型
- 消息按发送时间顺序追加到 `Messages` 列表，保证顺序

### 2. 医生接诊
- 待接诊状态（PENDING）下调用 `AcceptSession` 转为进行中（ONGOING）
- **关键约束**：未接诊时医生发送 `IsMedicalAdvice=true` 的消息返回 `ErrSessionNotAccepted`，确保不能提前给出诊疗意见

### 3. 会话超时关闭
- **未接诊超时**：`DefaultAcceptTimeout = 30分钟`，PENDING 超时自动关闭
- **无互动超时**：`DefaultInactivityTimeout = 60分钟`，ONGOING 超过无消息自动关闭
- **触发点**：`SendMessage` 自动检查 + `CheckSessionTimeout` / `CheckAllSessionsTimeout` 主动调用
- 关闭后普通 `SendMessage` 返回 `ErrSessionClosed`

### 4. 归档与追问限制
- `CompleteSession` 生成 `ArchiveRecord`，包含诊断、治疗建议、完整消息历史
- `SendFollowUp` 独立接口，只能由患者在 `FollowUpLimit`（默认 3 次）内调用
- 超限返回 `ErrFollowUpLimitExceeded`

## 📊 测试覆盖

- **50+ 单元测试用例**，运行 `go test ./internal/teleconsult/ -v` **全部通过**
- 覆盖：正常流程 / 边界条件 / 所有 17 个错误分支 / 并发访问

## 🔗 关键代码位置

| 内容 | 文件 |
|------|------|
| 数据结构与错误定义 | [types.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/teleconsult/types.go) |
| 核心业务逻辑 | [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/teleconsult/service.go) |
| 完整单元测试 | [service_test.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/teleconsult/service_test.go) |
| 需求文档 | [teleconsult.md](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/docs/teleconsult.md) |