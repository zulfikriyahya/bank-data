package domain

type CookieInput struct {
	Name           string  `json:"name"`
	Value          string  `json:"value"`
	Domain         string  `json:"domain"`
	Path           string  `json:"path"`
	ExpirationDate float64 `json:"expirationDate"`
	HTTPOnly       bool    `json:"httpOnly"`
	Secure         bool    `json:"secure"`
	SameSite       string  `json:"sameSite"` // "no_restriction" | "lax" | "strict" | "unspecified"
}

type LocalStorageEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ScraperAuthInput struct {
	Cookies      []CookieInput       `json:"cookies"`
	LocalStorage []LocalStorageEntry `json:"localStorage"`
}

// Format yang dipahami Playwright storage_state
type PlaywrightCookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

type PlaywrightOrigin struct {
	Origin       string              `json:"origin"`
	LocalStorage []LocalStorageEntry `json:"localStorage"`
}

type PlaywrightStorageState struct {
	Cookies []PlaywrightCookie `json:"cookies"`
	Origins []PlaywrightOrigin `json:"origins"`
}
