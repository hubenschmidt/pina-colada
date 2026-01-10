package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"agent/internal/config"
)

type VisitorService struct {
	env           string
	pushoverUser  string
	pushoverToken string
	httpClient    *http.Client
	recentIPs     map[string]time.Time
	mu            sync.Mutex
}

type VisitorInfo struct {
	IP                   string
	UserAgent            string
	Referrer             string
	URL                  string   `json:"url"`
	HardwareConcurrency  int      `json:"hardware_concurrency"`
	Platform             string   `json:"platform"`
	Language             string   `json:"language"`
	Languages            []string `json:"languages"`
	Timezone             string   `json:"timezone"`
	ScreenWidth          int      `json:"screen_width"`
	ScreenHeight         int      `json:"screen_height"`
	ColorDepth           int      `json:"color_depth"`
	ViewportWidth        int      `json:"viewport_width"`
	ViewportHeight       int      `json:"viewport_height"`
	DevicePixelRatio     float64  `json:"device_pixel_ratio"`
	TouchSupport         bool     `json:"touch_support"`
	ConnectionType       string   `json:"connection_type"`
	ConnectionDownlink   float64  `json:"connection_downlink"`
	PrefersDark          bool     `json:"prefers_dark"`
	PrefersReducedMotion bool     `json:"prefers_reduced_motion"`
	DNT                  bool     `json:"dnt"`
	CookiesEnabled       bool     `json:"cookies_enabled"`
}

func NewVisitorService(cfg *config.Config) *VisitorService {
	return &VisitorService{
		env:           cfg.Env,
		pushoverUser:  cfg.PushoverUser,
		pushoverToken: cfg.PushoverToken,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		recentIPs:     make(map[string]time.Time),
	}
}

func (s *VisitorService) NotifyVisitor(info VisitorInfo) error {
	if s.pushoverUser == "" || s.pushoverToken == "" {
		return nil
	}

	if s.env == "development" {
		return nil
	}

	if s.isDuplicateVisit(info.IP) {
		return nil
	}

	location := s.getGeoLocation(info.IP)

	if location == "Ossining, New York, United States" {
		return nil
	}

	device := parseDeviceType(info.UserAgent)
	referrer := info.Referrer
	if referrer == "" {
		referrer = "direct"
	}

	screen := fmt.Sprintf("%dx%d @%dbit", info.ScreenWidth, info.ScreenHeight, info.ColorDepth)
	viewport := fmt.Sprintf("%dx%d", info.ViewportWidth, info.ViewportHeight)

	connection := info.ConnectionType
	if connection == "" {
		connection = "unknown"
	}
	if info.ConnectionDownlink > 0 {
		connection = fmt.Sprintf("%s (%.1f Mbps)", connection, info.ConnectionDownlink)
	}

	// Build preferences string
	var prefs []string
	if info.PrefersDark {
		prefs = append(prefs, "dark")
	} else {
		prefs = append(prefs, "light")
	}
	if info.PrefersReducedMotion {
		prefs = append(prefs, "reduced-motion")
	}
	if info.DNT {
		prefs = append(prefs, "DNT")
	}
	if !info.CookiesEnabled {
		prefs = append(prefs, "no-cookies")
	}
	prefsStr := strings.Join(prefs, ", ")

	touch := "No"
	if info.TouchSupport {
		touch = "Yes"
	}

	pageURL := info.URL
	if pageURL == "" {
		pageURL = referrer
	}

	ipLookup := fmt.Sprintf("https://ipinfo.io/%s", info.IP)

	msg := fmt.Sprintf(
		"Location: %s\n"+
			"Device: %s | %s\n"+
			"Screen: %s\n"+
			"Viewport: %s | DPR: %.2fx\n"+
			"Cores: %d | Touch: %s\n"+
			"Lang: %s | TZ: %s\n"+
			"Connection: %s\n"+
			"Prefs: %s\n"+
			"URL: %s\n"+
			"Referrer: %s\n"+
			"IP: <a href=\"%s\">%s</a>",
		location,
		device, info.Platform,
		screen,
		viewport, info.DevicePixelRatio,
		info.HardwareConcurrency, touch,
		info.Language, info.Timezone,
		connection,
		prefsStr,
		pageURL,
		referrer,
		ipLookup, info.IP,
	)

	return s.sendPushover("Landing Page Visitor", msg)
}

func (s *VisitorService) isDuplicateVisit(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if last, ok := s.recentIPs[ip]; ok && time.Since(last) < 5*time.Minute {
		return true
	}
	s.recentIPs[ip] = time.Now()
	return false
}

func (s *VisitorService) getGeoLocation(ip string) string {
	if isPrivateIP(ip) {
		return "Local/Private"
	}

	resp, err := s.httpClient.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=city,regionName,country", ip))
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()

	var geo struct {
		City    string `json:"city"`
		Region  string `json:"regionName"`
		Country string `json:"country"`
	}
	if json.NewDecoder(resp.Body).Decode(&geo) != nil {
		return "Unknown"
	}

	return fmt.Sprintf("%s, %s, %s", geo.City, geo.Region, geo.Country)
}

func (s *VisitorService) sendPushover(title, message string) error {
	form := url.Values{
		"token":   {s.pushoverToken},
		"user":    {s.pushoverUser},
		"title":   {title},
		"message": {message},
		"html":    {"1"},
	}

	resp, err := s.httpClient.PostForm("https://api.pushover.net/1/messages.json", form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pushover status %d", resp.StatusCode)
	}
	return nil
}

func parseDeviceType(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") {
		return "Mobile"
	}
	if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") {
		return "Tablet"
	}
	return "Desktop"
}

func isPrivateIP(ip string) bool {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" {
		return true
	}
	return strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.")
}
