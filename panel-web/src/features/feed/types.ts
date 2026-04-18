export interface FeedItem {
  time: string;
  title: string;
  description?: string;
  category: string;
  source: string;
  link: string;
  virality?: number;
  clusterCount?: number;
}