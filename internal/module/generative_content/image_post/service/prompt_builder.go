// internal/module/generative_content/image_post/service/prompt_builder.go
package image_post_service

import (
	"fmt"
	"strings"

	business_knowledge_service "postmatic-api/internal/module/business/business_knowledge/service"
	business_product_service "postmatic-api/internal/module/business/business_product/service"
	business_role_service "postmatic-api/internal/module/business/business_role/service"
)

// PromptBuilder adalah struct untuk membangun prompt AI
type PromptBuilder struct {
	Business          business_knowledge_service.BusinessKnowledgeResponse
	Product           business_product_service.BusinessProductResponse
	Role              business_role_service.BusinessRoleResponse
	Adv               *AdvanceGenerateInput
	AdditionalPrompt  string
	DesignStyle       string
	Category          string
	Ratio             string
	HasReferenceImage bool
	HasLogo           bool
	RSS               *RSSInput
}

// RSSInput adalah struct untuk input RSS
type RSSInput struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	PublishedAt string  `json:"publishedAt"`
	ImageURL    *string `json:"imageUrl"`
	Summary     string  `json:"summary"`
	Publisher   string  `json:"publisher"`
}

// NewPromptBuilder creates a new PromptBuilder
func NewPromptBuilder(input PromptBuilderInput) *PromptBuilder {
	return &PromptBuilder{
		Business:         input.BusinessKnowledge,
		Product:          input.Product,
		Role:             input.Role,
		Adv:              input.Adv,
		AdditionalPrompt: input.AdditionalPrompt,
		DesignStyle:      input.DesignStyle,
		Category:         input.Category,
	}
}

// PromptBuilderInput adalah input untuk membangun prompt
type PromptBuilderInput struct {
	BusinessKnowledge business_knowledge_service.BusinessKnowledgeResponse
	Product           business_product_service.BusinessProductResponse
	Role              business_role_service.BusinessRoleResponse
	Adv               *AdvanceGenerateInput
	AdditionalPrompt  string
	Category          string
	DesignStyle       string
}

// ================== NEGATIVE PROMPT ==================

// BuildNegativePrompt returns comprehensive negative prompt to avoid AI artifacts
func (p *PromptBuilder) BuildNegativePrompt() string {
	return `
NEGATIVE PROMPT — HINDARI HAL BERIKUT:
- Teks tambahan / watermark / signature / timestamp
- Noise, blur, banding, moiré, posterization, compression artifacts
- Proporsi objek aneh (tangan/jari, mata, telinga, perspektif tidak logis)
- Komposisi tidak seimbang, crop kepala/produk, ruang negatif tidak perlu
- Warna oversaturated atau lighting tidak realistis
- Distorsi logo/ikon/brand assets, warna brand melenceng
- Ketajaman rendah, depth of field berantakan, edge halo
- Pengulangan pattern tak sengaja, object cloning yang tidak masuk akal
- Teks yang typo/salah ejaan
`
}

// ================== CAPTION INSTRUCTION ==================

// BuildCaptionInstruction builds instruction for AI caption generation
func (p *PromptBuilder) BuildCaptionInstruction() string {
	businessSummary := p.buildBusinessKnowledge()
	productSummary := p.buildProductKnowledge()
	roleSummary := p.buildRoleKnowledge()
	contentSummary := p.buildContentContext()

	var rssSection string
	if p.RSS != nil {
		rssSection = p.buildRSSPrompt()
	}

	return fmt.Sprintf(`
PERAN:
Anda adalah copywriter profesional berbahasa Indonesia untuk media sosial.

TUJUAN:
Tulis caption yang singkat, jelas, menarik, dan mendorong aksi.

KONTEKS:
%s
%s
%s
%s
%s

ATURAN PENULISAN:
- Output HANYA CAPTION, tanpa judul/emoji berlebihan/markdown/penjelasan.
- 2–3 kalimat padat + 1 CTA jelas.
- Gunakan tone & persona sesuai Role.
- Jika Role memiliki daftar hashtags "fixed", JANGAN ubah. Boleh tambahkan 2–4 hashtag relevan lainnya di akhir.
%s

KUALITAS:
- Bahasa Indonesia yang natural, tidak kaku, tidak clickbait murahan.
- Hindari tanda seru berlebihan, hindari frasa generik.
`,
		contentSummary,
		businessSummary,
		productSummary,
		roleSummary,
		rssSection,
		p.buildRSSCaptionNote())
}

func (p *PromptBuilder) buildRSSCaptionNote() string {
	if p.RSS != nil {
		return "- Kaitkan insight/topik RSS secara relevan & ringkas (bukan rangkuman panjang)."
	}
	return ""
}

// ================== IMAGE PROMPT FOR GENERATE ==================

// BuildImagePromptForGenerate builds prompt for generate mode with knowledge design
func (p *PromptBuilder) BuildImagePromptForGenerate() string {
	businessSummary := p.buildBusinessKnowledge()
	productSummary := p.buildProductKnowledge()
	roleSummary := p.buildRoleKnowledge()
	contentSummary := p.buildContentContext()
	compositionRules := p.buildCompositionRules()
	brandRules := p.buildBrandRules()
	negativePrompt := p.BuildNegativePrompt()

	ratio := p.Ratio
	if ratio == "" {
		ratio = "1:1"
	}
	designStyle := p.DesignStyle
	if designStyle == "" {
		designStyle = "modern"
	}

	var colorToneSection string
	if p.shouldAttachBusinessField("ColorTone") && p.Business.ColorTone != "" {
		colorToneSection = fmt.Sprintf("- Brand Color Tone: %s", p.Business.ColorTone)
		if p.HasReferenceImage {
			colorToneSection += fmt.Sprintf("\n- PENTING: Sesuaikan gambar referensi dengan color tone brand %s", p.Business.ColorTone)
		}
	}

	var referenceNote string
	if p.HasReferenceImage {
		referenceNote = `
CATATAN REF:
- Hormati struktur visual referensi, namun tingkatkan kualitas & konsistensi brand.
- Siapkan visual agar mudah diberi caption di tahap copywriting.`
	}

	return fmt.Sprintf(`
PERAN:
Anda adalah desainer konten sosial media profesional.

TUJUAN:
Buat komposisi final beraspek %s dengan gaya "%s".
Gunakan referensi visual terlampir sebagai acuan **layout, warna, tipografi, dekorasi**.
Integrasikan produk %s dan logo bisnis (jika diaktifkan) secara elegan & proporsional.
Optimalkan lighting, kontras, dan detail agar terlihat tajam & premium.
%s

SPESIFIKASI URUTAN LAMPIRAN GAMBAR (PENTING):
- #1: Foto Produk
- #2: Template/Referensi (opsional)
- #3: Logo (opsional)
- #4: Gambar RSS (opsional)
- #5: Mask (opsional; untuk instruksi khusus)
Selaraskan komposisi berdasarkan urutan di atas bila tersedia.

KONTEKS:
%s
%s
%s
%s

ATURAN KOMPOSISI & BRAND:
%s
%s
%s

OUTPUT:
- Hasilkan **gambar final** sesuai aspect ratio yang ditentukan.
- Jangan menambahkan teks baru kecuali benar-benar diperlukan untuk estetika (maks 1–3 kata dekoratif, bukan slogan).
- Hindari watermark/signature.
%s

PROMPT TAMBAHAN (opsional dari user):
%s
`,
		ratio,
		designStyle,
		p.getProductName(),
		colorToneSection,
		contentSummary,
		businessSummary,
		productSummary,
		roleSummary,
		compositionRules,
		brandRules,
		negativePrompt,
		referenceNote,
		p.getAdditionalPrompt())
}

// ================== IMAGE PROMPT FOR REGENERATE ==================

// BuildImagePromptForRegenerate builds prompt for regenerate mode (minimal invasive)
func (p *PromptBuilder) BuildImagePromptForRegenerate(userPrompt string) string {
	businessSummary := p.buildBusinessKnowledge()
	productSummary := p.buildProductKnowledge()
	roleSummary := p.buildRoleKnowledge()

	designStyle := p.DesignStyle
	if designStyle == "" {
		designStyle = "Default"
	}
	ratio := p.Ratio
	if ratio == "" {
		ratio = "1:1"
	}

	return fmt.Sprintf(`
PERAN:
Anda adalah editor gambar profesional untuk media sosial.

TUJUAN:
Edit gambar terlampir sesuai instruksi pengguna, **minimally invasive**:
ubah hanya bagian yang relevan, pertahankan komposisi & mood asli.

KONTEKS:
Design Style: %s
Ratio Target: %s
%s
%s
%s

INSTRUKSI PENGGUNA (PENTING):
%s

ATURAN:
- Pertahankan kualitas (ketajaman, noise rendah, warna akurat).
- Jangan menambahkan teks/elemen yang tidak diminta.
- Jika "attach logo" tidak diaktifkan dan sumber tanpa logo, **jangan** menambahkan logo.
`,
		designStyle,
		ratio,
		businessSummary,
		productSummary,
		roleSummary,
		userPrompt)
}

// ================== IMAGE PROMPT FOR RSS ==================

// BuildImagePromptForRSS builds prompt for RSS mode with trend integration
func (p *PromptBuilder) BuildImagePromptForRSS() string {
	if p.RSS == nil {
		return p.BuildImagePromptForGenerate()
	}

	businessSummary := p.buildBusinessKnowledge()
	productSummary := p.buildProductKnowledge()
	roleSummary := p.buildRoleKnowledge()
	contentSummary := p.buildContentContext()
	compositionRules := p.buildCompositionRules()
	brandRules := p.buildBrandRules()
	negativePrompt := p.BuildNegativePrompt()
	rssPrompt := p.buildRSSPrompt()

	ratio := p.Ratio
	if ratio == "" {
		ratio = "1:1"
	}
	designStyle := p.DesignStyle
	if designStyle == "" {
		designStyle = "sesuai referensi"
	}

	hasRSSImage := p.RSS.ImageURL != nil && *p.RSS.ImageURL != ""
	var rssImageNote string
	if hasRSSImage {
		rssImageNote = "Gambar RSS terlampir boleh diolah kreatif (blend, overlay, duotone) agar harmonis."
	}

	return fmt.Sprintf(`
PERAN:
Anda adalah desainer konten sosial media profesional.

TUJUAN:
Buat visual %s dengan gaya "%s".
Integrasikan **elemen tren dari RSS** secara **subtil & relevan** (ikon/warna/shape/pattern),
tanpa merusak komposisi utama. %s

SPESIFIKASI URUTAN LAMPIRAN GAMBAR:
- #1: Foto Produk
- #2: Gambar RSS (opsional)
- #3: Template/Referensi (opsional)
- #4: Logo (opsional)

KONTEKS:
%s
%s
%s
%s

RINGKASAN TREN (RSS):
%s

ATURAN KOMPOSISI & BRAND:
%s
%s
%s

OUTPUT:
- Hasilkan **gambar final** sesuai aspect ratio; visual RSS harus menjadi aksen, bukan fokus utama.
- Jangan menyalin teks panjang dari RSS; gunakan simbol/analogi visual.
- Tidak ada watermark/signature.

PROMPT TAMBAHAN (opsional dari user):
%s
`,
		ratio,
		designStyle,
		rssImageNote,
		contentSummary,
		businessSummary,
		productSummary,
		roleSummary,
		rssPrompt,
		compositionRules,
		brandRules,
		negativePrompt,
		p.getAdditionalPrompt())
}

// ================== IMAGE PROMPT FOR MASK ==================

// BuildImagePromptForMask builds prompt for mask mode with precise editing
func (p *PromptBuilder) BuildImagePromptForMask(userPrompt string) string {
	return fmt.Sprintf(`
PERAN:
Anda adalah editor gambar profesional.

TUGAS (MASKING):
- Interpretasi mask biner: **PUTIH = area untuk diedit**, **HITAM/transparan = JANGAN diubah**.
- Edit hanya area putih sesuai instruksi pengguna. Area lain harus tetap identik.

INSTRUKSI PENGGUNA:
%s

ATURAN KUALITAS:
- Pencahayaan, bayangan, perspektif harus konsisten.
- Hindari artefak/halo/tepi kasar pada perbatasan mask.
- Hasil akhir tajam dan realistis untuk media sosial.
`, userPrompt)
}

// ================== KNOWLEDGE BUILDERS ==================

func (p *PromptBuilder) buildBusinessKnowledge() string {
	if p.Adv == nil || p.Adv.BusinessKnowledge == nil {
		return ""
	}

	bk := p.Business
	flags := p.Adv.BusinessKnowledge

	var parts []string
	parts = append(parts, "BUSINESS KNOWLEDGE (filtered):")

	if flags.Name && bk.Name != "" {
		parts = append(parts, fmt.Sprintf("- Name: %s", bk.Name))
	}
	if flags.Description && bk.Description != "" {
		parts = append(parts, fmt.Sprintf("- Description: %s", bk.Description))
	}
	if flags.Category && bk.Category != "" {
		parts = append(parts, fmt.Sprintf("- Category: %s", bk.Category))
	}
	if flags.Location && bk.Location != "" {
		parts = append(parts, fmt.Sprintf("- Location: %s", bk.Location))
	}
	if flags.UniqueSellingPoint && bk.UniqueSellingPoint != "" {
		parts = append(parts, fmt.Sprintf("- USP: %s", bk.UniqueSellingPoint))
	}
	if flags.VisionMission && bk.VisionMission != "" {
		parts = append(parts, fmt.Sprintf("- Vision/Mission: %s", bk.VisionMission))
	}
	if flags.Website && bk.WebsiteUrl != nil && *bk.WebsiteUrl != "" {
		parts = append(parts, fmt.Sprintf("- Website: %s", *bk.WebsiteUrl))
	}
	if flags.ColorTone && bk.ColorTone != "" {
		parts = append(parts, fmt.Sprintf("- Brand Color Tone: %s", bk.ColorTone))
	}

	parts = append(parts, `
CATATAN BRAND:
- Ikuti palet & nuansa warna brand (jika tersedia).
- Logo harus tajam, tidak terdistorsi, dan kontras memadai terhadap background.`)

	return strings.Join(parts, "\n")
}

func (p *PromptBuilder) buildProductKnowledge() string {
	if p.Adv == nil || p.Adv.ProductKnowledge == nil {
		return ""
	}

	pd := p.Product
	flags := p.Adv.ProductKnowledge

	var parts []string
	parts = append(parts, "PRODUCT KNOWLEDGE (filtered):")

	if flags.Name && pd.Name != "" {
		parts = append(parts, fmt.Sprintf("- Name: %s", pd.Name))
	}
	if flags.Description && pd.Description != "" {
		parts = append(parts, fmt.Sprintf("- Description: %s", pd.Description))
	}
	if flags.Category && pd.Category != "" {
		parts = append(parts, fmt.Sprintf("- Category: %s", pd.Category))
	}
	if flags.Price && pd.Price > 0 {
		parts = append(parts, fmt.Sprintf("- Price: %s %d", pd.Currency, pd.Price))
	}

	parts = append(parts, `
CATATAN:
- Gunakan detail yang memperkuat **fit** terhadap referensi/template & brand.`)

	return strings.Join(parts, "\n")
}

func (p *PromptBuilder) buildRoleKnowledge() string {
	role := p.Role

	var parts []string
	parts = append(parts, "ROLE KNOWLEDGE:")

	if role.AudiencePersona != "" {
		parts = append(parts, fmt.Sprintf("- Persona: %s", role.AudiencePersona))
	}
	if role.CallToAction != "" {
		parts = append(parts, fmt.Sprintf("- CTA: %s", role.CallToAction))
	}
	if role.Goals != "" {
		parts = append(parts, fmt.Sprintf("- Goals: %s", role.Goals))
	}
	if role.TargetAudience != "" {
		parts = append(parts, fmt.Sprintf("- Target Audience: %s", role.TargetAudience))
	}
	if role.Tone != "" {
		parts = append(parts, fmt.Sprintf("- Tone: %s", role.Tone))
	}

	// Hashtags only if enabled in adv flags
	if p.Adv != nil && p.Adv.RoleKnowledge != nil && p.Adv.RoleKnowledge.Hashtags && len(role.Hashtags) > 0 {
		hashtags := make([]string, len(role.Hashtags))
		for i, h := range role.Hashtags {
			hashtags[i] = "#" + h
		}
		parts = append(parts, fmt.Sprintf("- Hashtags (fixed): %s", strings.Join(hashtags, ", ")))
	}

	parts = append(parts, `
CATATAN:
- Tone & persona membimbing mood visual (color/contrast/keberanian layout).`)

	return strings.Join(parts, "\n")
}

func (p *PromptBuilder) buildContentContext() string {
	var parts []string
	parts = append(parts, "IMAGE CONTENT:")

	if p.Category != "" {
		parts = append(parts, fmt.Sprintf("- Category: %s", p.Category))
	}
	if p.DesignStyle != "" {
		parts = append(parts, fmt.Sprintf("- Design Style: %s", p.DesignStyle))
	}
	if p.Ratio != "" {
		parts = append(parts, fmt.Sprintf("- Ratio: %s", p.Ratio))
	}

	parts = append(parts, `
CATATAN:
- Perlakukan ini sebagai constraint utama desain.`)

	return strings.Join(parts, "\n")
}

// ================== COMPOSITION & BRAND RULES ==================

func (p *PromptBuilder) buildCompositionRules() string {
	return `
ATURAN KOMPOSISI:
- Gunakan grid sederhana; jaga keseimbangan ruang negatif.
- Safe margin tepi (±40 px pada kanvas 1080px) untuk menghindari crop di feed.
- Produk adalah fokus visual: kontras, pencahayaan, dan ketajaman diprioritaskan.
- Hindari clutter; maksimalkan keterbacaan dan hierarki visual.`
}

func (p *PromptBuilder) buildBrandRules() string {
	return `
ATURAN BRAND:
- Logo diletakkan proporsional, tidak terdistorsi, dan tetap tajam.
- Hindari overlay efek berat pada logo (blur/emboss/glow berlebihan).`
}

// ================== RSS PROMPT ==================

func (p *PromptBuilder) buildRSSPrompt() string {
	if p.RSS == nil {
		return ""
	}

	rss := p.RSS
	imageURL := "N/A"
	if rss.ImageURL != nil && *rss.ImageURL != "" {
		imageURL = fmt.Sprintf("%s (attached)", *rss.ImageURL)
	}

	return fmt.Sprintf(`
RSS SUMMARY:
- Judul: %s
- Publisher: %s
- Terbit: %s
- URL: %s
- Gambar: %s
- Inti: %s`,
		rss.Title,
		rss.Publisher,
		rss.PublishedAt,
		rss.URL,
		imageURL,
		rss.Summary)
}

// ================== HELPER METHODS ==================

func (p *PromptBuilder) shouldAttachBusinessField(field string) bool {
	if p.Adv == nil || p.Adv.BusinessKnowledge == nil {
		return false
	}
	flags := p.Adv.BusinessKnowledge
	switch field {
	case "Name":
		return flags.Name
	case "Category":
		return flags.Category
	case "Description":
		return flags.Description
	case "Location":
		return flags.Location
	case "Logo":
		return flags.Logo
	case "UniqueSellingPoint":
		return flags.UniqueSellingPoint
	case "VisionMission":
		return flags.VisionMission
	case "Website":
		return flags.Website
	case "ColorTone":
		return flags.ColorTone
	}
	return false
}

func (p *PromptBuilder) getProductName() string {
	if p.Product.Name != "" {
		return p.Product.Name
	}
	return "klien"
}

func (p *PromptBuilder) getAdditionalPrompt() string {
	if p.AdditionalPrompt != "" {
		return p.AdditionalPrompt
	}
	return "N/A"
}

// ================== LEGACY COMPATIBILITY ==================

// BuildImagePrompt is legacy function for backward compatibility
func BuildImagePrompt(input PromptBuilderInput) string {
	builder := NewPromptBuilder(input)
	return builder.BuildImagePromptForGenerate()
}

// BuildCaptionPrompt is legacy function for backward compatibility
func BuildCaptionPrompt(input PromptBuilderInput) string {
	builder := NewPromptBuilder(input)
	return builder.BuildCaptionInstruction()
}
