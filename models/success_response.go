package models

import "github.com/sebasegovia01/base-template-go-gin/enums"

type SuccessResponse struct {
	Result struct {
		Status      enums.ResultStatus `json:"status"`
		Description string             `json:"description,omitempty"`
		Data        interface{}        `json:"data,omitempty"`
	} `json:"Result"`
}
