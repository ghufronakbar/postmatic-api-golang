Berikut legacy code typescript saya yang berhubungan dengan SDK openai/google genai

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

```typescript
// src/services/ai/GeminiImageService.ts
import fs from "fs";
import {
  type ContentListUnion,
  type GenerateContentResponse,
  GoogleGenAI,
  ImageConfig,
} from "@google/genai";
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
  ImageGemini3ProImageSize,
  ValidRatioGemini,
} from "src/validators/ImageContentValidator";
import { AiBaseService } from "../AiBaseService";
import { GEMINI_API_KEY } from "../../../constant/gemini";

/* ------------------------------- TYPE DEFINITIONS ------------------------------ */

interface GenerateImageParams {
  productImage: string; // path to product image
  templateImage: string; // path to template image
  logo: string | null; // path to logo image
  prompt: string | null; // custom prompt for image generation
  ratio: ValidRatioGemini;
  product: Partial<ProductKnowledge>;
  role: Partial<RoleKnowledge>;
  body: ImageContentDTO;
  business: Partial<BusinessKnowledge>;
  advancedGenerate: ImageContentAdvancedGenerateDTO;
  referenceImages?: string[];
}

interface GenerateImageRssParams {
  productImage: string;
  templateImage: string;
  logo: string | null;
  rssImage: string; // path to rss image
  prompt: string | null;
  ratio: ValidRatioGemini;
  product: Partial<ProductKnowledge>;
  role: Partial<RoleKnowledge>;
  body: ImageContentDTO;
  business: Partial<BusinessKnowledge>;
  rss: ImageContentRssDTO["rss"];
  advancedGenerate: ImageContentAdvancedGenerateDTO;
}

interface RegenerateImageParams {
  body: ImageContentRegenerateDTO & { image: string; ratio: ValidRatioGemini };
  logo: string | null;
  product: Partial<ProductKnowledge>;
  role: Partial<RoleKnowledge>;
  business: Partial<BusinessKnowledge>;
  advancedGenerate: ImageContentAdvancedGenerateDTO;
}

interface MaskImageParams {
  referenceImage: string; // path to original image
  mask: string; // path to mask image (biner/alpha)
  body: ImageContentMaskDTO;
}

/** Bentuk usage (mirip OpenAI.Images.ImagesResponse.Usage) */
export interface ImageUsage {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  input_tokens_details: {
    image_tokens: number;
    text_tokens: number;
  };
}

/* --------------------------------- SERVICE ----------------------------------- */

export class GeminiImageService extends AiBaseService {
  private static TOKEN_PER_IMAGE = 9000;

  private ai: GoogleGenAI;

  constructor() {
    super();
    this.ai = new GoogleGenAI({
      apiKey: GEMINI_API_KEY,
      // Untuk Vertex AI:
      // GOOGLE_GENAI_USE_VERTEXAI=true
      // GOOGLE_CLOUD_PROJECT=...
      // GOOGLE_CLOUD_LOCATION=global
    });
  }

  /* ----------------------------- PUBLIC METHODS ----------------------------- */

  async generateImages(
    params: GenerateImageParams,
  ): Promise<{ image: string; usage: ImageUsage }> {
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

    const negative = this.buildNegativePrompt();
    const finalPrompt = this.buildKnowledgeDesignInstruction({
      body,
      business,
      product,
      role,
      advancedGenerate,
      additionalPrompt: additionalPrompt ?? "",
      hasReferenceImage: Boolean(body.referenceImage),
      extraNotes: negative,
    });

    const imageParts = this.buildImageParts(
      [
        { file: productImage },
        templateImage ? { file: templateImage } : null,
        advancedGenerate?.businessKnowledge?.logo && logo
          ? { file: logo }
          : null,
        ...(referenceImages
          ? referenceImages.map((refImg) => ({ file: refImg }))
          : []),
      ].filter(Boolean) as { file: string }[],
    );

    const contents: ContentListUnion = [
      { role: "user", parts: [...imageParts, { text: finalPrompt }] },
    ];

    const res = await this.ai.models.generateContent({
      model: params.body.model,
      contents,
      config: {
        responseModalities: ["IMAGE"],
        imageConfig: {
          aspectRatio: this.aspectFromRatio(ratio),
          imageSize:
            params.body.model === "gemini-3-pro-image-preview"
              ? (params.body.imageSize as ImageGemini3ProImageSize)
              : undefined,
        },
      },
    });

    // fs.writeFileSync(
    //   "test-gemini-image-res.json",
    //   JSON.stringify(res, null, 2)
    // );
    const base64Image = this.pickFirstImageBase64(res) ?? "";
    // fs.writeFileSync("test-gemini-image-base64.txt", base64Image);
    const usage = this.toImageUsage(res, imageParts.length, finalPrompt);
    // fs.writeFileSync(
    //   "test-gemini-image-usage.json",
    //   JSON.stringify(usage, null, 2)
    // );
    return { image: base64Image, usage };
  }

  async generateImageFromRss(
    params: GenerateImageRssParams,
  ): Promise<{ image: string; usage: ImageUsage }> {
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

    const imageParts = this.buildImageParts(
      [
        { file: productImage },
        rssImage ? { file: rssImage } : null,
        templateImage ? { file: templateImage } : null,
        advancedGenerate?.businessKnowledge?.logo && logo
          ? { file: logo }
          : null,
      ].filter(Boolean) as { file: string }[],
    );

    const contents: ContentListUnion = [
      { role: "user", parts: [...imageParts, { text: finalPrompt }] },
    ];

    const res = await this.ai.models.generateContent({
      model: params.body.model,
      contents,
      config: {
        responseModalities: ["IMAGE"],
        imageConfig: {
          aspectRatio: this.aspectFromRatio(ratio),
          imageSize:
            params.body.model === "gemini-3-pro-image-preview"
              ? (params.body.imageSize as ImageGemini3ProImageSize)
              : undefined,
        },
        maxOutputTokens: 100000,
      },
    });

    const base64Image = this.pickFirstImageBase64(res) ?? "";
    const usage = this.toImageUsage(res, imageParts.length, finalPrompt);
    return { image: base64Image, usage };
  }

  async regenerateContent(
    params: RegenerateImageParams,
  ): Promise<{ image: string; usage: ImageUsage }> {
    const { body, business, product, role, logo, advancedGenerate } = params;

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

    const imageParts = this.buildImageParts(
      [
        { file: body.image },
        advancedGenerate?.businessKnowledge?.logo && logo
          ? { file: logo }
          : null,
      ].filter(Boolean) as { file: string }[],
    );

    const contents: ContentListUnion = [
      { role: "user", parts: [...imageParts, { text: finalPrompt }] },
    ];

    const res = await this.ai.models.generateContent({
      model: params.body.model,
      contents,
      config: {
        responseModalities: ["IMAGE"],
        imageConfig: {
          aspectRatio: this.aspectFromRatio(body.ratio),
          imageSize:
            params.body.model === "gemini-3-pro-image-preview"
              ? (params.body.imageSize as ImageGemini3ProImageSize)
              : undefined,
        },
        maxOutputTokens: 100000,
      },
    });

    const base64Image = this.pickFirstImageBase64(res) ?? "";
    const usage = this.toImageUsage(res, imageParts.length, finalPrompt);
    return { image: base64Image, usage };
  }

  async maskContent(
    params: MaskImageParams,
  ): Promise<{ image: string; usage: ImageUsage }> {
    const { mask, body, referenceImage } = params;
    const { prompt } = body;

    // Soft-mask di Gemini (instruksi + 2 gambar).
    // Untuk presisi pixel, rute ke Imagen.
    const imageParts = this.buildImageParts([
      { file: referenceImage },
      { file: mask },
    ]);

    const finalPrompt = this.buildMaskInstruction(
      `${prompt}\n\n[Instruksi]: Gambar kedua adalah MASK biner untuk gambar pertama. Edit hanya area putih/alpha>0 pada MASK; isi latar belakang secara natural.`,
    );

    const contents: ContentListUnion = [
      { role: "user", parts: [...imageParts, { text: finalPrompt }] },
    ];

    const res = await this.ai.models.generateContent({
      model: params.body.model,
      contents,
      config: {
        responseModalities: ["IMAGE"],
        imageConfig: {
          aspectRatio: this.aspectFromRatio(body.ratio),
          imageSize:
            params.body.model === "gemini-3-pro-image-preview"
              ? (params.body.imageSize as ImageGemini3ProImageSize)
              : undefined,
        },
        maxOutputTokens: 100000,
      },
    });

    const base64Image = this.pickFirstImageBase64(res) ?? "";
    const usage = this.toImageUsage(res, imageParts.length, finalPrompt);
    return { image: base64Image, usage };
  }

  /* ----------------------------- PRIVATE HELPERS ---------------------------- */

  /** Baca file → inlineData base64 dengan MIME hasil sniff . */
  private buildImageParts(files: { file: string }[]) {
    return files.map(({ file }) => {
      const data = fs.readFileSync(file);

      const finalMime = this.sniffImageMime(data);
      this.ensureSupportedMime(finalMime, file);

      return {
        inlineData: {
          data: data.toString("base64"),
          mimeType: finalMime,
        },
      };
    });
  }

  /** Ambil base64 gambar pertama dari response. */
  private pickFirstImageBase64(
    res: GenerateContentResponse,
  ): string | undefined {
    const cands = res?.candidates ?? [];
    const parts = cands[0]?.content?.parts ?? [];
    const img = parts.find((p) =>
      p?.inlineData?.mimeType?.startsWith("image/"),
    );
    return img?.inlineData?.data;
  }

  /** Mapping ratio → aspectRatio yang didukung Gemini. */
  private aspectFromRatio(r: ValidRatioGemini): string {
    return r;
  }

  /** Konversi usage SDK → ImageUsage mirip OpenAI. */
  private toImageUsage(
    res: GenerateContentResponse,
    imageCount: number,
    promptText: string,
  ): ImageUsage {
    const um = res?.usageMetadata;
    let textTokens = 0;
    let imageTokens = 0;

    const details = um?.promptTokensDetails as
      | Array<{ modality: string; tokenCount: number }>
      | undefined;

    if (details?.length) {
      for (const d of details) {
        const mod = (d.modality || "").toUpperCase();
        if (mod === "TEXT") textTokens += d.tokenCount || 0;
        if (mod === "IMAGE") imageTokens += d.tokenCount || 0;
      }
    } else {
      // fallback konservatif
      textTokens = Math.ceil((promptText || "").length / 4);
      imageTokens = GeminiImageService.TOKEN_PER_IMAGE * imageCount;
    }

    const inputTokens = um?.promptTokenCount ?? textTokens + imageTokens;
    const outputTokens = um?.candidatesTokenCount ?? 0;
    const totalTokens = um?.totalTokenCount ?? inputTokens + outputTokens;

    return {
      input_tokens: inputTokens,
      output_tokens: outputTokens,
      total_tokens: totalTokens,
      input_tokens_details: {
        image_tokens: imageTokens,
        text_tokens: textTokens,
      },
    };
  }

  // ===== MIME sniffing & validation =====

  private static readonly SUPPORTED_MIME = new Set([
    "image/png",
    "image/jpeg",
    "image/webp",
  ]);

  /** Deteksi MIME dari magic number (tanpa tergantung ekstensi file). */
  private sniffImageMime(buf: Buffer): string {
    // PNG: 89 50 4E 47 0D 0A 1A 0A
    if (
      buf.length >= 8 &&
      buf[0] === 0x89 &&
      buf[1] === 0x50 &&
      buf[2] === 0x4e &&
      buf[3] === 0x47 &&
      buf[4] === 0x0d &&
      buf[5] === 0x0a &&
      buf[6] === 0x1a &&
      buf[7] === 0x0a
    )
      return "image/png";

    // JPEG: FF D8 ...
    if (buf.length >= 3 && buf[0] === 0xff && buf[1] === 0xd8) {
      return "image/jpeg";
    }

    // WEBP: "RIFF" .... "WEBP"
    if (
      buf.length >= 12 &&
      buf[0] === 0x52 &&
      buf[1] === 0x49 &&
      buf[2] === 0x46 &&
      buf[3] === 0x46 &&
      buf[8] === 0x57 &&
      buf[9] === 0x45 &&
      buf[10] === 0x42 &&
      buf[11] === 0x50
    )
      return "image/webp";

    return "application/octet-stream";
  }

  /** Pastikan MIME termasuk yang didukung; kalau tidak, lempar error yang jelas. */
  private ensureSupportedMime(mime: string, filePath: string) {
    if (!GeminiImageService.SUPPORTED_MIME.has(mime)) {
      throw new Error(
        `Unsupported image MIME: ${mime} for file ${filePath}. ` +
          `Convert ke PNG/JPEG/WEBP sebelum dikirim (contoh: sharp(input).png().toFile(...)).`,
      );
    }
  }
}
```
