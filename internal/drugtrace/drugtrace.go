package drugtrace

import (
	"fmt"
	"sort"
	"time"
)

func (s *DrugTraceService) AddDrug(code, name, spec string) error {
	if code == "" {
		return ErrInvalidDrugCode
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.drugs[code] = &Drug{
		Code: code,
		Name: name,
		Spec: spec,
	}
	return nil
}

func (s *DrugTraceService) GetDrug(code string) (*Drug, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	drug, exists := s.drugs[code]
	if !exists {
		return nil, ErrDrugNotFound
	}
	return drug, nil
}

func (s *DrugTraceService) InboundBatch(
	drugCode, batchNumber string,
	quantity int,
	productionDate, expiryDate time.Time,
	supplier, operator string,
) error {
	if batchNumber == "" {
		return ErrInvalidBatchNumber
	}
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	if !productionDate.Before(expiryDate) {
		return ErrInvalidDateRange
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.drugs[drugCode]; !exists {
		return ErrDrugNotFound
	}

	if _, exists := s.batches[batchNumber]; exists {
		return ErrBatchAlreadyExists
	}

	now := time.Now()
	status := BatchStatusNormal
	if isDateExpired(expiryDate, now) {
		status = BatchStatusExpired
	}

	batch := &Batch{
		BatchNumber:    batchNumber,
		DrugCode:       drugCode,
		Quantity:       quantity,
		RemainingQty:   quantity,
		ProductionDate: productionDate,
		ExpiryDate:     expiryDate,
		Supplier:       supplier,
		Status:         status,
		InboundTime:    now,
	}

	s.batches[batchNumber] = batch
	s.drugBatches[drugCode] = append(s.drugBatches[drugCode], batch)

	s.flowIDCounter++
	flow := &StockFlow{
		ID:          fmt.Sprintf("FLOW-%d", s.flowIDCounter),
		BatchNumber: batchNumber,
		DrugCode:    drugCode,
		FlowType:    "inbound",
		Quantity:    quantity,
		Operator:    operator,
		Time:        now,
		Remark:      fmt.Sprintf("入库：供应商=%s", supplier),
	}
	s.stockFlows = append(s.stockFlows, flow)

	return nil
}

func (s *DrugTraceService) GetBatch(batchNumber string) (*Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	batch, exists := s.batches[batchNumber]
	if !exists {
		return nil, ErrBatchNotFound
	}
	return batch, nil
}

func (s *DrugTraceService) GetExpiringBatches(days int) ([]*Batch, error) {
	if days < 0 {
		return nil, fmt.Errorf("days must be non-negative")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	result := make([]*Batch, 0)
	for _, batch := range s.batches {
		if batch.Status == BatchStatusRecalled {
			continue
		}
		if isDateWithinDays(batch.ExpiryDate, now, days) {
			result = append(result, batch)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ExpiryDate.Before(result[j].ExpiryDate)
	})

	return result, nil
}

func (s *DrugTraceService) GetExpiredBatches() ([]*Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	result := make([]*Batch, 0)
	for _, batch := range s.batches {
		if batch.Status == BatchStatusExpired ||
			(batch.Status == BatchStatusNormal && isDateExpired(batch.ExpiryDate, now)) {
			result = append(result, batch)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ExpiryDate.Before(result[j].ExpiryDate)
	})

	return result, nil
}

type OutboundResult struct {
	Details    []*OutboundDetail
	TotalQty   int
	Success    bool
	Error      error
}

func (s *DrugTraceService) OutboundFIFO(
	drugCode string,
	quantity int,
	department, patient, operator string,
) (*OutboundResult, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if _, exists := s.drugs[drugCode]; !exists {
		return nil, ErrDrugNotFound
	}

	batches, exists := s.drugBatches[drugCode]
	if !exists || len(batches) == 0 {
		return nil, ErrInsufficientStock
	}

	availableBatches := make([]*Batch, 0)
	for _, batch := range batches {
		if batch.Status == BatchStatusRecalled {
			continue
		}
		if batch.Status == BatchStatusExpired {
			continue
		}
		if isDateExpired(batch.ExpiryDate, now) {
			batch.Status = BatchStatusExpired
			continue
		}
		if batch.RemainingQty > 0 {
			availableBatches = append(availableBatches, batch)
		}
	}

	if len(availableBatches) == 0 {
		return nil, ErrInsufficientStock
	}

	sort.Slice(availableBatches, func(i, j int) bool {
		if availableBatches[i].ExpiryDate.Equal(availableBatches[j].ExpiryDate) {
			return availableBatches[i].InboundTime.Before(availableBatches[j].InboundTime)
		}
		return availableBatches[i].ExpiryDate.Before(availableBatches[j].ExpiryDate)
	})

	totalAvailable := 0
	for _, batch := range availableBatches {
		totalAvailable += batch.RemainingQty
	}
	if totalAvailable < quantity {
		return nil, ErrInsufficientStock
	}

	result := &OutboundResult{
		Details: make([]*OutboundDetail, 0),
		Success: true,
	}

	remainingQty := quantity

	for _, batch := range availableBatches {
		if remainingQty <= 0 {
			break
		}

		if isDateExpired(batch.ExpiryDate, now) {
			batch.Status = BatchStatusExpired
			continue
		}

		takeQty := batch.RemainingQty
		if takeQty > remainingQty {
			takeQty = remainingQty
		}

		batch.RemainingQty -= takeQty
		remainingQty -= takeQty

		s.outboundCounter++
		detail := &OutboundDetail{
			ID:           fmt.Sprintf("OUT-%d", s.outboundCounter),
			BatchNumber:  batch.BatchNumber,
			DrugCode:     drugCode,
			Quantity:     takeQty,
			Department:   department,
			Patient:      patient,
			Operator:     operator,
			OutboundTime: now,
		}

		s.outboundDetails = append(s.outboundDetails, detail)
		result.Details = append(result.Details, detail)
		result.TotalQty += takeQty

		s.flowIDCounter++
		flow := &StockFlow{
			ID:          fmt.Sprintf("FLOW-%d", s.flowIDCounter),
			BatchNumber: batch.BatchNumber,
			DrugCode:    drugCode,
			FlowType:    "outbound",
			Quantity:    takeQty,
			Operator:    operator,
			Time:        now,
			Remark:      fmt.Sprintf("出库：科室=%s, 患者=%s", department, patient),
		}
		s.stockFlows = append(s.stockFlows, flow)
	}

	return result, nil
}

func (s *DrugTraceService) RecallBatch(batchNumber, reason, operator string) error {
	if batchNumber == "" {
		return ErrInvalidBatchNumber
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	batch, exists := s.batches[batchNumber]
	if !exists {
		return ErrBatchNotFound
	}

	if batch.Status == BatchStatusRecalled {
		return nil
	}

	now := time.Now()
	batch.Status = BatchStatusRecalled
	batch.RecallReason = reason
	batch.RecallTime = &now

	s.flowIDCounter++
	flow := &StockFlow{
		ID:          fmt.Sprintf("FLOW-%d", s.flowIDCounter),
		BatchNumber: batchNumber,
		DrugCode:    batch.DrugCode,
		FlowType:    "recall",
		Quantity:    batch.RemainingQty,
		Operator:    operator,
		Time:        now,
		Remark:      fmt.Sprintf("召回：原因=%s", reason),
	}
	s.stockFlows = append(s.stockFlows, flow)

	return nil
}

func (s *DrugTraceService) GetBatchFlowTrace(batchNumber string) ([]*OutboundDetail, error) {
	if batchNumber == "" {
		return nil, ErrInvalidBatchNumber
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.batches[batchNumber]; !exists {
		return nil, ErrBatchNotFound
	}

	result := make([]*OutboundDetail, 0)
	for _, detail := range s.outboundDetails {
		if detail.BatchNumber == batchNumber {
			result = append(result, detail)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].OutboundTime.Before(result[j].OutboundTime)
	})

	return result, nil
}

func (s *DrugTraceService) GetStockFlows() []*StockFlow {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*StockFlow, len(s.stockFlows))
	copy(result, s.stockFlows)
	return result
}

func (s *DrugTraceService) GetOutboundDetails() []*OutboundDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*OutboundDetail, len(s.outboundDetails))
	copy(result, s.outboundDetails)
	return result
}

func (s *DrugTraceService) GetDrugBatches(drugCode string) ([]*Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.drugs[drugCode]; !exists {
		return nil, ErrDrugNotFound
	}

	batches, exists := s.drugBatches[drugCode]
	if !exists {
		return []*Batch{}, nil
	}

	result := make([]*Batch, len(batches))
	copy(result, batches)

	sort.Slice(result, func(i, j int) bool {
		if result[i].ExpiryDate.Equal(result[j].ExpiryDate) {
			return result[i].InboundTime.Before(result[j].InboundTime)
		}
		return result[i].ExpiryDate.Before(result[j].ExpiryDate)
	})

	return result, nil
}

func (s *DrugTraceService) UpdateBatchStatus() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, batch := range s.batches {
		if batch.Status == BatchStatusNormal && isDateExpired(batch.ExpiryDate, now) {
			batch.Status = BatchStatusExpired
		}
	}
}

func (s *DrugTraceService) GetDrugStock(drugCode string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.drugs[drugCode]; !exists {
		return 0, ErrDrugNotFound
	}

	batches, exists := s.drugBatches[drugCode]
	if !exists {
		return 0, nil
	}

	total := 0
	for _, batch := range batches {
		if batch.Status == BatchStatusNormal {
			total += batch.RemainingQty
		}
	}

	return total, nil
}
