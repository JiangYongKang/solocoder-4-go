# 体检套餐预约模块需求文档

## 1. 模块概述

体检套餐预约模块实现了从体检项目配置、套餐组合、时段预约、检查结果录入到报告生成的完整业务流程。模块使用内存数据结构模拟所有业务实体，支持并发安全的预约操作和容量控制，确保体检过程的数据一致性和业务规则的正确性。

### 主要功能特性

- **项目组合配置**：支持创建体检项目和套餐，套餐可组合多个项目，并校验重复项目和无效项目
- **预约时段容量控制**：患者预约套餐需选择体检时段，同一时段预约人数不超过容量限制
- **检查结果录入**：按预约和项目录入检查结果，未包含在套餐中的项目不得录入
- **报告生成与异常提醒**：所有必要结果录入后可生成报告，对超出参考范围的指标给出异常提醒
- **并发安全**：基于读写锁实现并发场景下的数据一致性
- **状态流转**：预约状态从待检查到检查中再到已完成的自动流转

## 2. 核心结构体职责

### 2.1 基础实体

#### Patient（患者）
- **职责**：存储患者基本信息，用于预约时的身份识别
- **关键字段**：
  - `ID`：患者唯一标识
  - `Name`：患者姓名
  - `Gender`：性别
  - `Age`：年龄
  - `Phone`：联系电话

#### CheckItem（体检项目）
- **职责**：描述单个体检项目的完整信息，包括参考范围和价格
- **关键字段**：
  - `ID`：项目唯一标识
  - `Name`：项目名称
  - `Description`：项目描述
  - `Category`：项目分类（检验/影像/物理/功能）
  - `Unit`：计量单位
  - `MinValue` / `MaxValue`：正常参考范围
  - `Price`：项目价格

### 2.2 套餐与时段

#### CheckPackage（体检套餐）
- **职责**：作为核心业务实体，承载套餐的完整信息和项目组合
- **关键字段**：
  - `ID`：套餐唯一标识
  - `Name`：套餐名称
  - `Description`：套餐描述
  - `ItemIDs`：包含的项目ID列表
  - `Items`：项目详情映射表（map[项目ID]*CheckItem）
  - `TotalPrice`：套餐总价（各项目价格之和）
  - `CreatedAt`：创建时间

#### TimeSlot（预约时段）
- **职责**：管理体检时段的容量控制，记录当前预约人数
- **关键字段**：
  - `ID`：时段唯一标识
  - `Date`：体检日期
  - `StartTime` / `EndTime`：时段起止时间
  - `Capacity`：容量上限
  - `CurrentCount`：当前已预约人数

### 2.3 预约与结果

#### Appointment（体检预约）
- **职责**：存储患者的体检预约信息和状态流转
- **关键字段**：
  - `ID`：预约唯一标识
  - `PatientID` / `PatientName`：患者信息
  - `PackageID` / `PackageName`：套餐信息
  - `TimeSlotID`：预约时段ID
  - `TimeSlotInfo`：时段信息描述
  - `Status`：预约状态（待检查/检查中/已完成/已取消）
  - `CreatedAt`：创建时间

#### CheckResult（检查结果）
- **职责**：记录单个项目的检查结果和异常判定
- **关键字段**：
  - `ID`：结果唯一标识
  - `AppointmentID`：关联预约ID
  - `ItemID` / `ItemName`：关联项目信息
  - `Value`：结果值（字符串形式）
  - `NumericValue`：数值结果（可解析为数值时）
  - `IsNumeric`：是否为数值类型
  - `IsAbnormal`：是否异常
  - `Unit`：计量单位
  - `Reference`：参考范围描述
  - `Remarks`：备注
  - `RecordedAt`：录入时间

### 2.4 报告相关

#### AbnormalItem（异常指标）
- **职责**：汇总报告中的异常指标，便于快速查看
- **关键字段**：
  - `ItemID` / `ItemName`：项目信息
  - `Value`：检测值
  - `Unit`：单位
  - `Reference`：参考范围
  - `Remarks`：备注

#### Report（体检报告）
- **职责**：作为最终输出实体，汇总所有检查结果和异常提醒
- **关键字段**：
  - `ID`：报告唯一标识
  - `AppointmentID`：关联预约ID
  - `PatientID` / `PatientName`：患者信息
  - `PackageID` / `PackageName`：套餐信息
  - `Results`：所有检查结果（map[项目ID]*CheckResult）
  - `AbnormalItems`：异常指标列表
  - `GeneratedAt`：生成时间
  - `Summary`：报告摘要

### 2.5 核心服务

#### Store（内存数据存储）
- **职责**：基于内存的数据存储层，提供所有实体的CRUD操作和业务逻辑
- **核心能力**：
  - 读写锁（sync.RWMutex）保证并发安全
  - 患者的增删查
  - 体检项目的增删查和列表
  - 套餐的创建、查询、列表（含重复/无效项目校验）
  - 时段的创建、查询、列表（含容量控制）
  - 预约的创建、取消、查询、列表（含容量校验）
  - 检查结果的录入、查询（含套餐项目校验）
  - 报告的生成和查询（含异常指标判定）
  - 自动ID生成（带前缀：PAT/ITM/PKG/TMS/APT/RES/RPT）

## 3. 核心业务流程

### 3.1 项目与套餐配置流程

```
1. 创建体检项目（定义名称、分类、参考范围、价格）
   ↓
2. 创建体检套餐
   ↓
3. 校验：
   - 套餐至少包含1个项目
   - 无重复项目
   - 所有项目ID均有效（存在于项目库）
   ↓
4. 自动计算套餐总价（各项目价格之和）
   ↓
5. 保存套餐信息
```

### 3.2 体检预约流程

```
1. 患者选择套餐和时段发起预约
   ↓
2. 校验：
   - 患者存在
   - 套餐存在
   - 时段存在
   - 时段当前预约人数 < 容量上限
   ↓
3. 时段当前预约人数 +1
   ↓
4. 创建预约记录，状态设为 PENDING（待检查）
   ↓
5. 初始化结果存储结构
```

### 3.3 检查结果录入流程

```
1. 按预约和项目录入结果
   ↓
2. 校验：
   - 预约存在
   - 预约未被取消
   - 项目包含在套餐中
   - 该项目尚未录入过结果（防重复）
   ↓
3. 异常判定（数值型结果）：
   - 解析结果值为浮点数
   - 与参考范围 MinValue / MaxValue 比较
   - 超出范围标记 IsAbnormal = true
   ↓
4. 自动流转预约状态：
   - 录入第1个结果：PENDING → CHECKING
   - 录入最后1个结果：CHECKING → COMPLETED
   ↓
5. 保存检查结果
```

### 3.4 报告生成流程

```
1. 请求生成报告（指定预约ID）
   ↓
2. 校验：
   - 预约存在
   - 预约未被取消
   - 套餐存在
   - 所有套餐项目均已录入结果（结果数 = 套餐项目数）
   ↓
3. 汇总异常指标：
   - 遍历所有结果
   - 筛选 IsAbnormal = true 的项目
   - 构造 AbnormalItem 列表
   ↓
4. 生成报告摘要：
   - 无异常：提示所有指标正常
   - 有异常：提示异常项数量，建议咨询医生
   ↓
5. 保存并返回报告
```

## 4. 状态定义

### 4.1 预约状态

| 状态 | 常量标识 | 说明 |
|------|----------|------|
| 待检查 | `AppointmentStatusPending` | 预约已创建，尚未开始任何检查 |
| 检查中 | `AppointmentStatusChecking` | 已录入部分检查结果，尚未全部完成 |
| 已完成 | `AppointmentStatusCompleted` | 所有套餐项目的结果均已录入 |
| 已取消 | `AppointmentStatusCancelled` | 预约被取消，释放时段容量 |

### 4.2 项目分类

| 分类 | 常量标识 | 说明 |
|------|----------|------|
| 检验类 | `ItemCategoryLaboratory` | 血液、尿液等实验室检查 |
| 影像类 | `ItemCategoryImaging` | X光、CT、超声等影像检查 |
| 物理类 | `ItemCategoryPhysical` | 身高、体重、血压等物理检查 |
| 功能类 | `ItemCategoryFunctional` | 心电图、肺功能等功能检查 |

## 5. 容量控制机制

### 5.1 实现原理

每个 `TimeSlot` 维护 `Capacity`（容量上限）和 `CurrentCount`（当前预约数）两个字段：

```go
type TimeSlot struct {
    Capacity     int  // 容量上限
    CurrentCount int  // 当前已预约人数
}
```

### 5.2 预约校验

创建预约时进行容量检查：
1. 读取时段的 `CurrentCount`
2. 若 `CurrentCount >= Capacity`，返回 `ErrTimeSlotCapacityFull`
3. 否则 `CurrentCount++`，创建预约成功

### 5.3 取消释放

取消预约时：
1. 若预约状态不是已取消，则将对应时段的 `CurrentCount--`
2. 预约状态设为 `CANCELLED`
3. 重复取消操作具有幂等性，不会导致 `CurrentCount` 变为负数

### 5.4 并发安全

整个预约创建过程在 `Store.mu` 写锁保护下执行，确保并发预约时：
- 容量检查和计数递增是原子操作
- 不会出现超卖（实际预约数超过容量）
- 取消预约与创建预约不会产生竞态条件

## 6. 异常指标判定规则

### 6.1 数值型结果

对于可解析为浮点数的结果值：
- 若项目定义了 `MinValue > 0` 或 `MaxValue > 0`，则进行范围判定
- `结果 < MinValue` → 异常（偏低）
- `结果 > MaxValue` → 异常（偏高）
- `MinValue ≤ 结果 ≤ MaxValue` → 正常

示例：
- 白细胞参考范围 4.0-10.0 ×10⁹/L
  - 结果 6.5 → 正常
  - 结果 15.0 → 异常（偏高）
  - 结果 2.5 → 异常（偏低）

### 6.2 非数值型结果

对于无法解析为浮点数的结果（如X光报告、心电图描述）：
- `IsNumeric = false`
- `IsAbnormal = false`（不进行自动异常判定）
- 异常判定由医生在 `Remarks` 中人工标注

## 7. 错误处理

### 7.1 错误类型定义

| 错误变量 | 说明 | 触发场景 |
|---------|------|---------|
| `ErrItemNotFound` | 体检项目不存在 | 查询/套餐中引用无效项目 |
| `ErrPackageNotFound` | 套餐不存在 | 查询/预约时引用无效套餐 |
| `ErrTimeSlotNotFound` | 时段不存在 | 查询/预约时引用无效时段 |
| `ErrAppointmentNotFound` | 预约不存在 | 查询/录入结果/生成报告时 |
| `ErrPatientNotFound` | 患者不存在 | 预约时患者ID无效 |
| `ErrDuplicateItemInPackage` | 套餐中重复项目 | 创建套餐时包含重复项目ID |
| `ErrInvalidItemInPackage` | 套餐中无效项目 | 创建套餐时引用不存在的项目 |
| `ErrEmptyPackageItems` | 套餐项目为空 | 创建套餐时项目列表为空 |
| `ErrTimeSlotCapacityFull` | 时段容量已满 | 预约时段已达人数上限 |
| `ErrInvalidTimeRange` | 时间范围非法 | 创建时段时开始时间≥结束时间 |
| `ErrInvalidCapacity` | 容量非法 | 创建时段时容量≤0 |
| `ErrResultNotFound` | 结果不存在 | 查询结果ID无效 |
| `ErrItemNotInPackage` | 项目不在套餐中 | 录入未包含在套餐中的项目结果 |
| `ErrDuplicateResult` | 结果重复录入 | 同一预约重复录入同一项目结果 |
| `ErrResultsIncomplete` | 结果不完整 | 未录完所有项目就生成报告 |
| `ErrReportNotFound` | 报告不存在 | 查询报告时预约ID无效 |
| `ErrAppointmentCancelled` | 预约已取消 | 对已取消预约录入结果/生成报告 |
| `ErrReportAlreadyGenerated` | 报告已生成 | （预留）重复生成报告时 |

## 8. 使用示例

### 8.1 初始化环境

```go
package main

import (
    "fmt"
    "solocoder-4-go/internal/checkup"
    "time"
)

func main() {
    store := checkup.NewStore()

    // 添加患者
    patientID := store.AddPatient(&checkup.Patient{
        Name:   "张三",
        Gender: "男",
        Age:    35,
        Phone:  "13800138000",
    })

    // 添加体检项目
    wbc := store.AddItem(&checkup.AddItemRequest{
        Name:        "血常规-白细胞计数",
        Category:    checkup.ItemCategoryLaboratory,
        Unit:        "×10⁹/L",
        MinValue:    4.0,
        MaxValue:    10.0,
        Price:       25.0,
    })

    rbc := store.AddItem(&checkup.AddItemRequest{
        Name:        "血常规-红细胞计数",
        Category:    checkup.ItemCategoryLaboratory,
        Unit:        "×10¹²/L",
        MinValue:    4.3,
        MaxValue:    5.8,
        Price:       20.0,
    })

    alt := store.AddItem(&checkup.AddItemRequest{
        Name:        "肝功能-谷丙转氨酶",
        Category:    checkup.ItemCategoryLaboratory,
        Unit:        "U/L",
        MinValue:    0,
        MaxValue:    40,
        Price:       30.0,
    })

    chest := store.AddItem(&checkup.AddItemRequest{
        Name:        "胸部X光",
        Category:    checkup.ItemCategoryImaging,
        Price:       100.0,
    })

    ecg := store.AddItem(&checkup.AddItemRequest{
        Name:        "心电图",
        Category:    checkup.ItemCategoryFunctional,
        Price:       50.0,
    })
}
```

### 8.2 完整流程：配置套餐 → 预约 → 录入结果 → 生成报告

```go
// 1. 创建体检套餐
pkg, err := store.CreatePackage(&checkup.CreatePackageRequest{
    Name:        "全面体检套餐A",
    Description: "包含血液检查、肝功能、胸部X光和心电图",
    ItemIDs:     []string{wbc.ID, rbc.ID, alt.ID, chest.ID, ecg.ID},
})
if err != nil {
    fmt.Printf("创建套餐失败: %v\n", err)
    return
}
fmt.Printf("套餐创建成功，ID: %s，总价: %.1f元，包含%d个项目\n",
    pkg.ID, pkg.TotalPrice, len(pkg.Items))

// 2. 创建预约时段
date := time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)
start := time.Date(2026, 7, 15, 8, 0, 0, 0, time.Local)
end := time.Date(2026, 7, 15, 12, 0, 0, 0, time.Local)
slot, err := store.CreateTimeSlot(&checkup.CreateTimeSlotRequest{
    Date:      date,
    StartTime: start,
    EndTime:   end,
    Capacity:  50,
})
if err != nil {
    fmt.Printf("创建时段失败: %v\n", err)
    return
}
fmt.Printf("时段创建成功，ID: %s，容量: %d人\n", slot.ID, slot.Capacity)

// 3. 创建预约
appt, err := store.CreateAppointment(&checkup.CreateAppointmentRequest{
    PatientID:  patientID,
    PackageID:  pkg.ID,
    TimeSlotID: slot.ID,
})
if err != nil {
    fmt.Printf("预约失败: %v\n", err)
    return
}
fmt.Printf("预约成功，ID: %s，状态: %s，时段: %s\n",
    appt.ID, appt.Status, appt.TimeSlotInfo)

// 4. 录入检查结果（含1项异常）
_, err = store.RecordResult(&checkup.RecordResultRequest{
    AppointmentID: appt.ID,
    ItemID:        wbc.ID,
    Value:         "12.5",
    Remarks:       "偏高，建议复查",
})
fmt.Printf("录入白细胞: %v, 异常: %v\n", err == nil, true)

_, err = store.RecordResult(&checkup.RecordResultRequest{
    AppointmentID: appt.ID,
    ItemID:        rbc.ID,
    Value:         "4.8",
})
fmt.Printf("录入红细胞: %v\n", err == nil)

_, err = store.RecordResult(&checkup.RecordResultRequest{
    AppointmentID: appt.ID,
    ItemID:        alt.ID,
    Value:         "35",
})
fmt.Printf("录入谷丙转氨酶: %v\n", err == nil)

_, err = store.RecordResult(&checkup.RecordResultRequest{
    AppointmentID: appt.ID,
    ItemID:        chest.ID,
    Value:         "未见异常",
    Remarks:       "心肺正常",
})
fmt.Printf("录入胸部X光: %v\n", err == nil)

_, err = store.RecordResult(&checkup.RecordResultRequest{
    AppointmentID: appt.ID,
    ItemID:        ecg.ID,
    Value:         "窦性心律",
    Remarks:       "正常心电图",
})
fmt.Printf("录入心电图: %v\n", err == nil)

// 5. 查看预约状态
updatedAppt, _ := store.GetAppointment(appt.ID)
fmt.Printf("预约状态（录完所有项目后）: %s\n", updatedAppt.Status)

// 6. 生成体检报告
report, err := store.GenerateReport(&checkup.GenerateReportRequest{
    AppointmentID: appt.ID,
})
if err != nil {
    fmt.Printf("生成报告失败: %v\n", err)
    return
}
fmt.Printf("\n========== 体检报告 ==========\n")
fmt.Printf("报告ID: %s\n", report.ID)
fmt.Printf("患者: %s\n", report.PatientName)
fmt.Printf("套餐: %s\n", report.PackageName)
fmt.Printf("检查项目数: %d\n", len(report.Results))
fmt.Printf("异常指标数: %d\n", len(report.AbnormalItems))
fmt.Printf("摘要: %s\n", report.Summary)

if len(report.AbnormalItems) > 0 {
    fmt.Printf("\n---------- 异常指标提醒 ----------\n")
    for i, item := range report.AbnormalItems {
        fmt.Printf("%d. %s: %s %s (参考: %s)\n   备注: %s\n",
            i+1, item.ItemName, item.Value, item.Unit, item.Reference, item.Remarks)
    }
}
```

### 8.3 容量控制示例

```go
// 创建容量=2的时段
slot, _ := store.CreateTimeSlot(&checkup.CreateTimeSlotRequest{
    Date:      date,
    StartTime: start,
    EndTime:   end,
    Capacity:  2,
})

// 患者1预约（成功）
appt1, err := store.CreateAppointment(&checkup.CreateAppointmentRequest{
    PatientID:  patient1ID,
    PackageID:  pkg.ID,
    TimeSlotID: slot.ID,
})
fmt.Printf("患者1预约: %v\n", err == nil) // true

// 患者2预约（成功）
appt2, err := store.CreateAppointment(&checkup.CreateAppointmentRequest{
    PatientID:  patient2ID,
    PackageID:  pkg.ID,
    TimeSlotID: slot.ID,
})
fmt.Printf("患者2预约: %v\n", err == nil) // true

// 患者3预约（失败，容量已满）
appt3, err := store.CreateAppointment(&checkup.CreateAppointmentRequest{
    PatientID:  patient3ID,
    PackageID:  pkg.ID,
    TimeSlotID: slot.ID,
})
if err == checkup.ErrTimeSlotCapacityFull {
    fmt.Println("时段已满，患者3预约被拒绝")
}

// 患者2取消预约，释放容量
store.CancelAppointment(appt2.ID)
fmt.Printf("取消后时段剩余容量: %d\n", slot.Capacity - slot.CurrentCount) // 1

// 患者3再次预约（成功）
appt3, err = store.CreateAppointment(&checkup.CreateAppointmentRequest{
    PatientID:  patient3ID,
    PackageID:  pkg.ID,
    TimeSlotID: slot.ID,
})
fmt.Printf("患者3重新预约: %v\n", err == nil) // true
```

### 8.4 结果录入校验示例

```go
// 尝试录入不在套餐中的项目（失败）
_, err = store.RecordResult(&checkup.RecordResultRequest{
    AppointmentID: appt.ID,
    ItemID:        itemNotInPackage.ID,
    Value:         "xxx",
})
if err == checkup.ErrItemNotInPackage {
    fmt.Println("该项目不在套餐中，无法录入")
}

// 尝试重复录入同一项目（失败）
_, err = store.RecordResult(&checkup.RecordResultRequest{
    AppointmentID: appt.ID,
    ItemID:        wbc.ID, // 之前已录入过
    Value:         "7.0",
})
if err == checkup.ErrDuplicateResult {
    fmt.Println("该项目已录入过，不可重复录入")
}

// 未录完所有项目就生成报告（失败）
_, err = store.GenerateReport(&checkup.GenerateReportRequest{
    AppointmentID: appt.ID,
})
if err == checkup.ErrResultsIncomplete {
    fmt.Println("尚有项目未录入结果，无法生成报告")
}
```

### 8.5 查询功能示例

```go
// 查询单个套餐
pkg, _ := store.GetPackage(pkgID)
fmt.Printf("套餐: %s, 包含%d个项目\n", pkg.Name, len(pkg.Items))

// 查询单个预约
appt, _ := store.GetAppointment(apptID)
fmt.Printf("预约状态: %s, 时段: %s\n", appt.Status, appt.TimeSlotInfo)

// 查询某预约的所有检查结果
results, _ := store.GetResultsByAppointment(apptID)
fmt.Printf("已录入 %d 项结果\n", len(results))
for _, r := range results {
    abnormalFlag := ""
    if r.IsAbnormal {
        abnormalFlag = " ⚠️异常"
    }
    fmt.Printf("  - %s: %s %s%s\n", r.ItemName, r.Value, r.Unit, abnormalFlag)
}

// 查询报告
report, _ := store.GetReport(apptID)
fmt.Printf("报告摘要: %s\n", report.Summary)

// 列出所有套餐
for _, p := range store.ListPackages() {
    fmt.Printf("- %s (¥%.0f)\n", p.Name, p.TotalPrice)
}

// 列出所有时段
for _, s := range store.ListTimeSlots() {
    fmt.Printf("- %s %s-%s (%d/%d)\n",
        s.Date.Format("01-02"),
        s.StartTime.Format("15:04"),
        s.EndTime.Format("15:04"),
        s.CurrentCount, s.Capacity)
}

// 列出所有预约
for _, a := range store.ListAppointments() {
    fmt.Printf("- %s | %s | %s\n", a.PatientName, a.PackageName, a.Status)
}
```

## 9. 测试覆盖说明

模块包含完整的单元测试（49 个测试用例），覆盖以下场景：

### 9.1 正常流程测试
- `TestAddPatient_Success`：添加患者成功
- `TestAddItem_Success`：添加体检项目成功
- `TestCreatePackage_Success`：创建套餐成功（自动计算总价）
- `TestCreateTimeSlot_Success`：创建时段成功
- `TestCreateAppointment_Success`：创建预约成功（时段计数递增）
- `TestCancelAppointment_Success`：取消预约成功（释放容量）
- `TestRecordResult_Success_Normal`：录入正常结果
- `TestRecordResult_Success_Abnormal`：录入异常结果（自动判定）
- `TestRecordResult_AllItemsCompleted`：录完所有项目后状态自动变为COMPLETED
- `TestGenerateReport_Success_NoAbnormal`：生成无异常报告
- `TestGenerateReport_Success_WithAbnormal`：生成含异常提醒的报告
- `TestGetReport_Success`：查询报告成功
- `TestFullWorkflow`：完整流程（配置→预约→录入→报告）

### 9.2 边界条件测试
- `TestCreatePackage_EmptyItems`：空项目列表创建套餐
- `TestCreatePackage_DuplicateItems`：套餐中重复项目校验
- `TestCreatePackage_InvalidItem`：套餐中引用无效项目
- `TestCreateTimeSlot_InvalidTimeRange`：时段开始时间≥结束时间
- `TestCreateTimeSlot_InvalidCapacity`：容量≤0
- `TestCreateAppointment_CapacityFull`：时段容量已满
- `TestRecordResult_DuplicateResult`：同一项目重复录入
- `TestRecordResult_BelowMin`：结果低于参考下限（异常判定）
- `TestGenerateReport_ResultsIncomplete`：结果不完整时生成报告
- `TestConcurrentAppointments`：并发预约的容量控制（5人抢3个名额）
- `TestCancelAppointment_Idempotent`：重复取消预约（幂等性）
- `TestGenerateReport_Idempotent`：重复生成报告（幂等性）
- `TestBoundaryCapacityOne`：容量=1的边界测试

### 9.3 异常分支测试
- 所有实体的不存在场景（患者/项目/套餐/时段/预约/结果/报告）
- 套餐引用无效项目
- 录入不在套餐中的项目结果
- 对已取消预约录入结果/生成报告
- 各种查询操作的不存在场景
- 列表功能（ListItems/ListPackages/ListTimeSlots/ListAppointments）

## 10. 运行测试

```bash
cd solocoder-4-go
go test ./internal/checkup/ -v
```

预期输出：所有 49 个测试用例 PASS。
