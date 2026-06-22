package mail

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
)

// Message provides a handle to the email message being constructed.
// It exposes methods for setting addresses, content, attachments,
// and headers on the underlying message during send callbacks.
type Message struct {
	from        Address
	sender      Address
	returnPath  string
	to          []Address
	cc          []Address
	bcc         []Address
	replyTo     []Address
	subject     string
	htmlBody    string
	textBody    string
	priority    int
	headers     map[string][]string
	attachments []*Attachment
	embeds      []*Embed
}

// Embed represents an inline-embedded resource with a Content-ID.
type Embed struct {
	CID  string
	Data []byte
	Name string
	Mime string
}

// AttachOption configures an attachment.
type AttachOption func(*Attachment)

// NewMessage creates a new empty Message.
func NewMessage() *Message {
	return &Message{
		headers: make(map[string][]string),
	}
}

// From sets the sender address.
func (m *Message) From(address string, name ...string) *Message {
	m.from = Address{Email: address}

	if len(name) > 0 {
		m.from.Name = name[0]
	}

	return m
}

// GetFrom returns the sender address.
func (m *Message) GetFrom() Address {
	return m.from
}

// Sender sets the actual sender (Sender header), distinct from From.
func (m *Message) Sender(address string, name ...string) *Message {
	m.sender = Address{Email: address}

	if len(name) > 0 {
		m.sender.Name = name[0]
	}

	return m
}

// GetSender returns the Sender header address.
func (m *Message) GetSender() Address {
	return m.sender
}

// ReturnPath sets the Return-Path address for bounce handling.
func (m *Message) ReturnPath(address string) *Message {
	m.returnPath = address

	return m
}

// GetReturnPath returns the Return-Path address.
func (m *Message) GetReturnPath() string {
	return m.returnPath
}

// To adds primary recipients. If override is true, existing To addresses are replaced.
func (m *Message) To(addresses []Address, override ...bool) *Message {
	if len(override) > 0 && override[0] {
		m.to = nil
	}

	m.to = append(m.to, addresses...)

	return m
}

// ForgetTo removes all To recipients.
func (m *Message) ForgetTo() *Message {
	m.to = nil

	return m
}

// GetTo returns the To recipients.
func (m *Message) GetTo() []Address {
	return m.to
}

// CC adds carbon-copy recipients. If override is true, existing CC addresses are replaced.
func (m *Message) CC(addresses []Address, override ...bool) *Message {
	if len(override) > 0 && override[0] {
		m.cc = nil
	}

	m.cc = append(m.cc, addresses...)

	return m
}

// ForgetCC removes all CC recipients.
func (m *Message) ForgetCC() *Message {
	m.cc = nil

	return m
}

// GetCC returns the CC recipients.
func (m *Message) GetCC() []Address {
	return m.cc
}

// BCC adds blind-carbon-copy recipients. If override is true, existing BCC addresses are replaced.
func (m *Message) BCC(addresses []Address, override ...bool) *Message {
	if len(override) > 0 && override[0] {
		m.bcc = nil
	}

	m.bcc = append(m.bcc, addresses...)

	return m
}

// ForgetBCC removes all BCC recipients.
func (m *Message) ForgetBCC() *Message {
	m.bcc = nil

	return m
}

// GetBCC returns the BCC recipients.
func (m *Message) GetBCC() []Address {
	return m.bcc
}

// ReplyTo adds reply-to addresses.
func (m *Message) ReplyTo(addresses []Address) *Message {
	m.replyTo = append(m.replyTo, addresses...)

	return m
}

// GetReplyTo returns the Reply-To addresses.
func (m *Message) GetReplyTo() []Address {
	return m.replyTo
}

// Subject sets the email subject line.
func (m *Message) Subject(subject string) *Message {
	m.subject = subject

	return m
}

// GetSubject returns the subject line.
func (m *Message) GetSubject() string {
	return m.subject
}

// Priority sets the email priority (1 = highest, 5 = lowest).
func (m *Message) Priority(level int) *Message {
	m.priority = level

	return m
}

// GetPriority returns the priority level.
func (m *Message) GetPriority() int {
	return m.priority
}

// SetHTMLBody sets the HTML body content.
func (m *Message) SetHTMLBody(html string) *Message {
	m.htmlBody = html

	return m
}

// GetHTMLBody returns the HTML body.
func (m *Message) GetHTMLBody() string {
	return m.htmlBody
}

// SetTextBody sets the plain-text body content.
func (m *Message) SetTextBody(text string) *Message {
	m.textBody = text

	return m
}

// GetTextBody returns the plain-text body.
func (m *Message) GetTextBody() string {
	return m.textBody
}

// Attach adds a file attachment by path.
func (m *Message) Attach(path string, options ...AttachOption) *Message {
	a := &Attachment{Path: path}

	for _, opt := range options {
		opt(a)
	}

	m.attachments = append(m.attachments, a)

	return m
}

// AttachData adds an attachment from raw data.
func (m *Message) AttachData(data []byte, name string, options ...AttachOption) *Message {
	a := FromData(func() (io.Reader, error) {
		return bytes.NewReader(data), nil
	}, name)

	for _, opt := range options {
		opt(a)
	}

	m.attachments = append(m.attachments, a)

	return m
}

// AttachWith adds a pre-built Attachment to the message.
func (m *Message) AttachWith(a *Attachment) *Message {
	m.attachments = append(m.attachments, a)

	return m
}

// GetAttachments returns all file attachments.
func (m *Message) GetAttachments() []*Attachment {
	return m.attachments
}

// Embed embeds a file inline and returns the Content-ID reference.
func (m *Message) Embed(path string, options ...AttachOption) string {
	cid := generateCID()
	a := &Attachment{Path: path, Inline: true}

	for _, opt := range options {
		opt(a)
	}

	m.embeds = append(m.embeds, &Embed{
		CID:  cid,
		Name: a.Name,
		Mime: a.Mime,
	})

	return "cid:" + cid
}

// EmbedData embeds raw data inline and returns the Content-ID reference.
func (m *Message) EmbedData(data []byte, name string, mime ...string) string {
	cid := generateCID()
	mimeType := "application/octet-stream"

	if len(mime) > 0 {
		mimeType = mime[0]
	}

	m.embeds = append(m.embeds, &Embed{
		CID:  cid,
		Data: data,
		Name: name,
		Mime: mimeType,
	})

	return "cid:" + cid
}

// GetEmbeds returns all inline embeds.
func (m *Message) GetEmbeds() []*Embed {
	return m.embeds
}

// SetHeader sets a custom header value.
func (m *Message) SetHeader(name string, values ...string) *Message {
	m.headers[name] = values

	return m
}

// GetCustomHeaders returns all custom headers.
func (m *Message) GetCustomHeaders() map[string][]string {
	return m.headers
}

// AllRecipients returns a deduplicated list of all recipients (To + CC + BCC).
func (m *Message) AllRecipients() []Address {
	seen := make(map[string]struct{})

	var all []Address

	for _, lists := range [][]Address{m.to, m.cc, m.bcc} {
		for _, a := range lists {
			if _, ok := seen[a.Email]; !ok {
				seen[a.Email] = struct{}{}
				all = append(all, a)
			}
		}
	}

	return all
}

// WithName sets the attachment display filename.
func WithName(name string) AttachOption {
	return func(a *Attachment) {
		a.Name = name
	}
}

// WithMimeType sets the attachment MIME content type.
func WithMimeType(mime string) AttachOption {
	return func(a *Attachment) {
		a.Mime = mime
	}
}

func generateCID() string {
	b := make([]byte, 16)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(fmt.Errorf("mail: generate content id: %w", err))
	}

	return fmt.Sprintf("%x", b)
}
