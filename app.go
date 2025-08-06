package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

// getTMDBAPIKey gets the API key from runtime, environment, or config
func (a *App) getTMDBAPIKey() string {
	// First check if runtime API key is set (from frontend input)
	if a.runtimeAPIKey != "" {
		return a.runtimeAPIKey
	}
	// Fall back to environment/config
	return GetTMDBAPIKey()
}

// Movie represents a movie with all its details
type Movie struct {
	Title           string     `json:"title"`
	URL             string     `json:"url"`
	Rating          float64    `json:"rating"`
	FormattedRating string     `json:"formatted_rating"`
	PosterURL       string     `json:"poster_url"`
	BackdropURL     string     `json:"backdrop_url"`
	LogoURL         string     `json:"logo_url"`
	ReleaseDate     string     `json:"release_date"`
	ReleaseYear     string     `json:"release_year"`
	Runtime         int        `json:"runtime"`
	FormattedRuntime string    `json:"formatted_runtime"`
	Genres          []string   `json:"genres"`
	IMDBID          string     `json:"imdb_id"`
	Overview        string     `json:"overview"`
	Director        Person     `json:"director"`
	Cast            []Person   `json:"cast"`
	Users           []User     `json:"users"`
	Count           int        `json:"count"`
}

// Person represents a director or cast member
type Person struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// User represents a Letterboxd user
type User struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// TMDBMovie represents TMDB movie data
type TMDBMovie struct {
	ID           int    `json:"id"`
	VoteAverage  float64 `json:"vote_average"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	ReleaseDate  string `json:"release_date"`
	Runtime      int    `json:"runtime"`
	Genres       []struct {
		Name string `json:"name"`
	} `json:"genres"`
	IMDBID   string `json:"imdb_id"`
	Overview string `json:"overview"`
	Credits  struct {
		Crew []struct {
			Name string `json:"name"`
			Job  string `json:"job"`
			ID   int    `json:"id"`
		} `json:"crew"`
		Cast []struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		} `json:"cast"`
	} `json:"credits"`
	Images struct {
		Logos []struct {
			FilePath    string  `json:"file_path"`
			ISO6391     *string `json:"iso_639_1"`
		} `json:"logos"`
	} `json:"images"`
}

// TMDBSearchResult represents TMDB search response
type TMDBSearchResult struct {
	Results []struct {
		ID int `json:"id"`
	} `json:"results"`
}

// App struct
type App struct {
	ctx           context.Context
	runtimeAPIKey string // API key set at runtime from frontend
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SetTMDBAPIKey sets the TMDB API key at runtime
func (a *App) SetTMDBAPIKey(apiKey string) error {
	a.runtimeAPIKey = strings.TrimSpace(apiKey)
	return nil
}

// GetUserAvatar fetches the avatar URL for a Letterboxd user
func (a *App) GetUserAvatar(username string) (string, error) {
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	var avatarURL string
	var err error

	c.OnHTML("meta[property='og:image']", func(e *colly.HTMLElement) {
		content := e.Attr("content")
		if content != "" {
			avatarURL = content
		}
	})

	c.OnError(func(r *colly.Response, e error) {
		err = fmt.Errorf("could not fetch profile for '%s': %v", username, e)
	})

	visitErr := c.Visit(fmt.Sprintf("https://letterboxd.com/%s/", username))
	if visitErr != nil {
		return "", fmt.Errorf("could not visit profile for '%s': %v", username, visitErr)
	}

	if err != nil {
		return "", err
	}

	if avatarURL == "" {
		return "", fmt.Errorf("could not find avatar for user '%s'", username)
	}

	return avatarURL, nil
}

// GetWatchlist scrapes a user's Letterboxd watchlist
func (a *App) GetWatchlist(username string) (map[string]string, error) {
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	movies := make(map[string]string)
	var scrapeErr error

	c.OnHTML("li.poster-container", func(e *colly.HTMLElement) {
		posterDiv := e.ChildAttr("div.film-poster", "data-target-link")
		img := e.ChildAttr("div.film-poster img", "alt")
		
		if img != "" && posterDiv != "" {
			title := img
			fullURL := fmt.Sprintf("https://letterboxd.com%s", posterDiv)
			movies[title] = fullURL
		}
	})

	c.OnHTML("a.next", func(e *colly.HTMLElement) {
		nextHref := e.Attr("href")
		if nextHref != "" {
			time.Sleep(500 * time.Millisecond) // Rate limiting
			nextURL := fmt.Sprintf("https://letterboxd.com%s", nextHref)
			e.Request.Visit(nextURL)
		}
	})

	c.OnError(func(r *colly.Response, e error) {
		scrapeErr = e
	})

	startURL := fmt.Sprintf("https://letterboxd.com/%s/watchlist/", username)
	err := c.Visit(startURL)
	if err != nil {
		return nil, fmt.Errorf("could not visit watchlist for '%s': %v", username, err)
	}

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	if len(movies) == 0 {
		return nil, fmt.Errorf("no movies found in watchlist for '%s'", username)
	}

	return movies, nil
}

// GetTMDBDetails fetches movie details from TMDB API with improved search logic
func (a *App) GetTMDBDetails(movieTitle string) (TMDBMovie, error) {
	var tmdbData TMDBMovie

	apiKey := a.getTMDBAPIKey()
	if apiKey == "" || len(apiKey) < 10 {
		return tmdbData, fmt.Errorf("TMDB API key not configured")
	}

	originalTitle := movieTitle

	// Extract year from title if present
	var year string
	yearRegex := regexp.MustCompile(`\((\d{4})\)$`)
	matches := yearRegex.FindStringSubmatch(movieTitle)
	if len(matches) > 1 {
		year = matches[1]
		movieTitle = strings.TrimSpace(yearRegex.ReplaceAllString(movieTitle, ""))
	}

	// Try multiple search variations
	searchVariations := []string{movieTitle}
	
	// Add variation without special characters
	cleanTitle := regexp.MustCompile(`[^\w\s]`).ReplaceAllString(movieTitle, "")
	if cleanTitle != movieTitle {
		searchVariations = append(searchVariations, cleanTitle)
	}
	
	// Add variation with common word replacements
	commonReplacements := map[string]string{
		"&": "and",
		"'": "",
		"-": " ",
		":": "",
	}
	altTitle := movieTitle
	for old, new := range commonReplacements {
		altTitle = strings.ReplaceAll(altTitle, old, new)
	}
	altTitle = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(altTitle), " ")
	if altTitle != movieTitle {
		searchVariations = append(searchVariations, altTitle)
	}

	var movieID int
	var searchErr error

	// Try each search variation
	for i, searchTitle := range searchVariations {
		encodedTitle := url.QueryEscape(searchTitle)
		searchURL := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s", apiKey, encodedTitle)
		if year != "" {
			searchURL += "&year=" + year
		}

		log.Printf("TMDB search attempt %d for '%s': %s", i+1, originalTitle, strings.Replace(searchURL, apiKey, "***", 1))
		
		// Add rate limiting
		if i > 0 {
			time.Sleep(250 * time.Millisecond)
		}

		resp, err := http.Get(searchURL)
		if err != nil {
			searchErr = fmt.Errorf("network error: %v", err)
			continue
		}

		if resp.StatusCode == 429 {
			// Rate limited - wait and retry once
			resp.Body.Close()
			log.Printf("Rate limited, waiting 2 seconds...")
			time.Sleep(2 * time.Second)
			resp, err = http.Get(searchURL)
			if err != nil {
				searchErr = fmt.Errorf("retry failed: %v", err)
				continue
			}
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			searchErr = fmt.Errorf("API error: status code %d", resp.StatusCode)
			continue
		}

		var searchResult TMDBSearchResult
		if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
			resp.Body.Close()
			searchErr = fmt.Errorf("parse error: %v", err)
			continue
		}
		resp.Body.Close()

		if len(searchResult.Results) > 0 {
			movieID = searchResult.Results[0].ID
			log.Printf("Found movie '%s' with ID %d on attempt %d", originalTitle, movieID, i+1)
			break
		}
	}

	if movieID == 0 {
		log.Printf("No TMDB results found for '%s' after %d attempts. Last error: %v", originalTitle, len(searchVariations), searchErr)
		return tmdbData, fmt.Errorf("no movie found for: %s", originalTitle)
	}

	// Get detailed movie information with retry
	detailsURL := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s&append_to_response=credits,images", movieID, apiKey)
	
	var resp *http.Response
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		
		resp, err = http.Get(detailsURL)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	
	if err != nil {
		return tmdbData, fmt.Errorf("failed to get movie details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return tmdbData, fmt.Errorf("details API error: status code %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&tmdbData); err != nil {
		return tmdbData, fmt.Errorf("failed to parse movie details: %v", err)
	}

	return tmdbData, nil
}

// TestTMDBAPI tests if the TMDB API key is working
func (a *App) TestTMDBAPI() (string, error) {
	apiKey := a.getTMDBAPIKey()
	if apiKey == "" || len(apiKey) < 10 {
		return "", fmt.Errorf("TMDB API key not configured")
	}

	// Test with a simple search
	testURL := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=interstellar", apiKey)
	resp, err := http.Get(testURL)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to TMDB: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "", fmt.Errorf("Invalid TMDB API key")
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("TMDB API error: status code %d", resp.StatusCode)
	}

	var searchResult TMDBSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return "", fmt.Errorf("Failed to parse TMDB response: %v", err)
	}

	return fmt.Sprintf("TMDB API key is working! Found %d results for 'Interstellar'", len(searchResult.Results)), nil
}

// FindCommonMovies processes usernames and returns common movies with full details
func (a *App) FindCommonMovies(usernames []string) ([]Movie, error) {
	if len(usernames) == 0 {
		return nil, fmt.Errorf("no usernames provided")
	}

	// Validate users and get avatars
	userAvatars := make(map[string]string)
	for _, username := range usernames {
		avatar, err := a.GetUserAvatar(username)
		if err != nil {
			return nil, fmt.Errorf("could not find profile for user: '%s'. The profile may be private or the username is incorrect", username)
		}
		userAvatars[username] = avatar
	}

	// Scrape watchlists concurrently
	type WatchlistResult struct {
		Username string
		Movies   map[string]string
		Error    error
	}

	watchlistChan := make(chan WatchlistResult, len(usernames))
	var wg sync.WaitGroup

	for _, username := range usernames {
		wg.Add(1)
		go func(user string) {
			defer wg.Done()
			movies, err := a.GetWatchlist(user)
			watchlistChan <- WatchlistResult{
				Username: user,
				Movies:   movies,
				Error:    err,
			}
		}(username)
	}

	wg.Wait()
	close(watchlistChan)

	// Process results
	scrapedData := make(map[string]map[string]string)
	validUsers := []string{}

	for result := range watchlistChan {
		if result.Error != nil {
			return nil, fmt.Errorf("could not find a public watchlist for user: '%s'. The profile may be private, empty, or the username is incorrect", result.Username)
		}
		scrapedData[result.Username] = result.Movies
		validUsers = append(validUsers, result.Username)
	}

	// Find common movies
	movieCounts := make(map[string]struct {
		Users []string
		URL   string
	})

	for user, watchlist := range scrapedData {
		for movieTitle, movieURL := range watchlist {
			if _, exists := movieCounts[movieTitle]; !exists {
				movieCounts[movieTitle] = struct {
					Users []string
					URL   string
				}{
					Users: []string{user},
					URL:   movieURL,
				}
			} else {
				data := movieCounts[movieTitle]
				data.Users = append(data.Users, user)
				movieCounts[movieTitle] = data
			}
		}
	}

	// Process movies with 2+ users and get TMDB details
	var processedMovies []Movie
	for title, data := range movieCounts {
		if len(data.Users) >= 2 {
			tmdbDetails, err := a.GetTMDBDetails(title)
			
			var movie Movie
			movie.Title = title
			movie.URL = data.URL
			movie.Count = len(data.Users)

			// Create user objects
			for _, username := range data.Users {
				movie.Users = append(movie.Users, User{
					Name:   username,
					Avatar: userAvatars[username],
				})
			}

			if err != nil {
				log.Printf("Could not fetch TMDB details for '%s': %v", title, err)
				// Set default values
				movie.Rating = 0.0
				movie.FormattedRating = "N/A"
				movie.PosterURL = "https://placehold.co/500x750/1f1f1f/ffffff?text=No+Poster"
				movie.BackdropURL = movie.PosterURL
				movie.LogoURL = ""
				movie.ReleaseDate = "0000-00-00"
				movie.ReleaseYear = "----"
				movie.Runtime = 0
				movie.FormattedRuntime = ""
				movie.Genres = []string{}
				movie.IMDBID = ""
				movie.Overview = "No overview available."
				movie.Director = Person{Name: "N/A", ID: 0}
				movie.Cast = []Person{}
			} else {
				// Process TMDB data
				movie.Rating = tmdbDetails.VoteAverage
				if movie.Rating > 0 {
					movie.FormattedRating = fmt.Sprintf("%.1f", movie.Rating)
				} else {
					movie.FormattedRating = "N/A"
				}

				if tmdbDetails.PosterPath != "" {
					movie.PosterURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", tmdbDetails.PosterPath)
				} else {
					movie.PosterURL = "https://placehold.co/500x750/1f1f1f/ffffff?text=No+Poster"
				}

				if tmdbDetails.BackdropPath != "" {
					movie.BackdropURL = fmt.Sprintf("https://image.tmdb.org/t/p/original%s", tmdbDetails.BackdropPath)
				} else {
					movie.BackdropURL = movie.PosterURL
				}

				// Find logo
				logoPath := ""
				noLangLogoPath := ""
				for _, logo := range tmdbDetails.Images.Logos {
					if logo.ISO6391 != nil && *logo.ISO6391 == "en" {
						logoPath = logo.FilePath
						break
					}
					if noLangLogoPath == "" && (logo.ISO6391 == nil || *logo.ISO6391 == "xx") {
						noLangLogoPath = logo.FilePath
					}
				}
				if logoPath == "" {
					logoPath = noLangLogoPath
				}
				if logoPath != "" {
					movie.LogoURL = fmt.Sprintf("https://image.tmdb.org/t/p/original%s", logoPath)
				}

				movie.ReleaseDate = tmdbDetails.ReleaseDate
				if tmdbDetails.ReleaseDate != "" {
					parts := strings.Split(tmdbDetails.ReleaseDate, "-")
					if len(parts) > 0 {
						movie.ReleaseYear = parts[0]
					}
				}
				if movie.ReleaseYear == "" {
					movie.ReleaseYear = "----"
				}

				movie.Runtime = tmdbDetails.Runtime
				if movie.Runtime > 0 {
					movie.FormattedRuntime = fmt.Sprintf("%d min", movie.Runtime)
				}

				// Genres
				for _, genre := range tmdbDetails.Genres {
					movie.Genres = append(movie.Genres, genre.Name)
				}

				movie.IMDBID = tmdbDetails.IMDBID
				movie.Overview = tmdbDetails.Overview
				if movie.Overview == "" {
					movie.Overview = "No overview available."
				}

				// Director
				movie.Director = Person{Name: "N/A", ID: 0}
				for _, crew := range tmdbDetails.Credits.Crew {
					if crew.Job == "Director" {
						movie.Director = Person{Name: crew.Name, ID: crew.ID}
						break
					}
				}

				// Cast (first 5)
				for i, cast := range tmdbDetails.Credits.Cast {
					if i >= 5 {
						break
					}
					movie.Cast = append(movie.Cast, Person{Name: cast.Name, ID: cast.ID})
				}
			}

			processedMovies = append(processedMovies, movie)
		}
	}

	// Sort by count (descending) then by rating (descending)
	sort.Slice(processedMovies, func(i, j int) bool {
		if processedMovies[i].Count != processedMovies[j].Count {
			return processedMovies[i].Count > processedMovies[j].Count
		}
		return processedMovies[i].Rating > processedMovies[j].Rating
	})

	return processedMovies, nil
}
