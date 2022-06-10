package benchmark

import (
	"fmt"
	"time"
	"xsyn-services/passport/passlog"
)

type Record struct {
	key       string
	startedAt time.Time
	endedAt   time.Time
}

type Benchmarker struct {
	RecordMap map[string]*Record
}

// New create a new benchmark instance
func New() *Benchmarker {
	bm := &Benchmarker{
		RecordMap: make(map[string]*Record),
	}

	return bm
}

func (bm *Benchmarker) Start(key string) {
	// check record exists
	if _, ok := bm.RecordMap[key]; ok {
		passlog.L.Warn().Msgf(`Benchmark on key "%s" has been override`, key)
		return
	}

	// store record
	bm.RecordMap[key] = &Record{
		key:       key,
		startedAt: time.Now(),
	}
}

func (bm *Benchmarker) End(key string) {
	now := time.Now()

	r, ok := bm.RecordMap[key]
	if !ok {
		passlog.L.Warn().Msgf(`Benchmark on key "%s" does not exists`, key)
		return
	}

	r.endedAt = now
}

func (bm *Benchmarker) ReportGet() (time.Duration, []string, error) {
	if len(bm.RecordMap) == 0 {
		passlog.L.Warn().Msg("There is no benchmark record")
		return 0, nil, fmt.Errorf("benchmark record not found")
	}

	// calculate duration
	var totalTime time.Duration
	var reports []string

	for key, record := range bm.RecordMap {
		if record.startedAt.After(record.endedAt) {
			passlog.L.Warn().Msgf(`The end time of key "%s" is not set`, key)
			return 0, nil, fmt.Errorf("invalid benchmark record")
		}
		duration := record.endedAt.Sub(record.startedAt)
		reports = append(reports, fmt.Sprintf(`%s: %d ms`, key, duration.Milliseconds()))

		totalTime += duration
	}

	return totalTime, reports, nil
}

func (bm *Benchmarker) Report() {
	if len(bm.RecordMap) == 0 {
		passlog.L.Warn().Msg("There is no benchmark record to report")
		return
	}

	// calculate duration
	totalTime, reports, err := bm.ReportGet()
	if err != nil {
		passlog.L.Warn().Err(err).Msg("Failed to get benchmark report")
		return
	}

	// log report if alert baseline is not set or total time exceed alert baseline
	passlog.L.Info().Interface("report", reports).Int64("total ms", totalTime.Milliseconds()).Msg("Benchmark report")
}

func (bm *Benchmarker) Alert(millisecond int64) {
	if len(bm.RecordMap) == 0 {
		passlog.L.Warn().Msg("There is no benchmark record to alert")
		return
	}

	totalTime, reports, err := bm.ReportGet()
	if err != nil {
		passlog.L.Warn().Err(err).Msg("Failed to get benchmark report")
		return
	}

	if totalTime.Milliseconds() > millisecond {
		passlog.L.Warn().Interface("report", reports).Int64("total ms", totalTime.Milliseconds()).Int64("required ms", millisecond).Msg("Exceed required time")
	}
}
