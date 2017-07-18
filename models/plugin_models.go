package models

type PluginManifest struct {
	Id          string          `json:"id"`
	Version     string          `json:"version"`
	Build       int             `json:"build"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Author      string          `json:"author"`
	AuthorUrl   string          `json:"authorUrl"`
	AuthorEmail string          `json:"authorEmail"`
	Services    PluginServices  `json:"services"`
	Interface   PluginInterface `json:"interface"`
}

type PluginManifestRoute struct {
	Name   string `json:"name"`
	Route  string `json:"route"`
	Method string `json:"method"`
	Url    string `json:"url"`
}

type PluginServices struct {
	Routes []*PluginManifestRoute `json:"routes"`
	Bin    string                 `json:"bin"`
	Docs   string                 `json:"docs"`
}
type PluginInterface struct {
	Public       string `json:"public"`
	PublicVendor string `json:"publicVendor"`
	PublicStyle  string `json:"publicStyle"`
	Admin        string `json:"admin"`
	AdminVendor  string `json:"adminVendor"`
	AdminStyle   string `json:"adminStyle"`
}
