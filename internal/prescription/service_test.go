package prescription

import (
	"testing"
)

type TestContext struct {
	Store      *Store
	Service    *Service
	DoctorID   string
	PatientID  string
	PharmacyID string
	Med1ID     string
	Med2ID     string
	Med3ID     string
}

func setupTestStore() *TestContext {
	store := NewStore()
	svc := NewService(store)

	ctx := &TestContext{
		Store:   store,
		Service: svc,
	}

	ctx.DoctorID = store.AddDoctor(&Doctor{
		Name:       "张医生",
		Title:      "主任医师",
		Department: "内科",
		LicenseNo:  "DOC001",
	})

	ctx.PatientID = store.AddPatient(&Patient{
		Name:            "李患者",
		IDCard:          "110101199001011234",
		Phone:           "13800138000",
		Gender:          "男",
		Age:             35,
		MedicalRecordNo: "MR001",
	})

	ctx.PharmacyID = store.AddPharmacy(&Pharmacy{
		Name:      "和平药房",
		Address:   "北京市朝阳区和平路1号",
		Phone:     "010-12345678",
		LicenseNo: "PHY001",
	})

	ctx.Med1ID = store.AddMedicine(&Medicine{
		Name:          "阿莫西林胶囊",
		GenericName:   "Amoxicillin",
		Specification: "0.25g*24粒",
		Manufacturer:  "华北制药",
		UnitPrice:     25.50,
	})

	ctx.Med2ID = store.AddMedicine(&Medicine{
		Name:          "布洛芬缓释胶囊",
		GenericName:   "Ibuprofen",
		Specification: "0.3g*20粒",
		Manufacturer:  "中美史克",
		UnitPrice:     32.00,
	})

	ctx.Med3ID = store.AddMedicine(&Medicine{
		Name:          "维生素C片",
		GenericName:   "Vitamin C",
		Specification: "100mg*100片",
		Manufacturer:  "东北制药",
		UnitPrice:     8.50,
	})

	store.UpdateInventory(ctx.Med1ID, 100)
	store.UpdateInventory(ctx.Med2ID, 50)
	store.UpdateInventory(ctx.Med3ID, 200)

	return ctx
}

func TestCreatePrescription_Success(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{
				MedicineID:   ctx.Med1ID,
				Quantity:     2,
				Dosage:       "每次2粒",
				Frequency:    "每日3次",
				Duration:     "7天",
				Instructions: "饭后服用",
			},
			{
				MedicineID:   ctx.Med2ID,
				Quantity:     1,
				Dosage:       "每次1粒",
				Frequency:    "每日2次",
				Duration:     "3天",
				Instructions: "必要时服用",
			},
		},
		Diagnosis: "上呼吸道感染",
		Remark:    "注意休息，多饮水",
	}

	prescription, err := svc.CreatePrescription(req)
	if err != nil {
		t.Fatalf("CreatePrescription failed: %v", err)
	}

	if prescription == nil {
		t.Fatal("prescription should not be nil")
	}

	if prescription.ID == "" {
		t.Error("prescription ID should not be empty")
	}

	if prescription.Status != StatusPendingReview {
		t.Errorf("expected status %v, got %v", StatusPendingReview, prescription.Status)
	}

	if len(prescription.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(prescription.Items))
	}

	expectedAmount := 25.50*2 + 32.00*1
	if prescription.TotalAmount != expectedAmount {
		t.Errorf("expected total amount %.2f, got %.2f", expectedAmount, prescription.TotalAmount)
	}

	if prescription.DoctorID != ctx.DoctorID {
		t.Errorf("expected doctor ID %s, got %s", ctx.DoctorID, prescription.DoctorID)
	}

	if prescription.PatientID != ctx.PatientID {
		t.Errorf("expected patient ID %s, got %s", ctx.PatientID, prescription.PatientID)
	}

	if prescription.PharmacyID != ctx.PharmacyID {
		t.Errorf("expected pharmacy ID %s, got %s", ctx.PharmacyID, prescription.PharmacyID)
	}

	if prescription.Diagnosis != "上呼吸道感染" {
		t.Errorf("expected diagnosis '上呼吸道感染', got '%s'", prescription.Diagnosis)
	}

	if prescription.Remark != "注意休息，多饮水" {
		t.Errorf("expected remark '注意休息，多饮水', got '%s'", prescription.Remark)
	}

	inventory1, _ := store.GetInventory(ctx.Med1ID)
	if inventory1.Reserved != 2 {
		t.Errorf("expected reserved 2 for med1, got %d", inventory1.Reserved)
	}

	inventory2, _ := store.GetInventory(ctx.Med2ID)
	if inventory2.Reserved != 1 {
		t.Errorf("expected reserved 1 for med2, got %d", inventory2.Reserved)
	}
}

func TestCreatePrescription_NoItems(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items:      []MedicineItemRequest{},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

func TestCreatePrescription_InvalidDoctor(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   "INVALID_DOCTOR",
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestCreatePrescription_InvalidPatient(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  "INVALID_PATIENT",
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestCreatePrescription_InvalidPharmacy(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: "INVALID_PHARMACY",
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrPharmacyNotFound {
		t.Errorf("expected ErrPharmacyNotFound, got %v", err)
	}
}

func TestCreatePrescription_InvalidMedicine(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: "INVALID_MEDICINE", Quantity: 1},
		},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrMedicineNotFound {
		t.Errorf("expected ErrMedicineNotFound, got %v", err)
	}
}

func TestCreatePrescription_InvalidQuantity(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 0},
		},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrInvalidQuantity {
		t.Errorf("expected ErrInvalidQuantity, got %v", err)
	}

	req2 := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: -1},
		},
	}

	_, err2 := svc.CreatePrescription(req2)
	if err2 != ErrInvalidQuantity {
		t.Errorf("expected ErrInvalidQuantity for negative quantity, got %v", err2)
	}
}

func TestCreatePrescription_InsufficientStock(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1000},
		},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestReviewPrescription_Approved(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 2, Dosage: "每次2粒", Frequency: "每日3次"},
		},
		Diagnosis: "感冒",
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "王药师",
	}

	reviewed, err := svc.ReviewPrescription(reviewReq)
	if err != nil {
		t.Fatalf("ReviewPrescription failed: %v", err)
	}

	if reviewed.Status != StatusApproved {
		t.Errorf("expected status %v, got %v", StatusApproved, reviewed.Status)
	}

	if reviewed.ReviewedBy != "王药师" {
		t.Errorf("expected reviewed by '王药师', got '%s'", reviewed.ReviewedBy)
	}

	if reviewed.ReviewedAt == nil {
		t.Error("reviewed at should not be nil")
	}

	inventory, _ := store.GetInventory(ctx.Med1ID)
	if inventory.Reserved != 2 {
		t.Errorf("expected reserved 2, got %d", inventory.Reserved)
	}
}

func TestReviewPrescription_Rejected(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 5, Dosage: "每次2粒"},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       false,
		RejectReason:   "药品剂量超标，请医生重新开具",
		ReviewedBy:     "李药师",
	}

	reviewed, err := svc.ReviewPrescription(reviewReq)
	if err != nil {
		t.Fatalf("ReviewPrescription failed: %v", err)
	}

	if reviewed.Status != StatusRejected {
		t.Errorf("expected status %v, got %v", StatusRejected, reviewed.Status)
	}

	if reviewed.RejectReason != "药品剂量超标，请医生重新开具" {
		t.Errorf("expected reject reason, got '%s'", reviewed.RejectReason)
	}

	inventory, _ := store.GetInventory(ctx.Med1ID)
	if inventory.Reserved != 0 {
		t.Errorf("expected reserved 0 after rejection, got %d", inventory.Reserved)
	}
}

func TestReviewPrescription_RejectedWithoutReason(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	invBefore, _ := store.GetInventory(ctx.Med1ID)
	if invBefore.Reserved != 1 {
		t.Errorf("expected reserved 1 before review, got %d", invBefore.Reserved)
	}

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       false,
		RejectReason:   "",
		ReviewedBy:     "药师",
	}

	_, err := svc.ReviewPrescription(reviewReq)
	if err != ErrRejectReasonRequired {
		t.Errorf("expected ErrRejectReasonRequired, got %v", err)
	}

	updated, _ := svc.GetPrescription(prescription.ID)
	if updated.Status != StatusPendingReview {
		t.Errorf("expected status to remain PENDING_REVIEW, got %v", updated.Status)
	}

	invAfter, _ := store.GetInventory(ctx.Med1ID)
	if invAfter.Reserved != 1 {
		t.Errorf("expected reserved 1 after failed rejection, got %d", invAfter.Reserved)
	}
}

func TestReviewPrescription_RejectedWithEmptyReason_ReturnsError(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 5},
			{MedicineID: ctx.Med2ID, Quantity: 3},
		},
	}

	prescription, err := svc.CreatePrescription(createReq)
	if err != nil {
		t.Fatalf("CreatePrescription failed: %v", err)
	}

	invMed1Before, _ := store.GetInventory(ctx.Med1ID)
	invMed2Before, _ := store.GetInventory(ctx.Med2ID)
	if invMed1Before.Reserved != 5 {
		t.Errorf("expected med1 reserved 5, got %d", invMed1Before.Reserved)
	}
	if invMed2Before.Reserved != 3 {
		t.Errorf("expected med2 reserved 3, got %d", invMed2Before.Reserved)
	}

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       false,
		RejectReason:   "",
		ReviewedBy:     "审核药师",
	}

	result, err := svc.ReviewPrescription(reviewReq)
	if err != ErrRejectReasonRequired {
		t.Errorf("expected ErrRejectReasonRequired, got %v", err)
	}
	if result != nil {
		t.Error("expected result to be nil when error occurs")
	}

	updatedPrescription, _ := svc.GetPrescription(prescription.ID)
	if updatedPrescription.Status != StatusPendingReview {
		t.Errorf("prescription status should remain PENDING_REVIEW, got %v", updatedPrescription.Status)
	}
	if updatedPrescription.RejectReason != "" {
		t.Errorf("reject reason should remain empty, got '%s'", updatedPrescription.RejectReason)
	}
	if updatedPrescription.ReviewedAt != nil {
		t.Error("reviewed at should remain nil")
	}
	if updatedPrescription.ReviewedBy != "" {
		t.Errorf("reviewed by should remain empty, got '%s'", updatedPrescription.ReviewedBy)
	}

	invMed1After, _ := store.GetInventory(ctx.Med1ID)
	invMed2After, _ := store.GetInventory(ctx.Med2ID)
	if invMed1After.Reserved != 5 {
		t.Errorf("med1 reserved should remain 5, got %d", invMed1After.Reserved)
	}
	if invMed2After.Reserved != 3 {
		t.Errorf("med2 reserved should remain 3, got %d", invMed2After.Reserved)
	}
	if invMed1After.Quantity != 100 {
		t.Errorf("med1 quantity should remain 100, got %d", invMed1After.Quantity)
	}
	if invMed2After.Quantity != 50 {
		t.Errorf("med2 quantity should remain 50, got %d", invMed2After.Quantity)
	}
}

func TestReviewPrescription_InvalidPrescription(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	reviewReq := &ReviewRequest{
		PrescriptionID: "INVALID_PRESCRIPTION",
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}

	_, err := svc.ReviewPrescription(reviewReq)
	if err != ErrPrescriptionNotFound {
		t.Errorf("expected ErrPrescriptionNotFound, got %v", err)
	}
}

func TestReviewPrescription_WrongStatus(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	_, err := svc.ReviewPrescription(reviewReq)
	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestReviewPrescription_WrongPharmacy(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	otherPharmacyID := store.AddPharmacy(&Pharmacy{
		Name:    "其他药房",
		Address: "其他地址",
	})

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     otherPharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}

	_, err := svc.ReviewPrescription(reviewReq)
	if err != ErrPharmacyNotFound {
		t.Errorf("expected ErrPharmacyNotFound for wrong pharmacy, got %v", err)
	}
}

func TestDispensePrescription_Success(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 3},
			{MedicineID: ctx.Med2ID, Quantity: 2},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	dispenseReq := &DispenseRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		DispensedBy:    "张发药员",
	}

	dispensed, err := svc.DispensePrescription(dispenseReq)
	if err != nil {
		t.Fatalf("DispensePrescription failed: %v", err)
	}

	if dispensed.Status != StatusCompleted {
		t.Errorf("expected status %v, got %v", StatusCompleted, dispensed.Status)
	}

	if dispensed.DispensedBy != "张发药员" {
		t.Errorf("expected dispensed by '张发药员', got '%s'", dispensed.DispensedBy)
	}

	if dispensed.DispensedAt == nil {
		t.Error("dispensed at should not be nil")
	}

	inventory1, _ := store.GetInventory(ctx.Med1ID)
	if inventory1.Quantity != 97 {
		t.Errorf("expected med1 quantity 97, got %d", inventory1.Quantity)
	}
	if inventory1.Reserved != 0 {
		t.Errorf("expected med1 reserved 0, got %d", inventory1.Reserved)
	}

	inventory2, _ := store.GetInventory(ctx.Med2ID)
	if inventory2.Quantity != 48 {
		t.Errorf("expected med2 quantity 48, got %d", inventory2.Quantity)
	}
	if inventory2.Reserved != 0 {
		t.Errorf("expected med2 reserved 0, got %d", inventory2.Reserved)
	}
}

func TestDispensePrescription_Duplicate(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	dispenseReq := &DispenseRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		DispensedBy:    "发药员",
	}

	svc.DispensePrescription(dispenseReq)

	_, err := svc.DispensePrescription(dispenseReq)
	if err != ErrAlreadyDispensed {
		t.Errorf("expected ErrAlreadyDispensed, got %v", err)
	}
}

func TestDispensePrescription_NotApproved(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	dispenseReq := &DispenseRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		DispensedBy:    "发药员",
	}

	_, err := svc.DispensePrescription(dispenseReq)
	if err != ErrNotApproved {
		t.Errorf("expected ErrNotApproved, got %v", err)
	}
}

func TestDispensePrescription_RejectedCannotDispense(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       false,
		RejectReason:   "需要调整",
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	dispenseReq := &DispenseRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		DispensedBy:    "发药员",
	}

	_, err := svc.DispensePrescription(dispenseReq)
	if err != ErrNotApproved {
		t.Errorf("expected ErrNotApproved for rejected prescription, got %v", err)
	}
}

func TestDispensePrescription_InvalidPrescription(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	dispenseReq := &DispenseRequest{
		PrescriptionID: "INVALID_PRE",
		PharmacyID:     ctx.PharmacyID,
		DispensedBy:    "发药员",
	}

	_, err := svc.DispensePrescription(dispenseReq)
	if err != ErrPrescriptionNotFound {
		t.Errorf("expected ErrPrescriptionNotFound, got %v", err)
	}
}

func TestWithdrawPrescription_FromPending(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 5},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	withdrawReq := &WithdrawRequest{
		PrescriptionID: prescription.ID,
		Reason:         "诊断有误，需要重新开具",
	}

	withdrawn, err := svc.WithdrawPrescription(withdrawReq)
	if err != nil {
		t.Fatalf("WithdrawPrescription failed: %v", err)
	}

	if withdrawn.Status != StatusWithdrawn {
		t.Errorf("expected status %v, got %v", StatusWithdrawn, withdrawn.Status)
	}

	if withdrawn.WithdrawnReason != "诊断有误，需要重新开具" {
		t.Errorf("expected withdraw reason, got '%s'", withdrawn.WithdrawnReason)
	}

	if withdrawn.WithdrawnAt == nil {
		t.Error("withdrawn at should not be nil")
	}

	inventory, _ := store.GetInventory(ctx.Med1ID)
	if inventory.Reserved != 0 {
		t.Errorf("expected reserved 0 after withdrawal, got %d", inventory.Reserved)
	}
	if inventory.Quantity != 100 {
		t.Errorf("expected quantity 100 after withdrawal, got %d", inventory.Quantity)
	}
}

func TestWithdrawPrescription_FromApproved(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 10},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	withdrawReq := &WithdrawRequest{
		PrescriptionID: prescription.ID,
		Reason:         "患者放弃取药",
	}

	withdrawn, err := svc.WithdrawPrescription(withdrawReq)
	if err != nil {
		t.Fatalf("WithdrawPrescription failed: %v", err)
	}

	if withdrawn.Status != StatusWithdrawn {
		t.Errorf("expected status %v, got %v", StatusWithdrawn, withdrawn.Status)
	}

	inventory, _ := store.GetInventory(ctx.Med1ID)
	if inventory.Reserved != 0 {
		t.Errorf("expected reserved 0 after withdrawal, got %d", inventory.Reserved)
	}
	if inventory.Quantity != 100 {
		t.Errorf("expected quantity 100 after withdrawal, got %d", inventory.Quantity)
	}
}

func TestWithdrawPrescription_FromCompleted(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	dispenseReq := &DispenseRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		DispensedBy:    "发药员",
	}
	svc.DispensePrescription(dispenseReq)

	withdrawReq := &WithdrawRequest{
		PrescriptionID: prescription.ID,
		Reason:         "尝试撤回已发药处方",
	}

	_, err := svc.WithdrawPrescription(withdrawReq)
	if err != ErrCannotWithdraw {
		t.Errorf("expected ErrCannotWithdraw, got %v", err)
	}
}

func TestWithdrawPrescription_FromRejected(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	prescription, _ := svc.CreatePrescription(createReq)

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       false,
		RejectReason:   "驳回",
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	withdrawReq := &WithdrawRequest{
		PrescriptionID: prescription.ID,
		Reason:         "尝试撤回已驳回处方",
	}

	_, err := svc.WithdrawPrescription(withdrawReq)
	if err != ErrCannotWithdraw {
		t.Errorf("expected ErrCannotWithdraw for rejected prescription, got %v", err)
	}
}

func TestWithdrawPrescription_InvalidPrescription(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	withdrawReq := &WithdrawRequest{
		PrescriptionID: "INVALID_PRE",
		Reason:         "测试",
	}

	_, err := svc.WithdrawPrescription(withdrawReq)
	if err != ErrPrescriptionNotFound {
		t.Errorf("expected ErrPrescriptionNotFound, got %v", err)
	}
}

func TestWithdrawPrescription_EmptyReason(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 3},
		},
	}

	prescription, err := svc.CreatePrescription(createReq)
	if err != nil {
		t.Fatalf("CreatePrescription failed: %v", err)
	}

	invBefore, _ := store.GetInventory(ctx.Med1ID)
	if invBefore.Reserved != 3 {
		t.Errorf("expected reserved 3 before withdraw, got %d", invBefore.Reserved)
	}

	withdrawReq := &WithdrawRequest{
		PrescriptionID: prescription.ID,
		Reason:         "",
	}

	result, err := svc.WithdrawPrescription(withdrawReq)
	if err != ErrWithdrawReasonRequired {
		t.Errorf("expected ErrWithdrawReasonRequired, got %v", err)
	}
	if result != nil {
		t.Error("expected result to be nil when error occurs")
	}

	updatedPrescription, _ := svc.GetPrescription(prescription.ID)
	if updatedPrescription.Status != StatusPendingReview {
		t.Errorf("prescription status should remain PENDING_REVIEW, got %v", updatedPrescription.Status)
	}
	if updatedPrescription.WithdrawnReason != "" {
		t.Errorf("withdrawn reason should remain empty, got '%s'", updatedPrescription.WithdrawnReason)
	}
	if updatedPrescription.WithdrawnAt != nil {
		t.Error("withdrawn at should remain nil")
	}

	invAfter, _ := store.GetInventory(ctx.Med1ID)
	if invAfter.Reserved != 3 {
		t.Errorf("reserved should remain 3 after failed withdrawal, got %d", invAfter.Reserved)
	}
	if invAfter.Quantity != 100 {
		t.Errorf("quantity should remain 100, got %d", invAfter.Quantity)
	}
}

func TestWithdrawPrescription_EmptyReasonFromApproved(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 4},
			{MedicineID: ctx.Med2ID, Quantity: 2},
		},
	}

	prescription, err := svc.CreatePrescription(createReq)
	if err != nil {
		t.Fatalf("CreatePrescription failed: %v", err)
	}

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "药师",
	}
	svc.ReviewPrescription(reviewReq)

	invMed1Before, _ := store.GetInventory(ctx.Med1ID)
	invMed2Before, _ := store.GetInventory(ctx.Med2ID)
	if invMed1Before.Reserved != 4 {
		t.Errorf("expected med1 reserved 4, got %d", invMed1Before.Reserved)
	}
	if invMed2Before.Reserved != 2 {
		t.Errorf("expected med2 reserved 2, got %d", invMed2Before.Reserved)
	}

	withdrawReq := &WithdrawRequest{
		PrescriptionID: prescription.ID,
		Reason:         "",
	}

	result, err := svc.WithdrawPrescription(withdrawReq)
	if err != ErrWithdrawReasonRequired {
		t.Errorf("expected ErrWithdrawReasonRequired, got %v", err)
	}
	if result != nil {
		t.Error("expected result to be nil when error occurs")
	}

	updatedPrescription, _ := svc.GetPrescription(prescription.ID)
	if updatedPrescription.Status != StatusApproved {
		t.Errorf("prescription status should remain APPROVED, got %v", updatedPrescription.Status)
	}

	invMed1After, _ := store.GetInventory(ctx.Med1ID)
	invMed2After, _ := store.GetInventory(ctx.Med2ID)
	if invMed1After.Reserved != 4 {
		t.Errorf("med1 reserved should remain 4, got %d", invMed1After.Reserved)
	}
	if invMed2After.Reserved != 2 {
		t.Errorf("med2 reserved should remain 2, got %d", invMed2After.Reserved)
	}
	if invMed1After.Quantity != 100 {
		t.Errorf("med1 quantity should remain 100, got %d", invMed1After.Quantity)
	}
	if invMed2After.Quantity != 50 {
		t.Errorf("med2 quantity should remain 50, got %d", invMed2After.Quantity)
	}
}

func TestCreatePrescription_RollbackOnInvalidQuantity(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 2},
			{MedicineID: ctx.Med2ID, Quantity: 0},
		},
	}

	_, err := svc.CreatePrescription(createReq)
	if err != ErrInvalidQuantity {
		t.Errorf("expected ErrInvalidQuantity, got %v", err)
	}

	invMed1, _ := store.GetInventory(ctx.Med1ID)
	if invMed1.Reserved != 0 {
		t.Errorf("med1 reserved should be rolled back to 0, got %d", invMed1.Reserved)
	}
	if invMed1.Quantity != 100 {
		t.Errorf("med1 quantity should remain 100, got %d", invMed1.Quantity)
	}

	invMed2, _ := store.GetInventory(ctx.Med2ID)
	if invMed2.Reserved != 0 {
		t.Errorf("med2 reserved should be 0, got %d", invMed2.Reserved)
	}
}

func TestCreatePrescription_RollbackOnInvalidMedicine(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 3},
			{MedicineID: "INVALID_MED", Quantity: 1},
		},
	}

	_, err := svc.CreatePrescription(createReq)
	if err != ErrMedicineNotFound {
		t.Errorf("expected ErrMedicineNotFound, got %v", err)
	}

	invMed1, _ := store.GetInventory(ctx.Med1ID)
	if invMed1.Reserved != 0 {
		t.Errorf("med1 reserved should be rolled back to 0, got %d", invMed1.Reserved)
	}
	if invMed1.Quantity != 100 {
		t.Errorf("med1 quantity should remain 100, got %d", invMed1.Quantity)
	}
}

func TestCreatePrescription_RollbackOnInsufficientStock(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 5},
			{MedicineID: ctx.Med2ID, Quantity: 999},
		},
	}

	_, err := svc.CreatePrescription(createReq)
	if err != ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	invMed1, _ := store.GetInventory(ctx.Med1ID)
	if invMed1.Reserved != 0 {
		t.Errorf("med1 reserved should be rolled back to 0, got %d", invMed1.Reserved)
	}
	if invMed1.Quantity != 100 {
		t.Errorf("med1 quantity should remain 100, got %d", invMed1.Quantity)
	}

	invMed2, _ := store.GetInventory(ctx.Med2ID)
	if invMed2.Reserved != 0 {
		t.Errorf("med2 reserved should be 0, got %d", invMed2.Reserved)
	}
	if invMed2.Quantity != 50 {
		t.Errorf("med2 quantity should remain 50, got %d", invMed2.Quantity)
	}
}

func TestFullWorkflow(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 2, Dosage: "2粒", Frequency: "3次/日", Duration: "7天", Instructions: "饭后"},
			{MedicineID: ctx.Med3ID, Quantity: 1, Dosage: "2片", Frequency: "1次/日", Duration: "30天", Instructions: "每日一次"},
		},
		Diagnosis: "支气管炎",
		Remark:    "随诊",
	}

	prescription, err := svc.CreatePrescription(createReq)
	if err != nil {
		t.Fatalf("Step 1 - Create failed: %v", err)
	}
	if prescription.Status != StatusPendingReview {
		t.Fatalf("Step 1 - Expected PENDING_REVIEW, got %v", prescription.Status)
	}

	inv1Before, _ := store.GetInventory(ctx.Med1ID)
	if inv1Before.Reserved != 2 {
		t.Errorf("Step 1 - Expected med1 reserved 2, got %d", inv1Before.Reserved)
	}

	reviewReq := &ReviewRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		Approved:       true,
		ReviewedBy:     "赵药师",
	}
	reviewed, err := svc.ReviewPrescription(reviewReq)
	if err != nil {
		t.Fatalf("Step 2 - Review failed: %v", err)
	}
	if reviewed.Status != StatusApproved {
		t.Fatalf("Step 2 - Expected APPROVED, got %v", reviewed.Status)
	}

	dispenseReq := &DispenseRequest{
		PrescriptionID: prescription.ID,
		PharmacyID:     ctx.PharmacyID,
		DispensedBy:    "孙发药",
	}
	completed, err := svc.DispensePrescription(dispenseReq)
	if err != nil {
		t.Fatalf("Step 3 - Dispense failed: %v", err)
	}
	if completed.Status != StatusCompleted {
		t.Fatalf("Step 3 - Expected COMPLETED, got %v", completed.Status)
	}

	inv1After, _ := store.GetInventory(ctx.Med1ID)
	if inv1After.Quantity != 98 {
		t.Errorf("Step 3 - Expected med1 quantity 98, got %d", inv1After.Quantity)
	}
	if inv1After.Reserved != 0 {
		t.Errorf("Step 3 - Expected med1 reserved 0, got %d", inv1After.Reserved)
	}

	inv3After, _ := store.GetInventory(ctx.Med3ID)
	if inv3After.Quantity != 199 {
		t.Errorf("Step 3 - Expected med3 quantity 199, got %d", inv3After.Quantity)
	}
}

func TestConcurrentPrescriptionStock(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	done := make(chan bool, 50)
	var successCount int
	var failCount int

	for i := 0; i < 50; i++ {
		go func() {
			req := &CreatePrescriptionRequest{
				DoctorID:   ctx.DoctorID,
				PatientID:  ctx.PatientID,
				PharmacyID: ctx.PharmacyID,
				Items: []MedicineItemRequest{
					{MedicineID: ctx.Med1ID, Quantity: 3},
				},
			}
			_, err := svc.CreatePrescription(req)
			if err == nil {
				successCount++
			} else {
				failCount++
			}
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}

	inventory, _ := store.GetInventory(ctx.Med1ID)
	maxPossible := 100 / 3
	actualReserved := inventory.Reserved / 3

	if actualReserved > maxPossible {
		t.Errorf("Over-reservation detected: reserved %d items, max possible %d", actualReserved, maxPossible)
	}

	t.Logf("Concurrent test results: %d successes, %d failures, reserved %d items", successCount, failCount, inventory.Reserved)
}

func TestListPrescriptions(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	for i := 0; i < 3; i++ {
		req := &CreatePrescriptionRequest{
			DoctorID:   ctx.DoctorID,
			PatientID:  ctx.PatientID,
			PharmacyID: ctx.PharmacyID,
			Items: []MedicineItemRequest{
				{MedicineID: ctx.Med1ID, Quantity: 1},
			},
		}
		svc.CreatePrescription(req)
	}

	patientPrescriptions, err := svc.ListPrescriptionsByPatient(ctx.PatientID)
	if err != nil {
		t.Fatalf("ListPrescriptionsByPatient failed: %v", err)
	}
	if len(patientPrescriptions) != 3 {
		t.Errorf("Expected 3 prescriptions for patient, got %d", len(patientPrescriptions))
	}

	pharmacyPrescriptions, err := svc.ListPrescriptionsByPharmacy(ctx.PharmacyID)
	if err != nil {
		t.Fatalf("ListPrescriptionsByPharmacy failed: %v", err)
	}
	if len(pharmacyPrescriptions) != 3 {
		t.Errorf("Expected 3 prescriptions for pharmacy, got %d", len(pharmacyPrescriptions))
	}

	doctorPrescriptions, err := svc.ListPrescriptionsByDoctor(ctx.DoctorID)
	if err != nil {
		t.Fatalf("ListPrescriptionsByDoctor failed: %v", err)
	}
	if len(doctorPrescriptions) != 3 {
		t.Errorf("Expected 3 prescriptions for doctor, got %d", len(doctorPrescriptions))
	}
}

func TestListPrescriptions_InvalidIDs(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	_, err := svc.ListPrescriptionsByPatient("INVALID")
	if err != ErrPatientNotFound {
		t.Errorf("Expected ErrPatientNotFound, got %v", err)
	}

	_, err = svc.ListPrescriptionsByPharmacy("INVALID")
	if err != ErrPharmacyNotFound {
		t.Errorf("Expected ErrPharmacyNotFound, got %v", err)
	}

	_, err = svc.ListPrescriptionsByDoctor("INVALID")
	if err != ErrDoctorNotFound {
		t.Errorf("Expected ErrDoctorNotFound, got %v", err)
	}
}

func TestGetInventory_NonExistentMedicine(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store

	_, err := store.GetInventory("INVALID_MED")
	if err != ErrMedicineNotFound {
		t.Errorf("Expected ErrMedicineNotFound, got %v", err)
	}
}

func TestUpdateInventory(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store

	err := store.UpdateInventory(ctx.Med1ID, -200)
	if err != ErrInsufficientStock {
		t.Errorf("Expected ErrInsufficientStock for negative inventory, got %v", err)
	}

	inventory, _ := store.GetInventory(ctx.Med1ID)
	if inventory.Quantity != 100 {
		t.Errorf("Expected quantity to remain 100, got %d", inventory.Quantity)
	}
}

func TestMedicineItemDetails(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service
	med1, _ := store.GetMedicine(ctx.Med1ID)

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{
				MedicineID:   ctx.Med1ID,
				Quantity:     3,
				Dosage:       "每次1粒",
				Frequency:    "每日三次",
				Duration:     "一周",
				Instructions: "温开水送服",
			},
		},
	}

	prescription, _ := svc.CreatePrescription(req)

	if len(prescription.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(prescription.Items))
	}

	item := prescription.Items[0]
	if item.MedicineID != ctx.Med1ID {
		t.Errorf("Expected medicine ID %s, got %s", ctx.Med1ID, item.MedicineID)
	}
	if item.MedicineName != med1.Name {
		t.Errorf("Expected medicine name %s, got %s", med1.Name, item.MedicineName)
	}
	if item.Quantity != 3 {
		t.Errorf("Expected quantity 3, got %d", item.Quantity)
	}
	if item.UnitPrice != med1.UnitPrice {
		t.Errorf("Expected unit price %.2f, got %.2f", med1.UnitPrice, item.UnitPrice)
	}
	if item.Dosage != "每次1粒" {
		t.Errorf("Expected dosage '每次1粒', got '%s'", item.Dosage)
	}
	if item.Frequency != "每日三次" {
		t.Errorf("Expected frequency '每日三次', got '%s'", item.Frequency)
	}
	if item.Duration != "一周" {
		t.Errorf("Expected duration '一周', got '%s'", item.Duration)
	}
	if item.Instructions != "温开水送服" {
		t.Errorf("Expected instructions '温开水送服', got '%s'", item.Instructions)
	}
}

func TestPrescriptionStockPartialFailure(t *testing.T) {
	ctx := setupTestStore()
	store := ctx.Store
	svc := ctx.Service

	med2Inv, _ := store.GetInventory(ctx.Med2ID)
	t.Logf("Initial med2 stock: %d", med2Inv.Quantity)

	req := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
			{MedicineID: ctx.Med2ID, Quantity: 9999},
		},
	}

	_, err := svc.CreatePrescription(req)
	if err != ErrInsufficientStock {
		t.Errorf("Expected ErrInsufficientStock, got %v", err)
	}

	med1Inv, _ := store.GetInventory(ctx.Med1ID)
	if med1Inv.Reserved != 0 {
		t.Errorf("med1 should not be reserved on partial failure, got %d reserved", med1Inv.Reserved)
	}
}

func TestGetPrescription(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	createReq := &CreatePrescriptionRequest{
		DoctorID:   ctx.DoctorID,
		PatientID:  ctx.PatientID,
		PharmacyID: ctx.PharmacyID,
		Items: []MedicineItemRequest{
			{MedicineID: ctx.Med1ID, Quantity: 1},
		},
	}

	created, _ := svc.CreatePrescription(createReq)

	fetched, err := svc.GetPrescription(created.ID)
	if err != nil {
		t.Fatalf("GetPrescription failed: %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, fetched.ID)
	}

	_, err = svc.GetPrescription("NON_EXISTENT")
	if err != ErrPrescriptionNotFound {
		t.Errorf("Expected ErrPrescriptionNotFound, got %v", err)
	}
}
