# Agent-Go Endpoint Evaluation

**Date:** 2025-12-08
**Python endpoints:** 158
**Go endpoints:** 75
**Missing in Go:** 83

---

# Important - when implementing, please follow /csr-code-rules and /go-code-rules

## Summary by Resource

| Resource              | Python | Go  | Missing |
| --------------------- | ------ | --- | ------- |
| reports               | 14     | 0   | 14      |
| conversations         | 7      | 1   | 6       |
| organizations         | 19     | 6   | 13      |
| documents             | 12     | 9   | 3       |
| projects              | 7      | 1   | 6       |
| usage                 | 5      | 0   | 5       |
| individuals           | 13     | 6   | 7       |
| leads                 | 4      | 0   | 4       |
| notifications         | 4      | 1   | 3       |
| opportunities         | 5      | 2   | 3       |
| partnerships          | 5      | 2   | 3       |
| technologies          | 3      | 0   | 3       |
| preferences           | 5      | 3   | 2       |
| provenance            | 2      | 0   | 2       |
| costs                 | 2      | 0   | 2       |
| industries            | 2      | 1   | 1       |
| auth                  | 2      | 1   | 1       |
| jobs                  | 6      | 5   | 1       |
| tasks                 | 8      | 8   | 0       |
| contacts              | 6      | 6   | 0       |
| users                 | 2      | 2   | 0       |
| notes                 | 5      | 5   | 0       |
| comments              | 5      | 5   | 0       |
| accounts              | 3      | 3   | 0       |
| tags                  | 1      | 1   | 0       |
| salary_ranges         | 1      | 1   | 0       |
| funding_stages        | 1      | 1   | 0       |
| employee_count_ranges | 1      | 1   | 0       |
| revenue_ranges        | 1      | 1   | 0       |

---

## Detailed Missing Endpoints

### reports (0/14 in Go)

| Method | Path                     | Status  |
| ------ | ------------------------ | ------- |
| GET    | /canned/lead-pipeline    | Missing |
| GET    | /canned/account-overview | Missing |
| GET    | /canned/contact-coverage | Missing |
| GET    | /canned/notes-activity   | Missing |
| GET    | /canned/user-audit       | Missing |
| GET    | /fields/{entity}         | Missing |
| POST   | /custom/preview          | Missing |
| POST   | /custom/run              | Missing |
| POST   | /custom/export           | Missing |
| GET    | /saved                   | Missing |
| POST   | /saved                   | Missing |
| GET    | /saved/{report_id}       | Missing |
| PUT    | /saved/{report_id}       | Missing |
| DELETE | /saved/{report_id}       | Missing |

### conversations (1/7 in Go)

| Method | Path                   | Status  |
| ------ | ---------------------- | ------- |
| GET    | /                      | Done    |
| GET    | /all                   | Missing |
| GET    | /{thread_id}           | Missing |
| PATCH  | /{thread_id}           | Missing |
| DELETE | /{thread_id}           | Missing |
| POST   | /{thread_id}/unarchive | Missing |
| DELETE | /{thread_id}/permanent | Missing |

### organizations (6/19 in Go)

| Method | Path                            | Status  |
| ------ | ------------------------------- | ------- |
| GET    | /                               | Done    |
| GET    | /search                         | Done    |
| GET    | /{id}                           | Done    |
| POST   | /                               | Missing |
| PUT    | /{id}                           | Done    |
| DELETE | /{id}                           | Done    |
| GET    | /{id}/contacts                  | Missing |
| POST   | /{id}/contacts                  | Done    |
| PUT    | /{id}/contacts/{contact_id}     | Missing |
| DELETE | /{id}/contacts/{contact_id}     | Missing |
| GET    | /{id}/technologies              | Missing |
| POST   | /{id}/technologies              | Missing |
| DELETE | /{id}/technologies/{tech_id}    | Missing |
| GET    | /{id}/funding-rounds            | Missing |
| POST   | /{id}/funding-rounds            | Missing |
| DELETE | /{id}/funding-rounds/{round_id} | Missing |
| GET    | /{id}/signals                   | Missing |
| POST   | /{id}/signals                   | Missing |
| DELETE | /{id}/signals/{signal_id}       | Missing |

### documents (9/12 in Go)

| Method | Path              | Status  |
| ------ | ----------------- | ------- |
| GET    | /                 | Done    |
| GET    | /check-filename   | Done    |
| GET    | /{id}             | Done    |
| POST   | /                 | Done    |
| GET    | /{id}/download    | Missing |
| PUT    | /{id}             | Missing |
| DELETE | /{id}             | Missing |
| POST   | /{id}/link        | Done    |
| DELETE | /{id}/link        | Done    |
| GET    | /{id}/versions    | Done    |
| POST   | /{id}/versions    | Done    |
| PATCH  | /{id}/set-current | Done    |

### projects (1/7 in Go)

| Method | Path        | Status  |
| ------ | ----------- | ------- |
| GET    | /           | Done    |
| GET    | /{id}       | Missing |
| POST   | /           | Missing |
| PUT    | /{id}       | Missing |
| DELETE | /{id}       | Missing |
| GET    | /{id}/leads | Missing |
| GET    | /{id}/deals | Missing |

### usage (0/5 in Go)

| Method | Path              | Status  |
| ------ | ----------------- | ------- |
| GET    | /user             | Missing |
| GET    | /tenant           | Missing |
| GET    | /timeseries       | Missing |
| GET    | /analytics        | Missing |
| GET    | /developer-access | Missing |

### individuals (6/13 in Go)

| Method | Path                        | Status  |
| ------ | --------------------------- | ------- |
| GET    | /                           | Done    |
| GET    | /search                     | Done    |
| GET    | /{id}                       | Done    |
| POST   | /                           | Missing |
| PUT    | /{id}                       | Done    |
| DELETE | /{id}                       | Done    |
| GET    | /{id}/contacts              | Missing |
| POST   | /{id}/contacts              | Done    |
| PUT    | /{id}/contacts/{contact_id} | Missing |
| DELETE | /{id}/contacts/{contact_id} | Missing |
| GET    | /{id}/signals               | Missing |
| POST   | /{id}/signals               | Missing |
| DELETE | /{id}/signals/{signal_id}   | Missing |

### leads (0/4 in Go)

| Method | Path                   | Status  |
| ------ | ---------------------- | ------- |
| GET    | /                      | Missing |
| GET    | /statuses              | Missing |
| POST   | /{job_id}/apply        | Missing |
| POST   | /{job_id}/do-not-apply | Missing |

### notifications (1/4 in Go)

| Method | Path              | Status  |
| ------ | ----------------- | ------- |
| GET    | /count            | Done    |
| GET    | /                 | Missing |
| POST   | /mark-read        | Missing |
| POST   | /mark-entity-read | Missing |

### opportunities (2/5 in Go)

| Method | Path  | Status  |
| ------ | ----- | ------- |
| GET    | /     | Done    |
| POST   | /     | Missing |
| GET    | /{id} | Done    |
| PUT    | /{id} | Missing |
| DELETE | /{id} | Missing |

### partnerships (2/5 in Go)

| Method | Path  | Status  |
| ------ | ----- | ------- |
| GET    | /     | Done    |
| POST   | /     | Missing |
| GET    | /{id} | Done    |
| PUT    | /{id} | Missing |
| DELETE | /{id} | Missing |

### technologies (0/3 in Go)

| Method | Path  | Status  |
| ------ | ----- | ------- |
| GET    | /     | Missing |
| GET    | /{id} | Missing |
| POST   | /     | Missing |

### preferences (3/5 in Go)

| Method | Path       | Status  |
| ------ | ---------- | ------- |
| GET    | /timezones | Done    |
| GET    | /user      | Done    |
| PATCH  | /user      | Done    |
| GET    | /tenant    | Missing |
| PATCH  | /tenant    | Missing |

### provenance (0/2 in Go)

| Method | Path                       | Status  |
| ------ | -------------------------- | ------- |
| GET    | /{entity_type}/{entity_id} | Missing |
| POST   | /                          | Missing |

### costs (0/2 in Go)

| Method | Path     | Status  |
| ------ | -------- | ------- |
| GET    | /summary | Missing |
| GET    | /org     | Missing |

### industries (1/2 in Go)

| Method | Path | Status  |
| ------ | ---- | ------- |
| GET    | /    | Done    |
| POST   | /    | Missing |

### auth (1/2 in Go)

| Method | Path           | Status  |
| ------ | -------------- | ------- |
| GET    | /me            | Done    |
| POST   | /tenant/create | Missing |

### jobs (5/6 in Go)

| Method | Path                | Status  |
| ------ | ------------------- | ------- |
| GET    | /                   | Done    |
| POST   | /                   | Done    |
| GET    | /recent-resume-date | Missing |
| GET    | /{id}               | Done    |
| PUT    | /{id}               | Done    |
| DELETE | /{id}               | Done    |

---

## Complete in Go (no missing endpoints)

- tasks (8/8)
- contacts (6/6)
- users (2/2)
- notes (5/5)
- comments (5/5)
- accounts (3/3)
- tags (1/1)
- salary_ranges (1/1)
- funding_stages (1/1)
- employee_count_ranges (1/1)
- revenue_ranges (1/1)
