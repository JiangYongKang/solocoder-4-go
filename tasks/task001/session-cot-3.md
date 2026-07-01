让我先分析问题，制定修复计划。

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
先读取相关代码确认问题。







        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\types.go
          
让我继续读取 Store 结构体和 LockSlot、ConfirmAppointment 的完整代码。










        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\types.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
现在开始修复。首先修复问题1 - 删除死代码：







        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在分析问题3（锁粒度细化）和问题2（抽取公共函数）。由于这两个修改相互关联，我先实现锁粒度细化，再抽取公共函数。

首先修改 `ScheduleSlot` 结构体，添加号源级别的锁：









        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
让我先修改 `ScheduleSlot` 结构体，添加号源级别的互斥锁：












        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\types.go
          
changes: undefined
          
现在开始修改所有方法的锁策略。先修改 `LockSlot`：
















        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `ConfirmAppointment`：


















        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `ReleaseLock`：

















        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `CheckAndReleaseExpiredLocks`：



















        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `CreateAppointment`：


























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `ChangeAppointment`，这个方法需要操作两个号源，需要按号源ID排序获取锁以避免死锁：





























        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
让我先看看 `tryProcessWaitQueueForSlot` 的实现：





























        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
现在修改 `ChangeAppointment`。这个方法涉及两个号源，需要按号源ID排序获取锁以避免死锁：







































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `CancelAppointment`：








































        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `RecordNoShow`：









































        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在让我先运行一下测试，看看锁粒度细化后的情况：








































        
            
toolName: run_command
            
status: success
          
command: go test ./internal/appointment/ -v -count=1 2>&1 | Select-Object -First 100 undefined
          
让我看完整的测试结果：










































        
            
toolName: run_command
            
status: success
          
command: go test ./internal/appointment/ -v -count=1 2>&1 | Select-Object -Skip 100 undefined
          
出现了死锁！问题是锁顺序违反了「先 Store 锁，后号源锁」的原则。我需要调整策略：在需要获取 Store 写锁之前，先释放号源锁，然后按正确顺序重新获取，并做乐观检查。

让我重新修改 `CreateAppointment`：

































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修改 `ConfirmAppointment`，同样的问题：


























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在我需要先解决问题2 - 抽取公共校验函数，这样可以减少重复代码，然后再修正其他方法的锁问题。

























        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
让我先抽取公共校验函数，放在 service.go 开头的内部函数部分：




























        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
现在添加公共校验函数，放在 `NewService` 之后：



























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在重构 `LockSlot` 使用公共校验函数：





























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在重构 `ConfirmAppointment` 和 `CreateAppointment` 使用公共校验函数：





























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在重构 `CreateAppointment` 使用公共校验函数：































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在我需要修复 `ChangeAppointment`、`CancelAppointment`、`RecordNoShow` 中的死锁问题。为了保证代码可靠性，我采用更简单的策略：这些需要修改 Store 数据的方法一开始就获取 Store 写锁，然后获取号源锁，保证锁顺序正确。

先修复 `ChangeAppointment`：
































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修复 `CancelAppointment`：







































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在修复 `RecordNoShow`：









































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\appointment\service.go
          
changes: undefined
          
现在运行测试验证修复：











































        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/appointment/ -v -count=1 -timeout 30s undefined
          
所有测试通过！让我更新 todo。










































        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
---

## 修复总结

### ✅ 问题1：删除死代码

**问题**：[CreateAppointment](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/appointment/service.go#L446-L516) 中存在永远无法触发的锁清理代码。

**修复**：删除了死代码 `if slot.Status == SlotStatusLocked { svc.clearLockInternal(slot) }`，因为前面的逻辑已经完整处理了所有 LOCKED 状态。

---

### ✅ 问题2：抽取公共校验函数，消除重复代码

**问题**：`LockSlot`、`ConfirmAppointment`、`CreateAppointment` 三个方法存在大量重复的校验逻辑。

**修复**：在 [service.go](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/appointment/service.go#L16-L58) 中抽取了3个公共校验函数：

| 函数 | 职责 | 调用方 |
|------|------|--------|
| `checkSlotAvailableForLock(slot, now)` | 检查号源是否可锁定（状态、超时自动清理、剩余容量） | `LockSlot`, `CreateAppointment` |
| `checkSlotLockValid(slot, patientID, now)` | 检查锁定是否有效（状态、是否过期、所有者、容量） | `ConfirmAppointment` |
| `checkSlotHasCapacity(slot)` | 检查号源是否有剩余容量 | `ChangeAppointment` |

所有相关方法均已重构使用这些公共函数，后续校验策略变更只需修改一处。

---

### ✅ 问题3：细化锁粒度到号源级别

**问题**：全局互斥锁导致操作不同号源的并发请求被串行化，浪费了号源之间的天然并行性。

**修复**：

1. **数据结构修改**：[ScheduleSlot](file:///C:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/appointment/types.go#L50-L65) 结构体新增 `mu sync.Mutex` 字段

2. **锁策略分层**：
   - **Store 级 `sync.RWMutex`**：保护全局 map 结构的读写
   - **号源级 `sync.Mutex`**：保护单个号源的内部状态（Status、BookedCount 等）

3. **各类操作的锁策略**：

| 操作 | 锁策略 | 并发性 |
|------|--------|--------|
| 查询类 (QuerySchedules, GetSlot 等) | Store 读锁 | 完全并行 |
| 号源锁定 (LockSlot, ReleaseLock) | Store 读锁 → 号源锁 | 不同号源可并行 |
| 超时检查 (CheckAndReleaseExpiredLocks) | Store 读锁 → 逐个号源锁 | 逐个检查，不阻塞其他查询 |
| 预约创建 (CreateAppointment, ConfirmAppointment) | 乐观检查（Store读锁→号源锁）→ 悲观确认（Store写锁→号源锁） | 不同号源预检查可并行 |
| 改约/取消/爽约 | Store 写锁 → 号源锁 | 全局串行（操作复杂，保证正确性） |

4. **死锁预防**：严格遵守「先 Store 锁，后号源锁」的获取顺序；双号源操作（改约）按号源 ID 字典序排序后依次加锁。

---

### 🧪 测试结果

```
go test ./internal/appointment/ -v -count=1 -timeout 30s

=== RUN   TestConcurrentAppointment_NoOversell
--- PASS: TestConcurrentAppointment_NoOversell (0.00s)
=== RUN   TestConcurrentAppointment_MultipleSlots
--- PASS: TestConcurrentAppointment_MultipleSlots (0.00s)

PASS
ok      solocoder-4-go/internal/appointment     0.229s
```

**73/73 测试用例全部通过**，包括并发超售测试验证在细粒度锁下仍不会超售。