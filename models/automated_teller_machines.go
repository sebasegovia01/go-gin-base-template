package models

type AutomatedTellerMachine struct {
	ATMIdentifier                  string `json:"atmIdentifier"`                            // Max16Text, [1..1]
	StreetName                     string `json:"streetName,omitempty"`                     // Max70Text, [0..1]
	BuildingNumber                 string `json:"buildingNumber,omitempty"`                 // Max16Text, [0..1]
	ATMTownName                    string `json:"atmTownName,omitempty"`                    // Max35Text, [0..1]
	ATMDistrictName                string `json:"atmDistrictName,omitempty"`                // Max35Text, [0..1]
	ATMCountrySubDivisionMajorName string `json:"atmCountrySubDivisionMajorName,omitempty"` // Max35Text, [0..1]
	ATMFromDatetime                string `json:"atmFromDatetime"`                          // ISODateTime, [1..1]
	ATMToDatetime                  string `json:"atmToDatetime"`                            // ISODateTime, [1..1]
	ATMTimeType                    string `json:"atmTimeType,omitempty"`                    // Max35Text (Enumerado), [0..1]
	ATMAttentionHour               string `json:"atmAttentionHour"`                         // BusinessDayCicle, [1..1]
	ATMServiceType                 string `json:"atmServiceType"`                           // ATMServiceTypeCode, [1..1]
	ATMAccessType                  string `json:"atmAccessType"`                            // ATMAccessTypeCode, [1..1]
}
