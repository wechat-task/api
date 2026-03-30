package ilink

// QRCodeResponse is the response from get_bot_qrcode.
type QRCodeResponse struct {
	QRCode           string `json:"qrcode"`
	QRCodeImgContent string `json:"qrcode_img_content"`
	Ret              int    `json:"ret"`
}

// QRCodeStatusResponse is the response from get_qrcode_status.
type QRCodeStatusResponse struct {
	Ret         int    `json:"ret"`
	Status      string `json:"status"`        // "expired" | "wait" | "confirmed"
	BaseURL     string `json:"baseurl"`       // only when confirmed
	BotToken    string `json:"bot_token"`     // only when confirmed
	ILinkBotID  string `json:"ilink_bot_id"`  // only when confirmed
	ILinkUserID string `json:"ilink_user_id"` // only when confirmed
}

// GetUpdatesRequest is the POST body for getupdates.
type GetUpdatesRequest struct {
	GetUpdatesBuf string   `json:"get_updates_buf"`
	BaseInfo      BaseInfo `json:"base_info"`
}

// BaseInfo contains client metadata.
type BaseInfo struct {
	ChannelVersion string `json:"channel_version"`
}

// GetUpdatesResponse is the response from getupdates.
type GetUpdatesResponse struct {
	Ret                  int             `json:"ret"`
	Msgs                 []WeixinMessage `json:"msgs"`
	GetUpdatesBuf        string          `json:"get_updates_buf"`
	LongpollingTimeoutMs int             `json:"longpolling_timeout_ms"`
}

// WeixinMessage represents a single message.
type WeixinMessage struct {
	FromUserID   string `json:"from_user_id"`
	ToUserID     string `json:"to_user_id"`
	MessageType  int    `json:"message_type"`  // 1=user, 2=bot
	MessageState int    `json:"message_state"` // 2=FINISH
	ContextToken string `json:"context_token"`
	ItemList     []Item `json:"item_list"`
}

// Item represents one content item in a message.
type Item struct {
	Type     int       `json:"type"` // 1=text, 2=image, 3=voice, 4=file, 5=video
	TextItem *TextItem `json:"text_item,omitempty"`
}

// TextItem holds text content.
type TextItem struct {
	Text string `json:"text"`
}

// SendMessageRequest wraps the msg payload for sendmessage.
type SendMessageRequest struct {
	Msg OutboundMessage `json:"msg"`
}

// OutboundMessage is the message payload for sending.
type OutboundMessage struct {
	ToUserID     string `json:"to_user_id"`
	MessageType  int    `json:"message_type"`  // always 2 for outbound
	MessageState int    `json:"message_state"` // always 2 for FINISH
	ContextToken string `json:"context_token"`
	ItemList     []Item `json:"item_list"`
}

// GetUploadURLResponse is the response from getuploadurl.
type GetUploadURLResponse struct {
	Ret       int    `json:"ret"`
	UploadURL string `json:"upload_url"`
	AESKey    string `json:"aes_key"`
}

// APIResponse is a generic envelope.
type APIResponse struct {
	Ret int `json:"ret"`
}
