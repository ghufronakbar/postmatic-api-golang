// internal/module/headless/mailer/service.go
package mailer

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"postmatic-api/config"
	"postmatic-api/pkg/errs"

	"gopkg.in/gomail.v2"
)

// MAGIC DISINI:
// Perintah ini akan membaca semua file di folder 'templates'
// dan memasukkannya ke variable 'templateFS' saat compile.
//
//go:embed templates/*.html
var templateFS embed.FS

type MailerService struct {
	dialer *gomail.Dialer

	// constants
	sender       string
	appName      string
	logo         string
	address      string
	contactEmail string
}

func NewService(cfg *config.Config) *MailerService {
	d := gomail.NewDialer(
		cfg.SMTP_HOST,
		cfg.SMTP_PORT,
		cfg.SMTP_USER,
		cfg.SMTP_PASS,
	)

	return &MailerService{
		dialer:       d,
		sender:       cfg.SMTP_SENDER,
		appName:      cfg.APP_NAME,
		logo:         cfg.APP_LOGO,
		address:      cfg.APP_ADDRESS,
		contactEmail: cfg.APP_CONTACT_EMAIL,
	}
}

// Helper function untuk template (agar bisa passing parameter ke component)
var funcMap = template.FuncMap{
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errs.NewInternalServerError(errors.New("invalid dict call"))
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errs.NewInternalServerError(errors.New("dict keys must be strings"))
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
}

func (s *MailerService) SendEmail(ctx context.Context, input SendEmailInput) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.sender)
	m.SetHeader("To", input.To)
	m.SetHeader("Subject", input.Subject)

	bodyContent := input.Body
	contentType := "text/plain"

	if input.TemplateName != "" {
		// 1. Parse Layout DAN Template Spesifik
		// Kita butuh layout.html untuk kerangka, dan input.TemplateName untuk isinya
		tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/layout.html", "templates/"+input.TemplateName)
		if err != nil {
			return err
		}

		// 2. Siapkan Data Lengkap (Gabung data input + Konstanta Global)
		// Kita convert input.Data ke map agar bisa kita tambah field-nya
		// (Atau buat struct wrapper baru jika mau lebih strict type-nya)
		dataMap := make(map[string]interface{})

		// Inject Global Constants
		dataMap["AppName"] = s.appName
		dataMap["Logo"] = s.logo
		dataMap["Address"] = s.address
		dataMap["ContactEmail"] = s.contactEmail
		dataMap["Subject"] = input.Subject // Butuh subject di title head

		// Inject Input Data (Merge)
		// Asumsi input.Data adalah struct atau map, kita perlu merge manual atau gunakan reflection
		// Cara simpel: input.Data sebaiknya sudah map, atau kita cast manual fields-nya di controller
		// Untuk contoh ini, kita asumsikan input.Data sudah berupa map[string]interface{} atau struct yang diakses via reflection.
		// JIKA input.Data adalah STRUCT spesifik (seperti di AuthService), template tetap bisa akses fieldnya.
		// TAPI agar data global (Logo, dll) masuk, cara terbaik adalah membuat struct wrapper:

		type TemplateData struct {
			AppName      string
			Logo         string
			Address      string
			ContactEmail string
			Subject      string
			// Embed Data Spesifik user
			Data interface{}
		}

		// Namun, karena template Go mencari field di root ({{.Name}} bukan {{.Data.Name}}),
		// cara paling fleksibel adalah menggunakan Map untuk `Execute`.

		// Mari kita pakai Map untuk fleksibilitas maksimal di contoh ini:
		finalData := map[string]interface{}{
			"AppName":      s.appName,
			"Logo":         s.logo,
			"Address":      s.address,
			"ContactEmail": s.contactEmail,
			"Subject":      input.Subject,
		}

		// Merge data spesifik dari input (disini kita asumsikan input.Data adalah Map atau Struct)
		// Jika Struct, template tidak bisa merge secara langsung tanpa reflection ribet.
		// JADI, SOLUSI TERBAIK UNTUK GO TEMPLATE:
		// Akses data spesifik lewat key, misal {{ .Payload.Name }}
		finalData["Payload"] = input.Data

		// TAPI tunggu, template HTML di atas pakai {{.Name}} langsung.
		// Agar support direct access {{.Name}}, kita harus kirim struct yang memiliki semua field (Global + Spesifik).
		// KARENA ITU RIBET (harus bikin struct kombinasi), SAYA SARANKAN UBAH HTML SEDIKIT:

		// UBAH HTML: {{.Name}} -> {{.Name}} (Tetap sama), TAPI execute template harus pintar.
		// Cara termudah di Go: Struct Embedding secara dinamis itu susah.
		// Mari kita gunakan pendekatan Map Merge manual via library atau loop sederhana jika input.Data adalah Map.

		// JIKA ANDA MENGIRIM STRUCT DARI AUTH SERVICE:
		// Sebaiknya template memanggil {{ .Name }} dan input.Data yang dikirim sudah berisi AppName, Logo, dll dari Service Auth?
		// Tentu tidak efisien.

		// SOLUSI PRAGMATIS:
		// Ubah template HTML sedikit agar menerima {{ .Name }} dll,
		// lalu saat execute, kita passing data gabungan.

		// Code di bawah ini asumsikan input.Data adalah map[string]interface{} dari Auth Service.
		if inputMap, ok := input.Data.(map[string]interface{}); ok {
			for k, v := range inputMap {
				finalData[k] = v
			}
		} else {
			// Jika input.Data adalah Struct, kita assign sebagai "Data"
			// Maka di HTML harus ubah jadi {{ .Data.Name }}
			// ATAU: Kita gunakan json marshal/unmarshal hack untuk convert struct ke map
			// Ini sedikit overhead tapi sangat memudahkan templating.
			// (Implementasi hack convert struct to map ada di bawah)
			structMap, _ := structToMap(input.Data)
			for k, v := range structMap {
				finalData[k] = v
			}
		}

		var buffer bytes.Buffer
		// Execute "layout.html" karena itu root template-nya, dia akan panggil "content"
		// Perhatikan nama filenya mungkin jadi "layout.html" di internal template name
		if err := tmpl.ExecuteTemplate(&buffer, "layout", finalData); err != nil {
			return err
		}

		bodyContent = buffer.String()
		contentType = "text/html"
	}

	m.SetBody(contentType, bodyContent)

	if err := s.dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

// Helper cepat untuk convert struct ke map (agar template bisa akses {{.Name}} langsung)

func structToMap(data interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	return m, err
}
