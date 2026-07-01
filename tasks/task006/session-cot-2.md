我来修复这些问题。首先创建任务计划：



        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          
让我先读取完整的 drugtrace.go 文件，然后开始修复：





        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
现在开始修复。首先在 models.go 中添加日期精度比较的辅助函数：







        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\models.go
          
现在在 models.go 中添加日期精度比较的辅助函数：









        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\models.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          
现在修复 InboundBatch 中的过期判定：











        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          
现在修复 GetExpiringBatches、GetExpiredBatches 和 UpdateBatchStatus：










        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
changes: undefined
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
changes: undefined
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          
现在修复最重要的 OutboundFIFO 方法，统一时间基准并使用日期精度比较：













        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          
现在补充测试用例。先读取测试文件末尾，然后添加新的测试：















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          
现在在测试文件末尾添加新的测试用例：


















        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          
changes: undefined
          
我需要在测试文件中添加 `fmt` 导入，因为新测试使用了 `fmt.Sprintf`：

















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          
现在运行测试：

















        
            
toolName: run_command
            
status: success
          
command: go test ./internal/drugtrace/ -v undefined
          
有个未使用的变量，让我删除：



















        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          
changes: undefined
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/drugtrace/ -v undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 7 items
          
## 所有问题已修复完成！✅

### 修复内容总结：

### 1. **时间基准不一致问题** - [drugtrace.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace.go#L165-L280)

在 `OutboundFIFO` 方法开始时（加锁后）获取一次 `now`，并在整个出库流程中使用同一个时间基准，避免了循环中不同迭代获取不同时间导致的判定不一致问题。

### 2. **日期精度过期判定** - [models.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/models.go#L106-L119)

新增三个辅助函数：
- `dateOnly(t time.Time)`: 将时间截断到日期精度（年月日，时分秒清零）
- `isDateExpired(expiryDate, asOf time.Time)`: 按日期精度判断是否过期
  - **业务规则**: 仅当 `expiryDate` 的日期 **严格小于** 当前日期时才算过期
  - 当天到期的药品（有效期为今天 00:00:00 或 23:59:59）在当天仍可正常出库
- `isDateWithinDays(expiryDate, asOf, days)`: 按日期精度判断是否在 N 天内过期

### 3. **统一所有过期判定**

修复了以下方法中的过期判定，全部使用日期精度比较：
- [InboundBatch](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace.go#L63-L67)
- [GetExpiringBatches](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace.go#L119-L129)
- [GetExpiredBatches](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace.go#L142-L149)
- [UpdateBatchStatus](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace.go#L394-L400)

### 4. **新增测试用例** - [drugtrace_test.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace_test.go)

新增 10 个测试函数，覆盖：
- 日期精度辅助函数测试
- 有效期等于当天的批次可正常出库
- 跨日期边界的过期判定（昨天到期 vs 今天到期 vs 明天到期）
- 时间基准一致性测试（多批次出库时所有明细使用同一时间戳）

### 测试结果：

```
PASS
ok      solocoder-4-go/internal/drugtrace       0.278s
```

**28 个测试函数，60+ 个子测试全部通过！** 🎉