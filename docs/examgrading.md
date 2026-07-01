# 在线考试阅卷模块需求文档

## 1. 模块概述

在线考试阅卷模块实现了从试卷创建发布、考生答题提交、客观题自动评分到成绩复核的完整业务流程。模块使用内存数据结构模拟所有业务实体，支持并发安全的交卷操作和防重复提交机制，确保考试过程的公平性和数据的一致性。

### 主要功能特性

- **试卷发布**：教师可以创建包含客观题的试卷并发布，设置考试时间范围
- **答题提交**：考生在考试时间内提交答卷，答案按题目保存并关联试卷
- **防重复交卷**：同一考生对同一试卷只能成功交卷一次，重复提交返回明确错误
- **自动评分**：交卷后立即进行客观题自动评分，支持单选题、多选题、判断题
- **成绩复核**：考生对成绩有异议时可申请复核，教师处理后可调整成绩
- **并发安全**：基于读写锁实现并发场景下的数据一致性

## 2. 核心结构体职责

### 2.1 基础实体

#### Teacher（教师）
- **职责**：存储教师基本信息，用于试卷创建和复核处理时的身份识别
- **关键字段**：
  - `ID`：教师唯一标识
  - `Name`：教师姓名

#### Student（学生）
- **职责**：存储学生基本信息，用于答题和复核申请时的身份识别
- **关键字段**：
  - `ID`：学生唯一标识
  - `Name`：学生姓名

### 2.2 试卷相关

#### Question（题目）
- **职责**：描述单个考试题目的完整信息，包含题干、选项、分值和正确答案
- **关键字段**：
  - `ID`：题目唯一标识
  - `Type`：题目类型（单选题/多选题/判断题）
  - `Content`：题干内容
  - `Options`：选项列表（适用于选择题）
  - `Score`：题目分值
  - `CorrectAnswer`：正确答案（支持字符串、布尔值、字符串数组）

#### Exam（试卷）
- **职责**：作为核心业务实体，承载试卷的完整信息和状态管理
- **关键字段**：
  - `ID`：试卷唯一标识
  - `Title`：试卷标题
  - `Description`：试卷描述
  - `TeacherID/TeacherName`：出卷教师信息
  - `Questions`：题目列表
  - `StartTime/EndTime`：考试时间范围
  - `TotalScore`：试卷总分
  - `Status`：试卷状态（草稿/已发布）
  - `CreatedAt`：创建时间
  - `PublishedAt`：发布时间

### 2.3 答题相关

#### AnswerSubmission（单题答案）
- **职责**：描述单个题目的作答内容
- **关键字段**：
  - `QuestionID`：关联题目ID
  - `Answer`：考生答案（支持多种类型）

#### ExamSubmission（答卷）
- **职责**：存储考生的完整答卷信息和评分结果
- **关键字段**：
  - `ID`：答卷唯一标识
  - `ExamID`：关联试卷ID
  - `StudentID/StudentName`：考生信息
  - `Answers`：各题目答案（map[题目ID]答案）
  - `SubmittedAt`：提交时间
  - `Score`：得分
  - `TotalScore`：试卷总分
  - `GradedAt`：评分时间
  - `GradeDetails`：各题正误详情（map[题目ID]是否正确）

### 2.4 复核相关

#### ReviewRecord（复核记录）
- **职责**：存储成绩复核申请和处理结果
- **关键字段**：
  - `ID`：复核记录唯一标识
  - `SubmissionID`：关联答卷ID
  - `StudentID/ExamID`：关联学生和试卷
  - `Reason`：复核申请理由
  - `ExpectedScore`：考生期望分数
  - `Status`：复核状态（待处理/通过/驳回）
  - `HandlerID/HandlerName`：处理教师信息
  - `HandlerComment`：处理意见
  - `AdjustedScore`：调整后分数
  - `CreatedAt`：申请时间
  - `ProcessedAt`：处理时间

### 2.5 核心服务

#### Store（内存数据存储）
- **职责**：基于内存的数据存储层，提供所有实体的CRUD操作和业务逻辑
- **核心能力**：
  - 读写锁（sync.RWMutex）保证并发安全
  - 教师/学生的增删查
  - 试卷的创建、发布、查询
  - 答卷的提交、查询、列表
  - 复核记录的申请、处理、查询
  - 自动ID生成（带前缀：TEA/STU/QST/EXM/SUB/REV）
  - 防重复交卷机制（submissionMap）

## 3. 核心业务流程

### 3.1 试卷发布流程

```
1. 教师创建试卷
   ↓
2. 校验：教师存在、题目非空、时间范围合法
   ↓
3. 自动生成题目ID（未指定时）、计算总分
   ↓
4. 试卷状态设为 DRAFT（草稿）
   ↓
5. 教师发布试卷
   ↓
6. 校验：试卷存在
   ↓
7. 试卷状态设为 PUBLISHED（已发布），记录发布时间
```

### 3.2 交卷评分流程

```
1. 考生提交答卷
   ↓
2. 校验：试卷存在且已发布、在考试时间内、考生存在
   ↓
3. 防重复校验：检查 submissionMap 中是否已有该考生-试卷组合
   ↓
4. 校验：所有题目都已作答
   ↓
5. 自动评分：
   - 单选题/判断题：直接比较答案
   - 多选题：排序后比较（不考虑选项顺序）
   - 逐题记录正误详情
   ↓
6. 保存答卷，记录 submissionMap 防止重复提交
   ↓
7. 返回答卷和评分结果
```

### 3.3 成绩复核流程

```
1. 考生申请复核
   ↓
2. 校验：答卷存在、申请人为答卷本人、期望分数合法
   ↓
3. 创建复核记录，状态设为 PENDING（待处理）
   ↓
4. 教师处理复核
   ↓
5. 校验：复核记录存在且未处理、处理人存在
   ↓
6. 处理：
   - 通过：更新 AdjustedScore，同步更新答卷分数
   - 驳回：分数保持不变
   ↓
7. 记录处理人、处理意见、处理时间，更新状态
```

## 4. 状态定义

### 4.1 试卷状态

| 状态 | 常量标识 | 说明 |
|------|----------|------|
| 草稿 | `ExamStatusDraft` | 试卷刚创建，尚未发布 |
| 已发布 | `ExamStatusPublished` | 试卷已发布，考生可以答题 |

### 4.2 复核状态

| 状态 | 常量标识 | 说明 |
|------|----------|------|
| 待处理 | `ReviewStatusPending` | 复核申请已提交，等待处理 |
| 已通过 | `ReviewStatusApproved` | 复核通过，分数已调整 |
| 已驳回 | `ReviewStatusRejected` | 复核被驳回，分数保持不变 |

### 4.3 题目类型

| 类型 | 常量标识 | 说明 |
|------|----------|------|
| 单选题 | `QuestionTypeSingleChoice` | 四个选项选一个 |
| 多选题 | `QuestionTypeMultipleChoice` | 多个选项，答案不考虑顺序 |
| 判断题 | `QuestionTypeTrueFalse` | 对/错两种答案 |

## 5. 防重复交卷机制

### 5.1 实现原理

使用 `submissionMap` 映射表维护"考生ID-试卷ID"组合与答卷ID的对应关系：

```go
submissionMap map[string]string  // key: studentID + "-" + examID, value: submissionID
```

### 5.2 校验流程

1. 提交答卷前，拼接 `studentID + "-" + examID` 作为 key
2. 检查 `submissionMap` 中是否存在该 key
3. 若存在，返回 `ErrAlreadySubmitted` 错误
4. 若不存在，保存答卷后将该 key 写入 `submissionMap`

### 5.3 并发安全

整个交卷过程在 `Store.mu` 写锁保护下执行，确保并发提交时：
- 只有第一个提交成功
- 后续提交全部返回错误
- 不会出现覆盖原答卷的情况

## 6. 自动评分规则

### 6.1 单选题 (SingleChoice)
- 正确答案类型：`string`（如 "A"、"B"、"C"、"D"）
- 比较方式：`reflect.DeepEqual` 精确匹配
- 示例：正确答案 "D"，考生答案 "D" → 正确

### 6.2 判断题 (TrueFalse)
- 正确答案类型：`bool`（true/false）
- 比较方式：`reflect.DeepEqual` 精确匹配
- 示例：正确答案 `true`，考生答案 `true` → 正确

### 6.3 多选题 (MultipleChoice)
- 正确答案类型：`[]string`（如 ["A", "B", "D"]）
- 比较方式：先排序再比较，不考虑选项顺序
- 示例：正确答案 ["A", "B", "D"]，考生答案 ["D", "A", "B"] → 正确
- 示例：正确答案 ["A", "B", "D"]，考生答案 ["A", "B"] → 错误（数量不符）

## 7. 错误处理

### 7.1 错误类型定义

| 错误变量 | 说明 | 触发场景 |
|---------|------|---------|
| `ErrExamNotFound` | 试卷不存在 | 查询/发布/交卷时试卷ID无效 |
| `ErrExamNotPublished` | 试卷未发布 | 对草稿状态的试卷提交答卷 |
| `ErrExamNotStarted` | 考试未开始 | 在考试开始时间前提交答卷 |
| `ErrExamEnded` | 考试已结束 | 在考试结束时间后提交答卷 |
| `ErrStudentNotFound` | 学生不存在 | 交卷/申请复核时学生ID无效 |
| `ErrTeacherNotFound` | 教师不存在 | 创建试卷/处理复核时教师ID无效 |
| `ErrAlreadySubmitted` | 已提交过 | 同一考生对同一试卷重复提交 |

### 7.2 自定义错误类型

#### DuplicateSubmissionError

当检测到重复交卷时，返回 `*DuplicateSubmissionError` 类型的错误，包含已有提交记录的详细信息：

```go
type DuplicateSubmissionError struct {
    ExistingSubmissionID string  // 已有提交记录的ID
    StudentID            string  // 学生ID
    ExamID               string  // 试卷ID
}
```

**特点：**
- 实现了 `error` 接口，错误消息与 `ErrAlreadySubmitted` 一致
- 实现了 `Unwrap() error` 方法，支持 `errors.Is(err, ErrAlreadySubmitted)` 检查
- 调用方可以通过类型断言获取已有提交记录的 ID，便于后续查询

**使用示例：**

```go
_, err := store.SubmitExam(submitReq)
if err != nil {
    if errors.Is(err, ErrAlreadySubmitted) {
        var dupErr *DuplicateSubmissionError
        if errors.As(err, &dupErr) {
            fmt.Printf("已有提交记录ID: %s\n", dupErr.ExistingSubmissionID)
            // 通过返回的ID查询已有提交记录
            existingSubmission, _ := store.GetSubmission(dupErr.ExistingSubmissionID)
        }
    }
}
```
| `ErrSubmissionNotFound` | 答卷不存在 | 查询/申请复核时答卷ID无效 |
| `ErrReviewNotFound` | 复核记录不存在 | 查询/处理复核时记录ID无效 |
| `ErrInvalidTimeRange` | 时间范围非法 | 创建试卷时开始时间≥结束时间 |
| `ErrNoQuestions` | 无题目 | 创建试卷时空题目列表 |
| `ErrInvalidAnswers` | 答案不完整 | 提交答卷时缺少部分题目的答案 |
| `ErrReviewAlreadyProcessed` | 复核已处理 | 对已处理的复核记录重复处理 |
| `ErrInvalidScore` | 分数非法 | 期望分数或调整分数超出范围 |
| `ErrQuestionNotFound` | 题目不存在 | 试卷中找不到指定题目 |

## 8. 使用示例

### 8.1 初始化环境

```go
package main

import (
    "fmt"
    "solocoder-4-go/internal/examgrading"
    "time"
)

func main() {
    store := examgrading.NewStore()

    // 添加教师
    teacherID := store.AddTeacher(&examgrading.Teacher{
        Name: "王老师",
    })

    // 添加学生
    studentID := store.AddStudent(&examgrading.Student{
        Name: "张三",
    })

    student2ID := store.AddStudent(&examgrading.Student{
        Name: "李四",
    })
}
```

### 8.2 完整流程：创建试卷 → 发布 → 答题 → 评分 → 复核

```go
// 1. 教师创建试卷
now := time.Now()
exam, err := store.CreateExam(&examgrading.CreateExamRequest{
    Title:       "Go语言基础测试",
    Description: "期末测试",
    TeacherID:   teacherID,
    Questions: []*examgrading.Question{
        {
            Type:         examgrading.QuestionTypeSingleChoice,
            Content:      "Go语言的创始人是谁？",
            Options:      []string{"A. Ken Thompson", "B. Rob Pike", "C. Robert Griesemer", "D. 以上都是"},
            Score:        20,
            CorrectAnswer: "D",
        },
        {
            Type:         examgrading.QuestionTypeTrueFalse,
            Content:      "Go语言支持垃圾回收。",
            Score:        20,
            CorrectAnswer: true,
        },
        {
            Type:         examgrading.QuestionTypeMultipleChoice,
            Content:      "以下哪些是Go语言的特性？",
            Options:      []string{"A. 协程", "B. 泛型", "C. 指针", "D. 垃圾回收"},
            Score:        30,
            CorrectAnswer: []string{"A", "C", "D"},
        },
        {
            Type:         examgrading.QuestionTypeSingleChoice,
            Content:      "Go语言中启动协程的关键字是？",
            Options:      []string{"A. async", "B. go", "C. start", "D. run"},
            Score:        30,
            CorrectAnswer: "B",
        },
    },
    StartTime: now.Add(-1 * time.Minute),
    EndTime:   now.Add(2 * time.Hour),
})
if err != nil {
    fmt.Printf("创建试卷失败: %v\n", err)
    return
}
fmt.Printf("试卷创建成功，ID: %s，状态: %s，总分: %d\n", 
    exam.ID, exam.Status, exam.TotalScore)

// 2. 发布试卷
published, err := store.PublishExam(exam.ID)
if err != nil {
    fmt.Printf("发布失败: %v\n", err)
    return
}
fmt.Printf("试卷已发布，状态: %s\n", published.Status)

// 3. 考生提交答卷
q1ID := exam.Questions[0].ID
q2ID := exam.Questions[1].ID
q3ID := exam.Questions[2].ID
q4ID := exam.Questions[3].ID

submission, err := store.SubmitExam(&examgrading.SubmitExamRequest{
    ExamID:    exam.ID,
    StudentID: studentID,
    Answers: []examgrading.AnswerSubmission{
        {QuestionID: q1ID, Answer: "D"},
        {QuestionID: q2ID, Answer: true},
        {QuestionID: q3ID, Answer: []string{"A", "C", "D"}},
        {QuestionID: q4ID, Answer: "B"},
    },
})
if err != nil {
    fmt.Printf("交卷失败: %v\n", err)
    return
}
fmt.Printf("交卷成功，得分: %d/%d\n", submission.Score, submission.TotalScore)
fmt.Printf("各题详情: 第1题%v, 第2题%v, 第3题%v, 第4题%v\n",
    submission.GradeDetails[q1ID],
    submission.GradeDetails[q2ID],
    submission.GradeDetails[q3ID],
    submission.GradeDetails[q4ID])

// 4. 尝试重复交卷（应该失败）
_, err = store.SubmitExam(&examgrading.SubmitExamRequest{
    ExamID:    exam.ID,
    StudentID: studentID,
    Answers: []examgrading.AnswerSubmission{
        {QuestionID: q1ID, Answer: "D"},
    },
})
if err == examgrading.ErrAlreadySubmitted {
    fmt.Println("防重复交卷生效，重复提交被拒绝")
}

// 5. 考生申请复核
review, err := store.RequestReview(&examgrading.RequestReviewRequest{
    SubmissionID:  submission.ID,
    StudentID:     studentID,
    Reason:        "认为第2题答案应该是false",
    ExpectedScore: 80,
})
if err != nil {
    fmt.Printf("申请复核失败: %v\n", err)
    return
}
fmt.Printf("复核申请已提交，状态: %s\n", review.Status)

// 6. 教师处理复核（通过）
processed, err := store.ProcessReview(&examgrading.ProcessReviewRequest{
    ReviewID:       review.ID,
    HandlerID:      teacherID,
    Approved:       true,
    HandlerComment: "经过复核，第2题答案确实为false，同意调整分数",
    AdjustedScore:  80,
})
if err != nil {
    fmt.Printf("处理复核失败: %v\n", err)
    return
}
fmt.Printf("复核处理完成，状态: %s，调整后分数: %d\n", 
    processed.Status, processed.AdjustedScore)

// 7. 查看更新后的答卷
updatedSubmission, _ := store.GetSubmission(submission.ID)
fmt.Printf("答卷最终分数: %d\n", updatedSubmission.Score)
```

### 8.3 复核驳回示例

```go
// 教师驳回复核申请
processed, err := store.ProcessReview(&examgrading.ProcessReviewRequest{
    ReviewID:       review.ID,
    HandlerID:      teacherID,
    Approved:       false,
    HandlerComment: "答案正确，无需调整",
})
if err != nil {
    fmt.Printf("处理复核失败: %v\n", err)
    return
}
fmt.Printf("复核已驳回，状态: %s，分数保持: %d\n", 
    processed.Status, processed.AdjustedScore)
```

### 8.4 查询功能示例

```go
// 查询单个试卷
exam, _ := store.GetExam(examID)
fmt.Printf("试卷: %s\n", exam.Title)

// 查询单个答卷
submission, _ := store.GetSubmission(submissionID)
fmt.Printf("得分: %d\n", submission.Score)

// 按学生和试卷查询答卷
submission, _ = store.GetSubmissionByStudentAndExam(studentID, examID)
fmt.Printf("提交时间: %v\n", submission.SubmittedAt)

// 查询复核记录
review, _ := store.GetReview(reviewID)
fmt.Printf("复核状态: %s\n", review.Status)

// 列出教师的所有试卷
exams, _ := store.ListExamsByTeacher(teacherID)
fmt.Printf("教师共创建 %d 份试卷\n", len(exams))

// 列出试卷的所有答卷
submissions, _ := store.ListSubmissionsByExam(examID)
fmt.Printf("试卷共有 %d 份答卷\n", len(submissions))

// 列出答卷的所有复核记录
reviews, _ := store.ListReviewsBySubmission(submissionID)
fmt.Printf("答卷共有 %d 条复核记录\n", len(reviews))
```

### 8.5 考试时间控制示例

```go
// 创建未开始的考试
futureExam, _ := store.CreateExam(&examgrading.CreateExamRequest{
    Title:     "未来考试",
    TeacherID: teacherID,
    Questions: createSampleQuestions(),
    StartTime: time.Now().Add(1 * time.Hour),  // 1小时后开始
    EndTime:   time.Now().Add(3 * time.Hour),
})
store.PublishExam(futureExam.ID)

// 尝试提前交卷（应该失败）
_, err := store.SubmitExam(&examgrading.SubmitExamRequest{
    ExamID:    futureExam.ID,
    StudentID: studentID,
    Answers:   answers,
})
if err == examgrading.ErrExamNotStarted {
    fmt.Println("考试尚未开始，无法交卷")
}

// 创建已结束的考试
pastExam, _ := store.CreateExam(&examgrading.CreateExamRequest{
    Title:     "已结束考试",
    TeacherID: teacherID,
    Questions: createSampleQuestions(),
    StartTime: time.Now().Add(-3 * time.Hour),
    EndTime:   time.Now().Add(-1 * time.Hour),  // 1小时前已结束
})
store.PublishExam(pastExam.ID)

// 尝试延迟交卷（应该失败）
_, err = store.SubmitExam(&examgrading.SubmitExamRequest{
    ExamID:    pastExam.ID,
    StudentID: studentID,
    Answers:   answers,
})
if err == examgrading.ErrExamEnded {
    fmt.Println("考试已结束，无法交卷")
}
```

## 9. 测试覆盖说明

模块包含完整的单元测试（40 个测试用例），覆盖以下场景：

### 9.1 正常流程测试
- `TestCreateExam_Success`：试卷创建成功
- `TestPublishExam_Success`：试卷发布成功
- `TestSubmitExam_Success`：答题提交并自动评分成功（满分）
- `TestSubmitExam_PartialCorrect`：部分正确的评分
- `TestSubmitExam_MultipleChoiceOrder`：多选题答案顺序不影响
- `TestRequestReview_Success`：复核申请成功
- `TestProcessReview_Approved`：复核通过并调整分数
- `TestProcessReview_Rejected`：复核被驳回
- `TestFullWorkflow`：创建→发布→答题→复核完整流程

### 9.2 边界条件测试
- `TestCreateExam_DefaultScore`：题目分值默认值
- `TestCreateExam_NoQuestions`：空题目列表
- `TestCreateExam_InvalidTimeRange`：无效时间范围
- `TestSubmitExam_DuplicateSubmission`：防重复交卷
- `TestSubmitExam_NotPublished`：未发布试卷无法交卷
- `TestSubmitExam_NotStarted`：考试未开始无法交卷
- `TestSubmitExam_Ended`：考试已结束无法交卷
- `TestSubmitExam_MissingAnswers`：缺少部分题目的答案
- `TestRequestReview_InvalidScore`：期望分数超出范围
- `TestProcessReview_InvalidAdjustedScore`：调整分数超出范围
- `TestConcurrentSubmission`：并发提交时的防重复机制
- `TestZeroScore`：所有题目都答错的零分情况
- `TestQuestionIDGeneration`：题目ID自动生成
- `TestDuplicateQuestionID`：重复题目ID校验

### 9.3 异常分支测试
- 所有实体的不存在场景（教师/学生/试卷/答卷/复核）
- 重复处理复核记录
- 非答卷本人申请复核
- 无效的处理人（教师）
- 各种查询操作的不存在场景

## 10. 运行测试

```bash
cd solocoder-4-go
go test ./internal/examgrading/ -v
```

预期输出：所有 40 个测试用例 PASS。
