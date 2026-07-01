让我分析并修复 `matchPatientCondition` 方法的逻辑问题，然后补充相关测试用例。

        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          
让我先仔细阅读当前的 `matchPatientCondition` 方法：


        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          
现在我理解了问题。当前逻辑存在两个缺陷：

1. **儿童患者匹配逻辑问题**：第一个 if 分支在检查儿科床位时，没有考虑儿童患者可能同时有其他特殊病情（如 ICU）。当儿童需要 ICU 时，会被错误地排除非儿科床位。

2. **手术患者ICU限制**：手术患者被限制只能分配手术或普通床位，无法覆盖术后需要 ICU 监护的场景。

让我重新设计匹配逻辑，按优先级处理条件：


        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator.go
          
changes: undefined
          
现在更新 Todo 并补充多条件组合匹配的单元测试。





        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          
让我先读取测试文件末尾，然后添加新的测试用例。




        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator_test.go
          
现在添加多条件组合匹配场景的单元测试。





        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\bedalloc\allocator_test.go
          
changes: undefined
          
现在运行测试确保全部通过。





        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          

        
            
toolName: run_command
            
status: success
          
command: go test ./internal/bedalloc/ -v undefined
          
所有测试通过！现在运行诊断检查确保代码质量。







        
            
toolName: todo_write
            
status: success
          
Todos updated: 3 items
          

        
            
toolName: GetDiagnostics
            
status: success
          
          
诊断检查通过，无错误。

## 问题修复完成

已成功修复 `matchPatientCondition` 方法的逻辑问题，并补充了完整的单元测试。

### 修复内容

**1. 重构 [matchPatientCondition](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/bedalloc/allocator.go#L169-L195) 方法**

原逻辑存在两个问题：
- **儿童患者匹配问题**：第一个 if 分支在检查儿科床位时，没有考虑儿童患者可能同时有其他特殊病情（如 ICU），导致需要 ICU 的儿童被错误地排除非儿科床位
- **手术患者ICU限制**：手术患者被限制只能分配手术或普通床位，无法覆盖术后需要 ICU 监护的场景

修复后的逻辑按优先级处理：
1. **明确指定床位类型**：如果 `criteria.BedType != ""`，只匹配该类型（最高优先级）
2. **传染病硬性约束**：`infectious` 只能分配隔离床位
3. **ICU硬性约束**：`icu` 只能分配 ICU 床位
4. **儿童普通患者**：年龄<14 且病情为 `general`，只能分配儿科床位
5. **手术患者**：可分配手术、ICU 或普通床位（支持术后 ICU 监护）
6. **普通患者**：可分配普通或手术床位

**2. 补充 9 个多条件组合测试用例**

在 [allocator_test.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/bedalloc/allocator_test.go) 中新增测试：

| 测试函数 | 覆盖场景 |
|---------|---------|
| `TestMultiCondition_ChildWithICU` | 儿童ICU患者能分配ICU床位，不被儿科限制错误排除 |
| `TestMultiCondition_SurgeryWithICU` | 手术患者能分配ICU床位用于术后监护 |
| `TestMultiCondition_SurgeryCanChooseBedType` | 手术患者可明确选择不同床位类型 |
| `TestMultiCondition_InfectiousNoIsolationBed` | 传染病患者无隔离床位时返回错误，不回退 |
| `TestMultiCondition_ChildGeneralNoPediatricBed` | 儿童普通患者无儿科床位时返回错误 |
| `TestMultiCondition_SpecifiedBedTypeOverridesCondition` | 明确指定床位类型时忽略患者条件 |
| `TestMultiCondition_ChildSurgeryCanUseSurgeryBed` | 儿童手术患者能分配手术床位 |
| `TestMultiCondition_ICUHardConstraint` | ICU患者必须分配ICU床位的硬性约束 |
| `TestMultiCondition_GeneralPatientCanUseSurgeryBed` | 普通患者可使用手术床位 |

### 测试结果

```
go test ./internal/bedalloc/ -v
```

**所有 41 个测试用例全部通过 ✓**，代码诊断无错误。