package bedalloc

import (
	"sync"
	"testing"
	"time"
)

func setupTestAllocator() *BedAllocator {
	ba := NewBedAllocator()

	ba.AddWard("W001", "内科病区", "内科")
	ba.AddWard("W002", "外科病区", "外科")
	ba.AddWard("W003", "ICU病区", "重症医学科")
	ba.AddWard("W004", "儿科病区", "儿科")

	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B003", BedTypeSurgery, "102", 1)
	ba.AddBed("W001", "B004", BedTypeIsolation, "103", 2)
	ba.AddBed("W001", "B005", BedTypeGeneral, "104", 2)

	ba.AddBed("W002", "B101", BedTypeSurgery, "201", 3)
	ba.AddBed("W002", "B102", BedTypeGeneral, "202", 3)

	ba.AddBed("W003", "B201", BedTypeICU, "ICU-01", 4)
	ba.AddBed("W003", "B202", BedTypeICU, "ICU-02", 4)

	ba.AddBed("W004", "B301", BedTypePediatric, "301", 5)
	ba.AddBed("W004", "B302", BedTypePediatric, "302", 5)

	ba.AddPatient("P001", "张三", 35, "男", "general")
	ba.AddPatient("P002", "李四", 45, "女", "surgery")
	ba.AddPatient("P003", "王五", 8, "男", "general")
	ba.AddPatient("P004", "赵六", 60, "男", "icu")
	ba.AddPatient("P005", "钱七", 28, "女", "infectious")
	ba.AddPatient("P006", "孙八", 50, "男", "general")

	return ba
}

func TestAddWard(t *testing.T) {
	ba := NewBedAllocator()

	ward := ba.AddWard("W001", "内科病区", "内科")

	if ward.ID != "W001" {
		t.Errorf("Expected ward ID W001, got %s", ward.ID)
	}
	if ward.Name != "内科病区" {
		t.Errorf("Expected ward name 内科病区, got %s", ward.Name)
	}
	if ward.Department != "内科" {
		t.Errorf("Expected department 内科, got %s", ward.Department)
	}

	found, exists := ba.GetWard("W001")
	if !exists {
		t.Error("Expected ward to exist")
	}
	if found != ward {
		t.Error("Expected same ward instance")
	}
}

func TestAddBed(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")

	bed, err := ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if bed.ID != "B001" {
		t.Errorf("Expected bed ID B001, got %s", bed.ID)
	}
	if bed.Type != BedTypeGeneral {
		t.Errorf("Expected bed type general, got %s", bed.Type)
	}
	if bed.Status != BedStatusAvailable {
		t.Errorf("Expected bed status available, got %s", bed.Status)
	}
	if bed.RoomNumber != "101" {
		t.Errorf("Expected room number 101, got %s", bed.RoomNumber)
	}
	if bed.Floor != 1 {
		t.Errorf("Expected floor 1, got %d", bed.Floor)
	}

	found, exists := ba.GetBed("B001")
	if !exists {
		t.Error("Expected bed to exist")
	}
	if found != bed {
		t.Error("Expected same bed instance")
	}

	_, err = ba.AddBed("INVALID", "B002", BedTypeGeneral, "102", 1)
	if err != ErrWardNotFound {
		t.Errorf("Expected ErrWardNotFound, got %v", err)
	}
}

func TestAddPatient(t *testing.T) {
	ba := NewBedAllocator()

	patient := ba.AddPatient("P001", "张三", 35, "男", "general")

	if patient.ID != "P001" {
		t.Errorf("Expected patient ID P001, got %s", patient.ID)
	}
	if patient.Name != "张三" {
		t.Errorf("Expected patient name 张三, got %s", patient.Name)
	}
	if patient.Age != 35 {
		t.Errorf("Expected age 35, got %d", patient.Age)
	}

	found, exists := ba.GetPatient("P001")
	if !exists {
		t.Error("Expected patient to exist")
	}
	if found != patient {
		t.Error("Expected same patient instance")
	}
}

func TestAllocateBed_Basic(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	criteria := AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	}

	admission, err := ba.AllocateBed(criteria)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if admission.PatientID != "P001" {
		t.Errorf("Expected patient ID P001, got %s", admission.PatientID)
	}
	if admission.WardID != "W001" {
		t.Errorf("Expected ward ID W001, got %s", admission.WardID)
	}
	if !admission.Active {
		t.Error("Expected admission to be active")
	}
	if admission.DischargeTime != nil {
		t.Error("Expected discharge time to be nil")
	}
	if admission.AdmitTime != now {
		t.Errorf("Expected admit time %v, got %v", now, admission.AdmitTime)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Status != BedStatusOccupied {
		t.Errorf("Expected bed status occupied, got %s", bed.Status)
	}

	active, exists := ba.GetActiveAdmission("P001")
	if !exists {
		t.Error("Expected active admission to exist")
	}
	if active != admission {
		t.Error("Expected same admission instance")
	}
}

func TestAllocateBed_ByFloorAndRoomOrder(t *testing.T) {
	ba := setupTestAllocator()

	criteria := AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    time.Now(),
	}

	admission, err := ba.AllocateBed(criteria)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if admission.BedID != "B001" && admission.BedID != "B002" {
		t.Errorf("Expected bed B001 or B002 (floor 1), got %s", admission.BedID)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Floor != 1 {
		t.Errorf("Expected floor 1, got %d", bed.Floor)
	}
}

func TestAllocateBed_PediatricAgeMatch(t *testing.T) {
	ba := setupTestAllocator()

	criteria := AllocateCriteria{
		WardID:       "W001",
		BedType:      "",
		PatientID:    "P003",
		PatientAge:   8,
		PatientCondition: "general",
		AdmitTime:    time.Now(),
	}

	_, err := ba.AllocateBed(criteria)
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed for child in non-pediatric ward, got %v", err)
	}

	criteria.WardID = "W004"
	admission, err := ba.AllocateBed(criteria)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Type != BedTypePediatric {
		t.Errorf("Expected pediatric bed type, got %s", bed.Type)
	}
}

func TestAllocateBed_ICUMatch(t *testing.T) {
	ba := setupTestAllocator()

	criteria := AllocateCriteria{
		WardID:       "W003",
		BedType:      "",
		PatientID:    "P004",
		PatientAge:   60,
		PatientCondition: "icu",
		AdmitTime:    time.Now(),
	}

	admission, err := ba.AllocateBed(criteria)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Type != BedTypeICU {
		t.Errorf("Expected ICU bed type, got %s", bed.Type)
	}
}

func TestAllocateBed_InfectiousMatch(t *testing.T) {
	ba := setupTestAllocator()

	criteria := AllocateCriteria{
		WardID:       "W001",
		BedType:      "",
		PatientID:    "P005",
		PatientAge:   28,
		PatientCondition: "infectious",
		AdmitTime:    time.Now(),
	}

	admission, err := ba.AllocateBed(criteria)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Type != BedTypeIsolation {
		t.Errorf("Expected isolation bed type, got %s", bed.Type)
	}
}

func TestAllocateBed_ConflictDetection(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	criteria := AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	}

	_, err := ba.AllocateBed(criteria)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = ba.AllocateBed(criteria)
	if err != ErrPatientAlreadyAdmitted {
		t.Errorf("Expected ErrPatientAlreadyAdmitted, got %v", err)
	}
}

func TestAllocateBed_Errors(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		PatientID:    "INVALID",
		AdmitTime:    now,
	})
	if err != ErrPatientNotFound {
		t.Errorf("Expected ErrPatientNotFound, got %v", err)
	}

	_, err = ba.AllocateBed(AllocateCriteria{
		WardID:       "INVALID",
		PatientID:    "P001",
		AdmitTime:    now,
	})
	if err != ErrWardNotFound {
		t.Errorf("Expected ErrWardNotFound, got %v", err)
	}

	ba2 := NewBedAllocator()
	ba2.AddWard("W001", "Test", "Test")
	ba2.AddPatient("P001", "Test", 30, "男", "general")

	_, err = ba2.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeICU,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed, got %v", err)
	}
}

func TestAllocateBed_AllBedsOccupied(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddPatient("P001", "Test1", 30, "男", "general")
	ba.AddPatient("P002", "Test2", 30, "男", "general")

	criteria := AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    time.Now(),
	}

	_, err := ba.AllocateBed(criteria)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	criteria.PatientID = "P002"
	_, err = ba.AllocateBed(criteria)
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed when all beds occupied, got %v", err)
	}
}

func TestTransferBed_ToDifferentWard(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	admission1, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	oldBed, _ := ba.GetBed(admission1.BedID)
	oldBedID := oldBed.ID

	transferTime := now.Add(24 * time.Hour)
	admission2, err := ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W002",
		TargetBedType: BedTypeGeneral,
		TransferTime:  transferTime,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if admission2.WardID != "W002" {
		t.Errorf("Expected target ward W002, got %s", admission2.WardID)
	}
	if admission2.BedID == oldBedID {
		t.Error("Expected different bed after transfer")
	}
	if admission2.AdmitTime != transferTime {
		t.Errorf("Expected admit time %v, got %v", transferTime, admission2.AdmitTime)
	}
	if !admission2.Active {
		t.Error("Expected new admission to be active")
	}

	if admission1.Active {
		t.Error("Expected old admission to be inactive")
	}
	if admission1.DischargeTime == nil {
		t.Error("Expected old admission to have discharge time")
	}
	if *admission1.DischargeTime != transferTime {
		t.Errorf("Expected discharge time %v, got %v", transferTime, *admission1.DischargeTime)
	}

	oldBed, _ = ba.GetBed(oldBedID)
	if oldBed.Status != BedStatusAvailable {
		t.Errorf("Expected old bed status available, got %s", oldBed.Status)
	}

	newBed, _ := ba.GetBed(admission2.BedID)
	if newBed.Status != BedStatusOccupied {
		t.Errorf("Expected new bed status occupied, got %s", newBed.Status)
	}

	active, _ := ba.GetActiveAdmission("P001")
	if active != admission2 {
		t.Error("Expected active admission to be the new one")
	}
}

func TestTransferBed_SpecificBed(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	transferTime := now.Add(24 * time.Hour)
	admission, err := ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W001",
		TargetBedType: BedTypeGeneral,
		TargetBedID:   "B005",
		TransferTime:  transferTime,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if admission.BedID != "B005" {
		t.Errorf("Expected bed B005, got %s", admission.BedID)
	}

	bed, _ := ba.GetBed("B005")
	if bed.Status != BedStatusOccupied {
		t.Errorf("Expected bed B005 status occupied, got %s", bed.Status)
	}
}

func TestTransferBed_Errors(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	_, err := ba.TransferBed(TransferCriteria{
		PatientID:    "P001",
		TargetWardID: "W001",
		TransferTime: now,
	})
	if err != ErrPatientNotFound {
		t.Errorf("Expected ErrPatientNotFound for non-admitted patient, got %v", err)
	}

	_, err = ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "INVALID",
		TransferTime:  now.Add(time.Hour),
	})
	if err != ErrWardNotFound {
		t.Errorf("Expected ErrWardNotFound, got %v", err)
	}

	_, err = ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W001",
		TargetBedID:   "INVALID",
		TransferTime:  now.Add(time.Hour),
	})
	if err != ErrBedNotFound {
		t.Errorf("Expected ErrBedNotFound, got %v", err)
	}

	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P002",
		PatientAge:   45,
		PatientCondition: "general",
		AdmitTime:    now,
	})

	_, err = ba.TransferBed(TransferCriteria{
		PatientID:     "P002",
		TargetWardID:  "W003",
		TargetBedID:   "B201",
		TargetBedType: BedTypeGeneral,
		TransferTime:  now.Add(time.Hour),
	})
	if err != ErrBedTypeMismatch {
		t.Errorf("Expected ErrBedTypeMismatch, got %v", err)
	}
}

func TestTransferBed_SameBedError(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W001",
		TargetBedID:   admission.BedID,
		TransferTime:  now.Add(time.Hour),
	})
	if err != ErrTransferSameBed {
		t.Errorf("Expected ErrTransferSameBed, got %v", err)
	}
}

func TestTransferBed_OccupiedBed(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	admission2, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P002",
		PatientAge:   45,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W001",
		TargetBedID:   admission2.BedID,
		TransferTime:  now.Add(time.Hour),
	})
	if err != ErrBedOccupied {
		t.Errorf("Expected ErrBedOccupied, got %v", err)
	}
}

func TestTransferBed_AtomicOperation(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ba3 := NewBedAllocator()
	ba3.AddWard("W001", "Test", "Test")
	ba3.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba3.AddPatient("P001", "Test", 30, "男", "general")
	ba3.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    now,
	})

	_, err = ba3.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W001",
		TargetBedType: BedTypeICU,
		TransferTime:  now.Add(time.Hour),
	})
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed, got %v", err)
	}

	bed, _ := ba3.GetBed("B001")
	if bed.Status != BedStatusOccupied {
		t.Errorf("Expected bed B001 to remain occupied after failed transfer, got %s", bed.Status)
	}

	active, exists := ba3.GetActiveAdmission("P001")
	if !exists {
		t.Error("Expected active admission to still exist")
	}
	if active.BedID != "B001" {
		t.Errorf("Expected active admission to still be on B001, got %s", active.BedID)
	}
}

func TestDischargePatient(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	dischargeTime := now.Add(72 * time.Hour)
	discharged, err := ba.DischargePatient("P001", dischargeTime)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if discharged != admission {
		t.Error("Expected same admission instance")
	}
	if discharged.Active {
		t.Error("Expected admission to be inactive")
	}
	if discharged.DischargeTime == nil {
		t.Error("Expected discharge time to be set")
	}
	if *discharged.DischargeTime != dischargeTime {
		t.Errorf("Expected discharge time %v, got %v", dischargeTime, *discharged.DischargeTime)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Status != BedStatusCleaning {
		t.Errorf("Expected bed status cleaning, got %s", bed.Status)
	}

	_, exists := ba.GetActiveAdmission("P001")
	if exists {
		t.Error("Expected no active admission for discharged patient")
	}
}

func TestDischargePatient_Errors(t *testing.T) {
	ba := setupTestAllocator()

	_, err := ba.DischargePatient("INVALID", time.Now())
	if err != ErrPatientNotFound {
		t.Errorf("Expected ErrPatientNotFound, got %v", err)
	}

	_, err = ba.DischargePatient("P001", time.Now())
	if err != ErrPatientNotFound {
		t.Errorf("Expected ErrPatientNotFound for non-admitted patient, got %v", err)
	}
}

func TestMarkBedCleaned(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	admission, _ := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})

	ba.DischargePatient("P001", now.Add(72*time.Hour))

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Status != BedStatusCleaning {
		t.Errorf("Expected bed status cleaning, got %s", bed.Status)
	}

	err := ba.MarkBedCleaned(admission.BedID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	bed, _ = ba.GetBed(admission.BedID)
	if bed.Status != BedStatusAvailable {
		t.Errorf("Expected bed status available, got %s", bed.Status)
	}

	err = ba.MarkBedCleaned("INVALID")
	if err != ErrBedNotFound {
		t.Errorf("Expected ErrBedNotFound, got %v", err)
	}
}

func TestCalculateUtilization_EmptyWard(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)

	report, err := ba.CalculateUtilization("W001", startDate, endDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.TotalBeds != 0 {
		t.Errorf("Expected 0 total beds, got %d", report.TotalBeds)
	}
	if report.OccupiedDays != 0 {
		t.Errorf("Expected 0 occupied days, got %f", report.OccupiedDays)
	}
	if report.UtilizationRate != 0 {
		t.Errorf("Expected 0 utilization rate, got %f", report.UtilizationRate)
	}
}

func TestCalculateUtilization_SingleAdmission(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddPatient("P001", "Test", 30, "男", "general")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)

	admitTime := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	dischargeTime := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    admitTime,
	})
	ba.DischargePatient("P001", dischargeTime)

	report, err := ba.CalculateUtilization("W001", startDate, endDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.TotalBeds != 1 {
		t.Errorf("Expected 1 total bed, got %d", report.TotalBeds)
	}
	if report.OccupiedDays != 5 {
		t.Errorf("Expected 5 occupied days, got %f", report.OccupiedDays)
	}
	if report.AvailableDays != 5 {
		t.Errorf("Expected 5 available days, got %f", report.AvailableDays)
	}
	expectedRate := 5.0 / 10.0
	if report.UtilizationRate != expectedRate {
		t.Errorf("Expected utilization rate %f, got %f", expectedRate, report.UtilizationRate)
	}
}

func TestCalculateUtilization_PartialOverlap(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddPatient("P001", "Test", 30, "男", "general")

	startDate := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	admitTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dischargeTime := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    admitTime,
	})
	ba.DischargePatient("P001", dischargeTime)

	report, err := ba.CalculateUtilization("W001", startDate, endDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.OccupiedDays != 10 {
		t.Errorf("Expected 10 occupied days, got %f", report.OccupiedDays)
	}
	if report.UtilizationRate != 1.0 {
		t.Errorf("Expected utilization rate 1.0, got %f", report.UtilizationRate)
	}
}

func TestCalculateUtilization_OutsideRange(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddPatient("P001", "Test", 30, "男", "general")

	startDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	admitTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dischargeTime := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    admitTime,
	})
	ba.DischargePatient("P001", dischargeTime)

	report, err := ba.CalculateUtilization("W001", startDate, endDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.OccupiedDays != 0 {
		t.Errorf("Expected 0 occupied days, got %f", report.OccupiedDays)
	}
	if report.UtilizationRate != 0 {
		t.Errorf("Expected utilization rate 0, got %f", report.UtilizationRate)
	}
}

func TestCalculateUtilization_MultipleBeds(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeGeneral, "102", 1)
	ba.AddPatient("P001", "Test1", 30, "男", "general")
	ba.AddPatient("P002", "Test2", 30, "男", "general")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)

	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	ba.DischargePatient("P001", time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC))

	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P002",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	ba.DischargePatient("P002", time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC))

	report, err := ba.CalculateUtilization("W001", startDate, endDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedOccupied := 5.0 + 10.0
	if report.OccupiedDays != expectedOccupied {
		t.Errorf("Expected %f occupied days, got %f", expectedOccupied, report.OccupiedDays)
	}
	totalBedDays := 2.0 * 10.0
	expectedRate := expectedOccupied / totalBedDays
	if report.UtilizationRate != expectedRate {
		t.Errorf("Expected utilization rate %f, got %f", expectedRate, report.UtilizationRate)
	}
}

func TestCalculateUtilization_Errors(t *testing.T) {
	ba := setupTestAllocator()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := ba.CalculateUtilization("INVALID", startDate, startDate.AddDate(0, 0, 10))
	if err != ErrWardNotFound {
		t.Errorf("Expected ErrWardNotFound, got %v", err)
	}

	_, err = ba.CalculateUtilization("W001", startDate, startDate)
	if err != ErrInvalidTimeRange {
		t.Errorf("Expected ErrInvalidTimeRange for same start/end, got %v", err)
	}

	_, err = ba.CalculateUtilization("W001", startDate.AddDate(0, 0, 10), startDate)
	if err != ErrInvalidTimeRange {
		t.Errorf("Expected ErrInvalidTimeRange for end before start, got %v", err)
	}
}

func TestCalculateUtilization_StillAdmitted(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddPatient("P001", "Test", 30, "男", "general")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)

	admitTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   30,
		PatientCondition: "general",
		AdmitTime:    admitTime,
	})

	report, err := ba.CalculateUtilization("W001", startDate, endDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.OccupiedDays != 10 {
		t.Errorf("Expected 10 occupied days, got %f", report.OccupiedDays)
	}
	if report.UtilizationRate != 1.0 {
		t.Errorf("Expected utilization rate 1.0, got %f", report.UtilizationRate)
	}
}

func TestListActiveAdmissions(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	if len(ba.ListActiveAdmissions()) != 0 {
		t.Error("Expected 0 active admissions initially")
	}

	ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	ba.AllocateBed(AllocateCriteria{
		WardID:       "W002",
		BedType:      BedTypeGeneral,
		PatientID:    "P002",
		PatientAge:   45,
		PatientCondition: "general",
		AdmitTime:    now,
	})

	if len(ba.ListActiveAdmissions()) != 2 {
		t.Errorf("Expected 2 active admissions, got %d", len(ba.ListActiveAdmissions()))
	}

	ba.DischargePatient("P001", now.Add(24*time.Hour))

	if len(ba.ListActiveAdmissions()) != 1 {
		t.Errorf("Expected 1 active admission after discharge, got %d", len(ba.ListActiveAdmissions()))
	}
}

func TestGetAdmissionHistory(t *testing.T) {
	ba := setupTestAllocator()
	now := time.Now()

	if len(ba.GetAdmissionHistory("P001")) != 0 {
		t.Error("Expected empty history initially")
	}

	admission1, _ := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   35,
		PatientCondition: "general",
		AdmitTime:    now,
	})

	transferTime := now.Add(24 * time.Hour)
	admission2, _ := ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W002",
		TargetBedType: BedTypeGeneral,
		TransferTime:  transferTime,
	})

	ba.DischargePatient("P001", transferTime.Add(24*time.Hour))

	history := ba.GetAdmissionHistory("P001")
	if len(history) != 2 {
		t.Fatalf("Expected 2 admission records, got %d", len(history))
	}

	if history[0] != admission1 {
		t.Error("Expected first record to be admission1")
	}
	if history[1] != admission2 {
		t.Error("Expected second record to be admission2")
	}
	if !history[0].AdmitTime.Before(history[1].AdmitTime) {
		t.Error("Expected history to be sorted by admit time")
	}
}

func TestConcurrentAccess(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "Test", "Test")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeGeneral, "102", 1)
	ba.AddBed("W001", "B003", BedTypeGeneral, "103", 1)
	ba.AddPatient("P001", "Test1", 30, "男", "general")
	ba.AddPatient("P002", "Test2", 30, "男", "general")
	ba.AddPatient("P003", "Test3", 30, "男", "general")

	var wg sync.WaitGroup
	errors := make([]error, 0)
	var mu sync.Mutex

	allocate := func(patientID string) {
		defer wg.Done()
		_, err := ba.AllocateBed(AllocateCriteria{
			WardID:       "W001",
			BedType:      BedTypeGeneral,
			PatientID:    patientID,
			PatientAge:   30,
			PatientCondition: "general",
			AdmitTime:    time.Now(),
		})
		if err != nil {
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}
	}

	wg.Add(3)
	go allocate("P001")
	go allocate("P002")
	go allocate("P003")
	wg.Wait()

	if len(errors) != 0 {
		t.Errorf("Expected no errors during concurrent allocation, got %d errors: %v", len(errors), errors)
	}

	active := ba.ListActiveAdmissions()
	if len(active) != 3 {
		t.Errorf("Expected 3 active admissions, got %d", len(active))
	}

	ward, _ := ba.GetWard("W001")
	occupiedCount := 0
	for _, bed := range ward.Beds {
		if bed.Status == BedStatusOccupied {
			occupiedCount++
		}
	}
	if occupiedCount != 3 {
		t.Errorf("Expected 3 occupied beds, got %d", occupiedCount)
	}
}

func TestFullWorkflow(t *testing.T) {
	ba := NewBedAllocator()

	ba.AddWard("W001", "内科病区", "内科")
	ba.AddWard("W002", "外科病区", "外科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeICU, "102", 1)
	ba.AddBed("W002", "B101", BedTypeSurgery, "201", 2)
	ba.AddPatient("P001", "张三", 45, "男", "general")

	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:       "W001",
		BedType:      BedTypeGeneral,
		PatientID:    "P001",
		PatientAge:   45,
		PatientCondition: "general",
		AdmitTime:    now,
	})
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}
	if admission.BedID != "B001" {
		t.Errorf("Expected bed B001, got %s", admission.BedID)
	}

	transferTime := now.Add(24 * time.Hour)
	transferred, err := ba.TransferBed(TransferCriteria{
		PatientID:     "P001",
		TargetWardID:  "W002",
		TargetBedType: BedTypeSurgery,
		TransferTime:  transferTime,
	})
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	if transferred.BedID != "B101" {
		t.Errorf("Expected bed B101, got %s", transferred.BedID)
	}

	oldBed, _ := ba.GetBed("B001")
	if oldBed.Status != BedStatusAvailable {
		t.Errorf("Expected old bed B001 to be available, got %s", oldBed.Status)
	}

	dischargeTime := transferTime.Add(72 * time.Hour)
	discharged, err := ba.DischargePatient("P001", dischargeTime)
	if err != nil {
		t.Fatalf("Discharge failed: %v", err)
	}
	if *discharged.DischargeTime != dischargeTime {
		t.Errorf("Expected discharge time %v, got %v", dischargeTime, *discharged.DischargeTime)
	}

	report, err := ba.CalculateUtilization("W001", now, dischargeTime)
	if err != nil {
		t.Fatalf("Utilization calculation failed: %v", err)
	}
	if report.UtilizationRate <= 0 {
		t.Errorf("Expected positive utilization rate for W001, got %f", report.UtilizationRate)
	}

	report2, err := ba.CalculateUtilization("W002", now, dischargeTime)
	if err != nil {
		t.Fatalf("Utilization calculation failed: %v", err)
	}
	if report2.UtilizationRate <= 0 {
		t.Errorf("Expected positive utilization rate for W002, got %f", report2.UtilizationRate)
	}

	history := ba.GetAdmissionHistory("P001")
	if len(history) != 2 {
		t.Errorf("Expected 2 admission records in history, got %d", len(history))
	}
}

func TestMultiCondition_ChildWithICU(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "ICU病区", "重症医学科")
	ba.AddWard("W002", "儿科病区", "儿科")
	ba.AddBed("W001", "B001", BedTypeICU, "ICU-01", 4)
	ba.AddBed("W002", "B002", BedTypePediatric, "301", 5)
	ba.AddPatient("P001", "小明", 8, "男", "icu")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       8,
		PatientCondition: "icu",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected child with ICU condition to get ICU bed, got error: %v", err)
	}
	if admission.BedID != "B001" {
		t.Errorf("Expected ICU bed B001 for child with ICU condition, got %s", admission.BedID)
	}

	bed, _ := ba.GetBed("B001")
	if bed.Type != BedTypeICU {
		t.Errorf("Expected bed type ICU, got %s", bed.Type)
	}
}

func TestMultiCondition_SurgeryWithICU(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "ICU病区", "重症医学科")
	ba.AddWard("W002", "外科病区", "外科")
	ba.AddBed("W001", "B001", BedTypeICU, "ICU-01", 4)
	ba.AddBed("W002", "B002", BedTypeSurgery, "201", 3)
	ba.AddPatient("P001", "张三", 55, "男", "surgery")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       55,
		PatientCondition: "surgery",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected surgery patient to get ICU bed for post-op care, got error: %v", err)
	}
	if admission.BedID != "B001" {
		t.Errorf("Expected ICU bed B001 for surgery patient, got %s", admission.BedID)
	}

	bed, _ := ba.GetBed("B001")
	if bed.Type != BedTypeICU {
		t.Errorf("Expected bed type ICU, got %s", bed.Type)
	}
}

func TestMultiCondition_SurgeryCanChooseBedType(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "外科病区", "外科")
	ba.AddBed("W001", "B001", BedTypeSurgery, "201", 3)
	ba.AddBed("W001", "B002", BedTypeGeneral, "202", 3)
	ba.AddBed("W001", "B003", BedTypeICU, "ICU-01", 4)
	ba.AddPatient("P001", "张三", 55, "男", "surgery")

	admission1, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          BedTypeSurgery,
		PatientID:        "P001",
		PatientAge:       55,
		PatientCondition: "surgery",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected surgery patient to get surgery bed, got error: %v", err)
	}
	if admission1.BedID != "B001" {
		t.Errorf("Expected surgery bed B001, got %s", admission1.BedID)
	}

	ba.AddPatient("P002", "李四", 60, "女", "surgery")
	admission2, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          BedTypeICU,
		PatientID:        "P002",
		PatientAge:       60,
		PatientCondition: "surgery",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected surgery patient to get ICU bed when specified, got error: %v", err)
	}
	if admission2.BedID != "B003" {
		t.Errorf("Expected ICU bed B003, got %s", admission2.BedID)
	}
}

func TestMultiCondition_InfectiousNoIsolationBed(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeSurgery, "102", 1)
	ba.AddBed("W001", "B003", BedTypeICU, "ICU-01", 2)
	ba.AddPatient("P001", "钱七", 28, "女", "infectious")

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       28,
		PatientCondition: "infectious",
		AdmitTime:        time.Now(),
	})
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed when no isolation bed for infectious patient, got %v", err)
	}

	ba.AddBed("W001", "B004", BedTypeIsolation, "103", 2)
	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       28,
		PatientCondition: "infectious",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected infectious patient to get isolation bed when available, got error: %v", err)
	}
	if admission.BedID != "B004" {
		t.Errorf("Expected isolation bed B004 for infectious patient, got %s", admission.BedID)
	}
}

func TestMultiCondition_ChildGeneralNoPediatricBed(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeSurgery, "102", 1)
	ba.AddPatient("P001", "小明", 8, "男", "general")

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       8,
		PatientCondition: "general",
		AdmitTime:        time.Now(),
	})
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed for child general patient without pediatric beds, got %v", err)
	}

	ba.AddBed("W001", "B003", BedTypePediatric, "103", 1)
	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       8,
		PatientCondition: "general",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected child general patient to get pediatric bed, got error: %v", err)
	}
	if admission.BedID != "B003" {
		t.Errorf("Expected pediatric bed B003, got %s", admission.BedID)
	}
}

func TestMultiCondition_SpecifiedBedTypeOverridesCondition(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeIsolation, "102", 1)
	ba.AddPatient("P001", "张三", 35, "男", "general")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          BedTypeIsolation,
		PatientID:        "P001",
		PatientAge:       35,
		PatientCondition: "general",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected to get isolation bed when explicitly specified, got error: %v", err)
	}
	if admission.BedID != "B002" {
		t.Errorf("Expected isolation bed B002 when explicitly specified, got %s", admission.BedID)
	}

	bed, _ := ba.GetBed("B002")
	if bed.Type != BedTypeIsolation {
		t.Errorf("Expected bed type isolation, got %s", bed.Type)
	}
}

func TestMultiCondition_ChildSurgeryCanUseSurgeryBed(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "外科病区", "外科")
	ba.AddBed("W001", "B001", BedTypeSurgery, "201", 3)
	ba.AddBed("W001", "B002", BedTypePediatric, "202", 3)
	ba.AddPatient("P001", "小明", 10, "男", "surgery")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       10,
		PatientCondition: "surgery",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected child with surgery condition to get surgery bed, got error: %v", err)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Type != BedTypeSurgery && bed.Type != BedTypeICU && bed.Type != BedTypeGeneral {
		t.Errorf("Expected surgery, ICU, or general bed for child surgery patient, got %s", bed.Type)
	}
}

func TestMultiCondition_ICUHardConstraint(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeSurgery, "102", 1)
	ba.AddBed("W001", "B003", BedTypeIsolation, "103", 1)
	ba.AddBed("W001", "B004", BedTypePediatric, "104", 1)
	ba.AddPatient("P001", "张三", 55, "男", "icu")

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       55,
		PatientCondition: "icu",
		AdmitTime:        time.Now(),
	})
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed when no ICU bed for ICU patient, got %v", err)
	}

	ba.AddBed("W001", "B005", BedTypeICU, "ICU-01", 2)
	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       55,
		PatientCondition: "icu",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected ICU patient to get ICU bed, got error: %v", err)
	}
	if admission.BedID != "B005" {
		t.Errorf("Expected ICU bed B005, got %s", admission.BedID)
	}
}

func TestMultiCondition_GeneralPatientCanUseSurgeryBed(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "外科病区", "外科")
	ba.AddBed("W001", "B001", BedTypeSurgery, "201", 3)
	ba.AddPatient("P001", "张三", 45, "男", "general")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       45,
		PatientCondition: "general",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected general patient to be able to use surgery bed, got error: %v", err)
	}
	if admission.BedID != "B001" {
		t.Errorf("Expected surgery bed B001 for general patient, got %s", admission.BedID)
	}
}

func TestMultiCondition_UnknownPatientCondition(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeSurgery, "102", 1)
	ba.AddBed("W001", "B003", BedTypeICU, "ICU-01", 2)
	ba.AddBed("W001", "B004", BedTypeIsolation, "103", 2)
	ba.AddPatient("P001", "张三", 45, "男", "genral")

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       45,
		PatientCondition: "genral",
		AdmitTime:        time.Now(),
	})
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed for unknown patient condition 'genral', got %v", err)
	}
}

func TestMultiCondition_EmptyPatientCondition(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeSurgery, "102", 1)
	ba.AddBed("W001", "B003", BedTypeICU, "ICU-01", 2)
	ba.AddPatient("P001", "张三", 45, "男", "")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       45,
		PatientCondition: "",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected empty condition to be treated as general, got error: %v", err)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Type != BedTypeGeneral && bed.Type != BedTypeSurgery {
		t.Errorf("Expected general or surgery bed for empty condition, got %s", bed.Type)
	}
}

func TestMultiCondition_UnknownConditionNoBedTypeSpecified(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddPatient("P001", "张三", 35, "男", "invalid_condition")

	_, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       35,
		PatientCondition: "invalid_condition",
		AdmitTime:        time.Now(),
	})
	if err != ErrNoAvailableBed {
		t.Errorf("Expected ErrNoAvailableBed for invalid condition, got %v", err)
	}
}

func TestMultiCondition_UnknownConditionWithBedTypeSpecified(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "内科病区", "内科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddPatient("P001", "张三", 35, "男", "invalid_condition")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          BedTypeGeneral,
		PatientID:        "P001",
		PatientAge:       35,
		PatientCondition: "invalid_condition",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected bed type specification to override condition check, got error: %v", err)
	}
	if admission.BedID != "B001" {
		t.Errorf("Expected bed B001 when bed type is explicitly specified, got %s", admission.BedID)
	}
}

func TestMultiCondition_SingleMatchLogicSource(t *testing.T) {
	ba := NewBedAllocator()
	ba.AddWard("W001", "ICU病区", "重症医学科")
	ba.AddBed("W001", "B001", BedTypeGeneral, "101", 1)
	ba.AddBed("W001", "B002", BedTypeICU, "ICU-01", 2)
	ba.AddPatient("P001", "张三", 55, "男", "icu")

	admission, err := ba.AllocateBed(AllocateCriteria{
		WardID:           "W001",
		BedType:          "",
		PatientID:        "P001",
		PatientAge:       55,
		PatientCondition: "icu",
		AdmitTime:        time.Now(),
	})
	if err != nil {
		t.Fatalf("Expected ICU patient to get ICU bed, got error: %v", err)
	}

	bed, _ := ba.GetBed(admission.BedID)
	if bed.Type != BedTypeICU {
		t.Errorf("Expected ICU bed type, got %s. Bed type filtering should only happen in matchPatientCondition.", bed.Type)
	}
	if bed.ID != "B002" {
		t.Errorf("Expected bed B002, got %s", bed.ID)
	}
}


