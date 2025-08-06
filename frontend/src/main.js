import './style.css';
import './app.css';

import { FindCommonMovies, SetTMDBAPIKey } from '../wailsjs/go/main/App';

// Global variables for managing state
let currentMovies = [];
let currentSort = 'count';

// DOM Elements
const form = document.getElementById('matcher-form');
const mainContainer = document.getElementById('main-container');
const loaderContainer = document.getElementById('loader-container');
const resultsContainer = document.getElementById('results-container');
const movieList = document.getElementById('movie-list');
const topBar = document.getElementById('top-bar');
const errorMessage = document.getElementById('error-message');
const backdrop = document.getElementById('side-panel-backdrop');
const sidePanel = document.getElementById('side-panel');
const noResults = document.getElementById('no-results');

// Form submission handler
form.addEventListener('submit', async function(e) {
    e.preventDefault();
    
    // Get usernames from textarea
    const usernamesRaw = form.querySelector('textarea[name="usernames"]').value;
    const usernames = usernamesRaw.split(/\s+/).filter(name => name.trim() !== '');
    
    if (usernames.length === 0) {
        showError('Please enter at least one username.');
        return;
    }
    
    // Get API key from input and set it in the backend
    const apiKey = document.getElementById('tmdb-api-key').value.trim();
    if (apiKey) {
        try {
            await SetTMDBAPIKey(apiKey);
        } catch (error) {
            console.log('Note: Could not set API key in backend:', error);
        }
    }
    
    // Show loading state
    hideError();
    mainContainer.style.display = 'none';
    loaderContainer.style.display = 'flex';
    
    try {
        // Call Go backend function
        const movies = await FindCommonMovies(usernames);
        
        if (movies && movies.length > 0) {
            currentMovies = movies;
            displayMovies(movies);
            showResults();
        } else {
            showNoResults(usernames);
        }
    } catch (error) {
        console.error('Error finding common movies:', error);
        showError(error.toString().replace('Error: ', ''));
        resetToMainScreen();
    }
});

// Display movies in the UI
function displayMovies(movies) {
    movieList.innerHTML = '';
    
    movies.forEach((movie, index) => {
        const movieCard = createMovieCard(movie, index);
        movieList.appendChild(movieCard);
    });
}

// Create individual movie card
function createMovieCard(movie, index) {
    const li = document.createElement('li');
    li.className = 'movie-card';
    li.dataset.index = index;
    
    // User avatars
    const userAvatarsHtml = movie.users.map(user => 
        `<img class="user-avatar" src="${user.avatar}" title="${user.name}">`
    ).join('');
    
    // Genres (limit to first 3)
    const genresHtml = movie.genres.slice(0, 3).map(genre => 
        `<span class="genre-tag">${genre}</span>`
    ).join('');
    
    // Runtime display
    const runtimeHtml = movie.formatted_runtime 
        ? `<span>
            <svg class="icon" viewBox="0 0 24 24"><path d="M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z"></path><path d="M12.5 7H11v6l5.25 3.15.75-1.23-4.5-2.67z"></path></svg>
            ${movie.formatted_runtime}
        </span>` : '';
    
    // Rating display
    const ratingHtml = movie.rating > 0 
        ? `<span>
            <svg class="icon" viewBox="0 0 24 24"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"></path></svg>
            ${movie.formatted_rating}
        </span>` : '';
    
    // Stremio button
    const stremioHtml = movie.imdb_id 
        ? `<div class="overlay-info-right">
            <a href="stremio://detail/movie/${movie.imdb_id}" class="stremio-play-btn" onclick="event.stopPropagation();">
                <svg viewBox="0 0 1128.96 1044.37">
                    <defs>
                        <style>
                        .cls-1 { fill: #fecd00; }
                        .cls-2 { fill: #030201; }
                        .cls-3 { fill: #fefefe; }
                        </style>
                    </defs>
                    <path class="cls-1" d="M1126.36,868.09c-2.97-246.32-12.75-492.61-30.41-738.34,13.2-180.12-179.61-118.09-294.25-124.27-181.23,4.6-362.38,8.83-543.56,15.75C78.87,33.45-30.71-14.74,7.63,221.92c12.48,186.65,32.01,372.76,48.58,559.11,20.45,146.06-23.7,284.72,182.55,260.6,207.74,2.19,415.68-8.76,622.16-31.61,101.42-22.67,292.48,19.7,265.44-141.92Zm-88.68-293.51c-20.8,150.02-351.94,255.79-479.88,320.83-146.96,83.22-275.85,90.85-311.66-104.93-31.59-122.82-46.22-249.04-61.17-374.79C106.03,23.88,193.1,64.16,499.08,222.89c99.52,62.43,550.12,245.74,538.6,351.69Z"/>
                    <path class="cls-2" d="M1015.62,521.85c-80.01-84-194.83-125.6-294.11-183.04-133.93-69.74-265.32-145.59-404.88-203.36C121.88,39.98,149.64,209.86,173.39,342.55c22.29,110.93,24.11,225.89,49.13,336.69,49.51,307.15,122.9,325.76,391.54,193.39,100.85-60.65,515.41-191.9,401.55-350.78Zm-43.86,46.98c2.95,104.2-348.34,209.48-432.95,268.23-93.12,57.43-228.32,74.46-235.22-68.66-25.82-105.02-33.87-213.07-52.46-319.52-67.4-332.28-40.9-287.37,240.63-169.44,100.41,45.82,196.79,99.45,292.46,154.19,58.14,40.22,174.25,57.92,187.54,135.21Z"/>
                    <path class="cls-3" d="M234.49,191.98c70.58-3.61,134.05,39.55,198.83,62.39,121.99,49.24,236.68,114.58,350.9,179.25,340.6,144.95,160.26,202.7-66.15,318.4-120.52,36.61-251.82,172.63-379.43,109.17-37.96-46.5-35.5-118.82-53.63-175.57-10.96-103.96-31.01-206.57-45.39-309.99-3.53-59.55-44.29-128.32-5.14-183.66Z"/>
                </svg>
            </a>
        </div>` : '';
    
    li.innerHTML = `
        <div class="movie-card-users">
            ${userAvatarsHtml}
        </div>
        <img src="${movie.poster_url}" alt="Poster for ${movie.title}">
        <div class="movie-overlay">
            <div class="overlay-bottom-content">
                <div class="overlay-info-left">
                    <div class="movie-overlay-title">${movie.title}</div>
                    <div class="movie-overlay-meta">
                        <span>${movie.release_year}</span>
                        ${runtimeHtml}
                        ${ratingHtml}
                    </div>
                    <div class="movie-overlay-genres">
                        ${genresHtml}
                    </div>
                </div>
                ${stremioHtml}
            </div>
        </div>
    `;
    
    // Add click event listener
    li.addEventListener('click', () => openMoviePanel(movie));
    
    return li;
}

// Open movie detail panel
function openMoviePanel(movie) {
    // Set background image
    document.getElementById('panel-background').style.backgroundImage = `url(${movie.backdrop_url})`;
    
    // Set movie logo or title
    const logoEl = document.getElementById('panel-logo');
    if (movie.logo_url) {
        logoEl.src = movie.logo_url;
        logoEl.style.display = 'block';
        logoEl.alt = movie.title;
        // Hide text fallback if logo exists
        let textFallback = document.getElementById('panel-logo-text');
        if (textFallback) textFallback.style.display = 'none';
    } else {
        logoEl.style.display = 'none';
        // Show text fallback
        let textFallback = document.getElementById('panel-logo-text');
        if (!textFallback) {
            textFallback = document.createElement('h2');
            textFallback.id = 'panel-logo-text';
            textFallback.style.cssText = 'color: var(--text-primary); font-size: 2rem; font-weight: 900; text-align: center; margin: 1rem 0;';
            logoEl.parentNode.insertBefore(textFallback, logoEl.nextSibling);
        }
        textFallback.textContent = movie.title;
        textFallback.style.display = 'block';
    }
    
    // Set metadata
    document.getElementById('panel-year').textContent = movie.release_year;
    document.getElementById('panel-runtime').querySelector('span').textContent = movie.formatted_runtime;
    document.getElementById('panel-score').querySelector('span').textContent = movie.formatted_rating;
    document.getElementById('panel-overview').textContent = movie.overview;
    
    // Set director
    const directorContainer = document.getElementById('panel-director');
    directorContainer.innerHTML = '';
    if (movie.director && movie.director.id && movie.director.name !== 'N/A') {
        const directorItem = document.createElement('li');
        const directorLink = document.createElement('a');
        directorLink.href = `https://www.themoviedb.org/person/${movie.director.id}`;
        directorLink.target = '_blank';
        directorLink.rel = 'noopener noreferrer';
        directorLink.textContent = movie.director.name;
        directorItem.appendChild(directorLink);
        directorContainer.appendChild(directorItem);
    }
    
    // Set cast
    const castContainer = document.getElementById('panel-cast');
    castContainer.innerHTML = '';
    if (movie.cast) {
        movie.cast.forEach(actor => {
            const actorItem = document.createElement('li');
            const actorLink = document.createElement('a');
            actorLink.href = `https://www.themoviedb.org/person/${actor.id}`;
            actorLink.target = '_blank';
            actorLink.rel = 'noopener noreferrer';
            actorLink.textContent = actor.name;
            actorItem.appendChild(actorLink);
            castContainer.appendChild(actorItem);
        });
    }
    
    // Set genres
    const genresContainer = document.getElementById('panel-genres');
    genresContainer.innerHTML = '';
    if (movie.genres) {
        movie.genres.forEach(genre => {
            const tag = document.createElement('span');
            tag.className = 'genre-tag';
            tag.textContent = genre;
            genresContainer.appendChild(tag);
        });
    }
    
    // Set users
    const usersContainer = document.getElementById('panel-users');
    usersContainer.innerHTML = '';
    if (movie.users) {
        movie.users.forEach(user => {
            const userLink = document.createElement('a');
            userLink.className = 'panel-user-item';
            userLink.href = `https://letterboxd.com/${user.name}/`;
            userLink.target = '_blank';
            userLink.rel = 'noopener noreferrer';

            const avatar = document.createElement('img');
            avatar.className = 'panel-user-avatar';
            avatar.src = user.avatar;
            
            const name = document.createElement('span');
            name.className = 'panel-user-name';
            name.textContent = user.name;

            userLink.appendChild(avatar);
            userLink.appendChild(name);
            usersContainer.appendChild(userLink);
        });
    }
    
    // Set Stremio link
    const stremioLink = document.getElementById('panel-stremio-link');
    if (movie.imdb_id) {
        stremioLink.href = `stremio://detail/movie/${movie.imdb_id}`;
        stremioLink.style.display = 'flex';
    } else {
        stremioLink.style.display = 'none';
    }
    
    // Open panel
    sidePanel.classList.add('is-open');
    backdrop.classList.add('is-open');
}

// Close movie detail panel
function closePanel() {
    sidePanel.classList.remove('is-open');
    backdrop.classList.remove('is-open');
}

// Show results view
function showResults() {
    mainContainer.style.display = 'none';
    loaderContainer.style.display = 'none';
    resultsContainer.style.display = 'block';
    topBar.style.display = 'flex';
    noResults.style.display = 'none';
}

// Show no results view
function showNoResults(usernames) {
    mainContainer.style.display = 'none';
    loaderContainer.style.display = 'none';
    resultsContainer.style.display = 'block';
    topBar.style.display = 'flex';
    noResults.style.display = 'flex';
    document.getElementById('no-results-text').textContent = 
        `No movies were found on at least two watchlists for the users: ${usernames.join(', ')}`;
}

// Reset to main screen
function resetToMainScreen() {
    mainContainer.style.display = 'flex';
    loaderContainer.style.display = 'none';
    resultsContainer.style.display = 'none';
    topBar.style.display = 'none';
    noResults.style.display = 'none';
}

// Show error message
function showError(message) {
    errorMessage.textContent = message;
    errorMessage.style.display = 'block';
}

// Hide error message
function hideError() {
    errorMessage.style.display = 'none';
}

// Reset application
window.resetApp = function() {
    resetToMainScreen();
    closePanel();
    hideError();
    form.reset();
};

// Sort functionality
document.getElementById('sort-controls').addEventListener('click', function(e) {
    const button = e.target.closest('.sort-button');
    if (!button) return;

    const sortKey = button.dataset.sort;
    currentSort = sortKey;

    // Update active button
    document.querySelectorAll('.sort-button').forEach(btn => btn.classList.remove('active'));
    button.classList.add('active');

    // Sort movies
    const sortedMovies = [...currentMovies].sort((a, b) => {
        if (sortKey === 'year') {
            return b.release_date.localeCompare(a.release_date);
        } else if (sortKey === 'count') {
            if (a.count !== b.count) {
                return b.count - a.count;
            }
            return b.rating - a.rating;
        } else {
            return b[sortKey] - a[sortKey];
        }
    });

    displayMovies(sortedMovies);
});

// Panel backdrop click handler
backdrop.addEventListener('click', closePanel);

// Load saved API key on page load
document.addEventListener('DOMContentLoaded', function() {
    const savedApiKey = localStorage.getItem('tmdb-api-key');
    if (savedApiKey) {
        document.getElementById('tmdb-api-key').value = savedApiKey;
    }
});

// Save API key to localStorage when user types
document.getElementById('tmdb-api-key').addEventListener('input', function() {
    if (this.value.trim()) {
        localStorage.setItem('tmdb-api-key', this.value.trim());
    } else {
        localStorage.removeItem('tmdb-api-key');
    }
});

// Remove unused elements from DOM since we're not using them
document.querySelector('#app').remove();
