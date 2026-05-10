# Trustwired Internal System — Full Technical Requirements

> Complete specification for rebuilding this system in any language (originally built in Laravel 11 + Astro 5 + Vue 3).  
> Backend: **Go** with air · Frontend: **Astro + Vue 3** · Database: **MySQL 8**

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [User Roles & Permissions](#2-user-roles--permissions)
3. [Database Schema](#3-database-schema)
4. [API Endpoints](#4-api-endpoints)
5. [Business Logic Rules](#5-business-logic-rules)
6. [File Storage](#6-file-storage)
7. [Email Notifications](#7-email-notifications)
8. [SLA & Indicator Rules](#8-sla--indicator-rules)
9. [Invoice & Commission Rules](#9-invoice--commission-rules)
10. [Frontend Pages & Components](#10-frontend-pages--components)

---

## 1. System Overview

Trustwired is an internal workflow management system for client onboarding. It tracks every client from initial creation through two stages of onboarding, then into recurring billing.

**Core flow:**
```
Support creates client
  → Sales fills Stage 1 details + uploads agreements
  → Owner approves agreements
  → Sales uploads final signed copy
  → Stage 2 unlocks → CS executes 6 tasks
  → All tasks done → Sales activates account
  → Monthly invoicing begins
```

---

## 2. User Roles & Permissions

Roles are stored as a **JSON array** per user — one user can hold multiple roles simultaneously (e.g. `["sales","cs"]`).

| Role | Description |
|---|---|
| `support` | Creates clients and assigns Sales/CS PICs |
| `sales` | Fills Stage 1 fields, uploads agreements, uploads signed copy, manages invoices |
| `cs` | Executes Stage 2 tasks |
| `admin` | Approves agreements, approves deactivations, manages staff, views audit logs, marks invoices paid |

### Permission Matrix

| Action | support | sales | cs | admin |
|---|---|---|---|---|
| Create client | ✅ | ❌ | ❌ | ✅ |
| Assign Sales/CS to job | ✅ | ❌ | ❌ | ✅ |
| View all records | ✅ | ✅ | ✅ | ✅ |
| Fill Stage 1 fields | ❌ | ✅ | ❌ | ✅ |
| Upload agreements (SA/NDA) | ❌ | ✅ | ❌ | ✅ |
| Upload signed copy | ❌ | ✅ | ❌ | ✅ |
| Approve/reject agreements | ❌ | ❌ | ❌ | ✅ |
| Add remarks on agreements | ❌ | ❌ | ❌ | ✅ |
| Execute Stage 2 tasks | ❌ | ❌ | ✅ | ✅ |
| Activate account (1st time) | ❌ | ✅ | ❌ | ✅ |
| Pause/unpause account | ❌ | ✅ | ❌ | ✅ |
| Request deactivation | ❌ | ✅ | ❌ | ✅ |
| Approve/reject deactivation | ❌ | ❌ | ❌ | ✅ |
| Register new staff | ✅ | ❌ | ❌ | ✅ |
| Register admin role | ❌ | ❌ | ❌ | ✅ |
| Record/upload invoices | ❌ | ✅ | ❌ | ✅ |
| Mark invoices as paid | ❌ | ❌ | ❌ | ✅ |
| View audit logs | ❌ | ❌ | ❌ | ✅ |

---

## 3. Database Schema

### `users`
```sql
CREATE TABLE users (
  id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name               VARCHAR(255) NOT NULL,
  email              VARCHAR(255) NOT NULL UNIQUE,
  password           VARCHAR(255) NOT NULL,        -- bcrypt hashed
  role               JSON NOT NULL,                -- array e.g. ["sales","cs"]
  is_active          BOOLEAN NOT NULL DEFAULT TRUE,
  email_verified_at  TIMESTAMP NULL,
  remember_token     VARCHAR(100) NULL,
  created_at         TIMESTAMP NULL,
  updated_at         TIMESTAMP NULL
);
```

**Role values:** `support`, `sales`, `cs`, `admin`  
**Note:** A user can have multiple roles: `["sales","cs"]`

---

### `personal_access_tokens` (Sanctum / JWT)
```sql
CREATE TABLE personal_access_tokens (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  tokenable_type VARCHAR(255) NOT NULL,
  tokenable_id   BIGINT UNSIGNED NOT NULL,
  name           TEXT NOT NULL,
  token          VARCHAR(64) NOT NULL UNIQUE,
  abilities      TEXT NULL,
  last_used_at   TIMESTAMP NULL,
  expires_at     TIMESTAMP NULL,
  created_at     TIMESTAMP NULL,
  updated_at     TIMESTAMP NULL,
  INDEX tokenable (tokenable_type, tokenable_id)
);
```

---

### `clients`
```sql
CREATE TABLE clients (
  id                           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  company_name                 VARCHAR(255) NOT NULL,
  todo_list                    TEXT NULL,
  account_status               ENUM('inactive','active','paused') NOT NULL DEFAULT 'inactive',
  pending_account_status       VARCHAR(50) NULL,           -- 'inactive' when Sales requests deactivation
  pending_status_requested_by  BIGINT UNSIGNED NULL,
  pending_status_requested_at  TIMESTAMP NULL,
  created_by                   BIGINT UNSIGNED NOT NULL,
  created_at                   TIMESTAMP NULL,
  updated_at                   TIMESTAMP NULL,
  FOREIGN KEY (pending_status_requested_by) REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (created_by) REFERENCES users(id)
);
```

---

### `job_requests`
```sql
CREATE TABLE job_requests (
  id                   BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  client_id            BIGINT UNSIGNED NOT NULL,
  status               ENUM('pending','client_pending','pending_to_owner','completed') NOT NULL DEFAULT 'pending',
  current_stage        TINYINT NOT NULL DEFAULT 1,        -- 1 or 2
  indicator            ENUM('grey','yellow','green','red') NOT NULL DEFAULT 'grey',
  customer_pic         VARCHAR(255) NULL,                 -- Stage 1: Sales fills
  monthly_recurring    DECIMAL(10,2) NULL,                -- Stage 1: Sales fills
  account_type         VARCHAR(255) NULL,                 -- Stage 1: Standard/Premium/Enterprise
  assigned_sales_id    BIGINT UNSIGNED NULL,
  assigned_cs_id       BIGINT UNSIGNED NULL,
  signed_file_path     VARCHAR(255) NULL,                 -- final signed PDF
  signed_uploaded_at   TIMESTAMP NULL,
  signed_uploaded_by   BIGINT UNSIGNED NULL,
  sla_started_at       TIMESTAMP NULL,                    -- set when client is created
  sla_deadline         TIMESTAMP NULL,                    -- sla_started_at + 14 days
  last_activity_at     TIMESTAMP NULL,
  stage1_approved_at   TIMESTAMP NULL,                    -- set when signed copy uploaded
  stage1_approved_by   BIGINT UNSIGNED NULL,              -- Owner who approved SA+NDA
  created_by           BIGINT UNSIGNED NOT NULL,
  created_at           TIMESTAMP NULL,
  updated_at           TIMESTAMP NULL,
  FOREIGN KEY (client_id)          REFERENCES clients(id) ON DELETE CASCADE,
  FOREIGN KEY (assigned_sales_id)  REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (assigned_cs_id)     REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (signed_uploaded_by) REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (stage1_approved_by) REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (created_by)         REFERENCES users(id)
);
```

---

### `agreements`
```sql
CREATE TABLE agreements (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  job_request_id  BIGINT UNSIGNED NOT NULL,
  type            ENUM('service_agreement','nda') NOT NULL,
  version         INT UNSIGNED NOT NULL DEFAULT 1,         -- auto-increments per type per job
  file_path       VARCHAR(255) NOT NULL,
  status          ENUM('draft','pending_approval','approved','rejected') NOT NULL DEFAULT 'draft',
  uploaded_by     BIGINT UNSIGNED NOT NULL,
  approved_by     BIGINT UNSIGNED NULL,
  approved_at     TIMESTAMP NULL,
  notes           TEXT NULL,                               -- rejection reason
  owner_remarks   TEXT NULL,                               -- admin comments on the document
  created_at      TIMESTAMP NULL,
  updated_at      TIMESTAMP NULL,
  FOREIGN KEY (job_request_id) REFERENCES job_requests(id) ON DELETE CASCADE,
  FOREIGN KEY (uploaded_by)    REFERENCES users(id),
  FOREIGN KEY (approved_by)    REFERENCES users(id) ON DELETE SET NULL
);
```

---

### `tasks`
```sql
CREATE TABLE tasks (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  job_request_id  BIGINT UNSIGNED NOT NULL,
  task_type       ENUM(
                    'verify_details',
                    'business_flow',
                    'crm',
                    'business_accelerator',
                    'database_reactive',
                    'onboarding'
                  ) NOT NULL,
  status          ENUM('pending','in_progress','completed') NOT NULL DEFAULT 'pending',
  remarks         TEXT NULL,
  updated_by      BIGINT UNSIGNED NULL,
  completed_at    TIMESTAMP NULL,
  created_at      TIMESTAMP NULL,
  updated_at      TIMESTAMP NULL,
  FOREIGN KEY (job_request_id) REFERENCES job_requests(id) ON DELETE CASCADE,
  FOREIGN KEY (updated_by)     REFERENCES users(id) ON DELETE SET NULL
);
```

**Note:** All 6 tasks are created automatically when a job request is created.

---

### `audit_logs`
```sql
CREATE TABLE audit_logs (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id         BIGINT UNSIGNED NULL,
  action          VARCHAR(255) NOT NULL,       -- created, updated, deleted, login, logout, approved, rejected
  auditable_type  VARCHAR(255) NOT NULL,       -- model class name e.g. "Client"
  auditable_id    BIGINT UNSIGNED NOT NULL,
  old_values      JSON NULL,                   -- values before update
  new_values      JSON NULL,                   -- values after update / full record on create
  ip_address      VARCHAR(45) NULL,
  created_at      TIMESTAMP NULL,
  updated_at      TIMESTAMP NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
  INDEX idx_auditable (auditable_type, auditable_id)
);
```

---

### `invoices`
```sql
CREATE TABLE invoices (
  id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  invoice_number      VARCHAR(255) NOT NULL UNIQUE,        -- format: INV-YYYYMM-0001
  client_id           BIGINT UNSIGNED NOT NULL,
  job_request_id      BIGINT UNSIGNED NOT NULL,
  assigned_sales_id   BIGINT UNSIGNED NULL,
  assigned_cs_id      BIGINT UNSIGNED NULL,
  amount              DECIMAL(10,2) NOT NULL,
  sales_commission    DECIMAL(10,2) NOT NULL DEFAULT 0,    -- 10% of amount
  cs_commission       DECIMAL(10,2) NOT NULL DEFAULT 0,    -- 10% of amount
  billing_month       DATE NOT NULL,                       -- stored as YYYY-MM-01
  status              ENUM('pending','paid','overdue') NOT NULL DEFAULT 'pending',
  notes               TEXT NULL,
  file_path           VARCHAR(255) NULL,                   -- uploaded invoice PDF
  file_uploaded_at    TIMESTAMP NULL,
  paid_at             TIMESTAMP NULL,
  paid_by             BIGINT UNSIGNED NULL,
  created_by          BIGINT UNSIGNED NOT NULL,
  created_at          TIMESTAMP NULL,
  updated_at          TIMESTAMP NULL,
  UNIQUE KEY unique_client_month (client_id, billing_month),   -- one invoice per client per month
  FOREIGN KEY (client_id)         REFERENCES clients(id),
  FOREIGN KEY (job_request_id)    REFERENCES job_requests(id),
  FOREIGN KEY (assigned_sales_id) REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (assigned_cs_id)    REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (paid_by)           REFERENCES users(id) ON DELETE SET NULL,
  FOREIGN KEY (created_by)        REFERENCES users(id)
);
```

---

## 4. API Endpoints

**Base URL:** `/api`  
**Auth:** Bearer token in `Authorization: Bearer {token}` header  
**Content-Type:** `application/json` (multipart/form-data for file uploads)

---

### Authentication

#### `POST /api/login`
**Public**

Request:
```json
{ "email": "string", "password": "string" }
```

Response `200`:
```json
{
  "token": "1|abc123...",
  "user": { "id": 1, "name": "Admin", "email": "admin@...", "role": ["admin"] },
  "email_verified": true
}
```

Response `403` (unverified email):
```json
{ "message": "Your email address is not verified.", "email_verified": false }
```

Response `403` (inactive account):
```json
{ "message": "Your account has been deactivated. Contact the Owner." }
```

---

#### `POST /api/logout`
**Auth required**

Response `200`:
```json
{ "message": "Logged out successfully." }
```

---

#### `GET /api/me`
**Auth required**

Response `200`: Current user object

---

#### `POST /api/forgot-password`
**Public**

Request: `{ "email": "string" }`  
Response `200`: `{ "message": "If that email exists, a password reset link has been sent." }`

---

#### `POST /api/reset-password`
**Public**

Request:
```json
{
  "token": "string",
  "email": "string",
  "password": "string",
  "password_confirmation": "string"
}
```

Response `200`: `{ "message": "Password reset successfully." }`

---

#### `POST /api/email/resend-verification`
**Public**

Request: `{ "email": "string" }`  
Response `200`: `{ "message": "Verification email sent." }`

---

#### `GET /api/email/verify/{id}/{hash}?expires=&signature=`
**Public + Signed URL**

Response `200`: `{ "message": "Email verified successfully." }`

---

### Staff

#### `GET /api/staff?role=sales`
**Auth required**

Returns array of users. Optional `?role=` filter uses JSON contains (e.g. `?role=cs` returns users who have `cs` in their role array).

Response:
```json
[
  { "id": 1, "name": "Sales Rep", "email": "...", "role": ["sales","cs"], "is_active": true }
]
```

---

#### `POST /api/staff`
**Auth required — support or admin**

Request:
```json
{
  "name": "string",
  "email": "string",
  "password": "string (min 8)",
  "role": ["sales", "cs"]    // array, support can't create admin
}
```

Response `201`: Created user object. Also sends verification email.

---

#### `PATCH /api/staff/{id}/role`
**Auth required — admin only**

Request: `{ "role": ["sales", "cs"] }`  
Response `200`: Updated user object.

---

#### `POST /api/staff/{id}/deactivate`
**Auth required — admin only**  
Cannot deactivate self. Revokes all tokens immediately.

Response `200`: `{ "message": "...", "user": {...} }`

---

#### `POST /api/staff/{id}/activate`
**Auth required — admin only**

Response `200`: `{ "message": "...", "user": {...} }`

---

#### `DELETE /api/staff/{id}`
**Auth required — admin only**  
Cannot delete self. Sets FK references to NULL.

Response `200`: `{ "message": "Staff account deleted." }`

---

### Clients

#### `GET /api/clients`
**Auth required**

- `support` / `admin`: all clients
- `sales`: only clients where `assigned_sales_id = current_user.id`
- `cs`: only clients where `assigned_cs_id = current_user.id`

Response:
```json
{
  "clients": [...],
  "total_recurring": 15000.00   // sum of monthly_recurring for returned clients
}
```

---

#### `POST /api/clients`
**Auth required — support only**

Request:
```json
{
  "company_name": "string",
  "todo_list": "string (optional)",
  "assigned_sales_id": 2,
  "assigned_cs_id": 3
}
```

**Side effects on creation:**
1. Client created with `account_status = inactive`
2. JobRequest auto-created: `status=pending, stage=1, indicator=grey, sla_started_at=now, sla_deadline=now+14days`
3. 6 Tasks auto-created: all `status=pending`
4. AuditLog entries created for Client, JobRequest, and each Task

---

#### `GET /api/clients/{id}`
**Auth required — all roles**

Response: Client with job_requests (including tasks and agreements)

---

#### `PUT /api/clients/{id}`
**Auth required — support only**

Request: `{ "company_name": "string", "todo_list": "string" }`

---

#### `POST /api/clients/{id}/activate`
**Auth required — sales only**

Rules:
- If `account_status = inactive`: Stage 2 must be completed first
- If `account_status = paused`: activates directly, no Stage 2 check

Clears any pending deactivation request.

---

#### `POST /api/clients/{id}/pause`
**Auth required — sales only**

Toggles: `active → paused` or `paused → active`

---

#### `POST /api/clients/{id}/request-deactivate`
**Auth required — sales only**

Sets `pending_account_status = inactive`. Admin must approve.

---

#### `POST /api/clients/{id}/approve-deactivate`
**Auth required — admin only**

Sets `account_status = inactive`, clears pending fields.

---

#### `POST /api/clients/{id}/reject-deactivate`
**Auth required — admin only**

Clears pending fields, account_status unchanged.

---

### Job Requests

#### `GET /api/job-requests`
**Auth required — all roles**

Returns all jobs with SLA status appended to each:
```json
[
  {
    "id": 1,
    "...": "...",
    "sla": {
      "indicator": "yellow",
      "sla_deadline": "2026-05-16 10:00:00",
      "days_remaining": 11,
      "sla_overdue": false,
      "days_since_update": 1,
      "stale": false
    }
  }
]
```

---

#### `GET /api/job-requests/{id}`
**Auth required — all roles**

Full detail including client, assigned staff, tasks (with updatedBy), agreements (with uploader/approver), signedUploader, stage1Approver. SLA status appended.

---

#### `GET /api/job-requests/{id}/sla`
**Auth required — all roles**

Returns SLA status object only.

---

#### `PATCH /api/job-requests/{id}/stage1`
**Auth required — sales only**

Only works when `current_stage = 1`. Updates `last_activity_at`.

Request:
```json
{
  "customer_pic": "string",
  "monthly_recurring": 5000.00,
  "account_type": "Standard|Premium|Enterprise"
}
```

---

#### `POST /api/job-requests/{id}/signed-copy`
**Auth required — sales only**

**Rules before upload:**
- `current_stage` must be 1
- Both `service_agreement` and `nda` must have at least one `approved` version

**On upload:**
- Stores PDF file
- Sets `signed_file_path`, `signed_uploaded_at`, `signed_uploaded_by`
- Sets `stage1_approved_at = now`, `current_stage = 2`, `status = pending`, `indicator = yellow`
- Stage 2 is now unlocked

Request: `multipart/form-data` with field `file` (PDF only)

---

#### `GET /api/job-requests/{id}/signed-copy/download`
**Auth required — all roles**

Streams the signed copy PDF.

---

### Agreements

#### `GET /api/job-requests/{id}/agreements`
**Auth required — all roles**

Returns all agreements ordered by type, then version descending.

---

#### `POST /api/job-requests/{id}/agreements`
**Auth required — sales or admin**

Request: `multipart/form-data`
```
type: service_agreement | nda
file: PDF/DOC/DOCX (max 10MB)
notes: string (optional)
```

**Rules:**
- Sales: only when `current_stage = 1`
- Admin: any time, no stage restriction
- Auto-increments version per type per job
- Sales upload: `status = pending_approval`, sets job `status = pending_to_owner`
- Admin upload: `status = approved`, `approved_by = current_user`, auto-approved

---

#### `GET /api/agreements/{id}/download`
**Auth required — all roles**

Streams the agreement file.

---

#### `POST /api/agreements/{id}/remarks`
**Auth required — admin only**

Request: `{ "owner_remarks": "string" }`  
Response: Updated agreement.

---

#### `POST /api/agreements/{id}/approve`
**Auth required — admin only**

Rules: `status` must be `pending_approval`

Sets `status = approved`, `approved_by`, `approved_at`.

**Does NOT auto-complete Stage 1** — Sales must separately upload the signed copy.

---

#### `POST /api/agreements/{id}/reject`
**Auth required — admin only**

Rules: `status` must be `pending_approval`

Request: `{ "notes": "string (optional)" }`  
Sets `status = rejected`, saves rejection notes.  
Sets job `status = client_pending`.

---

### Tasks

#### `GET /api/job-requests/{id}/tasks`
**Auth required — all roles**

Returns all 6 tasks ordered by id.

---

#### `PATCH /api/tasks/{id}`
**Auth required — cs only**

Rules: Job's `current_stage` must be 2.

Request:
```json
{
  "status": "in_progress | completed",
  "remarks": "string (optional)"
}
```

When `completed`: sets `completed_at = now`, updates `last_activity_at` on job.

**When all 6 tasks = completed:**
- Sets `job_request.status = completed`
- Sets `job_request.indicator = green`

---

### Dashboard / Monitoring

#### `GET /api/dashboard`
**Auth required — all roles**

Response:
```json
{
  "summary": {
    "total_jobs": 4,
    "active_clients": 1,
    "overdue_jobs": 1,
    "pending_approvals": 2,
    "stale_jobs": 1,
    "stuck_stage2_jobs": 0,
    "missing_fields_jobs": 1
  },
  "overdue_jobs": [...],
  "pending_approvals": [...],
  "stale_jobs": [...],
  "stuck_tasks": [...],
  "missing_fields": [...]
}
```

**Flag definitions:**
- `overdue_jobs`: `status != completed` AND `sla_deadline < now`
- `pending_approvals`: agreements with `status = pending_approval`
- `stale_jobs`: `status != completed` AND `last_activity_at < now - 3 days`
- `stuck_tasks`: `current_stage = 2` AND `status != completed` AND has incomplete tasks
- `missing_fields`: `current_stage = 1` AND any of (`customer_pic`, `monthly_recurring`, `account_type`) is NULL

---

#### `GET /api/audit-logs?page=1`
**Auth required — admin only**

Paginated (50 per page). Returns:
```json
{
  "data": [{ "id":1, "user": {...}, "action": "created", "auditable_type": "Client", ... }],
  "total": 100,
  "per_page": 50,
  "current_page": 1,
  "last_page": 2
}
```

---

### Invoices

#### `GET /api/invoices?month=2026-05-01&status=pending`
**Auth required — sales or admin**

- Sales: only their assigned clients' invoices
- Admin: all invoices

---

#### `GET /api/invoices/active-clients?month=2026-05-01`
**Auth required — sales or admin**

Returns active clients with their invoice status for the given month. Includes `overdue_missing: true` flag if today > 5th and no invoice exists.

Response per client:
```json
{
  "client_id": 1,
  "company_name": "Acme Corp",
  "job_request_id": 1,
  "monthly_recurring": 5000.00,
  "assigned_sales": { "id": 2, "name": "Sales Rep", "email": "..." },
  "assigned_cs":    { "id": 3, "name": "CS Agent",  "email": "..." },
  "invoice": null | { invoice object },
  "invoiced": false,
  "overdue_missing": true
}
```

---

#### `GET /api/invoices/admin-overview?month=2026-05-01`
**Auth required — admin only**

Returns all active clients with invoice status for the month + totals summary.

---

#### `GET /api/invoices/commissions?month=2026-05-01`
**Auth required — all roles**

Returns commission breakdown per staff member for the month.

---

#### `POST /api/invoices`
**Auth required — sales or admin**

Request: `multipart/form-data` (file is optional at creation)
```
client_id:      integer
job_request_id: integer
amount:         decimal
billing_month:  date (YYYY-MM-DD)
notes:          string (optional)
file:           PDF (optional)
```

**On creation:**
- Auto-generates `invoice_number` format: `INV-YYYYMM-0001`
- Copies `assigned_sales_id` and `assigned_cs_id` from job_request
- Calculates: `sales_commission = amount * 0.10`, `cs_commission = amount * 0.10`
- One invoice per client per month (unique constraint enforced)

---

#### `POST /api/invoices/{id}/upload-file`
**Auth required — sales or admin**

Request: `multipart/form-data` with `file` (PDF)

---

#### `GET /api/invoices/{id}/download`
**Auth required — all roles**

Streams the invoice PDF.

---

#### `POST /api/invoices/{id}/pay`
**Auth required — admin only**

Sets `status = paid`, `paid_at = now`, `paid_by = current_user.id`.

---

#### `PATCH /api/invoices/{id}`
**Auth required — sales or admin**

Request: `{ "status": "pending|overdue", "notes": "string", "amount": decimal }`  
Note: `paid` status can only be set via `/pay` endpoint.

---

## 5. Business Logic Rules

### Client Creation (Support)
1. Create client with `account_status = inactive`
2. Auto-create one JobRequest: `stage=1, status=pending, indicator=grey, sla_started_at=now, sla_deadline=now+14days, last_activity_at=now`
3. assign pic for sales and client success (CS)
3. Auto-create 6 Tasks (all `pending`) linked to the job request
4. Write audit log for client, job_request, and each task (action=`created`)

### Stage 1 Completion Flow
1. Sales fills fields - monthly recurring, initial recurring, setup fee. The setup is fee does not go to any commission  `PATCH /stage1`
2. Sales uploads Service Agreement → `status=pending_approval`, job `status=pending_to_owner`
3. Sales uploads NDA → same
4. Admin approves both (separately via `POST /agreements/{id}/approve`)
5. After both approved, Sales uploads **signed copy** via `POST /signed-copy`
6. Signed copy upload triggers Stage 2 unlock: `current_stage=2, stage1_approved_at=now, indicator=yellow`


### Stage 2 Completion Flow
1. CS updates each of 6 tasks with status + remarks
2. Each update sets `last_activity_at` on job
3. When all 6 tasks = `completed`: job `status=completed, indicator=green`
4. Sales can then activate the account

### Account Status Transitions
```
inactive ──[Sales: activate, Stage 2 required]──► active
active   ──[Sales: pause]──► paused
paused   ──[Sales: activate]──► active (no Stage 2 check)
active/paused ──[Sales: request-deactivate]──► pending_account_status=inactive
  └── [Admin: approve-deactivate]──► account_status=inactive
  └── [Admin: reject-deactivate]──► clears pending, status unchanged
```

### Audit Logging
- Every write operation on Client, JobRequest, Agreement, Task is automatically logged
- Login and logout are logged explicitly
- `old_values` = fields before update (only changed fields)
- `new_values` = new values / full record on create
- Sensitive fields (passwords) are never logged

---

## 6. File Storage

All files use the **S3-compatible API** (DigitalOcean Spaces in production, local disk in dev).

| File type | Storage path |
|---|---|
| Agreements | `agreements/job-{job_id}/{filename}` |
| Signed copies | `signed-copies/job-{job_id}/{filename}` |
| Invoice PDFs | `invoices/{YYYY-MM-01}/{filename}` |

**Download approach:** Files are never served directly from public URLs. All downloads go through authenticated API endpoints that stream the file content. This keeps files private.

**Accepted types:**
- Agreements: PDF, DOC, DOCX (max 10MB)
- Signed copies: PDF only (max 10MB)
- Invoices: PDF only (max 10MB)

---

## 7. Email Notifications

Uses **Resend** as the email provider. All emails link to the **frontend URL**, not the API.

### Verification Email
- Sent when: new staff member is registered
- Link format: `{FRONTEND_URL}/verify-email?id={id}&hash={sha1_email}&expires={ts}&signature={sig}`
- Expires: 60 minutes
- User clicks link → frontend calls `GET /api/email/verify/{id}/{hash}?expires=&signature=`

### Password Reset Email
- Sent when: `POST /api/forgot-password` called
- Link format: `{FRONTEND_URL}/reset-password?token={token}&email={email}`
- Expires: 60 minutes
- User submits new password → frontend calls `POST /api/reset-password`

---

## 8. SLA & Indicator Rules

### SLA
- Duration: **14 days** from client creation
- `sla_started_at` = when client is created
- `sla_deadline` = `sla_started_at + 14 days`
- `last_activity_at` = updated on any write to the job

### Indicator Calculation
Run this logic on every job (recalculated hourly via cron job):

```
if status == "completed"          → green
if now > sla_deadline             → red
if now - last_activity_at > 3d   → red  (stale)
if stage == 1
   AND customer_pic IS NULL
   AND monthly_recurring IS NULL
   AND account_type IS NULL       → grey  (not started)
else                              → yellow (in progress)
```

### SLA Status Response Object
```json
{
  "indicator": "yellow",
  "sla_deadline": "2026-05-16 10:00:00",
  "days_remaining": 11,      // negative = overdue
  "sla_overdue": false,
  "days_since_update": 1,
  "stale": false             // true if days_since_update > 3
}
```

### Cron Job
Run every hour: recalculate and update `indicator` on all non-completed job requests.

---

## 9. Invoice & Commission Rules

### Invoice Number Format
`INV-YYYYMM-XXXX` where XXXX is zero-padded sequence for that month.  
Example: `INV-202605-0001`

### Commission Calculation
On every invoice creation or amount update:
```
sales_commission = amount × 0.10
cs_commission    = amount × 0.10
```
`assigned_sales_id` and `assigned_cs_id` are copied from the job request at invoice creation time.

### Overdue Rule
If viewing the **current month** AND today's date > **5th of the month** AND no invoice exists for an active client → that client's invoice slot is flagged as `overdue`.

### One Invoice Per Client Per Month
Unique constraint on `(client_id, billing_month)`. Attempting to create a second invoice for the same client+month returns `422`.

### Payment Flow
1. Sales records invoice + uploads PDF
2. Admin reviews, downloads PDF, clicks "Mark Paid"
3. `status = paid`, `paid_at = now`, `paid_by = admin_id`

---

## 10. Frontend Pages & Components

**Framework:** Astro 5 (SSR mode with Node adapter) + Vue 3 + Tailwind CSS  
**State management:** Pinia  
**HTTP client:** Axios with auto Bearer token injection

### Auth Store (Pinia)
```typescript
interface User {
  id: number
  name: string
  email: string
  role: string[]       // e.g. ["sales", "cs"]
  is_active: boolean
}

// Methods
login(email, password): Promise<void>   // stores token in localStorage
logout(): Promise<void>                 // revokes token, redirects to /login
fetchUser(): Promise<void>              // GET /me, populates user
```

### Role Checks (frontend)
```typescript
// Always use .includes() — role is an array
auth.user?.role?.includes('admin')
auth.user?.role?.includes('sales')
auth.user?.role?.some(r => ['sales','cs'].includes(r))
```

### Pages

| URL | Auth | Description |
|---|---|---|
| `/` | No | Redirect: token → `/dashboard`, else → `/login` |
| `/login` | No | Login form. Shows unverified banner with resend button. |
| `/forgot-password` | No | Email form → sends reset link |
| `/reset-password?token=&email=` | No | New password form |
| `/verify-email?id=&hash=&expires=&signature=` | No | Auto-verifies on mount |
| `/dashboard` | Yes | Monitoring page: stat cards + 5 flag sections |
| `/clients` | Yes | Client list table. Support sees + New Client button. Sales/CS see only their clients + commission summary. |
| `/job-requests` | Yes | Table of all job requests with live SLA countdown |
| `/job-requests/{id}` | Yes | Full job detail: Stage 1 form + agreements + Stage 2 tasks + account status actions |
| `/invoices` | Yes (sales/admin) | Sales: record invoices + upload PDFs. Admin: all clients table + mark paid + commission summary. |
| `/approvals` | Yes (admin) | Pending agreements list with approve/reject |
| `/staff` | Yes (support/admin) | Staff list with role pickers. Admin can deactivate/delete. |
| `/audit-logs` | Yes (admin) | Paginated audit log table |

### Key Component Behaviours

**Job Detail Page (`/job-requests/{id}`):**
- Stage tabs: Stage 1 and Stage 2 (Stage 2 locked 🔒 until signed copy uploaded)
- Auto-refreshes every 30 seconds
- Stage 1 tab: fields form (Sales editable), AgreementPanel, signed copy upload (Sales, after both agreements approved)
- Stage 2 tab: task list (CS editable), AccountStatusActions

**AgreementPanel:**
- Shows SA and NDA separately
- Each shows version history (highest version first)
- Sales: upload button (Stage 1 only), each agreement shows download button
- Admin: approve/reject buttons on `pending_approval` items, owner remarks textarea, upload button (any time, auto-approved)
- Download goes through API (authenticated) — not direct URL

**SignedCopyUpload:**
- Shown to Sales ONLY when: Stage 1 open AND both SA + NDA have at least one `approved` version AND no signed copy yet
- Uploading this triggers Stage 2 unlock

**SLA Countdown:**
- Updates every second in the browser
- Red when overdue, orange when ≤ 3 days, green when completed

**RolePicker:**
- Pill-shaped toggle buttons for each role
- Colours: admin=blue, support=sky, sales=green, cs=orange
- At least one role must remain selected

---

## Environment Variables

| Variable | Description |
|---|---|
| `APP_URL` | API base URL (e.g. `https://api.yourdomain.com`) |
| `FRONTEND_URL` | Frontend URL for email links |
| `DB_HOST` / `DB_DATABASE` / `DB_USERNAME` / `DB_PASSWORD` | MySQL connection |
| `FILESYSTEM_DISK` | `local` (dev) or `s3` (production) |
| `AWS_ACCESS_KEY_ID` | DigitalOcean Spaces access key |
| `AWS_SECRET_ACCESS_KEY` | DigitalOcean Spaces secret key |
| `AWS_DEFAULT_REGION` | Spaces region (e.g. `sgp1`) |
| `AWS_BUCKET` | Spaces bucket name |
| `AWS_ENDPOINT` | `https://{region}.digitaloceanspaces.com` |
| `MAIL_MAILER` | `resend` |
| `RESEND_API_KEY` | Resend API key |
| `MAIL_FROM_ADDRESS` | Verified sender email |
| `SANCTUM_STATELESS_DOMAINS` | Frontend domain (for Sanctum token auth) |
