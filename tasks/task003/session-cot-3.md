我来修复这两个问题。首先查看当前代码：



        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\models.go
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
首先添加新的错误变量：


















        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\models.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在修改 WithdrawPrescription 方法：























        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在查看 CreatePrescription 方法中的重复回滚代码：















































































        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          
现在添加 `rollbackReservedStock` 辅助函数并重构 CreatePrescription：

















































































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          
changes: undefined
          
现在用辅助函数替换三处重复的回滚代码：


















































































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在查看测试文件中与撤回相关的测试用例：

















































































        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service_test.go
          
现在添加测试用例。首先查看测试文件的末尾部分，找到合适的插入位置：

















































































        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service_test.go
          

        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service_test.go
          
现在添加测试用例。在 `TestWithdrawPrescription_InvalidPrescription` 之后插入新的测试：

















































































        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service_test.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在运行测试：


















































































        
            
toolName: run_command
            
status: success
          
command: cd c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go ; go test ./internal/prescription/ -v undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
所有问题已修复完成！**40 个测试用例全部通过** ✅

## 修复内容总结

### 1. 新增错误类型
在 [models.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/models.go#L114) 中新增 `ErrWithdrawReasonRequired` 错误变量。

### 2. 修改 WithdrawPrescription 撤回逻辑
在 [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L215-L241) 中增加前置校验：
- **修改前**：当 `req.Reason` 为空时，系统自动填充"未提供撤回原因"并继续执行
- **修改后**：当 `req.Reason` 为空时，直接返回 `ErrWithdrawReasonRequired` 错误，不修改处方状态，不释放库存

**关键代码** ([service.go#L219-L221](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L219-L221))：
```go
if req.Reason == "" {
    return nil, ErrWithdrawReasonRequired
}
```

### 3. 消除 CreatePrescription 回滚代码重复
添加 `rollbackReservedStock` 辅助函数 ([service.go#L15-L18](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L15-L18))：
```go
func (svc *Service) rollbackReservedStock(reservedStock map[string]int) {
    for medID, qty := range reservedStock {
        svc.store.releaseStock(medID, qty)
    }
}
```

将原来 3 处完全相同的回滚代码（每处 3 行，共 9 行）替换为对 `rollbackReservedStock` 的调用（每处 1 行）。修改后的代码 ([service.go#L85-L99](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L85-L99))：
```go
for _, item := range req.Items {
    if item.Quantity <= 0 {
        svc.rollbackReservedStock(reservedStock)  // 统一调用
        return nil, ErrInvalidQuantity
    }
    medicine, exists := svc.store.medicines[item.MedicineID]
    if !exists {
        svc.rollbackReservedStock(reservedStock)  // 统一调用
        return nil, ErrMedicineNotFound
    }
    if err := svc.store.reserveStock(medicine.ID, item.Quantity); err != nil {
        svc.rollbackReservedStock(reservedStock)  // 统一调用
        return nil, err
    }
    // ...
}
```

**好处**：回滚逻辑只需在一处维护，修改时不会遗漏。

### 4. 新增测试用例（6 个）

| 测试用例 | 验证内容 |
|---------|---------|
| `TestWithdrawPrescription_EmptyReason` | PENDING_REVIEW 状态下撤回空原因返回错误，处方状态和库存保持不变 |
| `TestWithdrawPrescription_EmptyReasonFromApproved` | APPROVED 状态下撤回空原因返回错误，处方状态和库存保持不变 |
| `TestCreatePrescription_RollbackOnInvalidQuantity` | 数量非法时已占用的库存正确回滚 |
| `TestCreatePrescription_RollbackOnInvalidMedicine` | 药品不存在时已占用的库存正确回滚 |
| `TestCreatePrescription_RollbackOnInsufficientStock` | 库存不足时已占用的库存正确回滚 |

加上之前修改的 2 个测试，本轮共新增/修改了 8 个测试用例。

### 测试结果
```
PASS
ok  solocoder-4-go/internal/prescription    0.211s
```
**40 个测试用例，全部通过**。