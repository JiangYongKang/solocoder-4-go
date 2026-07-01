package attendance

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Service struct {
	mu                  sync.RWMutex
	students            map[string]*Student
	classes             map[string]*Class
	rules               map[string]*AttendanceRule
	sessions            map[string]*AttendanceSession
	checkInRecords      map[string]*CheckInRecord
	leaveRequests       map[string]*LeaveRequest
	attendanceRecords   map[string]*AttendanceRecord
	notifications       map[string]*AbsenceNotification
	approvers           map[string]*Student
	idCounter           int64
}

func NewService() *Service {
	return &Service{
		students:          make(map[string]*Student),
		classes:           make(map[string]*Class),
		rules:             make(map[string]*AttendanceRule),
		sessions:          make(map[string]*AttendanceSession),
		checkInRecords:    make(map[string]*CheckInRecord),
		leaveRequests:     make(map[string]*LeaveRequest),
		attendanceRecords: make(map[string]*AttendanceRecord),
		notifications:     make(map[string]*AbsenceNotification),
		approvers:         make(map[string]*Student),
	}
}

func (s *Service) generateID(prefix string) string {
	s.idCounter++
	return fmt.Sprintf("%s%010d", prefix, s.idCounter)
}

func (s *Service) AddStudent(student *Student) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if student.ID == "" {
		student.ID = s.generateID("STU")
	}
	s.students[student.ID] = student
	return student.ID
}

func (s *Service) AddApprover(approver *Student) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if approver.ID == "" {
		approver.ID = s.generateID("APR")
	}
	s.approvers[approver.ID] = approver
	return approver.ID
}

func (s *Service) CreateClass(name, teacher string) *Class {
	s.mu.Lock()
	defer s.mu.Unlock()

	class := &Class{
		ID:       s.generateID("CLS"),
		Name:     name,
		Teacher:  teacher,
		Students: make(map[string]*Student),
	}
	s.classes[class.ID] = class
	return class
}

func (s *Service) AddStudentToClass(classID, studentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	class, exists := s.classes[classID]
	if !exists {
		return ErrClassNotFound
	}

	student, exists := s.students[studentID]
	if !exists {
		return ErrStudentNotFound
	}

	if _, exists := class.Students[studentID]; exists {
		return ErrDuplicateStudentInClass
	}

	class.Students[studentID] = student
	return nil
}

func (s *Service) CreateRule(name string, checkInStart, checkInEnd, checkOutStart, checkOutEnd time.Time, lateThreshold, earlyThreshold int) (*AttendanceRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !checkInStart.Before(checkInEnd) {
		return nil, ErrInvalidTimeRange
	}
	if !checkOutStart.Before(checkOutEnd) {
		return nil, ErrInvalidTimeRange
	}

	rule := &AttendanceRule{
		ID:                       s.generateID("RUL"),
		Name:                     name,
		CheckInStartTime:         checkInStart,
		CheckInEndTime:           checkInEnd,
		CheckOutStartTime:        checkOutStart,
		CheckOutEndTime:          checkOutEnd,
		LateThresholdMinutes:     lateThreshold,
		EarlyLeaveThresholdMinutes: earlyThreshold,
	}
	s.rules[rule.ID] = rule
	return rule, nil
}

func (s *Service) CreateSession(classID, ruleID string, date time.Time) (*AttendanceSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.classes[classID]; !exists {
		return nil, ErrClassNotFound
	}
	if _, exists := s.rules[ruleID]; !exists {
		return nil, fmt.Errorf("rule not found")
	}

	session := &AttendanceSession{
		ID:        s.generateID("SES"),
		ClassID:   classID,
		RuleID:    ruleID,
		Date:      date,
		CreatedAt: time.Now(),
		Active:    true,
	}
	s.sessions[session.ID] = session
	return session, nil
}

func (s *Service) CloseSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}
	session.Active = false
	return nil
}

type CheckInRequest struct {
	StudentID string
	SessionID string
	CheckTime time.Time
}

func (s *Service) CheckIn(req *CheckInRequest) (*CheckInRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}
	if !session.Active {
		return nil, ErrSessionNotActive
	}

	if _, exists := s.students[req.StudentID]; !exists {
		return nil, ErrStudentNotFound
	}

	class, exists := s.classes[session.ClassID]
	if !exists {
		return nil, ErrClassNotFound
	}

	if _, exists := class.Students[req.StudentID]; !exists {
		return nil, ErrStudentNotInClass
	}

	recordKey := req.StudentID + "-" + req.SessionID
	if record, exists := s.checkInRecords[recordKey]; exists {
		if record.CheckInTime != nil {
			return nil, ErrAlreadyCheckedIn
		}
		checkTime := req.CheckTime
		record.CheckInTime = &checkTime
		return record, nil
	}

	checkTime := req.CheckTime
	record := &CheckInRecord{
		StudentID: req.StudentID,
		SessionID: req.SessionID,
		CheckInTime: &checkTime,
	}
	s.checkInRecords[recordKey] = record
	return record, nil
}

func (s *Service) CheckOut(req *CheckInRequest) (*CheckInRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}
	if !session.Active {
		return nil, ErrSessionNotActive
	}

	recordKey := req.StudentID + "-" + req.SessionID
	record, exists := s.checkInRecords[recordKey]
	if !exists || record.CheckInTime == nil {
		return nil, ErrCheckInRequired
	}

	if record.CheckOutTime != nil {
		return nil, ErrAlreadyCheckedOut
	}

	checkTime := req.CheckTime
	record.CheckOutTime = &checkTime
	return record, nil
}

type LeaveRequestRequest struct {
	StudentID string
	SessionID string
	Type      LeaveType
	Reason    string
}

func (s *Service) SubmitLeaveRequest(req *LeaveRequestRequest) (*LeaveRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	student, exists := s.students[req.StudentID]
	if !exists {
		return nil, ErrStudentNotFound
	}

	if _, exists := s.sessions[req.SessionID]; !exists {
		return nil, ErrSessionNotFound
	}

	leave := &LeaveRequest{
		ID:          s.generateID("LEV"),
		StudentID:   req.StudentID,
		StudentName: student.Name,
		SessionID:   req.SessionID,
		Type:        req.Type,
		Reason:      req.Reason,
		Status:      LeaveStatusPending,
		CreatedAt:   time.Now(),
	}
	s.leaveRequests[leave.ID] = leave
	return leave, nil
}

type ProcessLeaveRequest struct {
	LeaveID     string
	ApproverID  string
	Approved    bool
	Comment     string
}

func (s *Service) ProcessLeaveRequest(req *ProcessLeaveRequest) (*LeaveRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	leave, exists := s.leaveRequests[req.LeaveID]
	if !exists {
		return nil, ErrLeaveRequestNotFound
	}

	if leave.Status != LeaveStatusPending {
		return nil, ErrLeaveAlreadyProcessed
	}

	approver, exists := s.approvers[req.ApproverID]
	if !exists {
		return nil, ErrApproverNotFound
	}

	if req.Approved {
		leave.Status = LeaveStatusApproved
	} else {
		leave.Status = LeaveStatusRejected
	}

	leave.ApproverID = approver.ID
	leave.ApproverName = approver.Name
	leave.Comment = req.Comment
	now := time.Now()
	leave.ProcessedAt = &now

	return leave, nil
}

func (s *Service) CalculateAttendance(sessionID string) ([]*AttendanceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	rule, exists := s.rules[session.RuleID]
	if !exists {
		return nil, fmt.Errorf("rule not found")
	}

	class, exists := s.classes[session.ClassID]
	if !exists {
		return nil, ErrClassNotFound
	}

	approvedLeaves := make(map[string]*LeaveRequest)
	for _, leave := range s.leaveRequests {
		if leave.SessionID == sessionID && leave.Status == LeaveStatusApproved {
			approvedLeaves[leave.StudentID] = leave
		}
	}

	var records []*AttendanceRecord
	sessionDate := session.Date

	for studentID, student := range class.Students {
		recordKey := studentID + "-" + sessionID
		checkInRecord := s.checkInRecords[recordKey]

		var status AttendanceStatus
		isLeave := false
		var leaveID string

		if leave, ok := approvedLeaves[studentID]; ok {
			status = StatusLeave
			isLeave = true
			leaveID = leave.ID
		} else if checkInRecord == nil || checkInRecord.CheckInTime == nil {
			status = StatusAbsent
		} else {
			status = s.calculateStatus(rule, sessionDate, checkInRecord)
		}

		record := &AttendanceRecord{
			ID:           s.generateID("REC"),
			StudentID:    studentID,
			StudentName:  student.Name,
			SessionID:    sessionID,
			ClassID:      session.ClassID,
			CheckInTime:  nil,
			CheckOutTime: nil,
			Status:       status,
			IsLeave:      isLeave,
			LeaveID:      leaveID,
			CalculatedAt: time.Now(),
		}

		if checkInRecord != nil {
			record.CheckInTime = checkInRecord.CheckInTime
			record.CheckOutTime = checkInRecord.CheckOutTime
		}

		s.attendanceRecords[record.ID] = record
		records = append(records, record)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].StudentID < records[j].StudentID
	})

	return records, nil
}

func (s *Service) calculateStatus(rule *AttendanceRule, sessionDate time.Time, record *CheckInRecord) AttendanceStatus {
	ruleCheckInStart := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		rule.CheckInStartTime.Hour(), rule.CheckInStartTime.Minute(), rule.CheckInStartTime.Second(), 0, sessionDate.Location())
	ruleCheckInEnd := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		rule.CheckInEndTime.Hour(), rule.CheckInEndTime.Minute(), rule.CheckInEndTime.Second(), 0, sessionDate.Location())
	ruleCheckOutStart := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		rule.CheckOutStartTime.Hour(), rule.CheckOutStartTime.Minute(), rule.CheckOutStartTime.Second(), 0, sessionDate.Location())
	ruleCheckOutEnd := time.Date(sessionDate.Year(), sessionDate.Month(), sessionDate.Day(),
		rule.CheckOutEndTime.Hour(), rule.CheckOutEndTime.Minute(), rule.CheckOutEndTime.Second(), 0, sessionDate.Location())

	lateThreshold := ruleCheckInStart.Add(time.Duration(rule.LateThresholdMinutes) * time.Minute)
	earlyThreshold := ruleCheckOutEnd.Add(-time.Duration(rule.EarlyLeaveThresholdMinutes) * time.Minute)

	checkInTime := *record.CheckInTime
	isLate := checkInTime.After(lateThreshold) && checkInTime.Before(ruleCheckInEnd.Add(time.Nanosecond))

	if record.CheckOutTime == nil {
		if isLate {
			return StatusLate
		}
		return StatusPresent
	}

	checkOutTime := *record.CheckOutTime
	isEarlyLeave := checkOutTime.Before(earlyThreshold) && checkOutTime.After(ruleCheckOutStart.Add(-time.Nanosecond))

	if isLate && isEarlyLeave {
		return StatusEarlyLeave
	} else if isLate {
		return StatusLate
	} else if isEarlyLeave {
		return StatusEarlyLeave
	}

	return StatusPresent
}

func (s *Service) GenerateAbsenceNotifications(sessionID string) ([]*AbsenceNotification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	class, exists := s.classes[session.ClassID]
	if !exists {
		return nil, ErrClassNotFound
	}

	var notifications []*AbsenceNotification

	for _, record := range s.attendanceRecords {
		if record.SessionID != sessionID {
			continue
		}

		if record.Status == StatusAbsent || record.Status == StatusLate || record.Status == StatusEarlyLeave {
			notification := &AbsenceNotification{
				ID:              s.generateID("NOT"),
				StudentID:       record.StudentID,
				StudentName:     record.StudentName,
				ClassID:         record.ClassID,
				ClassName:       class.Name,
				SessionID:       sessionID,
				SessionDate:     session.Date,
				Status:          record.Status,
				NotifiedAt:      time.Now(),
				NotificationSent: true,
			}
			s.notifications[notification.ID] = notification
			notifications = append(notifications, notification)
		}
	}

	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].StudentID < notifications[j].StudentID
	})

	return notifications, nil
}

func (s *Service) GetClassSummary(classID string, startDate, endDate time.Time) (*AttendanceSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !startDate.Before(endDate) {
		return nil, ErrInvalidTimeRange
	}

	class, exists := s.classes[classID]
	if !exists {
		return nil, ErrClassNotFound
	}

	var sessions []*AttendanceSession
	for _, session := range s.sessions {
		if session.ClassID == classID && !session.Date.Before(startDate) && session.Date.Before(endDate.AddDate(0, 0, 1)) {
			sessions = append(sessions, session)
		}
	}

	summary := &AttendanceSummary{
		ClassID:      classID,
		ClassName:    class.Name,
		TotalSessions: len(sessions),
	}

	totalRecords := 0
	for _, session := range sessions {
		for _, record := range s.attendanceRecords {
			if record.SessionID == session.ID {
				totalRecords++
				switch record.Status {
				case StatusPresent:
					summary.PresentCount++
				case StatusLate:
					summary.LateCount++
				case StatusEarlyLeave:
					summary.EarlyLeaveCount++
				case StatusAbsent:
					summary.AbsentCount++
				case StatusLeave:
					summary.LeaveCount++
				}
			}
		}
	}

	if totalRecords > 0 {
		summary.AttendanceRate = float64(summary.PresentCount+summary.LeaveCount) / float64(totalRecords) * 100
	}

	return summary, nil
}

func (s *Service) GetStudentSummary(studentID, classID string, startDate, endDate time.Time) (*AttendanceSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !startDate.Before(endDate) {
		return nil, ErrInvalidTimeRange
	}

	if _, exists := s.students[studentID]; !exists {
		return nil, ErrStudentNotFound
	}

	class, exists := s.classes[classID]
	if !exists {
		return nil, ErrClassNotFound
	}

	student, exists := class.Students[studentID]
	if !exists {
		return nil, ErrStudentNotInClass
	}

	var sessions []*AttendanceSession
	for _, session := range s.sessions {
		if session.ClassID == classID && !session.Date.Before(startDate) && session.Date.Before(endDate.AddDate(0, 0, 1)) {
			sessions = append(sessions, session)
		}
	}

	summary := &AttendanceSummary{
		StudentID:    studentID,
		StudentName:  student.Name,
		ClassID:      classID,
		ClassName:    class.Name,
		TotalSessions: len(sessions),
	}

	for _, session := range sessions {
		for _, record := range s.attendanceRecords {
			if record.SessionID == session.ID && record.StudentID == studentID {
				switch record.Status {
				case StatusPresent:
					summary.PresentCount++
				case StatusLate:
					summary.LateCount++
				case StatusEarlyLeave:
					summary.EarlyLeaveCount++
				case StatusAbsent:
					summary.AbsentCount++
				case StatusLeave:
					summary.LeaveCount++
				}
			}
		}
	}

	if summary.TotalSessions > 0 {
		summary.AttendanceRate = float64(summary.PresentCount+summary.LeaveCount) / float64(summary.TotalSessions) * 100
	}

	return summary, nil
}

func (s *Service) GetStudent(studentID string) (*Student, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	student, exists := s.students[studentID]
	return student, exists
}

func (s *Service) GetClass(classID string) (*Class, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	class, exists := s.classes[classID]
	return class, exists
}

func (s *Service) GetSession(sessionID string) (*AttendanceSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[sessionID]
	return session, exists
}

func (s *Service) GetCheckInRecord(studentID, sessionID string) (*CheckInRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	recordKey := studentID + "-" + sessionID
	record, exists := s.checkInRecords[recordKey]
	return record, exists
}

func (s *Service) GetLeaveRequest(leaveID string) (*LeaveRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	leave, exists := s.leaveRequests[leaveID]
	return leave, exists
}

func (s *Service) GetAttendanceRecord(recordID string) (*AttendanceRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, exists := s.attendanceRecords[recordID]
	return record, exists
}

func (s *Service) ListSessionsByClass(classID string) []*AttendanceSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []*AttendanceSession
	for _, session := range s.sessions {
		if session.ClassID == classID {
			sessions = append(sessions, session)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Date.Before(sessions[j].Date)
	})

	return sessions
}

func (s *Service) ListAttendanceRecords(sessionID string) []*AttendanceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []*AttendanceRecord
	for _, record := range s.attendanceRecords {
		if record.SessionID == sessionID {
			records = append(records, record)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].StudentID < records[j].StudentID
	})

	return records
}

func (s *Service) ListNotifications(sessionID string) []*AbsenceNotification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var notifications []*AbsenceNotification
	for _, notification := range s.notifications {
		if notification.SessionID == sessionID {
			notifications = append(notifications, notification)
		}
	}

	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].StudentID < notifications[j].StudentID
	})

	return notifications
}
