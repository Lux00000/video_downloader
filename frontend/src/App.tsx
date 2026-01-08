import { useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Download, AlertCircle, RefreshCw, Youtube, Clock, CheckCircle } from 'lucide-react';
import { useDisclaimer } from './hooks/useDisclaimer';
import {
  DisclaimerModal,
  UrlInput,
  VideoPreview,
  FormatSelector,
  DownloadButton,
} from './components';
import { analyzeUrl, downloadFile } from './api/client';
import type { VideoInfo, Format, AppState } from './types';

function App() {
  const { accepted, accept } = useDisclaimer();
  const [state, setState] = useState<AppState>('idle');
  const [video, setVideo] = useState<VideoInfo | null>(null);
  const [selectedFormat, setSelectedFormat] = useState<Format | null>(null);
  const [currentUrl, setCurrentUrl] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [downloadProgress, setDownloadProgress] = useState(0);

  const handleAnalyze = useCallback(async (url: string) => {
    setError(null);
    setState('analyzing');
    setVideo(null);
    setSelectedFormat(null);
    setCurrentUrl(url);
    setDownloadProgress(0);

    try {
      const info = await analyzeUrl(url);
      setVideo(info);
      const firstVideo = info.formats.find((f) => f.type === 'video');
      const firstAudio = info.formats.find((f) => f.type === 'audio');
      setSelectedFormat(firstVideo || firstAudio || info.formats[0] || null);
      setState('ready');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка');
      setState('error');
    }
  }, []);

  const handleDownload = useCallback(async () => {
    if (!currentUrl || !selectedFormat) return;

    setState('downloading');
    setDownloadProgress(0);

    try {
      await downloadFile(currentUrl, selectedFormat.id, selectedFormat.type, (progress) => {
        setDownloadProgress(progress);
      });
      
      // Download complete!
      setState('ready');
      setDownloadProgress(0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка скачивания');
      setState('error');
    }
  }, [currentUrl, selectedFormat]);

  const handleReset = () => {
    setState('idle');
    setVideo(null);
    setSelectedFormat(null);
    setCurrentUrl('');
    setError(null);
    setDownloadProgress(0);
  };

  if (accepted === null) {
    return (
      <div className="min-h-screen bg-gradient-main flex items-center justify-center">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ repeat: Infinity, duration: 1, ease: 'linear' }}
          className="w-8 h-8 border-2 border-cyan-500 border-t-transparent rounded-full"
        />
      </div>
    );
  }

  if (!accepted) {
    return <DisclaimerModal onAccept={accept} />;
  }

  return (
    <div className="min-h-screen bg-gradient-main">
      <div className="min-h-screen flex flex-col items-center px-6 py-8">
        {/* Spacer top */}
        <div className="flex-1 min-h-8 max-h-24" />

        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
          className="text-center mb-10"
        >
          <motion.div 
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.2, type: 'spring', stiffness: 200 }}
            className="inline-flex items-center justify-center w-20 h-20 rounded-3xl bg-gradient-to-br from-cyan-500 to-blue-600 mb-6 shadow-2xl shadow-cyan-500/30"
          >
            <Download className="w-10 h-10 text-white" />
          </motion.div>
          
          <motion.h1 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.3 }}
            className="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight"
          >
            Video Downloader
          </motion.h1>
          
          <motion.p 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 }}
            className="text-lg text-gray-400"
          >
            Скачивайте видео с популярных платформ
          </motion.p>
        </motion.div>

        {/* Supported Platforms */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5, duration: 0.5 }}
          className="flex flex-wrap justify-center gap-4 mb-20"
        >
          {/* YouTube - Active */}
          <div className="flex items-center justify-center gap-2 h-10 px-4 rounded-xl bg-white/5 border border-white/10">
            <Youtube className="w-4 h-4 text-red-500" />
            <span className="text-white font-medium text-xs">YouTube</span>
            <CheckCircle className="w-3.5 h-3.5 text-green-400" />
          </div>

          {/* Instagram - Coming Soon */}
          <div className="flex items-center justify-center gap-2 h-10 px-4 rounded-xl bg-white/5 border border-white/10 opacity-50">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="url(#ig-grad2)">
              <defs>
                <linearGradient id="ig-grad2" x1="0%" y1="100%" x2="100%" y2="0%">
                  <stop offset="0%" stopColor="#FFDC80" />
                  <stop offset="50%" stopColor="#F77737" />
                  <stop offset="100%" stopColor="#C13584" />
                </linearGradient>
              </defs>
              <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838c-3.403 0-6.162 2.759-6.162 6.162s2.759 6.163 6.162 6.163 6.162-2.759 6.162-6.163c0-3.403-2.759-6.162-6.162-6.162zm0 10.162c-2.209 0-4-1.79-4-4 0-2.209 1.791-4 4-4s4 1.791 4 4c0 2.21-1.791 4-4 4zm6.406-11.845c-.796 0-1.441.645-1.441 1.44s.645 1.44 1.441 1.44c.795 0 1.439-.645 1.439-1.44s-.644-1.44-1.439-1.44z" />
            </svg>
            <span className="text-gray-400 font-medium text-xs">Instagram</span>
            <Clock className="w-3.5 h-3.5 text-amber-400" />
          </div>

          {/* TikTok - Coming Soon */}
          <div className="flex items-center justify-center gap-2 h-10 px-4 rounded-xl bg-white/5 border border-white/10 opacity-50">
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="white">
              <path d="M19.59 6.69a4.83 4.83 0 0 1-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 0 1-5.2 1.74 2.89 2.89 0 0 1 2.31-4.64 2.93 2.93 0 0 1 .88.13V9.4a6.84 6.84 0 0 0-1-.05A6.33 6.33 0 0 0 5 20.1a6.34 6.34 0 0 0 10.86-4.43v-7a8.16 8.16 0 0 0 4.77 1.52v-3.4a4.85 4.85 0 0 1-1-.1z" />
            </svg>
            <span className="text-gray-400 font-medium text-xs">TikTok</span>
            <Clock className="w-3.5 h-3.5 text-amber-400" />
          </div>
        </motion.div>

        {/* URL Input */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6, duration: 0.5 }}
          className="w-full max-w-xl mt-4"
        >
          <UrlInput
            onAnalyze={handleAnalyze}
            isLoading={state === 'analyzing'}
            disabled={state === 'downloading'}
          />
        </motion.div>

        {/* Error Message */}
        <AnimatePresence>
          {state === 'error' && error && (
            <motion.div
              initial={{ opacity: 0, y: -10, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -10, scale: 0.95 }}
              className="w-full max-w-xl mt-6 p-5 rounded-2xl bg-red-500/10 border border-red-500/30 flex items-center gap-4"
            >
              <div className="p-2 rounded-xl bg-red-500/20 flex-shrink-0">
                <AlertCircle className="w-5 h-5 text-red-400" />
              </div>
              <p className="text-red-300 flex-1 text-sm">{error}</p>
              <button
                onClick={handleReset}
                className="p-2 hover:bg-red-500/20 rounded-xl transition-colors flex-shrink-0"
              >
                <RefreshCw className="w-5 h-5 text-red-400" />
              </button>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Video Info and Format Selection */}
        <AnimatePresence>
          {video && state !== 'error' && (
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -30 }}
              transition={{ duration: 0.5, ease: 'easeOut' }}
              className="w-full max-w-xl mt-12"
            >
              {/* Video Preview */}
              <div className="mb-8">
                <VideoPreview video={video} />
              </div>

              {/* Format Selector */}
              <motion.div 
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
                className="glass-card rounded-2xl p-6 mb-8"
              >
                <h3 className="text-lg font-semibold text-white mb-6">Выберите формат</h3>
                <FormatSelector
                  formats={video.formats}
                  selectedFormat={selectedFormat}
                  onSelect={setSelectedFormat}
                />
              </motion.div>

              {/* Download Button */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
                className="mb-6"
              >
                <DownloadButton
                  onClick={handleDownload}
                  disabled={!selectedFormat}
                  isDownloading={state === 'downloading'}
                  progress={downloadProgress}
                />
              </motion.div>

              {/* New Download Button */}
              <motion.button
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.4 }}
                onClick={handleReset}
                className="w-full py-4 text-gray-400 hover:text-white transition-colors flex items-center justify-center gap-2 rounded-2xl hover:bg-white/5"
              >
                <RefreshCw className="w-4 h-4 flex-shrink-0" />
                <span className="text-sm font-medium">Скачать другое видео</span>
              </motion.button>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Spacer bottom */}
        <div className="flex-1 min-h-12 max-h-32" />

        {/* Footer */}
        <motion.footer
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.8 }}
          className="pb-6"
        >
          <p className="text-gray-500 text-sm font-medium tracking-wide">
            by goudini
          </p>
        </motion.footer>
      </div>
    </div>
  );
}

export default App;
