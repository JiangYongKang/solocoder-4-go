package appointment

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type TestContext struct {
	Store     *Store
	Service   *Service
	Doctor1ID string
	Doctor2ID string
	Patient1ID string
	Patient2ID string
	Patient3ID string
	Slot1ID    string
	Slot2ID    string
	Slot3ID    string
	Schedule1ID string
	FutureDate string
}

func setupTestStore() *TestContext {
	store := NewStore()
	svc := NewService(store)

	ctx := &TestContext{
		Store:   store,
		Service: svc,
	}

	ctx.Doctor1ID = store.AddDoctor(&Doctor{
		Name:       "张医生",
		Title:      "主任医师",
		Department: "内科",
		LicenseNo:  "DOC001",
	})

	ctx.Doctor2ID = store.AddDoctor(&Doctor{
		Name:       "李医生",
		Title:      "副主任医师",
		Department: "外科",
		LicenseNo:  "DOC002",
	})

	ctx.Patient1ID = store.AddPatient(&Patient{
		Name:            "王小明",
		IDCard:          "110101199001011234",
		Phone:           "13800138000",
		Gender:          "男",
		Age:             35,
		MedicalRecordNo: "MR001",
	})

	ctx.Patient2ID = store.AddPatient(&Patient{
		Name:            "赵小红",
		IDCard:          "110101199203151234",
		Phone:           "13900139000",
		Gender:          "女",
		Age:             32,
		MedicalRecordNo: "MR002",
	})

	ctx.Patient3ID = store.AddPatient(&Patient{
		Name:            "孙大华",
		IDCard:          "110101198512201234",
		Phone:           "13700137000",
		Gender:          "男",
		Age:             39,
		MedicalRecordNo: "MR003",
	})

	ctx.FutureDate = time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")

	schedule1, _ := svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor1ID,
		Date:     ctx.FutureDate,
		Slots: []ScheduleSlotSpec{
			{StartTime: "08:00", EndTime: "09:00", TotalCapacity: 5},
			{StartTime: "09:00", EndTime: "10:00", TotalCapacity: 3},
			{StartTime: "10:00", EndTime: "11:00", TotalCapacity: 1},
		},
	})
	ctx.Schedule1ID = schedule1.ID
	ctx.Slot1ID = schedule1.Slots[0].ID
	ctx.Slot2ID = schedule1.Slots[1].ID
	ctx.Slot3ID = schedule1.Slots[2].ID

	return ctx
}

func setupTestStoreWithMockTime() (*TestContext, *time.Time) {
	baseTime := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	store := NewStore()
	store.NowFunc = func() time.Time { return baseTime }

	svc := NewService(store)
	ctx := &TestContext{
		Store:   store,
		Service: svc,
	}

	ctx.Doctor1ID = store.AddDoctor(&Doctor{
		Name:       "张医生",
		Title:      "主任医师",
		Department: "内科",
		LicenseNo:  "DOC001",
	})
	ctx.Doctor2ID = store.AddDoctor(&Doctor{
		Name:       "李医生",
		Title:      "副主任医师",
		Department: "外科",
		LicenseNo:  "DOC002",
	})

	ctx.Patient1ID = store.AddPatient(&Patient{
		Name:            "王小明",
		IDCard:          "110101199001011234",
		Phone:           "13800138000",
		Gender:          "男",
		Age:             35,
		MedicalRecordNo: "MR001",
	})
	ctx.Patient2ID = store.AddPatient(&Patient{
		Name:            "赵小红",
		IDCard:          "110101199203151234",
		Phone:           "13900139000",
		Gender:          "女",
		Age:             32,
		MedicalRecordNo: "MR002",
	})
	ctx.Patient3ID = store.AddPatient(&Patient{
		Name:            "孙大华",
		IDCard:          "110101198512201234",
		Phone:           "13700137000",
		Gender:          "男",
		Age:             39,
		MedicalRecordNo: "MR003",
	})

	ctx.FutureDate = "2025-06-22"

	schedule1, _ := svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor1ID,
		Date:     ctx.FutureDate,
		Slots: []ScheduleSlotSpec{
			{StartTime: "08:00", EndTime: "09:00", TotalCapacity: 5},
			{StartTime: "09:00", EndTime: "10:00", TotalCapacity: 3},
			{StartTime: "10:00", EndTime: "11:00", TotalCapacity: 1},
		},
	})
	ctx.Schedule1ID = schedule1.ID
	ctx.Slot1ID = schedule1.Slots[0].ID
	ctx.Slot2ID = schedule1.Slots[1].ID
	ctx.Slot3ID = schedule1.Slots[2].ID

	return ctx, &baseTime
}

func TestAddSchedule_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	schedule, err := svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor2ID,
		Date:     "2025-06-23",
		Slots: []ScheduleSlotSpec{
			{StartTime: "14:00", EndTime: "15:00", TotalCapacity: 10},
		},
	})

	if err != nil {
		t.Fatalf("AddSchedule failed: %v", err)
	}
	if schedule == nil {
		t.Fatal("schedule should not be nil")
	}
	if schedule.ID == "" {
		t.Error("schedule ID should not be empty")
	}
	if schedule.DoctorID != ctx.Doctor2ID {
		t.Errorf("expected doctor ID %s, got %s", ctx.Doctor2ID, schedule.DoctorID)
	}
	if schedule.Date != "2025-06-23" {
		t.Errorf("expected date 2025-06-23, got %s", schedule.Date)
	}
	if len(schedule.Slots) != 1 {
		t.Errorf("expected 1 slot, got %d", len(schedule.Slots))
	}
	if schedule.Slots[0].TotalCapacity != 10 {
		t.Errorf("expected capacity 10, got %d", schedule.Slots[0].TotalCapacity)
	}
	if schedule.Slots[0].Status != SlotStatusAvailable {
		t.Errorf("expected status AVAILABLE, got %s", schedule.Slots[0].Status)
	}
}

func TestAddSchedule_InvalidDoctor(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	_, err := svc.AddSchedule(&AddScheduleRequest{
		DoctorID: "INVALID_DOC",
		Date:     "2025-06-22",
		Slots:    []ScheduleSlotSpec{{StartTime: "08:00", EndTime: "09:00", TotalCapacity: 5}},
	})

	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestAddSchedule_InvalidDate(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	_, err := svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor1ID,
		Date:     "invalid-date",
		Slots:    []ScheduleSlotSpec{{StartTime: "08:00", EndTime: "09:00", TotalCapacity: 5}},
	})

	if err != ErrInvalidDate {
		t.Errorf("expected ErrInvalidDate, got %v", err)
	}
}

func TestAddSchedule_SkipInvalidCapacity(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	schedule, err := svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor2ID,
		Date:     "2025-06-23",
		Slots: []ScheduleSlotSpec{
			{StartTime: "08:00", EndTime: "09:00", TotalCapacity: 0},
			{StartTime: "09:00", EndTime: "10:00", TotalCapacity: -1},
			{StartTime: "10:00", EndTime: "11:00", TotalCapacity: 5},
		},
	})

	if err != nil {
		t.Fatalf("AddSchedule failed: %v", err)
	}
	if len(schedule.Slots) != 1 {
		t.Errorf("expected 1 valid slot, got %d", len(schedule.Slots))
	}
}

func TestQuerySchedules_All(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor2ID,
		Date:     "2025-06-22",
		Slots: []ScheduleSlotSpec{
			{StartTime: "14:00", EndTime: "15:00", TotalCapacity: 8},
		},
	})

	results, err := svc.QuerySchedules(&QueryScheduleRequest{})
	if err != nil {
		t.Fatalf("QuerySchedules failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(results))
	}
}

func TestQuerySchedules_ByDoctor(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor2ID,
		Date:     "2025-06-22",
		Slots: []ScheduleSlotSpec{
			{StartTime: "14:00", EndTime: "15:00", TotalCapacity: 8},
		},
	})

	results, err := svc.QuerySchedules(&QueryScheduleRequest{DoctorID: ctx.Doctor1ID})
	if err != nil {
		t.Fatalf("QuerySchedules failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 schedule for doctor1, got %d", len(results))
	}
	if results[0].DoctorID != ctx.Doctor1ID {
		t.Errorf("expected doctor ID %s, got %s", ctx.Doctor1ID, results[0].DoctorID)
	}
}

func TestQuerySchedules_ByDepartment(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor2ID,
		Date:     "2025-06-22",
		Slots: []ScheduleSlotSpec{
			{StartTime: "14:00", EndTime: "15:00", TotalCapacity: 8},
		},
	})

	results, err := svc.QuerySchedules(&QueryScheduleRequest{Department: "外科"})
	if err != nil {
		t.Fatalf("QuerySchedules failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 schedule for 外科, got %d", len(results))
	}
	if results[0].Department != "外科" {
		t.Errorf("expected department 外科, got %s", results[0].Department)
	}
}

func TestQuerySchedules_ByDate(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor1ID,
		Date:     "2025-06-23",
		Slots: []ScheduleSlotSpec{
			{StartTime: "08:00", EndTime: "09:00", TotalCapacity: 5},
		},
	})

	results, err := svc.QuerySchedules(&QueryScheduleRequest{Date: "2025-06-22"})
	if err != nil {
		t.Fatalf("QuerySchedules failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 schedule for date 2025-06-22, got %d", len(results))
	}
}

func TestQuerySchedules_CombinedFilters(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	svc.AddSchedule(&AddScheduleRequest{
		DoctorID: ctx.Doctor1ID,
		Date:     "2025-06-23",
		Slots: []ScheduleSlotSpec{
			{StartTime: "08:00", EndTime: "09:00", TotalCapacity: 5},
		},
	})

	results, err := svc.QuerySchedules(&QueryScheduleRequest{
		DoctorID:   ctx.Doctor1ID,
		Department: "内科",
		Date:       "2025-06-22",
	})
	if err != nil {
		t.Fatalf("QuerySchedules failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 schedule with combined filters, got %d", len(results))
	}
}

func TestQuerySchedules_InvalidDate(t *testing.T) {
	ctx := setupTestStore()
	svc := ctx.Service

	_, err := svc.QuerySchedules(&QueryScheduleRequest{Date: "bad-date"})
	if err != ErrInvalidDate {
		t.Errorf("expected ErrInvalidDate, got %v", err)
	}
}

func TestQuerySchedules_SlotRemainingAndFull(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	for i := 0; i < 5; i++ {
		patientID := ctx.Store.AddPatient(&Patient{Name: "Test"})
		_, _ = svc.CreateAppointment(&CreateAppointmentRequest{
			SlotID:    ctx.Slot1ID,
			PatientID: patientID,
		})
	}

	results, err := svc.QuerySchedules(&QueryScheduleRequest{DoctorID: ctx.Doctor1ID})
	if err != nil {
		t.Fatalf("QuerySchedules failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(results))
	}

	slots := results[0].Slots
	var slot1Result *SlotQueryResult
	for i := range slots {
		if slots[i].SlotID == ctx.Slot1ID {
			slot1Result = &slots[i]
			break
		}
	}

	if slot1Result == nil {
		t.Fatal("slot1 not found in results")
	}

	if slot1Result.RemainingCount != 0 {
		t.Errorf("expected remaining 0 for slot1, got %d", slot1Result.RemainingCount)
	}
	if !slot1Result.IsFull {
		t.Error("expected slot1 to be full")
	}
	if slot1Result.Status != SlotStatusBooked {
		t.Errorf("expected slot1 status BOOKED, got %s", slot1Result.Status)
	}
}

func TestLockSlot_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	slot, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	if err != nil {
		t.Fatalf("LockSlot failed: %v", err)
	}
	if slot.Status != SlotStatusLocked {
		t.Errorf("expected status LOCKED, got %s", slot.Status)
	}
	if slot.LockedBy != ctx.Patient1ID {
		t.Errorf("expected locked by %s, got %s", ctx.Patient1ID, slot.LockedBy)
	}
	if slot.LockedAt == nil {
		t.Error("LockedAt should not be nil")
	}
	if slot.LockExpireAt == nil {
		t.Error("LockExpireAt should not be nil")
	}
}

func TestLockSlot_InvalidPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: "INVALID_PATIENT",
	})

	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestLockSlot_InvalidSlot(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    "INVALID_SLOT",
		PatientID: ctx.Patient1ID,
	})

	if err != ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestLockSlot_AlreadyLocked(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("First LockSlot failed: %v", err)
	}

	_, err = svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient2ID,
	})

	if err != ErrSlotAlreadyLocked {
		t.Errorf("expected ErrSlotAlreadyLocked, got %v", err)
	}
}

func TestLockSlot_FullSlot(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	for i := 0; i < 5; i++ {
		patientID := ctx.Store.AddPatient(&Patient{Name: "Test"})
		_, _ = svc.CreateAppointment(&CreateAppointmentRequest{
			SlotID:    ctx.Slot1ID,
			PatientID: patientID,
		})
	}

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient3ID,
	})

	if err != ErrSlotNoCapacity {
		t.Errorf("expected ErrSlotNoCapacity, got %v", err)
	}
}

func TestLockSlot_ExpiredLockReuse(t *testing.T) {
	ctx, baseTime := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("First LockSlot failed: %v", err)
	}

	*baseTime = baseTime.Add(10 * time.Minute)

	slot, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient2ID,
	})

	if err != nil {
		t.Fatalf("LockSlot after expiry failed: %v", err)
	}
	if slot.LockedBy != ctx.Patient2ID {
		t.Errorf("expected locked by %s, got %s", ctx.Patient2ID, slot.LockedBy)
	}
}

func TestConfirmAppointment_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("LockSlot failed: %v", err)
	}

	appt, err := svc.ConfirmAppointment(&ConfirmAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	if err != nil {
		t.Fatalf("ConfirmAppointment failed: %v", err)
	}
	if appt.ID == "" {
		t.Error("appointment ID should not be empty")
	}
	if appt.Status != AppointmentStatusConfirmed {
		t.Errorf("expected status CONFIRMED, got %s", appt.Status)
	}
	if appt.PatientID != ctx.Patient1ID {
		t.Errorf("expected patient %s, got %s", ctx.Patient1ID, appt.PatientID)
	}
	if appt.SlotID != ctx.Slot1ID {
		t.Errorf("expected slot %s, got %s", ctx.Slot1ID, appt.SlotID)
	}
	if appt.ConfirmedAt == nil {
		t.Error("ConfirmedAt should not be nil")
	}

	ctx.Store.mu.RLock()
	slot := ctx.Store.slotMap[ctx.Slot1ID]
	ctx.Store.mu.RUnlock()
	if slot.BookedCount != 1 {
		t.Errorf("expected booked count 1, got %d", slot.BookedCount)
	}
	if slot.Status != SlotStatusAvailable {
		t.Errorf("expected slot status AVAILABLE, got %s", slot.Status)
	}
}

func TestConfirmAppointment_InvalidPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("LockSlot failed: %v", err)
	}

	_, err = svc.ConfirmAppointment(&ConfirmAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: "INVALID_PATIENT",
	})

	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestConfirmAppointment_InvalidSlot(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.ConfirmAppointment(&ConfirmAppointmentRequest{
		SlotID:    "INVALID_SLOT",
		PatientID: ctx.Patient1ID,
	})

	if err != ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestConfirmAppointment_NotLocked(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.ConfirmAppointment(&ConfirmAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	if err != ErrSlotNotLocked {
		t.Errorf("expected ErrSlotNotLocked, got %v", err)
	}
}

func TestConfirmAppointment_WrongPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("LockSlot failed: %v", err)
	}

	_, err = svc.ConfirmAppointment(&ConfirmAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient2ID,
	})

	if err != ErrInvalidLockOwner {
		t.Errorf("expected ErrInvalidLockOwner, got %v", err)
	}
}

func TestConfirmAppointment_ExpiredLock(t *testing.T) {
	ctx, baseTime := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("LockSlot failed: %v", err)
	}

	*baseTime = baseTime.Add(10 * time.Minute)

	_, err = svc.ConfirmAppointment(&ConfirmAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	if err != ErrSlotLockExpired {
		t.Errorf("expected ErrSlotLockExpired, got %v", err)
	}
}

func TestReleaseLock_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("LockSlot failed: %v", err)
	}

	err = svc.ReleaseLock(&ReleaseLockRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	if err != nil {
		t.Fatalf("ReleaseLock failed: %v", err)
	}

	ctx.Store.mu.RLock()
	slot := ctx.Store.slotMap[ctx.Slot1ID]
	ctx.Store.mu.RUnlock()
	if slot.Status != SlotStatusAvailable {
		t.Errorf("expected slot status AVAILABLE, got %s", slot.Status)
	}
	if slot.LockedBy != "" {
		t.Error("LockedBy should be cleared")
	}
}

func TestReleaseLock_NotLocked(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	err := svc.ReleaseLock(&ReleaseLockRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	if err != ErrSlotNotLocked {
		t.Errorf("expected ErrSlotNotLocked, got %v", err)
	}
}

func TestReleaseLock_WrongPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("LockSlot failed: %v", err)
	}

	err = svc.ReleaseLock(&ReleaseLockRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient2ID,
	})

	if err != ErrInvalidLockOwner {
		t.Errorf("expected ErrInvalidLockOwner, got %v", err)
	}
}

func TestReleaseLock_InvalidSlot(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	err := svc.ReleaseLock(&ReleaseLockRequest{
		SlotID:    "INVALID_SLOT",
		PatientID: ctx.Patient1ID,
	})

	if err != ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestCheckAndReleaseExpiredLocks(t *testing.T) {
	ctx, baseTime := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("LockSlot slot1 failed: %v", err)
	}

	_, err = svc.LockSlot(&LockSlotRequest{
		SlotID:    ctx.Slot2ID,
		PatientID: ctx.Patient2ID,
	})
	if err != nil {
		t.Fatalf("LockSlot slot2 failed: %v", err)
	}

	released := svc.CheckAndReleaseExpiredLocks()
	if released != 0 {
		t.Errorf("expected 0 released, got %d", released)
	}

	*baseTime = baseTime.Add(10 * time.Minute)

	released = svc.CheckAndReleaseExpiredLocks()
	if released != 2 {
		t.Errorf("expected 2 released, got %d", released)
	}

	ctx.Store.mu.RLock()
	slot1 := ctx.Store.slotMap[ctx.Slot1ID]
	slot2 := ctx.Store.slotMap[ctx.Slot2ID]
	ctx.Store.mu.RUnlock()

	if slot1.Status != SlotStatusAvailable {
		t.Errorf("slot1 expected AVAILABLE, got %s", slot1.Status)
	}
	if slot2.Status != SlotStatusAvailable {
		t.Errorf("slot2 expected AVAILABLE, got %s", slot2.Status)
	}
}

func TestCreateAppointment_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}
	if appt.Status != AppointmentStatusConfirmed {
		t.Errorf("expected CONFIRMED, got %s", appt.Status)
	}

	ctx.Store.mu.RLock()
	slot := ctx.Store.slotMap[ctx.Slot1ID]
	ctx.Store.mu.RUnlock()
	if slot.BookedCount != 1 {
		t.Errorf("expected booked count 1, got %d", slot.BookedCount)
	}
}

func TestCreateAppointment_SlotFull(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	for i := 0; i < 5; i++ {
		patientID := ctx.Store.AddPatient(&Patient{Name: "Test"})
		_, err := svc.CreateAppointment(&CreateAppointmentRequest{
			SlotID:    ctx.Slot1ID,
			PatientID: patientID,
		})
		if err != nil {
			t.Fatalf("CreateAppointment %d failed: %v", i, err)
		}
	}

	_, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient3ID,
	})

	if err != ErrSlotNoCapacity {
		t.Errorf("expected ErrSlotNoCapacity, got %v", err)
	}
}

func TestChangeAppointment_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt1, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	newAppt, err := svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt1.ID,
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient1ID,
	})

	if err != nil {
		t.Fatalf("ChangeAppointment failed: %v", err)
	}
	if newAppt.SlotID != ctx.Slot2ID {
		t.Errorf("expected new slot %s, got %s", ctx.Slot2ID, newAppt.SlotID)
	}
	if newAppt.ChangedFromID != appt1.ID {
		t.Errorf("expected ChangedFromID %s, got %s", appt1.ID, newAppt.ChangedFromID)
	}

	oldAppt, _ := svc.GetAppointment(appt1.ID)
	if oldAppt.Status != AppointmentStatusChanged {
		t.Errorf("old appt expected CHANGED, got %s", oldAppt.Status)
	}
	if oldAppt.ChangedToID != newAppt.ID {
		t.Errorf("old appt ChangedToID expected %s, got %s", newAppt.ID, oldAppt.ChangedToID)
	}

	ctx.Store.mu.RLock()
	oldSlot := ctx.Store.slotMap[ctx.Slot1ID]
	newSlot := ctx.Store.slotMap[ctx.Slot2ID]
	ctx.Store.mu.RUnlock()
	if oldSlot.BookedCount != 0 {
		t.Errorf("old slot expected booked 0, got %d", oldSlot.BookedCount)
	}
	if newSlot.BookedCount != 1 {
		t.Errorf("new slot expected booked 1, got %d", newSlot.BookedCount)
	}
}

func TestChangeAppointment_SameSlot(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	_, err = svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt.ID,
		NewSlotID:     ctx.Slot1ID,
		PatientID:     ctx.Patient1ID,
	})

	if err != ErrSameSlotChange {
		t.Errorf("expected ErrSameSlotChange, got %v", err)
	}
}

func TestChangeAppointment_InvalidAppointment(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: "INVALID_APPT",
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient1ID,
	})

	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestChangeAppointment_WrongPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	_, err = svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt.ID,
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient2ID,
	})

	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestChangeAppointment_NotConfirmedStatus(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	_, _ = svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     ctx.Patient1ID,
		Reason:        "测试",
	})

	_, err = svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt.ID,
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient1ID,
	})

	if err != ErrChangeNotAllowed {
		t.Errorf("expected ErrChangeNotAllowed, got %v", err)
	}
}

func TestChangeAppointment_TooLate(t *testing.T) {
	ctx, baseTime := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	*baseTime = time.Date(2025, 6, 22, 7, 30, 0, 0, time.UTC)

	_, err = svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt.ID,
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient1ID,
	})

	if err != ErrCannotChangePast {
		t.Errorf("expected ErrCannotChangePast, got %v", err)
	}
}

func TestChangeAppointment_NewSlotFull(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		patientID := ctx.Store.AddPatient(&Patient{Name: "Test"})
		_, _ = svc.CreateAppointment(&CreateAppointmentRequest{
			SlotID:    ctx.Slot2ID,
			PatientID: patientID,
		})
	}

	_, err = svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt.ID,
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient1ID,
	})

	if err != ErrSlotNoCapacity {
		t.Errorf("expected ErrSlotNoCapacity, got %v", err)
	}
}

func TestCancelAppointment_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	cancelled, err := svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     ctx.Patient1ID,
		Reason:        "临时有事",
	})

	if err != nil {
		t.Fatalf("CancelAppointment failed: %v", err)
	}
	if cancelled.Status != AppointmentStatusCancelled {
		t.Errorf("expected CANCELLED, got %s", cancelled.Status)
	}
	if cancelled.CancelledAt == nil {
		t.Error("CancelledAt should not be nil")
	}

	ctx.Store.mu.RLock()
	slot := ctx.Store.slotMap[ctx.Slot1ID]
	ctx.Store.mu.RUnlock()
	if slot.BookedCount != 0 {
		t.Errorf("slot expected booked 0, got %d", slot.BookedCount)
	}
	if slot.Status != SlotStatusAvailable {
		t.Errorf("slot expected AVAILABLE, got %s", slot.Status)
	}
}

func TestCancelAppointment_InvalidAppointment(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: "INVALID_APPT",
		PatientID:     ctx.Patient1ID,
		Reason:        "测试",
	})

	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestCancelAppointment_WrongPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	_, err = svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     ctx.Patient2ID,
		Reason:        "测试",
	})

	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestCancelAppointment_AlreadyCancelled(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	_, _ = svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     ctx.Patient1ID,
		Reason:        "测试",
	})

	_, err = svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     ctx.Patient1ID,
		Reason:        "再次取消",
	})

	if err != ErrChangeNotAllowed {
		t.Errorf("expected ErrChangeNotAllowed, got %v", err)
	}
}

func TestCancelAppointment_TooLate(t *testing.T) {
	ctx, baseTime := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	*baseTime = time.Date(2025, 6, 21, 21, 0, 0, 0, time.UTC)

	_, err = svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     ctx.Patient1ID,
		Reason:        "临时有事",
	})

	if err != ErrCannotCancelPast {
		t.Errorf("expected ErrCannotCancelPast, got %v", err)
	}
}

func TestRecordNoShow_Success(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	record, err := svc.RecordNoShow(&RecordNoShowRequest{
		AppointmentID: appt.ID,
		Remark:        "患者未按时就诊且未提前告知",
	})

	if err != nil {
		t.Fatalf("RecordNoShow failed: %v", err)
	}
	if record == nil {
		t.Fatal("record should not be nil")
	}
	if record.PatientID != ctx.Patient1ID {
		t.Errorf("expected patient %s, got %s", ctx.Patient1ID, record.PatientID)
	}
	if record.AppointmentID != appt.ID {
		t.Errorf("expected appointment %s, got %s", appt.ID, record.AppointmentID)
	}
	if record.Remark != "患者未按时就诊且未提前告知" {
		t.Errorf("expected remark, got %s", record.Remark)
	}

	apptAfter, _ := svc.GetAppointment(appt.ID)
	if apptAfter.Status != AppointmentStatusNoShow {
		t.Errorf("appt expected NO_SHOW, got %s", apptAfter.Status)
	}
	if !apptAfter.IsNoShow {
		t.Error("IsNoShow should be true")
	}

	ctx.Store.mu.RLock()
	slot := ctx.Store.slotMap[ctx.Slot1ID]
	ctx.Store.mu.RUnlock()
	if slot.BookedCount != 0 {
		t.Errorf("slot expected booked 0, got %d", slot.BookedCount)
	}
}

func TestRecordNoShow_NoRemark(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	_, err = svc.RecordNoShow(&RecordNoShowRequest{
		AppointmentID: appt.ID,
		Remark:        "",
	})

	if err != ErrNoShowRemarkRequired {
		t.Errorf("expected ErrNoShowRemarkRequired, got %v", err)
	}
}

func TestRecordNoShow_InvalidAppointment(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.RecordNoShow(&RecordNoShowRequest{
		AppointmentID: "INVALID_APPT",
		Remark:        "测试",
	})

	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestRecordNoShow_NotConfirmed(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("CreateAppointment failed: %v", err)
	}

	_, _ = svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     ctx.Patient1ID,
		Reason:        "测试",
	})

	_, err = svc.RecordNoShow(&RecordNoShowRequest{
		AppointmentID: appt.ID,
		Remark:        "测试",
	})

	if err != ErrAppointmentNotActive {
		t.Errorf("expected ErrAppointmentNotActive, got %v", err)
	}
}

func TestGetNoShowRecords(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt1, _ := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	appt2, _ := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot2ID,
		PatientID: ctx.Patient2ID,
	})
	appt3, _ := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot3ID,
		PatientID: ctx.Patient1ID,
	})

	svc.RecordNoShow(&RecordNoShowRequest{AppointmentID: appt1.ID, Remark: "爽约1"})
	svc.RecordNoShow(&RecordNoShowRequest{AppointmentID: appt2.ID, Remark: "爽约2"})
	svc.RecordNoShow(&RecordNoShowRequest{AppointmentID: appt3.ID, Remark: "爽约3"})

	allRecords, err := svc.GetNoShowRecords("")
	if err != nil {
		t.Fatalf("GetNoShowRecords(all) failed: %v", err)
	}
	if len(allRecords) != 3 {
		t.Errorf("expected 3 all records, got %d", len(allRecords))
	}

	patient1Records, err := svc.GetNoShowRecords(ctx.Patient1ID)
	if err != nil {
		t.Fatalf("GetNoShowRecords(patient1) failed: %v", err)
	}
	if len(patient1Records) != 2 {
		t.Errorf("expected 2 records for patient1, got %d", len(patient1Records))
	}
}

func TestGetNoShowRecords_InvalidPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.GetNoShowRecords("INVALID_PATIENT")
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestJoinWaitQueue_ByDoctor(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	item, err := svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient1ID,
		DoctorID:  ctx.Doctor1ID,
		Date:      ctx.FutureDate,
	})

	if err != nil {
		t.Fatalf("JoinWaitQueue failed: %v", err)
	}
	if item == nil {
		t.Fatal("item should not be nil")
	}
	if item.PatientID != ctx.Patient1ID {
		t.Errorf("expected patient %s, got %s", ctx.Patient1ID, item.PatientID)
	}
	if item.DoctorID != ctx.Doctor1ID {
		t.Errorf("expected doctor %s, got %s", ctx.Doctor1ID, item.DoctorID)
	}
}

func TestJoinWaitQueue_ByDepartment(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	item, err := svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID:  ctx.Patient1ID,
		Department: "内科",
		Date:       ctx.FutureDate,
	})

	if err != nil {
		t.Fatalf("JoinWaitQueue failed: %v", err)
	}
	if item.Department != "内科" {
		t.Errorf("expected department 内科, got %s", item.Department)
	}
}

func TestJoinWaitQueue_InvalidPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: "INVALID_PATIENT",
		DoctorID:  ctx.Doctor1ID,
		Date:      ctx.FutureDate,
	})

	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestJoinWaitQueue_InvalidDoctor(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient1ID,
		DoctorID:  "INVALID_DOC",
		Date:      ctx.FutureDate,
	})

	if err != ErrDoctorNotFound {
		t.Errorf("expected ErrDoctorNotFound, got %v", err)
	}
}

func TestJoinWaitQueue_InvalidDate(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient1ID,
		DoctorID:  ctx.Doctor1ID,
		Date:      "bad-date",
	})

	if err != ErrInvalidDate {
		t.Errorf("expected ErrInvalidDate, got %v", err)
	}
}

func TestGetWaitQueue(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, _ = svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient1ID,
		DoctorID:  ctx.Doctor1ID,
		Date:      ctx.FutureDate,
	})
	_, _ = svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient2ID,
		DoctorID:  ctx.Doctor1ID,
		Date:      ctx.FutureDate,
	})

	queue, err := svc.GetWaitQueue(ctx.Doctor1ID, "", ctx.FutureDate)
	if err != nil {
		t.Fatalf("GetWaitQueue failed: %v", err)
	}
	if len(queue) != 2 {
		t.Errorf("expected 2 items in queue, got %d", len(queue))
	}
	if queue[0].PatientID != ctx.Patient1ID {
		t.Errorf("expected first in queue %s, got %s", ctx.Patient1ID, queue[0].PatientID)
	}
}

func TestGetWaitQueue_InvalidDate(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, err := svc.GetWaitQueue(ctx.Doctor1ID, "", "bad-date")
	if err != ErrInvalidDate {
		t.Errorf("expected ErrInvalidDate, got %v", err)
	}
}

func TestAutoFillFromWaitQueue_OnCancel(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	patientBooked := ctx.Store.AddPatient(&Patient{Name: "已预约"})
	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot3ID,
		PatientID: patientBooked,
	})
	if err != nil {
		t.Fatalf("CreateAppointment for slot3 failed: %v", err)
	}

	ctx.Store.mu.RLock()
	slotBefore := ctx.Store.slotMap[ctx.Slot3ID]
	if slotBefore.BookedCount != 1 {
		ctx.Store.mu.RUnlock()
		t.Fatalf("slot3 expected booked 1, got %d", slotBefore.BookedCount)
	}
	if slotBefore.Status != SlotStatusBooked {
		ctx.Store.mu.RUnlock()
		t.Fatalf("slot3 expected BOOKED, got %s", slotBefore.Status)
	}
	ctx.Store.mu.RUnlock()

	_, _ = svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient1ID,
		DoctorID:  ctx.Doctor1ID,
		Date:      ctx.FutureDate,
	})

	_, _ = svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: appt.ID,
		PatientID:     patientBooked,
		Reason:        "测试取消释放候补",
	})

	ctx.Store.mu.RLock()
	appointments := ctx.Store.appointments
	patient1Appts := 0
	for _, a := range appointments {
		if a.PatientID == ctx.Patient1ID && a.SlotID == ctx.Slot3ID && a.Status == AppointmentStatusConfirmed {
			patient1Appts++
		}
	}
	ctx.Store.mu.RUnlock()

	if patient1Appts != 1 {
		t.Fatalf("expected 1 auto-filled appointment for patient1, got %d", patient1Appts)
	}

	ctx.Store.mu.RLock()
	slotAfter := ctx.Store.slotMap[ctx.Slot3ID]
	ctx.Store.mu.RUnlock()
	if slotAfter.BookedCount != 1 {
		t.Errorf("slot3 expected booked 1 after auto-fill, got %d", slotAfter.BookedCount)
	}
	if slotAfter.Status != SlotStatusBooked {
		t.Errorf("slot3 expected BOOKED after auto-fill, got %s", slotAfter.Status)
	}
}

func TestAutoFillFromWaitQueue_OnNoShow(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	patientBooked := ctx.Store.AddPatient(&Patient{Name: "已预约"})
	appt, _ := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot3ID,
		PatientID: patientBooked,
	})

	_, _ = svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID:  ctx.Patient1ID,
		Department: "内科",
		Date:       ctx.FutureDate,
	})

	_, err := svc.RecordNoShow(&RecordNoShowRequest{
		AppointmentID: appt.ID,
		Remark:        "未按时就诊",
	})
	if err != nil {
		t.Fatalf("RecordNoShow failed: %v", err)
	}

	ctx.Store.mu.RLock()
	appointments := ctx.Store.appointments
	patient1Appts := 0
	for _, a := range appointments {
		if a.PatientID == ctx.Patient1ID && a.SlotID == ctx.Slot3ID && a.Status == AppointmentStatusConfirmed {
			patient1Appts++
		}
	}
	ctx.Store.mu.RUnlock()

	if patient1Appts != 1 {
		t.Errorf("expected 1 auto-filled appointment for patient1, got %d", patient1Appts)
	}
}

func TestAutoFillFromWaitQueue_MultipleCandidates(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	ctx.Store.mu.RLock()
	slot3Before := ctx.Store.slotMap[ctx.Slot3ID]
	capacity := slot3Before.TotalCapacity
	ctx.Store.mu.RUnlock()

	bookedPatients := make([]string, capacity)
	apptIDs := make([]string, capacity)
	for i := 0; i < capacity; i++ {
		pid := ctx.Store.AddPatient(&Patient{Name: "Booked"})
		bookedPatients[i] = pid
		appt, _ := svc.CreateAppointment(&CreateAppointmentRequest{
			SlotID:    ctx.Slot3ID,
			PatientID: pid,
		})
		apptIDs[i] = appt.ID
	}

	waitPatients := make([]string, capacity+2)
	for i := 0; i < capacity+2; i++ {
		pid := ctx.Store.AddPatient(&Patient{Name: "Wait"})
		waitPatients[i] = pid
		_, _ = svc.JoinWaitQueue(&JoinWaitQueueRequest{
			PatientID: pid,
			DoctorID:  ctx.Doctor1ID,
			Date:      ctx.FutureDate,
		})
	}

	for i := 0; i < capacity; i++ {
		_, _ = svc.CancelAppointment(&CancelAppointmentRequest{
			AppointmentID: apptIDs[i],
			PatientID:     bookedPatients[i],
			Reason:        "批量取消",
		})
	}

	ctx.Store.mu.RLock()
	appointments := ctx.Store.appointments
	waitPatientBooked := make(map[string]bool)
	for _, appt := range appointments {
		if appt.SlotID == ctx.Slot3ID && appt.Status == AppointmentStatusConfirmed {
			for _, wp := range waitPatients {
				if appt.PatientID == wp {
					waitPatientBooked[wp] = true
				}
			}
		}
	}
	ctx.Store.mu.RUnlock()

	if len(waitPatientBooked) != capacity {
		t.Errorf("expected %d wait patients to be auto-filled, got %d", capacity, len(waitPatientBooked))
	}
}

func TestAutoFillFromWaitQueue_OnChange(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	appt, _ := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot3ID,
		PatientID: ctx.Patient3ID,
	})

	_, _ = svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient1ID,
		DoctorID:  ctx.Doctor1ID,
		Date:      ctx.FutureDate,
	})

	_, err := svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt.ID,
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient3ID,
	})
	if err != nil {
		t.Fatalf("ChangeAppointment failed: %v", err)
	}

	ctx.Store.mu.RLock()
	appointments := ctx.Store.appointments
	patient1Slot3Appts := 0
	for _, a := range appointments {
		if a.PatientID == ctx.Patient1ID && a.SlotID == ctx.Slot3ID && a.Status == AppointmentStatusConfirmed {
			patient1Slot3Appts++
		}
	}
	ctx.Store.mu.RUnlock()

	if patient1Slot3Appts != 1 {
		t.Errorf("expected 1 appointment for patient1 in slot3, got %d", patient1Slot3Appts)
	}
}

func TestGetAppointment(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	created, _ := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})

	fetched, err := svc.GetAppointment(created.ID)
	if err != nil {
		t.Fatalf("GetAppointment failed: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, fetched.ID)
	}

	_, err = svc.GetAppointment("NON_EXISTENT")
	if err != ErrAppointmentNotFound {
		t.Errorf("expected ErrAppointmentNotFound, got %v", err)
	}
}

func TestListAppointmentsByPatient(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	_, _ = svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot1ID,
		PatientID: ctx.Patient1ID,
	})
	_, _ = svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot2ID,
		PatientID: ctx.Patient1ID,
	})

	appointments, err := svc.ListAppointmentsByPatient(ctx.Patient1ID)
	if err != nil {
		t.Fatalf("ListAppointmentsByPatient failed: %v", err)
	}
	if len(appointments) != 2 {
		t.Errorf("expected 2 appointments, got %d", len(appointments))
	}

	_, err = svc.ListAppointmentsByPatient("INVALID_PATIENT")
	if err != ErrPatientNotFound {
		t.Errorf("expected ErrPatientNotFound, got %v", err)
	}
}

func TestGetSlot(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	slot, err := svc.GetSlot(ctx.Slot1ID)
	if err != nil {
		t.Fatalf("GetSlot failed: %v", err)
	}
	if slot.SlotID != ctx.Slot1ID {
		t.Errorf("expected slot ID %s, got %s", ctx.Slot1ID, slot.SlotID)
	}
	if slot.RemainingCount != 5 {
		t.Errorf("expected remaining 5, got %d", slot.RemainingCount)
	}
	if slot.IsFull {
		t.Error("slot should not be full")
	}

	_, err = svc.GetSlot("INVALID_SLOT")
	if err != ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestFullWorkflow(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	schedules, err := svc.QuerySchedules(&QueryScheduleRequest{DoctorID: ctx.Doctor1ID})
	if err != nil {
		t.Fatalf("Step 1 QuerySchedules failed: %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("Step 1 expected 1 schedule, got %d", len(schedules))
	}
	if len(schedules[0].Slots) != 3 {
		t.Fatalf("Step 1 expected 3 slots, got %d", len(schedules[0].Slots))
	}

	slot0 := schedules[0].Slots[0]
	appt, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    slot0.SlotID,
		PatientID: ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("Step 2 CreateAppointment failed: %v", err)
	}
	if appt.Status != AppointmentStatusConfirmed {
		t.Fatalf("Step 2 expected CONFIRMED, got %s", appt.Status)
	}

	slotAfter, _ := svc.GetSlot(slot0.SlotID)
	if slotAfter.BookedCount != 1 {
		t.Fatalf("Step 2 expected booked count 1, got %d", slotAfter.BookedCount)
	}

	changed, err := svc.ChangeAppointment(&ChangeAppointmentRequest{
		AppointmentID: appt.ID,
		NewSlotID:     ctx.Slot2ID,
		PatientID:     ctx.Patient1ID,
	})
	if err != nil {
		t.Fatalf("Step 3 ChangeAppointment failed: %v", err)
	}
	if changed.SlotID != ctx.Slot2ID {
		t.Fatalf("Step 3 expected slot %s, got %s", ctx.Slot2ID, changed.SlotID)
	}

	slot0AfterChange, _ := svc.GetSlot(slot0.SlotID)
	if slot0AfterChange.BookedCount != 0 {
		t.Fatalf("Step 3 expected old slot booked 0, got %d", slot0AfterChange.BookedCount)
	}
	slot2After, _ := svc.GetSlot(ctx.Slot2ID)
	if slot2After.BookedCount != 1 {
		t.Fatalf("Step 3 expected new slot booked 1, got %d", slot2After.BookedCount)
	}

	cancelled, err := svc.CancelAppointment(&CancelAppointmentRequest{
		AppointmentID: changed.ID,
		PatientID:     ctx.Patient1ID,
		Reason:        "临时有事取消",
	})
	if err != nil {
		t.Fatalf("Step 4 CancelAppointment failed: %v", err)
	}
	if cancelled.Status != AppointmentStatusCancelled {
		t.Fatalf("Step 4 expected CANCELLED, got %s", cancelled.Status)
	}

	slot2AfterCancel, _ := svc.GetSlot(ctx.Slot2ID)
	if slot2AfterCancel.BookedCount != 0 {
		t.Fatalf("Step 4 expected slot2 booked 0 after cancel, got %d", slot2AfterCancel.BookedCount)
	}

	appt2, err := svc.CreateAppointment(&CreateAppointmentRequest{
		SlotID:    ctx.Slot3ID,
		PatientID: ctx.Patient2ID,
	})
	if err != nil {
		t.Fatalf("Step 5 CreateAppointment for patient2 failed: %v", err)
	}

	_, _ = svc.JoinWaitQueue(&JoinWaitQueueRequest{
		PatientID: ctx.Patient3ID,
		DoctorID:  ctx.Doctor1ID,
		Date:      ctx.FutureDate,
	})

	_, err = svc.RecordNoShow(&RecordNoShowRequest{
		AppointmentID: appt2.ID,
		Remark:        "患者未到",
	})
	if err != nil {
		t.Fatalf("Step 5 RecordNoShow failed: %v", err)
	}

	ctx.Store.mu.RLock()
	patient3Appts := 0
	for _, a := range ctx.Store.appointments {
		if a.PatientID == ctx.Patient3ID && a.SlotID == ctx.Slot3ID && a.Status == AppointmentStatusConfirmed {
			patient3Appts++
		}
	}
	ctx.Store.mu.RUnlock()

	if patient3Appts != 1 {
		t.Fatalf("Step 5 expected patient3 auto-filled appointment, count=%d", patient3Appts)
	}
}

func TestConcurrentAppointment_NoOversell(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	slotID := ctx.Slot1ID
	slot, err := svc.GetSlot(slotID)
	if err != nil {
		t.Fatalf("GetSlot failed: %v", err)
	}

	capacity := slot.TotalCapacity
	numGoroutines := capacity * 5

	patientIDs := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		patientID := ctx.Store.generateID("PAT")
		ctx.Store.patients[patientID] = &Patient{
			ID:   patientID,
			Name: fmt.Sprintf("Patient%d", i),
		}
		patientIDs[i] = patientID
	}

	var wg sync.WaitGroup
	successCount := int64(0)
	failCount := int64(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			_, err := svc.CreateAppointment(&CreateAppointmentRequest{
				SlotID:    slotID,
				PatientID: pid,
			})
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}(patientIDs[i])
	}

	wg.Wait()

	slotAfter, err := svc.GetSlot(slotID)
	if err != nil {
		t.Fatalf("GetSlot after concurrent appointments failed: %v", err)
	}

	if slotAfter.BookedCount > capacity {
		t.Errorf("BookedCount %d exceeds capacity %d - oversell occurred!", slotAfter.BookedCount, capacity)
	}

	if slotAfter.BookedCount != capacity {
		t.Errorf("BookedCount %d does not equal capacity %d", slotAfter.BookedCount, capacity)
	}

	if successCount != int64(capacity) {
		t.Errorf("Expected %d successful appointments, got %d", capacity, successCount)
	}

	if failCount != int64(numGoroutines-capacity) {
		t.Errorf("Expected %d failed appointments, got %d", numGoroutines-capacity, failCount)
	}

	if successCount+failCount != int64(numGoroutines) {
		t.Errorf("Total count mismatch: success=%d + fail=%d = %d, expected %d", successCount, failCount, successCount+failCount, numGoroutines)
	}
}

func TestConcurrentAppointment_MultipleSlots(t *testing.T) {
	ctx, _ := setupTestStoreWithMockTime()
	svc := ctx.Service

	slotIDs := []string{ctx.Slot1ID, ctx.Slot2ID, ctx.Slot3ID}
	totalCapacity := 0
	for _, sid := range slotIDs {
		slot, _ := svc.GetSlot(sid)
		totalCapacity += slot.TotalCapacity
	}

	numGoroutines := totalCapacity * 3
	patientIDs := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		patientID := ctx.Store.generateID("PAT")
		ctx.Store.patients[patientID] = &Patient{
			ID:   patientID,
			Name: fmt.Sprintf("Patient%d", i),
		}
		patientIDs[i] = patientID
	}

	var wg sync.WaitGroup
	successCount := int64(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(pid string, slotIdx int) {
			defer wg.Done()
			_, err := svc.CreateAppointment(&CreateAppointmentRequest{
				SlotID:    slotIDs[slotIdx%len(slotIDs)],
				PatientID: pid,
			})
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}(patientIDs[i], i)
	}

	wg.Wait()

	for _, sid := range slotIDs {
		slot, _ := svc.GetSlot(sid)
		if slot.BookedCount > slot.TotalCapacity {
			t.Errorf("Slot %s: BookedCount %d exceeds capacity %d", sid, slot.BookedCount, slot.TotalCapacity)
		}
	}

	if successCount > int64(totalCapacity) {
		t.Errorf("Total success %d exceeds total capacity %d", successCount, totalCapacity)
	}
}
