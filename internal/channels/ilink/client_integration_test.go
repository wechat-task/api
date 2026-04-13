package ilink

import (
	"context"
	"fmt"
	"testing"
	"time"
)

const (
	testBotToken = "ee56c30a44c5@im.bot:060000edec8e1a3bdd9ecbaf9b26d11115bc83"
	testBaseURL  = "https://ilinkai.weixin.qq.com"
	testToUserID = "o9cq80_-YQ4SsQEctLny00QWNWd4@im.wechat"
)

func authenticatedTestClient() *Client {
	return NewAuthenticatedClient(testBotToken, testBaseURL)
}

func TestGetConfig(t *testing.T) {
	client := authenticatedTestClient()

	config, err := client.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	t.Logf("GetConfig response: %+v", config)

	if typingTicket, ok := config["typing_ticket"].(string); ok {
		t.Logf("typing_ticket: %s", typingTicket)
	}
}

func TestGetUpdates(t *testing.T) {
	client := authenticatedTestClient()

	resp, err := client.GetUpdates("")
	if err != nil {
		t.Fatalf("GetUpdates failed: %v", err)
	}

	t.Logf("GetUpdates ret=%d, msgs=%d, cursor=%s, timeout_ms=%d",
		resp.Ret, len(resp.Msgs), resp.GetUpdatesBuf, resp.LongpollingTimeoutMs)

	for i, msg := range resp.Msgs {
		t.Logf("  msg[%d]: from=%s, to=%s, type=%d, state=%d, context_token=%s",
			i, msg.FromUserID, msg.ToUserID, msg.MessageType, msg.MessageState, msg.ContextToken)
		for j, item := range msg.ItemList {
			if item.TextItem != nil {
				t.Logf("    item[%d]: type=%d, text=%s", j, item.Type, item.TextItem.Text)
			} else {
				t.Logf("    item[%d]: type=%d", j, item.Type)
			}
		}
	}
}

func TestSendTextMessage(t *testing.T) {
	client := authenticatedTestClient()

	text := fmt.Sprintf("[ilink test] send text message at %s", time.Now().Format(time.RFC3339))

	err := client.SendTextMessage(testToUserID, "", text)
	if err != nil {
		t.Fatalf("SendTextMessage failed: %v", err)
	}

	t.Logf("SendTextMessage sent successfully: %s", text)
}

func TestSendMessage(t *testing.T) {
	client := authenticatedTestClient()

	text := fmt.Sprintf("[ilink test] send message with items at %s", time.Now().Format(time.RFC3339))

	err := client.SendMessage(testToUserID, "", []Item{
		{Type: 1, TextItem: &TextItem{Text: text}},
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	t.Logf("SendMessage sent successfully: %s", text)
}

func TestSendTyping(t *testing.T) {
	client := authenticatedTestClient()

	// First get config to obtain typing_ticket
	config, err := client.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	typingTicket, ok := config["typing_ticket"].(string)
	if !ok || typingTicket == "" {
		t.Skip("No typing_ticket in config, skipping SendTyping test")
	}

	t.Logf("typing_ticket: %s", typingTicket)

	err = client.SendTyping(testToUserID, typingTicket)
	if err != nil {
		t.Fatalf("SendTyping failed: %v", err)
	}

	t.Logf("SendTyping sent successfully")
}

func TestGetUploadURL(t *testing.T) {
	client := authenticatedTestClient()

	resp, err := client.GetUploadURL()
	if err != nil {
		t.Fatalf("GetUploadURL failed: %v", err)
	}

	t.Logf("GetUploadURL: upload_url=%s, aes_key=%s", resp.UploadURL, resp.AESKey)
}

func TestPollLoop_GetMessages(t *testing.T) {
	client := authenticatedTestClient()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var messages []WeixinMessage

	handler := func(msg WeixinMessage) error {
		messages = append(messages, msg)
		t.Logf("Received message: from=%s, type=%d, state=%d, context_token=%s",
			msg.FromUserID, msg.MessageType, msg.MessageState, msg.ContextToken)
		for _, item := range msg.ItemList {
			if item.TextItem != nil {
				t.Logf("  text: %s", item.TextItem.Text)
			}
		}
		return nil
	}

	err := client.PollLoop(ctx, "", handler)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("PollLoop failed: %v", err)
	}

	t.Logf("PollLoop completed, received %d messages", len(messages))
}

func TestFullFlow_ReceiveAndReply(t *testing.T) {
	client := authenticatedTestClient()

	// Step 1: Poll for messages (wait up to 45s)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var lastMsg *WeixinMessage

	t.Log("Waiting for incoming messages (45s timeout)...")
	err := client.PollLoop(ctx, "", func(msg WeixinMessage) error {
		t.Logf("Received: from=%s, type=%d, context_token=%s",
			msg.FromUserID, msg.MessageType, msg.ContextToken)
		for _, item := range msg.ItemList {
			if item.TextItem != nil {
				t.Logf("  text: %s", item.TextItem.Text)
			}
		}
		lastMsg = &msg
		return nil
	})
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("PollLoop failed: %v", err)
	}

	// Step 2: If we got a message, try to reply
	if lastMsg == nil {
		t.Log("No messages received during poll window, sending standalone message instead")

		text := fmt.Sprintf("[ilink test] standalone message at %s", time.Now().Format(time.RFC3339))
		err := client.SendTextMessage(testToUserID, "", text)
		if err != nil {
			t.Fatalf("SendTextMessage (standalone) failed: %v", err)
		}
		t.Logf("Standalone message sent: %s", text)
		return
	}

	t.Logf("Got message from %s, replying...", lastMsg.FromUserID)

	// Reply using context_token from the received message
	replyText := fmt.Sprintf("[ilink test] reply at %s", time.Now().Format(time.RFC3339))
	err = client.SendTextMessage(lastMsg.FromUserID, lastMsg.ContextToken, replyText)
	if err != nil {
		t.Fatalf("SendTextMessage (reply) failed: %v", err)
	}

	t.Logf("Reply sent: %s", replyText)
}
