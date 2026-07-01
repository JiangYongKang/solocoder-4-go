package prescription

import (
	"time"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

type CreatePrescriptionRequest struct {
	DoctorID       string
	PatientID      string
	PharmacyID     string
	Items          []MedicineItemRequest
	Diagnosis      string
	Remark         string
}

type MedicineItemRequest struct {
	MedicineID   string
	Quantity     int
	Dosage       string
	Frequency    string
	Duration     string
	Instructions string
}

type ReviewRequest struct {
	PrescriptionID string
	PharmacyID     string
	Approved       bool
	RejectReason   string
	ReviewedBy     string
}

type DispenseRequest struct {
	PrescriptionID string
	PharmacyID     string
	DispensedBy    string
}

type WithdrawRequest struct {
	PrescriptionID string
	Reason         string
}

func (svc *Service) CreatePrescription(req *CreatePrescriptionRequest) (*Prescription, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	if len(req.Items) == 0 {
		return nil, ErrNoItems
	}

	doctor, exists := svc.store.doctors[req.DoctorID]
	if !exists {
		return nil, ErrDoctorNotFound
	}

	patient, exists := svc.store.patients[req.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	pharmacy, exists := svc.store.pharmacies[req.PharmacyID]
	if !exists {
		return nil, ErrPharmacyNotFound
	}

	prescriptionItems := make([]MedicineItem, 0, len(req.Items))
	reservedStock := make(map[string]int)
	var totalAmount float64 = 0

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		medicine, exists := svc.store.medicines[item.MedicineID]
		if !exists {
			return nil, ErrMedicineNotFound
		}

		inventoryItem, invExists := svc.store.inventory[item.MedicineID]
		if !invExists || inventoryItem.Quantity-inventoryItem.Reserved < item.Quantity {
			return nil, ErrInsufficientStock
		}

		prescriptionItems = append(prescriptionItems, MedicineItem{
			MedicineID:   medicine.ID,
			MedicineName: medicine.Name,
			Quantity:     item.Quantity,
			UnitPrice:    medicine.UnitPrice,
			Dosage:       item.Dosage,
			Frequency:    item.Frequency,
			Duration:     item.Duration,
			Instructions: item.Instructions,
		})

		reservedStock[medicine.ID] = item.Quantity
		totalAmount += medicine.UnitPrice * float64(item.Quantity)
	}

	for medicineID, quantity := range reservedStock {
		if err := svc.store.reserveStock(medicineID, quantity); err != nil {
			return nil, err
		}
	}

	prescription := &Prescription{
		ID:            svc.store.generateID("PRE"),
		DoctorID:      doctor.ID,
		DoctorName:    doctor.Name,
		PatientID:     patient.ID,
		PatientName:   patient.Name,
		PharmacyID:    pharmacy.ID,
		PharmacyName:  pharmacy.Name,
		Items:         prescriptionItems,
		TotalAmount:   totalAmount,
		Status:        StatusPendingReview,
		Diagnosis:     req.Diagnosis,
		Remark:        req.Remark,
		ReservedStock: reservedStock,
		CreatedAt:     time.Now(),
	}

	svc.store.addPrescription(prescription)
	return prescription, nil
}

func (svc *Service) ReviewPrescription(req *ReviewRequest) (*Prescription, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	prescription, exists := svc.store.prescriptions[req.PrescriptionID]
	if !exists {
		return nil, ErrPrescriptionNotFound
	}

	if prescription.Status != StatusPendingReview {
		return nil, ErrInvalidStatus
	}

	if prescription.PharmacyID != req.PharmacyID {
		return nil, ErrPharmacyNotFound
	}

	if _, exists := svc.store.pharmacies[req.PharmacyID]; !exists {
		return nil, ErrPharmacyNotFound
	}

	now := time.Now()
	if req.Approved {
		prescription.Status = StatusApproved
	} else {
		if req.RejectReason == "" {
			req.RejectReason = "未提供驳回原因"
		}
		prescription.Status = StatusRejected
		prescription.RejectReason = req.RejectReason
		for medicineID, quantity := range prescription.ReservedStock {
			svc.store.releaseStock(medicineID, quantity)
		}
	}

	prescription.ReviewedAt = &now
	prescription.ReviewedBy = req.ReviewedBy
	return prescription, nil
}

func (svc *Service) DispensePrescription(req *DispenseRequest) (*Prescription, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	prescription, exists := svc.store.prescriptions[req.PrescriptionID]
	if !exists {
		return nil, ErrPrescriptionNotFound
	}

	if prescription.Status == StatusDispensed || prescription.Status == StatusCompleted {
		return nil, ErrAlreadyDispensed
	}

	if prescription.Status != StatusApproved {
		return nil, ErrNotApproved
	}

	if prescription.PharmacyID != req.PharmacyID {
		return nil, ErrPharmacyNotFound
	}

	for medicineID, quantity := range prescription.ReservedStock {
		if err := svc.store.consumeStock(medicineID, quantity); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	prescription.Status = StatusCompleted
	prescription.DispensedAt = &now
	prescription.DispensedBy = req.DispensedBy
	return prescription, nil
}

func (svc *Service) WithdrawPrescription(req *WithdrawRequest) (*Prescription, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	prescription, exists := svc.store.prescriptions[req.PrescriptionID]
	if !exists {
		return nil, ErrPrescriptionNotFound
	}

	if prescription.Status != StatusPendingReview && prescription.Status != StatusApproved {
		return nil, ErrCannotWithdraw
	}

	if req.Reason == "" {
		req.Reason = "未提供撤回原因"
	}

	for medicineID, quantity := range prescription.ReservedStock {
		svc.store.releaseStock(medicineID, quantity)
	}

	now := time.Now()
	prescription.Status = StatusWithdrawn
	prescription.WithdrawnAt = &now
	prescription.WithdrawnReason = req.Reason
	return prescription, nil
}

func (svc *Service) GetPrescription(id string) (*Prescription, error) {
	return svc.store.GetPrescription(id)
}

func (svc *Service) ListPrescriptionsByPatient(patientID string) ([]*Prescription, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	if _, exists := svc.store.patients[patientID]; !exists {
		return nil, ErrPatientNotFound
	}

	var result []*Prescription
	for _, p := range svc.store.prescriptions {
		if p.PatientID == patientID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (svc *Service) ListPrescriptionsByPharmacy(pharmacyID string) ([]*Prescription, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	if _, exists := svc.store.pharmacies[pharmacyID]; !exists {
		return nil, ErrPharmacyNotFound
	}

	var result []*Prescription
	for _, p := range svc.store.prescriptions {
		if p.PharmacyID == pharmacyID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (svc *Service) ListPrescriptionsByDoctor(doctorID string) ([]*Prescription, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	if _, exists := svc.store.doctors[doctorID]; !exists {
		return nil, ErrDoctorNotFound
	}

	var result []*Prescription
	for _, p := range svc.store.prescriptions {
		if p.DoctorID == doctorID {
			result = append(result, p)
		}
	}
	return result, nil
}
