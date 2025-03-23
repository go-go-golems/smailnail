# Storing Messages in IMAP Mailboxes with Go

This tutorial provides a detailed walkthrough of how to store email messages in IMAP mailboxes using the Go programming language with the excellent `github.com/emersion/go-imap/v2` and `github.com/emersion/go-message` libraries.

## Introduction

The Internet Message Access Protocol (IMAP) is a standard email protocol used for accessing and manipulating email messages stored on a mail server. While most developers are familiar with retrieving messages from IMAP servers, storing messages is an equally important operation that requires understanding of both IMAP concepts and message formatting.

This tutorial focuses specifically on the message storage aspect of IMAP operations, showing how to construct RFC 5322-compliant messages and append them to IMAP mailboxes using Go libraries that handle the complex details of these standards.

## Prerequisites

To follow this tutorial, you should have:

- Go 1.18 or higher installed
- Basic understanding of email concepts and IMAP
- The following Go libraries:
  - `github.com/emersion/go-imap/v2` - Core IMAP functionality
  - `github.com/emersion/go-message` - Email message creation and manipulation

## Library Overview

### go-imap/v2

The `go-imap/v2` library provides a comprehensive implementation of the IMAP protocol. For message storage, we'll primarily use the following packages:

- `github.com/emersion/go-imap/v2` - Core IMAP types and data structures
- `github.com/emersion/go-imap/v2/imapclient` - IMAP client for connecting to servers

Key types for message storage include:
- `imapclient.Client` - The IMAP client connection
- `imap.AppendOptions` - Options for the APPEND command
- `imap.Flag` - Message flags (like \Seen, \Draft, etc.)

### go-message

The `go-message` library handles email message creation following RFC standards. For message construction, we'll use:

- `github.com/emersion/go-message` - Core message creation functionality
- `github.com/emersion/go-message/mail` - Mail-specific message formatting

Key types include:
- `mail.Header` - Email header manipulation
- `mail.Writer` - Creates multipart messages with attachments
- `mail.InlineWriter` - For alternative text/HTML content

## Connecting to an IMAP Server

Before storing messages, you need to establish a connection with the IMAP server and authenticate:

```go
package main

import (
    "log"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
)

func connectToIMAP(server, username, password string) (*imapclient.Client, error) {
    // Connect to the server
    client, err := imapclient.DialTLS(server, nil)
    if err != nil {
        return nil, err
    }
    
    // Wait for server greeting
    if err := client.WaitGreeting(); err != nil {
        client.Close()
        return nil, err
    }
    
    // Login
    if err := client.Login(username, password).Wait(); err != nil {
        client.Close()
        return nil, err
    }
    
    return client, nil
}
```

## Creating Simple Text Email Messages

For storing messages in IMAP mailboxes, you first need to construct the message. Let's start with simple text messages:

```go
package main

import (
    "bytes"
    "time"
    
    "github.com/emersion/go-message/mail"
)

func createTextMessage(from, to, subject, body string) ([]byte, error) {
    var buf bytes.Buffer
    
    // Create a new mail message
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetAddressList("From", []*mail.Address{{Address: from}})
    h.SetAddressList("To", []*mail.Address{{Address: to}})
    h.SetSubject(subject)
    
    // Create a message writer
    w, err := mail.CreateSingleInlineWriter(&buf, h)
    if err != nil {
        return nil, err
    }
    
    // Write the plain text body
    if _, err := w.Write([]byte(body)); err != nil {
        return nil, err
    }
    
    // Close the writer to finalize the message
    if err := w.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

## Creating Rich HTML Messages with Alternatives

For more complex messages with HTML content and plain text alternatives:

```go
func createHTMLMessage(from, to, subject, textBody, htmlBody string) ([]byte, error) {
    var buf bytes.Buffer
    
    // Create a new mail message
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetAddressList("From", []*mail.Address{{Address: from}})
    h.SetAddressList("To", []*mail.Address{{Address: to}})
    h.SetSubject(subject)
    
    // Create a multipart message with alternatives
    mw, err := mail.CreateWriter(&buf, h)
    if err != nil {
        return nil, err
    }
    
    // Create the alternative part
    altw, err := mw.CreateInline()
    if err != nil {
        return nil, err
    }
    
    // Add the plain text part
    th := mail.InlineHeader{}
    th.Set("Content-Type", "text/plain; charset=utf-8")
    tw, err := altw.CreatePart(th)
    if err != nil {
        return nil, err
    }
    if _, err := tw.Write([]byte(textBody)); err != nil {
        return nil, err
    }
    if err := tw.Close(); err != nil {
        return nil, err
    }
    
    // Add the HTML part
    hh := mail.InlineHeader{}
    hh.Set("Content-Type", "text/html; charset=utf-8")
    hw, err := altw.CreatePart(hh)
    if err != nil {
        return nil, err
    }
    if _, err := hw.Write([]byte(htmlBody)); err != nil {
        return nil, err
    }
    if err := hw.Close(); err != nil {
        return nil, err
    }
    
    // Close the alternative and message writers
    if err := altw.Close(); err != nil {
        return nil, err
    }
    if err := mw.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

## Adding Attachments to Messages

To create messages with file attachments:

```go
func createMessageWithAttachment(from, to, subject, body string, 
    filename string, fileContent []byte, contentType string) ([]byte, error) {
    
    var buf bytes.Buffer
    
    // Create the mail header
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetAddressList("From", []*mail.Address{{Address: from}})
    h.SetAddressList("To", []*mail.Address{{Address: to}})
    h.SetSubject(subject)
    
    // Create the multipart message
    mw, err := mail.CreateWriter(&buf, h)
    if err != nil {
        return nil, err
    }
    
    // Create the text part
    th := mail.InlineHeader{}
    th.Set("Content-Type", "text/plain; charset=utf-8")
    tw, err := mw.CreateSingleInline(th)
    if err != nil {
        return nil, err
    }
    if _, err := tw.Write([]byte(body)); err != nil {
        return nil, err
    }
    if err := tw.Close(); err != nil {
        return nil, err
    }
    
    // Create the attachment part
    ah := mail.AttachmentHeader{}
    ah.Set("Content-Type", contentType)
    ah.SetFilename(filename)
    aw, err := mw.CreateAttachment(ah)
    if err != nil {
        return nil, err
    }
    if _, err := aw.Write(fileContent); err != nil {
        return nil, err
    }
    if err := aw.Close(); err != nil {
        return nil, err
    }
    
    // Close the message writer
    if err := mw.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

## Storing Messages in IMAP Mailboxes

Now that we can create messages, let's see how to store them in IMAP mailboxes:

```go
func storeMessage(client *imapclient.Client, mailbox string, 
    messageData []byte, flags []imap.Flag, date time.Time) error {
    
    // Set the append options (flags and internal date)
    options := &imap.AppendOptions{
        Flags: flags,
        Time:  date,
    }
    
    // Create an append command
    cmd := client.Append(mailbox, int64(len(messageData)), options)
    
    // Write the message data
    if _, err := cmd.Write(messageData); err != nil {
        return err
    }
    
    // Close the writer to finalize the append
    if err := cmd.Close(); err != nil {
        return err
    }
    
    // Wait for the command to complete
    _, err := cmd.Wait()
    return err
}
```

## Working with Message Flags

IMAP messages have flags that indicate their status. Common flags include:

- `\Seen` - Message has been read
- `\Answered` - Message has been replied to
- `\Flagged` - Message is "flagged" for urgent/special attention
- `\Deleted` - Message is marked for deletion
- `\Draft` - Message is a draft
- `\Recent` - Message is "recent" (this is obsolete in IMAP4rev2)

Here's how to set flags when appending a message:

```go
func storeDraftMessage(client *imapclient.Client, mailbox string, messageData []byte) error {
    // Set draft flag
    options := &imap.AppendOptions{
        Flags: []imap.Flag{imap.FlagDraft},
        Time:  time.Now(),
    }
    
    cmd := client.Append(mailbox, int64(len(messageData)), options)
    if _, err := cmd.Write(messageData); err != nil {
        return err
    }
    if err := cmd.Close(); err != nil {
        return err
    }
    
    _, err := cmd.Wait()
    return err
}
```

## Complete Example: Creating and Storing a Message

Let's put everything together in a complete example:

```go
package main

import (
    "bytes"
    "fmt"
    "log"
    "time"
    
    "github.com/emersion/go-imap/v2"
    "github.com/emersion/go-imap/v2/imapclient"
    "github.com/emersion/go-message/mail"
)

func main() {
    // Connect to the IMAP server
    client, err := connectToIMAP("imap.example.com:993", "username", "password")
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()
    
    // Create a simple message
    messageData, err := createTextMessage(
        "sender@example.com",
        "recipient@example.com",
        "Hello from Go IMAP",
        "This is a test message created using go-imap and go-message libraries.",
    )
    if err != nil {
        log.Fatalf("Failed to create message: %v", err)
    }
    
    // Store the message in the Drafts mailbox with Draft flag
    err = storeMessage(
        client,
        "Drafts",
        messageData,
        []imap.Flag{imap.FlagDraft},
        time.Now(),
    )
    if err != nil {
        log.Fatalf("Failed to store message: %v", err)
    }
    
    fmt.Println("Message successfully stored in Drafts mailbox")
}
```

## Advanced Techniques

### Creating Messages with Inline Images

Inline images are embedded in HTML content and referenced with the `cid:` scheme:

```go
func createMessageWithInlineImage(from, to, subject, htmlBody string,
    imageName string, imageContent []byte, imageContentType string) ([]byte, error) {
    
    var buf bytes.Buffer
    
    // Create the mail header
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetAddressList("From", []*mail.Address{{Address: from}})
    h.SetAddressList("To", []*mail.Address{{Address: to}})
    h.SetSubject(subject)
    
    // Create the multipart message
    mw, err := mail.CreateWriter(&buf, h)
    if err != nil {
        return nil, err
    }
    
    // Create the alternative part
    altw, err := mw.CreateInline()
    if err != nil {
        return nil, err
    }
    
    // Add the plain text part
    th := mail.InlineHeader{}
    th.Set("Content-Type", "text/plain; charset=utf-8")
    tw, err := altw.CreatePart(th)
    if err != nil {
        return nil, err
    }
    if _, err := tw.Write([]byte("This message contains an inline image.")); err != nil {
        return nil, err
    }
    if err := tw.Close(); err != nil {
        return nil, err
    }
    
    // Add the HTML part (with image reference)
    hh := mail.InlineHeader{}
    hh.Set("Content-Type", "text/html; charset=utf-8")
    hw, err := altw.CreatePart(hh)
    if err != nil {
        return nil, err
    }
    if _, err := hw.Write([]byte(htmlBody)); err != nil {
        return nil, err
    }
    if err := hw.Close(); err != nil {
        return nil, err
    }
    
    // Close the alternative part
    if err := altw.Close(); err != nil {
        return nil, err
    }
    
    // Add the inline image as a related part
    ih := mail.AttachmentHeader{}
    ih.Set("Content-Type", imageContentType)
    ih.Set("Content-ID", "<"+imageName+">")
    ih.Set("Content-Disposition", "inline")
    iw, err := mw.CreateAttachment(ih)
    if err != nil {
        return nil, err
    }
    if _, err := iw.Write(imageContent); err != nil {
        return nil, err
    }
    if err := iw.Close(); err != nil {
        return nil, err
    }
    
    // Close the message writer
    if err := mw.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

In the HTML body, refer to the image using:
```html
<img src="cid:imageName" alt="Description" />
```

### Working with Message-ID and References Headers

For proper threading and reply handling:

```go
func setMessageIDAndReferences(header *mail.Header, inReplyTo, references []string) {
    // Generate a new Message-ID if not already set
    if header.Get("Message-ID") == "" {
        if err := header.GenerateMessageID(); err != nil {
            // Handle error or use a custom method to generate
            msgID := fmt.Sprintf("<%d.%d@example.com>", time.Now().Unix(), rand.Int63())
            header.SetMessageID(msgID[1:len(msgID)-1])
        }
    }
    
    // Set In-Reply-To header if provided
    if len(inReplyTo) > 0 {
        header.Set("In-Reply-To", inReplyTo[0])
    }
    
    // Set References header if provided
    if len(references) > 0 {
        header.SetMsgIDList("References", references)
    }
}
```

### Creating Calendar Invites

For calendar invites, use the appropriate MIME type and calendar attachment:

```go
func createCalendarInvite(from, to, subject string, icsData []byte) ([]byte, error) {
    var buf bytes.Buffer
    
    // Create the mail header
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetAddressList("From", []*mail.Address{{Address: from}})
    h.SetAddressList("To", []*mail.Address{{Address: to}})
    h.SetSubject(subject)
    
    // Set method for calendar processing
    h.Set("Content-Type", "text/calendar; method=REQUEST; charset=UTF-8")
    
    // Create the message writer
    w, err := mail.CreateSingleInlineWriter(&buf, h)
    if err != nil {
        return nil, err
    }
    
    // Write the iCalendar data
    if _, err := w.Write(icsData); err != nil {
        return nil, err
    }
    
    // Close the writer
    if err := w.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

## Batch Processing Messages

For efficient handling of multiple messages:

```go
func batchStoreMessages(client *imapclient.Client, mailbox string, messages [][]byte) error {
    for i, msgData := range messages {
        cmd := client.Append(mailbox, int64(len(msgData)), nil)
        if _, err := cmd.Write(msgData); err != nil {
            return fmt.Errorf("failed to write message %d: %w", i, err)
        }
        
        if err := cmd.Close(); err != nil {
            return fmt.Errorf("failed to close message %d: %w", i, err)
        }
        
        if _, err := cmd.Wait(); err != nil {
            return fmt.Errorf("failed to append message %d: %w", i, err)
        }
    }
    
    return nil
}
```

## Handling Special Characters and Encodings

The `go-message` library handles character encoding automatically, but for custom cases:

```go
func createInternationalMessage(from, to, subject, body string) ([]byte, error) {
    var buf bytes.Buffer
    
    // Create a new mail header
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetAddressList("From", []*mail.Address{{Address: from}})
    h.SetAddressList("To", []*mail.Address{{Address: to}})
    
    // Subject will be automatically encoded if it contains non-ASCII characters
    h.SetSubject(subject)
    
    // Create a message writer
    w, err := mail.CreateSingleInlineWriter(&buf, h)
    if err != nil {
        return nil, err
    }
    
    // Write the body (UTF-8 is handled automatically)
    if _, err := w.Write([]byte(body)); err != nil {
        return nil, err
    }
    
    // Close the writer
    if err := w.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

## Mailbox Creation and Management

Before storing messages, you might need to create or select mailboxes:

```go
func ensureMailboxExists(client *imapclient.Client, mailbox string) error {
    // Try to select the mailbox first to check if it exists
    cmd := client.Select(mailbox, &imap.SelectOptions{ReadOnly: true})
    err := cmd.Wait()
    
    if err == nil {
        // Mailbox exists, unselect it
        return client.Unselect().Wait()
    }
    
    // Attempt to create the mailbox
    err = client.Create(mailbox, nil).Wait()
    if err != nil {
        return fmt.Errorf("failed to create mailbox %s: %w", mailbox, err)
    }
    
    return nil
}
```

## Error Handling and Best Practices

When working with IMAP servers, robust error handling is crucial:

```go
func safelyStoreMessage(client *imapclient.Client, mailbox string, 
    messageData []byte, maxRetries int) error {
    
    var lastErr error
    for attempt := 0; attempt < maxRetries; attempt++ {
        cmd := client.Append(mailbox, int64(len(messageData)), nil)
        
        if _, err := cmd.Write(messageData); err != nil {
            lastErr = err
            // If it's a connection issue, try to reconnect
            if attempt < maxRetries-1 {
                time.Sleep(time.Second * time.Duration(attempt+1))
                continue
            }
            return err
        }
        
        if err := cmd.Close(); err != nil {
            lastErr = err
            if attempt < maxRetries-1 {
                time.Sleep(time.Second * time.Duration(attempt+1))
                continue
            }
            return err
        }
        
        _, err := cmd.Wait()
        if err == nil {
            return nil // Success
        }
        
        lastErr = err
        if attempt < maxRetries-1 {
            time.Sleep(time.Second * time.Duration(attempt+1))
        }
    }
    
    return fmt.Errorf("failed to store message after %d attempts: %w", maxRetries, lastErr)
}
```

## Performance Considerations

When storing large batches of messages, consider:

1. **Memory management**: Process and append messages in batches to avoid excessive memory usage
2. **Connection pooling**: Maintain a pool of connections for parallel processing
3. **Throttling**: Some IMAP servers have rate limits, so consider adding delays between operations

```go
func processLargeMessageBatch(client *imapclient.Client, mailbox string, 
    messageBatch [][]byte, batchSize int) error {
    
    totalMessages := len(messageBatch)
    for i := 0; i < totalMessages; i += batchSize {
        end := i + batchSize
        if end > totalMessages {
            end = totalMessages
        }
        
        currentBatch := messageBatch[i:end]
        if err := batchStoreMessages(client, mailbox, currentBatch); err != nil {
            return err
        }
        
        // Add a small delay between batches to avoid overwhelming the server
        if end < totalMessages {
            time.Sleep(100 * time.Millisecond)
        }
    }
    
    return nil
}
```

## Conclusion

Storing messages in IMAP mailboxes using Go involves constructing RFC-compliant messages with the `go-message` library and using the IMAP APPEND command via the `go-imap/v2` library. This tutorial covered:

1. Connecting to IMAP servers
2. Creating various types of email messages (text, HTML, attachments)
3. Setting message flags
4. Appending messages to mailboxes
5. Handling special cases, errors, and performance considerations

The `go-imap` and `go-message` libraries provide a robust, standards-compliant foundation for working with email messages and IMAP servers in Go applications.

## Further Resources

- [go-imap v2 Documentation](https://pkg.go.dev/github.com/emersion/go-imap/v2)
- [go-message Documentation](https://pkg.go.dev/github.com/emersion/go-message)
- [RFC 3501: IMAP4rev1](https://www.rfc-editor.org/rfc/rfc3501)
- [RFC 9051: IMAP4rev2](https://www.rfc-editor.org/rfc/rfc9051.html)
- [RFC 5322: Internet Message Format](https://www.rfc-editor.org/rfc/rfc5322) 