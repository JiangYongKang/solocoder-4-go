package checkup

import (
	"sync"
	"testing"
	"time"
)

func setupTestStore() *Store {
	store := NewStore()

	store.AddPatient(&Patient{
		Name:   "张三",
		Gender: "男",
		Age:    35,
		Phone:  "13800138000",
	})

	store.AddPatient(&Patient{
		Name:   "李四",
		Gender: "女",
		Age:    28,
		Phone:  "13900139000",
	})

	store.AddItem(&AddItemRequest{
		Name:        "血常规-白细胞计数",
		Description: "检测血液中白细胞数量",
		Category:    ItemCategoryLaboratory,
		Unit:        "×10⁹/L",
		MinValue:    floatPtr(4.0),
		MaxValue:    floatPtr(10.0),
		Price:       25.0,
	})

	store.AddItem(&AddItemRequest{
		Name:        "血常规-红细胞计数",
		Description: "检测血液中红细胞数量",
		Category:    ItemCategoryLaboratory,
		Unit:        "×10¹²/L",
		MinValue:    floatPtr(4.3),
		MaxValue:    floatPtr(5.8),
		Price:       20.0,
	})

	store.AddItem(&AddItemRequest{
		Name:        "肝功能-谷丙转氨酶",
		Description: "检测肝脏功能指标",
		Category:    ItemCategoryLaboratory,
		Unit:        "U/L",
		MinValue:    floatPtr(0),
		MaxValue:    floatPtr(40),
		Price:       30.0,
	})

	store.AddItem(&AddItemRequest{
		Name:        "胸部X光",
		Description: "胸部X线检查",
		Category:    ItemCategoryImaging,
		Unit:        "",
		MinValue:    nil,
		MaxValue:    nil,
		Price:       100.0,
	})

	store.AddItem(&AddItemRequest{
		Name:        "心电图",
		Description: "检测心脏电活动",
		Category:    ItemCategoryFunctional,
		Unit:        "",
		MinValue:    nil,
		MaxValue:    nil,
		Price:       50.0,
	})

	return store
}

func getFirstPatientID(store *Store) string {
	for id := range store.patients {
		return id
	}
	return ""
}

func getSecondPatientID(store *Store) string {
	keys := make([]string, 0, len(store.patients))
	for id := range store.patients {
		keys = append(keys, id)
	}
	if len(keys) >= 2 {
		return keys[1]
	}
	return ""
}

func getItemIDByName(store *Store, name string) string {
	for id, item := range store.items {
		if item.Name == name {
			return id
		}
	}
	return ""
}

func TestAddPatient_Success(t *testing.T) {
	store := NewStore()
	p := &Patient{Name: "测试患者"}
	id := store.AddPatient(p)
	if id == "" {
		t.Error("patient ID should not be empty")
	}
	got, err := store.GetPatient(id)
	if err != nil {
		t.Fatalf("GetPatient failed: %v", err)
	}
	if got.Name != "测试患者" {
		t.Errorf("expected name '测试患者', got '%s'", got.Name)
	}
}

func TestGetPatient_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetPatient("NONEXIST")
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestAddItem_Success(t *testing.T) {
	store := NewStore()
	item := store.AddItem(&AddItemRequest{
		Name:     "测试项目",
		Category: ItemCategoryPhysical,
		Unit:     "mmHg",
		MinValue: floatPtr(60),
		MaxValue: floatPtr(90),
		Price:    15.0,
	})
	if item.ID == "" {
		t.Error("item ID should not be empty")
	}
	if item.Name != "测试项目" {
		t.Errorf("expected name '测试项目', got '%s'", item.Name)
	}
	got, err := store.GetItem(item.ID)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}
	if got.Price != 15.0 {
		t.Errorf("expected price 15.0, got %v", got.Price)
	}
}

func TestGetItem_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetItem("NONEXIST")
	if err != ErrItemNotFound {
		t.Errorf("expected ErrItemNotFound, got %v", err)
	}
}

func TestListItems(t *testing.T) {
	store := setupTestStore()
	items := store.ListItems()
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}
}

func TestCreatePackage_Success(t *testing.T) {
	store := setupTestStore()
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	item2 := getItemIDByName(store, "血常规-红细胞计数")
	item3 := getItemIDByName(store, "肝功能-谷丙转氨酶")

	pkg, err := store.CreatePackage(&CreatePackageRequest{
		Name:        "基础体检套餐",
		Description: "包含基础血液检查和肝功能",
		ItemIDs:     []string{item1, item2, item3},
	})
	if err != nil {
		t.Fatalf("CreatePackage failed: %v", err)
	}
	if pkg.ID == "" {
		t.Error("package ID should not be empty")
	}
	if len(pkg.Items) != 3 {
		t.Errorf("expected 3 items in package, got %d", len(pkg.Items))
	}
	expectedPrice := 25.0 + 20.0 + 30.0
	if pkg.TotalPrice != expectedPrice {
		t.Errorf("expected total price %.1f, got %.1f", expectedPrice, pkg.TotalPrice)
	}
}

func TestCreatePackage_EmptyItems(t *testing.T) {
	store := setupTestStore()
	_, err := store.CreatePackage(&CreatePackageRequest{
		Name:    "空套餐",
		ItemIDs: []string{},
	})
	if err != ErrEmptyPackageItems {
		t.Errorf("expected ErrEmptyPackageItems, got %v", err)
	}
}

func TestCreatePackage_DuplicateItems(t *testing.T) {
	store := setupTestStore()
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	_, err := store.CreatePackage(&CreatePackageRequest{
		Name:    "重复项目套餐",
		ItemIDs: []string{item1, item1},
	})
	if err != ErrDuplicateItemInPackage {
		t.Errorf("expected ErrDuplicateItemInPackage, got %v", err)
	}
}

func TestCreatePackage_InvalidItem(t *testing.T) {
	store := setupTestStore()
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	_, err := store.CreatePackage(&CreatePackageRequest{
		Name:    "含无效项目套餐",
		ItemIDs: []string{item1, "INVALID_ITEM_ID"},
	})
	if err == nil {
		t.Fatal("expected error for invalid item, got nil")
	}
}

func TestGetPackage_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetPackage("NONEXIST")
	if err != ErrPackageNotFound {
		t.Errorf("expected ErrPackageNotFound, got %v", err)
	}
}

func TestListPackages(t *testing.T) {
	store := setupTestStore()
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	store.CreatePackage(&CreatePackageRequest{Name: "套餐A", ItemIDs: []string{item1}})
	item2 := getItemIDByName(store, "胸部X光")
	store.CreatePackage(&CreatePackageRequest{Name: "套餐B", ItemIDs: []string{item2}})
	pkgs := store.ListPackages()
	if len(pkgs) != 2 {
		t.Errorf("expected 2 packages, got %d", len(pkgs))
	}
}

func TestCreateTimeSlot_Success(t *testing.T) {
	store := NewStore()
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)

	slot, err := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  20,
	})
	if err != nil {
		t.Fatalf("CreateTimeSlot failed: %v", err)
	}
	if slot.ID == "" {
		t.Error("time slot ID should not be empty")
	}
	if slot.CurrentCount != 0 {
		t.Errorf("expected current count 0, got %d", slot.CurrentCount)
	}
	if slot.Capacity != 20 {
		t.Errorf("expected capacity 20, got %d", slot.Capacity)
	}
}

func TestCreateTimeSlot_InvalidTimeRange(t *testing.T) {
	store := NewStore()
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)

	_, err := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	if err != ErrInvalidTimeRange {
		t.Errorf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestCreateTimeSlot_InvalidCapacity(t *testing.T) {
	store := NewStore()
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)

	_, err := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  0,
	})
	if err != ErrInvalidCapacity {
		t.Errorf("expected ErrInvalidCapacity, got %v", err)
	}
}

func TestGetTimeSlot_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetTimeSlot("NONEXIST")
	if err != ErrTimeSlotNotFound {
		t.Errorf("expected ErrTimeSlotNotFound, got %v", err)
	}
}

func TestCreateAppointment_Success(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	item2 := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item1, item2},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})

	appt, err := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}
	if appt.ID == "" {
		t.Error("appointment ID should not be empty")
	}
	if appt.Status != AppointmentStatusPending {
		t.Errorf("expected status PENDING, got %v", appt.Status)
	}
	updatedSlot, _ := store.GetTimeSlot(slot.ID)
	if updatedSlot.CurrentCount != 1 {
		t.Errorf("expected slot current count 1, got %d", updatedSlot.CurrentCount)
	}
}

func TestCreateAppointment_PatientNotFound(t *testing.T) {
	store := setupTestStore()
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item1},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})

	_, err := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  "INVALID_PATIENT",
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestCreateAppointment_PackageNotFound(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})

	_, err := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  "INVALID_PACKAGE",
		TimeSlotID: slot.ID,
	})
	if err != ErrPackageNotFound {
		t.Errorf("expected ErrPackageNotFound, got %v", err)
	}
}

func TestCreateAppointment_TimeSlotNotFound(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item1},
	})

	_, err := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: "INVALID_SLOT",
	})
	if err != ErrTimeSlotNotFound {
		t.Errorf("expected ErrTimeSlotNotFound, got %v", err)
	}
}

func TestCreateAppointment_CapacityFull(t *testing.T) {
	store := setupTestStore()
	patient1 := getFirstPatientID(store)
	patient2 := getSecondPatientID(store)
	store.AddPatient(&Patient{Name: "王五"})
	var patient3 string
	for id := range store.patients {
		if id != patient1 && id != patient2 {
			patient3 = id
			break
		}
	}
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item1},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  2,
	})

	_, err := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patient1,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != nil {
		t.Fatalf("First appointment failed: %v", err)
	}

	_, err = store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patient2,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != nil {
		t.Fatalf("Second appointment failed: %v", err)
	}

	_, err = store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patient3,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != ErrTimeSlotCapacityFull {
		t.Errorf("expected ErrTimeSlotCapacityFull, got %v", err)
	}
}

func TestCancelAppointment_Success(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	item1 := getItemIDByName(store, "血常规-白细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item1},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})

	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	cancelled, err := store.CancelAppointment(appt.ID)
	if err != nil {
		t.Fatalf("CancelAppointment failed: %v", err)
	}
	if cancelled.Status != AppointmentStatusCancelled {
		t.Errorf("expected status CANCELLED, got %v", cancelled.Status)
	}
	updatedSlot, _ := store.GetTimeSlot(slot.ID)
	if updatedSlot.CurrentCount != 0 {
		t.Errorf("expected slot current count 0 after cancel, got %d", updatedSlot.CurrentCount)
	}
}

func TestCancelAppointment_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.CancelAppointment("NONEXIST")
	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestGetAppointment_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetAppointment("NONEXIST")
	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestRecordResult_Success_Normal(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
		Remarks:       "",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if result.IsAbnormal {
		t.Error("expected result to be normal (6.5 is within 4.0-10.0)")
	}
	if !result.IsNumeric {
		t.Error("expected result to be numeric")
	}
	if result.NumericValue != 6.5 {
		t.Errorf("expected numeric value 6.5, got %v", result.NumericValue)
	}

	updatedAppt, _ := store.GetAppointment(appt.ID)
	if updatedAppt.Status != AppointmentStatusChecking {
		t.Errorf("expected status CHECKING, got %v", updatedAppt.Status)
	}
}

func TestRecordResult_Success_Abnormal(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "15.0",
		Remarks:       "偏高，建议复查",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if !result.IsAbnormal {
		t.Error("expected result to be abnormal (15.0 is above 10.0)")
	}
}

func TestRecordResult_AllItemsCompleted(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemRBC,
		Value:         "5.0",
	})

	updatedAppt, _ := store.GetAppointment(appt.ID)
	if updatedAppt.Status != AppointmentStatusCompleted {
		t.Errorf("expected status COMPLETED, got %v", updatedAppt.Status)
	}
}

func TestRecordResult_AppointmentNotFound(t *testing.T) {
	store := setupTestStore()
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	_, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: "INVALID_APPT",
		ItemID:        itemWBC,
		Value:         "6.5",
	})
	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestRecordResult_ItemNotInPackage(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemECG := getItemIDByName(store, "心电图")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	_, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemECG,
		Value:         "窦性心律",
	})
	if err != ErrItemNotInPackage {
		t.Errorf("expected ErrItemNotInPackage, got %v", err)
	}
}

func TestRecordResult_DuplicateResult(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})

	_, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "7.0",
	})
	if err != ErrDuplicateResult {
		t.Errorf("expected ErrDuplicateResult, got %v", err)
	}
}

func TestRecordResult_CancelledAppointment(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	store.CancelAppointment(appt.ID)

	_, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})
	if err != ErrAppointmentCancelled {
		t.Errorf("expected ErrAppointmentCancelled, got %v", err)
	}
}

func TestGetResultsByAppointment(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})

	results, err := store.GetResultsByAppointment(appt.ID)
	if err != nil {
		t.Fatalf("GetResultsByAppointment failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestGetResultsByAppointment_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetResultsByAppointment("NONEXIST")
	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestGetResult_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetResult("NONEXIST")
	if err != ErrResultNotFound {
		t.Errorf("expected ErrResultNotFound, got %v", err)
	}
}

func TestGenerateReport_Success_NoAbnormal(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	itemALT := getItemIDByName(store, "肝功能-谷丙转氨酶")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "基础体检套餐",
		ItemIDs: []string{itemWBC, itemRBC, itemALT},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemRBC,
		Value:         "5.0",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemALT,
		Value:         "25",
	})

	report, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}
	if report.ID == "" {
		t.Error("report ID should not be empty")
	}
	if len(report.AbnormalItems) != 0 {
		t.Errorf("expected 0 abnormal items, got %d", len(report.AbnormalItems))
	}
	if len(report.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(report.Results))
	}
}

func TestGenerateReport_Success_WithAbnormal(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	itemALT := getItemIDByName(store, "肝功能-谷丙转氨酶")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "基础体检套餐",
		ItemIDs: []string{itemWBC, itemRBC, itemALT},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "15.0",
		Remarks:       "偏高",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemRBC,
		Value:         "5.0",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemALT,
		Value:         "65",
		Remarks:       "偏高",
	})

	report, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}
	if len(report.AbnormalItems) != 2 {
		t.Errorf("expected 2 abnormal items, got %d", len(report.AbnormalItems))
	}
}

func TestGenerateReport_ResultsIncomplete(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	itemALT := getItemIDByName(store, "肝功能-谷丙转氨酶")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "基础体检套餐",
		ItemIDs: []string{itemWBC, itemRBC, itemALT},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})

	_, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != ErrResultsIncomplete {
		t.Errorf("expected ErrResultsIncomplete, got %v", err)
	}
}

func TestGenerateReport_AppointmentNotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: "NONEXIST",
	})
	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestGenerateReport_CancelledAppointment(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	store.CancelAppointment(appt.ID)

	_, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != ErrAppointmentCancelled {
		t.Errorf("expected ErrAppointmentCancelled, got %v", err)
	}
}

func TestGetReport_NotFound(t *testing.T) {
	store := NewStore()
	_, err := store.GetReport("NONEXIST")
	if err != ErrReportNotFound {
		t.Errorf("expected ErrReportNotFound, got %v", err)
	}
}

func TestGetReport_Success(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemRBC,
		Value:         "5.0",
	})
	generated, _ := store.GenerateReport(&GenerateReportRequest{AppointmentID: appt.ID})

	report, err := store.GetReport(appt.ID)
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}
	if report.ID != generated.ID {
		t.Errorf("expected report ID %s, got %s", generated.ID, report.ID)
	}
}

func TestFullWorkflow(t *testing.T) {
	store := setupTestStore()

	patientID := getFirstPatientID(store)

	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	itemALT := getItemIDByName(store, "肝功能-谷丙转氨酶")
	itemChest := getItemIDByName(store, "胸部X光")
	itemECG := getItemIDByName(store, "心电图")

	pkg, err := store.CreatePackage(&CreatePackageRequest{
		Name:        "全面体检套餐A",
		Description: "包含血液检查、肝功能、影像和功能检查",
		ItemIDs:     []string{itemWBC, itemRBC, itemALT, itemChest, itemECG},
	})
	if err != nil {
		t.Fatalf("CreatePackage failed: %v", err)
	}
	if pkg.TotalPrice != 25+20+30+100+50 {
		t.Errorf("expected total price 225, got %.1f", pkg.TotalPrice)
	}

	date := time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 15, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 15, 12, 0, 0, 0, time.Local)
	slot, err := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  50,
	})
	if err != nil {
		t.Fatalf("CreateTimeSlot failed: %v", err)
	}

	appt, err := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}
	if appt.Status != AppointmentStatusPending {
		t.Errorf("expected status PENDING, got %v", appt.Status)
	}

	_, err = store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "12.5",
		Remarks:       "偏高，建议复查",
	})
	if err != nil {
		t.Fatalf("Record WBC failed: %v", err)
	}

	_, err = store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemRBC,
		Value:         "4.8",
	})
	if err != nil {
		t.Fatalf("Record RBC failed: %v", err)
	}

	_, err = store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemALT,
		Value:         "35",
	})
	if err != nil {
		t.Fatalf("Record ALT failed: %v", err)
	}

	_, err = store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemChest,
		Value:         "未见异常",
		Remarks:       "心肺正常",
	})
	if err != nil {
		t.Fatalf("Record Chest X-ray failed: %v", err)
	}

	_, err = store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemECG,
		Value:         "窦性心律",
		Remarks:       "正常心电图",
	})
	if err != nil {
		t.Fatalf("Record ECG failed: %v", err)
	}

	updatedAppt, _ := store.GetAppointment(appt.ID)
	if updatedAppt.Status != AppointmentStatusCompleted {
		t.Errorf("expected status COMPLETED, got %v", updatedAppt.Status)
	}

	report, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}
	if len(report.AbnormalItems) != 1 {
		t.Errorf("expected 1 abnormal item (WBC), got %d", len(report.AbnormalItems))
	}
	if len(report.AbnormalItems) > 0 && report.AbnormalItems[0].ItemName != "血常规-白细胞计数" {
		t.Errorf("expected abnormal item WBC, got %s", report.AbnormalItems[0].ItemName)
	}
}

func TestConcurrentAppointments(t *testing.T) {
	store := NewStore()
	store.AddPatient(&Patient{Name: "患者1"})
	store.AddPatient(&Patient{Name: "患者2"})
	store.AddPatient(&Patient{Name: "患者3"})
	store.AddPatient(&Patient{Name: "患者4"})
	store.AddPatient(&Patient{Name: "患者5"})

	item := store.AddItem(&AddItemRequest{
		Name:     "检查项目",
		Category: ItemCategoryLaboratory,
		MinValue: floatPtr(0), MaxValue: floatPtr(100), Price: 50,
	})
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "套餐",
		ItemIDs: []string{item.ID},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  3,
	})

	var wg sync.WaitGroup
	var successCount int
	var mu sync.Mutex

	patientIDs := make([]string, 0)
	for id := range store.patients {
		patientIDs = append(patientIDs, id)
	}

	for _, pid := range patientIDs {
		wg.Add(1)
		go func(patientID string) {
			defer wg.Done()
			_, err := store.CreateAppointment(&CreateAppointmentRequest{
				PatientID:  patientID,
				PackageID:  pkg.ID,
				TimeSlotID: slot.ID,
			})
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(pid)
	}
	wg.Wait()

	if successCount != 3 {
		t.Errorf("expected 3 successful appointments (capacity 3), got %d", successCount)
	}

	updatedSlot, _ := store.GetTimeSlot(slot.ID)
	if updatedSlot.CurrentCount != 3 {
		t.Errorf("expected slot current count 3, got %d", updatedSlot.CurrentCount)
	}
}

func TestRecordResult_BelowMin(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, _ := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "2.5",
		Remarks:       "偏低",
	})
	if !result.IsAbnormal {
		t.Error("expected result to be abnormal (2.5 is below 4.0)")
	}
}

func TestCancelAppointment_Idempotent(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.CancelAppointment(appt.ID)
	cancelled2, err := store.CancelAppointment(appt.ID)
	if err != nil {
		t.Fatalf("Second cancel failed: %v", err)
	}
	if cancelled2.Status != AppointmentStatusCancelled {
		t.Errorf("expected status CANCELLED, got %v", cancelled2.Status)
	}
	updatedSlot, _ := store.GetTimeSlot(slot.ID)
	if updatedSlot.CurrentCount != 0 {
		t.Errorf("expected slot current count 0 (not negative) after double cancel, got %d", updatedSlot.CurrentCount)
	}
}

func TestGenerateReport_Idempotent(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemRBC,
		Value:         "5.0",
	})

	report1, _ := store.GenerateReport(&GenerateReportRequest{AppointmentID: appt.ID})
	report2, _ := store.GenerateReport(&GenerateReportRequest{AppointmentID: appt.ID})
	if report1.ID != report2.ID {
		t.Errorf("expected same report ID on second GenerateReport call, got %s vs %s", report1.ID, report2.ID)
	}
}

func TestListAppointments(t *testing.T) {
	store := setupTestStore()
	patient1 := getFirstPatientID(store)
	patient2 := getSecondPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})

	store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patient1,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patient2,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	appts := store.ListAppointments()
	if len(appts) != 2 {
		t.Errorf("expected 2 appointments, got %d", len(appts))
	}
}

func TestListTimeSlots(t *testing.T) {
	store := NewStore()
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start1 := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end1 := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	start2 := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	end2 := time.Date(2026, 7, 1, 12, 0, 0, 0, time.Local)

	store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date: date, StartTime: start1, EndTime: end1, Capacity: 10,
	})
	store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date: date, StartTime: start2, EndTime: end2, Capacity: 10,
	})

	slots := store.ListTimeSlots()
	if len(slots) != 2 {
		t.Errorf("expected 2 time slots, got %d", len(slots))
	}
}

func TestBoundaryCapacityOne(t *testing.T) {
	store := setupTestStore()
	patient1 := getFirstPatientID(store)
	patient2 := getSecondPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  1,
	})

	_, err := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patient1,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != nil {
		t.Fatalf("First appointment should succeed: %v", err)
	}

	_, err = store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patient2,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	if err != ErrTimeSlotCapacityFull {
		t.Errorf("expected ErrTimeSlotCapacityFull for capacity=1, got %v", err)
	}
}

func TestRecordResult_ReferenceZeroLowerBound_NegativeValue(t *testing.T) {
	store := NewStore()
	patientID := store.AddPatient(&Patient{Name: "测试患者"})

	item := store.AddItem(&AddItemRequest{
		Name:     "血氧饱和度",
		Category: ItemCategoryLaboratory,
		Unit:     "%",
		MinValue: floatPtr(0),
		MaxValue: floatPtr(100),
		Price:    50.0,
	})
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item.ID},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item.ID,
		Value:         "-5",
		Remarks:       "仪器异常",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if !result.IsAbnormal {
		t.Error("expected result to be abnormal (-5 is below 0)")
	}
	if result.Reference != "0.00 - 100.00 %" {
		t.Errorf("expected reference '0.00 - 100.00 %%', got '%s'", result.Reference)
	}
}

func TestRecordResult_ReferenceZeroLowerBound_UpperLimitExceed(t *testing.T) {
	store := NewStore()
	patientID := store.AddPatient(&Patient{Name: "测试患者"})

	item := store.AddItem(&AddItemRequest{
		Name:     "血氧饱和度",
		Category: ItemCategoryLaboratory,
		Unit:     "%",
		MinValue: floatPtr(0),
		MaxValue: floatPtr(100),
		Price:    50.0,
	})
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item.ID},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item.ID,
		Value:         "105",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if !result.IsAbnormal {
		t.Error("expected result to be abnormal (105 is above 100)")
	}
}

func TestRecordResult_ReferenceZeroLowerBound_NormalValue(t *testing.T) {
	store := NewStore()
	patientID := store.AddPatient(&Patient{Name: "测试患者"})

	item := store.AddItem(&AddItemRequest{
		Name:     "血氧饱和度",
		Category: ItemCategoryLaboratory,
		Unit:     "%",
		MinValue: floatPtr(0),
		MaxValue: floatPtr(100),
		Price:    50.0,
	})
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item.ID},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item.ID,
		Value:         "98",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if result.IsAbnormal {
		t.Error("expected result to be normal (98 is within 0-100)")
	}
}

func TestRecordResult_OnlyUpperLimit(t *testing.T) {
	store := NewStore()
	patientID := store.AddPatient(&Patient{Name: "测试患者"})

	item := store.AddItem(&AddItemRequest{
		Name:     "心率",
		Category: ItemCategoryFunctional,
		Unit:     "次/分",
		MinValue: nil,
		MaxValue: floatPtr(100),
		Price:    30.0,
	})
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item.ID},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result1, _ := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item.ID,
		Value:         "120",
	})
	if !result1.IsAbnormal {
		t.Error("expected result to be abnormal (120 is above 100)")
	}
	if result1.Reference != "-∞ - 100.00 次/分" {
		t.Errorf("expected reference '-∞ - 100.00 次/分', got '%s'", result1.Reference)
	}
}

func TestRecordResult_OnlyLowerLimit(t *testing.T) {
	store := NewStore()
	patientID := store.AddPatient(&Patient{Name: "测试患者"})

	item := store.AddItem(&AddItemRequest{
		Name:     "身高",
		Category: ItemCategoryPhysical,
		Unit:     "cm",
		MinValue: floatPtr(100),
		MaxValue: nil,
		Price:    10.0,
	})
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item.ID},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})

	appt1, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	result1, _ := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt1.ID,
		ItemID:        item.ID,
		Value:         "90",
	})
	if !result1.IsAbnormal {
		t.Error("expected result to be abnormal (90 is below 100)")
	}
	if result1.Reference != "100.00 - +∞ cm" {
		t.Errorf("expected reference '100.00 - +∞ cm', got '%s'", result1.Reference)
	}

	appt2, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})
	result2, _ := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt2.ID,
		ItemID:        item.ID,
		Value:         "250",
	})
	if result2.IsAbnormal {
		t.Error("expected result to be normal (250 is above 100, no upper limit)")
	}
}

func TestRecordResult_NoReferenceRange(t *testing.T) {
	store := NewStore()
	patientID := store.AddPatient(&Patient{Name: "测试患者"})

	item := store.AddItem(&AddItemRequest{
		Name:     "X光检查",
		Category: ItemCategoryImaging,
		Unit:     "",
		MinValue: nil,
		MaxValue: nil,
		Price:    100.0,
	})
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "测试套餐",
		ItemIDs: []string{item.ID},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item.ID,
		Value:         "100",
		Remarks:       "心肺正常",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if result.IsAbnormal {
		t.Error("expected result not to be abnormal (no reference range)")
	}
	if result.Reference != "" {
		t.Errorf("expected empty reference, got '%s'", result.Reference)
	}
}

func TestRecordResult_ReportAlreadyGenerated(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemRBC := getItemIDByName(store, "血常规-红细胞计数")
	itemALT := getItemIDByName(store, "肝功能-谷丙转氨酶")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "血液检查套餐",
		ItemIDs: []string{itemWBC, itemRBC, itemALT},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "6.5",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemRBC,
		Value:         "5.0",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemALT,
		Value:         "35",
	})

	_, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	_, err = store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "7.0",
	})
	if err != ErrReportAlreadyGenerated {
		t.Errorf("expected ErrReportAlreadyGenerated, got %v", err)
	}
}

func TestFullWorkflow_ZeroReferenceRange(t *testing.T) {
	store := NewStore()
	patientID := store.AddPatient(&Patient{Name: "测试患者"})

	item1 := store.AddItem(&AddItemRequest{
		Name:     "血氧饱和度",
		Category: ItemCategoryLaboratory,
		Unit:     "%",
		MinValue: floatPtr(0),
		MaxValue: floatPtr(100),
		Price:    50.0,
	})
	item2 := store.AddItem(&AddItemRequest{
		Name:     "心率",
		Category: ItemCategoryFunctional,
		Unit:     "次/分",
		MinValue: nil,
		MaxValue: floatPtr(100),
		Price:    30.0,
	})
	item3 := store.AddItem(&AddItemRequest{
		Name:     "舒张压",
		Category: ItemCategoryPhysical,
		Unit:     "mmHg",
		MinValue: floatPtr(60),
		MaxValue: floatPtr(90),
		Price:    20.0,
	})

	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "综合测试套餐",
		ItemIDs: []string{item1.ID, item2.ID, item3.ID},
	})

	date := time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 15, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 15, 12, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  50,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item1.ID,
		Value:         "-5",
		Remarks:       "偏低，仪器可能异常",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item2.ID,
		Value:         "110",
		Remarks:       "偏快",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        item3.ID,
		Value:         "75",
	})

	report, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}
	if len(report.AbnormalItems) != 2 {
		t.Errorf("expected 2 abnormal items, got %d", len(report.AbnormalItems))
	}

	abnormalNames := make(map[string]bool)
	for _, ai := range report.AbnormalItems {
		abnormalNames[ai.ItemName] = true
	}
	if !abnormalNames["血氧饱和度"] {
		t.Error("expected 血氧饱和度 to be abnormal")
	}
	if !abnormalNames["心率"] {
		t.Error("expected 心率 to be abnormal")
	}
	if abnormalNames["舒张压"] {
		t.Error("expected 舒张压 to be normal")
	}
}

func TestRecordResult_NonNumeric_ReferenceEmpty(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemChest := getItemIDByName(store, "胸部X光")
	itemECG := getItemIDByName(store, "心电图")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "影像检查套餐",
		ItemIDs: []string{itemChest, itemECG},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemChest,
		Value:         "双肺纹理清晰，未见实质性病变",
		Remarks:       "心肺正常，建议每年复查",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if result.IsNumeric {
		t.Error("expected result to be non-numeric")
	}
	if result.NumericValue != 0 {
		t.Errorf("expected numeric value 0 for non-numeric, got %v", result.NumericValue)
	}
	if result.IsAbnormal {
		t.Error("expected non-numeric result not to be auto-flagged abnormal")
	}
	if result.Reference != "" {
		t.Errorf("expected reference to be empty for non-numeric result, got '%s'", result.Reference)
	}
	if result.Remarks != "心肺正常，建议每年复查" {
		t.Errorf("expected remarks preserved, got '%s'", result.Remarks)
	}
}

func TestRecordResult_NonNumeric_ECG_WithRemarks(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemChest := getItemIDByName(store, "胸部X光")
	itemECG := getItemIDByName(store, "心电图")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "影像检查套餐",
		ItemIDs: []string{itemChest, itemECG},
	})
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 1, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  10,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	result, err := store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemECG,
		Value:         "窦性心律不齐",
		Remarks:       "轻度不齐，建议进一步做动态心电图",
	})
	if err != nil {
		t.Fatalf("RecordResult failed: %v", err)
	}
	if result.IsNumeric {
		t.Error("expected ECG result to be non-numeric")
	}
	if result.IsAbnormal {
		t.Error("expected non-numeric ECG not to be auto-flagged abnormal")
	}
	if result.Reference != "" {
		t.Errorf("expected reference to be empty for non-numeric, got '%s'", result.Reference)
	}
	if result.Value != "窦性心律不齐" {
		t.Errorf("expected value preserved, got '%s'", result.Value)
	}
	if result.Remarks != "轻度不齐，建议进一步做动态心电图" {
		t.Errorf("expected remarks preserved, got '%s'", result.Remarks)
	}
}

func TestRecordResult_NonNumericAndNumeric_MixedReport(t *testing.T) {
	store := setupTestStore()
	patientID := getFirstPatientID(store)
	itemWBC := getItemIDByName(store, "血常规-白细胞计数")
	itemChest := getItemIDByName(store, "胸部X光")
	pkg, _ := store.CreatePackage(&CreatePackageRequest{
		Name:    "基础检查套餐",
		ItemIDs: []string{itemWBC, itemChest},
	})
	date := time.Date(2026, 7, 15, 0, 0, 0, 0, time.Local)
	start := time.Date(2026, 7, 15, 8, 0, 0, 0, time.Local)
	end := time.Date(2026, 7, 15, 12, 0, 0, 0, time.Local)
	slot, _ := store.CreateTimeSlot(&CreateTimeSlotRequest{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Capacity:  50,
	})
	appt, _ := store.CreateAppointment(&CreateAppointmentRequest{
		PatientID:  patientID,
		PackageID:  pkg.ID,
		TimeSlotID: slot.ID,
	})

	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemWBC,
		Value:         "15.0",
		Remarks:       "偏高",
	})
	store.RecordResult(&RecordResultRequest{
		AppointmentID: appt.ID,
		ItemID:        itemChest,
		Value:         "未见异常",
		Remarks:       "心肺正常",
	})

	report, err := store.GenerateReport(&GenerateReportRequest{
		AppointmentID: appt.ID,
	})
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}
	if len(report.Results) != 2 {
		t.Errorf("expected 2 results in report, got %d", len(report.Results))
	}
	if len(report.AbnormalItems) != 1 {
		t.Errorf("expected 1 abnormal item (WBC numeric), got %d", len(report.AbnormalItems))
	}

	chestResult := report.Results[itemChest]
	if chestResult == nil {
		t.Fatal("chest x-ray result should exist in report")
	}
	if chestResult.IsNumeric {
		t.Error("chest x-ray result should be non-numeric")
	}
	if chestResult.Reference != "" {
		t.Errorf("expected chest x-ray reference empty, got '%s'", chestResult.Reference)
	}
	if chestResult.Remarks != "心肺正常" {
		t.Errorf("expected chest x-ray remarks preserved, got '%s'", chestResult.Remarks)
	}
}
