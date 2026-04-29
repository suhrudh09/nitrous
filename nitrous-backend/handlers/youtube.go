package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"nitrous-backend/config"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type youtubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title                string `json:"title"`
			Description          string `json:"description"`
			ChannelTitle         string `json:"channelTitle"`
			LiveBroadcastContent string `json:"liveBroadcastContent"`
		} `json:"snippet"`
	} `json:"items"`
}

type youtubeVideosResponse struct {
	Items []struct {
		ID     string `json:"id"`
		Status struct {
			Embeddable bool `json:"embeddable"`
		} `json:"status"`
	} `json:"items"`
}

type openF1VideoResponse struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ChannelTitle string `json:"channelTitle"`
	EmbedURL     string `json:"embedUrl"`
	WatchURL     string `json:"watchUrl"`
	Query        string `json:"query"`
	Mode         string `json:"mode"`
	SessionKey   int    `json:"sessionKey"`
}

type cachedVideo struct {
	data      openF1VideoResponse
	expiresAt time.Time
}

var (
	youtubeCacheMu sync.Mutex
	youtubeCache   = map[string]cachedVideo{}
)

// GetOpenF1SessionVideo resolves the best embeddable YouTube video for live or recent OpenF1 session telemetry.
func GetOpenF1SessionVideo(c *gin.Context) {
	apiKey := strings.TrimSpace(config.AppConfig.YouTubeAPIKey)
	if apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "YouTube API key not configured"})
		return
	}

	mode := strings.ToLower(strings.TrimSpace(c.DefaultQuery("mode", "")))
	rawSessionKey := strings.TrimSpace(c.Query("sessionKey"))
	sessionKey := 0
	if rawSessionKey != "" {
		v, err := strconv.Atoi(rawSessionKey)
		if err != nil || v <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sessionKey"})
			return
		}
		sessionKey = v
	}

	if mode == "" {
		if sessionKey > 0 {
			mode = "recent"
		} else {
			mode = "live"
		}
	}
	if mode != "live" && mode != "recent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mode"})
		return
	}

	session, err := resolveTargetSession(mode, sessionKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	query := buildYouTubeQuery(session, mode)
	cacheKey := fmt.Sprintf("%s:%d:%s", mode, session.SessionKey, query)
	if cached, ok := getCachedVideo(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	resolved, err := findBestYouTubeVideo(apiKey, query, mode, session.SessionKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	setCachedVideo(cacheKey, resolved, mode)
	c.JSON(http.StatusOK, resolved)
}

func resolveTargetSession(mode string, sessionKey int) (openF1Session, error) {
	if sessionKey > 0 {
		session, ok, err := fetchOpenF1SessionByKey(sessionKey)
		if err != nil {
			return openF1Session{}, fmt.Errorf("failed to fetch session metadata")
		}
		if !ok {
			return openF1Session{}, fmt.Errorf("session not found")
		}
		return session, nil
	}

	session, ok, err := fetchOpenF1Session()
	if err != nil {
		return openF1Session{}, fmt.Errorf("failed to fetch live session metadata")
	}
	if !ok {
		return openF1Session{}, fmt.Errorf("no OpenF1 session available")
	}
	if mode == "live" && !isSessionActive(session) {
		return openF1Session{}, fmt.Errorf("no active live session")
	}

	return session, nil
}

func buildYouTubeQuery(session openF1Session, mode string) string {
	year := time.Now().Year()
	if t := parseRFC3339OrZero(session.DateStart); !t.IsZero() {
		year = t.Year()
	}

	parts := []string{"Formula 1", session.SessionName, session.CircuitShortName, session.CountryName, strconv.Itoa(year)}
	if mode == "live" {
		parts = append(parts, "live")
	} else {
		parts = append(parts, "highlights")
	}

	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			cleaned = append(cleaned, p)
		}
	}
	return strings.Join(cleaned, " ")
}

func findBestYouTubeVideo(apiKey, query, mode string, sessionKey int) (openF1VideoResponse, error) {
	base := "https://www.googleapis.com/youtube/v3/search"
	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("type", "video")
	params.Set("maxResults", "10")
	params.Set("order", "relevance")
	params.Set("q", query)
	params.Set("key", apiKey)
	if mode == "live" {
		params.Set("eventType", "live")
	}

	var searchResp youtubeSearchResponse
	if err := fetchJSON(base+"?"+params.Encode(), &searchResp); err != nil {
		if mode == "live" {
			// Fallback to general search when no live event is currently indexed.
			params.Del("eventType")
			if err2 := fetchJSON(base+"?"+params.Encode(), &searchResp); err2 != nil {
				return openF1VideoResponse{}, fmt.Errorf("youtube search failed")
			}
		} else {
			return openF1VideoResponse{}, fmt.Errorf("youtube search failed")
		}
	}

	type candidate struct {
		videoID              string
		title                string
		channel              string
		liveBroadcastContent string
		score                int
	}

	candidates := make([]candidate, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		videoID := strings.TrimSpace(item.ID.VideoID)
		if videoID == "" {
			continue
		}
		title := strings.TrimSpace(item.Snippet.Title)
		channel := strings.TrimSpace(item.Snippet.ChannelTitle)
		liveContent := strings.ToLower(strings.TrimSpace(item.Snippet.LiveBroadcastContent))

		score := 0
		lowerTitle := strings.ToLower(title)
		for _, token := range strings.Fields(strings.ToLower(query)) {
			if len(token) < 3 {
				continue
			}
			if strings.Contains(lowerTitle, token) {
				score += 2
			}
		}
		// Penalise official F1/FOM channels — they block embedding via Content ID
		// even when the YouTube API reports embeddable=true.
		lowerChannel := strings.ToLower(channel)
		isFOMChannel := lowerChannel == "formula 1" || lowerChannel == "f1" ||
			strings.HasPrefix(lowerChannel, "formula 1 ") ||
			strings.Contains(lowerChannel, "formula one management") ||
			strings.Contains(lowerChannel, "fom ")
		if isFOMChannel {
			score -= 20
		}
		// Boost well-known third-party motorsport channels that allow embedding
		knownEmbeddable := []string{"sky sports f1", "channel 4 f1", "espn f1", "motorsport", "wtf1", "racefans", "the race"}
		for _, ch := range knownEmbeddable {
			if strings.Contains(lowerChannel, ch) {
				score += 8
				break
			}
		}
		if mode == "live" && liveContent == "live" {
			score += 10
		}

		candidates = append(candidates, candidate{
			videoID:              videoID,
			title:                title,
			channel:              channel,
			liveBroadcastContent: liveContent,
			score:                score,
		})
	}
	if len(candidates) == 0 {
		return openF1VideoResponse{}, fmt.Errorf("no matching videos found")
	}

	ids := make([]string, 0, len(candidates))
	for _, c := range candidates {
		ids = append(ids, c.videoID)
	}
	embeddable, err := fetchEmbeddableVideoIDs(apiKey, ids)
	if err != nil {
		return openF1VideoResponse{}, fmt.Errorf("youtube video details failed")
	}

	filtered := make([]candidate, 0, len(candidates))
	for _, c := range candidates {
		if embeddable[c.videoID] {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		return openF1VideoResponse{}, fmt.Errorf("no embeddable videos found")
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].score > filtered[j].score
	})

	best := filtered[0]
	return openF1VideoResponse{
		VideoID:      best.videoID,
		Title:        best.title,
		ChannelTitle: best.channel,
		EmbedURL:     fmt.Sprintf("https://www.youtube.com/embed/%s?rel=0", best.videoID),
		WatchURL:     fmt.Sprintf("https://www.youtube.com/watch?v=%s", best.videoID),
		Query:        query,
		Mode:         mode,
		SessionKey:   sessionKey,
	}, nil
}

func fetchEmbeddableVideoIDs(apiKey string, ids []string) (map[string]bool, error) {
	base := "https://www.googleapis.com/youtube/v3/videos"
	params := url.Values{}
	params.Set("part", "status")
	params.Set("id", strings.Join(ids, ","))
	params.Set("key", apiKey)

	var resp youtubeVideosResponse
	if err := fetchJSON(base+"?"+params.Encode(), &resp); err != nil {
		return nil, err
	}

	out := make(map[string]bool, len(resp.Items))
	for _, item := range resp.Items {
		out[item.ID] = item.Status.Embeddable
	}
	return out, nil
}

func getCachedVideo(key string) (openF1VideoResponse, bool) {
	youtubeCacheMu.Lock()
	defer youtubeCacheMu.Unlock()
	cached, ok := youtubeCache[key]
	if !ok {
		return openF1VideoResponse{}, false
	}
	if time.Now().After(cached.expiresAt) {
		delete(youtubeCache, key)
		return openF1VideoResponse{}, false
	}
	return cached.data, true
}

func setCachedVideo(key string, data openF1VideoResponse, mode string) {
	ttl := 24 * time.Hour
	if mode == "live" {
		ttl = 3 * time.Minute
	}
	youtubeCacheMu.Lock()
	youtubeCache[key] = cachedVideo{data: data, expiresAt: time.Now().Add(ttl)}
	youtubeCacheMu.Unlock()
}
