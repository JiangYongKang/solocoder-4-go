# 电子处方流转模块需求文档

## 1. 模块概述

电子处方流转模块实现了从医生开具处方、药房审核、配药发药到处方完成的完整业务流程，同时支持处方撤回功能，并与药品库存进行联动管理，确保处方流转过程中药品库存的准确性和安全性。

### 主要功能特性

- **处方开具**：医生为患者创建包含药品明细和用量说明的处方
- **药房审核**：药房对处方进行审核，支持通过或驳回操作
- **配药发药**：审核通过的处方完成配药并发药
- **处方撤回**：支持在特定状态下撤回处方
- **库存联动**：处方流转各环节与库存占用/释放联动
- **并发安全**：基于读写锁实现并发场景下的数据一致性

## 2. 核心结构体职责

### 2.1 基础实体

#### Doctor（医生）
- **职责**：存储医生基本信息，用于处方开具时的身份识别
- **关键字段**：
  - `ID`：医生唯一标识
  - `Name`：医生姓名
  - `Title`：职称（如主任医师、副主任医师等）
  - `Department`：所属科室
  - `LicenseNo`：执业证书编号

#### Patient（患者）
- **职责**：存储患者基本信息，用于处方的归属标识
- **关键字段**：
  - `ID`：患者唯一标识
  - `Name`：患者姓名
  - `IDCard`：身份证号
  - `Phone`：联系电话
  - `Gender`：性别
  - `Age`：年龄
  - `MedicalRecordNo`：病历号

#### Pharmacy（药房）
- **职责**：存储药房信息，用于处方审核和发药的机构标识
- **关键字段**：
  - `ID`：药房唯一标识
  - `Name`：药房名称
  - `Address`：药房地址
  - `Phone`：联系电话
  - `LicenseNo`：经营许可证编号

#### Medicine（药品）
- **职责**：存储药品基础信息
- **关键字段**：
  - `ID`：药品唯一标识
  - `Name`：药品商品名
  - `GenericName`：通用名
  - `Specification`：规格（如 0.25g*24粒）
  - `Manufacturer`：生产厂家
  - `UnitPrice`：单价

### 2.2 库存相关

#### InventoryItem（库存条目）
- **职责**：管理单个药品的库存数量和预留数量
- **关键字段**：
  - `MedicineID`：关联药品ID
  - `Medicine`：药品引用
  - `Quantity`：实际库存数量
  - `Reserved`：已被处方占用的预留数量
  - `LastUpdated`：最后更新时间
- **关键逻辑**：
  - 可用库存 = Quantity - Reserved
  - 处方创建时增加 Reserved
  - 处方发药时同时减少 Quantity 和 Reserved
  - 处方撤回/驳回时减少 Reserved

### 2.3 处方相关

#### MedicineItem（处方药品明细项）
- **职责**：描述处方中单个药品的详细信息和使用说明
- **关键字段**：
  - `MedicineID`：药品ID
  - `MedicineName`：药品名称（冗余存储方便查询）
  - `Quantity`：开具数量
  - `UnitPrice`：开方时的单价
  - `Dosage`：每次用量（如"每次2粒"）
  - `Frequency`：用药频率（如"每日3次"）
  - `Duration`：用药时长（如"7天"）
  - `Instructions`：特殊说明（如"饭后服用"）

#### Prescription（处方）
- **职责**：作为核心业务实体，承载处方的完整信息和状态流转
- **关键字段**：
  - `ID`：处方唯一标识
  - `DoctorID/DoctorName`：开方医生信息
  - `PatientID/PatientName`：患者信息
  - `PharmacyID/PharmacyName`：目标药房信息
  - `Items`：药品明细列表
  - `TotalAmount`：处方总金额
  - `Status`：当前状态
  - `RejectReason`：审核驳回原因
  - `Diagnosis`：诊断信息
  - `Remark`：备注
  - `ReservedStock`：本处方占用的库存记录（map[药品ID]数量）
  - `CreatedAt`：创建时间
  - `ReviewedAt/ReviewedBy`：审核时间和审核人
  - `DispensedAt/DispensedBy`：发药时间和发药人
  - `WithdrawnAt/WithdrawnReason`：撤回时间和原因

### 2.4 核心服务

#### Store（内存数据存储）
- **职责**：基于内存的数据存储层，提供所有实体的CRUD操作和库存管理
- **核心能力**：
  - 读写锁（sync.RWMutex）保证并发安全
  - 医生/患者/药房/药品的增删查
  - 库存的增加、查询、预留、释放、扣减
  - 处方的存取和按维度查询
  - 自动ID生成（带前缀：DOC/PAT/PHY/MED/PRE）

#### Service（业务服务层）
- **职责**：封装处方流转的核心业务逻辑，对外提供统一的操作接口
- **核心方法**：
  - `CreatePrescription`：开具处方
  - `ReviewPrescription`：审核处方
  - `DispensePrescription`：配药发药
  - `WithdrawPrescription`：撤回处方
  - `GetPrescription`：查询单个处方
  - `ListPrescriptionsByPatient/Pharmacy/Doctor`：按维度列出处方

## 3. 处方状态流转

### 3.1 状态定义

| 状态 | 常量标识 | 说明 |
|------|----------|------|
| 待审核 | `StatusPendingReview` | 处方刚创建，等待药房审核 |
| 审核通过 | `StatusApproved` | 药房审核通过，等待发药 |
| 已驳回 | `StatusRejected` | 药房审核驳回，流程终止 |
| 已完成 | `StatusCompleted` | 药品已发放，流程完成 |
| 已撤回 | `StatusWithdrawn` | 处方被撤回，流程终止 |

### 3.2 状态流转图

```
                     ┌─────────────────┐
                     │  PENDING_REVIEW │◄─────────┐
                     │    (待审核)     │          │
                     └────────┬────────┘          │
                              │                   │
                     审核通过 │                   │ 撤回
                              ▼                   │
                     ┌─────────────────┐          │
                     │   APPROVED      │──────────┘
                     │   (审核通过)    │
                     └────────┬────────┘
                              │
                        发药  │  或  撤回
                              ▼
              ┌───────────────┴───────────────┐
              │                               │
              ▼                               ▼
    ┌─────────────────┐             ┌─────────────────┐
    │   COMPLETED     │             │   WITHDRAWN     │
    │    (已完成)     │             │    (已撤回)     │
    └─────────────────┘             └─────────────────┘
    
    另外：从 PENDING_REVIEW 审核驳回 → REJECTED (已驳回)
```

### 3.3 流转规则详解

1. **创建处方 → PENDING_REVIEW**
   - 触发操作：`CreatePrescription`
   - 前置条件：医生、患者、药房存在；所有药品库存充足；数量合法
   - 副作用：占用所有药品对应数量的库存（增加 Reserved）

2. **PENDING_REVIEW → APPROVED（审核通过）**
   - 触发操作：`ReviewPrescription(Approved=true)`
   - 前置条件：处方处于 PENDING_REVIEW 状态；审核药房与处方药房一致
   - 副作用：记录审核时间和审核人；库存保持占用

3. **PENDING_REVIEW → REJECTED（审核驳回）**
   - 触发操作：`ReviewPrescription(Approved=false)`
   - 前置条件：处方处于 PENDING_REVIEW 状态；审核药房与处方药房一致
   - 副作用：记录驳回原因；释放所有已占用的库存（减少 Reserved）；流程终止

4. **APPROVED → COMPLETED（配药发药）**
   - 触发操作：`DispensePrescription`
   - 前置条件：处方处于 APPROVED 状态；未重复发药；药房一致
   - 副作用：记录发药时间和发药人；扣减实际库存（同时减少 Quantity 和 Reserved）；流程完成

5. **PENDING_REVIEW → WITHDRAWN（待审核时撤回）**
   - 触发操作：`WithdrawPrescription`
   - 前置条件：处方处于 PENDING_REVIEW 或 APPROVED 状态
   - 副作用：记录撤回原因；释放所有已占用的库存；流程终止

6. **APPROVED → WITHDRAWN（审核通过后撤回）**
   - 触发操作：`WithdrawPrescription`
   - 前置条件：同上
   - 副作用：同上

### 3.4 禁止的流转

以下状态转换不允许，将返回错误：
- 已完成(COMPLETED) → 任何其他状态
- 已驳回(REJECTED) → 审核通过(APPROVED)或撤回(WITHDRAWN)
- 已撤回(WITHDRAWN) → 任何其他状态
- 待审核(PENDING_REVIEW) → 直接发药(跳过审核)
- 重复发药（防止多次发药）

## 4. 库存联动机制

### 4.1 库存状态说明

库存采用"实际数量 + 预留数量"双字段设计：
- `Quantity`：药房实际拥有的药品数量
- `Reserved`：被处方占用但尚未发药的数量
- `Available = Quantity - Reserved`：当前可用于新开处方的数量

### 4.2 库存操作矩阵

| 业务操作 | Quantity 变化 | Reserved 变化 | 说明 |
|---------|---------------|---------------|------|
| `UpdateInventory(+n)` | +n | 0 | 入库，增加实际库存 |
| 创建处方 | 0 | +n | 占用库存，预留不扣减 |
| 审核通过 | 0 | 0 | 维持占用状态 |
| 审核驳回 | 0 | -n | 释放被占用的库存 |
| 配药发药 | -n | -n | 实际扣减库存，同时清掉预留 |
| 处方撤回 | 0 | -n | 释放被占用的库存 |

### 4.3 库存校验点

1. **处方创建时**：逐个检查药品的可用库存(Quantity-Reserved)是否满足
2. **处方发药时**：检查预留库存(Reserved)是否足够（防止并发问题）
3. **入库时**：检查是否导致 Quantity 为负

### 4.4 并发安全

所有库存操作都在 `Store.mu` 读写锁保护下执行：
- 读操作（GetInventory, GetAvailableStock）使用读锁
- 写操作（创建/审核/发药/撤回）使用写锁
- 整个处方创建流程在单个写锁内完成，保证库存占用的原子性

## 5. 错误处理

### 5.1 错误类型定义

| 错误变量 | 说明 | 触发场景 |
|---------|------|---------|
| `ErrDoctorNotFound` | 医生不存在 | 开方时医生ID无效 |
| `ErrPatientNotFound` | 患者不存在 | 开方时患者ID无效 |
| `ErrPharmacyNotFound` | 药房不存在 | 开方/审核/发药时药房无效 |
| `ErrMedicineNotFound` | 药品不存在 | 开方时药品ID无效 |
| `ErrPrescriptionNotFound` | 处方不存在 | 查询/审核/发药/撤回时处方ID无效 |
| `ErrInsufficientStock` | 库存不足 | 开方时可用库存不够 |
| `ErrInvalidStatus` | 状态非法 | 在不允许的状态下执行操作（如重复审核） |
| `ErrAlreadyDispensed` | 已发药 | 对已完成的处方再次发药 |
| `ErrNoItems` | 无药品明细 | 开方时空药品列表 |
| `ErrInvalidQuantity` | 数量非法 | 药品数量 ≤ 0 |
| `ErrNotApproved` | 未审核通过 | 未审核或被驳回的处方尝试发药 |
| `ErrCannotWithdraw` | 不可撤回 | 已完成/已驳回的处方尝试撤回 |

## 6. 使用示例

### 6.1 初始化环境

```go
package main

import (
    "fmt"
    "solocoder-4-go/internal/prescription"
)

func main() {
    store := prescription.NewStore()
    svc := prescription.NewService(store)

    // 添加医生
    doctorID := store.AddDoctor(&prescription.Doctor{
        Name:       "张医生",
        Title:      "主任医师",
        Department: "内科",
        LicenseNo:  "DOC001",
    })

    // 添加患者
    patientID := store.AddPatient(&prescription.Patient{
        Name:   "李患者",
        IDCard: "110101199001011234",
        Phone:  "13800138000",
        Age:    35,
    })

    // 添加药房
    pharmacyID := store.AddPharmacy(&prescription.Pharmacy{
        Name:      "和平药房",
        Address:   "北京市朝阳区和平路1号",
        LicenseNo: "PHY001",
    })

    // 添加药品及库存
    med1ID := store.AddMedicine(&prescription.Medicine{
        Name:          "阿莫西林胶囊",
        Specification: "0.25g*24粒",
        Manufacturer:  "华北制药",
        UnitPrice:     25.50,
    })
    store.UpdateInventory(med1ID, 100)

    med2ID := store.AddMedicine(&prescription.Medicine{
        Name:          "布洛芬缓释胶囊",
        Specification: "0.3g*20粒",
        Manufacturer:  "中美史克",
        UnitPrice:     32.00,
    })
    store.UpdateInventory(med2ID, 50)
}
```

### 6.2 完整流程：开具 → 审核 → 发药

```go
// 1. 医生开具处方
prescription, err := svc.CreatePrescription(&prescription.CreatePrescriptionRequest{
    DoctorID:   doctorID,
    PatientID:  patientID,
    PharmacyID: pharmacyID,
    Items: []prescription.MedicineItemRequest{
        {
            MedicineID:   med1ID,
            Quantity:     2,
            Dosage:       "每次2粒",
            Frequency:    "每日3次",
            Duration:     "7天",
            Instructions: "饭后服用",
        },
        {
            MedicineID:   med2ID,
            Quantity:     1,
            Dosage:       "每次1粒",
            Frequency:    "每日2次",
            Duration:     "3天",
            Instructions: "必要时服用",
        },
    },
    Diagnosis: "上呼吸道感染",
    Remark:    "注意休息，多饮水",
})
if err != nil {
    fmt.Printf("开具处方失败: %v\n", err)
    return
}
fmt.Printf("处方创建成功，ID: %s，状态: %s，总金额: %.2f\n", 
    prescription.ID, prescription.Status, prescription.TotalAmount)

// 2. 药房审核通过
reviewed, err := svc.ReviewPrescription(&prescription.ReviewRequest{
    PrescriptionID: prescription.ID,
    PharmacyID:     pharmacyID,
    Approved:       true,
    ReviewedBy:     "王药师",
})
if err != nil {
    fmt.Printf("审核失败: %v\n", err)
    return
}
fmt.Printf("审核完成，状态: %s\n", reviewed.Status)

// 3. 配药发药
completed, err := svc.DispensePrescription(&prescription.DispenseRequest{
    PrescriptionID: prescription.ID,
    PharmacyID:     pharmacyID,
    DispensedBy:    "张发药员",
})
if err != nil {
    fmt.Printf("发药失败: %v\n", err)
    return
}
fmt.Printf("发药完成，状态: %s，发药人: %s\n", 
    completed.Status, completed.DispensedBy)
```

### 6.3 审核驳回示例

```go
// 药房审核驳回
rejected, err := svc.ReviewPrescription(&prescription.ReviewRequest{
    PrescriptionID: prescription.ID,
    PharmacyID:     pharmacyID,
    Approved:       false,
    RejectReason:   "药品剂量超标，请医生确认",
    ReviewedBy:     "李药师",
})
if err != nil {
    fmt.Printf("审核失败: %v\n", err)
    return
}
fmt.Printf("已驳回，原因: %s\n", rejected.RejectReason)
// 此时库存已自动释放
```

### 6.4 撤回处方示例

```go
// 在待审核或审核通过状态下撤回
withdrawn, err := svc.WithdrawPrescription(&prescription.WithdrawRequest{
    PrescriptionID: prescription.ID,
    Reason:         "患者放弃取药",
})
if err != nil {
    fmt.Printf("撤回失败: %v\n", err)
    return
}
fmt.Printf("已撤回，状态: %s，原因: %s\n", 
    withdrawn.Status, withdrawn.WithdrawnReason)
// 此时库存已自动释放
```

### 6.5 查询处方列表

```go
// 按患者查询
patientPrescriptions, _ := svc.ListPrescriptionsByPatient(patientID)
fmt.Printf("患者共有 %d 张处方\n", len(patientPrescriptions))

// 按药房查询
pharmacyPrescriptions, _ := svc.ListPrescriptionsByPharmacy(pharmacyID)
fmt.Printf("药房共有 %d 张处方\n", len(pharmacyPrescriptions))

// 按医生查询
doctorPrescriptions, _ := svc.ListPrescriptionsByDoctor(doctorID)
fmt.Printf("医生共开具 %d 张处方\n", len(doctorPrescriptions))
```

### 6.6 查看库存状态

```go
// 查看某药品库存
inv, err := store.GetInventory(med1ID)
if err == nil {
    available := inv.Quantity - inv.Reserved
    fmt.Printf("药品: %s，总库存: %d，已占用: %d，可用: %d\n",
        inv.Medicine.Name, inv.Quantity, inv.Reserved, available)
}
```

## 7. 测试覆盖说明

模块包含完整的单元测试（33 个测试用例），覆盖以下场景：

### 7.1 正常流程测试
- `TestCreatePrescription_Success`：处方创建成功
- `TestReviewPrescription_Approved`：审核通过
- `TestReviewPrescription_Rejected`：审核驳回
- `TestDispensePrescription_Success`：配药发药成功
- `TestWithdrawPrescription_FromPending`：待审核状态撤回
- `TestWithdrawPrescription_FromApproved`：审核通过后撤回
- `TestFullWorkflow`：开方→审核→发药完整流程

### 7.2 边界条件测试
- `TestCreatePrescription_InsufficientStock`：库存不足
- `TestCreatePrescription_NoItems`：空药品列表
- `TestCreatePrescription_InvalidQuantity`：数量 ≤ 0
- `TestDispensePrescription_Duplicate`：重复发药
- `TestConcurrentPrescriptionStock`：并发开方库存安全
- `TestUpdateInventory`：入库导致负数库存
- `TestPrescriptionStockPartialFailure`：部分药品不足时的原子性

### 7.3 异常分支测试
- 所有实体的不存在场景（医生/患者/药房/药品/处方）
- 审核时药房不匹配
- 重复审核（错误状态）
- 未审核处方尝试发药
- 已驳回处方尝试发药
- 已完成处方尝试撤回
- 已驳回处方尝试撤回
- 驳回时未提供原因
- 撤回时未提供原因

## 8. 运行测试

```bash
cd solocoder-4-go
go test ./internal/prescription/ -v
```

预期输出：所有 33 个测试用例 PASS。
