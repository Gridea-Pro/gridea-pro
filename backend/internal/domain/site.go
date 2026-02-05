package domain

// SiteData aggregates data for the frontend (View Model)
type SiteData struct {
	AppDir             string                 `json:"appDir"`
	Posts              []interface{}          `json:"posts"`
	Tags               []interface{}          `json:"tags"`
	Menus              []interface{}          `json:"menus"`
	Categories         []interface{}          `json:"categories"`
	Links              []interface{}          `json:"links"`
	ThemeConfig        ThemeConfig            `json:"themeConfig"`
	ThemeCustomConfig  map[string]interface{} `json:"themeCustomConfig"`
	CurrentThemeConfig []interface{}          `json:"currentThemeConfig"`
	Themes             []interface{}          `json:"themes"`
	Setting            Setting                `json:"setting"`
}
