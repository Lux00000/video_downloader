# Video Downloader

Веб-приложение для скачивания видео с популярных платформ.

![Video Downloader](https://img.shields.io/badge/version-1.0.0-blue)

## Поддерживаемые платформы

- ✅ YouTube (включая YouTube Music)
- ⏳ Instagram (в разработке)
- ⏳ TikTok (в разработке)

## Возможности

- 🎬 Скачивание видео в различных качествах (360p - 1080p)
- 🎵 Скачивание только аудио
- 🖼️ Превью видео перед скачиванием
- 📊 Отображение прогресса загрузки
- 🔒 HTTPS поддержка
- ⏱️ Без ограничений по длительности видео

## Технологии

### Backend
- Go 1.21+
- Chi router
- yt-dlp

### Frontend
- React 18
- TypeScript
- Tailwind CSS
- Framer Motion
- Vite

## Установка

### Требования
- Go 1.21+
- Node.js 18+
- yt-dlp
- ffmpeg
- Nginx (для production)

### Backend

```bash
cd backend
go mod download
go build -o viddown .
./viddown
```

### Frontend

```bash
cd frontend
npm install
npm run dev      # Development
npm run build    # Production
```

### Nginx конфигурация

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.crt;
    ssl_certificate_key /path/to/cert.key;

    location /download {
        alias /opt/viddown/frontend/dist;
        try_files $uri $uri/ /download/index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_read_timeout 600s;
        proxy_buffering off;
    }
}
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| PORT | 8080 | Порт API сервера |
| YTDLP_PATH | yt-dlp | Путь к yt-dlp |
| MAX_CONCURRENT | 3 | Макс. параллельных загрузок |
| RATE_LIMIT_RPM | 10 | Лимит запросов в минуту |

## API Endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | /api/health | Проверка статуса |
| POST | /api/analyze | Анализ видео по URL |
| GET | /api/download | Скачивание видео |
| GET | /api/thumbnail | Прокси для превью |

## Лицензия

MIT

## Автор

by goudini

