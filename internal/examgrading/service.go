package examgrading

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"
)

type Store struct {
	mu           sync.RWMutex
	teachers     map[string]*Teacher
	students     map[string]*Student
	exams        map[string]*Exam
	submissions  map[string]*ExamSubmission
	submissionMap map[string]string
	reviews      map[string]*ReviewRecord
	idCounter    int64
}

func NewStore() *Store {
	return &Store{
		teachers:     make(map[string]*Teacher),
		students:     make(map[string]*Student),
		exams:        make(map[string]*Exam),
		submissions:  make(map[string]*ExamSubmission),
		submissionMap: make(map[string]string),
		reviews:      make(map[string]*ReviewRecord),
	}
}

func (s *Store) generateID(prefix string) string {
	s.idCounter++
	return fmt.Sprintf("%s%010d", prefix, s.idCounter)
}

func (s *Store) AddTeacher(teacher *Teacher) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if teacher.ID == "" {
		teacher.ID = s.generateID("TEA")
	}
	s.teachers[teacher.ID] = teacher
	return teacher.ID
}

func (s *Store) AddStudent(student *Student) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if student.ID == "" {
		student.ID = s.generateID("STU")
	}
	s.students[student.ID] = student
	return student.ID
}

type CreateExamRequest struct {
	Title       string
	Description string
	TeacherID   string
	Questions   []*Question
	StartTime   time.Time
	EndTime     time.Time
}

func (s *Store) CreateExam(req *CreateExamRequest) (*Exam, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(req.Questions) == 0 {
		return nil, ErrNoQuestions
	}

	if !req.StartTime.Before(req.EndTime) {
		return nil, ErrInvalidTimeRange
	}

	teacher, exists := s.teachers[req.TeacherID]
	if !exists {
		return nil, ErrTeacherNotFound
	}

	questionIDs := make(map[string]bool)
	for _, q := range req.Questions {
		if q.ID == "" {
			q.ID = s.generateID("QST")
		}
		if questionIDs[q.ID] {
			return nil, fmt.Errorf("duplicate question ID: %s", q.ID)
		}
		questionIDs[q.ID] = true
	}

	var totalScore int
	for _, q := range req.Questions {
		if q.Score <= 0 {
			q.Score = 10
		}
		totalScore += q.Score
	}

	exam := &Exam{
		ID:          s.generateID("EXM"),
		Title:       req.Title,
		Description: req.Description,
		TeacherID:   teacher.ID,
		TeacherName: teacher.Name,
		Questions:   req.Questions,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		TotalScore:  totalScore,
		Status:      ExamStatusDraft,
		CreatedAt:   time.Now(),
	}

	s.exams[exam.ID] = exam
	return exam, nil
}

func (s *Store) PublishExam(examID string) (*Exam, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exam, exists := s.exams[examID]
	if !exists {
		return nil, ErrExamNotFound
	}

	now := time.Now()
	exam.Status = ExamStatusPublished
	exam.PublishedAt = &now
	return exam, nil
}

type SubmitExamRequest struct {
	ExamID    string
	StudentID string
	Answers   []AnswerSubmission
}

func (s *Store) SubmitExam(req *SubmitExamRequest) (*ExamSubmission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exam, exists := s.exams[req.ExamID]
	if !exists {
		return nil, ErrExamNotFound
	}

	if exam.Status != ExamStatusPublished {
		return nil, ErrExamNotPublished
	}

	now := time.Now()
	if now.Before(exam.StartTime) {
		return nil, ErrExamNotStarted
	}
	if now.After(exam.EndTime) {
		return nil, ErrExamEnded
	}

	student, exists := s.students[req.StudentID]
	if !exists {
		return nil, ErrStudentNotFound
	}

	submissionKey := req.StudentID + "-" + req.ExamID
	if _, exists := s.submissionMap[submissionKey]; exists {
		return nil, ErrAlreadySubmitted
	}

	answerMap := make(map[string]interface{})
	for _, ans := range req.Answers {
		answerMap[ans.QuestionID] = ans.Answer
	}

	for _, q := range exam.Questions {
		if _, exists := answerMap[q.ID]; !exists {
			return nil, ErrInvalidAnswers
		}
	}

	submission := &ExamSubmission{
		ID:          s.generateID("SUB"),
		ExamID:      exam.ID,
		StudentID:   student.ID,
		StudentName: student.Name,
		Answers:     answerMap,
		SubmittedAt: now,
		TotalScore:  exam.TotalScore,
		GradeDetails: make(map[string]bool),
	}

	score, gradeDetails := s.gradeExam(exam, answerMap)
	submission.Score = score
	submission.GradeDetails = gradeDetails
	submission.GradedAt = &now

	s.submissions[submission.ID] = submission
	s.submissionMap[submissionKey] = submission.ID

	return submission, nil
}

func (s *Store) gradeExam(exam *Exam, answers map[string]interface{}) (int, map[string]bool) {
	var totalScore int
	gradeDetails := make(map[string]bool)

	for _, q := range exam.Questions {
		correct := false
		studentAnswer := answers[q.ID]

		switch q.Type {
		case QuestionTypeSingleChoice, QuestionTypeTrueFalse:
			correct = reflect.DeepEqual(q.CorrectAnswer, studentAnswer)
		case QuestionTypeMultipleChoice:
			correct = compareMultipleChoiceAnswers(q.CorrectAnswer, studentAnswer)
		}

		gradeDetails[q.ID] = correct
		if correct {
			totalScore += q.Score
		}
	}

	return totalScore, gradeDetails
}

func compareMultipleChoiceAnswers(correct, student interface{}) bool {
	correctSlice, ok1 := correct.([]string)
	studentSlice, ok2 := student.([]string)
	if !ok1 || !ok2 {
		return reflect.DeepEqual(correct, student)
	}

	if len(correctSlice) != len(studentSlice) {
		return false
	}

	sort.Strings(correctSlice)
	sort.Strings(studentSlice)

	for i := range correctSlice {
		if correctSlice[i] != studentSlice[i] {
			return false
		}
	}
	return true
}

type RequestReviewRequest struct {
	SubmissionID  string
	StudentID     string
	Reason        string
	ExpectedScore int
}

func (s *Store) RequestReview(req *RequestReviewRequest) (*ReviewRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	submission, exists := s.submissions[req.SubmissionID]
	if !exists {
		return nil, ErrSubmissionNotFound
	}

	if submission.StudentID != req.StudentID {
		return nil, ErrStudentNotFound
	}

	if req.ExpectedScore < 0 || req.ExpectedScore > submission.TotalScore {
		return nil, ErrInvalidScore
	}

	review := &ReviewRecord{
		ID:            s.generateID("REV"),
		SubmissionID:  submission.ID,
		StudentID:     submission.StudentID,
		ExamID:        submission.ExamID,
		Reason:        req.Reason,
		ExpectedScore: req.ExpectedScore,
		Status:        ReviewStatusPending,
		AdjustedScore: submission.Score,
		CreatedAt:     time.Now(),
	}

	s.reviews[review.ID] = review
	return review, nil
}

type ProcessReviewRequest struct {
	ReviewID      string
	HandlerID     string
	Approved      bool
	HandlerComment string
	AdjustedScore int
}

func (s *Store) ProcessReview(req *ProcessReviewRequest) (*ReviewRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	review, exists := s.reviews[req.ReviewID]
	if !exists {
		return nil, ErrReviewNotFound
	}

	if review.Status != ReviewStatusPending {
		return nil, ErrReviewAlreadyProcessed
	}

	handler, exists := s.teachers[req.HandlerID]
	if !exists {
		return nil, ErrTeacherNotFound
	}

	submission, exists := s.submissions[review.SubmissionID]
	if !exists {
		return nil, ErrSubmissionNotFound
	}

	if req.Approved {
		if req.AdjustedScore < 0 || req.AdjustedScore > submission.TotalScore {
			return nil, ErrInvalidScore
		}
		review.Status = ReviewStatusApproved
		review.AdjustedScore = req.AdjustedScore
		submission.Score = req.AdjustedScore
	} else {
		review.Status = ReviewStatusRejected
	}

	review.HandlerID = handler.ID
	review.HandlerName = handler.Name
	review.HandlerComment = req.HandlerComment
	now := time.Now()
	review.ProcessedAt = &now

	return review, nil
}

func (s *Store) GetExam(examID string) (*Exam, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exam, exists := s.exams[examID]
	if !exists {
		return nil, ErrExamNotFound
	}
	return exam, nil
}

func (s *Store) GetSubmission(submissionID string) (*ExamSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	submission, exists := s.submissions[submissionID]
	if !exists {
		return nil, ErrSubmissionNotFound
	}
	return submission, nil
}

func (s *Store) GetSubmissionByStudentAndExam(studentID, examID string) (*ExamSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	submissionKey := studentID + "-" + examID
	submissionID, exists := s.submissionMap[submissionKey]
	if !exists {
		return nil, ErrSubmissionNotFound
	}

	submission, exists := s.submissions[submissionID]
	if !exists {
		return nil, ErrSubmissionNotFound
	}
	return submission, nil
}

func (s *Store) GetReview(reviewID string) (*ReviewRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	review, exists := s.reviews[reviewID]
	if !exists {
		return nil, ErrReviewNotFound
	}
	return review, nil
}

func (s *Store) ListExamsByTeacher(teacherID string) ([]*Exam, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.teachers[teacherID]; !exists {
		return nil, ErrTeacherNotFound
	}

	var result []*Exam
	for _, e := range s.exams {
		if e.TeacherID == teacherID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Store) ListSubmissionsByExam(examID string) ([]*ExamSubmission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.exams[examID]; !exists {
		return nil, ErrExamNotFound
	}

	var result []*ExamSubmission
	for _, sub := range s.submissions {
		if sub.ExamID == examID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (s *Store) ListReviewsBySubmission(submissionID string) ([]*ReviewRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.submissions[submissionID]; !exists {
		return nil, ErrSubmissionNotFound
	}

	var result []*ReviewRecord
	for _, r := range s.reviews {
		if r.SubmissionID == submissionID {
			result = append(result, r)
		}
	}
	return result, nil
}
