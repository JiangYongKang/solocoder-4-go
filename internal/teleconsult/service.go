package teleconsult

import (
	"time"
)

type CreateSessionRequest struct {
	PatientID      string
	DoctorID       string
	ChiefComplaint string
	InitialMessage string
}

type AcceptSessionRequest struct {
	SessionID string
	DoctorID  string
}

type SendMessageRequest struct {
	SessionID       string
	SenderID        string
	SenderType      MessageSender
	MessageType     MessageType
	Content         string
	ImageURL        string
	IsMedicalAdvice bool
}

type CompleteSessionRequest struct {
	SessionID      string
	DoctorID       string
	Diagnosis      string
	TreatmentAdvice string
	FollowUpPlan   string
}

type SendFollowUpRequest struct {
	ArchiveID   string
	PatientID   string
	MessageType MessageType
	Content     string
	ImageURL    string
}

func (s *Service) SetAcceptTimeout(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.acceptTimeout = d
}

func (s *Service) SetInactivityTimeout(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inactivityTimeout = d
}

func (s *Service) SetFollowUpLimit(limit int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.followUpLimit = limit
}

func (s *Service) AddPatient(p *Patient) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = s.generateID("PAT")
	}
	s.patients[p.ID] = p
	return p.ID
}

func (s *Service) AddDoctor(d *Doctor) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d.ID == "" {
		d.ID = s.generateID("DOC")
	}
	s.doctors[d.ID] = d
	return d.ID
}

func (s *Service) GetPatient(id string) (*Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.patients[id]
	if !exists {
		return nil, ErrPatientNotFound
	}
	return p, nil
}

func (s *Service) GetDoctor(id string) (*Doctor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, exists := s.doctors[id]
	if !exists {
		return nil, ErrDoctorNotFound
	}
	return d, nil
}

func (s *Service) CreateSession(req *CreateSessionRequest) (*ConsultSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	patient, exists := s.patients[req.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	doctor, exists := s.doctors[req.DoctorID]
	if !exists {
		return nil, ErrDoctorNotFound
	}
	if !doctor.Available {
		return nil, ErrDoctorNotAvailable
	}

	now := time.Now()
	session := &ConsultSession{
		ID:             s.generateID("SES"),
		PatientID:      patient.ID,
		PatientName:    patient.Name,
		DoctorID:       doctor.ID,
		DoctorName:     doctor.Name,
		Department:     doctor.Department,
		ChiefComplaint: req.ChiefComplaint,
		Status:         SessionStatusPending,
		Messages:       make([]*Message, 0),
		CreatedAt:      now,
		LastMessageAt:  now,
		FollowUpLimit:  s.followUpLimit,
	}

	if req.InitialMessage != "" {
		msg := &Message{
			ID:         s.generateID("MSG"),
			SessionID:  session.ID,
			SenderID:   patient.ID,
			SenderType: MessageSenderPatient,
			Type:       MessageTypeText,
			Content:    req.InitialMessage,
			CreatedAt:  now,
		}
		session.Messages = append(session.Messages, msg)
	}

	s.sessions[session.ID] = session
	return session, nil
}

func (s *Service) AcceptSession(req *AcceptSessionRequest) (*ConsultSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	doctor, exists := s.doctors[req.DoctorID]
	if !exists {
		return nil, ErrDoctorNotFound
	}

	if session.Status != SessionStatusPending {
		return nil, ErrAlreadyAccepted
	}

	if session.DoctorID != req.DoctorID {
		return nil, ErrInvalidDoctorAssignment
	}

	if !doctor.Available {
		return nil, ErrDoctorNotAvailable
	}

	now := time.Now()
	session.Status = SessionStatusOngoing
	session.AcceptedAt = &now

	sysMsg := &Message{
		ID:             s.generateID("MSG"),
		SessionID:      session.ID,
		SenderID:       "SYSTEM",
		SenderType:     MessageSenderSystem,
		Type:           MessageTypeSystem,
		Content:        "医生已接诊，开始问诊",
		CreatedAt:      now,
		IsMedicalAdvice: false,
	}
	session.Messages = append(session.Messages, sysMsg)
	session.LastMessageAt = now

	return session, nil
}

func (s *Service) checkTimeoutLocked(session *ConsultSession) {
	now := time.Now()

	if session.Status == SessionStatusPending {
		if now.Sub(session.CreatedAt) > s.acceptTimeout {
			s.closeSessionLocked(session, "超时未接诊自动关闭")
			return
		}
	}

	if session.Status == SessionStatusOngoing {
		if now.Sub(session.LastMessageAt) > s.inactivityTimeout {
			s.closeSessionLocked(session, "长期无互动自动关闭")
		}
	}
}

func (s *Service) CheckSessionTimeout(sessionID string) (*ConsultSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}
	s.checkTimeoutLocked(session)
	return session, nil
}

func (s *Service) CheckAllSessionsTimeout() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, session := range s.sessions {
		oldStatus := session.Status
		s.checkTimeoutLocked(session)
		if session.Status != oldStatus {
			count++
		}
	}
	return count
}

func (s *Service) SendMessage(req *SendMessageRequest) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	s.checkTimeoutLocked(session)

	if req.MessageType == MessageTypeText && len(req.Content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	if req.MessageType == MessageTypeText && req.Content == "" {
		return nil, ErrEmptyContent
	}
	if req.MessageType == MessageTypeImage && req.ImageURL == "" && req.Content == "" {
		return nil, ErrEmptyContent
	}

	if session.Status == SessionStatusClosed || session.Status == SessionStatusCompleted || session.Status == SessionStatusArchived {
		return nil, ErrSessionClosed
	}

	switch req.SenderType {
	case MessageSenderPatient:
		if session.PatientID != req.SenderID {
			return nil, ErrNotPatient
		}
	case MessageSenderDoctor:
		if session.DoctorID != req.SenderID {
			return nil, ErrNotDoctor
		}
		if session.Status == SessionStatusPending && req.IsMedicalAdvice {
			return nil, ErrSessionNotAccepted
		}
	default:
		return nil, ErrNotPatient
	}

	now := time.Now()
	msg := &Message{
		ID:              s.generateID("MSG"),
		SessionID:       session.ID,
		SenderID:        req.SenderID,
		SenderType:      req.SenderType,
		Type:            req.MessageType,
		Content:         req.Content,
		ImageURL:        req.ImageURL,
		CreatedAt:       now,
		IsMedicalAdvice: req.IsMedicalAdvice && req.SenderType == MessageSenderDoctor && session.Status == SessionStatusOngoing,
	}
	session.Messages = append(session.Messages, msg)
	session.LastMessageAt = now
	return msg, nil
}

func (s *Service) CompleteSession(req *CompleteSessionRequest) (*ArchiveRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if session.DoctorID != req.DoctorID {
		return nil, ErrNotDoctor
	}

	if session.Status != SessionStatusOngoing {
		return nil, ErrInvalidSessionStatus
	}

	now := time.Now()
	session.Status = SessionStatusCompleted
	session.ClosedAt = &now

	sysMsg := &Message{
		ID:         s.generateID("MSG"),
		SessionID:  session.ID,
		SenderID:   "SYSTEM",
		SenderType: MessageSenderSystem,
		Type:       MessageTypeSystem,
		Content:    "医生已结束问诊，生成问诊记录",
		CreatedAt:  now,
	}
	session.Messages = append(session.Messages, sysMsg)

	messagesCopy := make([]*Message, len(session.Messages))
	copy(messagesCopy, session.Messages)

	archive := &ArchiveRecord{
		ID:              s.generateID("ARC"),
		SessionID:       session.ID,
		PatientID:       session.PatientID,
		PatientName:     session.PatientName,
		DoctorID:        session.DoctorID,
		DoctorName:      session.DoctorName,
		Department:      session.Department,
		ChiefComplaint:  session.ChiefComplaint,
		Diagnosis:       req.Diagnosis,
		TreatmentAdvice: req.TreatmentAdvice,
		FollowUpPlan:    req.FollowUpPlan,
		Messages:        messagesCopy,
		FollowUpCount:   0,
		FollowUpLimit:   session.FollowUpLimit,
		ArchivedAt:      now,
	}
	s.archives[archive.ID] = archive
	session.ArchiveID = archive.ID
	session.Status = SessionStatusArchived

	return archive, nil
}

func (s *Service) SendFollowUp(req *SendFollowUpRequest) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	archive, exists := s.archives[req.ArchiveID]
	if !exists {
		return nil, ErrArchiveNotFound
	}

	if archive.PatientID != req.PatientID {
		return nil, ErrNotPatient
	}

	if archive.FollowUpCount >= archive.FollowUpLimit {
		return nil, ErrFollowUpLimitExceeded
	}

	if req.MessageType == MessageTypeText && req.Content == "" {
		return nil, ErrEmptyContent
	}
	if req.MessageType == MessageTypeImage && req.ImageURL == "" {
		return nil, ErrEmptyContent
	}

	if req.MessageType == MessageTypeText && len(req.Content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	now := time.Now()
	msg := &Message{
		ID:         s.generateID("MSG"),
		SessionID:  archive.SessionID,
		SenderID:   req.PatientID,
		SenderType: MessageSenderPatient,
		Type:       req.MessageType,
		Content:    req.Content,
		ImageURL:   req.ImageURL,
		CreatedAt:  now,
	}

	archive.Messages = append(archive.Messages, msg)
	archive.FollowUpCount++

	session, sessionExists := s.sessions[archive.SessionID]
	if sessionExists {
		session.Messages = append(session.Messages, msg)
		session.FollowUpCount = archive.FollowUpCount
		session.LastMessageAt = now
	}

	return msg, nil
}

func (s *Service) GetSession(id string) (*ConsultSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *Service) GetArchive(id string) (*ArchiveRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	archive, exists := s.archives[id]
	if !exists {
		return nil, ErrArchiveNotFound
	}
	return archive, nil
}

func (s *Service) ListSessionsByPatient(patientID string) ([]*ConsultSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.patients[patientID]; !exists {
		return nil, ErrPatientNotFound
	}

	var result []*ConsultSession
	for _, ses := range s.sessions {
		if ses.PatientID == patientID {
			result = append(result, ses)
		}
	}
	return result, nil
}

func (s *Service) ListSessionsByDoctor(doctorID string) ([]*ConsultSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.doctors[doctorID]; !exists {
		return nil, ErrDoctorNotFound
	}

	var result []*ConsultSession
	for _, ses := range s.sessions {
		if ses.DoctorID == doctorID {
			result = append(result, ses)
		}
	}
	return result, nil
}

func (s *Service) ListArchivesByPatient(patientID string) ([]*ArchiveRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.patients[patientID]; !exists {
		return nil, ErrPatientNotFound
	}

	var result []*ArchiveRecord
	for _, arc := range s.archives {
		if arc.PatientID == patientID {
			result = append(result, arc)
		}
	}
	return result, nil
}

func (s *Service) ListArchivesByDoctor(doctorID string) ([]*ArchiveRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.doctors[doctorID]; !exists {
		return nil, ErrDoctorNotFound
	}

	var result []*ArchiveRecord
	for _, arc := range s.archives {
		if arc.DoctorID == doctorID {
			result = append(result, arc)
		}
	}
	return result, nil
}

func (s *Service) closeSessionLocked(session *ConsultSession, reason string) {
	if session.Status == SessionStatusClosed || session.Status == SessionStatusCompleted || session.Status == SessionStatusArchived {
		return
	}
	now := time.Now()
	session.Status = SessionStatusClosed
	session.ClosedAt = &now
	session.CloseReason = reason

	sysMsg := &Message{
		ID:         s.generateID("MSG"),
		SessionID:  session.ID,
		SenderID:   "SYSTEM",
		SenderType: MessageSenderSystem,
		Type:       MessageTypeSystem,
		Content:    "会话已关闭：" + reason,
		CreatedAt:  now,
	}
	session.Messages = append(session.Messages, sysMsg)
}
