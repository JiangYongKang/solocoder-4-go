package checkup

import (
	"errors"
	"time"
)

var (
	ErrItemNotFound            = errors.New("checkup item not found")
	ErrPackageNotFound         = errors.New("checkup package not found")
	ErrTimeSlotNotFound        = errors.New("time slot not found")
	ErrAppointmentNotFound     = errors.New("appointment not found")
	ErrPatientNotFound         = errors.New("patient not found")
	ErrDuplicateItemInPackage  = errors.New("duplicate checkup item in package")
	ErrInvalidItemInPackage    = errors.New("invalid checkup item in package")
	ErrEmptyPackageItems       = errors.New("package must contain at least one item")
	ErrTimeSlotCapacityFull    = errors.New("time slot capacity is full")
	ErrInvalidTimeRange        = errors.New("invalid time range: start must be before end")
	ErrInvalidCapacity         = errors.New("capacity must be greater than 0")
	ErrResultNotFound          = errors.New("check result not found")
	ErrItemNotInPackage        = errors.New("checkup item is not included in the package")
	ErrDuplicateResult         = errors.New("result for this item already exists in this appointment")
	ErrResultsIncomplete       = errors.New("not all required results have been recorded")
	ErrReportNotFound          = errors.New("report not found")
	ErrAppointmentCancelled    = errors.New("appointment has been cancelled")
	ErrReportAlreadyGenerated  = errors.New("report has already been generated")
)

type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "PENDING"
	AppointmentStatusChecking  AppointmentStatus = "CHECKING"
	AppointmentStatusCompleted AppointmentStatus = "COMPLETED"
	AppointmentStatusCancelled AppointmentStatus = "CANCELLED"
)

type ItemCategory string

const (
	ItemCategoryLaboratory ItemCategory = "LABORATORY"
	ItemCategoryImaging    ItemCategory = "IMAGING"
	ItemCategoryPhysical   ItemCategory = "PHYSICAL"
	ItemCategoryFunctional ItemCategory = "FUNCTIONAL"
)

type Patient struct {
	ID     string
	Name   string
	Gender string
	Age    int
	Phone  string
}

type CheckItem struct {
	ID          string
	Name        string
	Description string
	Category    ItemCategory
	Unit        string
	MinValue    float64
	MaxValue    float64
	Price       float64
}

type CheckPackage struct {
	ID          string
	Name        string
	Description string
	ItemIDs     []string
	Items       map[string]*CheckItem
	TotalPrice  float64
	CreatedAt   time.Time
}

type TimeSlot struct {
	ID           string
	Date         time.Time
	StartTime    time.Time
	EndTime      time.Time
	Capacity     int
	CurrentCount int
}

type Appointment struct {
	ID           string
	PatientID    string
	PatientName  string
	PackageID    string
	PackageName  string
	TimeSlotID   string
	TimeSlotInfo string
	Status       AppointmentStatus
	CreatedAt    time.Time
}

type CheckResult struct {
	ID           string
	AppointmentID string
	ItemID       string
	ItemName     string
	Value        string
	NumericValue float64
	IsNumeric    bool
	IsAbnormal   bool
	Unit         string
	Reference    string
	Remarks      string
	RecordedAt   time.Time
}

type AbnormalItem struct {
	ItemID   string
	ItemName string
	Value    string
	Unit     string
	Reference string
	Remarks  string
}

type Report struct {
	ID              string
	AppointmentID   string
	PatientID       string
	PatientName     string
	PackageID       string
	PackageName     string
	Results         map[string]*CheckResult
	AbnormalItems   []*AbnormalItem
	GeneratedAt     time.Time
	Summary         string
}
