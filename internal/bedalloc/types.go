package bedalloc

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrWardNotFound         = errors.New("ward not found")
	ErrBedNotFound          = errors.New("bed not found")
	ErrBedOccupied          = errors.New("bed is already occupied")
	ErrBedNotOccupied       = errors.New("bed is not occupied")
	ErrPatientNotFound      = errors.New("patient not found")
	ErrPatientAlreadyAdmitted = errors.New("patient is already admitted")
	ErrNoAvailableBed       = errors.New("no available bed matching criteria")
	ErrTransferSameBed      = errors.New("cannot transfer to the same bed")
	ErrInvalidTimeRange     = errors.New("invalid time range: start must be before end")
	ErrBedTypeMismatch      = errors.New("bed type mismatch with patient requirement")
)

type BedType string

const (
	BedTypeGeneral  BedType = "general"
	BedTypeICU      BedType = "icu"
	BedTypeSurgery  BedType = "surgery"
	BedTypePediatric BedType = "pediatric"
	BedTypeIsolation BedType = "isolation"
)

type BedStatus string

const (
	BedStatusAvailable BedStatus = "available"
	BedStatusOccupied  BedStatus = "occupied"
	BedStatusCleaning  BedStatus = "cleaning"
	BedStatusMaintenance BedStatus = "maintenance"
)

type Bed struct {
	ID        string
	WardID    string
	Type      BedType
	Status    BedStatus
	RoomNumber string
	Floor     int
}

type Ward struct {
	ID         string
	Name       string
	Department string
	Beds       map[string]*Bed
}

type Patient struct {
	ID        string
	Name      string
	Age       int
	Gender    string
	Condition string
}

type AdmissionRecord struct {
	ID          string
	PatientID   string
	WardID      string
	BedID       string
	AdmitTime   time.Time
	DischargeTime *time.Time
	Active      bool
}

type AllocateCriteria struct {
	WardID       string
	BedType      BedType
	PatientID    string
	PatientAge   int
	PatientCondition string
	AdmitTime    time.Time
}

type TransferCriteria struct {
	PatientID     string
	TargetWardID  string
	TargetBedType BedType
	TargetBedID   string
	TransferTime  time.Time
}

type UtilizationReport struct {
	WardID        string
	WardName      string
	StartDate     time.Time
	EndDate       time.Time
	TotalBeds     int
	OccupiedDays  float64
	AvailableDays float64
	UtilizationRate float64
}

type BedAllocator struct {
	mu              sync.RWMutex
	wards           map[string]*Ward
	patients        map[string]*Patient
	admissions      map[string]*AdmissionRecord
	patientAdmitMap map[string]string
	nextID          int64
}

func NewBedAllocator() *BedAllocator {
	return &BedAllocator{
		wards:           make(map[string]*Ward),
		patients:        make(map[string]*Patient),
		admissions:      make(map[string]*AdmissionRecord),
		patientAdmitMap: make(map[string]string),
		nextID:          1,
	}
}

func (ba *BedAllocator) generateID(prefix string) string {
	id := ba.nextID
	ba.nextID++
	return prefix + "-" + time.Now().Format("20060102") + "-" + formatID(id)
}

func formatID(n int64) string {
	return time.Now().Format("150405") + "-" + string(rune('A'+int(n%26))) + string(rune('0'+int(n%10)))
}
