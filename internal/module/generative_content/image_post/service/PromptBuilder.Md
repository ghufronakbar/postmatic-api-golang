Sebelumnya saya memiliki legacy code di nodejs typescript untuk prompt builder seperti berikut

```typescript
import { BaseService } from "../BaseService";
import {
  BusinessKnowledge,
  ProductKnowledge,
  RoleKnowledge,
} from ".prisma/client";
import {
  ImageContentAdvancedGenerateDTO,
  ImageContentDTO,
  ImageContentRegenerateDTO,
  ImageContentRssDTO,
  ValidRatio,
} from "src/validators/ImageContentValidator";

export class AiBaseService extends BaseService {
  constructor() {
    super();
  }

  /* ------------------------------ PROMPT BUILDERS ------------------------------ */

  /** Negative prompt ketat untuk menghindari artefak umum di image gen/edit. */
  protected buildNegativePrompt(): string {
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
`.trim();
  }

  /** Instruksi caption yang lebih ketat, tetap output single-line caption saja. */
  protected buildCaptionInstruction(args: {
    businessSummary: string;
    productSummary: string;
    roleSummary: string;
    body: ImageContentDTO;
    rss: ImageContentRssDTO["rss"] | null;
  }): string {
    const { businessSummary, productSummary, roleSummary, body, rss } = args;

    return `
PERAN:
Anda adalah copywriter profesional berbahasa Indonesia untuk media sosial.

TUJUAN:
Tulis caption yang singkat, jelas, menarik, dan mendorong aksi.

KONTEKS:
${this.buildPromptImageContent(body)}
${businessSummary}
${productSummary}
${roleSummary}
${rss ? this.buildRssPrompt(rss) : ""}

ATURAN PENULISAN:
- Output HANYA CAPTION, tanpa judul/emoji berlebihan/markdown/penjelasan.
- 2–3 kalimat padat + 1 CTA jelas.
- Gunakan tone & persona sesuai Role.
- Jika Role memiliki daftar hashtags “fixed”, JANGAN ubah. Boleh tambahkan 2–4 hashtag relevan lainnya di akhir.
${
  rss
    ? "- Kaitkan insight/topik RSS secara relevan & ringkas (bukan rangkuman panjang)."
    : ""
}

KUALITAS:
- Bahasa Indonesia yang natural, tidak kaku, tidak clickbait murahan.
- Hindari tanda seru berlebihan, hindari frasa generik.
`.trim();
  }

  /** Instruksi utama untuk generate dari knowledge + referensi visual. */
  protected buildKnowledgeDesignInstruction(args: {
    body: ImageContentDTO;
    business: Partial<BusinessKnowledge>;
    product: Partial<ProductKnowledge>;
    role: Partial<RoleKnowledge>;
    advancedGenerate: ImageContentAdvancedGenerateDTO;
    additionalPrompt: string;
    hasReferenceImage: boolean;
    extraNotes: string;
  }): string {
    const {
      body,
      business,
      product,
      role,
      advancedGenerate,
      additionalPrompt,
      hasReferenceImage,
      extraNotes,
    } = args;

    const businessSummary = this.buildPromptBusinessKnowledge(
      business,
      advancedGenerate,
    );
    const productSummary = this.buildPromptProductKnowledge(
      product,
      advancedGenerate,
    );
    const { colorTone: attachBusinessColorTone } =
      advancedGenerate.businessKnowledge;
    const { colorTone } = business;
    const { referenceImage } = body;
    const roleSummary = this.buildPromptRoleKnowledge(role, advancedGenerate);
    const contentSummary = this.buildPromptImageContent(body);

    return `
PERAN:
Anda adalah desainer konten sosial media profesional.

TUJUAN:
Buat komposisi final beraspek ${body.ratio ?? "1:1"} dengan gaya “${
      body.designStyle ?? "vintage"
    }”.
Gunakan referensi visual terlampir sebagai acuan **layout, warna, tipografi, dekorasi**.
Integrasikan produk ${
      product.name ?? "klien"
    } dan logo bisnis (jika diaktifkan) secara elegan & proporsional.
Optimalkan lighting, kontras, dan detail agar terlihat tajam & premium.
${
  attachBusinessColorTone && colorTone
    ? `- Brand Color Tone: #${colorTone}`
    : ""
}
${
  attachBusinessColorTone && colorTone && referenceImage
    ? `- PENTING: Sesuaikan gambar referensi dengan color tone brand #${colorTone}`
    : ""
}
${
  referenceImage
    ? `- Jika saya menyertakan informasi bisnis yang relevan, harap gunakan seperti pada referensi gambar saya`
    : ""
}

SPESIFIKASI URUTAN LAMPIRAN GAMBAR (PENTING):
- #1: Foto Produk
- #2: Template/Referensi (opsional)
- #3: Logo (opsional)
- #4: Gambar RSS (opsional)
- #5: Mask (opsional; untuk instruksi khusus)
Selaraskan komposisi berdasarkan urutan di atas bila tersedia.

KONTEKS:
${contentSummary}
${businessSummary}
${productSummary}
${roleSummary}

ATURAN KOMPOSISI & BRAND:
${this.buildCompositionRules()}
${this.buildBrandRules(business)}
${extraNotes}

OUTPUT:
- Hasilkan **gambar final** sesuai aspect ratio yang ditentukan.
- Jangan menambahkan teks baru kecuali benar-benar diperlukan untuk estetika (maks 1–3 kata dekoratif, bukan slogan).
- Hindari watermark/signature.
${
  hasReferenceImage
    ? `CATATAN REF:
- Hormati struktur visual referensi, namun tingkatkan kualitas & konsistensi brand.
- Siapkan visual agar mudah diberi caption di tahap copywriting.`
    : ""
}

PROMPT TAMBAHAN (opsional dari user):
${additionalPrompt || "N/A"}
`.trim();
  }

  /** Instruksi generate berbasis tren/RSS + integrasi elemen visual subtil. */
  protected buildTrendDesignInstruction(args: {
    body: ImageContentDTO;
    business: Partial<BusinessKnowledge>;
    product: Partial<ProductKnowledge>;
    role: Partial<RoleKnowledge>;
    advancedGenerate: ImageContentAdvancedGenerateDTO;
    rssSummaryBlock: string;
    additionalPrompt: string;
    hasRssImage: boolean;
    extraNotes: string;
  }): string {
    const {
      body,
      business,
      product,
      role,
      advancedGenerate,
      rssSummaryBlock,
      additionalPrompt,
      hasRssImage,
      extraNotes,
    } = args;

    const businessSummary = this.buildPromptBusinessKnowledge(
      business,
      advancedGenerate,
    );
    const productSummary = this.buildPromptProductKnowledge(
      product,
      advancedGenerate,
    );
    const roleSummary = this.buildPromptRoleKnowledge(role, advancedGenerate);
    const contentSummary = this.buildPromptImageContent(body);

    return `
PERAN:
Anda adalah desainer konten sosial media profesional.

TUJUAN:
Buat visual ${body.ratio} dengan gaya “${
      body.designStyle ?? "sesuai referensi"
    }”.
Integrasikan **elemen tren dari RSS** secara **subtil & relevan** (ikon/warna/shape/pattern),
tanpa merusak komposisi utama. ${
      hasRssImage
        ? "Gambar RSS terlampir boleh diolah kreatif (blend, overlay, duotone) agar harmonis."
        : ""
    }

SPESIFIKASI URUTAN LAMPIRAN GAMBAR:
- #1: Foto Produk
- #2: Gambar RSS (opsional)
- #3: Template/Referensi (opsional)
- #4: Logo (opsional)

KONTEKS:
${contentSummary}
${businessSummary}
${productSummary}
${roleSummary}

RINGKASAN TREN (RSS):
${rssSummaryBlock}

ATURAN KOMPOSISI & BRAND:
${this.buildCompositionRules()}
${this.buildBrandRules(business)}
${extraNotes}

OUTPUT:
- Hasilkan **gambar final** sesuai aspect ratio; visual RSS harus menjadi aksen, bukan fokus utama.
- Jangan menyalin teks panjang dari RSS; gunakan simbol/analogi visual.
- Tidak ada watermark/signature.

PROMPT TAMBAHAN (opsional dari user):
${additionalPrompt || "N/A"}
`.trim();
  }

  /** Instruksi edit/regenerate: presisi pada perubahan, sisanya dipertahankan. */
  protected buildRegenerateInstruction(args: {
    body: ImageContentRegenerateDTO & { image: string; ratio: ValidRatio };
    businessSummary: string;
    productSummary: string;
    roleSummary: string;
  }): string {
    const { body, businessSummary, productSummary, roleSummary } = args;

    return `
PERAN:
Anda adalah editor gambar profesional untuk media sosial.

TUJUAN:
Edit gambar terlampir sesuai instruksi pengguna, **minimally invasive**:
ubah hanya bagian yang relevan, pertahankan komposisi & mood asli.

KONTEKS:
Design Style: ${body.designStyle ?? "Default"}
Ratio Target: ${body.ratio}
${businessSummary}
${productSummary}
${roleSummary}

INSTRUKSI PENGGUNA (PENTING):
${body.prompt || "N/A"}

ATURAN:
- Pertahankan kualitas (ketajaman, noise rendah, warna akurat).
- Jangan menambahkan teks/elemen yang tidak diminta.
- Jika "attach logo" tidak diaktifkan dan sumber tanpa logo, **jangan** menambahkan logo.
`.trim();
  }

  /** Instruksi masking yang konsisten: putih=EDIT, hitam/transparan=JANGAN ubah. */
  protected buildMaskInstruction(userPrompt: string): string {
    return `
PERAN:
Anda adalah editor gambar profesional.

TUGAS (MASKING):
- Interpretasi mask biner: **PUTIH = area untuk diedit**, **HITAM/transparan = JANGAN diubah**.
- Edit hanya area putih sesuai instruksi pengguna. Area lain harus tetap identik.

INSTRUKSI PENGGUNA:
${userPrompt}

ATURAN KUALITAS:
- Pencahayaan, bayangan, perspektif harus konsisten.
- Hindari artefak/halo/tepi kasar pada perbatasan mask.
- Hasil akhir tajam dan realistis untuk media sosial.
`.trim();
  }

  /** Ringkasan RSS yang ringkas dan terstruktur. */
  protected buildRssPrompt(params: ImageContentRssDTO["rss"]): string {
    const { publishedAt, summary, title, url, publisher, imageUrl } = params;
    return `
RSS SUMMARY:
- Judul: ${title}
- Publisher: ${publisher}
- Terbit: ${publishedAt}
- URL: ${url}
- Gambar: ${imageUrl ? `${imageUrl} (attached)` : "N/A"}
- Inti: ${summary}
`.trim();
  }

  /* ------------------------------ KNOWLEDGE BLOCKS ----------------------------- */

  protected buildPromptProductKnowledge(
    product: Partial<ProductKnowledge>,
    advancedGenerate: ImageContentAdvancedGenerateDTO,
  ): string {
    const { category, currency, description, name, price } = product;
    const {
      attachProductCategory,
      attachProductDescription,
      attachProductName,
      attachProductPrice,
    } = this.buildAttachAdvancedGenerate(advancedGenerate).productKnowledge;

    return `
PRODUCT KNOWLEDGE (filtered):
${attachProductName && name ? `- Name: ${name}` : ""}
${
  attachProductDescription && description ? `- Description: ${description}` : ""
}
${attachProductCategory && category ? `- Category: ${category}` : ""}
${
  attachProductPrice && typeof price === "number"
    ? `- Price: ${currency ?? ""} ${price}`
    : ""
}
CATATAN:
- Gunakan detail yang memperkuat **fit** terhadap referensi/template & brand.
`.trim();
  }

  protected buildPromptRoleKnowledge(
    role: Partial<RoleKnowledge>,
    advancedGenerate: ImageContentAdvancedGenerateDTO,
  ): string {
    const {
      audiencePersona,
      callToAction,
      goals,
      hashtags,
      targetAudience,
      tone,
    } = role;

    const { attachRoleHashtags } =
      this.buildAttachAdvancedGenerate(advancedGenerate).roleKnowledge;

    return `
ROLE KNOWLEDGE:
${audiencePersona ? `- Persona: ${audiencePersona}` : ""}
${callToAction ? `- CTA: ${callToAction}` : ""}
${goals ? `- Goals: ${goals}` : ""}
${targetAudience ? `- Target Audience: ${targetAudience}` : ""}
${tone ? `- Tone: ${tone}` : ""}
${
  attachRoleHashtags && hashtags?.length
    ? `- Hashtags (fixed): ${hashtags.map((h) => `#${h}`).join(", ")}`
    : ""
}
CATATAN:
- Tone & persona membimbing mood visual (color/contrast/keberanian layout).
`.trim();
  }

  private buildPromptImageContent(body: ImageContentDTO): string {
    const { category, designStyle, ratio } = body;
    return `
IMAGE CONTENT:
${category ? `- Category: ${category}` : ""}
${designStyle ? `- Design Style: ${designStyle}` : ""}
${ratio ? `- Ratio: ${ratio}` : ""}
CATATAN:
- Perlakukan ini sebagai constraint utama desain.
`.trim();
  }

  protected buildPromptBusinessKnowledge(
    business: Partial<BusinessKnowledge>,
    advancedGenerate: ImageContentAdvancedGenerateDTO,
  ): string {
    const {
      category,
      description,
      location,
      name,
      uniqueSellingPoint,
      visionMission,
      website,
      colorTone,
    } = business;

    const {
      attachBusinessCategory,
      attachBusinessDescription,
      attachBusinessName,
      attachBusinessUniqueSellingPoint,
      attachBusinessVisionMission,
      attachBusinessWebsite,
      attachBusinessLocation,
      attachBusinessColorTone,
    } = this.buildAttachAdvancedGenerate(advancedGenerate).businessKnowledge;

    return `
BUSINESS KNOWLEDGE (filtered):
${attachBusinessName && name ? `- Name: ${name}` : ""}
${
  attachBusinessDescription && description
    ? `- Description: ${description}`
    : ""
}
${attachBusinessCategory && category ? `- Category: ${category}` : ""}
${attachBusinessLocation && location ? `- Location: ${location}` : ""}
${
  attachBusinessUniqueSellingPoint && uniqueSellingPoint
    ? `- USP: ${uniqueSellingPoint}`
    : ""
}
${
  attachBusinessVisionMission && visionMission
    ? `- Vision/Mission: ${visionMission}`
    : ""
}
${attachBusinessWebsite && website ? `- Website: ${website}` : ""}
${
  attachBusinessColorTone && colorTone
    ? `- Brand Color Tone: #${colorTone}`
    : ""
}

CATATAN BRAND:
- Ikuti palet & nuansa warna brand (jika tersedia).
- Logo harus tajam, tidak terdistorsi, dan kontras memadai terhadap background.
`.trim();
  }

  /** Prompt ringkas untuk proses “regenerate” tanpa knowledge block lengkap (opsional). */
  protected buildRegeneratePrompt(data: ImageContentRegenerateDTO): string {
    const { designStyle, prompt, advancedGenerate } = data;
    const attachLogo = advancedGenerate?.businessKnowledge?.logo;
    return `
User wants to edit the attached image with:
- Design Style: ${designStyle ?? "Default"}
- Attach Logo: ${attachLogo ? "Yes" : "No"} 
- User Prompt: ${prompt ?? "N/A"}

NOTE:
- Apply only the requested changes; preserve the rest.
- Keep it sharp, clean, and social-ready.
`.trim();
  }

  /* ------------------------------ INTERNAL HELPERS ----------------------------- */

  /** Aturan komposisi umum (grid, safe area, hierarchy). */
  private buildCompositionRules(): string {
    return `
- Gunakan grid sederhana; jaga keseimbangan ruang negatif.
- Safe margin tepi (±40 px pada kanvas 1080px) untuk menghindari crop di feed.
- Produk adalah fokus visual: kontras, pencahayaan, dan ketajaman diprioritaskan.
- Hindari clutter; maksimalkan keterbacaan dan hierarki visual.
`.trim();
  }

  /** Aturan brand & logo (menghormati identitas visual). */
  private buildBrandRules(business: Partial<BusinessKnowledge>): string {
    const hasColor = Boolean(business?.colorTone);
    return `
- Logo diletakkan proporsional, tidak terdistorsi, dan tetap tajam.
- Hindari overlay efek berat pada logo (blur/emboss/glow berlebihan).
`.trim();
  }

  private buildAttachAdvancedGenerate(
    advancedGenerate: ImageContentAdvancedGenerateDTO,
  ) {
    const {
      category: attachBusinessCategory,
      description: attachBusinessDescription,
      logo,
      name: attachBusinessName,
      uniqueSellingPoint: attachBusinessUniqueSellingPoint,
      visionMission: attachBusinessVisionMission,
      website: attachBusinessWebsite,
      location: attachBusinessLocation,
      colorTone: attachBusinessColorTone,
    } = advancedGenerate.businessKnowledge;
    const attachBusinessLogo = logo;
    const {
      name: attachProductName,
      description: attachProductDescription,
      category: attachProductCategory,
      price: attachProductPrice,
    } = advancedGenerate.productKnowledge;
    const { hashtags: attachRoleHashtags } = advancedGenerate.roleKnowledge;
    return {
      businessKnowledge: {
        attachBusinessCategory,
        attachBusinessDescription,
        attachBusinessLogo,
        attachBusinessName,
        attachBusinessUniqueSellingPoint,
        attachBusinessVisionMission,
        attachBusinessWebsite,
        attachBusinessLocation,
        attachBusinessColorTone,
      },
      productKnowledge: {
        attachProductName,
        attachProductDescription,
        attachProductCategory,
        attachProductPrice,
      },
      roleKnowledge: { attachRoleHashtags },
    };
  }
  /* --------------------------------- UTILS --------------------------------- */

  getTokenUsage(tokens: number[]): number {
    return tokens.reduce((acc, token) => token + acc, 0);
  }
}

export const defaultAdvancedGenerate = {
  businessKnowledge: {
    name: true,
    category: true,
    description: true,
    location: true,
    logo: true,
    uniqueSellingPoint: true,
    website: true,
    visionMission: true,
    colorTone: true,
  },
  productKnowledge: {
    name: true,
    category: true,
    description: true,
    price: true,
  },
  roleKnowledge: {
    hashtags: true,
  },
};
```

yang dimana kurang lebih untuk memanggilnya seperti berikut

```typescript
import OpenAI, { toFile, Uploadable } from "openai";
import { OPENAI_API_KEY } from "../../../constant/openai";
import fs from "fs";
import { encoding_for_model } from "tiktoken";
import {
  BusinessKnowledge,
  ProductKnowledge,
  RoleKnowledge,
} from ".prisma/client";
import {
  ImageContentAdvancedGenerateDTO,
  ImageContentDTO,
  ImageContentMaskDTO,
  ImageContentRegenerateDTO,
  ImageContentRssDTO,
  ValidRatioOpenAi,
} from "src/validators/ImageContentValidator";
import { AiBaseService } from "../AiBaseService";
import { ImageEditParams } from "openai/resources/images";

export class OpenAiImageService extends AiBaseService {
  private static TOKEN_PER_IMAGE = 9000;
  private ai: OpenAI;
  constructor() {
    super();
    this.ai = new OpenAI({
      apiKey: OPENAI_API_KEY,
    });
  }

  async generateImages(params: GenerateImageParams): Promise<{
    image: string;
    usage: OpenAI.Images.ImagesResponse.Usage;
  }> {
    const {
      productImage,
      templateImage,
      logo,
      prompt: additionalPrompt,
      ratio,
      product,
      role,
      body,
      business,
      advancedGenerate,
      referenceImages,
    } = params;

    const attachLogo = advancedGenerate.businessKnowledge.logo;

    const images: Uploadable[] = [
      // PRODUCT IMAGE
      await toFile(fs.createReadStream(productImage), null, {
        type: "image/png",
      }),
    ];

    if (referenceImages) {
      for (const referenceImage of referenceImages) {
        images.push(
          await toFile(fs.createReadStream(referenceImage), null, {
            type: "image/png",
          }),
        );
      }
    }

    if (templateImage) {
      const templateImageFile = await toFile(
        fs.createReadStream(templateImage),
        null,
        { type: "image/png" },
      );
      images.push(templateImageFile);
    }

    if (attachLogo && logo) {
      const logoImage = await toFile(fs.createReadStream(logo), null, {
        type: "image/png",
      });
      images.push(logoImage);
    }

    const negative = this.buildNegativePrompt();
    const finalPrompt = this.buildKnowledgeDesignInstruction({
      body,
      business,
      product,
      role,
      advancedGenerate,
      additionalPrompt: additionalPrompt ?? "",
      // untuk bagian caption rule “jika ada reference image”
      hasReferenceImage: Boolean(body.referenceImage),
      extraNotes: negative,
    });

    const result = await this.ai.images.edit({
      model: "gpt-image-1",
      image: images,
      prompt: finalPrompt,
      size: this.getImageSize(ratio),
      quality: "high",
    });

    const base64Image = result?.data?.[0]?.b64_json ?? "";
    const textTokens = encoding_for_model("gpt-4").encode(finalPrompt).length;

    return {
      image: base64Image,
      usage: {
        input_tokens: textTokens,
        output_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
        total_tokens: OpenAiImageService.TOKEN_PER_IMAGE + textTokens,
        input_tokens_details: {
          image_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
          text_tokens: textTokens,
        },
      },
    };
  }

  async generateImageFromRss(params: GenerateImageRssParams): Promise<{
    image: string;
    usage: OpenAI.Images.ImagesResponse.Usage;
  }> {
    const {
      productImage,
      templateImage,
      logo,
      prompt: additionalPrompt,
      ratio,
      product,
      role,
      body,
      business,
      rssImage,
      rss,
      advancedGenerate,
    } = params;

    const attachLogo = advancedGenerate?.businessKnowledge?.logo;

    const images: Uploadable[] = [
      await toFile(fs.createReadStream(productImage), null, {
        type: "image/png",
      }),
    ];

    if (rssImage) {
      images.push(
        await toFile(fs.createReadStream(rssImage), null, {
          type: "image/png",
        }),
      );
    }

    if (attachLogo && logo) {
      images.push(
        await toFile(fs.createReadStream(logo), null, { type: "image/png" }),
      );
    }

    if (templateImage) {
      images.push(
        await toFile(fs.createReadStream(templateImage), null, {
          type: "image/png",
        }),
      );
    }

    const negative = this.buildNegativePrompt();
    const rssPrompt = this.buildRssPrompt(rss);
    const finalPrompt = this.buildTrendDesignInstruction({
      body,
      business,
      product,
      role,
      advancedGenerate,
      rssSummaryBlock: rssPrompt,
      additionalPrompt: additionalPrompt ?? "",
      hasRssImage: Boolean(rssImage),
      extraNotes: negative,
    });

    const result = await this.ai.images.edit({
      model: "gpt-image-1",
      image: images,
      prompt: finalPrompt,
      size: this.getImageSize(ratio),
      quality: "high",
    });

    const base64Image = result?.data?.[0]?.b64_json ?? "";
    const textTokens = encoding_for_model("gpt-4").encode(finalPrompt).length;

    return {
      image: base64Image,
      usage: {
        input_tokens: textTokens,
        output_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
        total_tokens: OpenAiImageService.TOKEN_PER_IMAGE + textTokens,
        input_tokens_details: {
          image_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
          text_tokens: textTokens,
        },
      },
    };
  }

  async regenerateContent(params: RegenerateImageParams) {
    const { body, business, product, role, logo, advancedGenerate } = params;
    const attachLogo = advancedGenerate?.businessKnowledge?.logo;

    const businessPrompt = this.buildPromptBusinessKnowledge(
      business,
      advancedGenerate,
    );
    const productPrompt = this.buildPromptProductKnowledge(
      product,
      advancedGenerate,
    );
    const rolePrompt = this.buildPromptRoleKnowledge(role, advancedGenerate);

    const finalPrompt = this.buildRegenerateInstruction({
      body,
      businessSummary: businessPrompt,
      productSummary: productPrompt,
      roleSummary: rolePrompt,
    });

    const images: Uploadable[] = [
      await toFile(fs.createReadStream(body.image), null, {
        type: "image/png",
      }),
    ];

    if (attachLogo && logo) {
      images.push(
        await toFile(fs.createReadStream(logo), null, {
          type: "image/png",
        }),
      );
    }

    const result = await this.ai.images.edit({
      model: "gpt-image-1",
      image: images,
      prompt: finalPrompt,
      size: this.getImageSize(body.ratio),
      quality: "high",
    });

    const base64Image = result?.data?.[0]?.b64_json ?? "";
    const textTokens = encoding_for_model("gpt-4").encode(finalPrompt).length;

    return {
      image: base64Image,
      usage: {
        input_tokens: textTokens,
        output_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
        total_tokens: OpenAiImageService.TOKEN_PER_IMAGE + textTokens,
        input_tokens_details: {
          image_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
          text_tokens: textTokens,
        },
      },
    };
  }

  async maskContent(params: MaskImageParams) {
    const { mask, body, referenceImage } = params;
    const { prompt } = body;

    const images: Uploadable[] = [
      await toFile(fs.createReadStream(referenceImage), "base.png", {
        type: "image/png",
      }),
    ];

    const maskImage = await toFile(fs.createReadStream(mask), "mask.png", {
      type: "image/png",
    });

    const finalPrompt = this.buildMaskInstruction(prompt);

    const result = await this.ai.images.edit({
      model: "gpt-image-1",
      image: images,
      prompt: finalPrompt,
      mask: maskImage,
      quality: "high",
      size: this.getImageSize(body.ratio),
    });

    const base64Image = result?.data?.[0]?.b64_json ?? "";
    const textTokens = encoding_for_model("gpt-4").encode(finalPrompt).length;

    return {
      image: base64Image,
      usage: {
        input_tokens: textTokens,
        output_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
        total_tokens: OpenAiImageService.TOKEN_PER_IMAGE + textTokens,
        input_tokens_details: {
          image_tokens: OpenAiImageService.TOKEN_PER_IMAGE,
          text_tokens: textTokens,
        },
      },
    };
  }

  private getImageSize(ratio: ValidRatioOpenAi): ImageEditParams["size"] {
    switch (ratio) {
      case "1:1":
        return "1024x1024";
      case "2:3":
        return "1024x1536";
      case "3:2":
        return "1536x1024";
      default:
        return "1024x1024";
    }
  }
}

/* ------------------------------- TYPE DEFINITIONS ------------------------------ */

interface GenerateImageParams {
  productImage: string; // path to product image
  templateImage: string; // path to template image
  logo: string | null; // path to logo image
  prompt: string | null; // custom prompt for image generation
  ratio: ValidRatioOpenAi;
  product: Partial<ProductKnowledge>;
  role: Partial<RoleKnowledge>;
  body: ImageContentDTO;
  business: Partial<BusinessKnowledge>;
  advancedGenerate: ImageContentAdvancedGenerateDTO;
  referenceImages?: string[];
}

interface GenerateImageRssParams {
  productImage: string; // path to product image
  templateImage: string; // path to template image
  logo: string | null; // path to logo image
  rssImage: string; // path to rss image
  prompt: string | null; // custom prompt for image generation
  ratio: ValidRatioOpenAi;
  product: Partial<ProductKnowledge>;
  role: Partial<RoleKnowledge>;
  body: ImageContentDTO;
  business: Partial<BusinessKnowledge>;
  rss: ImageContentRssDTO["rss"];
  advancedGenerate: ImageContentAdvancedGenerateDTO;
}

interface RegenerateImageParams {
  body: ImageContentRegenerateDTO & { image: string; ratio: ValidRatioOpenAi };
  logo: string | null; // path to logo image
  product: Partial<ProductKnowledge>;
  role: Partial<RoleKnowledge>;
  business: Partial<BusinessKnowledge>;
  advancedGenerate: ImageContentAdvancedGenerateDTO;
}

interface MaskImageParams {
  referenceImage: string; // path to reference image
  mask: string; // path to mask image
  body: ImageContentMaskDTO;
}
```
