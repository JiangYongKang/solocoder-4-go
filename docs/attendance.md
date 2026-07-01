# 班级考勤管理模块需求文档

## 1. 模块概述

班级考勤管理模块实现了从学生信息管理、考勤规则配置、签到签退、请假审批到考勤统计与异常通知的完整业务流程。模块使用内存数据结构模拟所有业务实体，支持并发安全的签到操作和防重复签到机制，确保考勤数据的准确性和一致性。

### 主要功能特性

- **学生签到签退**：学生可以在指定考勤场次中签到和签退，系统记录签到签退时间并防止重复操作
- **请假审批**：学生可以提交请假申请，审批通过后对应考勤场次按请假处理
- **迟到早退判定**：根据考勤规则和签到签退时间自动判断正常、迟到、早退或缺勤状态
- **考勤汇总**：支持按班级和学生汇总考勤结果，统计出勤率等指标
- **异常缺勤通知**：自动生成迟到、早退、缺勤学生的通知记录
- **并发安全**：基于读写锁实现并发场景下的数据一致性

## 2. 核心结构体职责

### 2.1 基础实体

#### Student（学生）
- **职责**：存储学生基本信息，用于签到和请假时的身份识别
- **关键字段**：
  - `ID`：学生唯一标识
  - `Name`：学生姓名

#### Class（班级）
- **职责**：管理班级信息和班级成员
- **关键字段**：
  - `ID`：班级唯一标识
  - `Name`：班级名称
  - `Teacher`：班主任/授课教师
  - `Students`：班级学生列表（map[学生ID]学生对象）

### 2.2 考勤规则相关

#### AttendanceRule（考勤规则）
- **职责**：定义考勤的时间规则和判定阈值
- **关键字段**：
  - `ID`：规则唯一标识
  - `Name`：规则名称
  - `CheckInStartTime`：签到开始时间
  - `CheckInEndTime`：签到截止时间
  - `CheckOutStartTime`：签退开始时间
  - `CheckOutEndTime`：签退截止时间
  - `LateThresholdMinutes`：迟到判定阈值（分钟）
  - `EarlyLeaveThresholdMinutes`：早退判定阈值（分钟）

#### AttendanceSession（考勤场次）
- **职责**：表示一次具体的考勤活动，关联班级和考勤规则
- **关键字段**：
  - `ID`：场次唯一标识
  - `ClassID`：关联班级ID
  - `RuleID`：关联规则ID
  - `Date`：考勤日期
  - `CreatedAt`：创建时间
  - `Active`：是否处于激活状态

### 2.3 签到与请假相关

#### CheckInRecord（签到记录）
- **职责**：记录学生单次考勤的签到和签退时间
- **关键字段**：
  - `StudentID`：学生ID
  - `SessionID`：考勤场次ID
  - `CheckInTime`：签到时间
  - `CheckOutTime`：签退时间

#### LeaveRequest（请假申请）
- **职责**：存储学生请假申请和审批结果
- **关键字段**：
  - `ID`：请假申请唯一标识
  - `StudentID/StudentName`：学生信息
  - `SessionID`：关联考勤场次ID
  - `Type`：请假类型（事假/病假/其他）
  - `Reason`：请假原因
  - `Status`：审批状态（待处理/已批准/已驳回）
  - `ApproverID/ApproverName`：审批人信息
  - `Comment`：审批意见
  - `CreatedAt`：申请时间
  - `ProcessedAt`：处理时间

### 2.4 考勤结果相关

#### AttendanceRecord（考勤记录）
- **职责**：存储单次考勤的最终判定结果
- **关键字段**：
  - `ID`：记录唯一标识
  - `StudentID/StudentName`：学生信息
  - `SessionID/ClassID`：关联场次和班级
  - `CheckInTime/CheckOutTime`：签到签退时间（从签到记录同步）
  - `Status`：考勤状态（正常/迟到/早退/迟到早退/缺勤/请假）
  - `IsLate`：是否迟到（独立标志，用于复合状态统计）
  - `IsEarlyLeave`：是否早退（独立标志，用于复合状态统计）
  - `IsLeave`：是否为请假
  - `LeaveID`：关联的请假申请ID
  - `CalculatedAt`：结果计算时间

#### AttendanceSummary（考勤汇总）
- **职责**：存储指定时间范围内的考勤统计结果
- **关键字段**：
  - `StudentID/StudentName`：学生信息（学生汇总时）
  - `ClassID/ClassName`：班级信息
  - `TotalSessions`：总考勤场次
  - `PresentCount`：正常出勤次数（仅完全正常，不含迟到早退）
  - `LateCount`：迟到次数（基于 IsLate 标志统计，包含复合状态）
  - `EarlyLeaveCount`：早退次数（基于 IsEarlyLeave 标志统计，包含复合状态）
  - `AbsentCount`：缺勤次数
  - `LeaveCount`：请假次数
  - `AttendanceRate`：出勤率 = (PresentCount + LateCount + EarlyLeaveCount + LeaveCount) / TotalSessions * 100

> **注意**：迟到和早退统计基于独立标志位，当学生同时迟到和早退时，LateCount 和 EarlyLeaveCount 会分别各加1，确保两种异常都被完整记录。

#### AbsenceNotification（缺勤通知）
- **职责**：存储异常考勤的通知记录
- **关键字段**：
  - `ID`：通知唯一标识
  - `StudentID/StudentName`：学生信息
  - `ClassID/ClassName`：班级信息
  - `SessionID/SessionDate`：场次信息
  - `Status`：异常状态（迟到/早退/缺勤）
  - `NotifiedAt`：通知生成时间
  - `NotificationSent`：是否已发送

> **注意**：当学生同时迟到和早退时，会生成两条独立的通知记录（一条迟到通知和一条早退通知），确保两种异常都被完整通知。

### 2.5 核心服务

#### Service（考勤服务）
- **职责**：基于内存的数据存储层，提供所有实体的CRUD操作和业务逻辑
- **核心能力**：
  - 学生、班级、审批人的管理
  - 考勤规则和场次的创建与管理
  - 学生签到签退操作
  - 请假申请提交与审批
  - 考勤状态自动判定
  - 考勤结果汇总统计
  - 异常缺勤通知生成

## 3. 考勤状态判定规则

### 3.1 状态类型

| 状态 | 说明 |
|------|------|
| `PRESENT`（正常） | 按时签到和签退 |
| `LATE`（迟到） | 签到时间晚于迟到阈值，但在签到截止时间前 |
| `EARLY_LEAVE`（早退） | 签退时间早于早退阈值，但在签退开始时间后 |
| `LATE_AND_EARLY_LEAVE`（迟到早退） | 同时满足迟到和早退条件，两种异常分别统计 |
| `ABSENT`（缺勤） | 未签到，或请假申请被驳回 |
| `LEAVE`（请假） | 请假申请已批准 |

### 3.2 判定逻辑

#### 时间计算基准
- 将考勤规则中的时间与考勤场次的日期结合，计算当天的具体时间点
- 所有时间比较统一截断到分钟精度，确保相同时钟时刻（如 09:00:00）无论纳秒值如何，判定结果一致
- 例如：规则中签到开始时间为 08:00，场次日期为 2024-06-01，则实际签到开始时间为 2024-06-01 08:00:00

#### 判定优先级
1. **请假优先**：如果学生有已批准的请假申请，直接判定为 `LEAVE`
2. **缺勤判定**：如果没有签到记录或请假申请被驳回，判定为 `ABSENT`
3. **迟到判定**：签到时间 > 签到开始时间 + 迟到阈值，且 ≤ 签到截止时间
4. **早退判定**：签退时间 < 签退截止时间 - 早退阈值，且 ≥ 签退开始时间
5. **迟到且早退**：同时满足迟到和早退条件，判定为 `LATE_AND_EARLY_LEAVE`，并设置 `IsLate` 和 `IsEarlyLeave` 两个独立标志
6. **正常出勤**：不满足以上异常条件

#### 时间比较精度
- 所有时间值在比较前统一使用 `Truncate(time.Minute)` 截断到分钟
- 确保在阈值临界点（如恰好 08:10:00）的判定结果稳定，不受纳秒级差异影响
- 迟到判定：`checkInTime.After(lateThreshold) && !checkInTime.After(ruleCheckInEnd)`
- 早退判定：`checkOutTime.Before(earlyThreshold) && !checkOutTime.Before(ruleCheckOutStart)`

### 3.3 示例场景

假设规则配置：
- 签到时间：08:00 - 08:30
- 签退时间：17:00 - 18:00
- 迟到阈值：10分钟
- 早退阈值：10分钟

则：
- 迟到阈值时间 = 08:00 + 10分钟 = 08:10
- 早退阈值时间 = 18:00 - 10分钟 = 17:50

| 签到时间 | 签退时间 | 判定结果 | 附加说明 |
|----------|----------|----------|----------|
| 08:05 | 17:55 | PRESENT（正常） | IsLate=false, IsEarlyLeave=false |
| 08:15 | - | LATE（迟到） | IsLate=true, IsEarlyLeave=false |
| 08:05 | 17:30 | EARLY_LEAVE（早退） | IsLate=false, IsEarlyLeave=true |
| 08:15 | 17:30 | LATE_AND_EARLY_LEAVE（迟到早退） | IsLate=true, IsEarlyLeave=true，分别计入迟到和早退统计 |
| 08:10:00.000 | 17:50:00.999 | PRESENT（正常） | 分钟级精度，纳秒差异不影响判定 |
| - | - | ABSENT（缺勤） | - |
| （已请假） | - | LEAVE（请假） | - |

## 4. 使用示例

### 4.1 初始化服务并配置基础数据

```go
package main

import (
    "time"
    "solocoder-4-go/internal/attendance"
)

func main() {
    // 初始化服务
    service := attendance.NewService()

    // 添加学生
    studentID1 := service.AddStudent(&attendance.Student{Name: "张三"})
    studentID2 := service.AddStudent(&attendance.Student{Name: "李四"})

    // 添加审批人
    approverID := service.AddApprover(&attendance.Student{Name: "王老师"})

    // 创建班级
    class := service.CreateClass("计算机科学1班", "王老师")

    // 将学生加入班级
    service.AddStudentToClass(class.ID, studentID1)
    service.AddStudentToClass(class.ID, studentID2)

    // 创建考勤规则
    checkInStart := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
    checkInEnd := time.Date(2024, 1, 1, 8, 30, 0, 0, time.UTC)
    checkOutStart := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)
    checkOutEnd := time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)
    rule, _ := service.CreateRule("日常考勤", checkInStart, checkInEnd, checkOutStart, checkOutEnd, 10, 10)

    // 创建考勤场次
    sessionDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
    session, _ := service.CreateSession(class.ID, rule.ID, sessionDate)
}
```

### 4.2 学生签到签退

```go
// 学生签到
checkInTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
_, err := service.CheckIn(&attendance.CheckInRequest{
    StudentID: studentID1,
    SessionID: session.ID,
    CheckTime: checkInTime,
})
if err != nil {
    // 处理错误：重复签到、场次未激活等
}

// 学生签退
checkOutTime := time.Date(2024, 6, 1, 17, 55, 0, 0, time.UTC)
_, err = service.CheckOut(&attendance.CheckInRequest{
    StudentID: studentID1,
    SessionID: session.ID,
    CheckTime: checkOutTime,
})
```

### 4.3 请假申请与审批

```go
// 学生提交请假申请
leave, err := service.SubmitLeaveRequest(&attendance.LeaveRequestRequest{
    StudentID: studentID2,
    SessionID: session.ID,
    Type:      attendance.LeaveTypeSick,
    Reason:    "身体不适，需要请假一天",
})

// 审批人处理请假申请
_, err = service.ProcessLeaveRequest(&attendance.ProcessLeaveRequest{
    LeaveID:    leave.ID,
    ApproverID: approverID,
    Approved:   true,
    Comment:    "批准请假，注意休息",
})
```

### 4.4 计算考勤结果与生成通知

```go
// 计算本场次所有学生的考勤结果
records, err := service.CalculateAttendance(session.ID)
for _, record := range records {
    fmt.Printf("学生：%s，状态：%s\n", record.StudentName, record.Status)
}

// 生成异常缺勤通知
notifications, err := service.GenerateAbsenceNotifications(session.ID)
for _, notification := range notifications {
    fmt.Printf("通知：%s %s %s\n", notification.StudentName, notification.Status, notification.SessionDate)
}
```

### 4.5 考勤汇总统计

```go
// 按学生汇总
startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
endDate := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
studentSummary, err := service.GetStudentSummary(studentID1, class.ID, startDate, endDate)
fmt.Printf("出勤率：%.2f%%\n", studentSummary.AttendanceRate)

// 按班级汇总
classSummary, err := service.GetClassSummary(class.ID, startDate, endDate)
fmt.Printf("班级出勤人次：%d\n", classSummary.PresentCount)
```

## 5. 错误码说明

| 错误码 | 说明 |
|--------|------|
| `ErrClassNotFound` | 班级不存在 |
| `ErrStudentNotFound` | 学生不存在 |
| `ErrStudentNotInClass` | 学生不在该班级中 |
| `ErrSessionNotFound` | 考勤场次不存在 |
| `ErrSessionNotActive` | 考勤场次未激活 |
| `ErrAlreadyCheckedIn` | 已签到，不能重复签到 |
| `ErrAlreadyCheckedOut` | 已签退，不能重复签退 |
| `ErrCheckInRequired` | 需先签到才能签退 |
| `ErrLeaveRequestNotFound` | 请假申请不存在 |
| `ErrLeaveAlreadyProcessed` | 请假申请已处理，不能重复处理 |
| `ErrInvalidTimeRange` | 时间范围无效（开始时间需早于结束时间） |
| `ErrApproverNotFound` | 审批人不存在 |
| `ErrDuplicateStudentInClass` | 学生已在班级中，不能重复添加 |
