package scheduler

// AutomationExecutor defines the interface for automation execution
type AutomationExecutor interface {
	ProcessDueAutomations()
	ExecuteForConfig(configID int64) error
	ResumePausedCrawlers()
}

// AutomationWorker wraps an executor for scheduler use
type AutomationWorker struct {
	executor AutomationExecutor
}

// NewAutomationWorker creates a new automation worker
func NewAutomationWorker(executor AutomationExecutor) *AutomationWorker {
	return &AutomationWorker{executor: executor}
}

// ProcessDueAutomations delegates to the executor
func (w *AutomationWorker) ProcessDueAutomations() {
	if w.executor == nil {
		return
	}
	w.executor.ProcessDueAutomations()
}

// ExecuteForConfig delegates to the executor (implements TestRunner interface)
func (w *AutomationWorker) ExecuteForConfig(configID int64) error {
	if w.executor == nil {
		return nil
	}
	return w.executor.ExecuteForConfig(configID)
}

// ResumePausedCrawlers delegates to the executor
func (w *AutomationWorker) ResumePausedCrawlers() {
	if w.executor == nil {
		return
	}
	w.executor.ResumePausedCrawlers()
}
