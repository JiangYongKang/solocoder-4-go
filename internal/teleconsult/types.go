package teleconsult

import (
	"errors"
	"sync"
	"time"
)

type SessionStatus string

const (
	SessionStatusPending   SessionStatus = "PENDING"
	SessionStatusOngoing   SessionStatus = "ONGOING"
	SessionStatusCompleted SessionStatus = "COMPLETED"
	SessionStatusClosed    SessionStatus = "CLOSED"
	SessionStatusArchived  SessionStatus = "ARCHIVED"
)

type MessageType string

const (
	MessageTypeText   MessageType = "TEXT"
	MessageTypeImage  MessageType = "IMAGE"
	MessageTypeSystem MessageType = "SYSTEM"
)

type MessageSender string

const (
	MessageSenderPatient MessageSender = "PATIENT"
	MessageSenderDoctor  MessageSender = "DOCTOR"
	MessageSenderSystem  MessageSender = "SYSTEM"
)

type Patient struct {
	ID     string
	Name   string
	Gender string
	Age    int
	Phone  string
}

type Doctor struct {
	ID         string
	Name       string
	Title      string
	Department string
	LicenseNo  string
	Available  bool
}

type Message struct {
	ID        string
	SessionID string
	SenderID  string
	SenderType MessageSender
	Type      MessageType
	Content   string
	ImageURL  string
	CreatedAt time.Time
	IsMedicalAdvice bool
}

type ArchiveRecord struct {
	ID             string
	SessionID      string
	PatientID      string
	PatientName    string
	DoctorID       string
	DoctorName     string
	Department     string
	ChiefComplaint string
	Diagnosis      string
	TreatmentAdvice string
	FollowUpPlan   string
	Messages       []*Message
	FollowUpCount  int
	FollowUpLimit  int
	ArchivedAt     time.Time
}

type ConsultSession struct {
	ID             string
	PatientID      string
	PatientName    string
	DoctorID       string
	DoctorName     string
	Department     string
	ChiefComplaint string
	Status         SessionStatus
	Messages       []*Message
	CreatedAt      time.Time
	AcceptedAt     *time.Time
	LastMessageAt  time.Time
	ClosedAt       *time.Time
	CloseReason    string
	FollowUpCount  int
	FollowUpLimit  int
	ArchiveID      string
}

var (
	ErrPatientNotFound         = errors.New("patient not found")
	ErrDoctorNotFound          = errors.New("doctor not found")
	ErrDoctorNotAvailable      = errors.New("doctor is not available")
	ErrSessionNotFound         = errors.New("session not found")
	ErrArchiveNotFound         = errors.New("archive not found")
	ErrInvalidSessionStatus    = errors.New("invalid session status")
	ErrSessionNotAccepted      = errors.New("session has not been accepted by a doctor")
	ErrSessionClosed           = errors.New("session is closed, cannot send messages")
	ErrEmptyContent            = errors.New("message content cannot be empty")
	ErrFollowUpLimitExceeded   = errors.New("follow-up limit exceeded")
	ErrNotPatient              = errors.New("sender must be the patient of the session")
	ErrNotDoctor               = errors.New("sender must be the assigned doctor of the session")
	ErrAlreadyAccepted         = errors.New("session has already been accepted")
	ErrInvalidDoctorAssignment = errors.New("doctor does not match the assigned doctor")
	ErrSelfFollowUpNotAllowed  = errors.New("follow-up messages must be sent by patient")
	ErrMessageTooLong          = errors.New("message content exceeds maximum length")
)

const (
	DefaultAcceptTimeout    = 30 * time.Minute
	DefaultInactivityTimeout = 60 * time.Minute
	DefaultFollowUpLimit    = 3
	MaxMessageLength        = 5000
)

type Service struct {
	mu               sync.RWMutex
	patients         map[string]*Patient
	doctors          map[string]*Doctor
	sessions         map[string]*ConsultSession
	archives         map[string]*ArchiveRecord
	idCounter        int64
	acceptTimeout    time.Duration
	inactivityTimeout time.Duration
	followUpLimit    int
}

func NewService() *Service {
	return &Service{
		patients:         make(map[string]*Patient),
		doctors:          make(map[string]*Doctor),
		sessions:         make(map[string]*ConsultSession),
		archives:         make(map[string]*ArchiveRecord),
		acceptTimeout:    DefaultAcceptTimeout,
		inactivityTimeout: DefaultInactivityTimeout,
		followUpLimit:    DefaultFollowUpLimit,
	}
}

func (s *Service) generateID(prefix string) string {
	s.idCounter++
	return prefix + "-" + time.Now().Format("20060102150405") + "-" + itoa(s.idCounter)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
