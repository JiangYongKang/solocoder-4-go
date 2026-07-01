package attendance

import (
	"errors"
	"time"
)

var (
	ErrClassNotFound           = errors.New("class not found")
	ErrStudentNotFound         = errors.New("student not found")
	ErrStudentNotInClass       = errors.New("student not in this class")
	ErrSessionNotFound         = errors.New("attendance session not found")
	ErrSessionNotActive        = errors.New("attendance session is not active")
	ErrAlreadyCheckedIn        = errors.New("student has already checked in")
	ErrAlreadyCheckedOut       = errors.New("student has already checked out")
	ErrCheckInRequired         = errors.New("must check in before checking out")
	ErrLeaveRequestNotFound    = errors.New("leave request not found")
	ErrLeaveAlreadyProcessed   = errors.New("leave request has already been processed")
	ErrInvalidTimeRange        = errors.New("invalid time range: start must be before end")
	ErrApproverNotFound        = errors.New("approver not found")
	ErrAttendanceRecordNotFound = errors.New("attendance record not found")
	ErrDuplicateStudentInClass = errors.New("student is already in this class")
)

type AttendanceStatus string

const (
	StatusPresent           AttendanceStatus = "PRESENT"
	StatusLate              AttendanceStatus = "LATE"
	StatusEarlyLeave        AttendanceStatus = "EARLY_LEAVE"
	StatusLateAndEarlyLeave AttendanceStatus = "LATE_AND_EARLY_LEAVE"
	StatusAbsent            AttendanceStatus = "ABSENT"
	StatusLeave             AttendanceStatus = "LEAVE"
)

type LeaveStatus string

const (
	LeaveStatusPending  LeaveStatus = "PENDING"
	LeaveStatusApproved LeaveStatus = "APPROVED"
	LeaveStatusRejected LeaveStatus = "REJECTED"
)

type LeaveType string

const (
	LeaveTypePersonal LeaveType = "PERSONAL"
	LeaveTypeSick     LeaveType = "SICK"
	LeaveTypeOther    LeaveType = "OTHER"
)

type Student struct {
	ID   string
	Name string
}

type Class struct {
	ID       string
	Name     string
	Teacher  string
	Students map[string]*Student
}

type AttendanceRule struct {
	ID                     string
	Name                   string
	CheckInStartTime       time.Time
	CheckInEndTime         time.Time
	CheckOutStartTime      time.Time
	CheckOutEndTime        time.Time
	LateThresholdMinutes   int
	EarlyLeaveThresholdMinutes int
}

type AttendanceSession struct {
	ID        string
	ClassID   string
	RuleID    string
	Date      time.Time
	CreatedAt time.Time
	Active    bool
}

type CheckInRecord struct {
	StudentID   string
	SessionID   string
	CheckInTime *time.Time
	CheckOutTime *time.Time
}

type LeaveRequest struct {
	ID           string
	StudentID    string
	StudentName  string
	SessionID    string
	Type         LeaveType
	Reason       string
	Status       LeaveStatus
	ApproverID   string
	ApproverName string
	Comment      string
	CreatedAt    time.Time
	ProcessedAt  *time.Time
}

type AttendanceRecord struct {
	ID            string
	StudentID     string
	StudentName   string
	SessionID     string
	ClassID       string
	CheckInTime   *time.Time
	CheckOutTime  *time.Time
	Status        AttendanceStatus
	IsLate        bool
	IsEarlyLeave  bool
	IsLeave       bool
	LeaveID       string
	CalculatedAt  time.Time
}

type AttendanceSummary struct {
	StudentID    string
	StudentName  string
	ClassID      string
	ClassName    string
	TotalSessions int
	PresentCount int
	LateCount    int
	EarlyLeaveCount int
	AbsentCount  int
	LeaveCount   int
	AttendanceRate float64
}

type AbsenceNotification struct {
	ID              string
	StudentID       string
	StudentName     string
	ClassID         string
	ClassName       string
	SessionID       string
	SessionDate     time.Time
	Status          AttendanceStatus
	NotifiedAt      time.Time
	NotificationSent bool
}
