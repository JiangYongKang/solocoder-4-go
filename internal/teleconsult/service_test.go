package teleconsult

import (
	"strings"
	"testing"
	"time"
)

func setupTestService() (*Service, string, string) {
	svc := NewService()
	patientID := svc.AddPatient(&Patient{
		Name:   "张三",
		Gender: "男",
		Age:    35,
		Phone:  "13800138000",
	})
	doctorID := svc.AddDoctor(&Doctor{
		Name:       "李医生",
		Title:      "主任医师",
		Department: "内科",
		LicenseNo:  "LIC12345",
		Available:  true,
	})
	return svc, patientID, doctorID
}

func TestAddAndGetPatient(t *testing.T) {
	svc := NewService()
	patientID := svc.AddPatient(&Patient{
		Name:   "张三",
		Gender: "男",
		Age:    35,
		Phone:  "13800138000",
	})
	if patientID == "" {
		t.Fatal("expected non-empty patient ID")
	}

	p, err := svc.GetPatient(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "张三" {
		t.Errorf("expected name 张三, got %s", p.Name)
	}

	_, err = svc.GetPatient("nonexistent")
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestAddAndGetDoctor(t *testing.T) {
	svc := NewService()
	doctorID := svc.AddDoctor(&Doctor{
		Name:       "李医生",
		Title:      "主任医师",
		Department: "内科",
		LicenseNo:  "LIC12345",
		Available:  true,
	})
	if doctorID == "" {
		t.Fatal("expected non-empty doctor ID")
	}

	d, err := svc.GetDoctor(doctorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Name != "李医生" {
		t.Errorf("expected name 李医生, got %s", d.Name)
	}

	_, err = svc.GetDoctor("nonexistent")
	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestCreateSession_Success(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, err := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛三天，伴有恶心",
		InitialMessage: "医生您好，我最近三天一直头痛，还有点恶心，请问是什么原因？",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.ID == "" {
		t.Fatal("expected non-empty session ID")
	}
	if session.Status != SessionStatusPending {
		t.Errorf("expected status PENDING, got %s", session.Status)
	}
	if session.PatientID != patientID {
		t.Errorf("expected patient ID %s, got %s", patientID, session.PatientID)
	}
	if session.DoctorID != doctorID {
		t.Errorf("expected doctor ID %s, got %s", doctorID, session.DoctorID)
	}
	if session.ChiefComplaint != "头痛三天，伴有恶心" {
		t.Errorf("unexpected chief complaint: %s", session.ChiefComplaint)
	}
	if len(session.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(session.Messages))
	}
	if session.Messages[0].Content != "医生您好，我最近三天一直头痛，还有点恶心，请问是什么原因？" {
		t.Errorf("unexpected initial message content")
	}
	if session.Messages[0].SenderType != MessageSenderPatient {
		t.Errorf("expected sender type PATIENT, got %s", session.Messages[0].SenderType)
	}
}

func TestCreateSession_WithoutInitialMessage(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, err := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "咳嗽",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(session.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(session.Messages))
	}
}

func TestCreateSession_InvalidPatient(t *testing.T) {
	svc, _, doctorID := setupTestService()
	_, err := svc.CreateSession(&CreateSessionRequest{
		PatientID:      "invalid-patient",
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestCreateSession_InvalidDoctor(t *testing.T) {
	svc, patientID, _ := setupTestService()
	_, err := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       "invalid-doctor",
		ChiefComplaint: "头痛",
	})
	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestCreateSession_DoctorNotAvailable(t *testing.T) {
	svc := NewService()
	patientID := svc.AddPatient(&Patient{Name: "张三"})
	doctorID := svc.AddDoctor(&Doctor{
		Name:      "李医生",
		Available: false,
	})

	_, err := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	if err != ErrDoctorNotAvailable {
		t.Errorf("expected ErrDoctorNotAvailable, got %v", err)
	}
}

func TestAcceptSession_Success(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	accepted, err := svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accepted.Status != SessionStatusOngoing {
		t.Errorf("expected status ONGOING, got %s", accepted.Status)
	}
	if accepted.AcceptedAt == nil {
		t.Error("expected AcceptedAt to be set")
	}
	if len(accepted.Messages) != 1 {
		t.Errorf("expected 1 system message, got %d", len(accepted.Messages))
	}
}

func TestAcceptSession_NotFound(t *testing.T) {
	svc, _, doctorID := setupTestService()
	_, err := svc.AcceptSession(&AcceptSessionRequest{
		SessionID: "nonexistent",
		DoctorID:  doctorID,
	})
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestAcceptSession_AlreadyAccepted(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	_, err := svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	if err != ErrAlreadyAccepted {
		t.Errorf("expected ErrAlreadyAccepted, got %v", err)
	}
}

func TestAcceptSession_WrongDoctor(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherDoctorID := svc.AddDoctor(&Doctor{Name: "王医生", Available: true})

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	_, err := svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  otherDoctorID,
	})
	if err != ErrInvalidDoctorAssignment {
		t.Errorf("expected ErrInvalidDoctorAssignment, got %v", err)
	}
}

func TestAcceptSession_DoctorNotFound(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	_, err := svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  "invalid-doctor",
	})
	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestAcceptSession_DoctorNotAvailable(t *testing.T) {
	svc := NewService()
	patientID := svc.AddPatient(&Patient{Name: "张三"})
	doctorID := svc.AddDoctor(&Doctor{
		Name:      "李医生",
		Available: true,
	})

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	d, _ := svc.GetDoctor(doctorID)
	d.Available = false

	_, err := svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	if err != ErrDoctorNotAvailable {
		t.Errorf("expected ErrDoctorNotAvailable, got %v", err)
	}
}

func TestSendMessage_PatientText_Success(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
		InitialMessage: "医生您好，我头痛三天了",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	msg, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "头痛加剧了，还伴有发烧",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content != "头痛加剧了，还伴有发烧" {
		t.Errorf("unexpected content: %s", msg.Content)
	}
	if msg.SenderType != MessageSenderPatient {
		t.Errorf("expected PATIENT sender, got %s", msg.SenderType)
	}
	if msg.ID == "" {
		t.Error("expected non-empty message ID")
	}

	updated, _ := svc.GetSession(session.ID)
	if len(updated.Messages) != 3 {
		t.Errorf("expected 3 messages (init+system+new), got %d", len(updated.Messages))
	}
}

func TestSendMessage_PatientImage_Success(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "皮疹",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	msg, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeImage,
		Content:     "这是皮疹的照片",
		ImageURL:    "https://example.com/rashes.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ImageURL != "https://example.com/rashes.jpg" {
		t.Errorf("expected image URL, got %s", msg.ImageURL)
	}
	if msg.Type != MessageTypeImage {
		t.Errorf("expected IMAGE type, got %s", msg.Type)
	}
}

func TestSendMessage_DoctorMedicalAdvice_Success(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	msg, err := svc.SendMessage(&SendMessageRequest{
		SessionID:       session.ID,
		SenderID:        doctorID,
		SenderType:      MessageSenderDoctor,
		MessageType:     MessageTypeText,
		Content:         "根据您的症状，可能是偏头痛。建议服用布洛芬缓解，多休息，避免强光。如症状持续请复诊。",
		IsMedicalAdvice: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !msg.IsMedicalAdvice {
		t.Error("expected IsMedicalAdvice to be true")
	}
}

func TestSendMessage_DoctorNotAccepted_RejectMedicalAdvice(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:       session.ID,
		SenderID:        doctorID,
		SenderType:      MessageSenderDoctor,
		MessageType:     MessageTypeText,
		Content:         "这是诊疗意见",
		IsMedicalAdvice: true,
	})
	if err != ErrSessionNotAccepted {
		t.Errorf("expected ErrSessionNotAccepted, got %v", err)
	}
}

func TestSendMessage_DoctorNonMedicalBeforeAccepted(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	msg, err := svc.SendMessage(&SendMessageRequest{
		SessionID:       session.ID,
		SenderID:        doctorID,
		SenderType:      MessageSenderDoctor,
		MessageType:     MessageTypeText,
		Content:         "您好，请稍等，我马上接诊",
		IsMedicalAdvice: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.IsMedicalAdvice {
		t.Error("expected IsMedicalAdvice to be false for pre-acceptance message")
	}
}

func TestSendMessage_SessionNotFound(t *testing.T) {
	svc, patientID, _ := setupTestService()
	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   "nonexistent",
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "test",
	})
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSendMessage_EmptyTextContent(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "",
	})
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestSendMessage_MessageTooLong(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	longContent := strings.Repeat("a", MaxMessageLength+1)
	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     longContent,
	})
	if err != ErrMessageTooLong {
		t.Errorf("expected ErrMessageTooLong, got %v", err)
	}
}

func TestSendMessage_EmptyImageURL(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeImage,
		ImageURL:    "",
		Content:     "",
	})
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestSendMessage_WrongPatient(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherPatientID := svc.AddPatient(&Patient{Name: "李四"})

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    otherPatientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "test",
	})
	if err != ErrNotPatient {
		t.Errorf("expected ErrNotPatient, got %v", err)
	}
}

func TestSendMessage_WrongDoctor(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherDoctorID := svc.AddDoctor(&Doctor{Name: "王医生", Available: true})

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    otherDoctorID,
		SenderType:  MessageSenderDoctor,
		MessageType: MessageTypeText,
		Content:     "test",
	})
	if err != ErrNotDoctor {
		t.Errorf("expected ErrNotDoctor, got %v", err)
	}
}

func TestSendMessage_ClosedSession(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	_, _ = svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服用布洛芬",
		FollowUpPlan:   "一周后复诊",
	})

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "再问一个问题",
	})
	if err != ErrSessionClosed {
		t.Errorf("expected ErrSessionClosed, got %v", err)
	}
}

func TestSendMessage_InvalidSenderType(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    "SYSTEM",
		SenderType:  MessageSenderSystem,
		MessageType: MessageTypeText,
		Content:     "系统消息",
	})
	if err != ErrInvalidSenderType {
		t.Errorf("expected ErrInvalidSenderType for SYSTEM sender, got %v", err)
	}

	_, err = svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  "UNKNOWN_TYPE",
		MessageType: MessageTypeText,
		Content:     "测试未知发送者",
	})
	if err != ErrInvalidSenderType {
		t.Errorf("expected ErrInvalidSenderType for UNKNOWN sender type, got %v", err)
	}
}

func TestSendMessage_MessagesOrdered(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "咨询",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	contents := []string{"msg1", "msg2", "msg3", "msg4"}
	for _, c := range contents {
		_, _ = svc.SendMessage(&SendMessageRequest{
			SessionID:   session.ID,
			SenderID:    patientID,
			SenderType:  MessageSenderPatient,
			MessageType: MessageTypeText,
			Content:     c,
		})
	}

	s, _ := svc.GetSession(session.ID)
	nonSystemMsgs := make([]string, 0)
	for _, m := range s.Messages {
		if m.Type != MessageTypeSystem {
			nonSystemMsgs = append(nonSystemMsgs, m.Content)
		}
	}
	if len(nonSystemMsgs) != len(contents) {
		t.Fatalf("expected %d non-system messages, got %d", len(contents), len(nonSystemMsgs))
	}
	for i, c := range contents {
		if nonSystemMsgs[i] != c {
			t.Errorf("message order wrong at %d: expected %s, got %s", i, c, nonSystemMsgs[i])
		}
	}
}

func TestAcceptTimeout_AutoClose(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	svc.SetAcceptTimeout(50 * time.Millisecond)

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	time.Sleep(100 * time.Millisecond)

	checked, err := svc.CheckSessionTimeout(session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checked.Status != SessionStatusClosed {
		t.Errorf("expected status CLOSED, got %s", checked.Status)
	}
	if checked.CloseReason != "超时未接诊自动关闭" {
		t.Errorf("unexpected close reason: %s", checked.CloseReason)
	}
}

func TestInactivityTimeout_AutoClose(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	svc.SetInactivityTimeout(50 * time.Millisecond)

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	time.Sleep(100 * time.Millisecond)

	checked, err := svc.CheckSessionTimeout(session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checked.Status != SessionStatusClosed {
		t.Errorf("expected status CLOSED, got %s", checked.Status)
	}
	if checked.CloseReason != "长期无互动自动关闭" {
		t.Errorf("unexpected close reason: %s", checked.CloseReason)
	}
}

func TestCheckAllSessionsTimeout(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	svc.SetAcceptTimeout(50 * time.Millisecond)
	svc.SetInactivityTimeout(50 * time.Millisecond)

	_, _ = svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛1",
	})

	session2, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛2",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session2.ID,
		DoctorID:  doctorID,
	})

	time.Sleep(100 * time.Millisecond)

	_, _ = svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛3",
	})

	count := svc.CheckAllSessionsTimeout()
	if count != 2 {
		t.Errorf("expected 2 sessions closed, got %d", count)
	}
}

func TestCheckSessionTimeout_NotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.CheckSessionTimeout("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSendMessage_TriggersTimeoutCheck(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	svc.SetInactivityTimeout(50 * time.Millisecond)

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	time.Sleep(100 * time.Millisecond)

	_, err := svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "test",
	})
	if err != ErrSessionClosed {
		t.Errorf("expected ErrSessionClosed due to inactivity, got %v", err)
	}
}

func TestCompleteSession_Success(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛三天，伴有恶心",
		InitialMessage: "医生您好，我头痛三天了",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	_, _ = svc.SendMessage(&SendMessageRequest{
		SessionID:       session.ID,
		SenderID:        doctorID,
		SenderType:      MessageSenderDoctor,
		MessageType:     MessageTypeText,
		Content:         "可能是偏头痛，建议休息",
		IsMedicalAdvice: true,
	})

	archive, err := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛（Migraine）",
		TreatmentAdvice: "1. 布洛芬 400mg 口服，必要时每 6-8 小时一次\n2. 保证充足睡眠\n3. 避免咖啡因和酒精",
		FollowUpPlan:   "如症状持续超过一周或加重，及时复诊；建议记录头痛日记。",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if archive.ID == "" {
		t.Fatal("expected non-empty archive ID")
	}
	if archive.Diagnosis != "偏头痛（Migraine）" {
		t.Errorf("unexpected diagnosis: %s", archive.Diagnosis)
	}
	if archive.TreatmentAdvice == "" {
		t.Error("expected treatment advice to be set")
	}
	if archive.FollowUpPlan == "" {
		t.Error("expected follow-up plan to be set")
	}
	if archive.FollowUpCount != 0 {
		t.Errorf("expected follow-up count 0, got %d", archive.FollowUpCount)
	}
	if archive.FollowUpLimit != DefaultFollowUpLimit {
		t.Errorf("expected follow-up limit %d, got %d", DefaultFollowUpLimit, archive.FollowUpLimit)
	}
	if len(archive.Messages) < 3 {
		t.Errorf("expected at least 3 messages in archive, got %d", len(archive.Messages))
	}
	if archive.ChiefComplaint != "头痛三天，伴有恶心" {
		t.Errorf("unexpected chief complaint in archive")
	}

	updatedSession, _ := svc.GetSession(session.ID)
	if updatedSession.Status != SessionStatusArchived {
		t.Errorf("expected session status ARCHIVED, got %s", updatedSession.Status)
	}
	if updatedSession.ArchiveID != archive.ID {
		t.Errorf("expected archive ID %s in session, got %s", archive.ID, updatedSession.ArchiveID)
	}
}

func TestCompleteSession_SessionNotFound(t *testing.T) {
	svc, _, doctorID := setupTestService()
	_, err := svc.CompleteSession(&CompleteSessionRequest{
		SessionID: "nonexistent",
		DoctorID:  doctorID,
		Diagnosis: "test",
	})
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestCompleteSession_NotDoctor(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherDoctorID := svc.AddDoctor(&Doctor{Name: "王医生", Available: true})

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})

	_, err := svc.CompleteSession(&CompleteSessionRequest{
		SessionID: session.ID,
		DoctorID:  otherDoctorID,
		Diagnosis: "test",
	})
	if err != ErrNotDoctor {
		t.Errorf("expected ErrNotDoctor, got %v", err)
	}
}

func TestCompleteSession_InvalidStatus(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})

	_, err := svc.CompleteSession(&CompleteSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
		Diagnosis: "test",
	})
	if err != ErrInvalidSessionStatus {
		t.Errorf("expected ErrInvalidSessionStatus, got %v", err)
	}
}

func TestSendFollowUp_Success(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	svc.SetFollowUpLimit(3)

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服用布洛芬",
		FollowUpPlan:   "一周后复诊",
	})

	msg, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "医生您好，我吃药后头痛已经缓解了，还需要继续吃吗？",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content != "医生您好，我吃药后头痛已经缓解了，还需要继续吃吗？" {
		t.Errorf("unexpected content: %s", msg.Content)
	}

	updatedArchive, _ := svc.GetArchive(archive.ID)
	if updatedArchive.FollowUpCount != 1 {
		t.Errorf("expected follow-up count 1, got %d", updatedArchive.FollowUpCount)
	}
}

func TestSendFollowUp_LimitExceeded(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	svc.SetFollowUpLimit(2)

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服药",
		FollowUpPlan:   "复诊",
	})

	_, _ = svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "追问1",
	})
	_, _ = svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "追问2",
	})

	_, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "追问3",
	})
	if err != ErrFollowUpLimitExceeded {
		t.Errorf("expected ErrFollowUpLimitExceeded, got %v", err)
	}

	updatedArchive, _ := svc.GetArchive(archive.ID)
	if updatedArchive.FollowUpCount != 2 {
		t.Errorf("expected follow-up count to stay at 2, got %d", updatedArchive.FollowUpCount)
	}
}

func TestSendFollowUp_ArchiveNotFound(t *testing.T) {
	svc, patientID, _ := setupTestService()
	_, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   "nonexistent",
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "test",
	})
	if err != ErrArchiveNotFound {
		t.Errorf("expected ErrArchiveNotFound, got %v", err)
	}
}

func TestSendFollowUp_NotPatient(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服药",
		FollowUpPlan:   "复诊",
	})

	_, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   doctorID,
		MessageType: MessageTypeText,
		Content:     "test",
	})
	if err != ErrNotPatient {
		t.Errorf("expected ErrNotPatient, got %v", err)
	}
}

func TestSendFollowUp_EmptyContent(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服药",
		FollowUpPlan:   "复诊",
	})

	_, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "",
	})
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestSendFollowUp_EmptyImageURL(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服药",
		FollowUpPlan:   "复诊",
	})

	_, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeImage,
	})
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestSendFollowUp_TooLongContent(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服药",
		FollowUpPlan:   "复诊",
	})

	longContent := strings.Repeat("a", MaxMessageLength+1)
	_, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     longContent,
	})
	if err != ErrMessageTooLong {
		t.Errorf("expected ErrMessageTooLong, got %v", err)
	}
}

func TestSendFollowUp_ImageSuccess(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "皮疹",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "过敏性皮疹",
		TreatmentAdvice: "外用药膏",
		FollowUpPlan:   "观察变化",
	})

	msg, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeImage,
		Content:     "这是最新的皮肤状况",
		ImageURL:    "https://example.com/followup.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ImageURL != "https://example.com/followup.jpg" {
		t.Errorf("unexpected image URL: %s", msg.ImageURL)
	}
}

func TestSendFollowUp_DoesNotUpdateOriginalSession(t *testing.T) {
	svc, patientID, doctorID := setupTestService()

	session, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{
		SessionID: session.ID,
		DoctorID:  doctorID,
	})
	archive, _ := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "偏头痛",
		TreatmentAdvice: "服药",
		FollowUpPlan:   "复诊",
	})

	sessionBefore, _ := svc.GetSession(session.ID)
	msgCountBefore := len(sessionBefore.Messages)
	followUpCountBefore := sessionBefore.FollowUpCount

	archiveBefore, _ := svc.GetArchive(archive.ID)
	archiveMsgCountBefore := len(archiveBefore.Messages)

	_, err := svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "追问一下",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessionAfter, _ := svc.GetSession(session.ID)
	if len(sessionAfter.Messages) != msgCountBefore {
		t.Errorf("expected session message count unchanged (%d), got %d — follow-up should not modify original session", msgCountBefore, len(sessionAfter.Messages))
	}
	if sessionAfter.FollowUpCount != followUpCountBefore {
		t.Errorf("expected session follow-up count unchanged (%d), got %d — follow-up should not modify original session", followUpCountBefore, sessionAfter.FollowUpCount)
	}

	archiveAfter, _ := svc.GetArchive(archive.ID)
	if len(archiveAfter.Messages) != archiveMsgCountBefore+1 {
		t.Errorf("expected archive message count %d, got %d", archiveMsgCountBefore+1, len(archiveAfter.Messages))
	}
	if archiveAfter.FollowUpCount != 1 {
		t.Errorf("expected archive follow-up count 1, got %d", archiveAfter.FollowUpCount)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.GetSession("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestGetArchive_NotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.GetArchive("nonexistent")
	if err != ErrArchiveNotFound {
		t.Errorf("expected ErrArchiveNotFound, got %v", err)
	}
}

func TestListSessionsByPatient(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherPatientID := svc.AddPatient(&Patient{Name: "李四"})

	_, _ = svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "咳嗽",
	})
	_, _ = svc.CreateSession(&CreateSessionRequest{
		PatientID:      otherPatientID,
		DoctorID:       doctorID,
		ChiefComplaint: "胃痛",
	})

	sessions, err := svc.ListSessionsByPatient(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions for patient, got %d", len(sessions))
	}

	_, err = svc.ListSessionsByPatient("nonexistent")
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestListSessionsByDoctor(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherDoctorID := svc.AddDoctor(&Doctor{Name: "王医生", Available: true})

	_, _ = svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       otherDoctorID,
		ChiefComplaint: "胃痛",
	})

	sessions, err := svc.ListSessionsByDoctor(doctorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 session for doctor, got %d", len(sessions))
	}

	_, err = svc.ListSessionsByDoctor("nonexistent")
	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestListArchivesByPatient(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherPatientID := svc.AddPatient(&Patient{Name: "李四"})

	s1, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{SessionID: s1.ID, DoctorID: doctorID})
	_, _ = svc.CompleteSession(&CompleteSessionRequest{
		SessionID: s1.ID, DoctorID: doctorID, Diagnosis: "偏头痛", TreatmentAdvice: "吃药", FollowUpPlan: "复诊",
	})

	s2, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "咳嗽",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{SessionID: s2.ID, DoctorID: doctorID})
	_, _ = svc.CompleteSession(&CompleteSessionRequest{
		SessionID: s2.ID, DoctorID: doctorID, Diagnosis: "感冒", TreatmentAdvice: "多喝热水", FollowUpPlan: "观察",
	})

	s3, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      otherPatientID,
		DoctorID:       doctorID,
		ChiefComplaint: "胃痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{SessionID: s3.ID, DoctorID: doctorID})
	_, _ = svc.CompleteSession(&CompleteSessionRequest{
		SessionID: s3.ID, DoctorID: doctorID, Diagnosis: "胃炎", TreatmentAdvice: "胃药", FollowUpPlan: "胃镜",
	})

	archives, err := svc.ListArchivesByPatient(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(archives) != 2 {
		t.Errorf("expected 2 archives for patient, got %d", len(archives))
	}

	_, err = svc.ListArchivesByPatient("nonexistent")
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestListArchivesByDoctor(t *testing.T) {
	svc, patientID, doctorID := setupTestService()
	otherDoctorID := svc.AddDoctor(&Doctor{Name: "王医生", Available: true})

	s1, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "头痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{SessionID: s1.ID, DoctorID: doctorID})
	_, _ = svc.CompleteSession(&CompleteSessionRequest{
		SessionID: s1.ID, DoctorID: doctorID, Diagnosis: "偏头痛", TreatmentAdvice: "吃药", FollowUpPlan: "复诊",
	})

	s2, _ := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       otherDoctorID,
		ChiefComplaint: "胃痛",
	})
	_, _ = svc.AcceptSession(&AcceptSessionRequest{SessionID: s2.ID, DoctorID: otherDoctorID})
	_, _ = svc.CompleteSession(&CompleteSessionRequest{
		SessionID: s2.ID, DoctorID: otherDoctorID, Diagnosis: "胃炎", TreatmentAdvice: "吃药", FollowUpPlan: "复诊",
	})

	archives, err := svc.ListArchivesByDoctor(doctorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(archives) != 1 {
		t.Errorf("expected 1 archive for doctor, got %d", len(archives))
	}

	_, err = svc.ListArchivesByDoctor("nonexistent")
	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestFullNormalFlow(t *testing.T) {
	svc := NewService()

	patientID := svc.AddPatient(&Patient{Name: "王五", Gender: "男", Age: 45, Phone: "13900139000"})
	doctorID := svc.AddDoctor(&Doctor{Name: "赵医生", Title: "副主任医师", Department: "神经内科", LicenseNo: "LIC99999", Available: true})

	session, err := svc.CreateSession(&CreateSessionRequest{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ChiefComplaint: "反复头痛2周，伴失眠",
		InitialMessage: "医生您好！我最近两周反复头痛，主要是前额和太阳穴胀痛，晚上睡不着觉，白天精神差，影响工作了。",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if session.Status != SessionStatusPending {
		t.Fatalf("expected PENDING, got %s", session.Status)
	}

	_, err = svc.AcceptSession(&AcceptSessionRequest{SessionID: session.ID, DoctorID: doctorID})
	if err != nil {
		t.Fatalf("AcceptSession failed: %v", err)
	}

	_, _ = svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "头痛一般下午开始加重，最近工作压力也比较大。",
	})

	_, _ = svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeImage,
		Content:     "这是我前两天做的脑部CT报告",
		ImageURL:    "https://med.example.com/ct/report_001.png",
	})

	_, _ = svc.SendMessage(&SendMessageRequest{
		SessionID:       session.ID,
		SenderID:        doctorID,
		SenderType:      MessageSenderDoctor,
		MessageType:     MessageTypeText,
		Content:         "您好！根据您的描述和检查结果，脑部CT未见异常。考虑为紧张性头痛，与工作压力大、睡眠不足有关。建议：1）调整作息，每天保证7-8小时睡眠；2）适当运动，如快走、瑜伽；3）头痛明显时可服对乙酰氨基酚 500mg，每日不超过4次；4）学习放松技巧，如深呼吸、冥想。",
		IsMedicalAdvice: true,
	})

	_, _ = svc.SendMessage(&SendMessageRequest{
		SessionID:   session.ID,
		SenderID:    patientID,
		SenderType:  MessageSenderPatient,
		MessageType: MessageTypeText,
		Content:     "好的，谢谢医生！我会按您说的做。",
	})

	archive, err := svc.CompleteSession(&CompleteSessionRequest{
		SessionID:      session.ID,
		DoctorID:       doctorID,
		Diagnosis:      "紧张性头痛（Tension-Type Headache）",
		TreatmentAdvice: "1. 对乙酰氨基酚 500mg 口服，必要时每次1片，每日不超过4片\n2. 规律作息，保证充足睡眠\n3. 每日有氧运动 30 分钟\n4. 学习放松训练（深呼吸、渐进性肌肉松弛）",
		FollowUpPlan:   "2周后复诊，评估症状改善情况。若头痛加重、出现呕吐、视力模糊等症状，需立即就诊。",
	})
	if err != nil {
		t.Fatalf("CompleteSession failed: %v", err)
	}

	_, err = svc.SendFollowUp(&SendFollowUpRequest{
		ArchiveID:   archive.ID,
		PatientID:   patientID,
		MessageType: MessageTypeText,
		Content:     "医生您好，按您的方法调整一周了，头痛明显减轻，睡眠也改善了！想请问药还要继续吃吗？",
	})
	if err != nil {
		t.Fatalf("SendFollowUp failed: %v", err)
	}

	archiveAfter, _ := svc.GetArchive(archive.ID)
	if archiveAfter.FollowUpCount != 1 {
		t.Errorf("expected follow-up count 1, got %d", archiveAfter.FollowUpCount)
	}
	if len(archiveAfter.Messages) < 7 {
		t.Errorf("expected at least 7 messages in archive, got %d", len(archiveAfter.Messages))
	}

	sessionAfter, _ := svc.GetSession(session.ID)
	if sessionAfter.Status != SessionStatusArchived {
		t.Errorf("expected ARCHIVED status, got %s", sessionAfter.Status)
	}
}

func TestConcurrentAccess(t *testing.T) {
	svc := NewService()
	patientID := svc.AddPatient(&Patient{Name: "并发测试"})
	doctorID := svc.AddDoctor(&Doctor{Name: "并发医生", Available: true})

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_, _ = svc.CreateSession(&CreateSessionRequest{
				PatientID:      patientID,
				DoctorID:       doctorID,
				ChiefComplaint: "并发测试",
			})
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	sessions, _ := svc.ListSessionsByPatient(patientID)
	if len(sessions) != 10 {
		t.Errorf("expected 10 sessions, got %d", len(sessions))
	}
}
