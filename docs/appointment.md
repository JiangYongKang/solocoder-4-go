# 门诊预约管理模块需求文档

## 1. 模块概述

门诊预约管理模块实现了门诊预约全流程的核心业务功能，包括医生排班管理、号源查询与锁定、预约创建与确认、患者改约与取消、爽约记录，以及基于候补队列的号源自动补位机制。模块基于内存数据结构，提供并发安全的操作接口，支持高并发场景下的号源一致性保证。

### 主要功能特性

- **排班管理**：支持为医生配置每日多时段号源排班
- **排班查询**：支持按医生、科室、日期多维度组合查询排班，并区分剩余号源与已满号源
- **号源锁定**：预约前锁定号源，防止并发重复占用，支持超时自动释放
- **预约创建**：锁定号源后确认生成正式预约记录
- **患者改约**：在规则允许范围内将预约改签到其他号源
- **预约取消**：规则允许范围内取消预约，号源重新释放
- **爽约记录**：患者未按时就诊时记录爽约，号源释放
- **候补队列**：号源满时支持加入候补队列，号源释放后按入队顺序自动补位
- **并发安全**：基于读写锁实现并发场景下的数据一致性

## 2. 核心结构体职责

### 2.1 基础实体

#### Doctor（医生）
- **职责**：存储医生基本信息，用于排班归属标识
- **关键字段**：
  - `ID`：医生唯一标识
  - `Name`：医生姓名
  - `Title`：职称（如主任医师、副主任医师等）
  - `Department`：所属科室
  - `LicenseNo`：执业证书编号

#### Patient（患者）
- **职责**：存储患者基本信息，用于预约归属标识
- **关键字段**：
  - `ID`：患者唯一标识
  - `Name`：患者姓名
  - `IDCard`：身份证号
  - `Phone`：联系电话
  - `Gender`：性别
  - `Age`：年龄
  - `MedicalRecordNo`：病历号

### 2.2 排班与号源

#### ScheduleSlot（号源时段）
- **职责**：描述医生某一天中单个时段的号源信息，是预约的最小单位
- **关键字段**：
  - `ID`：号源唯一标识
  - `DoctorID/DoctorName`：所属医生
  - `Department`：所属科室
  - `Date`：排班日期（YYYY-MM-DD）
  - `StartTime/EndTime`：时段起止时间（HH:MM）
  - `TotalCapacity`：该时段总号源数
  - `BookedCount`：已预约数
  - `Status`：号源状态（AVAILABLE/LOCKED/BOOKED 等）
  - `LockedBy`：锁定该号源的患者ID
  - `LockedAt/LockExpireAt`：锁定时间与超时时间
- **关键逻辑**：
  - 剩余号源 = TotalCapacity - BookedCount
  - 号源被锁定时，BookedCount 不变，但锁定期间其他患者无法再锁定
  - 确认预约时 BookedCount+1，锁定信息清除
  - 取消/爽约/改约时 BookedCount-1

#### Schedule（排班）
- **职责**：按天组织的医生排班，包含多个号源时段
- **关键字段**：
  - `ID`：排班唯一标识
  - `DoctorID/DoctorName`：所属医生
  - `Department`：所属科室
  - `Date`：排班日期
  - `Slots`：该天的号源时段列表

### 2.3 预约相关

#### Appointment（预约记录）
- **职责**：作为核心业务实体，承载单次预约的完整信息和状态流转
- **关键字段**：
  - `ID`：预约唯一标识
  - `PatientID/PatientName`：患者信息
  - `SlotID`：关联号源ID
  - `DoctorID/DoctorName`：医生信息
  - `Department`：科室
  - `Date/StartTime/EndTime`：预约时间（冗余存储方便查询）
  - `Status`：预约状态
  - `IsNoShow`：是否爽约
  - `CreatedAt`：创建时间
  - `ConfirmedAt`：确认时间
  - `CancelledAt`：取消时间
  - `CompletedAt`：完成时间
  - `ChangedFromID/ChangedToID`：改约关联（从哪个预约改来/改到哪个预约）

#### WaitQueueItem（候补队列项）
- **职责**：表示一个候补队列中的等待记录
- **关键字段**：
  - `ID`：候补项唯一标识
  - `PatientID/PatientName`：等待患者
  - `TargetDate`：目标日期
  - `DoctorID`：目标医生（可为空，表示按科室候诊）
  - `Department`：目标科室
  - `JoinedAt`：加入候补的时间（用于排序，先入先出）

#### NoShowRecord（爽约记录）
- **职责**：记录患者爽约信息，用于统计和后续管理
- **关键字段**：
  - `ID`：记录唯一标识
  - `PatientID/PatientName`：爽约患者
  - `AppointmentID`：关联的预约ID
  - `DoctorID`：医生
  - `Date/StartTime`：爽约的预约时间
  - `RecordedAt`：记录时间
  - `Remark`：爽约原因/备注

### 2.4 核心服务

#### Store（内存数据存储）
- **职责**：基于内存的数据存储层，提供所有实体的CRUD操作和业务数据管理
- **核心能力**：
  - 读写锁（sync.RWMutex）保证并发安全
  - 医生/患者的增删查
  - 排班/号源的增删查
  - 预约记录的存取和按维度查询
  - 候补队列的维护（按医生+日期 / 科室+日期分桶）
  - 爽约记录管理
  - 自动ID生成（带前缀：DOC/PAT/SCH/SLT/APT/WQ/NOS）
  - 可配置的超时规则：
    - `LockTimeout`：号源锁定超时（默认5分钟）
    - `ChangeDeadline`：改约截止时间（预约前24小时）
    - `CancelDeadline`：取消截止时间（预约前12小时）
  - `NowFunc`：可注入的时间函数，方便测试

#### Service（业务服务层）
- **职责**：封装预约管理的核心业务逻辑，对外提供统一的操作接口
- **核心方法**：
  - `AddSchedule`：为医生添加排班
  - `QuerySchedules`：多维度查询排班
  - `LockSlot`：锁定号源
  - `ConfirmAppointment`：确认锁定并生成预约
  - `ReleaseLock`：主动释放号源锁定
  - `CheckAndReleaseExpiredLocks`：批量释放超时锁定
  - `CreateAppointment`：一键创建预约（锁定+确认）
  - `ChangeAppointment`：改约
  - `CancelAppointment`：取消预约
  - `RecordNoShow`：记录爽约
  - `GetNoShowRecords`：查询爽约记录
  - `JoinWaitQueue`：加入候补队列
  - `GetWaitQueue`：查询候补队列
  - `GetAppointment`：查询单个预约
  - `ListAppointmentsByPatient`：按患者列出预约
  - `GetSlot`：查询单个号源状态

## 3. 号源状态流转

### 3.1 号源状态定义

| 状态 | 常量标识 | 说明 |
|------|----------|------|
| 可预约 | `SlotStatusAvailable` | 号源可被锁定/预约 |
| 锁定中 | `SlotStatusLocked` | 号源被患者锁定，等待确认 |
| 已满 | `SlotStatusBooked` | 号源全部约满（BookedCount >= TotalCapacity） |
| 已取消 | `SlotStatusCancelled` | 预留状态（当前版本未使用） |
| 爽约释放 | `SlotStatusNoShow` | 预留状态（当前版本未使用） |
| 已完成 | `SlotStatusCompleted` | 预留状态（当前版本未使用） |

### 3.2 号源状态流转图

```
                     ┌─────────────────┐
                     │   AVAILABLE     │◄───────────────────────────┐
                     │    (可预约)     │                            │
                     └────────┬────────┘                            │
                              │                                     │
                     锁定号源 │                                     │
                              ▼                                     │
                     ┌─────────────────┐                            │
                     │    LOCKED       │  锁定超时/主动释放         │
                     │    (锁定中)     ├────────────────────────────┘
                     └────────┬────────┘
                              │
                     确认预约 │   （BookedCount++）
                              ▼
          ┌───────────────────┴───────────────────┐
          │                                       │
 BookedCount < Capacity                  BookedCount >= Capacity
          │                                       │
          ▼                                       ▼
   ┌───────────────┐                      ┌───────────────┐
   │   AVAILABLE   │                      │   BOOKED      │
   │  (还有剩余)   │                      │    (已满)     │
   └───────┬───────┘                      └───────┬───────┘
           │                                       │
           │ 取消预约/爽约/改约 (BookedCount--)      │ 取消预约/爽约/改约
           │                                       │
           └───────────────────────────────────────┘
```

### 3.3 号源与候补队列的联动

当号源因取消/爽约/改约而释放时（BookedCount减少），系统自动触发候补队列补位流程：
1. 收集针对该号源的候补候选（按医生+日期 和 科室+日期 两个维度匹配）
2. 按 JoinedAt 时间升序排序（先入先出）
3. 依次为候选人生成预约，直到号源再次满或候选耗尽
4. 已成功补位的候选人从候补队列中移除

## 4. 预约状态流转

### 4.1 预约状态定义

| 状态 | 常量标识 | 说明 |
|------|----------|------|
| 待确认 | `AppointmentStatusPending` | 预留状态（当前版本锁定后直接确认） |
| 已确认 | `AppointmentStatusConfirmed` | 预约已确认，正常待就诊 |
| 已取消 | `AppointmentStatusCancelled` | 患者主动取消预约 |
| 已爽约 | `AppointmentStatusNoShow` | 患者未按时就诊 |
| 已完成 | `AppointmentStatusCompleted` | 预留状态（患者已就诊完成） |
| 已改约 | `AppointmentStatusChanged` | 预约已改签到其他号源 |

### 4.2 预约状态流转图

```
                    ┌─────────────────┐
                    │   PENDING       │
                    │   (待确认)      │
                    └────────┬────────┘
                             │ 确认预约
                             ▼
                    ┌─────────────────┐
          ┌─────────│   CONFIRMED     │──────────┐
          │         │    (已确认)     │          │
          │         └───────┬─────────┘          │
          │                 │                    │
          │ 规则内改约       │ 规则内取消         │ 爽约
          │                 │                    │
          ▼                 ▼                    ▼
  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
  │   CHANGED     │  │  CANCELLED    │  │   NO_SHOW     │
  │   (已改约)    │  │   (已取消)    │  │   (已爽约)    │
  └───────────────┘  └───────────────┘  └───────────────┘

  改约后：ChangedToID 指向新预约
  新预约：ChangedFromID 指向原预约
```

### 4.3 流转规则详解

1. **创建预约 → CONFIRMED**
   - 触发操作：`CreateAppointment`（或 `LockSlot` + `ConfirmAppointment`）
   - 前置条件：患者存在；号源存在；号源有剩余容量；号源未被其他患者锁定
   - 副作用：号源 BookedCount++；清除号源锁定信息

2. **CONFIRMED → CHANGED（改约）**
   - 触发操作：`ChangeAppointment`
   - 前置条件：预约处于 CONFIRMED 状态；当前时间距预约时间 >= ChangeDeadline（默认24小时）；新号源有容量
   - 副作用：
     - 原预约状态改为 CHANGED，记录 ChangedToID
     - 新预约创建，状态为 CONFIRMED，记录 ChangedFromID
     - 原号源 BookedCount--
     - 新号源 BookedCount++
     - 原号源释放后自动触发候补补位

3. **CONFIRMED → CANCELLED（取消）**
   - 触发操作：`CancelAppointment`
   - 前置条件：预约处于 CONFIRMED 状态；当前时间距预约时间 >= CancelDeadline（默认12小时）
   - 副作用：号源 BookedCount--；号源释放后自动触发候补补位

4. **CONFIRMED → NO_SHOW（爽约）**
   - 触发操作：`RecordNoShow`
   - 前置条件：预约处于 CONFIRMED 状态；必须提供爽约备注
   - 副作用：
     - 生成 NoShowRecord 记录
     - 预约状态改为 NO_SHOW，IsNoShow=true
     - 号源 BookedCount--
     - 号源释放后自动触发候补补位

### 4.4 禁止的流转

以下状态转换不允许，将返回错误：
- 已取消/已爽约/已改约 → 任何其他状态（终端状态不可逆转）
- 在截止时间之后尝试改约或取消
- 尝试改约到同一个号源
- 非本人尝试操作他人的预约

## 5. 号源锁定机制

### 5.1 锁定目的

号源锁定用于解决并发预约场景下的"超卖"问题：
1. 患者 A 查询到号源 X 有 1 个剩余号
2. 患者 B 同时查询到号源 X 有 1 个剩余号
3. 患者 A 先锁定号源 X
4. 患者 B 尝试锁定时失败，避免了重复预约

### 5.2 锁定流程

```
患者发起锁定 → 检查号源是否被锁定且未超时
       │
       ├── 是（被有效锁定）→ 返回 ErrSlotAlreadyLocked
       │
       └── 否（可用/超时锁定）→ 设置锁定信息
                                   │
                                   ├── LockedBy = 患者ID
                                   ├── LockedAt = 当前时间
                                   ├── LockExpireAt = 当前时间 + LockTimeout
                                   └── 返回锁定成功
```

### 5.3 锁定的生命周期

1. **锁定创建**：`LockSlot` 成功后创建
2. **确认转化为预约**：`ConfirmAppointment` 成功后清除锁定，生成预约
3. **主动释放**：`ReleaseLock` 由锁定者主动放弃
4. **超时自动释放**：`CheckAndReleaseExpiredLocks` 批量检查释放，或在下次操作号源时懒检查释放

### 5.4 并发安全保证

- 所有号源操作都在 Store 的写锁（sync.RWMutex 的 Lock/Unlock）保护下执行
- 锁定、确认、释放、查询都在锁内完成，保证状态一致性
- `CreateAppointment` 内部串行调用 LockSlot + ConfirmAppointment，每步都加锁

## 6. 候补队列机制

### 6.1 候补队列的组织

候补队列按 key 分桶存储，key 有两种格式：
- 按医生维度：`doc:{DoctorID}:{Date}`
- 按科室维度：`dept:{Department}:{Date}`

同一个患者可以同时加入多个候补队列（如既等张医生又等内科）。

### 6.2 号源释放时的候补匹配

当号源 slot (Doctor=D, Department=Dep, Date=Date) 释放时：
1. 收集 key = `doc:D:Date` 队列中的所有候补项
2. 收集 key = `dept:Dep:Date` 队列中的所有候补项
3. 合并后按 JoinedAt 升序排序（先入先出）
4. 依次尝试为每个候补项生成预约：
   - 患者不存在则跳过并移除
   - 号源还有剩余则生成预约，并移除候补项
   - 号源已满则停止处理

### 6.3 候补触发时机

候补补位在以下场景自动触发：
- 患者取消预约（`CancelAppointment`）
- 患者爽约（`RecordNoShow`）
- 患者改约（原号源释放后）

## 7. 错误处理

### 7.1 错误类型定义

| 错误变量 | 说明 | 触发场景 |
|---------|------|---------|
| `ErrDoctorNotFound` | 医生不存在 | 添加排班/候诊时医生ID无效 |
| `ErrPatientNotFound` | 患者不存在 | 锁定/预约/候诊时患者ID无效 |
| `ErrScheduleNotFound` | 排班不存在 | 查询排班时 |
| `ErrSlotNotFound` | 号源不存在 | 锁定/预约时号源ID无效 |
| `ErrAppointmentNotFound` | 预约不存在 | 改约/取消/爽约时预约ID无效 |
| `ErrSlotAlreadyLocked` | 号源已被锁定 | 其他患者已有效锁定号源 |
| `ErrSlotAlreadyBooked` | 号源已满 | 号源无剩余容量 |
| `ErrSlotNotLocked` | 号源未被锁定 | 确认预约时号源不在锁定状态 |
| `ErrSlotLockExpired` | 锁定已超时 | 确认预约时锁定已过期 |
| `ErrInvalidLockOwner` | 非锁定所有者 | 患者B尝试确认/释放患者A的锁定 |
| `ErrAppointmentNotPending` | 预约非待确认 | 预留 |
| `ErrAppointmentNotActive` | 预约非活跃 | 尝试对已取消/改约的预约做爽约操作 |
| `ErrCannotChangePast` | 超过改约截止 | 距离预约不足24小时时尝试改约 |
| `ErrCannotCancelPast` | 超过取消截止 | 距离预约不足12小时时尝试取消 |
| `ErrSameSlotChange` | 改约到同一号源 | 新旧号源相同 |
| `ErrChangeNotAllowed` | 状态不允许变更 | 预约非CONFIRMED状态时尝试改约/取消 |
| `ErrInvalidDate` | 日期格式错误 | 日期不符合YYYY-MM-DD格式 |
| `ErrSlotNoCapacity` | 号源无容量 | 号源剩余数为0时尝试预约 |
| `ErrWaitQueueItemNotFound` | 候补项不存在 | 预留 |
| `ErrPatientAlreadyBooked` | 患者已预约 | 预留（当前版本未强制禁止） |
| `ErrLockTimeoutInvalid` | 锁定超时配置无效 | Store.LockTimeout <= 0 |
| `ErrNoShowRemarkRequired` | 爽约备注必填 | 爽约时未提供Remark |

## 8. 使用示例

### 8.1 初始化环境

```go
package main

import (
    "fmt"
    "solocoder-4-go/internal/appointment"
)

func main() {
    store := appointment.NewStore()
    svc := appointment.NewService(store)

    doctor1ID := store.AddDoctor(&appointment.Doctor{
        Name:       "张医生",
        Title:      "主任医师",
        Department: "内科",
        LicenseNo:  "DOC001",
    })

    patient1ID := store.AddPatient(&appointment.Patient{
        Name:            "王小明",
        IDCard:          "110101199001011234",
        Phone:           "13800138000",
        Gender:          "男",
        Age:             35,
        MedicalRecordNo: "MR001",
    })

    schedule, _ := svc.AddSchedule(&appointment.AddScheduleRequest{
        DoctorID: doctor1ID,
        Date:     "2025-06-22",
        Slots: []appointment.ScheduleSlotSpec{
            {StartTime: "08:00", EndTime: "09:00", TotalCapacity: 5},
            {StartTime: "09:00", EndTime: "10:00", TotalCapacity: 3},
            {StartTime: "10:00", EndTime: "11:00", TotalCapacity: 1},
        },
    })
    fmt.Printf("排班创建成功：ID=%s，共 %d 个时段\n", schedule.ID, len(schedule.Slots))
}
```

### 8.2 排班查询

```go
results, err := svc.QuerySchedules(&appointment.QueryScheduleRequest{
    Department: "内科",
    Date:       "2025-06-22",
})
if err != nil {
    fmt.Printf("查询失败: %v\n", err)
    return
}

for _, sch := range results {
    fmt.Printf("医生 %s，日期 %s：\n", sch.DoctorName, sch.Date)
    for _, slot := range sch.Slots {
        status := "可预约"
        if slot.IsLocked {
            status = "锁定中"
        } else if slot.IsFull {
            status = "已满"
        }
        fmt.Printf("  %s-%s  剩余 %d/%d  [%s]  SlotID=%s\n",
            slot.StartTime, slot.EndTime,
            slot.RemainingCount, slot.TotalCapacity,
            status, slot.SlotID)
    }
}
```

### 8.3 号源锁定 + 确认（完整预约）

```go
slotID := schedule.Slots[0].ID

locked, err := svc.LockSlot(&appointment.LockSlotRequest{
    SlotID:    slotID,
    PatientID: patient1ID,
})
if err != nil {
    fmt.Printf("锁定失败: %v\n", err)
    return
}
fmt.Printf("号源锁定成功：%s\n", locked.LockExpireAt)

appointment, err := svc.ConfirmAppointment(&appointment.ConfirmAppointmentRequest{
    SlotID:    slotID,
    PatientID: patient1ID,
})
if err != nil {
    fmt.Printf("确认失败: %v\n", err)
    return
}
fmt.Printf("预约创建成功：ID=%s，状态=%s\n", appointment.ID, appointment.Status)
```

### 8.4 一键创建预约

```go
appointment, err := svc.CreateAppointment(&appointment.CreateAppointmentRequest{
    SlotID:    slotID,
    PatientID: patient1ID,
})
if err != nil {
    fmt.Printf("预约失败: %v\n", err)
    return
}
fmt.Printf("预约ID: %s, 状态: %s\n", appointment.ID, appointment.Status)
```

### 8.5 改约

```go
newAppt, err := svc.ChangeAppointment(&appointment.ChangeAppointmentRequest{
    AppointmentID: appointment.ID,
    NewSlotID:     schedule.Slots[1].ID,
    PatientID:     patient1ID,
})
if err != nil {
    fmt.Printf("改约失败: %v\n", err)
    return
}
fmt.Printf("改约成功，新时段：%s-%s\n", newAppt.StartTime, newAppt.EndTime)
```

### 8.6 取消预约

```go
cancelled, err := svc.CancelAppointment(&appointment.CancelAppointmentRequest{
    AppointmentID: appointment.ID,
    PatientID:     patient1ID,
    Reason:        "临时有事",
})
if err != nil {
    fmt.Printf("取消失败: %v\n", err)
    return
}
fmt.Printf("预约已取消：状态=%s\n", cancelled.Status)
```

### 8.7 加入候补 + 爽约自动补位

```go
fullSlotID := schedule.Slots[2].ID

_, _ = svc.CreateAppointment(&appointment.CreateAppointmentRequest{
    SlotID:    fullSlotID,
    PatientID: store.AddPatient(&appointment.Patient{Name: "其他患者"}),
})

_, err = svc.JoinWaitQueue(&appointment.JoinWaitQueueRequest{
    PatientID: patient1ID,
    DoctorID:  doctor1ID,
    Date:      "2025-06-22",
})
fmt.Println("加入候补队列成功")

otherAppt, _ := svc.CreateAppointment(&appointment.CreateAppointmentRequest{
    SlotID:    fullSlotID,
    PatientID: store.AddPatient(&appointment.Patient{Name: "爽约者"}),
})

_, err = svc.RecordNoShow(&appointment.RecordNoShowRequest{
    AppointmentID: otherAppt.ID,
    Remark:        "患者未到且未提前联系",
})

list, _ := svc.ListAppointmentsByPatient(patient1ID)
fmt.Printf("patient1 的预约数：%d（候补自动补位生成）\n", len(list))
```

### 8.8 定时清理超时锁定

```go
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        released := svc.CheckAndReleaseExpiredLocks()
        if released > 0 {
            fmt.Printf("释放了 %d 个超时锁定\n", released)
        }
    }
}()
```

## 9. 测试覆盖说明

模块包含完整的单元测试（67 个测试用例），覆盖以下场景：

### 9.1 正常流程测试
- `TestAddSchedule_Success`：添加排班成功
- `TestQuerySchedules_*`：排班多维度查询
- `TestLockSlot_Success`：号源锁定成功
- `TestConfirmAppointment_Success`：确认预约成功
- `TestReleaseLock_Success`：主动释放锁定
- `TestCreateAppointment_Success`：一键创建预约
- `TestChangeAppointment_Success`：改约成功
- `TestCancelAppointment_Success`：取消预约成功
- `TestRecordNoShow_Success`：记录爽约成功
- `TestJoinWaitQueue_*`：加入候补队列
- `TestGetWaitQueue`：查询候补队列
- `TestAutoFillFromWaitQueue_*`：候补自动补位（取消/爽约/改约/多候选）
- `TestFullWorkflow`：查询→预约→改约→取消→爽约→补位 完整流程

### 9.2 边界条件测试
- `TestAddSchedule_SkipInvalidCapacity`：号源容量为0或负数时跳过
- `TestQuerySchedules_SlotRemainingAndFull`：号源满时状态正确
- `TestLockSlot_FullSlot`：对已满号源尝试锁定
- `TestCreateAppointment_SlotFull`：号源满时创建预约失败
- `TestLockSlot_ExpiredLockReuse`：锁定超时后可被他人重新锁定
- `TestCheckAndReleaseExpiredLocks`：批量释放超时锁定
- `TestChangeAppointment_SameSlot`：改约到同一号源
- `TestChangeAppointment_NewSlotFull`：新号源已满时改约失败
- `TestAutoFillFromWaitQueue_MultipleCandidates`：多个候补中先入先出补位

### 9.3 异常分支测试
- 所有实体的不存在场景（医生/患者/号源/预约）
- 锁定相关：重复锁定、错误患者确认/释放、锁定超时后确认
- 改约相关：错误患者操作、状态不允许、超过截止时间
- 取消相关：错误患者操作、重复取消、超过截止时间
- 爽约相关：无备注、非确认状态、无效预约
- 候补相关：无效日期/医生/患者
- 排班相关：无效医生、无效日期格式

## 10. 运行测试

```bash
cd solocoder-4-go
go test ./internal/appointment/ -v
```

预期输出：67 个测试用例全部 PASS。
