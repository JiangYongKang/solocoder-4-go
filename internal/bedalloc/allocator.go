package bedalloc

import (
	"sort"
	"time"
)

func (ba *BedAllocator) AddWard(id, name, department string) *Ward {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	ward := &Ward{
		ID:         id,
		Name:       name,
		Department: department,
		Beds:       make(map[string]*Bed),
	}
	ba.wards[id] = ward
	return ward
}

func (ba *BedAllocator) AddBed(wardID, bedID string, bedType BedType, roomNumber string, floor int) (*Bed, error) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	ward, exists := ba.wards[wardID]
	if !exists {
		return nil, ErrWardNotFound
	}

	bed := &Bed{
		ID:         bedID,
		WardID:     wardID,
		Type:       bedType,
		Status:     BedStatusAvailable,
		RoomNumber: roomNumber,
		Floor:      floor,
	}
	ward.Beds[bedID] = bed
	return bed, nil
}

func (ba *BedAllocator) AddPatient(id, name string, age int, gender, condition string) *Patient {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	patient := &Patient{
		ID:        id,
		Name:      name,
		Age:       age,
		Gender:    gender,
		Condition: condition,
	}
	ba.patients[id] = patient
	return patient
}

func (ba *BedAllocator) GetWard(wardID string) (*Ward, bool) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	ward, exists := ba.wards[wardID]
	return ward, exists
}

func (ba *BedAllocator) GetBed(bedID string) (*Bed, bool) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	for _, ward := range ba.wards {
		if bed, exists := ward.Beds[bedID]; exists {
			return bed, true
		}
	}
	return nil, false
}

func (ba *BedAllocator) GetPatient(patientID string) (*Patient, bool) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	patient, exists := ba.patients[patientID]
	return patient, exists
}

func (ba *BedAllocator) GetActiveAdmission(patientID string) (*AdmissionRecord, bool) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	admitID, exists := ba.patientAdmitMap[patientID]
	if !exists {
		return nil, false
	}

	admission, exists := ba.admissions[admitID]
	return admission, exists && admission.Active
}

func (ba *BedAllocator) AllocateBed(criteria AllocateCriteria) (*AdmissionRecord, error) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	if _, exists := ba.patients[criteria.PatientID]; !exists {
		return nil, ErrPatientNotFound
	}

	if _, exists := ba.patientAdmitMap[criteria.PatientID]; exists {
		return nil, ErrPatientAlreadyAdmitted
	}

	ward, exists := ba.wards[criteria.WardID]
	if !exists {
		return nil, ErrWardNotFound
	}

	availableBeds := ba.findAvailableBeds(ward, criteria)
	if len(availableBeds) == 0 {
		return nil, ErrNoAvailableBed
	}

	selectedBed := availableBeds[0]
	selectedBed.Status = BedStatusOccupied

	admission := &AdmissionRecord{
		ID:          ba.generateID("ADM"),
		PatientID:   criteria.PatientID,
		WardID:      criteria.WardID,
		BedID:       selectedBed.ID,
		AdmitTime:   criteria.AdmitTime,
		DischargeTime: nil,
		Active:      true,
	}

	ba.admissions[admission.ID] = admission
	ba.patientAdmitMap[criteria.PatientID] = admission.ID

	return admission, nil
}

func (ba *BedAllocator) findAvailableBeds(ward *Ward, criteria AllocateCriteria) []*Bed {
	var beds []*Bed

	for _, bed := range ward.Beds {
		if bed.Status != BedStatusAvailable {
			continue
		}

		if criteria.BedType != "" && bed.Type != criteria.BedType {
			continue
		}

		if !ba.matchPatientCondition(bed, criteria) {
			continue
		}

		beds = append(beds, bed)
	}

	sort.Slice(beds, func(i, j int) bool {
		if beds[i].Floor != beds[j].Floor {
			return beds[i].Floor < beds[j].Floor
		}
		return beds[i].RoomNumber < beds[j].RoomNumber
	})

	return beds
}

func (ba *BedAllocator) matchPatientCondition(bed *Bed, criteria AllocateCriteria) bool {
	if criteria.BedType != "" {
		return bed.Type == criteria.BedType
	}

	if criteria.PatientCondition == "infectious" {
		return bed.Type == BedTypeIsolation
	}

	if criteria.PatientCondition == "icu" {
		return bed.Type == BedTypeICU
	}

	if criteria.PatientAge < 14 && criteria.PatientCondition == "general" {
		return bed.Type == BedTypePediatric
	}

	if criteria.PatientCondition == "surgery" {
		return bed.Type == BedTypeSurgery || bed.Type == BedTypeICU || bed.Type == BedTypeGeneral
	}

	if criteria.PatientCondition == "general" {
		return bed.Type == BedTypeGeneral || bed.Type == BedTypeSurgery
	}

	return true
}

func (ba *BedAllocator) TransferBed(criteria TransferCriteria) (*AdmissionRecord, error) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	admitID, exists := ba.patientAdmitMap[criteria.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	oldAdmission := ba.admissions[admitID]
	if !oldAdmission.Active {
		return nil, ErrPatientNotFound
	}

	targetWard, exists := ba.wards[criteria.TargetWardID]
	if !exists {
		return nil, ErrWardNotFound
	}

	var targetBed *Bed
	if criteria.TargetBedID != "" {
		targetBed, exists = targetWard.Beds[criteria.TargetBedID]
		if !exists {
			return nil, ErrBedNotFound
		}
		if targetBed.ID == oldAdmission.BedID && targetWard.ID == oldAdmission.WardID {
			return nil, ErrTransferSameBed
		}
		if targetBed.Status != BedStatusAvailable {
			return nil, ErrBedOccupied
		}
		if criteria.TargetBedType != "" && targetBed.Type != criteria.TargetBedType {
			return nil, ErrBedTypeMismatch
		}
	} else {
		availableBeds := ba.findAvailableBeds(targetWard, AllocateCriteria{
			BedType: criteria.TargetBedType,
		})
		if len(availableBeds) == 0 {
			return nil, ErrNoAvailableBed
		}
		targetBed = availableBeds[0]
	}

	oldBed := ba.wards[oldAdmission.WardID].Beds[oldAdmission.BedID]
	oldBed.Status = BedStatusAvailable
	targetBed.Status = BedStatusOccupied

	oldAdmission.Active = false
	dischargeTime := criteria.TransferTime
	oldAdmission.DischargeTime = &dischargeTime

	newAdmission := &AdmissionRecord{
		ID:            ba.generateID("ADM"),
		PatientID:     criteria.PatientID,
		WardID:        criteria.TargetWardID,
		BedID:         targetBed.ID,
		AdmitTime:     criteria.TransferTime,
		DischargeTime: nil,
		Active:        true,
	}

	ba.admissions[newAdmission.ID] = newAdmission
	ba.patientAdmitMap[criteria.PatientID] = newAdmission.ID

	return newAdmission, nil
}

func (ba *BedAllocator) DischargePatient(patientID string, dischargeTime time.Time) (*AdmissionRecord, error) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	admitID, exists := ba.patientAdmitMap[patientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	admission := ba.admissions[admitID]
	if !admission.Active {
		return nil, ErrPatientNotFound
	}

	ward, exists := ba.wards[admission.WardID]
	if !exists {
		return nil, ErrWardNotFound
	}

	bed, exists := ward.Beds[admission.BedID]
	if !exists {
		return nil, ErrBedNotFound
	}
	if bed.Status != BedStatusOccupied {
		return nil, ErrBedNotOccupied
	}

	bed.Status = BedStatusCleaning
	admission.Active = false
	admission.DischargeTime = &dischargeTime
	delete(ba.patientAdmitMap, patientID)

	return admission, nil
}

func (ba *BedAllocator) MarkBedCleaned(bedID string) error {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	for _, ward := range ba.wards {
		if bed, exists := ward.Beds[bedID]; exists {
			if bed.Status == BedStatusCleaning {
				bed.Status = BedStatusAvailable
			}
			return nil
		}
	}
	return ErrBedNotFound
}

func (ba *BedAllocator) CalculateUtilization(wardID string, startDate, endDate time.Time) (*UtilizationReport, error) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	if !startDate.Before(endDate) {
		return nil, ErrInvalidTimeRange
	}

	ward, exists := ba.wards[wardID]
	if !exists {
		return nil, ErrWardNotFound
	}

	totalBeds := len(ward.Beds)
	if totalBeds == 0 {
		return &UtilizationReport{
			WardID:          wardID,
			WardName:        ward.Name,
			StartDate:       startDate,
			EndDate:         endDate,
			TotalBeds:       0,
			OccupiedDays:    0,
			AvailableDays:   0,
			UtilizationRate: 0,
		}, nil
	}

	totalDays := endDate.Sub(startDate).Hours() / 24
	var occupiedDays float64

	for _, admission := range ba.admissions {
		if admission.WardID != wardID {
			continue
		}

		admitTime := admission.AdmitTime
		dischargeTime := endDate
		if admission.DischargeTime != nil && admission.DischargeTime.Before(endDate) {
			dischargeTime = *admission.DischargeTime
		}

		if dischargeTime.Before(startDate) {
			continue
		}
		if admitTime.After(endDate) {
			continue
		}

		actualStart := admitTime
		if actualStart.Before(startDate) {
			actualStart = startDate
		}

		actualEnd := dischargeTime
		if actualEnd.After(endDate) {
			actualEnd = endDate
		}

		stayDays := actualEnd.Sub(actualStart).Hours() / 24
		if stayDays > 0 {
			occupiedDays += stayDays
		}
	}

	totalBedDays := float64(totalBeds) * totalDays
	var utilizationRate float64
	if totalBedDays > 0 {
		utilizationRate = occupiedDays / totalBedDays
	}

	availableDays := totalBedDays - occupiedDays
	if availableDays < 0 {
		availableDays = 0
	}

	return &UtilizationReport{
		WardID:          wardID,
		WardName:        ward.Name,
		StartDate:       startDate,
		EndDate:         endDate,
		TotalBeds:       totalBeds,
		OccupiedDays:    occupiedDays,
		AvailableDays:   availableDays,
		UtilizationRate: utilizationRate,
	}, nil
}

func (ba *BedAllocator) ListActiveAdmissions() []*AdmissionRecord {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	var admissions []*AdmissionRecord
	for _, admitID := range ba.patientAdmitMap {
		if admission, exists := ba.admissions[admitID]; exists {
			admissions = append(admissions, admission)
		}
	}
	return admissions
}

func (ba *BedAllocator) GetAdmissionHistory(patientID string) []*AdmissionRecord {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	var history []*AdmissionRecord
	for _, admission := range ba.admissions {
		if admission.PatientID == patientID {
			history = append(history, admission)
		}
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].AdmitTime.Before(history[j].AdmitTime)
	})

	return history
}
