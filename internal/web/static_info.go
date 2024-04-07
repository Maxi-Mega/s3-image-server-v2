package web

type StaticInfo struct {
	WindowTitle            string `json:"window_title"`
	ApplicationTitle       string `json:"application_title"`
	LogoBase64             string `json:"logo_base_64"`
	ScaleInitialPercentage int    `json:"scale_initial_percentage"`
	MaxImagesDisplayCount  int    `json:"max_images_display_count"`
	TileServerURL          string `json:"tile_server_url"`
}
