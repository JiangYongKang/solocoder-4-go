我需要修复两个问题：1）出勤率计算重复计数；2）判定基准偏移。让我先分析问题，然后制定修复方案。









        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          
让我先查看当前的 calculateStatus 方法和汇总逻辑：









        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service.go
          
现在我理解了问题。让我先修复出勤率计算，避免重复计数：











        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service.go
          
changes: undefined
          
现在修复 GetStudentSummary 方法中的出勤率计算：













        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service.go
          
changes: undefined
          
现在更新 todo 列表，然后修复 calculateStatus 方法的判定基准问题：



















        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          
现在修复 calculateStatus 方法，恢复原有参照时间点，仅对学生签到签退时间进行精度修正：























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service.go
          
changes: undefined
          
现在更新 todo 列表，然后修改测试用例：
























        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          
现在添加测试用例，验证出勤率不超过100%：


























        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service_test.go
          
现在添加测试用例验证出勤率不超过100%：































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\attendance\service_test.go
          
changes: undefined
          
现在更新 todo 列表并运行测试：





















        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          
现在运行测试确保所有测试通过：























        
            
toolName: run_command
            
status: success
          
command: go test ./internal/attendance/ -v -count=1 undefined
          
所有测试都通过了！现在更新 todo 列表：

























        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          
所有修复已完成，52个测试全部通过！以下是修复总结：

## 已修复的问题

### 1. 出勤率计算重复计数问题

**问题**：复合状态（迟到且早退）的记录会同时计入 LateCount 和 EarlyLeaveCount，而出勤率分子是简单相加，导致重复计数，出勤率可能超过100%。

**解决方案**：
- 引入独立的 `attendedCount` 计数器，每条记录只算一次出勤
- 根据 Status 判断是否出勤：PRESENT、LATE、EARLY_LEAVE、LATE_AND_EARLY_LEAVE、LEAVE 各计1次
- LateCount 和 EarlyLeaveCount 仅用于异常统计，不参与出勤率计算

**代码位置**：
- [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/attendance/service.go#L550-L584) - 班级汇总
- [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/attendance/service.go#L626-L658) - 学生汇总

### 2. 判定基准偏移问题

**问题**：`calculateStatus` 方法中对所有时间（包括参照时间点）都进行了 `truncateToMinute`，改变了原有的判定基准。

**解决方案**：
- 参照时间点（ruleCheckInStart、ruleCheckInEnd、ruleCheckOutStart、ruleCheckOutEnd、lateThreshold、earlyThreshold）保持原始计算方式，不进行截断
- 仅对学生的签到签退时间进行 `truncateToMinute`，确保相同时钟时刻判定一致
- 时间比较逻辑保持 `!After` 和 `!Before` 的稳定写法

**代码位置**：
- [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/attendance/service.go#L396-L431) - calculateStatus 方法

### 3. 新增出勤率约束测试

**新增3个测试用例**：
- `TestAttendanceRate_NotExceed100Percent` - 多学生混合状态下出勤率不超过100%
- `TestAttendanceRate_SingleStudentMultipleSessions` - 单学生多场次出勤率验证
- `TestAttendanceRate_AllStudentsLateAndEarlyLeave` - 全部学生都是复合状态时出勤率恰好为100%

**测试结果**：
```
PASS
ok      solocoder-4-go/internal/attendance      0.222s
```
共 **52个测试用例** 全部通过。

运行命令：`go test ./internal/attendance/ -v`