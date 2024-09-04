package models

import "time"

// PersonalCustomerIdentification representa la información de identificación de un cliente personal
type PersonalCustomerIdentification struct {
	CustomerIdentification string    `json:"customerIdentification"`
	CustomerFirstName      string    `json:"customerFirstName"`
	CustomerMiddleName     string    `json:"customerMiddleName"`
	CustomerLastName       string    `json:"customerLastName"`
	CustomerSecondLastName string    `json:"customerSecondLastName"`
	CustomerInitDate       time.Time `json:"customerInitDate"`
}

// PersonalCustomerAdditionalInfo representa la información adicional de un cliente personal
type PersonalCustomerAdditionalInfo struct {
	CustomerStreetName                  string `json:"customerStreetName,omitempty"`
	CustomerBuildingNumber              string `json:"customerBuildingNumber,omitempty"`
	CustomerDistrictName                string `json:"customerDistrictName,omitempty"`
	CustomerCountrySubDivisionMajorName string `json:"customerCountrySubDivisionMajorName,omitempty"`
	CustomerEmailAddress                string `json:"customerEmailAddress,omitempty"`
	CustomerPhoneNumber                 string `json:"customerPhoneNumber,omitempty"`
	LegalRepresentativeFirstName        string `json:"legalRepresentativeFirstName,omitempty"`
	LegalRepresentativeMiddleName       string `json:"legalRepresentativeMiddleName,omitempty"`
	LegalRepresentativeLastName         string `json:"legalRepresentativeLastName,omitempty"`
	LegalRepresentativeSecondLastName   string `json:"legalRepresentativeSecondLastName,omitempty"`
	LegalRepresentativeIdentification   string `json:"legalRepresentativeIdentification,omitempty"`
}

// LegalEntityIdentification representa la información de identificación de una persona jurídica
type LegalEntityIdentification struct {
	LegalEntityFirstName                string    `json:"legalEntityFirstName"`
	LegalEntityMiddleName               string    `json:"legalEntityMiddleName"`
	LegalEntityLastName                 string    `json:"legalEntityLastName"`
	LegalEntitySecondLastName           string    `json:"legalEntitySecondLastName"`
	LegalEntityPropietaryIdentification string    `json:"legalEntityPropietaryIdentification"`
	LegalEntityBusinessType             string    `json:"legalEntityBusinessType"`
	LegalEntityInitDate                 time.Time `json:"legalEntityInitDate"`
}

// LegalEntityAdditionalInfo representa la información adicional de una persona jurídica
type LegalEntityAdditionalInfo struct {
	LegalEntityStreetName                   string `json:"legalEntityStreetName,omitempty"`
	LegalEntityBuildingNumber               string `json:"legalEntityBuildingNumber,omitempty"`
	LegalEntityDistrictName                 string `json:"legalEntityDistrictName,omitempty"`
	LegalEntityCountrySubDivisionMajorName  string `json:"legalEntityCountrySubDivisionMajorName,omitempty"`
	LegalEntityWebsite                      string `json:"legalEntityWebsite,omitempty"`
	LegalEntityEmailAddress                 string `json:"legalEntityEmailAddress,omitempty"`
	LegalEntityPhoneNumber                  string `json:"legalEntityPhoneNumber,omitempty"`
	LegalEntityRepresentativeFirstName      string `json:"legalEntityRepresentativeFirstName,omitempty"`
	LegalEntityRepresentativeMiddleName     string `json:"legalEntityRepresentativeMiddleName,omitempty"`
	LegalEntityRepresentativeLastName       string `json:"legalEntityRepresentativeLastName,omitempty"`
	LegalEntityRepresentativeSecondLastName string `json:"legalEntityRepresentativeSecondLastName,omitempty"`
	LegalEntityRepresentativeIdentification string `json:"legalEntityRepresentativeIdentification,omitempty"`
}

// DataCustomer engloba todos los tipos de datos de cliente
type Customer struct {
	PersonalIdentification    PersonalCustomerIdentification `json:"personalIdentification"`
	PersonalAdditionalInfo    PersonalCustomerAdditionalInfo `json:"personalAdditionalInfo"`
	LegalEntityIdentification LegalEntityIdentification      `json:"legalEntityIdentification,omitempty"`
	LegalEntityAdditionalInfo LegalEntityAdditionalInfo      `json:"legalEntityAdditionalInfo,omitempty"`
}
