export interface FeedItem {
  time: string;
  title: string;
  description?: string;
  descriptionTr?: string;
  hook?: string;
  summary?: string;
  importance?: string;
  sentiment?: string;
  category: string;
  source: string;
  link: string;
  virality?: number;
  clusterCount?: number;
}
