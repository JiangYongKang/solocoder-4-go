package appointment

import (
	"sort"
	"time"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

type AddScheduleRequest struct {
	DoctorID   string
	Date       string
	Slots      []ScheduleSlotSpec
}

type ScheduleSlotSpec struct {
	StartTime     string
	EndTime       string
	TotalCapacity int
}

type QueryScheduleRequest struct {
	DoctorID   string
	Department string
	Date       string
}

type ScheduleQueryResult struct {
	ScheduleID    string
	DoctorID      string
	DoctorName    string
	Department    string
	Date          string
	Slots         []SlotQueryResult
}

type SlotQueryResult struct {
	SlotID         string
	StartTime      string
	EndTime        string
	TotalCapacity  int
	BookedCount    int
	RemainingCount int
	Status         SlotStatus
	IsFull         bool
	IsLocked       bool
}

type LockSlotRequest struct {
	SlotID    string
	PatientID string
}

type ConfirmAppointmentRequest struct {
	SlotID    string
	PatientID string
}

type ReleaseLockRequest struct {
	SlotID    string
	PatientID string
}

type CreateAppointmentRequest struct {
	SlotID    string
	PatientID string
}

type ChangeAppointmentRequest struct {
	AppointmentID string
	NewSlotID     string
	PatientID     string
}

type CancelAppointmentRequest struct {
	AppointmentID string
	PatientID     string
	Reason        string
}

type RecordNoShowRequest struct {
	AppointmentID string
	Remark        string
}

type JoinWaitQueueRequest struct {
	PatientID  string
	DoctorID   string
	Department string
	Date       string
}

func (svc *Service) AddSchedule(req *AddScheduleRequest) (*Schedule, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	doctor, exists := svc.store.doctors[req.DoctorID]
	if !exists {
		return nil, ErrDoctorNotFound
	}

	if !isValidDate(req.Date) {
		return nil, ErrInvalidDate
	}

	scheduleID := svc.store.generateID("SCH")
	schedule := &Schedule{
		ID:         scheduleID,
		DoctorID:   doctor.ID,
		DoctorName: doctor.Name,
		Department: doctor.Department,
		Date:       req.Date,
		Slots:      make([]*ScheduleSlot, 0, len(req.Slots)),
	}

	for _, slotSpec := range req.Slots {
		if slotSpec.TotalCapacity <= 0 {
			continue
		}
		slotID := svc.store.generateID("SLT")
		slot := &ScheduleSlot{
			ID:            slotID,
			DoctorID:      doctor.ID,
			DoctorName:    doctor.Name,
			Department:    doctor.Department,
			Date:          req.Date,
			StartTime:     slotSpec.StartTime,
			EndTime:       slotSpec.EndTime,
			TotalCapacity: slotSpec.TotalCapacity,
			BookedCount:   0,
			Status:        SlotStatusAvailable,
		}
		schedule.Slots = append(schedule.Slots, slot)
		svc.store.slotMap[slotID] = slot
		svc.store.slotScheduleMap[slotID] = scheduleID
	}

	svc.store.schedules[scheduleID] = schedule
	return schedule, nil
}

func (svc *Service) QuerySchedules(req *QueryScheduleRequest) ([]*ScheduleQueryResult, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	now := svc.store.NowFunc()

	if req.Date != "" && !isValidDate(req.Date) {
		return nil, ErrInvalidDate
	}

	var results []*ScheduleQueryResult

	for _, schedule := range svc.store.schedules {
		if req.DoctorID != "" && schedule.DoctorID != req.DoctorID {
			continue
		}
		if req.Department != "" && schedule.Department != req.Department {
			continue
		}
		if req.Date != "" && schedule.Date != req.Date {
			continue
		}

		slotResults := make([]SlotQueryResult, 0, len(schedule.Slots))
		for _, slot := range schedule.Slots {
			slotResults = append(slotResults, svc.buildSlotQueryResult(slot, now))
		}

		sort.Slice(slotResults, func(i, j int) bool {
			return slotResults[i].StartTime < slotResults[j].StartTime
		})

		results = append(results, &ScheduleQueryResult{
			ScheduleID: schedule.ID,
			DoctorID:   schedule.DoctorID,
			DoctorName: schedule.DoctorName,
			Department: schedule.Department,
			Date:       schedule.Date,
			Slots:      slotResults,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Date != results[j].Date {
			return results[i].Date < results[j].Date
		}
		return results[i].DoctorName < results[j].DoctorName
	})

	return results, nil
}

func (svc *Service) buildSlotQueryResult(slot *ScheduleSlot, now time.Time) SlotQueryResult {
	effectiveStatus := slot.Status
	isLocked := slot.Status == SlotStatusLocked

	if slot.Status == SlotStatusLocked && slot.LockExpireAt != nil && now.After(*slot.LockExpireAt) {
		effectiveStatus = SlotStatusAvailable
		isLocked = false
	}

	remaining := slot.TotalCapacity - slot.BookedCount
	isFull := remaining <= 0

	if isFull && effectiveStatus == SlotStatusAvailable {
		effectiveStatus = SlotStatusBooked
	}

	return SlotQueryResult{
		SlotID:         slot.ID,
		StartTime:      slot.StartTime,
		EndTime:        slot.EndTime,
		TotalCapacity:  slot.TotalCapacity,
		BookedCount:    slot.BookedCount,
		RemainingCount: remaining,
		Status:         effectiveStatus,
		IsFull:         isFull,
		IsLocked:       isLocked,
	}
}

func (svc *Service) LockSlot(req *LockSlotRequest) (*ScheduleSlot, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	if svc.store.LockTimeout <= 0 {
		return nil, ErrLockTimeoutInvalid
	}

	if _, exists := svc.store.patients[req.PatientID]; !exists {
		return nil, ErrPatientNotFound
	}

	slot, exists := svc.store.slotMap[req.SlotID]
	if !exists {
		return nil, ErrSlotNotFound
	}

	now := svc.store.NowFunc()

	if slot.Status == SlotStatusLocked {
		if slot.LockExpireAt != nil && now.After(*slot.LockExpireAt) {
			svc.clearLockInternal(slot)
		} else {
			return nil, ErrSlotAlreadyLocked
		}
	}

	remaining := slot.TotalCapacity - slot.BookedCount
	if remaining <= 0 {
		return nil, ErrSlotNoCapacity
	}

	lockedAt := now
	expireAt := now.Add(svc.store.LockTimeout)

	slot.Status = SlotStatusLocked
	slot.LockedBy = req.PatientID
	slot.LockedAt = &lockedAt
	slot.LockExpireAt = &expireAt

	return slot, nil
}

func (svc *Service) ConfirmAppointment(req *ConfirmAppointmentRequest) (*Appointment, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	patient, exists := svc.store.patients[req.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	slot, exists := svc.store.slotMap[req.SlotID]
	if !exists {
		return nil, ErrSlotNotFound
	}

	now := svc.store.NowFunc()

	if slot.Status != SlotStatusLocked {
		return nil, ErrSlotNotLocked
	}

	if slot.LockExpireAt != nil && now.After(*slot.LockExpireAt) {
		svc.clearLockInternal(slot)
		return nil, ErrSlotLockExpired
	}

	if slot.LockedBy != req.PatientID {
		return nil, ErrInvalidLockOwner
	}

	remaining := slot.TotalCapacity - slot.BookedCount
	if remaining <= 0 {
		svc.clearLockInternal(slot)
		return nil, ErrSlotNoCapacity
	}

	apptID := svc.store.generateID("APT")
	confirmedAt := now
	appointment := &Appointment{
		ID:            apptID,
		PatientID:     patient.ID,
		PatientName:   patient.Name,
		SlotID:        slot.ID,
		DoctorID:      slot.DoctorID,
		DoctorName:    slot.DoctorName,
		Department:    slot.Department,
		Date:          slot.Date,
		StartTime:     slot.StartTime,
		EndTime:       slot.EndTime,
		Status:        AppointmentStatusConfirmed,
		IsNoShow:      false,
		CreatedAt:     now,
		ConfirmedAt:   &confirmedAt,
	}

	slot.BookedCount++
	svc.clearLockInternal(slot)

	if slot.BookedCount >= slot.TotalCapacity {
		slot.Status = SlotStatusBooked
	} else {
		slot.Status = SlotStatusAvailable
	}

	svc.store.appointments[apptID] = appointment
	svc.store.patientApptMap[patient.ID] = append(svc.store.patientApptMap[patient.ID], apptID)

	return appointment, nil
}

func (svc *Service) ReleaseLock(req *ReleaseLockRequest) error {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	slot, exists := svc.store.slotMap[req.SlotID]
	if !exists {
		return ErrSlotNotFound
	}

	if slot.Status != SlotStatusLocked {
		return ErrSlotNotLocked
	}

	if slot.LockedBy != req.PatientID {
		return ErrInvalidLockOwner
	}

	svc.clearLockInternal(slot)
	if slot.BookedCount >= slot.TotalCapacity {
		slot.Status = SlotStatusBooked
	} else {
		slot.Status = SlotStatusAvailable
	}
	return nil
}

func (svc *Service) clearLockInternal(slot *ScheduleSlot) {
	slot.Status = SlotStatusAvailable
	slot.LockedBy = ""
	slot.LockedAt = nil
	slot.LockExpireAt = nil
}

func (svc *Service) CheckAndReleaseExpiredLocks() int {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	now := svc.store.NowFunc()
	releasedCount := 0

	for _, slot := range svc.store.slotMap {
		if slot.Status == SlotStatusLocked && slot.LockExpireAt != nil && now.After(*slot.LockExpireAt) {
			svc.clearLockInternal(slot)
			if slot.BookedCount >= slot.TotalCapacity {
				slot.Status = SlotStatusBooked
			} else {
				slot.Status = SlotStatusAvailable
			}
			releasedCount++
		}
	}

	return releasedCount
}

func (svc *Service) CreateAppointment(req *CreateAppointmentRequest) (*Appointment, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	if svc.store.LockTimeout <= 0 {
		return nil, ErrLockTimeoutInvalid
	}

	patient, exists := svc.store.patients[req.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	slot, exists := svc.store.slotMap[req.SlotID]
	if !exists {
		return nil, ErrSlotNotFound
	}

	now := svc.store.NowFunc()

	if slot.Status == SlotStatusLocked {
		if slot.LockExpireAt != nil && now.After(*slot.LockExpireAt) {
			svc.clearLockInternal(slot)
		} else {
			return nil, ErrSlotAlreadyLocked
		}
	}

	remaining := slot.TotalCapacity - slot.BookedCount
	if remaining <= 0 {
		return nil, ErrSlotNoCapacity
	}

	apptID := svc.store.generateID("APT")
	confirmedAt := now
	appointment := &Appointment{
		ID:            apptID,
		PatientID:     patient.ID,
		PatientName:   patient.Name,
		SlotID:        slot.ID,
		DoctorID:      slot.DoctorID,
		DoctorName:    slot.DoctorName,
		Department:    slot.Department,
		Date:          slot.Date,
		StartTime:     slot.StartTime,
		EndTime:       slot.EndTime,
		Status:        AppointmentStatusConfirmed,
		IsNoShow:      false,
		CreatedAt:     now,
		ConfirmedAt:   &confirmedAt,
	}

	slot.BookedCount++
	if slot.Status == SlotStatusLocked {
		svc.clearLockInternal(slot)
	}

	if slot.BookedCount >= slot.TotalCapacity {
		slot.Status = SlotStatusBooked
	} else {
		slot.Status = SlotStatusAvailable
	}

	svc.store.appointments[apptID] = appointment
	svc.store.patientApptMap[patient.ID] = append(svc.store.patientApptMap[patient.ID], apptID)

	return appointment, nil
}

func (svc *Service) ChangeAppointment(req *ChangeAppointmentRequest) (*Appointment, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	patient, exists := svc.store.patients[req.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	oldAppt, exists := svc.store.appointments[req.AppointmentID]
	if !exists {
		return nil, ErrAppointmentNotFound
	}

	if oldAppt.PatientID != req.PatientID {
		return nil, ErrPatientNotFound
	}

	if oldAppt.Status != AppointmentStatusConfirmed {
		return nil, ErrChangeNotAllowed
	}

	if oldAppt.SlotID == req.NewSlotID {
		return nil, ErrSameSlotChange
	}

	now := svc.store.NowFunc()

	apptDate, err := time.Parse("2006-01-02", oldAppt.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}
	apptStart, err := time.Parse("15:04", oldAppt.StartTime)
	if err != nil {
		return nil, ErrInvalidDate
	}
	apptDateTime := time.Date(apptDate.Year(), apptDate.Month(), apptDate.Day(),
		apptStart.Hour(), apptStart.Minute(), 0, 0, now.Location())

	if now.After(apptDateTime.Add(-svc.store.ChangeDeadline)) {
		return nil, ErrCannotChangePast
	}

	newSlot, exists := svc.store.slotMap[req.NewSlotID]
	if !exists {
		return nil, ErrSlotNotFound
	}

	newSlotDate, err := time.Parse("2006-01-02", newSlot.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}
	newSlotStart, err := time.Parse("15:04", newSlot.StartTime)
	if err != nil {
		return nil, ErrInvalidDate
	}
	newSlotDateTime := time.Date(newSlotDate.Year(), newSlotDate.Month(), newSlotDate.Day(),
		newSlotStart.Hour(), newSlotStart.Minute(), 0, 0, now.Location())

	if now.After(newSlotDateTime.Add(-svc.store.ChangeDeadline)) {
		return nil, ErrCannotChangePast
	}

	if newSlot.Status == SlotStatusLocked {
		if newSlot.LockExpireAt != nil && now.After(*newSlot.LockExpireAt) {
			svc.clearLockInternal(newSlot)
		} else if newSlot.LockedBy != req.PatientID {
			return nil, ErrSlotAlreadyLocked
		}
	}

	newRemaining := newSlot.TotalCapacity - newSlot.BookedCount
	if newRemaining <= 0 {
		return nil, ErrSlotNoCapacity
	}

	oldSlot, oldSlotExists := svc.store.slotMap[oldAppt.SlotID]
	if oldSlotExists {
		if oldSlot.BookedCount > 0 {
			oldSlot.BookedCount--
		}
		if oldSlot.Status == SlotStatusBooked {
			oldSlot.Status = SlotStatusAvailable
		}
	}

	newSlot.BookedCount++
	if newSlot.BookedCount >= newSlot.TotalCapacity {
		newSlot.Status = SlotStatusBooked
	}

	oldAppt.Status = AppointmentStatusChanged
	oldAppt.ChangedToID = ""

	newApptID := svc.store.generateID("APT")
	newAppt := &Appointment{
		ID:            newApptID,
		PatientID:     patient.ID,
		PatientName:   patient.Name,
		SlotID:        newSlot.ID,
		DoctorID:      newSlot.DoctorID,
		DoctorName:    newSlot.DoctorName,
		Department:    newSlot.Department,
		Date:          newSlot.Date,
		StartTime:     newSlot.StartTime,
		EndTime:       newSlot.EndTime,
		Status:        AppointmentStatusConfirmed,
		IsNoShow:      false,
		CreatedAt:     now,
		ConfirmedAt:   &now,
		ChangedFromID: oldAppt.ID,
	}

	oldAppt.ChangedToID = newApptID

	svc.store.appointments[newApptID] = newAppt
	svc.store.patientApptMap[patient.ID] = append(svc.store.patientApptMap[patient.ID], newApptID)

	if oldSlotExists {
		svc.tryProcessWaitQueueForSlot(oldSlot)
	}

	return newAppt, nil
}

func (svc *Service) CancelAppointment(req *CancelAppointmentRequest) (*Appointment, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	if _, exists := svc.store.patients[req.PatientID]; !exists {
		return nil, ErrPatientNotFound
	}

	appt, exists := svc.store.appointments[req.AppointmentID]
	if !exists {
		return nil, ErrAppointmentNotFound
	}

	if appt.PatientID != req.PatientID {
		return nil, ErrAppointmentNotFound
	}

	if appt.Status != AppointmentStatusConfirmed {
		return nil, ErrChangeNotAllowed
	}

	now := svc.store.NowFunc()
	apptDate, err := time.Parse("2006-01-02", appt.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}
	apptStart, err := time.Parse("15:04", appt.StartTime)
	if err != nil {
		return nil, ErrInvalidDate
	}
	apptDateTime := time.Date(apptDate.Year(), apptDate.Month(), apptDate.Day(),
		apptStart.Hour(), apptStart.Minute(), 0, 0, now.Location())

	if now.After(apptDateTime.Add(-svc.store.CancelDeadline)) {
		return nil, ErrCannotCancelPast
	}

	slot, slotExists := svc.store.slotMap[appt.SlotID]
	if slotExists {
		if slot.BookedCount > 0 {
			slot.BookedCount--
		}
		if slot.Status == SlotStatusBooked {
			slot.Status = SlotStatusAvailable
		}
	}

	cancelledAt := now
	appt.Status = AppointmentStatusCancelled
	appt.CancelledAt = &cancelledAt

	if slotExists {
		svc.tryProcessWaitQueueForSlot(slot)
	}

	return appt, nil
}

func (svc *Service) RecordNoShow(req *RecordNoShowRequest) (*NoShowRecord, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	if req.Remark == "" {
		return nil, ErrNoShowRemarkRequired
	}

	appt, exists := svc.store.appointments[req.AppointmentID]
	if !exists {
		return nil, ErrAppointmentNotFound
	}

	if appt.Status != AppointmentStatusConfirmed {
		return nil, ErrAppointmentNotActive
	}

	now := svc.store.NowFunc()
	recordID := svc.store.generateID("NOS")

	record := &NoShowRecord{
		ID:            recordID,
		PatientID:     appt.PatientID,
		PatientName:   appt.PatientName,
		AppointmentID: appt.ID,
		DoctorID:      appt.DoctorID,
		Date:          appt.Date,
		StartTime:     appt.StartTime,
		RecordedAt:    now,
		Remark:        req.Remark,
	}

	appt.Status = AppointmentStatusNoShow
	appt.IsNoShow = true

	slot, slotExists := svc.store.slotMap[appt.SlotID]
	if slotExists {
		if slot.BookedCount > 0 {
			slot.BookedCount--
		}
		if slot.Status == SlotStatusBooked {
			slot.Status = SlotStatusAvailable
		}
	}

	svc.store.noShowRecords = append(svc.store.noShowRecords, record)

	if slotExists {
		svc.tryProcessWaitQueueForSlot(slot)
	}

	return record, nil
}

func (svc *Service) GetNoShowRecords(patientID string) ([]*NoShowRecord, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	if patientID != "" {
		if _, exists := svc.store.patients[patientID]; !exists {
			return nil, ErrPatientNotFound
		}
	}

	var results []*NoShowRecord
	for _, record := range svc.store.noShowRecords {
		if patientID != "" && record.PatientID != patientID {
			continue
		}
		results = append(results, record)
	}

	return results, nil
}

func (svc *Service) JoinWaitQueue(req *JoinWaitQueueRequest) (*WaitQueueItem, error) {
	svc.store.mu.Lock()
	defer svc.store.mu.Unlock()

	patient, exists := svc.store.patients[req.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	if req.DoctorID != "" {
		if _, exists := svc.store.doctors[req.DoctorID]; !exists {
			return nil, ErrDoctorNotFound
		}
	}

	if !isValidDate(req.Date) {
		return nil, ErrInvalidDate
	}

	key := waitQueueKey(req.DoctorID, req.Date, req.Department)
	now := svc.store.NowFunc()

	itemID := svc.store.generateID("WQ")
	item := &WaitQueueItem{
		ID:          itemID,
		PatientID:   patient.ID,
		PatientName: patient.Name,
		TargetDate:  req.Date,
		DoctorID:    req.DoctorID,
		Department:  req.Department,
		JoinedAt:    now,
	}

	svc.store.waitQueues[key] = append(svc.store.waitQueues[key], item)
	return item, nil
}

func (svc *Service) GetWaitQueue(doctorID, department, date string) ([]*WaitQueueItem, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	if date != "" && !isValidDate(date) {
		return nil, ErrInvalidDate
	}

	key := waitQueueKey(doctorID, date, department)
	items := svc.store.waitQueues[key]

	result := make([]*WaitQueueItem, len(items))
	copy(result, items)

	sort.Slice(result, func(i, j int) bool {
		return result[i].JoinedAt.Before(result[j].JoinedAt)
	})

	return result, nil
}

func (svc *Service) tryProcessWaitQueueForSlot(slot *ScheduleSlot) {
	if slot.BookedCount >= slot.TotalCapacity {
		return
	}

	candidates := svc.getWaitQueueCandidatesForSlot(slot)
	if len(candidates) == 0 {
		return
	}

	for _, candidate := range candidates {
		if slot.BookedCount >= slot.TotalCapacity {
			break
		}

		patient, patientExists := svc.store.patients[candidate.PatientID]
		if !patientExists {
			svc.removeFromWaitQueue(candidate)
			continue
		}

		now := svc.store.NowFunc()
		apptID := svc.store.generateID("APT")

		appointment := &Appointment{
			ID:            apptID,
			PatientID:     patient.ID,
			PatientName:   patient.Name,
			SlotID:        slot.ID,
			DoctorID:      slot.DoctorID,
			DoctorName:    slot.DoctorName,
			Department:    slot.Department,
			Date:          slot.Date,
			StartTime:     slot.StartTime,
			EndTime:       slot.EndTime,
			Status:        AppointmentStatusConfirmed,
			IsNoShow:      false,
			CreatedAt:     now,
			ConfirmedAt:   &now,
		}

		slot.BookedCount++
		if slot.BookedCount >= slot.TotalCapacity {
			slot.Status = SlotStatusBooked
		}

		svc.store.appointments[apptID] = appointment
		svc.store.patientApptMap[patient.ID] = append(svc.store.patientApptMap[patient.ID], apptID)

		svc.removeFromWaitQueue(candidate)
	}
}

func (svc *Service) getWaitQueueCandidatesForSlot(slot *ScheduleSlot) []*WaitQueueItem {
	var candidates []*WaitQueueItem

	docKey := waitQueueKey(slot.DoctorID, slot.Date, "")
	if items, ok := svc.store.waitQueues[docKey]; ok {
		candidates = append(candidates, items...)
	}

	deptKey := waitQueueKey("", slot.Date, slot.Department)
	if items, ok := svc.store.waitQueues[deptKey]; ok {
		candidates = append(candidates, items...)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].JoinedAt.Before(candidates[j].JoinedAt)
	})

	return candidates
}

func (svc *Service) removeFromWaitQueue(target *WaitQueueItem) {
	keys := []string{
		waitQueueKey(target.DoctorID, target.TargetDate, ""),
		waitQueueKey("", target.TargetDate, target.Department),
	}

	for _, key := range keys {
		items, ok := svc.store.waitQueues[key]
		if !ok {
			continue
		}
		for i, item := range items {
			if item.ID == target.ID {
				svc.store.waitQueues[key] = append(items[:i], items[i+1:]...)
				return
			}
		}
	}
}

func (svc *Service) GetAppointment(id string) (*Appointment, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	appt, exists := svc.store.appointments[id]
	if !exists {
		return nil, ErrAppointmentNotFound
	}
	return appt, nil
}

func (svc *Service) ListAppointmentsByPatient(patientID string) ([]*Appointment, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	if _, exists := svc.store.patients[patientID]; !exists {
		return nil, ErrPatientNotFound
	}

	ids := svc.store.patientApptMap[patientID]
	result := make([]*Appointment, 0, len(ids))
	for _, id := range ids {
		if appt, ok := svc.store.appointments[id]; ok {
			result = append(result, appt)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Date != result[j].Date {
			return result[i].Date < result[j].Date
		}
		return result[i].StartTime < result[j].StartTime
	})

	return result, nil
}

func (svc *Service) GetSlot(id string) (*SlotQueryResult, error) {
	svc.store.mu.RLock()
	defer svc.store.mu.RUnlock()

	slot, exists := svc.store.slotMap[id]
	if !exists {
		return nil, ErrSlotNotFound
	}

	result := svc.buildSlotQueryResult(slot, svc.store.NowFunc())
	return &result, nil
}

func isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}
