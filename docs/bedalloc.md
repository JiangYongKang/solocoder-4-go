# 住院床位调度模块

## 1. 模块概述

住院床位调度模块（bedalloc）是一个基于内存数据结构的医院床位管理系统，支持病区床位分配、转科换床、床位占用冲突检测、出院释放和床位利用率统计等核心功能。模块使用 Go 语言实现，提供线程安全的并发访问支持。

## 2. 模块功能

### 2.1 病区床位分配
- 支持按指定病区、床位类型分配可用床位
- 根据患者年龄、病情等条件自动匹配合适的床位类型
  - 14岁以下儿童自动匹配儿科床位
  - ICU 患者自动匹配 ICU 床位
  - 传染病患者自动匹配隔离床位
  - 手术患者优先匹配手术床位
- 分配成功后床位状态自动转为占用
- 按楼层和房间号排序，优先分配低楼层床位

### 2.2 转科换床
- 支持患者在不同病区间转移
- 支持同一病区内更换床位
- 整个过程为原子操作，确保不会产生中间冲突
- 自动释放原床位，占用新床位
- 创建新的住院记录，原有记录标记为结束

### 2.3 床位占用冲突检测
- 同一床位不能同时分配给多个患者
- 同一患者不能同时有多个有效住院记录
- 重复入院返回明确错误
- 转科时目标床位已被占用返回错误
- 禁止转移到当前已占用的同一床位

### 2.4 出院释放与床位利用率统计
- 患者出院后自动释放床位，进入清洁状态
- 床位清洁完成后恢复可用状态
- 支持按病区统计指定时间范围内的床位利用率
- 正确处理跨统计周期的住院记录
- 支持仍在住院患者的统计计算

## 3. 核心结构体职责

### 3.1 BedAllocator
**职责**：床位调度核心管理器，提供所有对外接口。

**主要字段**：
- `wards`：病区映射表
- `patients`：患者信息映射表
- `admissions`：住院记录映射表
- `patientAdmitMap`：患者当前有效住院记录映射
- `mu`：读写锁，保证并发安全

### 3.2 Ward（病区）
**职责**：表示一个病区，管理该病区内的所有床位。

**字段**：
- `ID`：病区唯一标识
- `Name`：病区名称
- `Department`：所属科室
- `Beds`：该病区的床位映射表

### 3.3 Bed（床位）
**职责**：表示一个具体的床位，包含床位属性和状态。

**字段**：
- `ID`：床位唯一标识
- `WardID`：所属病区ID
- `Type`：床位类型（普通、ICU、手术、儿科、隔离）
- `Status`：床位状态（可用、占用、清洁、维护）
- `RoomNumber`：房间号
- `Floor`：楼层

### 3.4 Patient（患者）
**职责**：存储患者基本信息。

**字段**：
- `ID`：患者唯一标识
- `Name`：患者姓名
- `Age`：年龄
- `Gender`：性别
- `Condition`：病情描述

### 3.5 AdmissionRecord（住院记录）
**职责**：记录一次完整的住院过程。

**字段**：
- `ID`：记录唯一标识
- `PatientID`：患者ID
- `WardID`：病区ID
- `BedID`：床位ID
- `AdmitTime`：入院时间
- `DischargeTime`：出院时间（可为空表示仍在住院）
- `Active`：是否为当前有效记录

## 4. 床位状态流转

```
Available（可用）
    |
    | 分配床位
    v
Occupied（占用）
    |
    | 转科换床 / 出院
    v
Available（可用）  <----+
    |                  |
    | 出院              | 清洁完成
    v                  |
Cleaning（清洁中） -----+
    |
    | 床位维护
    v
Maintenance（维护中）
```

### 状态说明：
- **Available**：床位空闲可分配
- **Occupied**：床位被患者占用
- **Cleaning**：患者出院后等待清洁消毒
- **Maintenance**：床位处于维修保养状态

## 5. 错误定义

| 错误变量 | 错误说明 |
|---------|---------|
| `ErrWardNotFound` | 病区不存在 |
| `ErrBedNotFound` | 床位不存在 |
| `ErrBedOccupied` | 床位已被占用 |
| `ErrBedNotOccupied` | 床位未被占用 |
| `ErrPatientNotFound` | 患者不存在 |
| `ErrPatientAlreadyAdmitted` | 患者已在住院中 |
| `ErrNoAvailableBed` | 没有符合条件的可用床位 |
| `ErrTransferSameBed` | 不能转移到同一床位 |
| `ErrInvalidTimeRange` | 无效的时间范围 |
| `ErrBedTypeMismatch` | 床位类型不匹配 |

## 6. 使用示例

### 6.1 初始化调度器

```go
ba := bedalloc.NewBedAllocator()
```

### 6.2 创建病区和床位

```go
ba.AddWard("W001", "内科病区", "内科")
ba.AddBed("W001", "B001", bedalloc.BedTypeGeneral, "101", 1)
ba.AddBed("W001", "B002", bedalloc.BedTypeICU, "102", 1)
```

### 6.3 添加患者

```go
ba.AddPatient("P001", "张三", 45, "男", "general")
```

### 6.4 分配床位

```go
now := time.Now()
admission, err := ba.AllocateBed(bedalloc.AllocateCriteria{
    WardID:           "W001",
    BedType:          bedalloc.BedTypeGeneral,
    PatientID:        "P001",
    PatientAge:       45,
    PatientCondition: "general",
    AdmitTime:        now,
})
```

### 6.5 转科换床

```go
transferTime := now.Add(24 * time.Hour)
newAdmission, err := ba.TransferBed(bedalloc.TransferCriteria{
    PatientID:     "P001",
    TargetWardID:  "W002",
    TargetBedType: bedalloc.BedTypeSurgery,
    TransferTime:  transferTime,
})
```

### 6.6 患者出院

```go
dischargeTime := transferTime.Add(72 * time.Hour)
discharged, err := ba.DischargePatient("P001", dischargeTime)
```

### 6.7 标记床位清洁完成

```go
err := ba.MarkBedCleaned("B001")
```

### 6.8 统计床位利用率

```go
startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
endDate := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)
report, err := ba.CalculateUtilization("W001", startDate, endDate)

fmt.Printf("床位利用率: %.2f%%\n", report.UtilizationRate*100)
fmt.Printf("占用天数: %.1f\n", report.OccupiedDays)
fmt.Printf("可用天数: %.1f\n", report.AvailableDays)
```

### 6.9 查询住院历史

```go
history := ba.GetAdmissionHistory("P001")
for _, record := range history {
    fmt.Printf("入院时间: %v, 床位: %s\n", record.AdmitTime, record.BedID)
}
```

## 7. 核心接口列表

| 方法 | 功能说明 |
|-----|---------|
| `NewBedAllocator()` | 创建新的床位调度器 |
| `AddWard()` | 添加病区 |
| `AddBed()` | 添加床位 |
| `AddPatient()` | 添加患者 |
| `AllocateBed()` | 分配床位 |
| `TransferBed()` | 转科换床 |
| `DischargePatient()` | 患者出院 |
| `MarkBedCleaned()` | 标记床位清洁完成 |
| `CalculateUtilization()` | 计算床位利用率 |
| `GetActiveAdmission()` | 获取患者当前住院记录 |
| `GetAdmissionHistory()` | 获取患者住院历史 |
| `ListActiveAdmissions()` | 列出所有有效住院记录 |
| `GetWard()` | 获取病区信息 |
| `GetBed()` | 获取床位信息 |
| `GetPatient()` | 获取患者信息 |
