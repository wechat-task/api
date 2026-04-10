package lark

// WebhookRequest is the body sent to the Lark webhook URL.
type WebhookRequest struct {
	MsgType string `json:"msg_type"`
	Content any    `json:"content"`
	Topic   string `json:"topic,omitempty"`
	Sign    string `json:"sign,omitempty"`
	TS      string `json:"timestamp,omitempty"`
}

// TextContent is the content for a text message.
type TextContent struct {
	Text string `json:"text"`
}

// PostContent is the content for a rich text (post) message.
type PostContent struct {
	Post PostBody `json:"post"`
}

// PostBody wraps language-specific post content.
type PostBody struct {
	ZhCn *PostDetail `json:"zh_cn,omitempty"`
	EnUs *PostDetail `json:"en_us,omitempty"`
}

// PostDetail holds title and content sections.
type PostDetail struct {
	Title   string       `json:"title"`
	Content [][]PostItem `json:"content"`
}

// PostItem is one element in a rich text section.
type PostItem struct {
	Tag  string `json:"tag"`
	Text string `json:"text,omitempty"`
	Href string `json:"href,omitempty"`
}
