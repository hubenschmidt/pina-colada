package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
)

// Scheduler manages scheduled automation jobs
type Scheduler struct {
	cron          *cron.Cron
	worker        *AutomationWorker
	digestService *DigestService
}

// NewScheduler creates a new scheduler
func NewScheduler(
	worker *AutomationWorker,
	digestService *DigestService,
) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		worker:        worker,
		digestService: digestService,
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() error {
	_, err := s.cron.AddFunc("* * * * *", s.checkDueAutomations)
	if err != nil {
		return err
	}

	_, err = s.cron.AddFunc("0 11 * * *", s.runDailyDigest)
	if err != nil {
		return err
	}

	s.cron.Start()
	log.Println("Scheduler started")
	return nil
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) checkDueAutomations() {
	if s.worker == nil {
		return
	}
	s.worker.ProcessDueAutomations()
}

func (s *Scheduler) runDailyDigest() {
	if s.digestService == nil {
		return
	}
	s.digestService.SendDailyDigests()
}
