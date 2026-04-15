export type PresetView = {
  id: string;
  label: string;
  description: string;
};

export const PRESET_VIEWS: PresetView[] = [
  {
    id: "turkey-critical",
    label: "Türkiye Kritik Gelişmeler",
    description: "Türkiye’yi doğrudan etkileyen yüksek önemde gelişmeler"
  },
  {
    id: "global-high-impact",
    label: "Küresel Önemli Gelişmeler",
    description: "Yüksek etkili uluslararası gelişmeler"
  },
  {
    id: "economy-markets",
    label: "Ekonomi & Piyasalar",
    description: "Makro, enerji ve piyasa etkili haberler"
  },
  {
    id: "tech-watch",
    label: "Teknoloji",
    description: "Yüksek etkili teknoloji ve ürün lansmanları"
  }
];