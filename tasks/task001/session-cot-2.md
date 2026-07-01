我先分析问题，然后制定修复计划。
        
            
toolName: todo_write
            
status: success
          
Todos updated: 6 items
          
让我先读取相关代码确认问题点。


        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
让我看看测试用例中错误的部分。





        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service_test.go
          
问题确认完毕，现在开始修复。首先重写 `CreateAppointment` 保证原子性。







        
            
toolName: todo_write
            
status: success
          
Todos updated: 6 items
          
让我先读取 LockSlot 的完整实现，以便整合到 CreateAppointment 中。







        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
现在重写 `CreateAppointment` 方法，将锁定和确认整合在一个锁内，保证原子性：









        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修复 `ChangeAppointment` 中的同号源检查逻辑，需要调整检查的位置（必须在获取旧预约之后）。









        
            
toolName: todo_write
            
status: success
          
Todos updated: 6 items
          
让我先读取 ChangeAppointment 的完整代码。











        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
现在修复 ChangeAppointment，将检查移到获取旧预约之后，并比较号源ID：










        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修复测试用例 `TestChangeAppointment_SameSlot`，传入正确的号源ID：

















        
            
toolName: todo_write
            
status: success
          
Todos updated: 6 items
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service_test.go
          
changes: undefined
          
现在补充并发超售测试。先看看测试文件的末尾，找到合适的插入位置：



















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service_test.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service_test.go
          
现在添加并发超售测试，放在 `TestFullWorkflow` 之后：























        
            
toolName: todo_write
            
status: success
          
Todos updated: 6 items
          
现在添加并发超售测试：













        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service_test.go
          
changes: undefined
          
让我检查测试文件的 import 部分，确保我们使用的包都已导入：

























        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service_test.go
          
需要添加必要的 import：



























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service_test.go
          
changes: undefined
          
现在运行测试验证所有修复：































        
            
toolName: todo_write
            
status: success
          
Todos updated: 6 items
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/appointment/ -v -count=1 undefined
          
所有测试都通过了！让我更新 todo。



































        
            
toolName: todo_write
            
status: success
          
Todos updated: 6 items
          
---

## 修复总结

### ✅ 问题1：CreateAppointment 原子性问题

**问题描述**：原实现先调用 `LockSlot` 再调用 `ConfirmAppointment`，两次独立加解锁之间存在并发间隙，可能导致号源超售。

**修复方案**：在 [CreateAppointment](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/appointment/service.go#L396-L463) 方法中将锁定、检查、确认逻辑整合在**同一个互斥锁**内，确保整个预约过程的原子性。

```go
// 修复后：单个锁内完成所有操作
svc.store.mu.Lock()
defer svc.store.mu.Unlock()

// 1. 锁定检查
// 2. 容量检查
// 3. 创建预约记录
// 4. 更新号源状态
// 5. 清除锁定（如有）
```

---

### ✅ 问题2：ChangeAppointment 同号源检查逻辑错误

**问题描述**：原代码错误地比较预约ID和号源ID（`req.AppointmentID == req.NewSlotID`），两者属于不同的ID体系（前缀分别是 `APT` 和 `SLT`），永远不会相等。

**修复方案**：在获取旧预约记录后，比较**旧号源ID**和**新号源ID**：
```go
// 修复前（错误）
if req.AppointmentID == req.NewSlotID { ... }

// 修复后（正确）
if oldAppt.SlotID == req.NewSlotID { ... }
```
修复位置：[ChangeAppointment](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/appointment/service.go#L487-L489)

---

### ✅ 问题3：TestChangeAppointment_SameSlot 测试用例错误

**问题描述**：测试用例错误地向 `NewSlotID` 传入了 `appt.ID`（预约ID），而不是真正的号源ID。

**修复方案**：[TestChangeAppointment_SameSlot](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/appointment/service_test.go#L929-L933) 中传入正确的号源ID `ctx.Slot1ID`。

---

### ✅ 问题4：新增并发超售测试

新增两个并发测试用例，验证在高并发场景下不会出现号源超售：

1. **TestConcurrentAppointment_NoOversell** - 单号源并发预约测试
   - 号源容量：3
   - 并发 goroutine：15 个（容量 × 5）
   - 验证点：
     - `BookedCount == 3`（不超售，不浪费）
     - 恰好 3 个成功，12 个失败
     - 总数一致

2. **TestConcurrentAppointment_MultipleSlots** - 多号源并发预约测试
   - 3 个号源，每个容量 3，总容量 9
   - 并发 goroutine：27 个
   - 验证点：每个号源的 `BookedCount` 都不超过容量

---

### 🧪 测试结果

```
go test ./internal/appointment/ -v -count=1

PASS: TestAddSchedule_Success
...
PASS: TestChangeAppointment_SameSlot
...
PASS: TestFullWorkflow
PASS: TestConcurrentAppointment_NoOversell
PASS: TestConcurrentAppointment_MultipleSlots

PASS
ok      solocoder-4-go/internal/appointment     0.266s
```

**71/71 测试用例全部通过**（原有 67 个 + 新增 2 个并发测试 + 2 个修复后原有测试）。