# 远程问诊会话模块 (teleconsult)

## 一、模块功能概述

本模块实现了基于内存数据结构的远程问诊会话管理系统，提供完整的图文咨询、医生接诊、会话超时、记录归档及患者追问等功能。

### 主要功能点

| 功能 | 说明 |
|------|------|
| 图文咨询 | 患者发起问诊，支持发送文字消息和图片消息，消息按会话时间顺序保存 |
| 医生接诊 | 医生接诊后会话进入进行中状态；未接诊前医生不得发送正式诊疗意见 |
| 会话超时 | 超过允许时间未接诊（默认 30 分钟）或长期无互动（默认 60 分钟）的会话自动关闭 |
| 记录归档 | 会话结束后生成归档记录，包含完整诊断、治疗建议和消息历史 |
| 追问限制 | 患者归档后可在限定次数内（默认 3 次）追加追问消息 |

---

## 二、核心结构体职责

### 2.1 Patient（患者）
| 字段 | 类型 | 说明 |
|------|------|------|
| ID | string | 患者唯一标识 |
| Name | string | 患者姓名 |
| Gender | string | 性别 |
| Age | int | 年龄 |
| Phone | string | 联系电话 |

### 2.2 Doctor（医生）
| 字段 | 类型 | 说明 |
|------|------|------|
| ID | string | 医生唯一标识 |
| Name | string | 医生姓名 |
| Title | string | 职称 |
| Department | string | 所属科室 |
| LicenseNo | string | 执业证书编号 |
| Available | bool | 是否可接诊 |

### 2.3 Message（消息）
| 字段 | 类型 | 说明 |
|------|------|------|
| ID | string | 消息唯一标识 |
| SessionID | string | 所属会话 ID |
| SenderID | string | 发送者 ID |
| SenderType | MessageSender | 发送者类型：PATIENT/DOCTOR/SYSTEM |
| Type | MessageType | 消息类型：TEXT/IMAGE/SYSTEM |
| Content | string | 文字内容或图片描述 |
| ImageURL | string | 图片 URL（图片消息） |
| CreatedAt | time.Time | 发送时间 |
| IsMedicalAdvice | bool | 是否为正式诊疗意见 |

### 2.4 ConsultSession（问诊会话）
| 字段 | 类型 | 说明 |
|------|------|------|
| ID | string | 会话唯一标识 |
| PatientID | string | 患者 ID |
| PatientName | string | 患者姓名（冗余） |
| DoctorID | string | 医生 ID |
| DoctorName | string | 医生姓名（冗余） |
| Department | string | 科室 |
| ChiefComplaint | string | 主诉 |
| Status | SessionStatus | 会话状态 |
| Messages | []*Message | 消息列表（按发送时间排序） |
| CreatedAt | time.Time | 创建时间 |
| AcceptedAt | *time.Time | 接诊时间 |
| LastMessageAt | time.Time | 最后消息时间 |
| ClosedAt | *time.Time | 关闭时间 |
| CloseReason | string | 关闭原因 |
| FollowUpCount | int | 已追问次数 |
| FollowUpLimit | int | 追问次数上限 |
| ArchiveID | string | 归档记录 ID |

### 2.5 ArchiveRecord（归档记录）
| 字段 | 类型 | 说明 |
|------|------|------|
| ID | string | 归档唯一标识 |
| SessionID | string | 所属会话 ID |
| PatientID / PatientName | string | 患者信息 |
| DoctorID / DoctorName | string | 医生信息 |
| Department | string | 科室 |
| ChiefComplaint | string | 主诉 |
| Diagnosis | string | 诊断结论 |
| TreatmentAdvice | string | 治疗建议 |
| FollowUpPlan | string | 随访计划 |
| Messages | []*Message | 完整消息历史（含后续追问） |
| FollowUpCount | int | 已追问次数 |
| FollowUpLimit | int | 追问次数上限 |
| ArchivedAt | time.Time | 归档时间 |

### 2.6 Service（问诊服务）
内存存储 + 读写锁保护，管理所有实体和业务逻辑。

---

## 三、会话状态流转

```
                      (超时未接诊)
     ┌───────────────────────────────────────┐
     │                                       ▼
  PENDING ──医生接诊──▶ ONGOING ──完成问诊──▶ COMPLETED/ARCHIVED
     │                     │
     │                     │(长期无互动)
     │                     ▼
     └───────────────▶ CLOSED
                       (关闭后不可发送普通消息)
```

### 状态说明

| 状态 | 含义 | 允许的操作 |
|------|------|-----------|
| **PENDING**（待接诊） | 患者发起，等待医生接诊 | 患者可发送消息；医生可发送非诊疗性消息；医生可接诊 |
| **ONGOING**（进行中） | 医生已接诊 | 医患双方均可发送消息；医生可标记诊疗意见；医生可结束问诊 |
| **CLOSED**（已关闭） | 超时或异常关闭 | 仅可通过归档记录的追问通道发送追问（如已归档） |
| **COMPLETED**（已完成） | 医生主动结束（短暂过渡态） | 同 ARCHIVED |
| **ARCHIVED**（已归档） | 生成归档记录 | 患者可在次数限制内发送追问消息；不可发送普通会话消息 |

### 关键约束
- 只有 ONGOING 状态下，医生发送的消息才能标记 `IsMedicalAdvice=true`
- PENDING 状态下如设置 `IsMedicalAdvice=true`，返回 `ErrSessionNotAccepted`
- 一旦进入 CLOSED / COMPLETED / ARCHIVED，普通 `SendMessage` 返回 `ErrSessionClosed`
- 追问消息走独立接口 `SendFollowUp`，不受会话关闭限制（但受次数限制）

---

## 四、默认配置常量

| 常量 | 默认值 | 说明 |
|------|--------|------|
| DefaultAcceptTimeout | 30 分钟 | 待接诊超时时长 |
| DefaultInactivityTimeout | 60 分钟 | 进行中无互动超时时长 |
| DefaultFollowUpLimit | 3 次 | 归档后可追问次数上限 |
| MaxMessageLength | 5000 字符 | 单条文字消息最大长度 |

可通过以下方法动态调整：
- `SetAcceptTimeout(d time.Duration)`
- `SetInactivityTimeout(d time.Duration)`
- `SetFollowUpLimit(limit int)`

---

## 五、使用示例

### 5.1 初始化服务

```go
svc := teleconsult.NewService()
```

### 5.2 添加患者和医生

```go
patientID := svc.AddPatient(&teleconsult.Patient{
    Name:   "张三",
    Gender: "男",
    Age:    35,
    Phone:  "13800138000",
})

doctorID := svc.AddDoctor(&teleconsult.Doctor{
    Name:       "李医生",
    Title:      "主任医师",
    Department: "神经内科",
    LicenseNo:  "LIC12345",
    Available:  true,
})
```

### 5.3 患者发起问诊

```go
session, err := svc.CreateSession(&teleconsult.CreateSessionRequest{
    PatientID:      patientID,
    DoctorID:       doctorID,
    ChiefComplaint: "反复头痛2周，伴失眠",
    InitialMessage: "医生您好！我最近两周反复头痛...",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("会话ID:", session.ID, "状态:", session.Status) // PENDING
```

### 5.4 医生接诊

```go
accepted, err := svc.AcceptSession(&teleconsult.AcceptSessionRequest{
    SessionID: session.ID,
    DoctorID:  doctorID,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("接诊后状态:", accepted.Status) // ONGOING
```

### 5.5 发送图文消息

```go
// 患者发送文字
_, _ = svc.SendMessage(&teleconsult.SendMessageRequest{
    SessionID:   session.ID,
    SenderID:    patientID,
    SenderType:  teleconsult.MessageSenderPatient,
    MessageType: teleconsult.MessageTypeText,
    Content:     "头痛一般下午开始加重。",
})

// 患者发送图片
_, _ = svc.SendMessage(&teleconsult.SendMessageRequest{
    SessionID:   session.ID,
    SenderID:    patientID,
    SenderType:  teleconsult.MessageSenderPatient,
    MessageType: teleconsult.MessageTypeImage,
    Content:     "CT报告",
    ImageURL:    "https://med.example.com/ct/001.png",
})

// 医生发送正式诊疗意见
_, _ = svc.SendMessage(&teleconsult.SendMessageRequest{
    SessionID:       session.ID,
    SenderID:        doctorID,
    SenderType:      teleconsult.MessageSenderDoctor,
    MessageType:     teleconsult.MessageTypeText,
    Content:         "诊断为紧张性头痛，建议...",
    IsMedicalAdvice: true,
})
```

### 5.6 超时检查（建议定时调用）

```go
// 单会话检查
updated, _ := svc.CheckSessionTimeout(session.ID)

// 批量检查所有会话
closedCount := svc.CheckAllSessionsTimeout()
fmt.Printf("本次自动关闭 %d 个会话\n", closedCount)
```

### 5.7 医生完成问诊并归档

```go
archive, err := svc.CompleteSession(&teleconsult.CompleteSessionRequest{
    SessionID:      session.ID,
    DoctorID:       doctorID,
    Diagnosis:      "紧张性头痛（Tension-Type Headache）",
    TreatmentAdvice: "1. 对乙酰氨基酚 500mg 口服...\n2. 规律作息...",
    FollowUpPlan:   "2周后复诊，若头痛加重立即就诊。",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("归档ID:", archive.ID)
```

### 5.8 患者在归档后追加追问

```go
followUpMsg, err := svc.SendFollowUp(&teleconsult.SendFollowUpRequest{
    ArchiveID:   archive.ID,
    PatientID:   patientID,
    MessageType: teleconsult.MessageTypeText,
    Content:     "医生您好，按您的方法调整一周了，头痛明显减轻！药还要继续吃吗？",
})
if err == teleconsult.ErrFollowUpLimitExceeded {
    fmt.Println("已超出可追问次数，请重新发起问诊")
} else if err != nil {
    log.Fatal(err)
}
fmt.Println("追问已发送，剩余次数:", archive.FollowUpLimit - archive.FollowUpCount - 1)
```

### 5.9 查询相关接口

```go
// 获取单会话/归档
ses, _ := svc.GetSession(sessionID)
arc, _ := svc.GetArchive(archiveID)

// 按患者/医生查询
patientSessions, _ := svc.ListSessionsByPatient(patientID)
doctorSessions, _ := svc.ListSessionsByDoctor(doctorID)
patientArchives, _ := svc.ListArchivesByPatient(patientID)
doctorArchives, _ := svc.ListArchivesByDoctor(doctorID)
```

---

## 六、错误变量速查

| 错误变量 | 触发场景 |
|----------|---------|
| `ErrPatientNotFound` | 患者 ID 不存在 |
| `ErrDoctorNotFound` | 医生 ID 不存在 |
| `ErrDoctorNotAvailable` | 医生标记为不可接诊 |
| `ErrSessionNotFound` | 会话 ID 不存在 |
| `ErrArchiveNotFound` | 归档 ID 不存在 |
| `ErrInvalidSessionStatus` | 操作与会话状态不匹配（如未接诊就结束问诊） |
| `ErrSessionNotAccepted` | 医生未接诊时发送带 `IsMedicalAdvice=true` 的消息 |
| `ErrSessionClosed` | 会话已关闭/已归档，调用普通 `SendMessage` |
| `ErrEmptyContent` | 消息内容为空（文字或图片 URL） |
| `ErrMessageTooLong` | 文字消息超过 `MaxMessageLength` |
| `ErrFollowUpLimitExceeded` | 归档后追问次数达上限 |
| `ErrNotPatient` | 发送者不是会话中的患者 |
| `ErrNotDoctor` | 发送者不是会话中的接诊医生 |
| `ErrAlreadyAccepted` | 会话已接诊，重复调用 `AcceptSession` |
| `ErrInvalidDoctorAssignment` | 接诊医生与会话指定医生不一致 |
