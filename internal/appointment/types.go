package appointment

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type SlotStatus string

const (
	SlotStatusAvailable SlotStatus = "AVAILABLE"
	SlotStatusLocked    SlotStatus = "LOCKED"
	SlotStatusBooked    SlotStatus = "BOOKED"
	SlotStatusCancelled SlotStatus = "CANCELLED"
	SlotStatusNoShow    SlotStatus = "NO_SHOW"
	SlotStatusCompleted SlotStatus = "COMPLETED"
)

type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "PENDING"
	AppointmentStatusConfirmed AppointmentStatus = "CONFIRMED"
	AppointmentStatusCancelled AppointmentStatus = "CANCELLED"
	AppointmentStatusNoShow    AppointmentStatus = "NO_SHOW"
	AppointmentStatusCompleted AppointmentStatus = "COMPLETED"
	AppointmentStatusChanged   AppointmentStatus = "CHANGED"
)

type Doctor struct {
	ID         string
	Name       string
	Title      string
	Department string
	LicenseNo  string
}

type Patient struct {
	ID              string
	Name            string
	IDCard          string
	Phone           string
	Gender          string
	Age             int
	MedicalRecordNo string
}

type ScheduleSlot struct {
	ID             string
	DoctorID       string
	DoctorName     string
	Department     string
	Date           string
	StartTime      string
	EndTime        string
	TotalCapacity  int
	BookedCount    int
	Status         SlotStatus
	LockedBy       string
	LockedAt       *time.Time
	LockExpireAt   *time.Time
}

type Schedule struct {
	ID             string
	DoctorID       string
	DoctorName     string
	Department     string
	Date           string
	Slots          []*ScheduleSlot
}

type Appointment struct {
	ID             string
	PatientID      string
	PatientName    string
	SlotID         string
	DoctorID       string
	DoctorName     string
	Department     string
	Date           string
	StartTime      string
	EndTime        string
	Status         AppointmentStatus
	IsNoShow       bool
	CreatedAt      time.Time
	ConfirmedAt    *time.Time
	CancelledAt    *time.Time
	CompletedAt    *time.Time
	ChangedFromID  string
	ChangedToID    string
}

type WaitQueueItem struct {
	ID           string
	PatientID    string
	PatientName  string
	TargetDate   string
	DoctorID     string
	Department   string
	JoinedAt     time.Time
}

type NoShowRecord struct {
	ID              string
	PatientID       string
	PatientName     string
	AppointmentID   string
	DoctorID        string
	Date            string
	StartTime       string
	RecordedAt      time.Time
	Remark          string
}

var (
	ErrDoctorNotFound         = errors.New("doctor not found")
	ErrPatientNotFound        = errors.New("patient not found")
	ErrScheduleNotFound       = errors.New("schedule not found")
	ErrSlotNotFound           = errors.New("slot not found")
	ErrAppointmentNotFound    = errors.New("appointment not found")
	ErrSlotAlreadyLocked      = errors.New("slot is already locked")
	ErrSlotAlreadyBooked      = errors.New("slot is already fully booked")
	ErrSlotNotLocked          = errors.New("slot is not locked")
	ErrSlotLockExpired        = errors.New("slot lock has expired")
	ErrInvalidLockOwner       = errors.New("lock does not belong to this patient")
	ErrAppointmentNotPending  = errors.New("appointment is not in pending status")
	ErrAppointmentNotActive   = errors.New("appointment is not active")
	ErrCannotChangePast       = errors.New("cannot change appointment for past date")
	ErrCannotCancelPast       = errors.New("cannot cancel appointment for past date")
	ErrSameSlotChange         = errors.New("cannot change to the same slot")
	ErrChangeNotAllowed       = errors.New("appointment status does not allow changes")
	ErrInvalidDate            = errors.New("invalid date format, expected YYYY-MM-DD")
	ErrSlotNoCapacity         = errors.New("slot has no remaining capacity")
	ErrWaitQueueItemNotFound  = errors.New("wait queue item not found")
	ErrPatientAlreadyBooked   = errors.New("patient already has an appointment for this slot")
	ErrLockTimeoutInvalid     = errors.New("lock timeout must be greater than 0")
	ErrNoShowRemarkRequired   = errors.New("no-show remark is required")
)

type Store struct {
	mu             sync.RWMutex
	doctors        map[string]*Doctor
	patients       map[string]*Patient
	schedules      map[string]*Schedule
	slotMap        map[string]*ScheduleSlot
	slotScheduleMap map[string]string
	appointments   map[string]*Appointment
	patientApptMap map[string][]string
	noShowRecords  []*NoShowRecord
	waitQueues     map[string][]*WaitQueueItem
	idCounter      int64

	LockTimeout    time.Duration
	ChangeDeadline time.Duration
	CancelDeadline time.Duration
	NowFunc        func() time.Time
}

func NewStore() *Store {
	return &Store{
		doctors:        make(map[string]*Doctor),
		patients:       make(map[string]*Patient),
		schedules:      make(map[string]*Schedule),
		slotMap:        make(map[string]*ScheduleSlot),
		slotScheduleMap: make(map[string]string),
		appointments:   make(map[string]*Appointment),
		patientApptMap: make(map[string][]string),
		waitQueues:     make(map[string][]*WaitQueueItem),
		LockTimeout:    5 * time.Minute,
		ChangeDeadline: 24 * time.Hour,
		CancelDeadline: 12 * time.Hour,
		NowFunc:        time.Now,
	}
}

func (s *Store) generateID(prefix string) string {
	s.idCounter++
	return fmt.Sprintf("%s%010d", prefix, s.idCounter)
}

func (s *Store) AddDoctor(doctor *Doctor) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if doctor.ID == "" {
		doctor.ID = s.generateID("DOC")
	}
	s.doctors[doctor.ID] = doctor
	return doctor.ID
}

func (s *Store) AddPatient(patient *Patient) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if patient.ID == "" {
		patient.ID = s.generateID("PAT")
	}
	s.patients[patient.ID] = patient
	return patient.ID
}

func (s *Store) GetDoctor(id string) (*Doctor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doctor, exists := s.doctors[id]
	if !exists {
		return nil, ErrDoctorNotFound
	}
	return doctor, nil
}

func (s *Store) GetPatient(id string) (*Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	patient, exists := s.patients[id]
	if !exists {
		return nil, ErrPatientNotFound
	}
	return patient, nil
}

func (s *Store) ListDoctors() []*Doctor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Doctor, 0, len(s.doctors))
	for _, d := range s.doctors {
		result = append(result, d)
	}
	return result
}

func (s *Store) ListPatients() []*Patient {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Patient, 0, len(s.patients))
	for _, p := range s.patients {
		result = append(result, p)
	}
	return result
}

func waitQueueKey(doctorID, date, department string) string {
	if doctorID != "" {
		return "doc:" + doctorID + ":" + date
	}
	return "dept:" + department + ":" + date
}
