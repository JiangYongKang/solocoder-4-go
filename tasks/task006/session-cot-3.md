我来修复这些问题。首先创建任务计划：




        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
先读取相关代码：









        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\models.go
          
现在开始修复。首先修复 `isDateWithinDays` 函数，增加下限排除：











        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\models.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在修复 `GetExpiringBatches`，排除已过期状态的批次：

















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在删除 OutboundFIFO 扣减阶段的冗余过期检查：


















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在补充测试用例。先更新 `TestIsDateWithinDays`，然后添加新的测试：

















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          
现在更新 `TestIsDateWithinDays` 测试，因为函数逻辑变了：

















        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          
changes: undefined
          
现在在测试文件末尾添加新的测试用例：



















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\drugtrace\drugtrace_test.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在运行测试：



















        
            
toolName: run_command
            
status: success
          
command: go test ./internal/drugtrace/ -v undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
## 所有问题已修复完成！✅

### 修复内容总结：

### 1. **临期判定函数增加下限排除** - [models.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/models.go#L114-L120)

修复了 `isDateWithinDays` 函数，增加下限检查：

```go
// 修复前（有缺陷）：只检查上限
return expiryDateOnly.Before(cutoff) || expiryDateOnly.Equal(cutoff)

// 修复后：同时检查上下限
return (expiryDateOnly.After(asOfOnly) || expiryDateOnly.Equal(asOfOnly)) &&
       (expiryDateOnly.Before(cutoff) || expiryDateOnly.Equal(cutoff)
```

**业务规则**：临期批次必须满足 `当前日期 ≤ 到期日期 ≤ 当前日期 + days，早已过期的批次不会被误判为临期。

---

### 2. **GetExpiringBatches 排除已过期批次** - [drugtrace.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace.go#L121-L129)

在遍历时同时排除已过期和已召回状态的批次：

```go
// 修复前
if batch.Status == BatchStatusRecalled {
    continue
}

// 修复后
if batch.Status == BatchStatusRecalled || batch.Status == BatchStatusExpired {
    continue
}
```

**效果**：同一个批次不会同时出现在已过期列表和临期列表中。

---

### 3. **删除 OutboundFIFO 冗余过期检查** - [drugtrace.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace.go#L231-L234)

删除了扣减阶段的冗余过期检查。因为在筛选阶段已经将所有过期批次排除出 `availableBatches` 集合，扣减阶段的检查在任何情况下都不会被触发，属于死代码。

---

### 4. **补充测试用例** - [drugtrace_test.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/drugtrace/drugtrace_test.go)

新增 4 个测试函数：

- `TestGetExpiringBatchesExcludesExpiredStatus`: 验证临期查询排除已过期状态批次
- `TestGetExpiringBatchesExcludesLongExpiredNotUpdated`: 验证早已过期但状态未更新的批次不被包含
- `TestBatchNotInBothExpiredAndExpiringLists`: 验证批次不会同时出现在两个列表中
- `TestGetExpiringBatchesExcludesRecalled`: 验证召回批次不被包含

同时更新了 `TestIsDateWithinDays` 测试，补充已过期日期的下限排除测试用例。

---

### 测试结果：

```
PASS
ok      solocoder-4-go/internal/drugtrace       0.253s
```

**32 个测试函数，70+ 个子测试全部通过！** 🎉