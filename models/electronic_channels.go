package models

// WebChannel represents the electronic channel for web services.
type WebChannel struct {
	WebChannelType       string `json:"webChannelType" validate:"required"`       // Tipo de Canal de Atención Electrónica
	WebURLAddress        string `json:"webURLAddress" validate:"required"`        // Enlace Sitio Web
	WebAvailableServices string `json:"webAvailableServices" validate:"required"` // Servicios disponibles Web y App
	WebAttentionHours    string `json:"webAttentionHours,omitempty"`              // Horario de atención Web y App (Opcional)
	WebPlatformType      string `json:"webPlatformType,omitempty"`                // Tipo de Plataforma Web y App (Opcional)
}

// EmailChannel represents the electronic channel for email services.
type EmailChannel struct {
	EmailAvailableServices string `json:"emailAvailableServices" validate:"required"` // Servicios disponibles Email
	EmailAddress           string `json:"emailAddress" validate:"required"`           // Dirección Email
	EmailAttentionHours    string `json:"emailAttentionHours" validate:"required"`    // Horario de atención Email
}

// SocialMediaChannel represents the electronic channel for social media services.
type SocialMediaChannel struct {
	SocialMediaAvailableServices string `json:"socialMediaAvailableServices,omitempty"` // Servicios disponibles Redes Sociales (Opcional)
	SocialMediaAccount           string `json:"socialMediaAccount,omitempty"`           // Cuenta Red Social (Opcional)
	SocialMediaAttentionHours    string `json:"socialMediaAttentionHours,omitempty"`    // Horario de atención Redes Sociales (Opcional)
}

// ElectronicChannels aggregates the web, email, and social media channels
type ElectronicChannels struct {
	WebChannel         WebChannel         `json:"webChannel,omitempty"`         // Web channel data
	EmailChannel       EmailChannel       `json:"emailChannel,omitempty"`       // Email channel data
	SocialMediaChannel SocialMediaChannel `json:"socialMediaChannel,omitempty"` // Social media channel data
}
