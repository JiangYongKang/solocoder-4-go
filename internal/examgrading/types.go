package examgrading

import (
	"errors"
	"time"
)

var (
	ErrExamNotFound           = errors.New("exam not found")
	ErrExamNotPublished       = errors.New("exam is not published")
	ErrExamNotStarted         = errors.New("exam has not started yet")
	ErrExamEnded              = errors.New("exam has already ended")
	ErrStudentNotFound        = errors.New("student not found")
	ErrTeacherNotFound        = errors.New("teacher not found")
	ErrAlreadySubmitted       = errors.New("student has already submitted this exam")
	ErrSubmissionNotFound     = errors.New("submission not found")
	ErrReviewNotFound         = errors.New("review record not found")
	ErrInvalidTimeRange       = errors.New("invalid time range: start must be before end")
	ErrNoQuestions            = errors.New("exam must contain at least one question")
	ErrInvalidAnswers         = errors.New("submission must contain answers for all questions")
	ErrAnswerKeyMismatch      = errors.New("answer key must have entries for all questions")
	ErrReviewAlreadyProcessed = errors.New("review has already been processed")
	ErrInvalidScore           = errors.New("invalid score: must be between 0 and total score")
	ErrQuestionNotFound       = errors.New("question not found in exam")
)

type QuestionType string

const (
	QuestionTypeSingleChoice QuestionType = "SINGLE_CHOICE"
	QuestionTypeMultipleChoice QuestionType = "MULTIPLE_CHOICE"
	QuestionTypeTrueFalse    QuestionType = "TRUE_FALSE"
)

type ExamStatus string

const (
	ExamStatusDraft     ExamStatus = "DRAFT"
	ExamStatusPublished ExamStatus = "PUBLISHED"
)

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "PENDING"
	ReviewStatusApproved ReviewStatus = "APPROVED"
	ReviewStatusRejected ReviewStatus = "REJECTED"
)

type Teacher struct {
	ID   string
	Name string
}

type Student struct {
	ID   string
	Name string
}

type Question struct {
	ID           string
	Type         QuestionType
	Content      string
	Options      []string
	Score        int
	CorrectAnswer interface{}
}

type Exam struct {
	ID            string
	Title         string
	Description   string
	TeacherID     string
	TeacherName   string
	Questions     []*Question
	StartTime     time.Time
	EndTime       time.Time
	TotalScore    int
	Status        ExamStatus
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

type AnswerSubmission struct {
	QuestionID string
	Answer     interface{}
}

type ExamSubmission struct {
	ID            string
	ExamID        string
	StudentID     string
	StudentName   string
	Answers       map[string]interface{}
	SubmittedAt   time.Time
	Score         int
	TotalScore    int
	GradedAt      *time.Time
	GradeDetails  map[string]bool
}

type ReviewRecord struct {
	ID             string
	SubmissionID   string
	StudentID      string
	ExamID         string
	Reason         string
	ExpectedScore  int
	Status         ReviewStatus
	HandlerID      string
	HandlerName    string
	HandlerComment string
	AdjustedScore  int
	CreatedAt      time.Time
	ProcessedAt    *time.Time
}
