# Python Code Rules Violations Report
# Generated for modules/agent directory

## Summary
- Total violations found: 26
- Files analyzed: 123

### Violations by Type

- **complex_function**: 7
- **continue_statement**: 4
- **guard_clause**: 1
- **long_function**: 6
- **nested_conditional**: 8

## Detailed Violations

### Complex Function

#### agent/src/agent/graph.py

- **Line 68**: KISS - Keep It Simple
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

#### agent/src/lib/serialization.py

- **Line 10**: KISS - Keep It Simple
  ```python
  def model_to_dict
  ```
  - Note: Function has 6 conditionals - consider breaking into smaller functions

#### agent/src/services/report_builder.py

- **Line 150**: KISS - Keep It Simple
  ```python
  def _extract_field_value
  ```
  - Note: Function has 8 conditionals - consider breaking into smaller functions

### Continue Statement

#### agent/src/services/report_builder.py

- **Line 224**: Avoid continue; refactor with small pure functions, early returns
  ```python
  continue
  ```

- **Line 226**: Avoid continue; refactor with small pure functions, early returns
  ```python
  continue  # Skip join field filters for now
  ```

- **Line 228**: Avoid continue; refactor with small pure functions, early returns
  ```python
  continue
  ```

- **Line 252**: Avoid continue; refactor with small pure functions, early returns
  ```python
  continue
  ```

### Guard Clause

#### agent/src/api/routes/accounts.py

- **Line 26**: Avoid else statements; use guard clauses
  ```python
  elif has_ind:
  ```

### Long Function

#### agent/src/agent/graph.py

- **Line 68**: KISS - Keep It Simple
  ```python
  def make_websocket_stream_adapter
  ```
  - Note: Function is 83 lines - consider breaking into smaller functions

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

#### agent/src/api/routes/individuals.py

- **Line 141**: KISS - Keep It Simple
  ```python
  def _ind_to_dict
  ```
  - Note: Function is 56 lines - consider breaking into smaller functions

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

#### agent/src/agent/graph.py

- **Line 282**: Avoid nested conditionals; use guard clauses
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

#### agent/src/services/report_builder.py

- **Line 169**: Avoid nested conditionals; use guard clauses
  ```python
  if hasattr(val, "isoformat"):
  ```

- **Line 268**: Avoid nested conditionals; use guard clauses
  ```python
  if entity_type:
  ```
