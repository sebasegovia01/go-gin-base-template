package models

type PresentialChannel struct {
	PresentialChannelIdentifier           string `json:"presentialChannelIdentifier"`                     // Max16Text, [1..1]
	PresentialChannelType                 string `json:"presentialChannelType"`                           // Max4Text, [1..1]
	StreetName                            string `json:"streetName,omitempty"`                            // Max70Text, [0..1]
	BuildingNumber                        string `json:"buildingNumber,omitempty"`                        // Max16Text, [0..1]
	PresentialTownName                    string `json:"presentialTownName,omitempty"`                    // Max35Text, [0..1]
	PresentialDistrictName                string `json:"presentialDistrictName,omitempty"`                // Max35Text, [0..1]
	PresentialCountrySubDivisionMajorName string `json:"presentialCountrySubDivisionMajorName,omitempty"` // Max35Text, [0..1]
	PresentialFromDatetime                string `json:"presentialFromDatetime"`                          // ISODateTime, [1..1]
	PresentialToDatetime                  string `json:"presentialToDatetime"`                            // ISODateTime, [1..1]
	PresentialTimeType                    string `json:"presentialTimeType,omitempty"`                    // Max35Text (Enumerado), [0..1]
	PresentialAttentionHours              string `json:"presentialAttentionHours"`                        // BusinessDayCicle, [1..1]
	PresentialAvailableServices           string `json:"presentialAvailableServices,omitempty"`           // SupplementaryData, [0..1]
	PresentialWeekDayCode                 string `json:"presentialWeekDayCode"`                           // Day6Code, [1..7]
}
