好的，我来修复这两个问题。


        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
先读取需要修改的文件：



        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\types.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service.go
          
好的，现在开始修改。先添加新的错误类型：




        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\types.go
          
changes: undefined
          
现在修改 SendMessage 的 default 分支和 SendFollowUp：





        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service.go
          
changes: undefined
          
现在修改 SendFollowUp，去掉同步写回 session 的逻辑：







        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service.go
          
changes: undefined
          
现在更新测试用例。先查看受影响的测试：




        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: search_by_regex
            
status: success
          
query: TestSendFollowUp_UpdatesSessionMessages
          
现在修改测试用例并重命名，同时添加新的测试：





        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          
changes: undefined
          
现在添加一个测试 ErrInvalidSenderType 的用例：





        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\teleconsult\service_test.go
          
changes: undefined
          
现在运行测试验证修复：







        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/teleconsult/ -v undefined
          
所有测试通过！让我完成任务并更新文档中的错误速查表：







        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
所有问题已修复完成！以下是修改总结：

## 修复内容

### 1. 错误类型语义修复 — `ErrInvalidSenderType`

**之前的问题**：`SendMessage` 中 `switch` 的 `default` 分支返回 `ErrNotPatient`，当传入 `SYSTEM` 或其他类型时，错误信息"不是患者"在语义上具有误导性。

**修复**：
- 在 [types.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/teleconsult/types.go#L119) 新增错误变量：
  ```go
  ErrInvalidSenderType = errors.New("invalid message sender type, must be PATIENT or DOCTOR")
  ```
- 在 [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/teleconsult/service.go#L281-L283) 的 `SendMessage` default 分支改为返回 `ErrInvalidSenderType`

---

### 2. SendFollowUp 不再同步写回原始 Session

**之前的问题**：追问消息同步追加到 `session.Messages`，存在两个隐患：
- 若后续新增会话清理逻辑，`sessionExists` 判断后的引用存在空指针风险
- 业务语义上，归档后的追问属于"归档追加"，不应再修改已结束的原始会话

**修复**：
- 在 [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/teleconsult/service.go#L401-L404) 的 `SendFollowUp` 中移除了这段逻辑：
  ```go
  // 已删除：
  session, sessionExists := s.sessions[archive.SessionID]
  if sessionExists {
      session.Messages = append(session.Messages, msg)
      session.FollowUpCount = archive.FollowUpCount
      session.LastMessageAt = now
  }
  ```
- 追问消息现在**仅保留在 ArchiveRecord.Messages** 中

---

### 3. 测试用例更新（52 个测试全部通过）

| 变更 | 说明 |
|------|------|
| 新增 `TestSendMessage_InvalidSenderType` | 覆盖 `SYSTEM` 和未知类型两种场景 |
| 重命名 `TestSendFollowUp_UpdatesSessionMessages` → `TestSendFollowUp_DoesNotUpdateOriginalSession` | 反向验证：session 的消息数和 FollowUpCount **保持不变**，而 archive 的消息数正确递增 |

运行 `go test ./internal/teleconsult/ -v`：**52 PASS, 0 FAIL** ✅