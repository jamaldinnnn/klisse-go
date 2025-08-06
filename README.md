# 🎬 Klisse Desktop

**Find common movies from multiple Letterboxd users' watchlists**

A fast, native desktop application that helps you discover movies that everyone in your group wants to watch. Perfect for movie nights, film clubs, or just finding your next shared viewing experience!

## ⚡ Quick Start

### 📥 Download (Recommended)

1. **Go to [Releases](../../releases)**
2. **Download the latest version** for your operating system:
   - **Windows**: `klisse-windows-amd64.exe`
   - **macOS**: `klisse-macos-universal.app` 
   - **Linux**: `klisse-linux-amd64`
3. **Run the app** - that's it! No installation required.

### 🚀 How to Use

1. **Enter Letterboxd usernames** (one per line or space-separated)
2. **Optional**: Add your [TMDB API key](https://www.themoviedb.org/settings/api) for movie posters and details
3. **Click "Find Matches"** - the app will find movies on 2+ watchlists
4. **Browse results** - click any movie for detailed information
5. **Open in Stremio** - click the Stremio button to watch

### 🎯 Features

- **Multi-user support** - Compare 2 or more Letterboxd watchlists
- **Rich movie data** - Posters, ratings, cast, crew, and descriptions
- **Smart search** - Advanced TMDB integration with multiple search strategies
- **Beautiful interface** - Clean, responsive design matching Letterboxd's aesthetic
- **Stremio integration** - One-click movie opening in Stremio
- **Fast & native** - Desktop performance with Go backend
- **No browser required** - Avoids CORS issues of web-based solutions

### 🔑 TMDB API Key (Optional)

For the best experience with movie posters and details:

1. Create a free account at [TMDB](https://www.themoviedb.org/)
2. Get your API key from [Settings → API](https://www.themoviedb.org/settings/api)
3. Enter it in the app's API key field
4. Your key is saved locally and remembered between sessions

**Without an API key**, the app still works but shows placeholder images.

### 🛠️ For Developers

If you want to build from source or contribute:

#### Prerequisites
- **Go 1.21+**
- **Node.js 18+**
- **Wails v2**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

#### Build Commands
```bash
# Development
wails dev

# Production build
wails build

# Cross-platform builds (done automatically via GitHub Actions)
wails build -platform windows/amd64
wails build -platform linux/amd64
wails build -platform darwin/universal
```

### 📋 System Requirements

- **Windows**: Windows 10/11
- **macOS**: macOS 10.15 or later
- **Linux**: Modern Linux distribution with GTK3

### ⚠️ Important Notes

- Only works with **public** Letterboxd watchlists
- Respects Letterboxd's rate limits with built-in delays
- TMDB API has rate limits (handled automatically with retries)

### 🐛 Issues & Support

Found a bug or have a feature request? [Open an issue](../../issues)!

### 📜 License

This project is for educational purposes. Please respect Letterboxd's and TMDB's terms of service.

---

**Original web version**: This desktop app is based on a Flask web application, converted to native desktop for better performance and user experience.