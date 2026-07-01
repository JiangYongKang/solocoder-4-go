我来修复这两个问题。首先创建任务计划。


        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
让我先查看当前的代码，然后进行修复。

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service.go
          
现在我来修复第一个问题 - 在排序前创建切片的副本：



        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service.go
          
changes: undefined
          
现在修复第二个问题。我需要创建一个自定义错误类型来包含已有提交记录的信息。让我先看看当前的错误定义：



        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\types.go
          
我需要创建一个自定义错误类型来包含已有提交记录的信息。先修改 types.go：



        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\types.go
          
changes: undefined
          
现在修改 SubmitExam 方法，在重复交卷时返回包含已有提交记录信息的错误：




        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service.go
          
changes: undefined
          
现在添加测试用例。先更新任务状态，然后读取测试文件的末尾来添加新的测试：






        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
让我查看测试文件的末尾：







        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
让我找到文件的最后部分：









        
            
toolName: search_by_regex
            
status: success
          
query: func TestZeroScore
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
现在在测试文件末尾添加两个新的测试用例：











        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
changes: undefined
          
现在需要在测试文件中添加 `errors` 包的导入：













        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
changes: undefined
          
现在更新任务状态并运行测试：

























        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/examgrading/ -v undefined
          
有两个测试需要更新，因为现在返回的是自定义错误类型，需要使用 `errors.Is` 来检查。让我修复这些测试：





        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
changes: undefined
          
现在修复 TestFullWorkflow 中的类似问题：







        
            
toolName: search_by_regex
            
status: success
          
query: Step 4 - Expected ErrAlreadySubmitted
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\examgrading\service_test.go
          
changes: undefined
          
现在再次运行测试：













        
            
toolName: run_command
            
status: success
          
command: go test ./internal/examgrading/ -v undefined
          
所有测试都通过了！现在更新需求文档中关于错误处理的部分：







        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在更新需求文档，添加关于自定义错误类型的说明：









        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\docs\examgrading.md
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\docs\examgrading.md
          
changes: undefined
          
文档编辑出现了格式问题，让我修复它：












        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\docs\examgrading.md
          
我需要修复文档的格式问题。表格被打断了，让我重新整理：










        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\docs\examgrading.md
          
changes: undefined
          
现在文档格式已经修复。让我最后再运行一次测试确认所有测试通过：









        
            
toolName: run_command
            
status: success
          
command: go test ./internal/examgrading/ -v -run "TestMultipleChoiceOriginalOrderPreserved|TestDuplicateSubmissionReturnsExistingID|TestSubmitExam_DuplicateSubmission|TestFullWorkflow" undefined
          
## 修复完成总结

所有问题已成功修复，**42 个测试用例全部通过**。以下是修复内容：

### 1. 修复多选题答案排序问题

**问题**：`compareMultipleChoiceAnswers` 函数中 `sort.Strings` 会原地修改切片，导致 `Exam` 结构体中原始答案顺序被永久改变。

**修复**（[service.go L241-L247](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/examgrading/service.go#L241-L247)）：
```go
correctCopy := make([]string, len(correctSlice))
copy(correctCopy, correctSlice)
studentCopy := make([]string, len(studentSlice))
copy(studentCopy, studentSlice)

sort.Strings(correctCopy)
sort.Strings(studentCopy)
```
在排序前创建切片副本，对副本排序比较，不修改原始数据。

### 2. 修复重复交卷返回信息不足问题

**问题**：重复交卷时只返回 `ErrAlreadySubmitted` 错误，调用方无法获知已有提交记录的 ID。

**修复**：
- 在 [types.go L27-L39](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/examgrading/types.go#L27-L39) 新增 `DuplicateSubmissionError` 自定义错误类型，包含：
  - `ExistingSubmissionID`：已有提交记录 ID
  - `StudentID`：学生 ID
  - `ExamID`：试卷 ID
  - 实现 `Unwrap() error` 方法，支持 `errors.Is(err, ErrAlreadySubmitted)` 检查

- 在 [service.go L169-L175](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/examgrading/service.go#L169-L175) 修改 `SubmitExam` 方法，返回自定义错误：
```go
if existingSubmissionID, exists := s.submissionMap[submissionKey]; exists {
    return nil, &DuplicateSubmissionError{
        ExistingSubmissionID: existingSubmissionID,
        StudentID:            req.StudentID,
        ExamID:               req.ExamID,
    }
}
```

### 3. 新增测试用例

**TestMultipleChoiceOriginalOrderPreserved**（[service_test.go L1787-L1878](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/examgrading/service_test.go#L1787-L1878)）：
- 创建原始答案顺序为 `["B", "D", "A", "C"]` 的多选题
- 多次提交答卷后，验证原始答案顺序未被修改
- 测试输出：`Original answer order: [B D A C], After grading: [B D A C]`

**TestDuplicateSubmissionReturnsExistingID**（[service_test.go L1880-L1956](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/examgrading/service_test.go#L1880-L1956)）：
- 首次提交成功后再次提交
- 验证返回 `*DuplicateSubmissionError` 类型
- 验证 `ExistingSubmissionID`、`StudentID`、`ExamID` 字段正确
- 验证可以通过返回的 `ExistingSubmissionID` 查询到原有提交记录
- 测试输出：`Duplicate submission correctly returned existing ID: SUB0000000009`

### 4. 更新需求文档

在 [examgrading.md L259-L292](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/docs/examgrading.md#L259-L292) 新增 `7.2 自定义错误类型` 章节，详细说明 `DuplicateSubmissionError` 的使用方法。

### 调用方使用示例

```go
_, err := store.SubmitExam(submitReq)
if err != nil {
    if errors.Is(err, examgrading.ErrAlreadySubmitted) {
        var dupErr *examgrading.DuplicateSubmissionError
        if errors.As(err, &dupErr) {
            // 获取已有提交记录ID并查询
            existingSubmission, _ := store.GetSubmission(dupErr.ExistingSubmissionID)
            fmt.Printf("已有提交记录: %+v\n", existingSubmission)
        }
    }
}
```