package models

// PhoneChannel represents the electronic channel for phone services.
type PhoneChannel struct {
	PhoneAvailableServices []string `json:"phoneAvailableServices,omitempty"` // Servicios disponibles Telefono
	PhoneNumber            string   `json:"phoneNumber,omitempty"`            // Número Telefónico
	PhoneAttentionHours    string   `json:"phoneAttentionHours,omitempty"`    // Horario de atención Telefónica
}

// SMSChannel represents the electronic channel for SMS services.
type SMSChannel struct {
	SMSAvailableServices     []string `json:"smsAvailableServices,omitempty"`     // Servicios disponibles SMS
	SMSAvailableServicesCode []string `json:"smsAvailableServicesCode,omitempty"` // Casilla SMS
	SMSAttentionHours        string   `json:"smsAttentionHours,omitempty"`        // Horario de atención SMS
}

// ElectronicChannels aggregates the phone and SMS channels.
type ElectronicChannels struct {
	PhoneChannel PhoneChannel `json:"phoneChannel,omitempty"` // Phone channel data
	SMSChannel   SMSChannel   `json:"smsChannel,omitempty"`   // SMS channel data
}
