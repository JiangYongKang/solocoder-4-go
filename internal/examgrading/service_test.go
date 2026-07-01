package examgrading

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func setupTestStore() *Store {
	store := NewStore()

	store.AddTeacher(&Teacher{
		Name: "王老师",
	})

	store.AddStudent(&Student{
		Name: "张三",
	})

	store.AddStudent(&Student{
		Name: "李四",
	})

	return store
}

func getFirstID(store *Store, entityType string) string {
	switch entityType {
	case "teacher":
		for id := range store.teachers {
			return id
		}
	case "student":
		for id := range store.students {
			return id
		}
	case "student2":
		keys := make([]string, 0, len(store.students))
		for id := range store.students {
			keys = append(keys, id)
		}
		return keys[1]
	}
	return ""
}

func createSampleQuestions() []*Question {
	return []*Question{
		{
			Type:         QuestionTypeSingleChoice,
			Content:      "Go语言的创始人是谁？",
			Options:      []string{"A. Ken Thompson", "B. Rob Pike", "C. Robert Griesemer", "D. 以上都是"},
			Score:        20,
			CorrectAnswer: "D",
		},
		{
			Type:         QuestionTypeTrueFalse,
			Content:      "Go语言是一种面向对象的编程语言。",
			Score:        20,
			CorrectAnswer: true,
		},
		{
			Type:         QuestionTypeMultipleChoice,
			Content:      "以下哪些是Go语言的特性？",
			Options:      []string{"A. 垃圾回收", "B. 协程", "C. 泛型", "D. 指针"},
			Score:        30,
			CorrectAnswer: []string{"A", "B", "D"},
		},
		{
			Type:         QuestionTypeSingleChoice,
			Content:      "Go语言中用于并发同步的关键字是？",
			Options:      []string{"A. async", "B. await", "C. go", "D. sync"},
			Score:        30,
			CorrectAnswer: "C",
		},
	}
}

func TestCreateExam_Success(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	req := &CreateExamRequest{
		Title:       "Go语言基础测试",
		Description: "测试学生对Go语言基础知识的掌握程度",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now,
		EndTime:     now.Add(2 * time.Hour),
	}

	exam, err := store.CreateExam(req)
	if err != nil {
		t.Fatalf("CreateExam failed: %v", err)
	}

	if exam.ID == "" {
		t.Error("exam ID should not be empty")
	}

	if exam.Title != "Go语言基础测试" {
		t.Errorf("expected title 'Go语言基础测试', got '%s'", exam.Title)
	}

	if exam.TeacherID != teacherID {
		t.Errorf("expected teacher ID %s, got %s", teacherID, exam.TeacherID)
	}

	if exam.TeacherName != "王老师" {
		t.Errorf("expected teacher name '王老师', got '%s'", exam.TeacherName)
	}

	if len(exam.Questions) != 4 {
		t.Errorf("expected 4 questions, got %d", len(exam.Questions))
	}

	if exam.TotalScore != 100 {
		t.Errorf("expected total score 100, got %d", exam.TotalScore)
	}

	if exam.Status != ExamStatusDraft {
		t.Errorf("expected status %v, got %v", ExamStatusDraft, exam.Status)
	}

	if exam.PublishedAt != nil {
		t.Error("published at should be nil for draft exam")
	}
}

func TestCreateExam_NoQuestions(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	req := &CreateExamRequest{
		Title:       "空试卷",
		TeacherID:   teacherID,
		Questions:   []*Question{},
		StartTime:   now,
		EndTime:     now.Add(1 * time.Hour),
	}

	_, err := store.CreateExam(req)
	if err != ErrNoQuestions {
		t.Errorf("expected ErrNoQuestions, got %v", err)
	}
}

func TestCreateExam_InvalidTimeRange(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	req := &CreateExamRequest{
		Title:       "无效时间测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(2 * time.Hour),
		EndTime:     now,
	}

	_, err := store.CreateExam(req)
	if err != ErrInvalidTimeRange {
		t.Errorf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestCreateExam_InvalidTeacher(t *testing.T) {
	store := setupTestStore()

	now := time.Now()
	req := &CreateExamRequest{
		Title:       "无效教师测试",
		TeacherID:   "INVALID_TEACHER",
		Questions:   createSampleQuestions(),
		StartTime:   now,
		EndTime:     now.Add(1 * time.Hour),
	}

	_, err := store.CreateExam(req)
	if err != ErrTeacherNotFound {
		t.Errorf("expected ErrTeacherNotFound, got %v", err)
	}
}

func TestCreateExam_DefaultScore(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	req := &CreateExamRequest{
		Title:     "默认分数测试",
		TeacherID: teacherID,
		Questions: []*Question{
			{
				Type:         QuestionTypeSingleChoice,
				Content:      "测试题",
				Options:      []string{"A", "B"},
				CorrectAnswer: "A",
			},
		},
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}

	exam, err := store.CreateExam(req)
	if err != nil {
		t.Fatalf("CreateExam failed: %v", err)
	}

	if exam.Questions[0].Score != 10 {
		t.Errorf("expected default score 10, got %d", exam.Questions[0].Score)
	}

	if exam.TotalScore != 10 {
		t.Errorf("expected total score 10, got %d", exam.TotalScore)
	}
}

func TestPublishExam_Success(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "待发布测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now,
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)

	published, err := store.PublishExam(exam.ID)
	if err != nil {
		t.Fatalf("PublishExam failed: %v", err)
	}

	if published.Status != ExamStatusPublished {
		t.Errorf("expected status %v, got %v", ExamStatusPublished, published.Status)
	}

	if published.PublishedAt == nil {
		t.Error("published at should not be nil")
	}
}

func TestPublishExam_InvalidExam(t *testing.T) {
	store := setupTestStore()

	_, err := store.PublishExam("INVALID_EXAM")
	if err != ErrExamNotFound {
		t.Errorf("expected ErrExamNotFound, got %v", err)
	}
}

func TestSubmitExam_Success(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "答题测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	q1ID := exam.Questions[0].ID
	q2ID := exam.Questions[1].ID
	q3ID := exam.Questions[2].ID
	q4ID := exam.Questions[3].ID

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: q1ID, Answer: "D"},
			{QuestionID: q2ID, Answer: true},
			{QuestionID: q3ID, Answer: []string{"A", "B", "D"}},
			{QuestionID: q4ID, Answer: "C"},
		},
	}

	submission, err := store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}

	if submission.ID == "" {
		t.Error("submission ID should not be empty")
	}

	if submission.StudentID != studentID {
		t.Errorf("expected student ID %s, got %s", studentID, submission.StudentID)
	}

	if submission.StudentName != "张三" {
		t.Errorf("expected student name '张三', got '%s'", submission.StudentName)
	}

	if submission.ExamID != exam.ID {
		t.Errorf("expected exam ID %s, got %s", exam.ID, submission.ExamID)
	}

	if submission.Score != 100 {
		t.Errorf("expected score 100, got %d", submission.Score)
	}

	if submission.TotalScore != 100 {
		t.Errorf("expected total score 100, got %d", submission.TotalScore)
	}

	if submission.GradedAt == nil {
		t.Error("graded at should not be nil")
	}

	if !submission.GradeDetails[q1ID] {
		t.Error("question 1 should be correct")
	}
	if !submission.GradeDetails[q2ID] {
		t.Error("question 2 should be correct")
	}
	if !submission.GradeDetails[q3ID] {
		t.Error("question 3 should be correct")
	}
	if !submission.GradeDetails[q4ID] {
		t.Error("question 4 should be correct")
	}
}

func TestSubmitExam_PartialCorrect(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "部分正确测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	q1ID := exam.Questions[0].ID
	q2ID := exam.Questions[1].ID
	q3ID := exam.Questions[2].ID
	q4ID := exam.Questions[3].ID

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: q1ID, Answer: "D"},
			{QuestionID: q2ID, Answer: false},
			{QuestionID: q3ID, Answer: []string{"A", "B"}},
			{QuestionID: q4ID, Answer: "C"},
		},
	}

	submission, err := store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}

	expectedScore := 20 + 0 + 0 + 30
	if submission.Score != expectedScore {
		t.Errorf("expected score %d, got %d", expectedScore, submission.Score)
	}

	if !submission.GradeDetails[q1ID] {
		t.Error("question 1 should be correct")
	}
	if submission.GradeDetails[q2ID] {
		t.Error("question 2 should be wrong")
	}
	if submission.GradeDetails[q3ID] {
		t.Error("question 3 should be wrong")
	}
	if !submission.GradeDetails[q4ID] {
		t.Error("question 4 should be correct")
	}
}

func TestSubmitExam_DuplicateSubmission(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "防重复提交测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	_, err := store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("first submission failed: %v", err)
	}

	_, err = store.SubmitExam(submitReq)
	if !errors.Is(err, ErrAlreadySubmitted) {
		t.Errorf("expected ErrAlreadySubmitted, got %v", err)
	}

	submissions, _ := store.ListSubmissionsByExam(exam.ID)
	if len(submissions) != 1 {
		t.Errorf("expected 1 submission, got %d", len(submissions))
	}
}

func TestSubmitExam_NotPublished(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "未发布测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	_, err := store.SubmitExam(submitReq)
	if err != ErrExamNotPublished {
		t.Errorf("expected ErrExamNotPublished, got %v", err)
	}
}

func TestSubmitExam_NotStarted(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "未开始测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(1 * time.Hour),
		EndTime:     now.Add(2 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	_, err := store.SubmitExam(submitReq)
	if err != ErrExamNotStarted {
		t.Errorf("expected ErrExamNotStarted, got %v", err)
	}
}

func TestSubmitExam_Ended(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "已结束测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-2 * time.Hour),
		EndTime:     now.Add(-1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	_, err := store.SubmitExam(submitReq)
	if err != ErrExamEnded {
		t.Errorf("expected ErrExamEnded, got %v", err)
	}
}

func TestSubmitExam_InvalidStudent(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "无效学生测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: "INVALID_STUDENT",
		Answers:   answers,
	}

	_, err := store.SubmitExam(submitReq)
	if err != ErrStudentNotFound {
		t.Errorf("expected ErrStudentNotFound, got %v", err)
	}
}

func TestSubmitExam_InvalidExam(t *testing.T) {
	store := setupTestStore()
	studentID := getFirstID(store, "student")

	submitReq := &SubmitExamRequest{
		ExamID:    "INVALID_EXAM",
		StudentID: studentID,
		Answers:   []AnswerSubmission{},
	}

	_, err := store.SubmitExam(submitReq)
	if err != ErrExamNotFound {
		t.Errorf("expected ErrExamNotFound, got %v", err)
	}
}

func TestSubmitExam_MissingAnswers(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "缺少答案测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: exam.Questions[0].ID, Answer: "A"},
		},
	}

	_, err := store.SubmitExam(submitReq)
	if err != ErrInvalidAnswers {
		t.Errorf("expected ErrInvalidAnswers, got %v", err)
	}
}

func TestSubmitExam_MultipleChoiceOrder(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	questions := []*Question{
		{
			Type:         QuestionTypeMultipleChoice,
			Content:      "多选题测试",
			Options:      []string{"A", "B", "C", "D"},
			Score:        100,
			CorrectAnswer: []string{"A", "B", "C"},
		},
	}

	createReq := &CreateExamRequest{
		Title:       "多选题顺序测试",
		TeacherID:   teacherID,
		Questions:   questions,
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	qID := exam.Questions[0].ID

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: qID, Answer: []string{"C", "A", "B"}},
		},
	}

	submission, err := store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}

	if submission.Score != 100 {
		t.Errorf("expected score 100, got %d", submission.Score)
	}

	if !submission.GradeDetails[qID] {
		t.Error("multiple choice answer should be correct regardless of order")
	}
}

func TestRequestReview_Success(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "复核测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "认为第三题答案有误",
		ExpectedScore: 100,
	}

	review, err := store.RequestReview(reviewReq)
	if err != nil {
		t.Fatalf("RequestReview failed: %v", err)
	}

	if review.ID == "" {
		t.Error("review ID should not be empty")
	}

	if review.SubmissionID != submission.ID {
		t.Errorf("expected submission ID %s, got %s", submission.ID, review.SubmissionID)
	}

	if review.Reason != "认为第三题答案有误" {
		t.Errorf("expected reason, got '%s'", review.Reason)
	}

	if review.ExpectedScore != 100 {
		t.Errorf("expected expected score 100, got %d", review.ExpectedScore)
	}

	if review.Status != ReviewStatusPending {
		t.Errorf("expected status %v, got %v", ReviewStatusPending, review.Status)
	}

	if review.AdjustedScore != submission.Score {
		t.Errorf("expected adjusted score %d, got %d", submission.Score, review.AdjustedScore)
	}
}

func TestRequestReview_InvalidSubmission(t *testing.T) {
	store := setupTestStore()
	studentID := getFirstID(store, "student")

	reviewReq := &RequestReviewRequest{
		SubmissionID:  "INVALID_SUBMISSION",
		StudentID:     studentID,
		Reason:        "测试",
		ExpectedScore: 80,
	}

	_, err := store.RequestReview(reviewReq)
	if err != ErrSubmissionNotFound {
		t.Errorf("expected ErrSubmissionNotFound, got %v", err)
	}
}

func TestRequestReview_WrongStudent(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")
	student2ID := getFirstID(store, "student2")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "错误学生复核测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     student2ID,
		Reason:        "测试",
		ExpectedScore: 80,
	}

	_, err := store.RequestReview(reviewReq)
	if err != ErrStudentNotFound {
		t.Errorf("expected ErrStudentNotFound, got %v", err)
	}
}

func TestRequestReview_InvalidScore(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "无效分数测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "测试",
		ExpectedScore: 200,
	}

	_, err := store.RequestReview(reviewReq)
	if err != ErrInvalidScore {
		t.Errorf("expected ErrInvalidScore, got %v", err)
	}

	reviewReq2 := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "测试",
		ExpectedScore: -10,
	}

	_, err2 := store.RequestReview(reviewReq2)
	if err2 != ErrInvalidScore {
		t.Errorf("expected ErrInvalidScore for negative, got %v", err2)
	}
}

func TestProcessReview_Approved(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "复核通过测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	q1ID := exam.Questions[0].ID
	q2ID := exam.Questions[1].ID
	q3ID := exam.Questions[2].ID
	q4ID := exam.Questions[3].ID

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: q1ID, Answer: "A"},
			{QuestionID: q2ID, Answer: false},
			{QuestionID: q3ID, Answer: []string{"A"}},
			{QuestionID: q4ID, Answer: "C"},
		},
	}

	submission, _ := store.SubmitExam(submitReq)
	originalScore := submission.Score

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "请求加分",
		ExpectedScore: 80,
	}

	review, _ := store.RequestReview(reviewReq)

	processReq := &ProcessReviewRequest{
		ReviewID:      review.ID,
		HandlerID:     teacherID,
		Approved:      true,
		HandlerComment: "同意复核，调整分数",
		AdjustedScore: 80,
	}

	processed, err := store.ProcessReview(processReq)
	if err != nil {
		t.Fatalf("ProcessReview failed: %v", err)
	}

	if processed.Status != ReviewStatusApproved {
		t.Errorf("expected status %v, got %v", ReviewStatusApproved, processed.Status)
	}

	if processed.AdjustedScore != 80 {
		t.Errorf("expected adjusted score 80, got %d", processed.AdjustedScore)
	}

	if processed.HandlerID != teacherID {
		t.Errorf("expected handler ID %s, got %s", teacherID, processed.HandlerID)
	}

	if processed.HandlerName != "王老师" {
		t.Errorf("expected handler name '王老师', got '%s'", processed.HandlerName)
	}

	if processed.ProcessedAt == nil {
		t.Error("processed at should not be nil")
	}

	updatedSubmission, _ := store.GetSubmission(submission.ID)
	if updatedSubmission.Score != 80 {
		t.Errorf("expected submission score updated to 80, got %d", updatedSubmission.Score)
	}

	t.Logf("Original score: %d, Adjusted score: %d", originalScore, 80)
}

func TestProcessReview_Rejected(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "复核驳回测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "请求加分",
		ExpectedScore: 100,
	}

	review, _ := store.RequestReview(reviewReq)

	processReq := &ProcessReviewRequest{
		ReviewID:      review.ID,
		HandlerID:     teacherID,
		Approved:      false,
		HandlerComment: "答案正确，无需调整",
	}

	processed, err := store.ProcessReview(processReq)
	if err != nil {
		t.Fatalf("ProcessReview failed: %v", err)
	}

	if processed.Status != ReviewStatusRejected {
		t.Errorf("expected status %v, got %v", ReviewStatusRejected, processed.Status)
	}

	if processed.AdjustedScore != submission.Score {
		t.Errorf("expected adjusted score to remain %d, got %d", submission.Score, processed.AdjustedScore)
	}

	updatedSubmission, _ := store.GetSubmission(submission.ID)
	if updatedSubmission.Score != submission.Score {
		t.Errorf("submission score should not change, got %d", updatedSubmission.Score)
	}
}

func TestProcessReview_AlreadyProcessed(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "重复处理测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "测试",
		ExpectedScore: 100,
	}

	review, _ := store.RequestReview(reviewReq)

	processReq := &ProcessReviewRequest{
		ReviewID:      review.ID,
		HandlerID:     teacherID,
		Approved:      true,
		AdjustedScore: 100,
	}

	_, err := store.ProcessReview(processReq)
	if err != nil {
		t.Fatalf("first process failed: %v", err)
	}

	_, err = store.ProcessReview(processReq)
	if err != ErrReviewAlreadyProcessed {
		t.Errorf("expected ErrReviewAlreadyProcessed, got %v", err)
	}
}

func TestProcessReview_InvalidReview(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	processReq := &ProcessReviewRequest{
		ReviewID:      "INVALID_REVIEW",
		HandlerID:     teacherID,
		Approved:      true,
		AdjustedScore: 80,
	}

	_, err := store.ProcessReview(processReq)
	if err != ErrReviewNotFound {
		t.Errorf("expected ErrReviewNotFound, got %v", err)
	}
}

func TestProcessReview_InvalidHandler(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "无效处理人测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "测试",
		ExpectedScore: 100,
	}

	review, _ := store.RequestReview(reviewReq)

	processReq := &ProcessReviewRequest{
		ReviewID:      review.ID,
		HandlerID:     "INVALID_HANDLER",
		Approved:      true,
		AdjustedScore: 80,
	}

	_, err := store.ProcessReview(processReq)
	if err != ErrTeacherNotFound {
		t.Errorf("expected ErrTeacherNotFound, got %v", err)
	}
}

func TestProcessReview_InvalidAdjustedScore(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "无效调整分数测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "测试",
		ExpectedScore: 100,
	}

	review, _ := store.RequestReview(reviewReq)

	processReq := &ProcessReviewRequest{
		ReviewID:      review.ID,
		HandlerID:     teacherID,
		Approved:      true,
		AdjustedScore: 200,
	}

	_, err := store.ProcessReview(processReq)
	if err != ErrInvalidScore {
		t.Errorf("expected ErrInvalidScore, got %v", err)
	}
}

func TestGetExam(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "获取试卷测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now,
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)

	fetched, err := store.GetExam(exam.ID)
	if err != nil {
		t.Fatalf("GetExam failed: %v", err)
	}

	if fetched.ID != exam.ID {
		t.Errorf("expected ID %s, got %s", exam.ID, fetched.ID)
	}

	_, err = store.GetExam("NON_EXISTENT")
	if err != ErrExamNotFound {
		t.Errorf("expected ErrExamNotFound, got %v", err)
	}
}

func TestGetSubmission(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "获取答卷测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	fetched, err := store.GetSubmission(submission.ID)
	if err != nil {
		t.Fatalf("GetSubmission failed: %v", err)
	}

	if fetched.ID != submission.ID {
		t.Errorf("expected ID %s, got %s", submission.ID, fetched.ID)
	}

	_, err = store.GetSubmission("NON_EXISTENT")
	if err != ErrSubmissionNotFound {
		t.Errorf("expected ErrSubmissionNotFound, got %v", err)
	}
}

func TestGetSubmissionByStudentAndExam(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "按学生和试卷获取答卷测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	fetched, err := store.GetSubmissionByStudentAndExam(studentID, exam.ID)
	if err != nil {
		t.Fatalf("GetSubmissionByStudentAndExam failed: %v", err)
	}

	if fetched.ID != submission.ID {
		t.Errorf("expected ID %s, got %s", submission.ID, fetched.ID)
	}

	_, err = store.GetSubmissionByStudentAndExam("INVALID_STUDENT", exam.ID)
	if err != ErrSubmissionNotFound {
		t.Errorf("expected ErrSubmissionNotFound, got %v", err)
	}
}

func TestGetReview(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "获取复核测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "测试",
		ExpectedScore: 100,
	}

	review, _ := store.RequestReview(reviewReq)

	fetched, err := store.GetReview(review.ID)
	if err != nil {
		t.Fatalf("GetReview failed: %v", err)
	}

	if fetched.ID != review.ID {
		t.Errorf("expected ID %s, got %s", review.ID, fetched.ID)
	}

	_, err = store.GetReview("NON_EXISTENT")
	if err != ErrReviewNotFound {
		t.Errorf("expected ErrReviewNotFound, got %v", err)
	}
}

func TestListExamsByTeacher(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	for i := 0; i < 3; i++ {
		createReq := &CreateExamRequest{
			Title:     fmt.Sprintf("试卷%d", i+1),
			TeacherID: teacherID,
			Questions: createSampleQuestions(),
			StartTime: now,
			EndTime:   now.Add(1 * time.Hour),
		}
		store.CreateExam(createReq)
	}

	exams, err := store.ListExamsByTeacher(teacherID)
	if err != nil {
		t.Fatalf("ListExamsByTeacher failed: %v", err)
	}

	if len(exams) != 3 {
		t.Errorf("expected 3 exams, got %d", len(exams))
	}

	_, err = store.ListExamsByTeacher("INVALID_TEACHER")
	if err != ErrTeacherNotFound {
		t.Errorf("expected ErrTeacherNotFound, got %v", err)
	}
}

func TestListSubmissionsByExam(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")
	student2ID := getFirstID(store, "student2")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "列出答卷测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	for _, sid := range []string{studentID, student2ID} {
		answers := make([]AnswerSubmission, len(exam.Questions))
		for i, q := range exam.Questions {
			answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
		}

		submitReq := &SubmitExamRequest{
			ExamID:    exam.ID,
			StudentID: sid,
			Answers:   answers,
		}
		store.SubmitExam(submitReq)
	}

	submissions, err := store.ListSubmissionsByExam(exam.ID)
	if err != nil {
		t.Fatalf("ListSubmissionsByExam failed: %v", err)
	}

	if len(submissions) != 2 {
		t.Errorf("expected 2 submissions, got %d", len(submissions))
	}

	_, err = store.ListSubmissionsByExam("INVALID_EXAM")
	if err != ErrExamNotFound {
		t.Errorf("expected ErrExamNotFound, got %v", err)
	}
}

func TestListReviewsBySubmission(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "列出复核测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	submission, _ := store.SubmitExam(submitReq)

	for i := 0; i < 2; i++ {
		reviewReq := &RequestReviewRequest{
			SubmissionID:  submission.ID,
			StudentID:     studentID,
			Reason:        fmt.Sprintf("复核请求%d", i+1),
			ExpectedScore: 100,
		}
		store.RequestReview(reviewReq)
	}

	reviews, err := store.ListReviewsBySubmission(submission.ID)
	if err != nil {
		t.Fatalf("ListReviewsBySubmission failed: %v", err)
	}

	if len(reviews) != 2 {
		t.Errorf("expected 2 reviews, got %d", len(reviews))
	}

	_, err = store.ListReviewsBySubmission("INVALID_SUBMISSION")
	if err != ErrSubmissionNotFound {
		t.Errorf("expected ErrSubmissionNotFound, got %v", err)
	}
}

func TestFullWorkflow(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "Go语言期末测试",
		Description: "期末考试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(2 * time.Hour),
	}

	exam, err := store.CreateExam(createReq)
	if err != nil {
		t.Fatalf("Step 1 - Create exam failed: %v", err)
	}
	if exam.Status != ExamStatusDraft {
		t.Fatalf("Step 1 - Expected status DRAFT, got %v", exam.Status)
	}

	published, err := store.PublishExam(exam.ID)
	if err != nil {
		t.Fatalf("Step 2 - Publish exam failed: %v", err)
	}
	if published.Status != ExamStatusPublished {
		t.Fatalf("Step 2 - Expected status PUBLISHED, got %v", published.Status)
	}

	q1ID := exam.Questions[0].ID
	q2ID := exam.Questions[1].ID
	q3ID := exam.Questions[2].ID
	q4ID := exam.Questions[3].ID

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: q1ID, Answer: "D"},
			{QuestionID: q2ID, Answer: true},
			{QuestionID: q3ID, Answer: []string{"A", "B", "D"}},
			{QuestionID: q4ID, Answer: "A"},
		},
	}

	submission, err := store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("Step 3 - Submit exam failed: %v", err)
	}

	expectedScore := 20 + 20 + 30 + 0
	if submission.Score != expectedScore {
		t.Errorf("Step 3 - Expected score %d, got %d", expectedScore, submission.Score)
	}

	_, err = store.SubmitExam(submitReq)
	if !errors.Is(err, ErrAlreadySubmitted) {
		t.Errorf("Step 4 - Expected ErrAlreadySubmitted for duplicate submission, got %v", err)
	}

	reviewReq := &RequestReviewRequest{
		SubmissionID:  submission.ID,
		StudentID:     studentID,
		Reason:        "第4题答案应该是A，请求复核",
		ExpectedScore: 100,
	}

	review, err := store.RequestReview(reviewReq)
	if err != nil {
		t.Fatalf("Step 5 - Request review failed: %v", err)
	}
	if review.Status != ReviewStatusPending {
		t.Fatalf("Step 5 - Expected status PENDING, got %v", review.Status)
	}

	processReq := &ProcessReviewRequest{
		ReviewID:      review.ID,
		HandlerID:     teacherID,
		Approved:      true,
		HandlerComment: "经过复核，第4题答案确实为A，调整分数",
		AdjustedScore: 100,
	}

	processed, err := store.ProcessReview(processReq)
	if err != nil {
		t.Fatalf("Step 6 - Process review failed: %v", err)
	}
	if processed.Status != ReviewStatusApproved {
		t.Fatalf("Step 6 - Expected status APPROVED, got %v", processed.Status)
	}

	updatedSubmission, _ := store.GetSubmission(submission.ID)
	if updatedSubmission.Score != 100 {
		t.Errorf("Step 7 - Expected submission score updated to 100, got %d", updatedSubmission.Score)
	}

	t.Log("Full workflow completed successfully")
}

func TestConcurrentSubmission(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "并发提交测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	done := make(chan bool, 20)
	var successCount int
	var failCount int

	for i := 0; i < 20; i++ {
		go func() {
			submitReq := &SubmitExamRequest{
				ExamID:    exam.ID,
				StudentID: studentID,
				Answers:   answers,
			}
			_, err := store.SubmitExam(submitReq)
			if err == nil {
				successCount++
			} else {
				failCount++
			}
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful submission, got %d", successCount)
	}

	if failCount != 19 {
		t.Errorf("Expected 19 failed submissions, got %d", failCount)
	}

	submissions, _ := store.ListSubmissionsByExam(exam.ID)
	if len(submissions) != 1 {
		t.Errorf("Expected 1 submission in store, got %d", len(submissions))
	}

	t.Logf("Concurrent test results: %d success, %d failures", successCount, failCount)
}

func TestQuestionIDGeneration(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	questions := []*Question{
		{
			Type:         QuestionTypeSingleChoice,
			Content:      "题1",
			Options:      []string{"A", "B"},
			Score:        50,
			CorrectAnswer: "A",
		},
		{
			ID:           "CUSTOM_QID",
			Type:         QuestionTypeSingleChoice,
			Content:      "题2",
			Options:      []string{"A", "B"},
			Score:        50,
			CorrectAnswer: "B",
		},
	}

	createReq := &CreateExamRequest{
		Title:     "题目ID生成测试",
		TeacherID: teacherID,
		Questions: questions,
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}

	exam, err := store.CreateExam(createReq)
	if err != nil {
		t.Fatalf("CreateExam failed: %v", err)
	}

	if exam.Questions[0].ID == "" {
		t.Error("question 1 should have auto-generated ID")
	}

	if exam.Questions[1].ID != "CUSTOM_QID" {
		t.Errorf("question 2 should retain custom ID, got %s", exam.Questions[1].ID)
	}
}

func TestDuplicateQuestionID(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")

	now := time.Now()
	questions := []*Question{
		{
			ID:           "DUP_QID",
			Type:         QuestionTypeSingleChoice,
			Content:      "题1",
			Options:      []string{"A", "B"},
			Score:        50,
			CorrectAnswer: "A",
		},
		{
			ID:           "DUP_QID",
			Type:         QuestionTypeSingleChoice,
			Content:      "题2",
			Options:      []string{"A", "B"},
			Score:        50,
			CorrectAnswer: "B",
		},
	}

	createReq := &CreateExamRequest{
		Title:     "重复题目ID测试",
		TeacherID: teacherID,
		Questions: questions,
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}

	_, err := store.CreateExam(createReq)
	if err == nil {
		t.Error("expected error for duplicate question ID, got nil")
	}
}

func TestZeroScore(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "零分测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	q1ID := exam.Questions[0].ID
	q2ID := exam.Questions[1].ID
	q3ID := exam.Questions[2].ID
	q4ID := exam.Questions[3].ID

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: q1ID, Answer: "A"},
			{QuestionID: q2ID, Answer: false},
			{QuestionID: q3ID, Answer: []string{"C"}},
			{QuestionID: q4ID, Answer: "A"},
		},
	}

	submission, err := store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}

	if submission.Score != 0 {
		t.Errorf("expected score 0, got %d", submission.Score)
	}

	for qID, correct := range submission.GradeDetails {
		if correct {
			t.Errorf("question %s should be wrong", qID)
		}
	}
}

func TestMultipleChoiceOriginalOrderPreserved(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	originalAnswer := []string{"B", "D", "A", "C"}
	originalCopy := make([]string, len(originalAnswer))
	copy(originalCopy, originalAnswer)

	questions := []*Question{
		{
			Type:         QuestionTypeMultipleChoice,
			Content:      "多选题测试",
			Options:      []string{"A", "B", "C", "D"},
			Score:        100,
			CorrectAnswer: originalAnswer,
		},
	}

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:     "多选题原始顺序测试",
		TeacherID: teacherID,
		Questions: questions,
		StartTime: now.Add(-1 * time.Minute),
		EndTime:   now.Add(1 * time.Hour),
	}

	exam, err := store.CreateExam(createReq)
	if err != nil {
		t.Fatalf("CreateExam failed: %v", err)
	}
	store.PublishExam(exam.ID)

	qID := exam.Questions[0].ID

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers: []AnswerSubmission{
			{QuestionID: qID, Answer: []string{"A", "B", "C", "D"}},
		},
	}

	_, err = store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("SubmitExam failed: %v", err)
	}

	updatedExam, _ := store.GetExam(exam.ID)
	updatedAnswer, ok := updatedExam.Questions[0].CorrectAnswer.([]string)
	if !ok {
		t.Fatal("correct answer should be []string")
	}

	for i := range originalCopy {
		if updatedAnswer[i] != originalCopy[i] {
			t.Errorf("original answer order changed at index %d: expected %s, got %s",
				i, originalCopy[i], updatedAnswer[i])
		}
	}

	t.Logf("Original answer order: %v, After grading: %v", originalCopy, updatedAnswer)

	submitReq2 := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: getFirstID(store, "student2"),
		Answers: []AnswerSubmission{
			{QuestionID: qID, Answer: []string{"C", "B"}},
		},
	}

	_, err = store.SubmitExam(submitReq2)
	if err != nil {
		t.Fatalf("Second SubmitExam failed: %v", err)
	}

	updatedExam2, _ := store.GetExam(exam.ID)
	updatedAnswer2, ok := updatedExam2.Questions[0].CorrectAnswer.([]string)
	if !ok {
		t.Fatal("correct answer should be []string")
	}

	for i := range originalCopy {
		if updatedAnswer2[i] != originalCopy[i] {
			t.Errorf("original answer order changed after second submission at index %d: expected %s, got %s",
				i, originalCopy[i], updatedAnswer2[i])
		}
	}

	t.Logf("After second submission, answer order: %v", updatedAnswer2)
}

func TestDuplicateSubmissionReturnsExistingID(t *testing.T) {
	store := setupTestStore()
	teacherID := getFirstID(store, "teacher")
	studentID := getFirstID(store, "student")

	now := time.Now()
	createReq := &CreateExamRequest{
		Title:       "重复提交返回ID测试",
		TeacherID:   teacherID,
		Questions:   createSampleQuestions(),
		StartTime:   now.Add(-1 * time.Minute),
		EndTime:     now.Add(1 * time.Hour),
	}

	exam, _ := store.CreateExam(createReq)
	store.PublishExam(exam.ID)

	answers := make([]AnswerSubmission, len(exam.Questions))
	for i, q := range exam.Questions {
		answers[i] = AnswerSubmission{QuestionID: q.ID, Answer: q.CorrectAnswer}
	}

	submitReq := &SubmitExamRequest{
		ExamID:    exam.ID,
		StudentID: studentID,
		Answers:   answers,
	}

	firstSubmission, err := store.SubmitExam(submitReq)
	if err != nil {
		t.Fatalf("First submission failed: %v", err)
	}

	_, err = store.SubmitExam(submitReq)
	if err == nil {
		t.Fatal("expected error for duplicate submission, got nil")
	}

	var dupErr *DuplicateSubmissionError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected *DuplicateSubmissionError, got %T", err)
	}

	if dupErr.ExistingSubmissionID != firstSubmission.ID {
		t.Errorf("expected existing submission ID %s, got %s",
			firstSubmission.ID, dupErr.ExistingSubmissionID)
	}

	if dupErr.StudentID != studentID {
		t.Errorf("expected student ID %s, got %s", studentID, dupErr.StudentID)
	}

	if dupErr.ExamID != exam.ID {
		t.Errorf("expected exam ID %s, got %s", exam.ID, dupErr.ExamID)
	}

	if dupErr.Error() != ErrAlreadySubmitted.Error() {
		t.Errorf("expected error message '%s', got '%s'",
			ErrAlreadySubmitted.Error(), dupErr.Error())
	}

	if !errors.Is(err, ErrAlreadySubmitted) {
		t.Error("expected error to wrap ErrAlreadySubmitted")
	}

	fetchedSubmission, err := store.GetSubmission(dupErr.ExistingSubmissionID)
	if err != nil {
		t.Fatalf("Failed to fetch submission using returned ID: %v", err)
	}

	if fetchedSubmission.ID != firstSubmission.ID {
		t.Errorf("fetched submission ID mismatch: expected %s, got %s",
			firstSubmission.ID, fetchedSubmission.ID)
	}

	t.Logf("Duplicate submission correctly returned existing ID: %s", dupErr.ExistingSubmissionID)
}
