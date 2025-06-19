package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Raw   string `json:"raw"`
}

type EmailAddressList []EmailAddress

func (e EmailAddressList) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *EmailAddressList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %v to []byte", value)
	}
	return json.Unmarshal(bytes, e)
}

type HeadersMap map[string]string

func (h HeadersMap) Value() (driver.Value, error) {
	return json.Marshal(h)
}

func (h *HeadersMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %v to []byte", value)
	}
	return json.Unmarshal(bytes, h)
}

type Message struct {
	Base
	InboxID string           `json:"inbox_id" db:"inbox_id" validate:"required"`
	Subject string           `json:"subject" db:"subject"`
	HTML    string           `json:"html" db:"html"`
	Text    string           `json:"text" db:"text"`
	From    EmailAddressList `json:"from" db:"from"`
	To      EmailAddressList `json:"to" db:"to"`
	Cc      EmailAddressList `json:"cc" db:"cc"`
	Bcc     EmailAddressList `json:"bcc" db:"bcc"`
	IsRead  bool             `json:"is_read" db:"is_read"`

	// New fields
	MessageID      string     `json:"message_id" db:"message_id"` // RFC Message-ID
	RawSHA1        string     `json:"raw_sha1" db:"raw_sha1"`     // SHA1 of raw content
	RawPath        string     `json:"raw_path" db:"raw_path"`     // Path on disk
	References     []string   `json:"references" db:"references"` // References from the email headers, TEXT[] in PostgreSQL
	InReplyTo      string     `json:"in_reply_to" db:"in_reply_to"`
	Headers        HeadersMap `json:"headers" db:"headers"`
	SpamScore      float64    `json:"spam_score" db:"spam_score"`
	SpamThreshold  float64    `json:"spam_threshold" db:"spam_threshold"`
	SpamAction     string     `json:"spam_action" db:"spam_action"`
	HasAttachments bool       `json:"has_attachments" db:"has_attachments"`
}

type Attachment struct {
	ID          string `json:"id" db:"id"`
	MessageID   string `json:"message_id" db:"message_id"`
	Name        string `json:"name" db:"name"`
	MimeType    string `json:"mimetype" db:"mimetype"`
	ContentType string `json:"content_type" db:"content_type"`
	Path        string `json:"path" db:"path"`
}
