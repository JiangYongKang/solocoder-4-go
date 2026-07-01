让我分析问题并进行修复。首先查看当前的相关代码。



        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          
我清楚地看到了两个问题：

1. **重复筛选逻辑**：`findAvailableBeds` 第148-150行和 `matchPatientCondition` 第170-172行都有床位类型检查，逻辑重复
2. **兜底分支问题**：第194行 `return true` 导致未知条件的患者被静默匹配任意床位

让我修复这两个问题：




        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          
changes: undefined
          
现在修复第二个问题：未知患者条件的兜底分支。同时检查 `TransferBed` 中是否也有类似的重复逻辑。







        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          
现在修复第二个问题：将 `matchPatientCondition` 末尾的 `return true` 改为 `return false`，确保未知条件的患者不会被静默匹配任意床位。







        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          

        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          
changes: undefined
          
现在补充未知患者条件处理的单元测试。









        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator_test.go
          
我注意到 `TransferBed` 函数中调用 `findAvailableBeds` 时没有传递 `PatientCondition`，这会导致空字符串被当作未知条件处理。让我修复 `matchPatientCondition` 函数，区分空条件和未知条件：









        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          
changes: undefined
          
现在添加未知患者条件处理的单元测试。









        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator_test.go
          
changes: undefined
          
现在运行测试确保全部通过。









        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/bedalloc/ -v undefined
          
所有测试通过！让我运行诊断检查确保代码质量。











        
            
toolName: GetDiagnostics
            
status: success
          
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 4 items
          
## 问题修复完成

已成功修复床位调度模块的两个问题，所有测试通过。

### 修复内容

**1. 消除重复筛选逻辑**

修复 [findAvailableBeds](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/bedalloc/allocator.go#L140-L163) 函数，移除了重复的床位类型检查：

```go
// 移除了重复的检查：
// if criteria.BedType != "" && bed.Type != criteria.BedType {
//     continue
// }

// 现在只通过 matchPatientCondition 统一处理所有匹配逻辑
if !ba.matchPatientCondition(bed, criteria) {
    continue
}
```

这样床位类型的筛选逻辑只在 [matchPatientCondition](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/bedalloc/allocator.go#L165-L191) 中维护，避免了两处逻辑不一致的风险。

**2. 修复未知患者条件静默匹配问题**

修改 [matchPatientCondition](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/bedalloc/allocator.go#L165-L191) 函数末尾的兜底分支：

```go
// 修复前（有问题）：
return true  // 未知条件会静默匹配任意床位

// 修复后：
if criteria.PatientCondition == "general" || criteria.PatientCondition == "" {
    return bed.Type == BedTypeGeneral || bed.Type == BedTypeSurgery
}

return false  // 真正的未知条件（如拼写错误）返回 false
```

同时将空字符串 `""` 视为普通患者条件处理，保持与 `TransferBed` 等调用方的兼容性。

**3. 新增 5 个单元测试**

| 测试函数 | 覆盖场景 |
|---------|---------|
| `TestMultiCondition_UnknownPatientCondition` | 拼写错误的条件（如 "genral"）返回 ErrNoAvailableBed |
| `TestMultiCondition_EmptyPatientCondition` | 空条件被视为普通患者处理 |
| `TestMultiCondition_UnknownConditionNoBedTypeSpecified` | 未指定床位类型时，未知条件被拒绝 |
| `TestMultiCondition_UnknownConditionWithBedTypeSpecified` | 明确指定床位类型时，即使条件未知也可分配 |
| `TestMultiCondition_SingleMatchLogicSource` | 验证床位类型筛选只在 matchPatientCondition 中发生 |

### 测试结果

```
go test ./internal/bedalloc/ -v
```

**所有 46 个测试用例全部通过 ✓**，代码诊断无错误。