package models

import "time"

type ATM struct {
	ID                             int       `json:"id"`
	ATMIdentifier                  string    `json:"atmidentifier"`
	ATMAddressStreetName           string    `json:"atmaddress_streetname"`
	ATMAddressBuildingNumber       string    `json:"atmaddress_buildingnumber"`
	ATMTownName                    string    `json:"atmtownname"`
	ATMDistrictName                string    `json:"atmdistrictname"`
	ATMCountrySubdivisionMajorName string    `json:"atmcountrysubdivisionmajorname"`
	ATMFromDateTime                time.Time `json:"atmfromdatetime"`
	ATMToDateTime                  time.Time `json:"atmtodatetime"`
	ATMTimeType                    string    `json:"atmtimetype"`
	ATMAttentionHour               string    `json:"atmattentionhour"`
	ATMServiceType                 string    `json:"atmservicetype"`
	ATMAccessType                  string    `json:"atmaccesstype"`
}
