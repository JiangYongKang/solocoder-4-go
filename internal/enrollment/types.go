package enrollment

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrStudentNotFound          = errors.New("student not found")
	ErrCourseNotFound           = errors.New("course not found")
	ErrEnrollmentNotFound       = errors.New("enrollment not found")
	ErrCourseFull               = errors.New("course is full")
	ErrPrerequisiteNotCompleted = errors.New("prerequisite courses not completed")
	ErrAlreadyEnrolled          = errors.New("student is already enrolled in this course")
	ErrAlreadyInWaitlist        = errors.New("student is already in the waitlist")
	ErrNotEnrolled              = errors.New("student is not enrolled in this course")
	ErrNotInWaitlist            = errors.New("student is not in the waitlist")
	ErrOutsideTimeWindow        = errors.New("operation is outside the enrollment time window")
	ErrInvalidTimeRange         = errors.New("invalid time range: start must be before end")
	ErrMaxCreditsExceeded       = errors.New("maximum credits exceeded")
)

type EnrollmentStatus string

const (
	StatusEnrolled EnrollmentStatus = "ENROLLED"
	StatusDropped  EnrollmentStatus = "DROPPED"
	StatusWaitlist EnrollmentStatus = "WAITLIST"
)

type Student struct {
	ID                string
	Name              string
	Major             string
	CompletedCourses  map[string]bool
	MaxCredits        int
}

type Course struct {
	ID           string
	Name         string
	Credits      int
	Capacity     int
	Prerequisite []string
}

type Enrollment struct {
	ID         string
	StudentID  string
	CourseID   string
	Status     EnrollmentStatus
	EnrolledAt time.Time
	DroppedAt  *time.Time
}

type WaitlistEntry struct {
	StudentID string
	JoinedAt  time.Time
}

type TimeWindow struct {
	Start time.Time
	End   time.Time
}

type Store struct {
	mu            sync.RWMutex
	students      map[string]*Student
	courses       map[string]*Course
	enrollments   map[string]*Enrollment
	waitlists     map[string][]*WaitlistEntry
	idCounter     int64
	enrollmentWindow *TimeWindow
}

func NewStore() *Store {
	return &Store{
		students:    make(map[string]*Student),
		courses:     make(map[string]*Course),
		enrollments: make(map[string]*Enrollment),
		waitlists:   make(map[string][]*WaitlistEntry),
	}
}

func (s *Store) SetEnrollmentWindow(start, end time.Time) error {
	if !start.Before(end) {
		return ErrInvalidTimeRange
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enrollmentWindow = &TimeWindow{Start: start, End: end}
	return nil
}

func (s *Store) withinEnrollmentWindow() bool {
	if s.enrollmentWindow == nil {
		return true
	}
	now := time.Now()
	return now.After(s.enrollmentWindow.Start) && now.Before(s.enrollmentWindow.End)
}

func (s *Store) AddStudent(student *Student) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if student.CompletedCourses == nil {
		student.CompletedCourses = make(map[string]bool)
	}
	if student.MaxCredits == 0 {
		student.MaxCredits = 20
	}
	s.students[student.ID] = student
}

func (s *Store) AddCourse(course *Course) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.courses[course.ID] = course
}

func (s *Store) GetStudent(studentID string) (*Student, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	student, exists := s.students[studentID]
	return student, exists
}

func (s *Store) GetCourse(courseID string) (*Course, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	course, exists := s.courses[courseID]
	return course, exists
}
