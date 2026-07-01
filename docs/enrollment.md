# 学校选课系统模块

## 1. 模块概述

学校选课系统模块（enrollment）是一个基于内存数据结构的学生课程管理系统，支持课程容量限制、先修课校验、退课候补递补、学分统计和选课时间窗口控制等核心功能。模块使用 Go 语言实现，提供线程安全的并发访问支持。

## 2. 模块功能

### 2.1 课程容量限制
- 每门课程配置最大选课容量
- 学生选课时实时检查课程剩余容量
- 容量已满时学生不能直接选课，自动进入候补队列
- 支持查询课程当前剩余可选座位数

### 2.2 先修课校验
- 课程可以配置多门先修课程
- 学生选课时检查是否已完成所有先修课程
- 未完成先修课程的学生不得直接选入该课程
- 课程满员时允许未完成先修课的学生进入候补队列（学生可能在候补期间完成先修课）
- 候补递补时重新校验先修课资格，不合格者自动跳过

### 2.3 退课候补递补
- 课程满员后允许学生进入候补队列
- 候补队列严格按照加入顺序（FIFO）排列
- 已有学生退课时自动按候补顺序尝试递补
- 递补时重新校验学生资格（先修课、学分限制）
- 资格不符的候补学生自动被移出队列，继续尝试下一位
- 支持学生主动退出候补队列

### 2.4 学分统计与限制
- 支持统计学生当前已选课程的总学分
- 每个学生可配置最大允许选课学分
- 选课时检查学分上限，超出时禁止选课
- 候补递补时重新校验学分限制
- 退课后自动更新学生已选学分统计

### 2.5 选课时间窗口控制
- 支持配置选课开放的起止时间窗口
- 学生只能在开放时间窗口内进行选课或退课操作
- 时间窗口外的选课/退课操作返回明确错误
- 未配置时间窗口时默认始终允许操作
- 支持动态修改时间窗口配置

## 3. 核心结构体职责

### 3.1 Store
**职责**：选课系统核心管理器，提供所有对外接口，维护内存中的所有数据。

**主要字段**：
- `students`：学生信息映射表
- `courses`：课程信息映射表
- `enrollments`：选课记录映射表
- `waitlists`：各课程的候补队列映射表
- `enrollmentWindow`：选课时间窗口配置
- `mu`：读写锁，保证并发安全
- `idCounter`：ID生成计数器

### 3.2 Student（学生）
**职责**：表示一个学生，包含学生基本信息和选课相关属性。

**字段**：
- `ID`：学生唯一标识
- `Name`：学生姓名
- `Major`：所属专业
- `CompletedCourses`：已完成课程集合（用于先修课校验）
- `MaxCredits`：最大允许选课学分，默认20

### 3.3 Course（课程）
**职责**：表示一门课程，包含课程基本信息和选课限制。

**字段**：
- `ID`：课程唯一标识
- `Name`：课程名称
- `Credits`：课程学分
- `Capacity`：最大选课容量
- `Prerequisite`：先修课程ID列表

### 3.4 Enrollment（选课记录）
**职责**：记录一次选课过程，包含选课状态和时间信息。

**字段**：
- `ID`：记录唯一标识
- `StudentID`：学生ID
- `CourseID`：课程ID
- `Status`：选课状态（已选、已退、候补中）
- `EnrolledAt`：选课/加入候补时间
- `DroppedAt`：退课时间（可为空）

### 3.5 WaitlistEntry（候补给项）
**职责**：表示候补队列中的一条记录。

**字段**：
- `StudentID`：学生ID
- `JoinedAt`：加入候补队列时间

### 3.6 TimeWindow（时间窗口）
**职责**：表示选课开放的时间范围。

**字段**：
- `Start`：窗口开始时间
- `End`：窗口结束时间

## 4. 选课状态流转

```
                    +-------------------+
                    |                   |
                    |   课程未满        |
                    |                   |
                    v                   |
WAITLIST ------> ENROLLED --------> DROPPED
     ^              ^                   |
     |              |                   |
     |              | 课程未满          |
     |              +-------------------+
     |
     | 课程已满
     +-------------------+
                         |
                         | 退课触发递补
                         |
                         v
                  资格校验通过
```

### 状态说明：
- **ENROLLED**：学生已成功选入课程，正常上课状态
- **DROPPED**：学生已退课，选课记录保留用于历史查询
- **WAITLIST**：学生在候补队列中等待递补

### 流转规则：
1. **选课**：课程未满 → ENROLLED；课程已满 → WAITLIST
2. **退课**：ENROLLED → DROPPED，同时触发候补递补
3. **候补递补**：WAITLIST → ENROLLED（资格校验通过后）
4. **退出候补**：WAITLIST → 直接移除，不保留记录

## 5. 错误定义

| 错误变量 | 错误说明 |
|---------|---------|
| `ErrStudentNotFound` | 学生不存在 |
| `ErrCourseNotFound` | 课程不存在 |
| `ErrEnrollmentNotFound` | 选课记录不存在 |
| `ErrCourseFull` | 课程容量已满，进入候补队列 |
| `ErrPrerequisiteNotCompleted` | 未完成先修课程 |
| `ErrAlreadyEnrolled` | 学生已选该课程 |
| `ErrAlreadyInWaitlist` | 学生已在该课程候补队列 |
| `ErrNotEnrolled` | 学生未选该课程 |
| `ErrNotInWaitlist` | 学生不在该课程候补队列 |
| `ErrOutsideTimeWindow` | 操作不在选课时间窗口内 |
| `ErrInvalidTimeRange` | 无效的时间范围（开始时间晚于结束时间） |
| `ErrMaxCreditsExceeded` | 超出最大允许选课学分 |

## 6. 使用示例

### 6.1 初始化选课系统

```go
store := enrollment.NewStore()
```

### 6.2 添加学生

```go
store.AddStudent(&enrollment.Student{
    ID:   "S001",
    Name: "Alice",
    Major: "Computer Science",
    CompletedCourses: map[string]bool{
        "C001": true,
    },
    MaxCredits: 20,
})
```

### 6.3 添加课程

```go
store.AddCourse(&enrollment.Course{
    ID:           "C002",
    Name:         "Data Structures",
    Credits:      4,
    Capacity:     30,
    Prerequisite: []string{"C001"},
})
```

### 6.4 配置选课时间窗口

```go
start := time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2024, 9, 15, 23, 59, 59, 0, time.UTC)
err := store.SetEnrollmentWindow(start, end)
if err != nil {
    log.Fatal(err)
}
```

### 6.5 学生选课

```go
enrollment, err := store.Enroll("S001", "C002")
if err != nil {
    if err == enrollment.ErrCourseFull {
        fmt.Println("课程已满，已加入候补队列")
        fmt.Printf("候补位置: %d\n", enrollment.ID)
    } else {
        log.Fatal(err)
    }
} else {
    fmt.Println("选课成功")
}
```

### 6.6 查询已选课程和学分

```go
courses, err := store.GetEnrolledCourses("S001")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("已选 %d 门课程:\n", len(courses))

credits, err := store.GetStudentCredits("S001")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("已选学分: %d\n", credits)
```

### 6.7 查询课程剩余座位

```go
seats, err := store.GetAvailableSeats("C002")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("剩余座位: %d\n", seats)
```

### 6.8 查询候补位置

```go
position, err := store.GetWaitlistPosition("S003", "C002")
if err != nil {
    if err == enrollment.ErrNotInWaitlist {
        fmt.Println("不在候补队列中")
    } else {
        log.Fatal(err)
    }
} else {
    fmt.Printf("候补位置: %d\n", position)
}
```

### 6.9 学生退课

```go
promoted, err := store.Drop("S001", "C002")
if err != nil {
    log.Fatal(err)
}

if len(promoted) > 0 {
    fmt.Printf("候补学生 %s 已递补成功\n", promoted[0].StudentID)
} else {
    fmt.Println("退课成功，无候补学生需要递补")
}
```

### 6.10 完整使用场景示例

```go
package main

import (
    "fmt"
    "time"
    
    "solocoder-4-go/internal/enrollment"
)

func main() {
    store := enrollment.NewStore()

    store.AddStudent(&enrollment.Student{
        ID:   "S001",
        Name: "Alice",
        CompletedCourses: map[string]bool{"C001": true},
        MaxCredits: 10,
    })
    store.AddStudent(&enrollment.Student{
        ID:   "S002",
        Name: "Bob",
        CompletedCourses: map[string]bool{"C001": true},
        MaxCredits: 20,
    })
    store.AddStudent(&enrollment.Student{
        ID:   "S003",
        Name: "Charlie",
        CompletedCourses: map[string]bool{},
        MaxCredits: 20,
    })

    store.AddCourse(&enrollment.Course{
        ID:           "C001",
        Name:         "Introduction to Programming",
        Credits:      4,
        Capacity:     2,
        Prerequisite: []string{},
    })
    store.AddCourse(&enrollment.Course{
        ID:           "C002",
        Name:         "Data Structures",
        Credits:      4,
        Capacity:     1,
        Prerequisite: []string{"C001"},
    })

    start := time.Now().Add(-1 * time.Hour)
    end := time.Now().Add(1 * time.Hour)
    store.SetEnrollmentWindow(start, end)

    if _, err := store.Enroll("S001", "C002"); err != nil {
        fmt.Println("S001 选课 C002 失败:", err)
    } else {
        fmt.Println("S001 成功选入 C002")
    }

    if _, err := store.Enroll("S002", "C002"); err == enrollment.ErrCourseFull {
        fmt.Println("S002 加入 C002 候补队列")
    }

    if _, err := store.Enroll("S003", "C002"); err == enrollment.ErrCourseFull {
        fmt.Println("S003 加入 C002 候补队列（未完成先修课，递补时会被跳过）")
    }

    credits, _ := store.GetStudentCredits("S001")
    fmt.Printf("S001 当前学分: %d\n", credits)

    promoted, err := store.Drop("S001", "C002")
    if err != nil {
        fmt.Println("退课失败:", err)
    }
    if len(promoted) > 0 {
        fmt.Printf("候补学生 %s 递补成功\n", promoted[0].StudentID)
    }
}
```

## 7. 核心接口列表

| 方法 | 功能说明 |
|-----|---------|
| `NewStore()` | 创建新的选课系统管理器 |
| `AddStudent()` | 添加学生 |
| `AddCourse()` | 添加课程 |
| `GetStudent()` | 获取学生信息 |
| `GetCourse()` | 获取课程信息 |
| `SetEnrollmentWindow()` | 设置选课时间窗口 |
| `Enroll()` | 学生选课（课程满时自动进入候补） |
| `Drop()` | 学生退课（自动触发候补递补） |
| `GetEnrolledCourses()` | 获取学生已选课程列表 |
| `GetStudentCredits()` | 获取学生已选学分总数 |
| `GetAvailableSeats()` | 获取课程剩余可选座位数 |
| `GetWaitlistPosition()` | 获取学生在某课程的候补位置 |
| `GetEnrollments()` | 获取所有选课记录 |
