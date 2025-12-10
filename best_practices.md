
<b>Pattern 1: Validate and normalize request path params and context values early; check types explicitly (e.g., UUID format and string type for user identifiers) and return precise HTTP status codes for bad input.</b>

Example code before:
```
id := r.PathValue("consentId")
if id == "" {
  http.Error(w, "missing id", http.StatusBadRequest)
  return
}
userEmail := r.Context().Value("userEmail") // assume string
// use userEmail.(string) later
```

Example code after:
```
id := r.PathValue("consentId")
if _, err := uuid.Parse(id); err != nil {
  http.Error(w, "invalid id format", http.StatusBadRequest)
  return
}
v := r.Context().Value("userEmail")
email, ok := v.(string)
if !ok || email == "" {
  http.Error(w, "unauthorized: missing user email", http.StatusUnauthorized)
  return
}
```

<details><summary>Examples for relevant past discussions:</summary>

- https://github.com/OpenDIF/opendif-core/pull/381#discussion_r2601721336
- https://github.com/OpenDIF/opendif-core/pull/329#discussion_r2549888606
</details>


___

<b>Pattern 2: Distinguish validation errors from server errors in handlers and map them to 4xx vs 5xx responses; return 400 for invalid payloads or required field violations instead of a generic 500.</b>

Example code before:
```
if err := svc.Create(ctx, req); err != nil {
  http.Error(w, "failed to create", http.StatusInternalServerError)
  return
}
```

Example code after:
```
if err := svc.Create(ctx, req); err != nil {
  if isValidationErr(err) || strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") {
    http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
    return
  }
  http.Error(w, "failed to create: "+err.Error(), http.StatusInternalServerError)
  return
}
```

<details><summary>Examples for relevant past discussions:</summary>

- https://github.com/OpenDIF/opendif-core/pull/329#discussion_r2549888606
</details>


___

<b>Pattern 3: Avoid runtime panics and nil dereferences in DTO mapping; when transferring optional fields (e.g., pointers), preserve pointer types or provide safe fallbacks instead of dereferencing blindly.</b>

Example code before:
```
// AppName is optional in storage
resp.AppName = *cr.AppName // may panic if nil
```

Example code after:
```
// Keep pointer to avoid deref panic or supply default
resp.AppName = cr.AppName
// or:
resp.AppName = func() *string { return cr.AppName }()
```

<details><summary>Examples for relevant past discussions:</summary>

- https://github.com/OpenDIF/opendif-core/pull/374#discussion_r2598318607
</details>


___

<b>Pattern 4: Wrap multi-step database writes in a transaction to preserve consistency (e.g., delete-then-create or batch updates); use the transaction handle for all operations and roll back on error.</b>

Example code before:
```
// delete old records
db.Where("schema_id = ?", sid).Delete(&PolicyMetadata{})
// create new records
if err := db.Create(&records).Error; err != nil {
  return err // may leave partial state
}
```

Example code after:
```
err := db.Transaction(func(tx *gorm.DB) error {
  if err := tx.Where("schema_id = ?", sid).Delete(&PolicyMetadata{}).Error; err != nil {
    return err
  }
  if len(records) > 0 {
    if err := tx.Create(&records).Error; err != nil {
      return err
    }
  }
  return nil
})
if err != nil { return err }
```

<details><summary>Examples for relevant past discussions:</summary>

- https://github.com/OpenDIF/opendif-core/pull/232#discussion_r2467783937
</details>


___

<b>Pattern 5: Optimize service lookups and data access patterns to avoid O(n) scans and N+1 queries; use maps or indexed queries for frequent lookups and batch-load records when possible.</b>

Example code before:
```
// O(n) linear scan for each request
var found *Provider
for _, p := range h.providers {
  if p.ServiceKey == key && p.SchemaID == schemaID { found = p; break }
}
```

Example code after:
```
// Build index for O(1) lookups
idxKey := func(key, schemaID string) string { return key + ":" + schemaID }
providerIdx := map[string]*Provider{}
// populate index on init/update
p := providerIdx[idxKey(key, schemaID)]
```

<details><summary>Examples for relevant past discussions:</summary>

- https://github.com/OpenDIF/opendif-core/pull/238#discussion_r2472440399
- https://github.com/OpenDIF/opendif-core/pull/238#discussion_r2472445841
- https://github.com/OpenDIF/opendif-core/pull/232#discussion_r2467764328
</details>


___

<b>Pattern 6: Ensure robust initialization and build-time checks; constructors should return errors instead of panicking, and top-level startup should validate prerequisites (env vars, toolchains) before proceeding.</b>

Example code before:
```
func NewV1Handler(db *gorm.DB) *V1Handler {
  url := os.Getenv("PDP_SERVICE_URL")
  if url == "" { panic("missing PDP url") }
  return &V1Handler{...}
}
```

Example code after:
```
func NewV1Handler(db *gorm.DB) (*V1Handler, error) {
  url := os.Getenv("PDP_SERVICE_URL")
  if url == "" {
    return nil, fmt.Errorf("PDP_SERVICE_URL is required")
  }
  return &V1Handler{...}, nil
}
// In main:
h, err := NewV1Handler(db)
if err != nil { slog.Error("init failed", "error", err); os.Exit(1) }
```

<details><summary>Examples for relevant past discussions:</summary>

- https://github.com/OpenDIF/opendif-core/pull/232#discussion_r2467754656
- https://github.com/OpenDIF/opendif-core/pull/232#discussion_r2467761012
- https://github.com/OpenDIF/opendif-core/pull/348#discussion_r2565973257
</details>


___
