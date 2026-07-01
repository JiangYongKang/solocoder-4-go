我来修复这些问题。首先让我查看当前的代码实现：


        
            
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
          
首先添加新的错误变量到 models.go：





        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\models.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在修改 ReviewPrescription 方法：











        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在重构 CreatePrescription 方法，将库存检查和实际预留合并到同一次遍历中：











        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在需要修改测试文件。首先查看现有的 `TestReviewPrescription_RejectedWithoutReason` 测试，然后修改并添加新的测试用例：











        
            
toolName: view_files
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service_test.go
          
现在修改这个测试，并添加新的测试用例：












        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go\internal\prescription\service_test.go
          
changes: undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
现在运行测试验证所有修改：















        
            
toolName: run_command
            
status: success
          
command: cd c:\Users\vince\GoletaLab\SoloCoder-4\solocoder-4-go ; go test ./internal/prescription/ -v undefined
          

        
            
toolName: todo_write
            
status: success
          
Todos updated: 5 items
          
所有问题已修复完成！**34 个测试用例全部通过** ✅

## 修复内容总结

### 1. 新增错误类型
在 [models.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/models.go#L113) 中新增 `ErrRejectReasonRequired` 错误变量，用于标识驳回处方时缺少原因的参数错误。

### 2. 修改 ReviewPrescription 驳回逻辑
在 [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L157-L161) 中，驳回处方时增加前置校验：
- **修改前**：当 `req.RejectReason` 为空时，系统自动填充"未提供驳回原因"并继续执行驳回
- **修改后**：当 `req.RejectReason` 为空时，直接返回 `ErrRejectReasonRequired` 错误，不修改处方状态，不释放库存

### 3. 重构 CreatePrescription 库存预留逻辑
在 [service.go](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L75-L115) 中，将原来的两次遍历合并为一次：
- **修改前**：第一次遍历检查库存可用性并收集 `reservedStock` map，第二次遍历调用 `reserveStock` 实际预留
- **修改后**：同一次遍历中，检查药品信息后**立即调用 `reserveStock` 预留库存**，消除 TOCTOU 风险
- 增加了失败回滚逻辑：遍历过程中任何一步失败（数量非法、药品不存在、库存不足），都会释放已成功预留的所有药品库存

### 4. 测试用例更新
- 修改了 `TestReviewPrescription_RejectedWithoutReason`：验证驳回时空原因返回 `ErrRejectReasonRequired`，处方状态保持不变，库存保持占用
- 新增 `TestReviewPrescription_RejectedWithEmptyReason_ReturnsError`：更全面地验证多药品处方在驳回空原因时，所有状态（处方状态、驳回原因、审核信息、库存占用和数量）都保持不变

### 关键代码片段

**驳回原因校验** ([service.go#L157-L161](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L157-L161))：
```go
if !req.Approved {
    if req.RejectReason == "" {
        return nil, ErrRejectReasonRequired
    }
}
```

**库存预留合并+失败回滚** ([service.go#L79-L100](file:///c:/Users/vince/GoletaLab/SoloCoder-4/solocoder-4-go/internal/prescription/service.go#L79-L100))：
```go
for _, item := range req.Items {
    if item.Quantity <= 0 {
        for medID, qty := range reservedStock {
            svc.store.releaseStock(medID, qty)
        }
        return nil, ErrInvalidQuantity
    }
    // ... 药品存在性检查 ...
    if err := svc.store.reserveStock(medicine.ID, item.Quantity); err != nil {
        for medID, qty := range reservedStock {
            svc.store.releaseStock(medID, qty)
        }
        return nil, err
    }
    // ... 继续处理 ...
}
```