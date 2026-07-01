我来修复这个问题。首先创建任务计划。





        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\types.go
          
现在修改 `Error()` 方法，在错误消息中包含已有提交记录的 ID 信息：







        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\types.go
          
changes: undefined
          
需要在 types.go 中添加 `fmt` 包的导入：










        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\types.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\types.go
          
changes: undefined
          
现在更新测试用例，验证错误消息中包含 ID 信息：















        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
changes: undefined
          
需要在测试文件中添加 `strings` 包的导入：











        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
changes: undefined
          
现在更新需求文档中关于错误消息的说明，然后运行测试：

















        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\docs\examgrading.md
          
changes: undefined
          
现在运行测试：

















        
            
toolName: run_command
            
status: success
          
command: go test ./internal/examgrading/ -v undefined
          
所有测试都通过了！更新任务状态：


















        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          
## 修复完成总结

所有问题已成功修复，**42 个测试用例全部通过**。以下是修复内容：

### 1. 修改错误消息格式

**问题**：`DuplicateSubmissionError` 的 `Error()` 方法只返回通用提示，没有包含已有提交记录的 ID 信息，调用方仅通过错误文本无法知道已有提交是哪一个。

**修复**（[types.go L33-L35](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/examgrading/types.go#L33-L35)）：

```go
func (e *DuplicateSubmissionError) Error() string {
    return fmt.Sprintf("%s; existing submission ID: %s", ErrAlreadySubmitted.Error(), e.ExistingSubmissionID)
}
```

**效果**：错误消息格式变为：
```
student has already submitted this exam; existing submission ID: SUB0000000009
```

这样调用方即使不使用类型断言，也能直接从错误文本中看到已有提交记录的 ID。

### 2. 更新测试用例

**TestDuplicateSubmissionReturnsExistingID**（[service_test.go L1937-L1947](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/examgrading/service_test.go#L1937-L1947)）：

- 验证错误消息包含通用提示文本
- 验证错误消息包含已有提交记录的 ID
- 测试输出：`Error message: student has already submitted this exam; existing submission ID: SUB0000000009`

### 3. 向后兼容性

- 通过 `Unwrap() error` 方法，`errors.Is(err, ErrAlreadySubmitted)` 仍然正常工作
- 调用方可以选择：
  - **简单方式**：直接打印或解析错误文本获取 ID
  - **结构化方式**：使用 `errors.As(err, &dupErr)` 进行类型断言获取结构化信息

### 调用方使用示例

```go
_, err := store.SubmitExam(submitReq)
if err != nil {
    // 方式1：直接从错误文本中获取
    fmt.Println(err) // 输出: student has already submitted this exam; existing submission ID: SUB0000000009

    // 方式2：使用 errors.Is 检查错误类型
    if errors.Is(err, examgrading.ErrAlreadySubmitted) {
        // 方式3：使用 errors.As 获取结构化信息
        var dupErr *examgrading.DuplicateSubmissionError
        if errors.As(err, &dupErr) {
            fmt.Printf("已有提交记录ID: %s\n", dupErr.ExistingSubmissionID)
        }
    }
}
```