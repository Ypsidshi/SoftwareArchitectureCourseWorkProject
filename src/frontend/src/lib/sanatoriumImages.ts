import type { Sanatorium } from "@/types";

const fallbackImages = [
  "/images/sanatorium-1.jpg",
  "/images/sanatorium-2.jpg",
  "/images/sanatorium-3.jpg",
  "/images/sanatorium-4.jpg",
  "/images/sanatorium-5.jpg",
  "/images/resort-1.jpg",
];

function hashText(input: string): number {
  let hash = 0;
  for (let i = 0; i < input.length; i += 1) {
    hash = (hash * 31 + input.charCodeAt(i)) >>> 0;
  }
  return hash;
}

export function getSanatoriumCover(item: Sanatorium): string {
  if (item.image_urls?.[0]) return item.image_urls[0];
  const idx = hashText(`${item.id}-${item.city}`) % fallbackImages.length;
  return fallbackImages[idx];
}

export function getSanatoriumFallback(item: Sanatorium, offset = 0): string {
  const idx = (hashText(`${item.id}-${item.city}`) + offset) % fallbackImages.length;
  return fallbackImages[idx];
}

export function getSanatoriumGallery(item: Sanatorium): string[] {
  if (item.image_urls?.length > 1) return item.image_urls.slice(1);
  return [];
}
