package attendance

import (
	"sort"
	"testing"
	"time"
)

func setupTestService() *Service {
	service := NewService()

	service.AddStudent(&Student{Name: "张三"})
	service.AddStudent(&Student{Name: "李四"})
	service.AddStudent(&Student{Name: "王五"})
	service.AddApprover(&Student{Name: "王老师"})

	return service
}

func getSortedStudentIDs(service *Service) []string {
	keys := make([]string, 0, len(service.students))
	for id := range service.students {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	return keys
}

func getFirstID(service *Service, entityType string) string {
	switch entityType {
	case "student":
		ids := getSortedStudentIDs(service)
		if len(ids) > 0 {
			return ids[0]
		}
	case "student2":
		ids := getSortedStudentIDs(service)
		if len(ids) > 1 {
			return ids[1]
		}
	case "student3":
		ids := getSortedStudentIDs(service)
		if len(ids) > 2 {
			return ids[2]
		}
	case "approver":
		for id := range service.approvers {
			return id
		}
	}
	return ""
}

func createDefaultRule(t *testing.T, service *Service) *AttendanceRule {
	t.Helper()

	checkInStart := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	checkInEnd := time.Date(2024, 1, 1, 8, 30, 0, 0, time.UTC)
	checkOutStart := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)
	checkOutEnd := time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)

	rule, err := service.CreateRule("默认考勤规则", checkInStart, checkInEnd, checkOutStart, checkOutEnd, 10, 10)
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}
	return rule
}

func setupClassWithStudents(t *testing.T, service *Service) (string, string) {
	t.Helper()

	class := service.CreateClass("计算机科学1班", "王老师")
	studentID1 := getFirstID(service, "student")
	studentID2 := getFirstID(service, "student2")
	studentID3 := getFirstID(service, "student3")

	err := service.AddStudentToClass(class.ID, studentID1)
	if err != nil {
		t.Fatalf("AddStudentToClass failed: %v", err)
	}
	err = service.AddStudentToClass(class.ID, studentID2)
	if err != nil {
		t.Fatalf("AddStudentToClass failed: %v", err)
	}
	err = service.AddStudentToClass(class.ID, studentID3)
	if err != nil {
		t.Fatalf("AddStudentToClass failed: %v", err)
	}

	return class.ID, studentID1
}

func setupSession(t *testing.T, service *Service, classID, ruleID string) *AttendanceSession {
	t.Helper()

	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	session, err := service.CreateSession(classID, ruleID, date)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	return session
}

func TestAddStudent(t *testing.T) {
	service := NewService()

	studentID := service.AddStudent(&Student{Name: "测试学生"})
	if studentID == "" {
		t.Error("student ID should not be empty")
	}

	student, exists := service.GetStudent(studentID)
	if !exists {
		t.Fatal("student should exist")
	}
	if student.Name != "测试学生" {
		t.Errorf("expected name '测试学生', got '%s'", student.Name)
	}
}

func TestCreateClass(t *testing.T) {
	service := NewService()

	class := service.CreateClass("测试班级", "测试老师")
	if class.ID == "" {
		t.Error("class ID should not be empty")
	}
	if class.Name != "测试班级" {
		t.Errorf("expected name '测试班级', got '%s'", class.Name)
	}
	if class.Teacher != "测试老师" {
		t.Errorf("expected teacher '测试老师', got '%s'", class.Teacher)
	}
	if class.Students == nil {
		t.Error("students map should be initialized")
	}
}

func TestAddStudentToClass_Success(t *testing.T) {
	service := setupTestService()
	class := service.CreateClass("测试班级", "测试老师")
	studentID := getFirstID(service, "student")

	err := service.AddStudentToClass(class.ID, studentID)
	if err != nil {
		t.Fatalf("AddStudentToClass failed: %v", err)
	}

	classFromGet, _ := service.GetClass(class.ID)
	if _, exists := classFromGet.Students[studentID]; !exists {
		t.Error("student should be in class")
	}
}

func TestAddStudentToClass_ClassNotFound(t *testing.T) {
	service := setupTestService()
	studentID := getFirstID(service, "student")

	err := service.AddStudentToClass("invalid-class", studentID)
	if err != ErrClassNotFound {
		t.Errorf("expected ErrClassNotFound, got %v", err)
	}
}

func TestAddStudentToClass_StudentNotFound(t *testing.T) {
	service := setupTestService()
	class := service.CreateClass("测试班级", "测试老师")

	err := service.AddStudentToClass(class.ID, "invalid-student")
	if err != ErrStudentNotFound {
		t.Errorf("expected ErrStudentNotFound, got %v", err)
	}
}

func TestAddStudentToClass_DuplicateStudent(t *testing.T) {
	service := setupTestService()
	class := service.CreateClass("测试班级", "测试老师")
	studentID := getFirstID(service, "student")

	err := service.AddStudentToClass(class.ID, studentID)
	if err != nil {
		t.Fatalf("first AddStudentToClass failed: %v", err)
	}

	err = service.AddStudentToClass(class.ID, studentID)
	if err != ErrDuplicateStudentInClass {
		t.Errorf("expected ErrDuplicateStudentInClass, got %v", err)
	}
}

func TestCreateRule_Success(t *testing.T) {
	service := NewService()

	checkInStart := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	checkInEnd := time.Date(2024, 1, 1, 8, 30, 0, 0, time.UTC)
	checkOutStart := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)
	checkOutEnd := time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)

	rule, err := service.CreateRule("测试规则", checkInStart, checkInEnd, checkOutStart, checkOutEnd, 10, 10)
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	if rule.ID == "" {
		t.Error("rule ID should not be empty")
	}
	if rule.LateThresholdMinutes != 10 {
		t.Errorf("expected LateThresholdMinutes 10, got %d", rule.LateThresholdMinutes)
	}
}

func TestCreateRule_InvalidTimeRange(t *testing.T) {
	service := NewService()

	checkInStart := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	checkInEnd := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	checkOutStart := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)
	checkOutEnd := time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)

	_, err := service.CreateRule("测试规则", checkInStart, checkInEnd, checkOutStart, checkOutEnd, 10, 10)
	if err != ErrInvalidTimeRange {
		t.Errorf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestCreateSession_Success(t *testing.T) {
	service := setupTestService()
	classID, _ := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)

	session := setupSession(t, service, classID, rule.ID)

	if session.ClassID != classID {
		t.Errorf("expected class ID %s, got %s", classID, session.ClassID)
	}
	if session.RuleID != rule.ID {
		t.Errorf("expected rule ID %s, got %s", rule.ID, session.RuleID)
	}
	if !session.Active {
		t.Error("session should be active")
	}
}

func TestCreateSession_ClassNotFound(t *testing.T) {
	service := NewService()
	rule := createDefaultRule(t, service)
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	_, err := service.CreateSession("invalid-class", rule.ID, date)
	if err != ErrClassNotFound {
		t.Errorf("expected ErrClassNotFound, got %v", err)
	}
}

func TestCheckIn_Success(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	record, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkTime,
	})

	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}
	if record.CheckInTime == nil {
		t.Fatal("CheckInTime should not be nil")
	}
	if !record.CheckInTime.Equal(checkTime) {
		t.Errorf("expected check time %v, got %v", checkTime, record.CheckInTime)
	}
}

func TestCheckIn_AlreadyCheckedIn(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkTime,
	})
	if err != nil {
		t.Fatalf("first CheckIn failed: %v", err)
	}

	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkTime,
	})
	if err != ErrAlreadyCheckedIn {
		t.Errorf("expected ErrAlreadyCheckedIn, got %v", err)
	}
}

func TestCheckIn_SessionNotFound(t *testing.T) {
	service := setupTestService()
	studentID := getFirstID(service, "student")

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: "invalid-session",
		CheckTime: time.Now(),
	})
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestCheckIn_SessionNotActive(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	err := service.CloseSession(session.ID)
	if err != nil {
		t.Fatalf("CloseSession failed: %v", err)
	}

	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: time.Now(),
	})
	if err != ErrSessionNotActive {
		t.Errorf("expected ErrSessionNotActive, got %v", err)
	}
}

func TestCheckIn_StudentNotInClass(t *testing.T) {
	service := setupTestService()
	classID, _ := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	newStudent := service.AddStudent(&Student{Name: "新学生"})

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: newStudent,
		SessionID: session.ID,
		CheckTime: time.Now(),
	})
	if err != ErrStudentNotInClass {
		t.Errorf("expected ErrStudentNotInClass, got %v", err)
	}
}

func TestCheckOut_Success(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	checkOutTime := time.Date(2024, 6, 1, 17, 30, 0, 0, time.UTC)
	record, err := service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	if record.CheckOutTime == nil {
		t.Fatal("CheckOutTime should not be nil")
	}
	if !record.CheckOutTime.Equal(checkOutTime) {
		t.Errorf("expected check out time %v, got %v", checkOutTime, record.CheckOutTime)
	}
}

func TestCheckOut_CheckInRequired(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	_, err := service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: time.Now(),
	})
	if err != ErrCheckInRequired {
		t.Errorf("expected ErrCheckInRequired, got %v", err)
	}
}

func TestCheckOut_AlreadyCheckedOut(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	checkOutTime := time.Date(2024, 6, 1, 17, 30, 0, 0, time.UTC)
	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("first CheckOut failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != ErrAlreadyCheckedOut {
		t.Errorf("expected ErrAlreadyCheckedOut, got %v", err)
	}
}

func TestSubmitLeaveRequest_Success(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	leave, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: studentID,
		SessionID: session.ID,
		Type:      LeaveTypeSick,
		Reason:    "身体不适，需要请假",
	})
	if err != nil {
		t.Fatalf("SubmitLeaveRequest failed: %v", err)
	}

	if leave.ID == "" {
		t.Error("leave ID should not be empty")
	}
	if leave.Status != LeaveStatusPending {
		t.Errorf("expected status PENDING, got %s", leave.Status)
	}
	if leave.Type != LeaveTypeSick {
		t.Errorf("expected type SICK, got %s", leave.Type)
	}
	if leave.Reason != "身体不适，需要请假" {
		t.Errorf("expected reason mismatch")
	}
}

func TestSubmitLeaveRequest_StudentNotFound(t *testing.T) {
	service := setupTestService()
	classID, _ := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	_, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: "invalid-student",
		SessionID: session.ID,
		Type:      LeaveTypeSick,
		Reason:    "请假",
	})
	if err != ErrStudentNotFound {
		t.Errorf("expected ErrStudentNotFound, got %v", err)
	}
}

func TestProcessLeaveRequest_Approved(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)
	approverID := getFirstID(service, "approver")

	leave, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: studentID,
		SessionID: session.ID,
		Type:      LeaveTypeSick,
		Reason:    "身体不适",
	})
	if err != nil {
		t.Fatalf("SubmitLeaveRequest failed: %v", err)
	}

	processed, err := service.ProcessLeaveRequest(&ProcessLeaveRequest{
		LeaveID:    leave.ID,
		ApproverID: approverID,
		Approved:   true,
		Comment:    "批准请假",
	})
	if err != nil {
		t.Fatalf("ProcessLeaveRequest failed: %v", err)
	}

	if processed.Status != LeaveStatusApproved {
		t.Errorf("expected status APPROVED, got %s", processed.Status)
	}
	if processed.Comment != "批准请假" {
		t.Errorf("expected comment '批准请假', got '%s'", processed.Comment)
	}
	if processed.ProcessedAt == nil {
		t.Error("ProcessedAt should not be nil")
	}
}

func TestProcessLeaveRequest_Rejected(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)
	approverID := getFirstID(service, "approver")

	leave, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: studentID,
		SessionID: session.ID,
		Type:      LeaveTypePersonal,
		Reason:    "有事",
	})
	if err != nil {
		t.Fatalf("SubmitLeaveRequest failed: %v", err)
	}

	processed, err := service.ProcessLeaveRequest(&ProcessLeaveRequest{
		LeaveID:    leave.ID,
		ApproverID: approverID,
		Approved:   false,
		Comment:    "理由不充分，不予批准",
	})
	if err != nil {
		t.Fatalf("ProcessLeaveRequest failed: %v", err)
	}

	if processed.Status != LeaveStatusRejected {
		t.Errorf("expected status REJECTED, got %s", processed.Status)
	}
}

func TestProcessLeaveRequest_AlreadyProcessed(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)
	approverID := getFirstID(service, "approver")

	leave, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: studentID,
		SessionID: session.ID,
		Type:      LeaveTypeSick,
		Reason:    "身体不适",
	})
	if err != nil {
		t.Fatalf("SubmitLeaveRequest failed: %v", err)
	}

	_, err = service.ProcessLeaveRequest(&ProcessLeaveRequest{
		LeaveID:    leave.ID,
		ApproverID: approverID,
		Approved:   true,
		Comment:    "批准",
	})
	if err != nil {
		t.Fatalf("first ProcessLeaveRequest failed: %v", err)
	}

	_, err = service.ProcessLeaveRequest(&ProcessLeaveRequest{
		LeaveID:    leave.ID,
		ApproverID: approverID,
		Approved:   true,
		Comment:    "再次处理",
	})
	if err != ErrLeaveAlreadyProcessed {
		t.Errorf("expected ErrLeaveAlreadyProcessed, got %v", err)
	}
}

func TestProcessLeaveRequest_ApproverNotFound(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	leave, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: studentID,
		SessionID: session.ID,
		Type:      LeaveTypeSick,
		Reason:    "身体不适",
	})
	if err != nil {
		t.Fatalf("SubmitLeaveRequest failed: %v", err)
	}

	_, err = service.ProcessLeaveRequest(&ProcessLeaveRequest{
		LeaveID:    leave.ID,
		ApproverID: "invalid-approver",
		Approved:   true,
		Comment:    "批准",
	})
	if err != ErrApproverNotFound {
		t.Errorf("expected ErrApproverNotFound, got %v", err)
	}
}

func TestCalculateAttendance_Present(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	checkOutTime := time.Date(2024, 6, 1, 17, 55, 0, 0, time.UTC)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	var studentRecord *AttendanceRecord
	for _, r := range records {
		if r.StudentID == studentID {
			studentRecord = r
			break
		}
	}

	if studentRecord == nil {
		t.Fatal("student record not found")
	}

	if studentRecord.Status != StatusPresent {
		t.Errorf("expected status PRESENT, got %s", studentRecord.Status)
	}
	if studentRecord.IsLeave {
		t.Error("IsLeave should be false")
	}
}

func TestCalculateAttendance_Late(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	var studentRecord *AttendanceRecord
	for _, r := range records {
		if r.StudentID == studentID {
			studentRecord = r
			break
		}
	}

	if studentRecord == nil {
		t.Fatal("student record not found")
	}

	if studentRecord.Status != StatusLate {
		t.Errorf("expected status LATE, got %s", studentRecord.Status)
	}
}

func TestCalculateAttendance_EarlyLeave(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	checkOutTime := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	var studentRecord *AttendanceRecord
	for _, r := range records {
		if r.StudentID == studentID {
			studentRecord = r
			break
		}
	}

	if studentRecord == nil {
		t.Fatal("student record not found")
	}

	if studentRecord.Status != StatusEarlyLeave {
		t.Errorf("expected status EARLY_LEAVE, got %s", studentRecord.Status)
	}
}

func TestCalculateAttendance_Absent(t *testing.T) {
	service := setupTestService()
	classID, _ := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	absentCount := 0
	for _, r := range records {
		if r.Status == StatusAbsent {
			absentCount++
		}
	}

	if absentCount != 3 {
		t.Errorf("expected 3 absent records, got %d", absentCount)
	}
}

func TestCalculateAttendance_Leave(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)
	approverID := getFirstID(service, "approver")

	leave, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: studentID,
		SessionID: session.ID,
		Type:      LeaveTypeSick,
		Reason:    "身体不适",
	})
	if err != nil {
		t.Fatalf("SubmitLeaveRequest failed: %v", err)
	}

	_, err = service.ProcessLeaveRequest(&ProcessLeaveRequest{
		LeaveID:    leave.ID,
		ApproverID: approverID,
		Approved:   true,
		Comment:    "批准",
	})
	if err != nil {
		t.Fatalf("ProcessLeaveRequest failed: %v", err)
	}

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	var studentRecord *AttendanceRecord
	for _, r := range records {
		if r.StudentID == studentID {
			studentRecord = r
			break
		}
	}

	if studentRecord == nil {
		t.Fatal("student record not found")
	}

	if studentRecord.Status != StatusLeave {
		t.Errorf("expected status LEAVE, got %s", studentRecord.Status)
	}
	if !studentRecord.IsLeave {
		t.Error("IsLeave should be true")
	}
	if studentRecord.LeaveID != leave.ID {
		t.Errorf("expected leave ID %s, got %s", leave.ID, studentRecord.LeaveID)
	}
}

func TestCalculateAttendance_RejectedLeaveAsAbsent(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)
	approverID := getFirstID(service, "approver")

	leave, err := service.SubmitLeaveRequest(&LeaveRequestRequest{
		StudentID: studentID,
		SessionID: session.ID,
		Type:      LeaveTypePersonal,
		Reason:    "有事",
	})
	if err != nil {
		t.Fatalf("SubmitLeaveRequest failed: %v", err)
	}

	_, err = service.ProcessLeaveRequest(&ProcessLeaveRequest{
		LeaveID:    leave.ID,
		ApproverID: approverID,
		Approved:   false,
		Comment:    "不批准",
	})
	if err != nil {
		t.Fatalf("ProcessLeaveRequest failed: %v", err)
	}

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	var studentRecord *AttendanceRecord
	for _, r := range records {
		if r.StudentID == studentID {
			studentRecord = r
			break
		}
	}

	if studentRecord == nil {
		t.Fatal("student record not found")
	}

	if studentRecord.Status != StatusAbsent {
		t.Errorf("expected status ABSENT for rejected leave, got %s", studentRecord.Status)
	}
}

func TestGenerateAbsenceNotifications(t *testing.T) {
	service := setupTestService()
	classID, studentID1 := setupClassWithStudents(t, service)
	studentID2 := getFirstID(service, "student2")
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime1 := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session.ID,
		CheckTime: checkInTime1,
	})
	if err != nil {
		t.Fatalf("CheckIn for student1 failed: %v", err)
	}

	checkInTime2 := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	checkOutTime2 := time.Date(2024, 6, 1, 17, 55, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID2,
		SessionID: session.ID,
		CheckTime: checkInTime2,
	})
	if err != nil {
		t.Fatalf("CheckIn for student2 failed: %v", err)
	}
	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID2,
		SessionID: session.ID,
		CheckTime: checkOutTime2,
	})
	if err != nil {
		t.Fatalf("CheckOut for student2 failed: %v", err)
	}

	_, err = service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	notifications, err := service.GenerateAbsenceNotifications(session.ID)
	if err != nil {
		t.Fatalf("GenerateAbsenceNotifications failed: %v", err)
	}

	if len(notifications) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(notifications))
	}

	lateCount := 0
	absentCount := 0
	for _, n := range notifications {
		if n.Status == StatusLate {
			lateCount++
		} else if n.Status == StatusAbsent {
			absentCount++
		}
		if !n.NotificationSent {
			t.Error("NotificationSent should be true")
		}
	}

	if lateCount != 1 {
		t.Errorf("expected 1 late notification, got %d", lateCount)
	}
	if absentCount != 1 {
		t.Errorf("expected 1 absent notification, got %d", absentCount)
	}
}

func TestGetStudentSummary(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)
	date3 := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	session1, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession 1 failed: %v", err)
	}
	session2, err := service.CreateSession(classID, rule.ID, date2)
	if err != nil {
		t.Fatalf("CreateSession 2 failed: %v", err)
	}
	session3, err := service.CreateSession(classID, rule.ID, date3)
	if err != nil {
		t.Fatalf("CreateSession 3 failed: %v", err)
	}

	checkInTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session1.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn session1 failed: %v", err)
	}

	checkInTime2 := time.Date(2024, 6, 2, 8, 15, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session2.ID,
		CheckTime: checkInTime2,
	})
	if err != nil {
		t.Fatalf("CheckIn session2 failed: %v", err)
	}

	_, err = service.CalculateAttendance(session1.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 1 failed: %v", err)
	}
	_, err = service.CalculateAttendance(session2.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 2 failed: %v", err)
	}
	_, err = service.CalculateAttendance(session3.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 3 failed: %v", err)
	}

	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	summary, err := service.GetStudentSummary(studentID, classID, startDate, endDate)
	if err != nil {
		t.Fatalf("GetStudentSummary failed: %v", err)
	}

	if summary.TotalSessions != 3 {
		t.Errorf("expected 3 total sessions, got %d", summary.TotalSessions)
	}
	if summary.PresentCount != 1 {
		t.Errorf("expected 1 present, got %d", summary.PresentCount)
	}
	if summary.LateCount != 1 {
		t.Errorf("expected 1 late, got %d", summary.LateCount)
	}
	if summary.AbsentCount != 1 {
		t.Errorf("expected 1 absent, got %d", summary.AbsentCount)
	}

	expectedRate := float64(2) / float64(3) * 100
	if summary.AttendanceRate != expectedRate {
		t.Errorf("expected attendance rate %.2f, got %.2f", expectedRate, summary.AttendanceRate)
	}
}

func TestGetClassSummary(t *testing.T) {
	service := setupTestService()
	classID, studentID1 := setupClassWithStudents(t, service)
	studentID2 := getFirstID(service, "student2")
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	session1, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	checkInTime := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session1.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn student1 failed: %v", err)
	}

	checkInTime2 := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID2,
		SessionID: session1.ID,
		CheckTime: checkInTime2,
	})
	if err != nil {
		t.Fatalf("CheckIn student2 failed: %v", err)
	}

	_, err = service.CalculateAttendance(session1.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	summary, err := service.GetClassSummary(classID, startDate, endDate)
	if err != nil {
		t.Fatalf("GetClassSummary failed: %v", err)
	}

	if summary.TotalSessions != 1 {
		t.Errorf("expected 1 total sessions, got %d", summary.TotalSessions)
	}
	if summary.PresentCount != 1 {
		t.Errorf("expected 1 present, got %d", summary.PresentCount)
	}
	if summary.LateCount != 1 {
		t.Errorf("expected 1 late, got %d", summary.LateCount)
	}
	if summary.AbsentCount != 1 {
		t.Errorf("expected 1 absent, got %d", summary.AbsentCount)
	}
}

func TestGetStudentSummary_InvalidTimeRange(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)

	startDate := time.Date(2024, 6, 5, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	_, err := service.GetStudentSummary(studentID, classID, startDate, endDate)
	if err != ErrInvalidTimeRange {
		t.Errorf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestCalculateAttendance_SessionNotFound(t *testing.T) {
	service := setupTestService()

	_, err := service.CalculateAttendance("invalid-session")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestGenerateAbsenceNotifications_SessionNotFound(t *testing.T) {
	service := setupTestService()

	_, err := service.GenerateAbsenceNotifications("invalid-session")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestCloseSession(t *testing.T) {
	service := setupTestService()
	classID, _ := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	err := service.CloseSession(session.ID)
	if err != nil {
		t.Fatalf("CloseSession failed: %v", err)
	}

	sessionFromGet, exists := service.GetSession(session.ID)
	if !exists {
		t.Fatal("session should exist")
	}
	if sessionFromGet.Active {
		t.Error("session should be inactive after close")
	}
}

func TestCloseSession_NotFound(t *testing.T) {
	service := setupTestService()

	err := service.CloseSession("invalid-session")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestCheckIn_StudentNotFound(t *testing.T) {
	service := setupTestService()
	classID, _ := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: "invalid-student",
		SessionID: session.ID,
		CheckTime: time.Now(),
	})
	if err != ErrStudentNotFound {
		t.Errorf("expected ErrStudentNotFound, got %v", err)
	}
}

func TestListSessionsByClass(t *testing.T) {
	service := setupTestService()
	classID, _ := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	_, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession 1 failed: %v", err)
	}
	_, err = service.CreateSession(classID, rule.ID, date2)
	if err != nil {
		t.Fatalf("CreateSession 2 failed: %v", err)
	}

	sessions := service.ListSessionsByClass(classID)
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}

	if !sessions[0].Date.Before(sessions[1].Date) {
		t.Error("sessions should be sorted by date")
	}
}

func TestCalculateAttendance_LateAndEarlyLeave(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	checkOutTime := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	var studentRecord *AttendanceRecord
	for _, r := range records {
		if r.StudentID == studentID {
			studentRecord = r
			break
		}
	}

	if studentRecord == nil {
		t.Fatal("student record not found")
	}

	if studentRecord.Status != StatusLateAndEarlyLeave {
		t.Errorf("expected status LATE_AND_EARLY_LEAVE for both late and early leave, got %s", studentRecord.Status)
	}
	if !studentRecord.IsLate {
		t.Error("IsLate should be true for late and early leave")
	}
	if !studentRecord.IsEarlyLeave {
		t.Error("IsEarlyLeave should be true for late and early leave")
	}
}

func TestCalculateAttendance_LateAndEarlyLeave_Summary(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	session1, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	checkInTime := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	checkOutTime := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)

	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session1.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session1.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	_, err = service.CalculateAttendance(session1.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	summary, err := service.GetStudentSummary(studentID, classID, startDate, endDate)
	if err != nil {
		t.Fatalf("GetStudentSummary failed: %v", err)
	}

	if summary.LateCount != 1 {
		t.Errorf("expected 1 late in summary, got %d", summary.LateCount)
	}
	if summary.EarlyLeaveCount != 1 {
		t.Errorf("expected 1 early leave in summary, got %d", summary.EarlyLeaveCount)
	}
	if summary.PresentCount != 0 {
		t.Errorf("expected 0 present in summary, got %d", summary.PresentCount)
	}
	if summary.AbsentCount != 0 {
		t.Errorf("expected 0 absent in summary, got %d", summary.AbsentCount)
	}
}

func TestCalculateAttendance_TimeBoundaryPrecision(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime1 := time.Date(2024, 6, 1, 8, 10, 0, 0, time.UTC)
	checkInTime2 := time.Date(2024, 6, 1, 8, 10, 0, 500, time.UTC)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime1,
	})
	if err != nil {
		t.Fatalf("CheckIn 1 failed: %v", err)
	}

	records1, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 1 failed: %v", err)
	}

	var record1 *AttendanceRecord
	for _, r := range records1 {
		if r.StudentID == studentID {
			record1 = r
			break
		}
	}

	service2 := setupTestService()
	classID2, studentID2 := setupClassWithStudents(t, service2)
	rule2 := createDefaultRule(t, service2)
	session2 := setupSession(t, service2, classID2, rule2.ID)

	_, err = service2.CheckIn(&CheckInRequest{
		StudentID: studentID2,
		SessionID: session2.ID,
		CheckTime: checkInTime2,
	})
	if err != nil {
		t.Fatalf("CheckIn 2 failed: %v", err)
	}

	records2, err := service2.CalculateAttendance(session2.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 2 failed: %v", err)
	}

	var record2 *AttendanceRecord
	for _, r := range records2 {
		if r.StudentID == studentID2 {
			record2 = r
			break
		}
	}

	if record1.Status != record2.Status {
		t.Errorf("same minute should have same status: got %s and %s", record1.Status, record2.Status)
	}
	if record1.IsLate != record2.IsLate {
		t.Errorf("same minute should have same IsLate: got %v and %v", record1.IsLate, record2.IsLate)
	}
}

func TestCalculateAttendance_ExactThresholdTime(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 10, 0, 0, time.UTC)
	checkOutTime := time.Date(2024, 6, 1, 17, 50, 0, 0, time.UTC)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	records, err := service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	var studentRecord *AttendanceRecord
	for _, r := range records {
		if r.StudentID == studentID {
			studentRecord = r
			break
		}
	}

	if studentRecord == nil {
		t.Fatal("student record not found")
	}

	if studentRecord.Status != StatusPresent {
		t.Errorf("expected status PRESENT for exact threshold time, got %s", studentRecord.Status)
	}
	if studentRecord.IsLate {
		t.Error("IsLate should be false for exact threshold time")
	}
	if studentRecord.IsEarlyLeave {
		t.Error("IsEarlyLeave should be false for exact threshold time")
	}
}

func TestGenerateAbsenceNotifications_LateAndEarlyLeave(t *testing.T) {
	service := setupTestService()
	classID, studentID1 := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)
	session := setupSession(t, service, classID, rule.ID)

	checkInTime := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	checkOutTime := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)

	_, err := service.CheckIn(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	_, err = service.CalculateAttendance(session.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	notifications, err := service.GenerateAbsenceNotifications(session.ID)
	if err != nil {
		t.Fatalf("GenerateAbsenceNotifications failed: %v", err)
	}

	lateCount := 0
	earlyCount := 0
	for _, n := range notifications {
		if n.StudentID == studentID1 {
			if n.Status == StatusLate {
				lateCount++
			} else if n.Status == StatusEarlyLeave {
				earlyCount++
			}
		}
	}

	if lateCount != 1 {
		t.Errorf("expected 1 late notification, got %d", lateCount)
	}
	if earlyCount != 1 {
		t.Errorf("expected 1 early leave notification, got %d", earlyCount)
	}
}

func TestTruncateToMinute(t *testing.T) {
	testCases := []struct {
		input    time.Time
		expected time.Time
	}{
		{
			time.Date(2024, 6, 1, 8, 10, 30, 500, time.UTC),
			time.Date(2024, 6, 1, 8, 10, 0, 0, time.UTC),
		},
		{
			time.Date(2024, 6, 1, 8, 10, 59, 999999999, time.UTC),
			time.Date(2024, 6, 1, 8, 10, 0, 0, time.UTC),
		},
		{
			time.Date(2024, 6, 1, 8, 10, 0, 0, time.UTC),
			time.Date(2024, 6, 1, 8, 10, 0, 0, time.UTC),
		},
	}

	for i, tc := range testCases {
		result := truncateToMinute(tc.input)
		if !result.Equal(tc.expected) {
			t.Errorf("test case %d: expected %v, got %v", i, tc.expected, result)
		}
	}
}

func TestGetClassSummary_WithLateAndEarlyLeave(t *testing.T) {
	service := setupTestService()
	classID, studentID1 := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	session1, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	checkInTime := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	checkOutTime := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)

	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session1.ID,
		CheckTime: checkInTime,
	})
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session1.ID,
		CheckTime: checkOutTime,
	})
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	_, err = service.CalculateAttendance(session1.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	summary, err := service.GetClassSummary(classID, startDate, endDate)
	if err != nil {
		t.Fatalf("GetClassSummary failed: %v", err)
	}

	if summary.LateCount < 1 {
		t.Errorf("expected at least 1 late in class summary, got %d", summary.LateCount)
	}
	if summary.EarlyLeaveCount < 1 {
		t.Errorf("expected at least 1 early leave in class summary, got %d", summary.EarlyLeaveCount)
	}
}

func TestAttendanceRate_NotExceed100Percent(t *testing.T) {
	service := setupTestService()
	classID, studentID1 := setupClassWithStudents(t, service)
	studentID2 := getFirstID(service, "student2")
	studentID3 := getFirstID(service, "student3")
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	session1, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	checkInTime1 := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	checkOutTime1 := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session1.ID,
		CheckTime: checkInTime1,
	})
	if err != nil {
		t.Fatalf("CheckIn student1 failed: %v", err)
	}
	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID1,
		SessionID: session1.ID,
		CheckTime: checkOutTime1,
	})
	if err != nil {
		t.Fatalf("CheckOut student1 failed: %v", err)
	}

	checkInTime2 := time.Date(2024, 6, 1, 8, 12, 0, 0, time.UTC)
	checkOutTime2 := time.Date(2024, 6, 1, 17, 48, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID2,
		SessionID: session1.ID,
		CheckTime: checkInTime2,
	})
	if err != nil {
		t.Fatalf("CheckIn student2 failed: %v", err)
	}
	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID2,
		SessionID: session1.ID,
		CheckTime: checkOutTime2,
	})
	if err != nil {
		t.Fatalf("CheckOut student2 failed: %v", err)
	}

	checkInTime3 := time.Date(2024, 6, 1, 8, 5, 0, 0, time.UTC)
	checkOutTime3 := time.Date(2024, 6, 1, 17, 55, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID3,
		SessionID: session1.ID,
		CheckTime: checkInTime3,
	})
	if err != nil {
		t.Fatalf("CheckIn student3 failed: %v", err)
	}
	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID3,
		SessionID: session1.ID,
		CheckTime: checkOutTime3,
	})
	if err != nil {
		t.Fatalf("CheckOut student3 failed: %v", err)
	}

	_, err = service.CalculateAttendance(session1.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	summary, err := service.GetClassSummary(classID, startDate, endDate)
	if err != nil {
		t.Fatalf("GetClassSummary failed: %v", err)
	}

	if summary.AttendanceRate > 100.0 {
		t.Errorf("attendance rate should not exceed 100%%, got %.2f%%", summary.AttendanceRate)
	}

	if summary.LateCount != 2 {
		t.Errorf("expected 2 late counts, got %d", summary.LateCount)
	}
	if summary.EarlyLeaveCount != 2 {
		t.Errorf("expected 2 early leave counts, got %d", summary.EarlyLeaveCount)
	}
	if summary.PresentCount != 1 {
		t.Errorf("expected 1 present count, got %d", summary.PresentCount)
	}
}

func TestAttendanceRate_SingleStudentMultipleSessions(t *testing.T) {
	service := setupTestService()
	classID, studentID := setupClassWithStudents(t, service)
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)
	date3 := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)

	session1, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession 1 failed: %v", err)
	}
	session2, err := service.CreateSession(classID, rule.ID, date2)
	if err != nil {
		t.Fatalf("CreateSession 2 failed: %v", err)
	}
	session3, err := service.CreateSession(classID, rule.ID, date3)
	if err != nil {
		t.Fatalf("CreateSession 3 failed: %v", err)
	}

	checkInTime1 := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
	checkOutTime1 := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session1.ID,
		CheckTime: checkInTime1,
	})
	if err != nil {
		t.Fatalf("CheckIn session1 failed: %v", err)
	}
	_, err = service.CheckOut(&CheckInRequest{
		StudentID: studentID,
		SessionID: session1.ID,
		CheckTime: checkOutTime1,
	})
	if err != nil {
		t.Fatalf("CheckOut session1 failed: %v", err)
	}

	checkInTime2 := time.Date(2024, 6, 2, 8, 5, 0, 0, time.UTC)
	_, err = service.CheckIn(&CheckInRequest{
		StudentID: studentID,
		SessionID: session2.ID,
		CheckTime: checkInTime2,
	})
	if err != nil {
		t.Fatalf("CheckIn session2 failed: %v", err)
	}

	_, err = service.CalculateAttendance(session1.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 1 failed: %v", err)
	}
	_, err = service.CalculateAttendance(session2.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 2 failed: %v", err)
	}
	_, err = service.CalculateAttendance(session3.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance 3 failed: %v", err)
	}

	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC)

	summary, err := service.GetStudentSummary(studentID, classID, startDate, endDate)
	if err != nil {
		t.Fatalf("GetStudentSummary failed: %v", err)
	}

	if summary.AttendanceRate > 100.0 {
		t.Errorf("attendance rate should not exceed 100%%, got %.2f%%", summary.AttendanceRate)
	}

	expectedRate := float64(2) / float64(3) * 100
	if summary.AttendanceRate != expectedRate {
		t.Errorf("expected attendance rate %.2f%%, got %.2f%%", expectedRate, summary.AttendanceRate)
	}
}

func TestAttendanceRate_AllStudentsLateAndEarlyLeave(t *testing.T) {
	service := setupTestService()
	classID, studentID1 := setupClassWithStudents(t, service)
	studentID2 := getFirstID(service, "student2")
	studentID3 := getFirstID(service, "student3")
	rule := createDefaultRule(t, service)

	date1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	session1, err := service.CreateSession(classID, rule.ID, date1)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	students := []string{studentID1, studentID2, studentID3}
	for _, sid := range students {
		checkInTime := time.Date(2024, 6, 1, 8, 15, 0, 0, time.UTC)
		checkOutTime := time.Date(2024, 6, 1, 17, 45, 0, 0, time.UTC)
		_, err = service.CheckIn(&CheckInRequest{
			StudentID: sid,
			SessionID: session1.ID,
			CheckTime: checkInTime,
		})
		if err != nil {
			t.Fatalf("CheckIn for student %s failed: %v", sid, err)
		}
		_, err = service.CheckOut(&CheckInRequest{
			StudentID: sid,
			SessionID: session1.ID,
			CheckTime: checkOutTime,
		})
		if err != nil {
			t.Fatalf("CheckOut for student %s failed: %v", sid, err)
		}
	}

	_, err = service.CalculateAttendance(session1.ID)
	if err != nil {
		t.Fatalf("CalculateAttendance failed: %v", err)
	}

	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC)

	summary, err := service.GetClassSummary(classID, startDate, endDate)
	if err != nil {
		t.Fatalf("GetClassSummary failed: %v", err)
	}

	if summary.AttendanceRate > 100.0 {
		t.Errorf("attendance rate should not exceed 100%%, got %.2f%%", summary.AttendanceRate)
	}

	if summary.AttendanceRate != 100.0 {
		t.Errorf("expected 100%% attendance rate, got %.2f%%", summary.AttendanceRate)
	}

	if summary.LateCount != 3 {
		t.Errorf("expected 3 late counts, got %d", summary.LateCount)
	}
	if summary.EarlyLeaveCount != 3 {
		t.Errorf("expected 3 early leave counts, got %d", summary.EarlyLeaveCount)
	}
	if summary.PresentCount != 0 {
		t.Errorf("expected 0 present count, got %d", summary.PresentCount)
	}
}
