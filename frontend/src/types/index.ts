export interface Format {
  id: string;
  type: 'audio' | 'video' | 'video_only';
  quality: string;
  ext: string;
  size?: number;
}

export interface VideoInfo {
  platform: 'youtube' | 'instagram' | 'tiktok';
  title: string;
  duration: number;
  thumbnail: string;
  formats: Format[];
}

export interface ConfigResponse {
  authRequired: boolean;
  maxConcurrent: number;
  platforms: string[];
}

export interface AnalyzeRequest {
  url: string;
}

export interface ErrorResponse {
  error: string;
}

export type AppState = 'idle' | 'analyzing' | 'ready' | 'downloading' | 'error';


