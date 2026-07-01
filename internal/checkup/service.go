package checkup

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Store struct {
	mu            sync.RWMutex
	patients      map[string]*Patient
	items         map[string]*CheckItem
	packages      map[string]*CheckPackage
	timeSlots     map[string]*TimeSlot
	appointments  map[string]*Appointment
	results       map[string][]*CheckResult
	apptResultMap map[string]map[string]*CheckResult
	reports       map[string]*Report
	idCounter     int64
}

func NewStore() *Store {
	return &Store{
		patients:      make(map[string]*Patient),
		items:         make(map[string]*CheckItem),
		packages:      make(map[string]*CheckPackage),
		timeSlots:     make(map[string]*TimeSlot),
		appointments:  make(map[string]*Appointment),
		results:       make(map[string][]*CheckResult),
		apptResultMap: make(map[string]map[string]*CheckResult),
		reports:       make(map[string]*Report),
	}
}

func (s *Store) generateID(prefix string) string {
	s.idCounter++
	return fmt.Sprintf("%s%010d", prefix, s.idCounter)
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

func (s *Store) GetPatient(id string) (*Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.patients[id]
	if !exists {
		return nil, ErrPatientNotFound
	}
	return p, nil
}

type AddItemRequest struct {
	Name        string
	Description string
	Category    ItemCategory
	Unit        string
	MinValue    float64
	MaxValue    float64
	Price       float64
}

func (s *Store) AddItem(req *AddItemRequest) *CheckItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := &CheckItem{
		ID:          s.generateID("ITM"),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Unit:        req.Unit,
		MinValue:    req.MinValue,
		MaxValue:    req.MaxValue,
		Price:       req.Price,
	}
	s.items[item.ID] = item
	return item
}

func (s *Store) GetItem(id string) (*CheckItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.items[id]
	if !exists {
		return nil, ErrItemNotFound
	}
	return item, nil
}

func (s *Store) ListItems() []*CheckItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*CheckItem, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

type CreatePackageRequest struct {
	Name        string
	Description string
	ItemIDs     []string
}

func (s *Store) CreatePackage(req *CreatePackageRequest) (*CheckPackage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(req.ItemIDs) == 0 {
		return nil, ErrEmptyPackageItems
	}

	seen := make(map[string]bool)
	for _, id := range req.ItemIDs {
		if seen[id] {
			return nil, ErrDuplicateItemInPackage
		}
		seen[id] = true
	}

	itemMap := make(map[string]*CheckItem)
	var totalPrice float64
	for _, id := range req.ItemIDs {
		item, exists := s.items[id]
		if !exists {
			return nil, fmt.Errorf("%w: %s", ErrInvalidItemInPackage, id)
		}
		itemMap[id] = item
		totalPrice += item.Price
	}

	pkg := &CheckPackage{
		ID:          s.generateID("PKG"),
		Name:        req.Name,
		Description: req.Description,
		ItemIDs:     append([]string{}, req.ItemIDs...),
		Items:       itemMap,
		TotalPrice:  totalPrice,
		CreatedAt:   time.Now(),
	}
	s.packages[pkg.ID] = pkg
	return pkg, nil
}

func (s *Store) GetPackage(id string) (*CheckPackage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pkg, exists := s.packages[id]
	if !exists {
		return nil, ErrPackageNotFound
	}
	return pkg, nil
}

func (s *Store) ListPackages() []*CheckPackage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*CheckPackage, 0, len(s.packages))
	for _, pkg := range s.packages {
		result = append(result, pkg)
	}
	return result
}

type CreateTimeSlotRequest struct {
	Date      time.Time
	StartTime time.Time
	EndTime   time.Time
	Capacity  int
}

func (s *Store) CreateTimeSlot(req *CreateTimeSlotRequest) (*TimeSlot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !req.StartTime.Before(req.EndTime) {
		return nil, ErrInvalidTimeRange
	}
	if req.Capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	slot := &TimeSlot{
		ID:           s.generateID("TMS"),
		Date:         req.Date,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Capacity:     req.Capacity,
		CurrentCount: 0,
	}
	s.timeSlots[slot.ID] = slot
	return slot, nil
}

func (s *Store) GetTimeSlot(id string) (*TimeSlot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	slot, exists := s.timeSlots[id]
	if !exists {
		return nil, ErrTimeSlotNotFound
	}
	return slot, nil
}

func (s *Store) ListTimeSlots() []*TimeSlot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*TimeSlot, 0, len(s.timeSlots))
	for _, slot := range s.timeSlots {
		result = append(result, slot)
	}
	return result
}

type CreateAppointmentRequest struct {
	PatientID  string
	PackageID  string
	TimeSlotID string
}

func (s *Store) CreateAppointment(req *CreateAppointmentRequest) (*Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	patient, exists := s.patients[req.PatientID]
	if !exists {
		return nil, ErrPatientNotFound
	}

	pkg, exists := s.packages[req.PackageID]
	if !exists {
		return nil, ErrPackageNotFound
	}

	slot, exists := s.timeSlots[req.TimeSlotID]
	if !exists {
		return nil, ErrTimeSlotNotFound
	}

	if slot.CurrentCount >= slot.Capacity {
		return nil, ErrTimeSlotCapacityFull
	}

	slot.CurrentCount++

	timeSlotInfo := fmt.Sprintf("%s %s-%s",
		slot.Date.Format("2006-01-02"),
		slot.StartTime.Format("15:04"),
		slot.EndTime.Format("15:04"),
	)

	appt := &Appointment{
		ID:           s.generateID("APT"),
		PatientID:    patient.ID,
		PatientName:  patient.Name,
		PackageID:    pkg.ID,
		PackageName:  pkg.Name,
		TimeSlotID:   slot.ID,
		TimeSlotInfo: timeSlotInfo,
		Status:       AppointmentStatusPending,
		CreatedAt:    time.Now(),
	}
	s.appointments[appt.ID] = appt
	s.results[appt.ID] = make([]*CheckResult, 0)
	s.apptResultMap[appt.ID] = make(map[string]*CheckResult)
	return appt, nil
}

func (s *Store) GetAppointment(id string) (*Appointment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	appt, exists := s.appointments[id]
	if !exists {
		return nil, ErrAppointmentNotFound
	}
	return appt, nil
}

func (s *Store) CancelAppointment(id string) (*Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	appt, exists := s.appointments[id]
	if !exists {
		return nil, ErrAppointmentNotFound
	}

	if appt.Status == AppointmentStatusCancelled {
		return appt, nil
	}

	slot, slotExists := s.timeSlots[appt.TimeSlotID]
	if slotExists && slot.CurrentCount > 0 {
		slot.CurrentCount--
	}

	appt.Status = AppointmentStatusCancelled
	return appt, nil
}

func (s *Store) ListAppointments() []*Appointment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Appointment, 0, len(s.appointments))
	for _, appt := range s.appointments {
		result = append(result, appt)
	}
	return result
}

type RecordResultRequest struct {
	AppointmentID string
	ItemID        string
	Value         string
	Remarks       string
}

func (s *Store) RecordResult(req *RecordResultRequest) (*CheckResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	appt, exists := s.appointments[req.AppointmentID]
	if !exists {
		return nil, ErrAppointmentNotFound
	}

	if appt.Status == AppointmentStatusCancelled {
		return nil, ErrAppointmentCancelled
	}

	if _, exists := s.packages[appt.PackageID]; !exists {
		return nil, ErrPackageNotFound
	}
	pkg := s.packages[appt.PackageID]

	if _, inPackage := pkg.Items[req.ItemID]; !inPackage {
		return nil, ErrItemNotInPackage
	}

	if _, alreadyExists := s.apptResultMap[req.AppointmentID][req.ItemID]; alreadyExists {
		return nil, ErrDuplicateResult
	}

	item := pkg.Items[req.ItemID]

	numericValue, err := strconv.ParseFloat(req.Value, 64)
	isNumeric := err == nil

	var isAbnormal bool
	var reference string
	if isNumeric {
		if item.MinValue != 0 || item.MaxValue != 0 {
			reference = fmt.Sprintf("%.2f - %.2f %s", item.MinValue, item.MaxValue, item.Unit)
			if numericValue < item.MinValue || numericValue > item.MaxValue {
				isAbnormal = true
			}
		}
	} else {
		reference = req.Remarks
	}

	result := &CheckResult{
		ID:            s.generateID("RES"),
		AppointmentID: req.AppointmentID,
		ItemID:        req.ItemID,
		ItemName:      item.Name,
		Value:         req.Value,
		NumericValue:  numericValue,
		IsNumeric:     isNumeric,
		IsAbnormal:    isAbnormal,
		Unit:          item.Unit,
		Reference:     reference,
		Remarks:       req.Remarks,
		RecordedAt:    time.Now(),
	}

	s.results[req.AppointmentID] = append(s.results[req.AppointmentID], result)
	s.apptResultMap[req.AppointmentID][req.ItemID] = result

	resultCount := len(s.apptResultMap[req.AppointmentID])
	requiredCount := len(pkg.Items)
	if resultCount == requiredCount {
		appt.Status = AppointmentStatusCompleted
	} else if resultCount > 0 {
		appt.Status = AppointmentStatusChecking
	}

	return result, nil
}

func (s *Store) GetResult(id string) (*CheckResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, results := range s.results {
		for _, r := range results {
			if r.ID == id {
				return r, nil
			}
		}
	}
	return nil, ErrResultNotFound
}

func (s *Store) GetResultsByAppointment(appointmentID string) ([]*CheckResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, exists := s.appointments[appointmentID]; !exists {
		return nil, ErrAppointmentNotFound
	}
	results, exists := s.results[appointmentID]
	if !exists {
		return []*CheckResult{}, nil
	}
	return append([]*CheckResult{}, results...), nil
}

type GenerateReportRequest struct {
	AppointmentID string
}

func (s *Store) GenerateReport(req *GenerateReportRequest) (*Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	appt, exists := s.appointments[req.AppointmentID]
	if !exists {
		return nil, ErrAppointmentNotFound
	}

	if appt.Status == AppointmentStatusCancelled {
		return nil, ErrAppointmentCancelled
	}

	pkg, exists := s.packages[appt.PackageID]
	if !exists {
		return nil, ErrPackageNotFound
	}

	apptResults, exists := s.apptResultMap[req.AppointmentID]
	if !exists {
		apptResults = make(map[string]*CheckResult)
	}

	if len(apptResults) < len(pkg.Items) {
		return nil, ErrResultsIncomplete
	}

	if _, reportExists := s.reports[req.AppointmentID]; reportExists {
		return s.reports[req.AppointmentID], nil
	}

	resultsCopy := make(map[string]*CheckResult)
	abnormalItems := make([]*AbnormalItem, 0)
	abnormalCount := 0

	for itemID, result := range apptResults {
		resultsCopy[itemID] = result
		if result.IsAbnormal {
			abnormalCount++
			abnormalItems = append(abnormalItems, &AbnormalItem{
				ItemID:    result.ItemID,
				ItemName:  result.ItemName,
				Value:     result.Value,
				Unit:      result.Unit,
				Reference: result.Reference,
				Remarks:   result.Remarks,
			})
		}
	}

	var summary string
	if abnormalCount == 0 {
		summary = fmt.Sprintf("本次体检共 %d 项检查，所有指标均在正常范围内。", len(pkg.Items))
	} else {
		summary = fmt.Sprintf("本次体检共 %d 项检查，其中 %d 项指标异常，请及时关注并咨询医生。", len(pkg.Items), abnormalCount)
	}

	report := &Report{
		ID:            s.generateID("RPT"),
		AppointmentID: appt.ID,
		PatientID:     appt.PatientID,
		PatientName:   appt.PatientName,
		PackageID:     pkg.ID,
		PackageName:   pkg.Name,
		Results:       resultsCopy,
		AbnormalItems: abnormalItems,
		GeneratedAt:   time.Now(),
		Summary:       summary,
	}
	s.reports[req.AppointmentID] = report
	return report, nil
}

func (s *Store) GetReport(appointmentID string) (*Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	report, exists := s.reports[appointmentID]
	if !exists {
		return nil, ErrReportNotFound
	}
	return report, nil
}
