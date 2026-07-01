package prescription

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type PrescriptionStatus string

const (
	StatusCreated       PrescriptionStatus = "CREATED"
	StatusPendingReview PrescriptionStatus = "PENDING_REVIEW"
	StatusApproved      PrescriptionStatus = "APPROVED"
	StatusRejected      PrescriptionStatus = "REJECTED"
	StatusDispensed     PrescriptionStatus = "DISPENSED"
	StatusCompleted     PrescriptionStatus = "COMPLETED"
	StatusWithdrawn     PrescriptionStatus = "WITHDRAWN"
)

type Doctor struct {
	ID          string
	Name        string
	Title       string
	Department  string
	LicenseNo   string
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

type Pharmacy struct {
	ID        string
	Name      string
	Address   string
	Phone     string
	LicenseNo string
}

type Medicine struct {
	ID            string
	Name          string
	GenericName   string
	Specification string
	Manufacturer  string
	UnitPrice     float64
}

type InventoryItem struct {
	MedicineID  string
	Medicine    *Medicine
	Quantity    int
	Reserved    int
	LastUpdated time.Time
}

type MedicineItem struct {
	MedicineID   string
	MedicineName string
	Quantity     int
	UnitPrice    float64
	Dosage       string
	Frequency    string
	Duration     string
	Instructions string
}

type Prescription struct {
	ID              string
	DoctorID        string
	DoctorName      string
	PatientID       string
	PatientName     string
	PharmacyID      string
	PharmacyName    string
	Items           []MedicineItem
	TotalAmount     float64
	Status          PrescriptionStatus
	RejectReason    string
	Diagnosis       string
	Remark          string
	ReservedStock   map[string]int
	CreatedAt       time.Time
	ReviewedAt      *time.Time
	ReviewedBy      string
	DispensedAt     *time.Time
	DispensedBy     string
	WithdrawnAt     *time.Time
	WithdrawnReason string
}

var (
	ErrDoctorNotFound        = errors.New("doctor not found")
	ErrPatientNotFound       = errors.New("patient not found")
	ErrPharmacyNotFound      = errors.New("pharmacy not found")
	ErrMedicineNotFound      = errors.New("medicine not found")
	ErrPrescriptionNotFound  = errors.New("prescription not found")
	ErrInsufficientStock     = errors.New("insufficient inventory stock")
	ErrInvalidStatus         = errors.New("invalid prescription status")
	ErrAlreadyDispensed      = errors.New("prescription already dispensed")
	ErrNoItems               = errors.New("prescription must contain at least one medicine item")
	ErrInvalidQuantity       = errors.New("invalid medicine quantity, must be greater than 0")
	ErrNotApproved           = errors.New("prescription is not approved")
	ErrCannotWithdraw        = errors.New("prescription cannot be withdrawn in current status")
	ErrRejectReasonRequired  = errors.New("reject reason is required when rejecting a prescription")
	ErrWithdrawReasonRequired = errors.New("withdraw reason is required when withdrawing a prescription")
)

type Store struct {
	mu            sync.RWMutex
	doctors       map[string]*Doctor
	patients      map[string]*Patient
	pharmacies    map[string]*Pharmacy
	medicines     map[string]*Medicine
	inventory     map[string]*InventoryItem
	prescriptions map[string]*Prescription
	idCounter     int64
}

func NewStore() *Store {
	return &Store{
		doctors:       make(map[string]*Doctor),
		patients:      make(map[string]*Patient),
		pharmacies:    make(map[string]*Pharmacy),
		medicines:     make(map[string]*Medicine),
		inventory:     make(map[string]*InventoryItem),
		prescriptions: make(map[string]*Prescription),
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

func (s *Store) AddPharmacy(pharmacy *Pharmacy) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pharmacy.ID == "" {
		pharmacy.ID = s.generateID("PHY")
	}
	s.pharmacies[pharmacy.ID] = pharmacy
	return pharmacy.ID
}

func (s *Store) AddMedicine(medicine *Medicine) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if medicine.ID == "" {
		medicine.ID = s.generateID("MED")
	}
	s.medicines[medicine.ID] = medicine
	return medicine.ID
}

func (s *Store) UpdateInventory(medicineID string, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.medicines[medicineID]; !exists {
		return ErrMedicineNotFound
	}
	item, exists := s.inventory[medicineID]
	if !exists {
		item = &InventoryItem{
			MedicineID:  medicineID,
			Medicine:    s.medicines[medicineID],
			Quantity:    0,
			Reserved:    0,
			LastUpdated: time.Now(),
		}
		s.inventory[medicineID] = item
	}
	item.Quantity += quantity
	if item.Quantity < 0 {
		item.Quantity -= quantity
		return ErrInsufficientStock
	}
	item.LastUpdated = time.Now()
	return nil
}

func (s *Store) GetInventory(medicineID string) (*InventoryItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, exists := s.medicines[medicineID]; !exists {
		return nil, ErrMedicineNotFound
	}
	item, exists := s.inventory[medicineID]
	if !exists {
		return &InventoryItem{
			MedicineID: medicineID,
			Medicine:   s.medicines[medicineID],
			Quantity:   0,
			Reserved:   0,
		}, nil
	}
	return item, nil
}

func (s *Store) GetAvailableStock(medicineID string) (int, error) {
	item, err := s.GetInventory(medicineID)
	if err != nil {
		return 0, err
	}
	return item.Quantity - item.Reserved, nil
}

func (s *Store) reserveStock(medicineID string, quantity int) error {
	item, exists := s.inventory[medicineID]
	if !exists {
		return ErrInsufficientStock
	}
	if item.Quantity-item.Reserved < quantity {
		return ErrInsufficientStock
	}
	item.Reserved += quantity
	item.LastUpdated = time.Now()
	return nil
}

func (s *Store) releaseStock(medicineID string, quantity int) {
	item, exists := s.inventory[medicineID]
	if !exists {
		return
	}
	if item.Reserved >= quantity {
		item.Reserved -= quantity
	} else {
		item.Reserved = 0
	}
	item.LastUpdated = time.Now()
}

func (s *Store) consumeStock(medicineID string, quantity int) error {
	item, exists := s.inventory[medicineID]
	if !exists {
		return ErrInsufficientStock
	}
	if item.Reserved < quantity {
		return ErrInsufficientStock
	}
	if item.Quantity < quantity {
		return ErrInsufficientStock
	}
	item.Reserved -= quantity
	item.Quantity -= quantity
	item.LastUpdated = time.Now()
	return nil
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

func (s *Store) GetPharmacy(id string) (*Pharmacy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pharmacy, exists := s.pharmacies[id]
	if !exists {
		return nil, ErrPharmacyNotFound
	}
	return pharmacy, nil
}

func (s *Store) GetMedicine(id string) (*Medicine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	medicine, exists := s.medicines[id]
	if !exists {
		return nil, ErrMedicineNotFound
	}
	return medicine, nil
}

func (s *Store) GetPrescription(id string) (*Prescription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prescription, exists := s.prescriptions[id]
	if !exists {
		return nil, ErrPrescriptionNotFound
	}
	return prescription, nil
}

func (s *Store) addPrescription(prescription *Prescription) {
	s.prescriptions[prescription.ID] = prescription
}
