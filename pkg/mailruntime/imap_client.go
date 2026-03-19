package mailruntime

import (
	"context"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// IMAPOptions holds connection parameters for an IMAP server.
type IMAPOptions struct {
	Host     string
	Port     int
	TLS      bool
	Username string
	Password string
}

// MailboxInfo is a lightweight descriptor returned by LIST.
type MailboxInfo struct {
	Name      string   `json:"name"`
	Flags     []string `json:"flags"`
	Delimiter string   `json:"delimiter"`
}

// MailboxStatus is returned by STATUS.
type MailboxStatus struct {
	Messages    uint32 `json:"messages"`
	Unseen      uint32 `json:"unseen"`
	UIDNext     uint32 `json:"uidNext"`
	UIDValidity uint32 `json:"uidValidity"`
	Recent      uint32 `json:"recent"`
}

// MessageEnvelope mirrors IMAP ENVELOPE data.
type MessageEnvelope struct {
	Date      string   `json:"date"`
	Subject   string   `json:"subject"`
	From      []string `json:"from"`
	To        []string `json:"to"`
	CC        []string `json:"cc"`
	BCC       []string `json:"bcc"`
	ReplyTo   []string `json:"replyTo"`
	MessageID string   `json:"messageId"`
	InReplyTo string   `json:"inReplyTo"`
}

// AttachmentInfo describes a message attachment.
type AttachmentInfo struct {
	Part        string `json:"part"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        uint32 `json:"size"`
	CID         string `json:"cid"`
}

// FetchedMessage holds fetched data for a message.
type FetchedMessage struct {
	UID          uint32            `json:"uid"`
	Flags        []string          `json:"flags"`
	Size         int64             `json:"size"`
	InternalDate string            `json:"internalDate"`
	Envelope     *MessageEnvelope  `json:"envelope,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	BodyText     string            `json:"bodyText,omitempty"`
	BodyHTML     string            `json:"bodyHTML,omitempty"`
	BodyRaw      []byte            `json:"bodyRaw,omitempty"`
	Attachments  []AttachmentInfo  `json:"attachments,omitempty"`

	client      *IMAPClient
	mailboxName string
}

// IMAPClient wraps go-imap/v2 with higher-level operations.
type IMAPClient struct {
	c            *imapclient.Client
	capabilities map[string]bool
	selectedBox  string
}

// Connect opens an IMAP connection and logs in.
func Connect(_ context.Context, opts IMAPOptions) (*IMAPClient, error) {
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	log.Debug().Str("addr", addr).Msg("connecting to IMAP")

	var (
		c   *imapclient.Client
		err error
	)

	if opts.TLS {
		c, err = imapclient.DialTLS(addr, nil)
	} else {
		c, err = imapclient.DialInsecure(addr, nil)
	}
	if err != nil {
		return nil, errors.Wrap(err, "dial IMAP")
	}

	if err := c.Login(opts.Username, opts.Password).Wait(); err != nil {
		_ = c.Logout().Wait()
		return nil, &MailError{Name: "AuthError", Message: err.Error(), Source: "imap"}
	}

	caps, err := c.Capability().Wait()
	if err != nil {
		_ = c.Logout().Wait()
		return nil, errors.Wrap(err, "CAPABILITY")
	}

	capMap := make(map[string]bool)
	for cap_ := range caps {
		capMap[strings.ToLower(string(cap_))] = true
	}

	log.Debug().Interface("caps", capMap).Msg("IMAP capabilities")

	return &IMAPClient{c: c, capabilities: capMap}, nil
}

func (ic *IMAPClient) Capabilities() map[string]bool {
	return ic.capabilities
}

func (ic *IMAPClient) Logout() error {
	return ic.c.Logout().Wait()
}

func (ic *IMAPClient) List(pattern string) ([]MailboxInfo, error) {
	if pattern == "" {
		pattern = "*"
	}
	cmd := ic.c.List("", pattern, nil)
	data, err := cmd.Collect()
	if err != nil {
		return nil, errors.Wrap(err, "LIST")
	}
	out := make([]MailboxInfo, 0, len(data))
	for _, mb := range data {
		flags := make([]string, 0, len(mb.Attrs))
		for _, f := range mb.Attrs {
			flags = append(flags, string(f))
		}
		delim := ""
		if mb.Delim != 0 {
			delim = string(mb.Delim)
		}
		out = append(out, MailboxInfo{
			Name:      mb.Mailbox,
			Flags:     flags,
			Delimiter: delim,
		})
	}
	return out, nil
}

func (ic *IMAPClient) Status(name string) (*MailboxStatus, error) {
	data, err := ic.c.Status(name, &imap.StatusOptions{
		NumMessages: true,
		NumUnseen:   true,
		UIDNext:     true,
		UIDValidity: true,
	}).Wait()
	if err != nil {
		return nil, errors.Wrap(err, "STATUS")
	}
	st := &MailboxStatus{}
	if data.NumMessages != nil {
		st.Messages = *data.NumMessages
	}
	if data.NumUnseen != nil {
		st.Unseen = *data.NumUnseen
	}
	if data.UIDNext != 0 {
		st.UIDNext = uint32(data.UIDNext)
	}
	if data.UIDValidity != 0 {
		st.UIDValidity = data.UIDValidity
	}
	return st, nil
}

func (ic *IMAPClient) SelectMailbox(name string, readOnly bool) (*imap.SelectData, error) {
	opts := &imap.SelectOptions{ReadOnly: readOnly}
	data, err := ic.c.Select(name, opts).Wait()
	if err != nil {
		return nil, &MailError{Name: "NoSuchMailboxError", Message: err.Error(), Source: "imap"}
	}
	ic.selectedBox = name
	return data, nil
}

func (ic *IMAPClient) UnselectMailbox() error {
	if ic.selectedBox == "" {
		return nil
	}
	if ic.capabilities["unselect"] {
		if err := ic.c.Unselect().Wait(); err != nil {
			return errors.Wrap(err, "UNSELECT")
		}
	} else if err := ic.c.UnselectAndExpunge().Wait(); err != nil {
		log.Debug().Err(err).Msg("UnselectAndExpunge fallback error")
	}
	ic.selectedBox = ""
	return nil
}

// SearchCriteria is a higher-level representation of IMAP search criteria.
type SearchCriteria struct {
	All      bool
	Seen     bool
	Unseen   bool
	Flagged  bool
	Answered bool
	Deleted  bool
	Draft    bool
	From     string
	To       string
	CC       string
	BCC      string
	Subject  string
	Text     string
	Body     string
	Since    *time.Time
	Before   *time.Time
	On       *time.Time
	Larger   int64
	Smaller  int64
	Header   map[string]string
	UID      *imap.UIDSet
	Not      *SearchCriteria
	Or       []*SearchCriteria
	And      []*SearchCriteria
	Raw      string
}

func (ic *IMAPClient) Search(criteria *SearchCriteria) ([]imap.UID, error) {
	imapCriteria := buildIMAPCriteria(criteria)
	data, err := ic.c.UIDSearch(imapCriteria, nil).Wait()
	if err != nil {
		return nil, errors.Wrap(err, "UID SEARCH")
	}
	return data.AllUIDs(), nil
}

type FetchField string

const (
	FetchUID           FetchField = "uid"
	FetchFlags         FetchField = "flags"
	FetchInternalDate  FetchField = "internalDate"
	FetchSize          FetchField = "size"
	FetchEnvelope      FetchField = "envelope"
	FetchHeaders       FetchField = "headers"
	FetchBodyText      FetchField = "body.text"
	FetchBodyHTML      FetchField = "body.html"
	FetchBodyRaw       FetchField = "body.raw"
	FetchAttachments   FetchField = "attachments"
	FetchBodyStructure FetchField = "bodyStructure"
)

func (ic *IMAPClient) Fetch(uids []imap.UID, fields []FetchField) ([]*FetchedMessage, error) {
	if len(uids) == 0 {
		return nil, nil
	}

	uidSet := imap.UIDSetNum(uids...)
	fetchOpts := buildFetchOptions(fields)

	cmd := ic.c.Fetch(uidSet, fetchOpts)
	var msgs []*FetchedMessage
	for {
		msgData := cmd.Next()
		if msgData == nil {
			break
		}
		buf, err := msgData.Collect()
		if err != nil {
			return nil, errors.Wrap(err, "collecting message")
		}
		fm, err := collectMessage(buf, fields, ic)
		if err != nil {
			return nil, errors.Wrap(err, "processing message")
		}
		fm.mailboxName = ic.selectedBox
		msgs = append(msgs, fm)
	}
	if err := cmd.Close(); err != nil {
		return nil, errors.Wrap(err, "FETCH")
	}
	return msgs, nil
}

func (ic *IMAPClient) StoreFlags(uids []imap.UID, op imap.StoreFlagsOp, flags []imap.Flag, silent bool) error {
	uidSet := imap.UIDSetNum(uids...)
	opts := &imap.StoreFlags{
		Op:     op,
		Flags:  flags,
		Silent: silent,
	}
	return ic.c.Store(uidSet, opts, nil).Close()
}

func (ic *IMAPClient) MoveUIDs(uids []imap.UID, dest string) error {
	uidSet := imap.UIDSetNum(uids...)
	_, err := ic.c.Move(uidSet, dest).Wait()
	return errors.Wrap(err, "MOVE")
}

func (ic *IMAPClient) CopyUIDs(uids []imap.UID, dest string) error {
	uidSet := imap.UIDSetNum(uids...)
	_, err := ic.c.Copy(uidSet, dest).Wait()
	return errors.Wrap(err, "COPY")
}

func (ic *IMAPClient) DeleteUIDs(uids []imap.UID, expunge bool) error {
	if err := ic.StoreFlags(uids, imap.StoreFlagsAdd, []imap.Flag{imap.FlagDeleted}, true); err != nil {
		return errors.Wrap(err, "store \\Deleted")
	}
	if expunge {
		return ic.Expunge(uids)
	}
	return nil
}

func (ic *IMAPClient) Expunge(uids []imap.UID) error {
	if len(uids) > 0 && ic.capabilities["uidplus"] {
		uidSet := imap.UIDSetNum(uids...)
		return ic.c.UIDExpunge(uidSet).Close()
	}
	return ic.c.Expunge().Close()
}

func (ic *IMAPClient) Append(mailbox string, msg []byte, flags []imap.Flag, date *time.Time) (imap.UID, error) {
	opts := &imap.AppendOptions{Flags: flags}
	if date != nil {
		opts.Time = *date
	}
	cmd := ic.c.Append(mailbox, int64(len(msg)), opts)
	if _, err := cmd.Write(msg); err != nil {
		return 0, errors.Wrap(err, "append write")
	}
	if err := cmd.Close(); err != nil {
		return 0, errors.Wrap(err, "APPEND close")
	}
	appendData, err := cmd.Wait()
	if err != nil {
		return 0, errors.Wrap(err, "APPEND wait")
	}
	return appendData.UID, nil
}

func (ic *IMAPClient) CreateMailbox(name string) error {
	return ic.c.Create(name, nil).Wait()
}

func (ic *IMAPClient) RenameMailbox(old, newName string) error {
	return ic.c.Rename(old, newName).Wait()
}

func (ic *IMAPClient) DeleteMailbox(name string) error {
	return ic.c.Delete(name).Wait()
}

func (ic *IMAPClient) Subscribe(name string) error {
	return ic.c.Subscribe(name).Wait()
}

func (ic *IMAPClient) Unsubscribe(name string) error {
	return ic.c.Unsubscribe(name).Wait()
}

func (ic *IMAPClient) SelectedMailbox() string {
	return ic.selectedBox
}

func (ic *IMAPClient) FetchBodyPart(uid imap.UID, part []int) ([]byte, error) {
	uidSet := imap.UIDSetNum(uid)
	section := &imap.FetchItemBodySection{
		Part: part,
		Peek: true,
	}
	opts := &imap.FetchOptions{
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{section},
	}
	cmd := ic.c.Fetch(uidSet, opts)
	defer func() {
		if closeErr := cmd.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("closing IMAP fetch body part command")
		}
	}()

	msgData := cmd.Next()
	if msgData == nil {
		return nil, fmt.Errorf("message UID %d not found", uid)
	}
	buf, err := msgData.Collect()
	if err != nil {
		return nil, errors.Wrap(err, "collect body part")
	}
	data := buf.FindBodySection(section)
	if data == nil {
		return nil, fmt.Errorf("body section not found in response")
	}
	return data, nil
}

func (ic *IMAPClient) FetchRaw(uid imap.UID) ([]byte, error) {
	uidSet := imap.UIDSetNum(uid)
	section := &imap.FetchItemBodySection{Peek: true}
	opts := &imap.FetchOptions{
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{section},
	}
	cmd := ic.c.Fetch(uidSet, opts)
	defer func() {
		if closeErr := cmd.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("closing IMAP fetch raw command")
		}
	}()

	msgData := cmd.Next()
	if msgData == nil {
		return nil, fmt.Errorf("message UID %d not found", uid)
	}
	buf, err := msgData.Collect()
	if err != nil {
		return nil, errors.Wrap(err, "collect raw message")
	}
	data := buf.FindBodySection(section)
	if data == nil {
		return nil, fmt.Errorf("body section not found in response")
	}
	return data, nil
}

func (ic *IMAPClient) FetchBodyPartReader(uid imap.UID, part []int) (io.ReadCloser, error) {
	data, err := ic.FetchBodyPart(uid, part)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(strings.NewReader(string(data))), nil
}

func buildIMAPCriteria(c *SearchCriteria) *imap.SearchCriteria {
	if c == nil {
		return &imap.SearchCriteria{}
	}
	sc := &imap.SearchCriteria{}

	if c.Seen {
		sc.Flag = append(sc.Flag, imap.FlagSeen)
	}
	if c.Unseen {
		sc.NotFlag = append(sc.NotFlag, imap.FlagSeen)
	}
	if c.Flagged {
		sc.Flag = append(sc.Flag, imap.FlagFlagged)
	}
	if c.Answered {
		sc.Flag = append(sc.Flag, imap.FlagAnswered)
	}
	if c.Deleted {
		sc.Flag = append(sc.Flag, imap.FlagDeleted)
	}
	if c.Draft {
		sc.Flag = append(sc.Flag, imap.FlagDraft)
	}

	if c.From != "" {
		sc.Header = append(sc.Header, imap.SearchCriteriaHeaderField{Key: "From", Value: c.From})
	}
	if c.To != "" {
		sc.Header = append(sc.Header, imap.SearchCriteriaHeaderField{Key: "To", Value: c.To})
	}
	if c.CC != "" {
		sc.Header = append(sc.Header, imap.SearchCriteriaHeaderField{Key: "Cc", Value: c.CC})
	}
	if c.BCC != "" {
		sc.Header = append(sc.Header, imap.SearchCriteriaHeaderField{Key: "Bcc", Value: c.BCC})
	}
	if c.Subject != "" {
		sc.Header = append(sc.Header, imap.SearchCriteriaHeaderField{Key: "Subject", Value: c.Subject})
	}
	if c.Text != "" {
		sc.Text = append(sc.Text, c.Text)
	}
	if c.Body != "" {
		sc.Body = append(sc.Body, c.Body)
	}
	for k, v := range c.Header {
		sc.Header = append(sc.Header, imap.SearchCriteriaHeaderField{Key: k, Value: v})
	}
	if c.Since != nil {
		sc.Since = *c.Since
	}
	if c.Before != nil {
		sc.Before = *c.Before
	}
	if c.On != nil {
		t := *c.On
		start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		end := start.Add(24 * time.Hour)
		sc.Since = start
		sc.Before = end
	}
	if c.Larger > 0 {
		sc.Larger = c.Larger
	}
	if c.Smaller > 0 {
		sc.Smaller = c.Smaller
	}
	if c.UID != nil {
		sc.UID = []imap.UIDSet{*c.UID}
	}
	if c.Not != nil {
		sc.Not = append(sc.Not, *buildIMAPCriteria(c.Not))
	}
	if len(c.Or) >= 2 {
		sc.Or = append(sc.Or, [2]imap.SearchCriteria{
			*buildIMAPCriteria(c.Or[0]),
			*buildIMAPCriteria(c.Or[1]),
		})
	}
	for _, a := range c.And {
		sub := buildIMAPCriteria(a)
		sc.And(sub)
	}
	return sc
}

func buildFetchOptions(fields []FetchField) *imap.FetchOptions {
	opts := &imap.FetchOptions{UID: true}
	for _, f := range fields {
		switch f {
		case FetchUID:
			opts.UID = true
		case FetchFlags:
			opts.Flags = true
		case FetchInternalDate:
			opts.InternalDate = true
		case FetchSize:
			opts.RFC822Size = true
		case FetchEnvelope:
			opts.Envelope = true
		case FetchHeaders:
			opts.BodySection = append(opts.BodySection, &imap.FetchItemBodySection{
				Specifier: imap.PartSpecifierHeader,
				Peek:      true,
			})
		case FetchBodyText, FetchBodyHTML:
			opts.BodySection = append(opts.BodySection, &imap.FetchItemBodySection{
				Specifier: imap.PartSpecifierText,
				Peek:      true,
			})
		case FetchBodyRaw:
			opts.BodySection = append(opts.BodySection, &imap.FetchItemBodySection{Peek: true})
		case FetchBodyStructure, FetchAttachments:
			opts.BodyStructure = &imap.FetchItemBodyStructure{Extended: true}
		}
	}
	return opts
}

func collectMessage(msg *imapclient.FetchMessageBuffer, _ []FetchField, ic *IMAPClient) (*FetchedMessage, error) {
	fm := &FetchedMessage{
		UID:    uint32(msg.UID),
		client: ic,
	}
	if msg.Flags != nil {
		for _, f := range msg.Flags {
			fm.Flags = append(fm.Flags, string(f))
		}
	}
	if msg.RFC822Size > 0 {
		fm.Size = msg.RFC822Size
	}
	if !msg.InternalDate.IsZero() {
		fm.InternalDate = msg.InternalDate.Format(time.RFC3339)
	}
	if msg.Envelope != nil {
		fm.Envelope = convertEnvelope(msg.Envelope)
	}
	for _, section := range msg.BodySection {
		data := section.Bytes
		if section.Section == nil {
			fm.BodyRaw = data
			continue
		}
		switch section.Section.Specifier {
		case imap.PartSpecifierHeader:
			fm.Headers = parseHeadersMap(string(data))
		case imap.PartSpecifierText:
			fm.BodyText = string(data)
		case imap.PartSpecifierMIME, imap.PartSpecifierNone:
			fm.BodyRaw = data
		default:
			fm.BodyRaw = data
		}
	}
	if msg.BodyStructure != nil {
		fm.Attachments = extractAttachments(msg.BodyStructure, "")
	}
	return fm, nil
}

func convertEnvelope(e *imap.Envelope) *MessageEnvelope {
	me := &MessageEnvelope{
		Subject:   e.Subject,
		MessageID: e.MessageID,
	}
	if !e.Date.IsZero() {
		me.Date = e.Date.Format(time.RFC3339)
	}
	for _, a := range e.From {
		me.From = append(me.From, addressString(a))
	}
	for _, a := range e.To {
		me.To = append(me.To, addressString(a))
	}
	for _, a := range e.Cc {
		me.CC = append(me.CC, addressString(a))
	}
	for _, a := range e.Bcc {
		me.BCC = append(me.BCC, addressString(a))
	}
	for _, a := range e.ReplyTo {
		me.ReplyTo = append(me.ReplyTo, addressString(a))
	}
	return me
}

func addressString(a imap.Address) string {
	if a.Name != "" {
		return fmt.Sprintf("%s <%s@%s>", a.Name, a.Mailbox, a.Host)
	}
	return fmt.Sprintf("%s@%s", a.Mailbox, a.Host)
}

func parseHeadersMap(raw string) map[string]string {
	headers := make(map[string]string)
	for _, line := range strings.Split(raw, "\r\n") {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key != "" {
			headers[key] = val
		}
	}
	return headers
}

func extractAttachments(bs imap.BodyStructure, partID string) []AttachmentInfo {
	var out []AttachmentInfo
	switch v := bs.(type) {
	case *imap.BodyStructureMultiPart:
		for i, child := range v.Children {
			childID := fmt.Sprintf("%d", i+1)
			if partID != "" {
				childID = partID + "." + childID
			}
			out = append(out, extractAttachments(child, childID)...)
		}
	case *imap.BodyStructureSinglePart:
		disp := v.Disposition()
		if disp != nil && strings.EqualFold(disp.Value, "attachment") {
			filename := ""
			if fn, ok := disp.Params["filename"]; ok {
				filename = fn
			}
			cid := ""
			if v.ID != "" {
				cid = strings.Trim(v.ID, "<>")
			}
			out = append(out, AttachmentInfo{
				Part:        partID,
				Filename:    filename,
				ContentType: v.MediaType(),
				Size:        v.Size,
				CID:         cid,
			})
		}
	}
	return out
}

// ParseCriteria converts a goja value into SearchCriteria for hosts that expose criteria objects to JS.
func ParseCriteria(vm *goja.Runtime, val goja.Value) *SearchCriteria {
	if goja.IsUndefined(val) || goja.IsNull(val) {
		return &SearchCriteria{All: true}
	}
	if val.ExportType().Kind().String() == "string" {
		return &SearchCriteria{Raw: val.String()}
	}

	obj := val.ToObject(vm)
	c := &SearchCriteria{}

	getBool := func(key string) (bool, bool) {
		v := obj.Get(key)
		if v == nil || goja.IsUndefined(v) {
			return false, false
		}
		return v.ToBoolean(), true
	}

	if v, ok := getBool("all"); ok && v {
		c.All = true
	}
	if v, ok := getBool("seen"); ok && v {
		c.Seen = true
	}
	if v, ok := getBool("unseen"); ok && v {
		c.Unseen = true
	}
	if v, ok := getBool("flagged"); ok && v {
		c.Flagged = true
	}
	if v, ok := getBool("answered"); ok && v {
		c.Answered = true
	}
	if v, ok := getBool("deleted"); ok && v {
		c.Deleted = true
	}
	if v, ok := getBool("draft"); ok && v {
		c.Draft = true
	}

	if v := obj.Get("from"); v != nil && !goja.IsUndefined(v) {
		c.From = StringOrRegex(v)
	}
	if v := obj.Get("to"); v != nil && !goja.IsUndefined(v) {
		c.To = StringOrRegex(v)
	}
	if v := obj.Get("cc"); v != nil && !goja.IsUndefined(v) {
		c.CC = StringOrRegex(v)
	}
	if v := obj.Get("bcc"); v != nil && !goja.IsUndefined(v) {
		c.BCC = StringOrRegex(v)
	}
	if v := obj.Get("subject"); v != nil && !goja.IsUndefined(v) {
		c.Subject = StringOrRegex(v)
	}
	if v := obj.Get("text"); v != nil && !goja.IsUndefined(v) {
		c.Text = v.String()
	}
	if v := obj.Get("body"); v != nil && !goja.IsUndefined(v) {
		c.Body = v.String()
	}
	if v := obj.Get("since"); v != nil && !goja.IsUndefined(v) {
		t := ParseJSDate(v)
		c.Since = &t
	}
	if v := obj.Get("before"); v != nil && !goja.IsUndefined(v) {
		t := ParseJSDate(v)
		c.Before = &t
	}
	if v := obj.Get("on"); v != nil && !goja.IsUndefined(v) {
		t := ParseJSDate(v)
		c.On = &t
	}
	if v := obj.Get("larger"); v != nil && !goja.IsUndefined(v) {
		c.Larger = v.ToInteger()
	}
	if v := obj.Get("smaller"); v != nil && !goja.IsUndefined(v) {
		c.Smaller = v.ToInteger()
	}
	if v := obj.Get("header"); v != nil && !goja.IsUndefined(v) {
		hObj := v.ToObject(vm)
		c.Header = make(map[string]string)
		for _, key := range hObj.Keys() {
			c.Header[key] = hObj.Get(key).String()
		}
	}
	if v := obj.Get("uid"); v != nil && !goja.IsUndefined(v) {
		uidSet := ParseUIDSet(v)
		c.UID = &uidSet
	}
	if v := obj.Get("not"); v != nil && !goja.IsUndefined(v) {
		c.Not = ParseCriteria(vm, v)
	}
	if v := obj.Get("or"); v != nil && !goja.IsUndefined(v) {
		arr := v.ToObject(vm)
		for _, k := range arr.Keys() {
			c.Or = append(c.Or, ParseCriteria(vm, arr.Get(k)))
		}
	}
	if v := obj.Get("and"); v != nil && !goja.IsUndefined(v) {
		arr := v.ToObject(vm)
		for _, k := range arr.Keys() {
			c.And = append(c.And, ParseCriteria(vm, arr.Get(k)))
		}
	}
	return c
}

// StringOrRegex extracts a string from a JS string or regex.
func StringOrRegex(v goja.Value) string {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return ""
	}
	if obj, ok := v.(*goja.Object); ok {
		if src := obj.Get("source"); src != nil && !goja.IsUndefined(src) {
			return regexToSubstring(src.String())
		}
	}
	return v.String()
}

func regexToSubstring(pattern string) string {
	pattern = strings.TrimPrefix(pattern, "^")
	pattern = strings.TrimSuffix(pattern, "$")
	if idx := strings.Index(pattern, "|"); idx >= 0 {
		pattern = pattern[:idx]
	}
	re := regexp.MustCompile(`[\\^$.*+?()[\]{}|]`)
	return re.ReplaceAllString(pattern, "")
}

// ParseJSDate parses a JS Date object, YYYY-MM-DD string, relative day string, or RFC3339 string.
func ParseJSDate(v goja.Value) time.Time {
	if obj, ok := v.(*goja.Object); ok {
		if getTime := obj.Get("getTime"); getTime != nil {
			if fn, ok := goja.AssertFunction(getTime); ok {
				ret, err := fn(v)
				if err == nil {
					ms := ret.ToInteger()
					return time.Unix(ms/1000, (ms%1000)*1e6).UTC()
				}
			}
		}
	}
	s := v.String()
	if strings.HasSuffix(s, "d") {
		if days, err := strconv.Atoi(strings.TrimSuffix(s, "d")); err == nil {
			return time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
		}
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

// ParseUIDSet parses a JS value to imap.UIDSet.
func ParseUIDSet(v goja.Value) imap.UIDSet {
	if arr, ok := v.(*goja.Object); ok && arr.Get("length") != nil {
		var uids []imap.UID
		length := int(arr.Get("length").ToInteger())
		for i := 0; i < length; i++ {
			el := arr.Get(strconv.Itoa(i))
			if el != nil {
				if uid, ok := jsValueToUID(el); ok {
					uids = append(uids, uid)
				}
			}
		}
		return imap.UIDSetNum(uids...)
	}
	s := v.String()
	var set imap.UIDSet
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, ":") {
			bounds := strings.SplitN(part, ":", 2)
			var lo, hi imap.UID
			if bounds[0] != "*" {
				n, _ := strconv.ParseUint(bounds[0], 10, 32)
				lo = imap.UID(n)
			}
			if bounds[1] != "*" {
				n, _ := strconv.ParseUint(bounds[1], 10, 32)
				hi = imap.UID(n)
			}
			set.AddRange(lo, hi)
			continue
		}
		if n, err := strconv.ParseUint(part, 10, 32); err == nil {
			set.AddNum(imap.UID(n))
		}
	}
	return set
}

func jsValueToUID(v goja.Value) (imap.UID, bool) {
	raw := v.ToInteger()
	if raw <= 0 || raw > math.MaxUint32 {
		return 0, false
	}
	return imap.UID(uint32(raw)), true
}
