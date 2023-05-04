package uploader

type AgentInfo struct {
	TenantName string   `json:"tenantName"`
	AgentName string `json:"agentName"`
}

type AppError struct {
	Message string `json:"message"`
	Error string `json:"error"`
}

type TempUploadUrlResponse struct {
	URL string `json:"url"`
	ObjectKey string `json:"objectKey"`
}

type TempUploadUrlRequest struct {
	Filename string `json:"filename"`
	ContentType string `json:"contentType"`
}
