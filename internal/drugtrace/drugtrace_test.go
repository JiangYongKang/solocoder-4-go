package drugtrace

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestAddDrug(t *testing.T) {
	s := NewDrugTraceService()

	t.Run("success", func(t *testing.T) {
		err := s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")
		if err != nil {
			t.Fatalf("AddDrug failed: %v", err)
		}

		drug, err := s.GetDrug("DRUG001")
		if err != nil {
			t.Fatalf("GetDrug failed: %v", err)
		}
		if drug.Code != "DRUG001" || drug.Name != "阿莫西林胶囊" {
			t.Errorf("drug info mismatch, got %+v", drug)
		}
	})

	t.Run("invalid drug code", func(t *testing.T) {
		err := s.AddDrug("", "test", "test")
		if !errors.Is(err, ErrInvalidDrugCode) {
			t.Errorf("expected ErrInvalidDrugCode, got %v", err)
		}
	})
}

func TestGetDrugNotFound(t *testing.T) {
	s := NewDrugTraceService()
	_, err := s.GetDrug("NOTEXIST")
	if !errors.Is(err, ErrDrugNotFound) {
		t.Errorf("expected ErrDrugNotFound, got %v", err)
	}
}

func TestInboundBatch(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	prodDate := time.Now().AddDate(0, -1, 0)
	expDate := time.Now().AddDate(1, 0, 0)

	t.Run("success", func(t *testing.T) {
		err := s.InboundBatch("DRUG001", "BATCH001", 100, prodDate, expDate, "供应商A", "张三")
		if err != nil {
			t.Fatalf("InboundBatch failed: %v", err)
		}

		batch, err := s.GetBatch("BATCH001")
		if err != nil {
			t.Fatalf("GetBatch failed: %v", err)
		}
		if batch.RemainingQty != 100 {
			t.Errorf("expected remaining qty 100, got %d", batch.RemainingQty)
		}
		if batch.Status != BatchStatusNormal {
			t.Errorf("expected status Normal, got %v", batch.Status)
		}

		flows := s.GetStockFlows()
		if len(flows) != 1 {
			t.Errorf("expected 1 stock flow, got %d", len(flows))
		}
		if flows[0].FlowType != "inbound" {
			t.Errorf("expected flow type inbound, got %s", flows[0].FlowType)
		}
	})

	t.Run("duplicate batch", func(t *testing.T) {
		err := s.InboundBatch("DRUG001", "BATCH001", 100, prodDate, expDate, "供应商A", "张三")
		if !errors.Is(err, ErrBatchAlreadyExists) {
			t.Errorf("expected ErrBatchAlreadyExists, got %v", err)
		}
	})

	t.Run("drug not found", func(t *testing.T) {
		err := s.InboundBatch("INVALID", "BATCH002", 100, prodDate, expDate, "供应商A", "张三")
		if !errors.Is(err, ErrDrugNotFound) {
			t.Errorf("expected ErrDrugNotFound, got %v", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		err := s.InboundBatch("DRUG001", "BATCH003", 0, prodDate, expDate, "供应商A", "张三")
		if !errors.Is(err, ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity, got %v", err)
		}

		err = s.InboundBatch("DRUG001", "BATCH003", -10, prodDate, expDate, "供应商A", "张三")
		if !errors.Is(err, ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("invalid date range", func(t *testing.T) {
		badExpDate := time.Now().AddDate(0, -2, 0)
		err := s.InboundBatch("DRUG001", "BATCH004", 100, prodDate, badExpDate, "供应商A", "张三")
		if !errors.Is(err, ErrInvalidDateRange) {
			t.Errorf("expected ErrInvalidDateRange, got %v", err)
		}
	})

	t.Run("empty batch number", func(t *testing.T) {
		err := s.InboundBatch("DRUG001", "", 100, prodDate, expDate, "供应商A", "张三")
		if !errors.Is(err, ErrInvalidBatchNumber) {
			t.Errorf("expected ErrInvalidBatchNumber, got %v", err)
		}
	})

	t.Run("expired batch on arrival", func(t *testing.T) {
		expiredProd := time.Now().AddDate(-2, 0, 0)
		expiredExp := time.Now().AddDate(-1, 0, 0)
		err := s.InboundBatch("DRUG001", "BATCH005", 100, expiredProd, expiredExp, "供应商B", "张三")
		if err != nil {
			t.Fatalf("InboundBatch should not fail for expired batch: %v", err)
		}

		batch, _ := s.GetBatch("BATCH005")
		if batch.Status != BatchStatusExpired {
			t.Errorf("expected status Expired, got %v", batch.Status)
		}
	})
}

func TestGetBatchNotFound(t *testing.T) {
	s := NewDrugTraceService()
	_, err := s.GetBatch("NOTEXIST")
	if !errors.Is(err, ErrBatchNotFound) {
		t.Errorf("expected ErrBatchNotFound, got %v", err)
	}
}

func TestGetExpiringBatches(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	prodDate := now.AddDate(0, -1, 0)

	batches := []struct {
		no      string
		expDays int
	}{
		{"B1", 10},
		{"B2", 30},
		{"B3", 90},
		{"B4", 365},
	}

	for _, b := range batches {
		expDate := now.AddDate(0, 0, b.expDays)
		_ = s.InboundBatch("DRUG001", b.no, 100, prodDate, expDate, "供应商A", "张三")
	}

	t.Run("30 days warning", func(t *testing.T) {
		result, err := s.GetExpiringBatches(30)
		if err != nil {
			t.Fatalf("GetExpiringBatches failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 batches, got %d", len(result))
		}
		if result[0].BatchNumber != "B1" || result[1].BatchNumber != "B2" {
			t.Errorf("unexpected batch order: %v, %v", result[0].BatchNumber, result[1].BatchNumber)
		}
	})

	t.Run("7 days warning", func(t *testing.T) {
		result, err := s.GetExpiringBatches(7)
		if err != nil {
			t.Fatalf("GetExpiringBatches failed: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 batches, got %d", len(result))
		}
	})

	t.Run("negative days", func(t *testing.T) {
		_, err := s.GetExpiringBatches(-1)
		if err == nil {
			t.Error("expected error for negative days")
		}
	})

	t.Run("recalled batch excluded", func(t *testing.T) {
		_ = s.RecallBatch("B1", "质量问题", "管理员")
		result, err := s.GetExpiringBatches(30)
		if err != nil {
			t.Fatalf("GetExpiringBatches failed: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 batch after recall, got %d", len(result))
		}
		if result[0].BatchNumber != "B2" {
			t.Errorf("expected B2, got %s", result[0].BatchNumber)
		}
	})
}

func TestGetExpiredBatches(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()

	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(-2, 0, 0), now.AddDate(-1, 0, 0), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B2", 100,
		now.AddDate(-3, 0, 0), now.AddDate(0, 0, -180), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B3", 100,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")

	result, err := s.GetExpiredBatches()
	if err != nil {
		t.Fatalf("GetExpiredBatches failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 expired batches, got %d", len(result))
	}
	if result[0].BatchNumber != "B1" || result[1].BatchNumber != "B2" {
		t.Errorf("unexpected batch order")
	}
}

func TestOutboundFIFO(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	prodDate := now.AddDate(0, -6, 0)

	_ = s.InboundBatch("DRUG001", "B1", 50,
		prodDate, now.AddDate(0, 0, 30), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B2", 100,
		prodDate, now.AddDate(0, 0, 60), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B3", 75,
		prodDate, now.AddDate(0, 0, 90), "供应商A", "张三")

	t.Run("basic FIFO by expiry", func(t *testing.T) {
		result, err := s.OutboundFIFO("DRUG001", 70, "内科", "患者A", "李四")
		if err != nil {
			t.Fatalf("OutboundFIFO failed: %v", err)
		}
		if !result.Success {
			t.Error("outbound should succeed")
		}
		if result.TotalQty != 70 {
			t.Errorf("expected total qty 70, got %d", result.TotalQty)
		}
		if len(result.Details) != 2 {
			t.Errorf("expected 2 details, got %d", len(result.Details))
		}
		if result.Details[0].BatchNumber != "B1" || result.Details[0].Quantity != 50 {
			t.Errorf("first detail should be B1 qty 50, got %s qty %d",
				result.Details[0].BatchNumber, result.Details[0].Quantity)
		}
		if result.Details[1].BatchNumber != "B2" || result.Details[1].Quantity != 20 {
			t.Errorf("second detail should be B2 qty 20, got %s qty %d",
				result.Details[1].BatchNumber, result.Details[1].Quantity)
		}

		b1, _ := s.GetBatch("B1")
		if b1.RemainingQty != 0 {
			t.Errorf("B1 remaining should be 0, got %d", b1.RemainingQty)
		}
		b2, _ := s.GetBatch("B2")
		if b2.RemainingQty != 80 {
			t.Errorf("B2 remaining should be 80, got %d", b2.RemainingQty)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		_, err := s.OutboundFIFO("DRUG001", 0, "内科", "患者B", "李四")
		if !errors.Is(err, ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity, got %v", err)
		}

		_, err = s.OutboundFIFO("DRUG001", -5, "内科", "患者B", "李四")
		if !errors.Is(err, ErrInvalidQuantity) {
			t.Errorf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("drug not found", func(t *testing.T) {
		_, err := s.OutboundFIFO("INVALID", 10, "内科", "患者B", "李四")
		if !errors.Is(err, ErrDrugNotFound) {
			t.Errorf("expected ErrDrugNotFound, got %v", err)
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		_, err := s.OutboundFIFO("DRUG001", 1000, "内科", "患者B", "李四")
		if !errors.Is(err, ErrInsufficientStock) {
			t.Errorf("expected ErrInsufficientStock, got %v", err)
		}
	})
}

func TestOutboundBlockExpired(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B_EXPIRED", 100,
		now.AddDate(-2, 0, 0), now.AddDate(-1, 0, 0), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_GOOD", 100,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")

	_, err := s.OutboundFIFO("DRUG001", 80, "内科", "患者A", "李四")
	if err != nil {
		t.Fatalf("OutboundFIFO failed: %v", err)
	}

	bExpired, _ := s.GetBatch("B_EXPIRED")
	if bExpired.RemainingQty != 100 {
		t.Errorf("expired batch should not be used, remaining should be 100, got %d", bExpired.RemainingQty)
	}

	bGood, _ := s.GetBatch("B_GOOD")
	if bGood.RemainingQty != 20 {
		t.Errorf("good batch should have 20 remaining, got %d", bGood.RemainingQty)
	}
}

func TestOutboundWithSameExpiry(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	expDate := now.AddDate(1, 0, 0)

	_ = s.InboundBatch("DRUG001", "B1", 50,
		now.AddDate(0, -2, 0), expDate, "供应商A", "张三")
	time.Sleep(10 * time.Millisecond)
	_ = s.InboundBatch("DRUG001", "B2", 50,
		now.AddDate(0, -1, 0), expDate, "供应商A", "张三")

	result, err := s.OutboundFIFO("DRUG001", 70, "内科", "患者A", "李四")
	if err != nil {
		t.Fatalf("OutboundFIFO failed: %v", err)
	}
	if len(result.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(result.Details))
	}
	if result.Details[0].BatchNumber != "B1" {
		t.Errorf("earlier inbound batch should be used first, got %s", result.Details[0].BatchNumber)
	}
}

func TestRecallBatch(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")

	t.Run("success", func(t *testing.T) {
		err := s.RecallBatch("B1", "发现质量问题", "管理员")
		if err != nil {
			t.Fatalf("RecallBatch failed: %v", err)
		}

		batch, _ := s.GetBatch("B1")
		if batch.Status != BatchStatusRecalled {
			t.Errorf("expected status Recalled, got %v", batch.Status)
		}
		if batch.RecallReason != "发现质量问题" {
			t.Errorf("unexpected recall reason: %s", batch.RecallReason)
		}
		if batch.RecallTime == nil {
			t.Error("recall time should be set")
		}

		flows := s.GetStockFlows()
		if len(flows) != 2 {
			t.Errorf("expected 2 flows, got %d", len(flows))
		}
		if flows[1].FlowType != "recall" {
			t.Errorf("expected flow type recall, got %s", flows[1].FlowType)
		}
	})

	t.Run("batch not found", func(t *testing.T) {
		err := s.RecallBatch("NOTEXIST", "reason", "admin")
		if !errors.Is(err, ErrBatchNotFound) {
			t.Errorf("expected ErrBatchNotFound, got %v", err)
		}
	})

	t.Run("empty batch number", func(t *testing.T) {
		err := s.RecallBatch("", "reason", "admin")
		if !errors.Is(err, ErrInvalidBatchNumber) {
			t.Errorf("expected ErrInvalidBatchNumber, got %v", err)
		}
	})

	t.Run("idempotent recall", func(t *testing.T) {
		err := s.RecallBatch("B1", "reason", "admin")
		if err != nil {
			t.Errorf("recall should be idempotent, got error: %v", err)
		}
	})
}

func TestRecallBlocksOutbound(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B2", 50,
		now.AddDate(0, -1, 0), now.AddDate(0, 0, 60), "供应商A", "张三")

	_ = s.RecallBatch("B1", "质量问题", "管理员")

	_, err := s.OutboundFIFO("DRUG001", 40, "内科", "患者A", "李四")
	if err != nil {
		t.Fatalf("OutboundFIFO failed: %v", err)
	}

	b1, _ := s.GetBatch("B1")
	if b1.RemainingQty != 100 {
		t.Errorf("recalled batch should not be used, remaining should be 100, got %d", b1.RemainingQty)
	}
	b2, _ := s.GetBatch("B2")
	if b2.RemainingQty != 10 {
		t.Errorf("non-recalled batch should have 10 remaining, got %d", b2.RemainingQty)
	}
}

func TestRecallOnlyNormalAvailable(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")

	_ = s.RecallBatch("B1", "质量问题", "管理员")

	_, err := s.OutboundFIFO("DRUG001", 10, "内科", "患者A", "李四")
	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock when all batches recalled, got %v", err)
	}
}

func TestGetBatchFlowTrace(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(0, -1, 0), now.AddDate(0, 0, 30), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B2", 100,
		now.AddDate(0, -1, 0), now.AddDate(0, 0, 60), "供应商A", "张三")

	_, _ = s.OutboundFIFO("DRUG001", 100, "内科", "患者A", "李四")
	time.Sleep(10 * time.Millisecond)
	_, _ = s.OutboundFIFO("DRUG001", 40, "外科", "患者B", "王五")

	t.Run("success", func(t *testing.T) {
		trace, err := s.GetBatchFlowTrace("B1")
		if err != nil {
			t.Fatalf("GetBatchFlowTrace failed: %v", err)
		}
		if len(trace) != 1 {
			t.Errorf("expected 1 trace record for B1, got %d", len(trace))
		}
		if trace[0].Quantity != 100 || trace[0].Patient != "患者A" {
			t.Errorf("unexpected trace data: %+v", trace[0])
		}
	})

	t.Run("batch not found", func(t *testing.T) {
		_, err := s.GetBatchFlowTrace("NOTEXIST")
		if !errors.Is(err, ErrBatchNotFound) {
			t.Errorf("expected ErrBatchNotFound, got %v", err)
		}
	})

	t.Run("empty batch number", func(t *testing.T) {
		_, err := s.GetBatchFlowTrace("")
		if !errors.Is(err, ErrInvalidBatchNumber) {
			t.Errorf("expected ErrInvalidBatchNumber, got %v", err)
		}
	})

	t.Run("recalled batch trace", func(t *testing.T) {
		_ = s.RecallBatch("B1", "质量问题", "管理员")
		trace, err := s.GetBatchFlowTrace("B1")
		if err != nil {
			t.Fatalf("GetBatchFlowTrace failed: %v", err)
		}
		if len(trace) != 1 {
			t.Errorf("recalled batch should still have trace, got %d", len(trace))
		}
	})
}

func TestGetDrugBatches(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B2", 100,
		now.AddDate(0, -1, 0), now.AddDate(0, 0, 60), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(0, -1, 0), now.AddDate(0, 0, 30), "供应商A", "张三")

	t.Run("sorted by expiry", func(t *testing.T) {
		batches, err := s.GetDrugBatches("DRUG001")
		if err != nil {
			t.Fatalf("GetDrugBatches failed: %v", err)
		}
		if len(batches) != 2 {
			t.Errorf("expected 2 batches, got %d", len(batches))
		}
		if batches[0].BatchNumber != "B1" || batches[1].BatchNumber != "B2" {
			t.Errorf("batches should be sorted by expiry date")
		}
	})

	t.Run("drug not found", func(t *testing.T) {
		_, err := s.GetDrugBatches("INVALID")
		if !errors.Is(err, ErrDrugNotFound) {
			t.Errorf("expected ErrDrugNotFound, got %v", err)
		}
	})
}

func TestGetDrugStock(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")
	_ = s.AddDrug("DRUG002", "布洛芬缓释胶囊", "0.3g*20粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B2", 50,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")

	t.Run("normal stock", func(t *testing.T) {
		stock, err := s.GetDrugStock("DRUG001")
		if err != nil {
			t.Fatalf("GetDrugStock failed: %v", err)
		}
		if stock != 150 {
			t.Errorf("expected stock 150, got %d", stock)
		}
	})

	t.Run("recalled excluded", func(t *testing.T) {
		_ = s.RecallBatch("B1", "reason", "admin")
		stock, err := s.GetDrugStock("DRUG001")
		if err != nil {
			t.Fatalf("GetDrugStock failed: %v", err)
		}
		if stock != 50 {
			t.Errorf("expected stock 50 after recall, got %d", stock)
		}
	})

	t.Run("no batches", func(t *testing.T) {
		stock, err := s.GetDrugStock("DRUG002")
		if err != nil {
			t.Fatalf("GetDrugStock failed: %v", err)
		}
		if stock != 0 {
			t.Errorf("expected stock 0, got %d", stock)
		}
	})

	t.Run("drug not found", func(t *testing.T) {
		_, err := s.GetDrugStock("INVALID")
		if !errors.Is(err, ErrDrugNotFound) {
			t.Errorf("expected ErrDrugNotFound, got %v", err)
		}
	})
}

func TestUpdateBatchStatus(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	prodDate := now.AddDate(-1, 0, 0)
	expDate := now.AddDate(-1, 0, 0).Add(time.Hour)

	_ = s.InboundBatch("DRUG001", "B1", 100, prodDate, now.AddDate(1, 0, 0), "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B2", 100, prodDate, expDate, "供应商A", "张三")

	b1, _ := s.GetBatch("B1")
	if b1.Status != BatchStatusNormal {
		t.Errorf("B1 should be normal, got %v", b1.Status)
	}

	s.UpdateBatchStatus()

	b1, _ = s.GetBatch("B1")
	if b1.Status != BatchStatusNormal {
		t.Errorf("B1 should still be normal, got %v", b1.Status)
	}

	b2, _ := s.GetBatch("B2")
	if b2.Status != BatchStatusExpired {
		t.Errorf("B2 should be expired, got %v", b2.Status)
	}
}

func TestGetOutboundDetails(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	now := time.Now()
	_ = s.InboundBatch("DRUG001", "B1", 100,
		now.AddDate(0, -1, 0), now.AddDate(1, 0, 0), "供应商A", "张三")

	_, _ = s.OutboundFIFO("DRUG001", 30, "内科", "患者A", "李四")
	_, _ = s.OutboundFIFO("DRUG001", 20, "外科", "患者B", "王五")

	details := s.GetOutboundDetails()
	if len(details) != 2 {
		t.Errorf("expected 2 outbound details, got %d", len(details))
	}
	if details[0].Patient != "患者A" || details[1].Patient != "患者B" {
		t.Errorf("unexpected outbound details order")
	}
}

func TestFullWorkflow(t *testing.T) {
	s := NewDrugTraceService()

	err := s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	err = s.InboundBatch("DRUG001", "B20240101", 500,
		now.AddDate(0, -3, 0), now.AddDate(0, 0, 30), "华北制药", "库管员A")
	if err != nil {
		t.Fatal(err)
	}
	err = s.InboundBatch("DRUG001", "B20240201", 300,
		now.AddDate(0, -2, 0), now.AddDate(0, 0, 60), "华北制药", "库管员A")
	if err != nil {
		t.Fatal(err)
	}

	expiring, err := s.GetExpiringBatches(30)
	if err != nil {
		t.Fatal(err)
	}
	if len(expiring) != 1 || expiring[0].BatchNumber != "B20240101" {
		t.Errorf("expected B20240101 in expiring list")
	}

	result, err := s.OutboundFIFO("DRUG001", 600, "内科", "张小明", "药师李")
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalQty != 600 {
		t.Errorf("expected total 600, got %d", result.TotalQty)
	}
	if len(result.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(result.Details))
	}

	err = s.RecallBatch("B20240201", "检测出杂质超标", "质量管理员")
	if err != nil {
		t.Fatal(err)
	}

	trace, err := s.GetBatchFlowTrace("B20240201")
	if err != nil {
		t.Fatal(err)
	}
	if len(trace) != 1 {
		t.Errorf("expected 1 trace record, got %d", len(trace))
	}
	if trace[0].Patient != "张小明" || trace[0].Quantity != 100 {
		t.Errorf("unexpected trace data: %+v", trace[0])
	}

	_, err = s.OutboundFIFO("DRUG001", 10, "内科", "新患者", "药师李")
	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("expected insufficient stock after recall, got %v", err)
	}
}

func TestBatchStatusString(t *testing.T) {
	tests := []struct {
		status BatchStatus
		want   string
	}{
		{BatchStatusNormal, "normal"},
		{BatchStatusExpired, "expired"},
		{BatchStatusRecalled, "recalled"},
		{BatchStatus(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("BatchStatus(%d).String() = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestDateOnlyHelper(t *testing.T) {
	t1 := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.Local)
	result := dateOnly(t1)
	expected := time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)
	if !result.Equal(expected) {
		t.Errorf("dateOnly() = %v, want %v", result, expected)
	}
}

func TestIsDateExpired(t *testing.T) {
	today := time.Date(2024, 6, 15, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		expiry   time.Time
		asOf     time.Time
		expected bool
	}{
		{
			name:     "expiry before asOf date - expired",
			expiry:   time.Date(2024, 6, 14, 23, 59, 59, 999999999, time.Local),
			asOf:     today,
			expected: true,
		},
		{
			name:     "expiry same day as asOf - not expired",
			expiry:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local),
			asOf:     today,
			expected: false,
		},
		{
			name:     "expiry same day different time - not expired",
			expiry:   time.Date(2024, 6, 15, 23, 59, 59, 999999999, time.Local),
			asOf:     today,
			expected: false,
		},
		{
			name:     "expiry day after asOf - not expired",
			expiry:   time.Date(2024, 6, 16, 0, 0, 0, 0, time.Local),
			asOf:     today,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateExpired(tt.expiry, tt.asOf)
			if result != tt.expected {
				t.Errorf("isDateExpired(%v, %v) = %v, want %v", tt.expiry, tt.asOf, result, tt.expected)
			}
		})
	}
}

func TestIsDateWithinDays(t *testing.T) {
	today := time.Date(2024, 6, 15, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		expiry   time.Time
		days     int
		expected bool
	}{
		{
			name:     "expiry today - within 0 days",
			expiry:   time.Date(2024, 6, 15, 23, 59, 59, 999999999, time.Local),
			days:     0,
			expected: true,
		},
		{
			name:     "expiry in 30 days - within 30 days",
			expiry:   time.Date(2024, 7, 15, 0, 0, 0, 0, time.Local),
			days:     30,
			expected: true,
		},
		{
			name:     "expiry in 31 days - not within 30 days",
			expiry:   time.Date(2024, 7, 16, 0, 0, 0, 0, time.Local),
			days:     30,
			expected: false,
		},
		{
			name:     "expiry yesterday - not within 30 days (already expired)",
			expiry:   time.Date(2024, 6, 14, 23, 59, 59, 999999999, time.Local),
			days:     30,
			expected: false,
		},
		{
			name:     "expiry long ago - not within 30 days (already expired)",
			expiry:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			days:     30,
			expected: false,
		},
		{
			name:     "expiry day after tomorrow - within 30 days",
			expiry:   time.Date(2024, 6, 17, 0, 0, 0, 0, time.Local),
			days:     30,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateWithinDays(tt.expiry, today, tt.days)
			if result != tt.expected {
				t.Errorf("isDateWithinDays(%v, today, %d) = %v, want %v", tt.expiry, tt.days, result, tt.expected)
			}
		})
	}
}

func TestOutboundSameDayExpiry(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	todayEvening := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())
	tomorrow := today.AddDate(0, 0, 1)

	_ = s.InboundBatch("DRUG001", "B_TODAY", 100,
		today.AddDate(-1, 0, 0), todayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_TOMORROW", 100,
		today.AddDate(-1, 0, 0), tomorrow, "供应商A", "张三")

	batch, err := s.GetBatch("B_TODAY")
	if err != nil {
		t.Fatalf("GetBatch failed: %v", err)
	}
	if batch.Status != BatchStatusNormal {
		t.Errorf("batch expiring today should be Normal, got %v", batch.Status)
	}

	result, err := s.OutboundFIFO("DRUG001", 80, "内科", "患者A", "李四")
	if err != nil {
		t.Fatalf("OutboundFIFO failed: %v", err)
	}
	if result.TotalQty != 80 {
		t.Errorf("expected 80, got %d", result.TotalQty)
	}
	if result.Details[0].BatchNumber != "B_TODAY" {
		t.Errorf("should use today's expiry batch first, got %s", result.Details[0].BatchNumber)
	}

	bToday, _ := s.GetBatch("B_TODAY")
	if bToday.RemainingQty != 20 {
		t.Errorf("expected 20 remaining, got %d", bToday.RemainingQty)
	}
}

func TestOutboundDayAfterExpiry(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	yesterdayEvening := time.Date(today.Year(), today.Month(), today.Day()-1, 23, 59, 59, 999999999, today.Location())
	tomorrow := today.AddDate(0, 0, 1)

	_ = s.InboundBatch("DRUG001", "B_YESTERDAY", 100,
		today.AddDate(-1, 0, 0), yesterdayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_TOMORROW", 100,
		today.AddDate(-1, 0, 0), tomorrow, "供应商A", "张三")

	batch, err := s.GetBatch("B_YESTERDAY")
	if err != nil {
		t.Fatalf("GetBatch failed: %v", err)
	}
	if batch.Status != BatchStatusExpired {
		t.Errorf("batch expired yesterday should be Expired, got %v", batch.Status)
	}

	result, err := s.OutboundFIFO("DRUG001", 80, "内科", "患者A", "李四")
	if err != nil {
		t.Fatalf("OutboundFIFO failed: %v", err)
	}
	if result.TotalQty != 80 {
		t.Errorf("expected 80, got %d", result.TotalQty)
	}
	if result.Details[0].BatchNumber != "B_TOMORROW" {
		t.Errorf("should use tomorrow's expiry batch, got %s", result.Details[0].BatchNumber)
	}

	bYesterday, _ := s.GetBatch("B_YESTERDAY")
	if bYesterday.RemainingQty != 100 {
		t.Errorf("expired batch should not be used, remaining should be 100, got %d", bYesterday.RemainingQty)
	}
}

func TestCrossMidnightExpiryBoundary(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	expiryDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	_ = s.InboundBatch("DRUG001", "B1", 100,
		today.AddDate(-1, 0, 0), expiryDate, "供应商A", "张三")

	batch, err := s.GetBatch("B1")
	if err != nil {
		t.Fatalf("GetBatch failed: %v", err)
	}
	if batch.Status != BatchStatusNormal {
		t.Errorf("batch expiring at midnight today should be Normal, got %v", batch.Status)
	}

	result, err := s.OutboundFIFO("DRUG001", 50, "内科", "患者A", "李四")
	if err != nil {
		t.Fatalf("OutboundFIFO failed: %v", err)
	}
	if result.TotalQty != 50 {
		t.Errorf("expected 50, got %d", result.TotalQty)
	}

	stock, _ := s.GetDrugStock("DRUG001")
	if stock != 50 {
		t.Errorf("expected 50 stock remaining, got %d", stock)
	}
}

func TestUpdateBatchStatusDatePrecision(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	todayEvening := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())
	yesterdayEvening := time.Date(today.Year(), today.Month(), today.Day()-1, 23, 59, 59, 999999999, today.Location())
	tomorrow := today.AddDate(0, 0, 1)

	_ = s.InboundBatch("DRUG001", "B_TODAY", 100,
		today.AddDate(-1, 0, 0), todayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_YESTERDAY", 100,
		today.AddDate(-2, 0, 0), yesterdayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_TOMORROW", 100,
		today.AddDate(-1, 0, 0), tomorrow, "供应商A", "张三")

	bToday, _ := s.GetBatch("B_TODAY")
	bToday.Status = BatchStatusNormal
	bYesterday, _ := s.GetBatch("B_YESTERDAY")
	bYesterday.Status = BatchStatusNormal

	s.UpdateBatchStatus()

	bToday, _ = s.GetBatch("B_TODAY")
	if bToday.Status != BatchStatusNormal {
		t.Errorf("B_TODAY should still be Normal after UpdateBatchStatus, got %v", bToday.Status)
	}

	bYesterday, _ = s.GetBatch("B_YESTERDAY")
	if bYesterday.Status != BatchStatusExpired {
		t.Errorf("B_YESTERDAY should be Expired after UpdateBatchStatus, got %v", bYesterday.Status)
	}

	bTomorrow, _ := s.GetBatch("B_TOMORROW")
	if bTomorrow.Status != BatchStatusNormal {
		t.Errorf("B_TOMORROW should be Normal, got %v", bTomorrow.Status)
	}
}

func TestGetExpiredBatchesDatePrecision(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	todayEvening := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())
	yesterdayEvening := time.Date(today.Year(), today.Month(), today.Day()-1, 23, 59, 59, 999999999, today.Location())

	_ = s.InboundBatch("DRUG001", "B_TODAY", 100,
		today.AddDate(-1, 0, 0), todayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_YESTERDAY", 100,
		today.AddDate(-2, 0, 0), yesterdayEvening, "供应商A", "张三")

	bToday, _ := s.GetBatch("B_TODAY")
	bToday.Status = BatchStatusNormal

	result, err := s.GetExpiredBatches()
	if err != nil {
		t.Fatalf("GetExpiredBatches failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 expired batch, got %d", len(result))
	}
	if result[0].BatchNumber != "B_YESTERDAY" {
		t.Errorf("expected B_YESTERDAY, got %s", result[0].BatchNumber)
	}
}

func TestGetExpiringBatchesDatePrecision(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	todayEvening := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())
	in30Days := today.AddDate(0, 0, 30)
	in31Days := today.AddDate(0, 0, 31)

	_ = s.InboundBatch("DRUG001", "B_TODAY", 100,
		today.AddDate(-1, 0, 0), todayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_30DAYS", 100,
		today.AddDate(-1, 0, 0), in30Days, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_31DAYS", 100,
		today.AddDate(-1, 0, 0), in31Days, "供应商A", "张三")

	result, err := s.GetExpiringBatches(30)
	if err != nil {
		t.Fatalf("GetExpiringBatches failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 expiring batches, got %d", len(result))
	}
}

func TestOutboundConsistentTimeBase(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	todayEvening := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())
	tomorrow := today.AddDate(0, 0, 1)

	for i := 0; i < 20; i++ {
		batchNo := fmt.Sprintf("B_%d", i)
		expiry := todayEvening
		if i%2 == 0 {
			expiry = tomorrow
		}
		_ = s.InboundBatch("DRUG001", batchNo, 10,
			today.AddDate(-1, 0, 0), expiry, "供应商A", "张三")
	}

	result, err := s.OutboundFIFO("DRUG001", 150, "内科", "患者A", "李四")
	if err != nil {
		t.Fatalf("OutboundFIFO failed: %v", err)
	}
	if result.TotalQty != 150 {
		t.Errorf("expected 150, got %d", result.TotalQty)
	}

	for _, detail := range result.Details {
		if !detail.OutboundTime.Equal(result.Details[0].OutboundTime) {
			t.Errorf("all outbound details should have the same timestamp, got %v vs %v",
				detail.OutboundTime, result.Details[0].OutboundTime)
		}
	}
}

func TestGetExpiringBatchesExcludesExpiredStatus(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	todayEvening := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())
	yesterdayEvening := time.Date(today.Year(), today.Month(), today.Day()-1, 23, 59, 59, 999999999, today.Location())
	in7Days := today.AddDate(0, 0, 7)

	_ = s.InboundBatch("DRUG001", "B_EXPIRED", 100,
		today.AddDate(-1, 0, 0), yesterdayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_NORMAL", 100,
		today.AddDate(-1, 0, 0), in7Days, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_TODAY", 100,
		today.AddDate(-1, 0, 0), todayEvening, "供应商A", "张三")

	expiring, err := s.GetExpiringBatches(30)
	if err != nil {
		t.Fatalf("GetExpiringBatches failed: %v", err)
	}
	if len(expiring) != 2 {
		t.Errorf("expected 2 expiring batches, got %d", len(expiring))
	}

	batchNos := make(map[string]bool)
	for _, b := range expiring {
		batchNos[b.BatchNumber] = true
	}
	if batchNos["B_EXPIRED"] {
		t.Error("expired status batch should not be in expiring list")
	}
	if !batchNos["B_NORMAL"] || !batchNos["B_TODAY"] {
		t.Error("normal batches should be in expiring list")
	}
}

func TestGetExpiringBatchesExcludesLongExpiredNotUpdated(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	longAgoExpiry := today.AddDate(-1, 0, 0)
	in7Days := today.AddDate(0, 0, 7)

	_ = s.InboundBatch("DRUG001", "B_LONG_AGO", 100,
		today.AddDate(-2, 0, 0), longAgoExpiry, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_NORMAL", 100,
		today.AddDate(-1, 0, 0), in7Days, "供应商A", "张三")

	bLongAgo, _ := s.GetBatch("B_LONG_AGO")
	bLongAgo.Status = BatchStatusNormal

	expiring, err := s.GetExpiringBatches(30)
	if err != nil {
		t.Fatalf("GetExpiringBatches failed: %v", err)
	}

	for _, b := range expiring {
		if b.BatchNumber == "B_LONG_AGO" {
			t.Error("long expired batch should not be in expiring list")
		}
	}

	if len(expiring) != 1 || expiring[0].BatchNumber != "B_NORMAL" {
		t.Errorf("expected only B_NORMAL, got %v", expiring)
	}
}

func TestBatchNotInBothExpiredAndExpiringLists(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	yesterdayEvening := time.Date(today.Year(), today.Month(), today.Day()-1, 23, 59, 59, 999999999, today.Location())
	in3Days := today.AddDate(0, 0, 3)
	in7Days := today.AddDate(0, 0, 7)

	_ = s.InboundBatch("DRUG001", "B_YESTERDAY", 100,
		today.AddDate(-1, 0, 0), yesterdayEvening, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_SOON", 100,
		today.AddDate(-1, 0, 0), in3Days, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_NORMAL", 100,
		today.AddDate(-1, 0, 0), in7Days, "供应商A", "张三")

	expired, err := s.GetExpiredBatches()
	if err != nil {
		t.Fatalf("GetExpiredBatches failed: %v", err)
	}

	expiring, err := s.GetExpiringBatches(30)
	if err != nil {
		t.Fatalf("GetExpiringBatches failed: %v", err)
	}

	expiredMap := make(map[string]bool)
	for _, b := range expired {
		expiredMap[b.BatchNumber] = true
	}

	for _, b := range expiring {
		if expiredMap[b.BatchNumber] {
			t.Errorf("batch %s should not be in both expired and expiring lists", b.BatchNumber)
		}
	}

	if len(expired) != 1 || expired[0].BatchNumber != "B_YESTERDAY" {
		t.Errorf("expired list should contain only B_YESTERDAY")
	}

	if len(expiring) != 2 {
		t.Errorf("expiring list should contain B_SOON and B_NORMAL")
	}
}

func TestGetExpiringBatchesExcludesRecalled(t *testing.T) {
	s := NewDrugTraceService()
	_ = s.AddDrug("DRUG001", "阿莫西林胶囊", "0.5g*24粒")

	today := time.Now()
	in7Days := today.AddDate(0, 0, 7)
	in14Days := today.AddDate(0, 0, 14)

	_ = s.InboundBatch("DRUG001", "B_RECALL", 100,
		today.AddDate(-1, 0, 0), in7Days, "供应商A", "张三")
	_ = s.InboundBatch("DRUG001", "B_NORMAL", 100,
		today.AddDate(-1, 0, 0), in14Days, "供应商A", "张三")

	_ = s.RecallBatch("B_RECALL", "质量问题", "管理员")

	expiring, err := s.GetExpiringBatches(30)
	if err != nil {
		t.Fatalf("GetExpiringBatches failed: %v", err)
	}

	if len(expiring) != 1 || expiring[0].BatchNumber != "B_NORMAL" {
		t.Errorf("expected only B_NORMAL, got %v", expiring)
	}
}
