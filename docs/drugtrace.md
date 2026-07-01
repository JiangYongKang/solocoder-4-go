# 药品批次追溯模块需求文档

## 1. 模块概述

药品批次追溯模块（drugtrace）是医院药品管理系统的核心组件，用于实现药品从入库到出库的全流程跟踪管理。模块支持批次入库、有效期预警、先进先出出库、召回管理和流向追溯等关键功能，确保药品使用安全，符合药事管理规范。

**包路径**: `internal/drugtrace/`

## 2. 功能需求

### 2.1 批号入库管理

**功能描述**: 支持按药品和批号登记入库信息，包括数量、生产日期、有效期和供应来源。

**输入参数**:
- `drugCode`: 药品编码
- `batchNumber`: 批号
- `quantity`: 入库数量（必须 > 0）
- `productionDate`: 生产日期
- `expiryDate`: 有效期（必须晚于生产日期）
- `supplier`: 供应来源
- `operator`: 操作员

**业务规则**:
- 同一批号不可重复入库
- 入库时自动检测有效期，已过期的批次状态标记为 `Expired`
- 自动记录库存流水

### 2.2 有效期预警

**功能描述**: 支持查询临近过期和已经过期的批次，出库时自动阻止使用过期批次。

**查询功能**:
- `GetExpiringBatches(days int)`: 查询未来 N 天内即将过期的批次
- `GetExpiredBatches()`: 查询所有已过期的批次
- `UpdateBatchStatus()`: 批量更新批次状态（正常 → 过期）

**预警规则**:
- 过期批次在出库时自动被过滤，不参与库存扣减
- 临近过期批次可正常使用，但应在预警列表中提示

### 2.3 出库先进先出（FIFO）

**功能描述**: 药品出库时优先扣减更早到期的批次，同有效期时优先扣减更早入库的批次。

**扣减优先级**:
1. 优先按有效期从早到晚排序
2. 有效期相同时按入库时间从早到晚排序
3. 自动跳过已过期和已召回的批次

**输出结果**:
- 出库明细列表（每个批次的扣减数量）
- 总出库数量
- 每条出库记录包含：批次号、药品编码、数量、科室、患者、操作员、时间

### 2.4 批次召回与流向查询

**功能描述**: 对有质量问题的批次执行召回锁定，并查询该批次已流向的科室或患者记录。

**召回功能**:
- `RecallBatch(batchNumber, reason, operator)`: 标记批次为召回状态
- 被召回的批次不得继续出库
- 召回操作记录库存流水

**流向查询**:
- `GetBatchFlowTrace(batchNumber)`: 查询该批次所有出库记录
- 返回按时间排序的流向明细，包含科室、患者等信息

## 3. 核心结构体职责

### 3.1 Drug（药品）

```go
type Drug struct {
    Code string  // 药品编码（唯一标识）
    Name string  // 药品名称
    Spec string  // 规格
}
```

**职责**: 存储药品基础信息，作为批次管理的主数据。

### 3.2 Batch（批次）

```go
type Batch struct {
    BatchNumber    string       // 批号
    DrugCode       string       // 关联药品编码
    Quantity       int          // 入库总数量
    RemainingQty   int          // 剩余可用数量
    ProductionDate time.Time    // 生产日期
    ExpiryDate     time.Time    // 有效期
    Supplier       string       // 供应商
    Status         BatchStatus  // 批次状态
    InboundTime    time.Time    // 入库时间
    RecallReason   string       // 召回原因
    RecallTime     *time.Time   // 召回时间
}
```

**职责**: 管理单个药品批次的全生命周期，包括库存、状态和追溯信息。

### 3.3 BatchStatus（批次状态）

```go
type BatchStatus int

const (
    BatchStatusNormal   BatchStatus = iota  // 正常
    BatchStatusExpired  BatchStatus = iota  // 已过期
    BatchStatusRecalled BatchStatus = iota  // 已召回
)
```

**职责**: 定义批次的状态枚举，控制批次的可用性。

### 3.4 StockFlow（库存流水）

```go
type StockFlow struct {
    ID          string    // 流水号
    BatchNumber string    // 关联批号
    DrugCode    string    // 药品编码
    FlowType    string    // 类型：inbound/outbound/recall
    Quantity    int       // 变动数量
    Operator    string    // 操作员
    Time        time.Time // 操作时间
    Remark      string    // 备注
}
```

**职责**: 记录每一次库存变动，用于审计和追溯。

### 3.5 OutboundDetail（出库明细）

```go
type OutboundDetail struct {
    ID           string    // 出库单号
    BatchNumber  string    // 批号
    DrugCode     string    // 药品编码
    Quantity     int       // 出库数量
    Department   string    // 领用科室
    Patient      string    // 使用患者
    Operator     string    // 发药药师
    OutboundTime time.Time // 出库时间
}
```

**职责**: 记录每一笔出库的详细流向，用于药品使用追溯。

### 3.6 DrugTraceService（主服务）

```go
type DrugTraceService struct {
    mu              sync.RWMutex       // 读写锁，保证并发安全
    drugs           map[string]*Drug   // 药品字典
    batches         map[string]*Batch  // 批次字典
    drugBatches     map[string][]*Batch // 药品-批次索引
    stockFlows      []*StockFlow       // 库存流水历史
    outboundDetails []*OutboundDetail  // 出库明细历史
}
```

**职责**: 提供所有业务操作的入口，维护内存数据结构，保证并发安全。

## 4. 批次状态变化

### 4.1 状态流转图

```
          入库时检测
Normal ──────────────→ Expired
  │                     ↑
  │ 手动召回           │ 自动更新
  └────────→ Recalled   │
            │           │
            └───────────┘
          召回后不再变化
```

### 4.2 状态转换规则

| 当前状态 | 事件 | 目标状态 | 说明 |
|---------|------|---------|------|
| - | 入库（有效期未过） | Normal | 正常入库 |
| - | 入库（有效期已过） | Expired | 入库时自动标记 |
| Normal | 到达有效期 | Expired | `UpdateBatchStatus()` 或出库时自动更新 |
| Normal | 手动召回 | Recalled | `RecallBatch()` 触发 |
| Expired | 手动召回 | Recalled | 已过期批次仍可被召回 |
| Recalled | 任何操作 | Recalled | 终止状态，不可逆转 |

## 5. 使用示例

### 5.1 初始化服务

```go
import "solocoder-4-go/internal/drugtrace"

service := drugtrace.NewDrugTraceService()
```

### 5.2 添加药品

```go
err := service.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")
if err != nil {
    log.Fatal(err)
}
```

### 5.3 批次入库

```go
prodDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
expDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

err = service.InboundBatch(
    "DRUG001",
    "B20240101",
    500,
    prodDate,
    expDate,
    "华北制药",
    "库管员A",
)
if err != nil {
    log.Fatal(err)
}
```

### 5.4 查询有效期预警

```go
// 查询30天内即将过期的批次
expiring, err := service.GetExpiringBatches(30)
if err != nil {
    log.Fatal(err)
}
for _, batch := range expiring {
    fmt.Printf("批号 %s 将于 %s 过期\n", batch.BatchNumber, batch.ExpiryDate.Format("2006-01-02"))
}
```

### 5.5 药品出库（FIFO）

```go
result, err := service.OutboundFIFO(
    "DRUG001",
    600,
    "内科",
    "张小明",
    "药师李",
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("总出库数量: %d\n", result.TotalQty)
for _, detail := range result.Details {
    fmt.Printf("批次 %s 出库 %d 盒\n", detail.BatchNumber, detail.Quantity)
}
```

### 5.6 批次召回

```go
err = service.RecallBatch("B20240101", "检测出杂质超标", "质量管理员")
if err != nil {
    log.Fatal(err)
}
```

### 5.7 查询批次流向

```go
trace, err := service.GetBatchFlowTrace("B20240101")
if err != nil {
    log.Fatal(err)
}

fmt.Println("批次流向记录:")
for _, record := range trace {
    fmt.Printf("时间: %s, 科室: %s, 患者: %s, 数量: %d\n",
        record.OutboundTime.Format("2006-01-02 15:04:05"),
        record.Department,
        record.Patient,
        record.Quantity,
    )
}
```

### 5.8 查询库存

```go
stock, err := service.GetDrugStock("DRUG001")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("当前可用库存: %d\n", stock)
```

## 6. 错误码定义

| 错误变量 | 说明 |
|---------|------|
| `ErrBatchNotFound` | 批次不存在 |
| `ErrDrugNotFound` | 药品不存在 |
| `ErrInsufficientStock` | 库存不足 |
| `ErrBatchExpired` | 批次已过期 |
| `ErrBatchRecalled` | 批次已召回 |
| `ErrInvalidQuantity` | 数量无效（≤0） |
| `ErrInvalidDateRange` | 生产日期晚于有效期 |
| `ErrBatchAlreadyExists` | 批号已存在 |
| `ErrInvalidBatchNumber` | 批号为空 |
| `ErrInvalidDrugCode` | 药品编码为空 |

## 7. 测试覆盖

模块包含完整的单元测试（`drugtrace_test.go`），覆盖以下场景：

- **正常流程**: 药品管理、入库、出库、召回、查询等
- **边界条件**: 有效期临界值、库存刚好充足、多批次组合出库等
- **异常分支**: 参数校验、库存不足、过期/召回批次阻止出库等
- **并发安全**: 使用 `sync.RWMutex` 保证线程安全

运行测试:

```bash
go test ./internal/drugtrace/ -v
```
