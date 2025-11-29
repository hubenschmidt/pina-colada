# Python Code Rules Violations Report
# Generated for modules/agent directory

## Summary
- Total violations found: 29
- Files analyzed: 119

### Violations by Type

- **complex_function**: 10
- **guard_clause**: 1
- **long_function**: 10
- **nested_conditional**: 8

## Detailed Violations

### Complex Function

#### agent/src/agent/evaluators/_base_evaluator.py

- **Line 96**: KISS - Keep It Simple
  ```python
  def evaluator_node
  ```
  - Note: Function has 7 conditionals - consider breaking into smaller functions

#### agent/src/agent/graph.py

- **Line 56**: KISS - Keep It Simple
  ```python
  def make_websocket_stream_adapter
  ```
  - Note: Function has 8 conditionals - consider breaking into smaller functions

#### agent/src/agent/orchestrator.py

- **Line 77**: KISS - Keep It Simple
  ```python
  def _ensure_tool_pairs_intact
  ```
  - Note: Function has 6 conditionals - consider breaking into smaller functions

#### agent/src/agent/tools/worker_tools.py

- **Line 353**: KISS - Keep It Simple
  ```python
  def _parse_tool_input
  ```
  - Note: Function has 6 conditionals - consider breaking into smaller functions

- **Line 390**: KISS - Keep It Simple
  ```python
  def check_applied_jobs
  ```
  - Note: Function has 6 conditionals - consider breaking into smaller functions

- **Line 575**: KISS - Keep It Simple
  ```python
  def _matches_applied_job
  ```
  - Note: Function has 6 conditionals - consider breaking into smaller functions

#### agent/src/agent/util/get_success_criteria.py

- **Line 1**: KISS - Keep It Simple
  ```python
  def get_success_criteria
  ```
  - Note: Function has 11 conditionals - consider breaking into smaller functions

#### agent/src/controllers/job_controller.py

- **Line 46**: KISS - Keep It Simple
  ```python
  def _job_to_response_dict
  ```
  - Note: Function has 9 conditionals - consider breaking into smaller functions

#### agent/src/lib/auth.py

- **Line 80**: KISS - Keep It Simple
  ```python
  def require_auth
  ```
  - Note: Function has 8 conditionals - consider breaking into smaller functions

#### agent/src/lib/serialization.py

- **Line 10**: KISS - Keep It Simple
  ```python
  def model_to_dict
  ```
  - Note: Function has 6 conditionals - consider breaking into smaller functions

### Guard Clause

#### agent/src/api/routes/accounts.py

- **Line 26**: Avoid else statements; use guard clauses
  ```python
  elif has_ind:
  ```

### Long Function

#### agent/src/agent/evaluators/_base_evaluator.py

- **Line 96**: KISS - Keep It Simple
  ```python
  def evaluator_node
  ```
  - Note: Function is 104 lines - consider breaking into smaller functions

#### agent/src/agent/graph.py

- **Line 56**: KISS - Keep It Simple
  ```python
  def make_websocket_stream_adapter
  ```
  - Note: Function is 97 lines - consider breaking into smaller functions

#### agent/src/agent/tools/worker_tools.py

- **Line 390**: KISS - Keep It Simple
  ```python
  def check_applied_jobs
  ```
  - Note: Function is 56 lines - consider breaking into smaller functions

- **Line 719**: KISS - Keep It Simple
  ```python
  def update_job_status
  ```
  - Note: Function is 66 lines - consider breaking into smaller functions

#### agent/src/agent/util/get_success_criteria.py

- **Line 1**: KISS - Keep It Simple
  ```python
  def get_success_criteria
  ```
  - Note: Function is 186 lines - consider breaking into smaller functions

#### agent/src/api/routes/individuals.py

- **Line 141**: KISS - Keep It Simple
  ```python
  def _ind_to_dict
  ```
  - Note: Function is 56 lines - consider breaking into smaller functions

#### agent/src/controllers/job_controller.py

- **Line 46**: KISS - Keep It Simple
  ```python
  def _job_to_response_dict
  ```
  - Note: Function is 87 lines - consider breaking into smaller functions

#### agent/src/lib/auth.py

- **Line 80**: KISS - Keep It Simple
  ```python
  def require_auth
  ```
  - Note: Function is 59 lines - consider breaking into smaller functions

#### agent/src/lib/serialization.py

- **Line 10**: KISS - Keep It Simple
  ```python
  def model_to_dict
  ```
  - Note: Function is 76 lines - consider breaking into smaller functions

#### agent/src/services/supabase_client.py

- **Line 78**: KISS - Keep It Simple
  ```python
  def fetch_applied_jobs
  ```
  - Note: Function is 53 lines - consider breaking into smaller functions

### Nested Conditional

#### agent/src/agent/evaluators/_base_evaluator.py

- **Line 165**: Avoid nested conditionals; use guard clauses
  ```python
  if eval_result.score < 60:
  ```

#### agent/src/agent/graph.py

- **Line 294**: Avoid nested conditionals; use guard clauses
  ```python
  if _is_disconnect_error(send_err):
  ```

#### agent/src/agent/orchestrator.py

- **Line 98**: Avoid nested conditionals; use guard clauses
  ```python
  if orphaned:
  ```

#### agent/src/agent/tools/worker_tools.py

- **Line 375**: Avoid nested conditionals; use guard clauses
  ```python
  if len(parts) > 1:
  ```

- **Line 377**: Avoid nested conditionals; use guard clauses
  ```python
  if potential_title:
  ```

#### agent/src/controllers/job_controller.py

- **Line 63**: Avoid nested conditionals; use guard clauses
  ```python
  elif job.lead.account.individuals:
  ```

#### agent/src/repositories/contact_repository.py

- **Line 151**: Avoid nested conditionals; use guard clauses
  ```python
  if existing.is_primary != is_primary:
  ```

#### agent/src/repositories/job_repository.py

- **Line 261**: Avoid nested conditionals; use guard clauses
  ```python
  if org and org.account_id:
  ```
