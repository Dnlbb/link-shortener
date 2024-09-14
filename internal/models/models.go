package models

type RequestModifyPost struct {
	Body string `json:"url"`
}

type ResponseModifyPost struct {
	Body string `json:"result"`
}
type ResponseModifyConflictPost string

type RequestModifyGet struct {
	Body string `json:"short-url"`
}

type ResponseModifyGet struct {
	Body string `json:"result"`
}

type ReqBatch []MiniBatchReq

type MiniBatchReq struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

type RespBatch []MiniBatchResp

type MiniBatchResp struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

type ResponseToOwner struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UserDelUrls []string
