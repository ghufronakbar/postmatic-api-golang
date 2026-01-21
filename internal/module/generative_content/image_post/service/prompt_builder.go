// internal/module/generative_content/image_post/service/prompt_builder.go
package image_post_service

import (
	"fmt"
	"strings"

	business_knowledge_service "postmatic-api/internal/module/business/business_knowledge/service"
	business_product_service "postmatic-api/internal/module/business/business_product/service"
	business_role_service "postmatic-api/internal/module/business/business_role/service"
)

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

// BuildImagePrompt builds prompt untuk image generation
func BuildImagePrompt(input PromptBuilderInput) string {
	var parts []string

	// Base instruction
	parts = append(parts, "Buatkan desain gambar poster/konten sosial media dengan detail berikut:")

	// Business knowledge
	if input.Adv != nil && input.Adv.BusinessKnowledge != nil {
		bk := input.BusinessKnowledge
		flags := input.Adv.BusinessKnowledge

		if flags.Name && bk.Name != "" {
			parts = append(parts, fmt.Sprintf("- Nama Bisnis: %s", bk.Name))
		}
		if flags.Category && bk.Category != "" {
			parts = append(parts, fmt.Sprintf("- Kategori Bisnis: %s", bk.Category))
		}
		if flags.Description && bk.Description != "" {
			parts = append(parts, fmt.Sprintf("- Deskripsi Bisnis: %s", bk.Description))
		}
		if flags.Location && bk.Location != "" {
			parts = append(parts, fmt.Sprintf("- Lokasi: %s", bk.Location))
		}
		if flags.UniqueSellingPoint && bk.UniqueSellingPoint != "" {
			parts = append(parts, fmt.Sprintf("- Keunggulan: %s", bk.UniqueSellingPoint))
		}
		if flags.VisionMission && bk.VisionMission != "" {
			parts = append(parts, fmt.Sprintf("- Visi Misi: %s", bk.VisionMission))
		}
		if flags.ColorTone && bk.ColorTone != "" {
			parts = append(parts, fmt.Sprintf("- Warna Dominan: %s", bk.ColorTone))
		}
	}

	// Product knowledge
	if input.Adv != nil && input.Adv.ProductKnowledge != nil {
		pd := input.Product
		flags := input.Adv.ProductKnowledge

		if flags.Name && pd.Name != "" {
			parts = append(parts, fmt.Sprintf("- Nama Produk: %s", pd.Name))
		}
		if flags.Category && pd.Category != "" {
			parts = append(parts, fmt.Sprintf("- Kategori Produk: %s", pd.Category))
		}
		if flags.Description && pd.Description != "" {
			parts = append(parts, fmt.Sprintf("- Deskripsi Produk: %s", pd.Description))
		}
		if flags.Price && pd.Price > 0 {
			parts = append(parts, fmt.Sprintf("- Harga: %s %d", pd.Currency, pd.Price))
		}
	}

	// Style dan category
	if input.DesignStyle != "" {
		parts = append(parts, fmt.Sprintf("- Gaya Desain: %s", input.DesignStyle))
	}
	if input.Category != "" {
		parts = append(parts, fmt.Sprintf("- Jenis Konten: %s", input.Category))
	}

	// Additional prompt
	if input.AdditionalPrompt != "" {
		parts = append(parts, fmt.Sprintf("\nInstruksi tambahan: %s", input.AdditionalPrompt))
	}

	return strings.Join(parts, "\n")
}

// BuildCaptionPrompt builds prompt untuk caption generation
func BuildCaptionPrompt(input PromptBuilderInput) string {
	var parts []string

	parts = append(parts, "Buatkan caption sosial media yang menarik untuk posting gambar dengan konteks berikut:")

	// Business knowledge (semua untuk caption)
	bk := input.BusinessKnowledge
	if bk.Name != "" {
		parts = append(parts, fmt.Sprintf("- Nama Bisnis: %s", bk.Name))
	}
	if bk.Category != "" {
		parts = append(parts, fmt.Sprintf("- Kategori Bisnis: %s", bk.Category))
	}
	if bk.Description != "" {
		parts = append(parts, fmt.Sprintf("- Deskripsi: %s", bk.Description))
	}

	// Product
	pd := input.Product
	if pd.Name != "" {
		parts = append(parts, fmt.Sprintf("- Produk: %s", pd.Name))
	}
	if pd.Description != "" {
		parts = append(parts, fmt.Sprintf("- Detail Produk: %s", pd.Description))
	}

	// Role knowledge
	role := input.Role
	if role.Tone != "" {
		parts = append(parts, fmt.Sprintf("- Tone: %s", role.Tone))
	}
	if role.CallToAction != "" {
		parts = append(parts, fmt.Sprintf("- Call to Action: %s", role.CallToAction))
	}
	if role.TargetAudience != "" {
		parts = append(parts, fmt.Sprintf("- Target Audiens: %s", role.TargetAudience))
	}

	// Hashtags
	if input.Adv != nil && input.Adv.RoleKnowledge != nil && input.Adv.RoleKnowledge.Hashtags && len(role.Hashtags) > 0 {
		parts = append(parts, fmt.Sprintf("- Sertakan hashtags: %s", strings.Join(role.Hashtags, " ")))
	}

	parts = append(parts, "\nBuat caption yang singkat, menarik, dan sesuai dengan konteks di atas.")

	return strings.Join(parts, "\n")
}
