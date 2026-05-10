 Here's the full diagram across all tables:                
                                                                
  users                                                         
  │  id · name · email · role · is_active                       
  │                                                             
  ├── personal_access_tokens                                    
  │     tokenable_id ──────────────────────→ users.id           
  │                                                             
  ├── project_staff                                             
  │     user_id ──────────────────────────→ users.id            
  │     project_id ───────────────────────→ projects.id         
  │
  ├── audit_logs                                                
  │     user_id ──────────────────────────→ users.id        
  │                                                             
  └── customer_crm                                        [CRM
  Space]                                                        
        id · company_name · account_status ·                
  recurring_start_date
        created_by ────────────────────────→ users.id
        pending_status_requested_by ───────→ users.id           
        │
        └── job_requests                                        
              client_id ─────────────────→ customer_crm.id      
              assigned_sales_id ─────────→ users.id
              assigned_cs_id ────────────→ users.id             
              signed_uploaded_by ────────→ users.id             
              stage1_approved_by ────────→ users.id
              created_by ────────────────→ users.id             
              │                                                 
              ├── agreements
              │     job_request_id ──────→ job_requests.id      
              │     uploaded_by ─────────→ users.id         
              │     approved_by ─────────→ users.id
              │                                                 
              ├── tasks
              │     job_request_id ──────→ job_requests.id      
              │     updated_by ──────────→ users.id         
              │                                                 
              └── invoices
                    job_request_id ──────→ job_requests.id      
                    client_id ───────────→ customer_crm.id  
                    assigned_sales_id ───→ users.id
                    assigned_cs_id ──────→ users.id             
                    paid_by ─────────────→ users.id
                    created_by ──────────→ users.id             
                                                                
  projects
  │  id · name · slug · is_active                               
  │                                                         
  └── project_staff
        project_id ─────────────────────→ projects.id
        user_id ────────────────────────→ users.id
                                                                
  users and projects are the two root tables — everything else  
  flows down from them. The entire CRM data chain hangs off     
  customer_crm, making it easy to isolate per-project later     
  (Investiland and Addhoc TWD will each have their own      
  equivalent root table).