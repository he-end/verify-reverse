# Ekosistem Validator

Dokumentasi penggunaan `service.Validator` — dari pemuatan instance, registrasi rule kustom, membaca report error, hingga integrasi bersih di Handler/Adapter.

---

## 1. Arsitektur

```
┌─────────────────────────────────────────────┐
│  service.Validator                          │
│  ┌───────────────────────────────┐          │
│  │  validator.Validate (engine)  │          │
│  │  • RegisterTagNameFunc        │          │
│  │  • RegisterValidation         │          │
│  │  • RegisterStructValidation   │          │
│  └──────────────┬────────────────┘          │
│                 │                            │
│  ┌──────────────▼────────────────┐          │
│  │  .Struct(s)  .Var(v, tag)    │          │
│  │         ↓                     │          │
│  │      *Report                  │          │
│  └──────────────────────────────┘          │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  Report                                     │
│  ┌───────────────────────────────────────┐  │
│  │  []FieldError                         │  │
│  │  • Field   : string (json field name) │  │
│  │  • Actual  : string (tag yg gagal)    │  │
│  │  • Detail  : string (pesan manusiawi) │  │
│  └───────────────────────────────────────┘  │
│  Methods: .HasErrors() .Error() .ToMap()    │
└─────────────────────────────────────────────┘
```

---

## 2. Pemuatan Config (`Validator`)

### 2.1 Instance Default (singleton)

Instance global sudah diinisialisasi di `init()` dan dapat langsung dipakai di mana saja. Cocok untuk aplikasi kecil — satu validator untuk seluruh service.

```go
import "github.com/he-end/verify-reverse/auth/service"

val := service.Default()
```

**Cara kerjanya:**
- `init()` di `validation.go` memanggil `NewValidator()`.
- `RegisterTagNameFunc` otomatis terpasang — field name di-report pakai `json` tag, bukan nama field Go.

### 2.2 Instance Kustom (isolated)

Gunakan `NewValidator()` jika butuh validator terpisah (misal: untuk test suite, atau rule yang berbeda per bounded context).

```go
val := service.NewValidator()

// daftarkan custom rule untuk domain ini saja
val.RegisterValidation("countrycode", validateCountryCode)
```

### 2.3 Wiring di Route

```go
// auth/route.go
func (route routeConf) WebhookWA(r *http.ServeMux) {
    logg := logger.GetLoggerRuntimeStore()
    waService := service.SetupWAService(...)

    regSim := registersimulator.LoadHandler(
        waService,
        *responser.Responser,
        *logg,
        service.Default(),   // ← inject validator
    )
    r.HandleFunc("/api/v1.0/register", regSim.RegisterViaWA)
}
```

### 2.4 Akses Engine Bawaan

Ketika butuh perilaku yang belum di-wrap oleh `Validator`:

```go
rawEngine := service.Default().Engine()
rawEngine.RegisterTagNameFunc(customFn)
```

---

## 3. Registrasi Validator (Custom Rules)

Ada dua level registrasi: **field-level** (single field) dan **struct-level** (lintas field).

### 3.1 Tag-Based Field Validation

Gunakan struct tag `validate:"..."` pada model request. `go-playground/validator` mendukung puluhan built-in tag.

```go
type RegisterViaWAReqBody struct {
    Number     *string `json:"number"      validate:"required,numeric,len=10"`
    Name       *string `json:"name"        validate:"required,min=2,max=50"`
    Pwd        *string `json:"pwd"         validate:"required,min=8"`
    ConfirmPwd *string `json:"confirm_pwd" validate:"required,eqfield=Pwd"`
    Email      *string `json:"email"       validate:"omitempty,email"`
    Plan       *string `json:"plan"        validate:"required,oneof=free pro enterprise"`
}
```

**Tag yang didukung penuh oleh Report (Detail auto-generated):**

| Tag         | Contoh Tag                              | Detail Output                          |
|------------|-----------------------------------------|----------------------------------------|
| `required` | `validate:"required"`                   | `field is required`                    |
| `min`      | `validate:"min=8"`                      | `must be at least 8`                   |
| `max`      | `validate:"max=100"`                    | `must be at most 100`                  |
| `len`      | `validate:"len=10"`                     | `must be exactly 10 characters long`   |
| `email`    | `validate:"email"`                      | `must be a valid email address`        |
| `url`      | `validate:"url"`                        | `must be a valid URL`                  |
| `uuid`     | `validate:"uuid"`                       | `must be a valid UUID`                 |
| `oneof`    | `validate:"oneof=free pro enterprise"`  | `must be one of [free pro enterprise]` |
| `eqfield`  | `validate:"eqfield=Pwd"`               | `must equal Pwd`                       |
| `numeric`  | `validate:"numeric"`                    | `must be numeric`                      |
| `boolean`  | `validate:"boolean"`                    | `must be a boolean value`              |
| `json`     | `validate:"json"`                       | `must be valid JSON`                   |

Tag lain (`gt`, `gte`, `lt`, `lte`, `ne`, `nefield`, `alpha`, `alphanum`) juga otomatis dikonversi ke pesan yang dapat dibaca.

### 3.2 Custom Validation Function (tag baru)

```go
// file: auth/handler/register_simulator/build_validator.go

func (h *handler) registerValidator(val *service.Validator) {
    // custom field-level rule
    val.RegisterValidation("countrycode", func(fl validator.FieldLevel) bool {
        code := fl.Field().String()
        _, ok := allowedCountryCodes[code]
        return ok
    })
}
```

Kemudian gunakan di struct:

```go
type Address struct {
    Country string `json:"country" validate:"required,countrycode"`
}
```

### 3.3 Struct-Level Validation (lintas field)

Digunakan ketika validasi satu field bergantung pada field lain:

```go
// file: auth/handler/register_simulator/build_validator.go

func (h *handler) registerValidator(val *service.Validator) {
    val.RegisterStructLevelValidation(regValRegisterViaWA)
}

func regValRegisterViaWA(sl validator.StructLevel) {
    reg := sl.Current().Interface().(RegisterViaWAReqBody)

    if reg.Number == nil {
        sl.ReportError(reg.Number, "number", "Number", "required", "")
    }
    if *reg.Pwd != *reg.ConfirmPwd {
        sl.ReportError(reg.ConfirmPwd, "confirm_pwd", "ConfirmPwd", "eqfield", "Pwd")
    }
}
```

---

## 4. Report Error (Converter)

### 4.1 Struktur Data

```go
type FieldError struct {
    Field  string `json:"field"`   // nama dari json tag
    Actual string `json:"actual"`  // tag rule yang gagal, e.g. "required"
    Detail string `json:"detail"`  // deskripsi manusiawi, e.g. "field is required"
}

type Report struct {
    Errors []FieldError `json:"errors"`
}
```

### 4.2 Metode Report

| Metode          | Return             | Keterangan                                        |
|----------------|--------------------|----------------------------------------------------|
| `.HasErrors()` | `bool`             | `true` jika ada error validasi                     |
| `.IsEmpty()`   | `bool`             | kebalikan dari `HasErrors`                         |
| `.ToMap()`     | `map[string]string`| `{"field": "detail", ...}` untuk form response    |
| `.Error()`     | `string`           | implementasi `error` interface                     |

### 4.3 Cara Membaca Report

```go
report := val.Struct(reqBody)

if report.HasErrors() {
    for _, fe := range report.Errors {
        fmt.Printf("%s: %s (tag: %s)\n", fe.Field, fe.Detail, fe.Actual)
    }
}
// Output:
// number: field is required (tag: required)
// name: must be at least 2 (tag: min)
```

### 4.4 Serialisasi JSON (response API)

```json
// Request:  {}
// Response: 422 Unprocessable Entity
{
  "errors": [
    {"field": "number",       "actual": "required", "detail": "field is required"},
    {"field": "name",         "actual": "required", "detail": "field is required"},
    {"field": "pwd",          "actual": "required", "detail": "field is required"},
    {"field": "confirm_pwd",  "actual": "required", "detail": "field is required"}
  ]
}
```

### 4.5 Single‑Variable Validation (`Var`)

```go
report := val.Var("invalid-email", "email")
if report.HasErrors() {
    fmt.Println(report.Errors[0].Detail) // "must be a valid email address"
}
```

---

## 5. Pola Bersih di Handler/Adapter

### 5.1 Pattern Dasar — Validasi → Report → Response

```go
// auth/handler/register_simulator/register.go

func (h *handler) RegisterViaWA(w http.ResponseWriter, r *http.Request) {
    type ReqBody struct {
        Number *string `json:"number" validate:"required,numeric"`
        Name   *string `json:"name"   validate:"required,min=2,max=50"`
    }

    defer r.Body.Close()

    var req ReqBody
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.resp.Error(w, "invalid request body", "", err.Error(), http.StatusBadRequest)
        return
    }

    // ───── validasi ─────
    if report := h.val.Struct(req); report.HasErrors() {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(report)
        return
    }

    // ───── bisnis logic ─────
    // ...
}
```

### 5.2 Ringkasan Flow Handler

```
HTTP Request
    │
    ▼
Decode JSON body (json.Decoder)
    │
    ├─ error ──► 400 Bad Request
    │
    ▼
h.val.Struct(req)
    │
    ├─ report.HasErrors() ──► 422 Unprocessable Entity
    │                           {"errors": [...]}
    │
    ▼
Bisnis Logic
    │
    ├─ service error ──► 400 / 500 (via resp.Error)
    │
    ▼
200 OK → JSON response sukses
```

### 5.3 Pemisahan Kategori Response

Gunakan `h.resp.Error()` hanya untuk error di luar validasi — DB down, API eksternal gagal, dll. Untuk validation error, gunakan `*Report` langsung sebagai JSON response body.

| Kategori        | HTTP Status | Body Generator       |
|----------------|-------------|----------------------|
| Validation fail | 422         | `*Report` → JSON     |
| Infrastructure  | 4xx / 5xx   | `resp.Error()`       |
| Success         | 200         | inline struct → JSON |

---

## 6. Menambah Detail Message Kustom

Edit fungsi `buildDetail()` di `auth/service/validation.go`:

```go
func buildDetail(fe validator.FieldError) string {
    param := fe.Param()
    tag := fe.ActualTag()

    switch tag {
    case "required":
        return "field is required"

    // ─── tambahkan di sini ───
    case "countrycode":
        return "must be a valid ISO 3166-1 country code"

    case "strongpwd":
        return "must contain uppercase, digit, and special character"
    // ─────────────────────────

    default:
        return fmt.Sprintf("failed validation on %s", tag)
    }
}
```

---

## 7. Testing

Karena `Report` adalah struct biasa tanpa dependency eksternal, unit test sangat sederhana:

```go
func TestRegisterValidation(t *testing.T) {
    val := service.NewValidator()

    tests := []struct {
        name   string
        input  interface{}
        wantOK bool
    }{
        {"empty number",     &Req{Number: nil,       Name: ptr("")},     false},
        {"name too short",   &Req{Number: ptr("089"), Name: ptr("A")},   false},
        {"valid",            &Req{Number: ptr("0895"), Name: ptr("Hend")}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            report := val.Struct(tt.input)
            if report.HasErrors() == tt.wantOK {
                t.Errorf("HasErrors()=%v, want=%v | %v",
                    report.HasErrors(), !tt.wantOK, report.Error())
            }
        })
    }
}

func ptr(s string) *string { return &s }
```

---

## 8. Ringkasan File Terkait

| File | Peran |
|------|-------|
| `auth/service/validation.go` | `Validator` struct, `Report`, `FieldError`, converter |
| `auth/handler/register_simulator/build_validator.go` | Registrasi custom rules per handler |
| `auth/handler/register_simulator/register.go` | Pemakaian `h.val.Struct()` di handler |
| `auth/route.go` | Wiring `service.Default()` ke handler constructor |
