package drugtrace

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrBatchNotFound         = errors.New("batch not found")
	ErrDrugNotFound          = errors.New("drug not found")
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrBatchExpired          = errors.New("batch has expired")
	ErrBatchRecalled         = errors.New("batch has been recalled")
	ErrInvalidQuantity       = errors.New("invalid quantity, must be positive")
	ErrInvalidDateRange      = errors.New("production date must be before expiry date")
	ErrBatchAlreadyExists    = errors.New("batch already exists")
	ErrInvalidBatchNumber    = errors.New("invalid batch number")
	ErrInvalidDrugCode       = errors.New("invalid drug code")
)

type BatchStatus int

const (
	BatchStatusNormal   BatchStatus = iota
	BatchStatusExpired  BatchStatus = iota
	BatchStatusRecalled BatchStatus = iota
)

type Drug struct {
	Code string
	Name string
	Spec string
}

type Batch struct {
	BatchNumber    string
	DrugCode       string
	Quantity       int
	RemainingQty   int
	ProductionDate time.Time
	ExpiryDate     time.Time
	Supplier       string
	Status         BatchStatus
	InboundTime    time.Time
	RecallReason   string
	RecallTime     *time.Time
}

type StockFlow struct {
	ID          string
	BatchNumber string
	DrugCode    string
	FlowType    string
	Quantity    int
	Operator    string
	Time        time.Time
	Remark      string
}

type OutboundDetail struct {
	ID           string
	BatchNumber  string
	DrugCode     string
	Quantity     int
	Department   string
	Patient      string
	Operator     string
	OutboundTime time.Time
}

type DrugTraceService struct {
	mu              sync.RWMutex
	drugs           map[string]*Drug
	batches         map[string]*Batch
	drugBatches     map[string][]*Batch
	stockFlows      []*StockFlow
	outboundDetails []*OutboundDetail
	flowIDCounter   int
	outboundCounter int
}

func NewDrugTraceService() *DrugTraceService {
	return &DrugTraceService{
		drugs:           make(map[string]*Drug),
		batches:         make(map[string]*Batch),
		drugBatches:     make(map[string][]*Batch),
		stockFlows:      make([]*StockFlow, 0),
		outboundDetails: make([]*OutboundDetail, 0),
	}
}

func (s *BatchStatus) String() string {
	switch *s {
	case BatchStatusNormal:
		return "normal"
	case BatchStatusExpired:
		return "expired"
	case BatchStatusRecalled:
		return "recalled"
	default:
		return "unknown"
	}
}
