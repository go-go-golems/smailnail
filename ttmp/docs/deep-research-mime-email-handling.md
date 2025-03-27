> https://chatgpt.com/share/67e555d6-221c-8012-b58f-c026540b9e40

Great. I’ll now create a detailed tutorial on composing MIME emails using `emersion/go-message`, covering multipart formatting (plain text, HTML, and attachments), encoding practices, email header structure, and how to serialize the message to bytes. I’ll then show how to append these messages into IMAP mailboxes using `go-imap`, continuing from the previous deep-dive.

I’ll let you know as soon as it’s ready for your review.

# Creating and Serializing MIME Emails with Go’s `emersion/go-message`

In this tutorial we’ll build MIME-structured email messages in Go using the `emersion/go-message` library, and then append those messages to an IMAP mailbox using `emersion/go-imap`. We’ll cover how MIME email structure works – including headers, multipart boundaries, encodings, and nested parts – and provide Go code examples for common scenarios:

- Plain text only email (single part)
- Plain text + HTML alternatives (`multipart/alternative`)
- Emails with attachments (`multipart/mixed`)
- HTML emails with inline images (`multipart/related`)

We assume a senior-level understanding of Go and internet email basics. We won’t be sending via SMTP or handling authentication (our focus is constructing the messages and storing them via IMAP). All examples use static sender/recipient addresses for clarity.

## MIME Email Structure Overview

An email message consists of a **header section** and a **body**, as defined by RFC 5322 (formerly RFC 2822) ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=,%C2%B6)). The header is a series of lines of the form `Name: value` (From, To, Subject, Date, etc.), followed by a blank line, then the body. **MIME (Multipurpose Internet Mail Extensions)** extends this format to allow structured, multi-part content and non-ASCII data in emails ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=A%20common%20misunderstanding%20about%20email,much%20like%20a%20file%20system)). Instead of a single flat text body, MIME treats an email body as a tree of parts (similar to a file system) that can contain different content types (text, HTML, images, etc.) or even nested multipart containers ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=A%20common%20misunderstanding%20about%20email,much%20like%20a%20file%20system)).

**MIME Headers:** Special headers introduced by MIME define the structure and encoding of the message. Important ones include:

- **MIME-Version:** Indicates the MIME protocol version (always `1.0` in practice) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=Mime)) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=Mime)).
- **Content-Type:** Describes the media type of the content. This can be a simple type like `text/plain` or a multipart container. If it’s multipart, a `boundary` parameter is included to delineate parts ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=Content)) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,parameter)). For example, `Content-Type: multipart/mixed; boundary="ABC123"` means the body contains multiple parts separated by the string `--ABC123`. The boundary is a unique string not appearing in the content ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,parameter)). Each MIME part within has its own headers (at least a Content-Type, and possibly others).
- **Content-Transfer-Encoding:** Indicates how the part’s data is encoded for safe transport over SMTP’s text-only channel ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=The%20encoding%20method%20used%20to,transfer%20the%20data)). Common values are:
  - `7bit` – no encoding beyond ASCII (default for plain text) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=The%20encoding%20method%20used%20to,transfer%20the%20data)).
  - `quoted-printable` – for text with special or 8-bit characters (encodes such characters as `=XY` codes, ensuring lines stay ASCII and <= 76 chars) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=The%20encoding%20method%20used%20to,transfer%20the%20data)).
  - `base64` – for binary data (encodes content in base64 text) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,through%20text%20based%20email%20systems)).
- **Content-Disposition:** Indicates presentation hint (usually `inline` or `attachment`) and optionally filename for a part ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=)). Parts meant to be displayed as part of the message (e.g. the email body or inline images) often have `Content-Disposition: inline`, whereas file attachments use `Content-Disposition: attachment; filename="file.pdf"` ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=)). This header helps mail clients decide which parts to show in the main view and which to treat as attachments ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=The%20Content,two%20values%3A%20inline%20or%20attachment)). If no Content-Disposition is provided, most clients treat the part as `inline` by default ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=attachment%2C%20then%20the%20content%20of,if%20the%20value%20were%20inline)).
- **Content-ID:** A unique identifier for a MIME part, used to reference that part from other parts. For example, an HTML part can include `<img src="cid:IMG1">` and an image part with `Content-ID: <IMG1>` to embed an inline image.

**Multipart Types:** When Content-Type is `multipart/*`, it indicates the message has multiple child parts. The common multipart subtypes are:

- **multipart/alternative:** The parts are different representations of the *same content* (e.g. plain text, HTML, and perhaps rich text). Recipients should display only one alternative – typically the last part is the richest/most preferred format ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,client%20is%20supposed%20to%20prioritise)) ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=multipart%2Falternative%20text%2Fplain%20text%2Fhtml)). For example, an email might have a text version and an HTML version of the message. The email client will choose the HTML part if it can display HTML, or fall back to text/plain if not ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=The%20reason%20for%20sending%20the,are%20capable%20of%20displaying%20HTML)) ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=multipart%2Falternative%20text%2Fplain%20text%2Fhtml)).
- **multipart/mixed:** The parts are independent pieces of content, e.g. a main message and attachments ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,additional%20secret%20content%2C%20and%20more)). This is used when you want to include attachments or just combine several parts arbitrarily. Email clients typically show all parts in a mixed message (or present attachments separately). For instance, a `multipart/mixed` email could contain a body (text or HTML) and some attachments (documents, images, etc.) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,additional%20secret%20content%2C%20and%20more)).
- **multipart/related:** The parts are meant to be treated as a single composite message, where some parts are “resources” for others ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=There%20is%20also%20multipart%2Frelated%3A%20which,are%20bundled%20together)). This is commonly used for an HTML message with inline images: the HTML part and image parts are grouped in a multipart/related so that the images are fetched from the same message rather than as separate attachments. The HTML can refer to image parts by their Content-ID. The first part in a related group is typically the “root” (e.g. the HTML), and subsequent parts are resources (images, audio, etc.). By nesting a multipart/related inside a multipart/alternative, one can have an email with an HTML+images part and a plain text part as an alternative ([MIME type to satisfy  HTML, email, images and plain text? - Stack Overflow](https://stackoverflow.com/questions/10631856/mime-type-to-satisfy-html-email-images-and-plain-text#:~:text=,)) ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=To%20make%20matters%20even%20more,multimedia%20content%20within%20the%20HTML)).

**Nested Structure:** These multipart containers can be nested. For example, a message with both an HTML body (with inline images) and a plain text alternative would have a top-level `multipart/alternative`. Inside that, the HTML part is itself a `multipart/related` containing the HTML and image parts ([MIME type to satisfy  HTML, email, images and plain text? - Stack Overflow](https://stackoverflow.com/questions/10631856/mime-type-to-satisfy-html-email-images-and-plain-text#:~:text=,)) ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=To%20make%20matters%20even%20more,multimedia%20content%20within%20the%20HTML)). Attachments would typically be added at the top as a multipart/mixed, with the first subpart being the multipart/alternative (for the body) and later parts the attachments ([MIME type to satisfy  HTML, email, images and plain text? - Stack Overflow](https://stackoverflow.com/questions/10631856/mime-type-to-satisfy-html-email-images-and-plain-text#:~:text=The%20answer%20to%20Mail%20multipart%2Falternative,message%2C%20like)). The MIME structure is effectively a tree that the email client traverses to decide what to display. If the structure isn’t organized properly (for example, putting an inline image in a multipart/mixed alongside the text parts), some clients may not display the content as intended ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=)). Generally, attachments should be separate from alternative parts, and inline media should be within a related container of the HTML part.

To summarize, MIME gives us a flexible framework to build complex emails. Now, let’s see how to construct these structures using `emersion/go-message`. This library provides streaming APIs to build messages part by part, ensuring proper formatting (it will generate boundaries, apply encodings, and format headers appropriately). After constructing the message and obtaining its raw bytes, we’ll use `emersion/go-imap` to append it to an IMAP mailbox.

## Plain Text Email (Single Part)

First, let’s create the simplest type of email: a plain text message with no attachments. This will be a single-part email with `Content-Type: text/plain`. In this case, we don’t need any multipart boundaries because there is only one body part. We will still use MIME headers to specify the content type and encoding.

Key points for a plain text email:

- **Content-Type:** `text/plain; charset=UTF-8` (we include a charset parameter to specify text encoding, UTF-8 being the usual choice).
- **Content-Transfer-Encoding:** likely `quoted-printable` if the text contains any non-ASCII characters or very long lines. The go-message library will handle this automatically for us – by default it will choose quoted-printable for text parts that aren’t strictly 7-bit ASCII ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=if%20%21h.Has%28%22Content)).
- **Content-Disposition:** optional for the main body (if unspecified or set to inline, it’s treated as body content by email clients ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=attachment%2C%20then%20the%20content%20of,if%20the%20value%20were%20inline))). The go-message API will set it to inline for us by default for the text part ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=h.Set%28%22Content)).
- We must include standard headers like From, To, Subject, Date, and MIME-Version.

Using `go-message`, we can either build the message by writing it out part-by-part, or use a convenience writer. For a single-part email, the library provides `mail.CreateSingleInlineWriter`, which writes the headers and returns an `io.WriteCloser` for the body content. This takes care of setting up proper transfer encoding for the part.

Below is an example:

```go
import (
    "bytes"
    "log"
    "time"

    "github.com/emersion/go-message/mail"
)

func buildPlainTextEmail() []byte {
    var buf bytes.Buffer

    // 1. Prepare the main email headers
    h := mail.Header{}
    h.SetDate(time.Now())  // Date: <current date>
    h.SetSubject("Test Plain Text Email")
    h.SetAddressList("From", []*mail.Address{{Name: "Alice", Address: "alice@example.org"}})
    h.SetAddressList("To", []*mail.Address{{Name: "Bob", Address: "bob@example.com"}})
    h.Set("MIME-Version", "1.0")

    // Set the Content-Type for the single part
    h.Set("Content-Type", "text/plain; charset=UTF-8")
    // (No need to set Content-Transfer-Encoding manually; CreateSingleInlineWriter will handle it)

    // 2. Create the writer for a single-part message
    wc, err := mail.CreateSingleInlineWriter(&buf, h)
    if err != nil {
        log.Fatal("CreateSingleInlineWriter:", err)
    }

    // 3. Write the plain text body
    _, err = wc.Write([]byte("Hello Bob,\r\nThis is a plain text email example.\r\nRegards,\r\nAlice"))
    if err != nil {
        log.Fatal("writing body:", err)
    }
    wc.Close()

    // The buffer now contains the raw email message
    return buf.Bytes()
}
```

Let’s break down what happens here. We created a `mail.Header` and used its methods to set standard fields. The `SetAddressList` method formats the From/To fields with names and addresses. We explicitly set `MIME-Version: 1.0` (which is required in any MIME message ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=Mime))) and the Content-Type of the body. Then `mail.CreateSingleInlineWriter(&buf, h)` writes out the headers to the buffer and prepares to write the body. We obtain an `io.WriteCloser` `wc` to write the body content. After writing the body text, we close `wc` to finalize the message.

The `CreateSingleInlineWriter` function automatically set the appropriate **Content-Transfer-Encoding** for the part based on the content type. Since we specified a `text/plain` content type, go-message will default to quoted-printable encoding for safety (if our text was pure ASCII, it would still work, but quoted-printable ensures any non-ASCII or `=` characters are safely encoded) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=if%20%21h.Has%28%22Content)). We did not manually add a Content-Transfer-Encoding header – the library inserted one for us (you can verify by examining the resulting bytes). For a plain text part with UTF-8, this will be `Content-Transfer-Encoding: quoted-printable` ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=if%20strings.HasPrefix%28t%2C%20)). If the content were entirely 7-bit safe, it might choose 7bit, but quoted-printable is the safe default for text with any special characters.

After building, `buf.Bytes()` contains the raw message. For example, it will look something like:

```
MIME-Version: 1.0
Date: Tue, 23 Mar 2025 13:36:11 -0400
Subject: Test Plain Text Email
From: Alice <alice@example.org>
To: Bob <bob@example.com>
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

Hello Bob,
This is a plain text email example.
Regards,
Alice
```

All headers are ASCII and terminated by CRLF, as required by RFC 5322 ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=,%C2%B6)). There’s a blank line separating the headers from the body ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=Where%20do%20email%20headers%20end,and%20the%20actual%20email%20begins)). The body text after encoding may contain `=20` or other QP encodings if there were trailing spaces or non-ASCII characters; in our simple text above, it might not need any encoding changes beyond normalizing line breaks to `\r\n`.

This single-part message is now ready to be sent or stored. Next, let’s build a more complex email with both text and HTML.

## Plain Text + HTML Email (Multipart/Alternative)

A common practice is to send emails with both a plain text version and an HTML version of the content. This is done with a **multipart/alternative** container: one part is `text/plain` and the other is `text/html`. Both parts represent the same message content in different formats ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,client%20is%20supposed%20to%20prioritise)). Email clients will pick the best format they can display (most will choose the HTML part if available and the user hasn’t opted to view plain text only) ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=The%20receiving%20client%20should%20only,it%20is%20capable%20of%20displaying)) ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=multipart%2Falternative%20text%2Fplain%20text%2Fhtml)). This improves compatibility with clients that can’t display HTML or for users who prefer plain text.

In a `multipart/alternative`, the **order of parts matters** – the rule is that the last part is the most preferred representation ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,client%20is%20supposed%20to%20prioritise)). So the plain text part should come first, then the HTML part. We’ll construct an email with that structure.

Using `go-message`, we have a couple of options. We could use the general `mail.Writer` (which defaults to `multipart/mixed` for attachments) and then create an alternative part within it. But an easier way here is to use `mail.CreateInlineWriter`, which specifically creates a `multipart/alternative` message. The term “Inline” in this library refers to the main body content (as opposed to attachments). In fact, `CreateInlineWriter` will set the Content-Type to `multipart/alternative` ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header%20%3D%20header,modify%20the%20caller%27s%20view)) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header.Set%28%22Content)) and give us an `InlineWriter` to add the alternative parts.

Let’s create an email with both plain text and HTML:

```go
import (
    "bytes"
    "log"
    "time"
    "io"

    "github.com/emersion/go-message/mail"
)

func buildAltEmail() []byte {
    var buf bytes.Buffer

    // 1. Set up common headers (From, To, Subject, Date, MIME-Version)
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetSubject("Test HTML Email")
    h.SetAddressList("From", []*mail.Address{{Name: "Alice", Address: "alice@example.org"}})
    h.SetAddressList("To", []*mail.Address{{Name: "Bob", Address: "bob@example.com"}})
    h.Set("MIME-Version", "1.0")
    // Note: we don't set Content-Type here; CreateInlineWriter will do that.

    // 2. Create a multipart/alternative writer
    iw, err := mail.CreateInlineWriter(&buf, h)
    if err != nil {
        log.Fatal("CreateInlineWriter:", err)
    }

    // 3. Create the plain text part
    var textHeader mail.InlineHeader
    textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
    textPart, err := iw.CreatePart(textHeader)
    if err != nil {
        log.Fatal("CreatePart text:", err)
    }
    io.WriteString(textPart, "Hello Bob,\r\nThis is the *text* version of the email.\r\nRegards,\r\nAlice")
    textPart.Close()

    // 4. Create the HTML part
    var htmlHeader mail.InlineHeader
    htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
    htmlPart, err := iw.CreatePart(htmlHeader)
    if err != nil {
        log.Fatal("CreatePart html:", err)
    }
    io.WriteString(htmlPart, `<p>Hello Bob,</p><p>This is the <b>HTML</b> version of the email.</p><p>Regards,<br>Alice</p>`)
    htmlPart.Close()

    // 5. Close the multipart/alternative writer
    iw.Close()

    return buf.Bytes()
}
```

Let’s go through this. We set up the header (From, To, etc.) similarly as before. We don’t manually specify Content-Type for the whole message; `mail.CreateInlineWriter` does that for us, internally setting `Content-Type: multipart/alternative; boundary=...` ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header%20%3D%20header,modify%20the%20caller%27s%20view)) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header.Set%28%22Content)). It returns an `InlineWriter` (`iw`) that we use to add parts.

We then create two parts using `iw.CreatePart(...)`. For each part, we use an `InlineHeader` to set the part’s Content-Type. The first part’s header is `text/plain; charset=UTF-8`, the second is `text/html; charset=UTF-8`. We write the content for each: the text part gets a plain text string, and the HTML part gets some HTML markup. Notice we included some simple HTML tags in the HTML string (paragraphs, bold, line break). After writing to each part’s WriteCloser, we close it.

Finally, we call `iw.Close()` to finalize the multipart/alternative. This writes the closing boundary marker for the multipart.

The resulting raw message (in `buf`) will have a structure like:

```
MIME-Version: 1.0
Date: ... 
Subject: Test HTML Email
From: Alice <alice@example.org>
To: Bob <bob@example.com>
Content-Type: multipart/alternative; boundary="XYZ..."

--XYZ...
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

Hello Bob,
This is the *text* version of the email.
Regards,
Alice

--XYZ...
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

<p>Hello Bob,</p><p>This is the <b>HTML</b> version of the email.</p><p>Regards,<br>Alice</p>

--XYZ-- 
```

A few things to note:

- The top-level `Content-Type` is `multipart/alternative` with a boundary string (shown as XYZ... above) that the library generated. Each part is separated by a line with `--boundary`. The final boundary appears with `--boundary--` to indicate the end ([go-message/entity_test.go at master · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/master/entity_test.go#:~:text=const%20testMultipartBody%20%3D%20%22)) ([go-message/entity_test.go at master · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/master/entity_test.go#:~:text=)).
- Each part has its own headers: our specified Content-Type plus a `Content-Transfer-Encoding: quoted-printable` (added by the library) and `Content-Disposition: inline` (added by the library via `initInlineHeader` for inline parts ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=h.Set%28%22Content))). Quoted-printable is used because we have UTF-8 text (in the HTML, for example, the `<b>` etc. are ASCII, but if there were any non-ASCII characters or just to be safe, it’s encoded). If you inspect the raw bytes, characters like `*` in the text or `<` in HTML may appear literally since they’re ASCII, but any special bytes would be QP encoded.
- The plain text part and HTML part content appear after their headers. The email client will choose one of these parts to display. Modern clients will render the HTML part (the second part) and ignore the plain text part, while simpler clients (or if a user opts for “text-only”) will show the first part. Having both versions is a best practice for broad compatibility ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=The%20reason%20for%20sending%20the,are%20capable%20of%20displaying%20HTML)).
- The boundary string is chosen automatically to be unique. We didn’t have to specify it (if needed, `mail.CreateInlineWriter` lets the library handle it).

At this point, we have a well-formed multipart/alternative message. If we needed to add attachments to this email, the typical approach is to make the top-level a `multipart/mixed`, with the first part being this multipart/alternative, and subsequent parts the attachments ([MIME type to satisfy  HTML, email, images and plain text? - Stack Overflow](https://stackoverflow.com/questions/10631856/mime-type-to-satisfy-html-email-images-and-plain-text#:~:text=The%20answer%20to%20Mail%20multipart%2Falternative,message%2C%20like)). The go-message library can handle that by using a mixed Writer and then calling `CreateInline()` to nest an alternative (as we will see next for attachments). 

## Email with Attachments (Multipart/Mixed)

Now we’ll create an email that has one or more file attachments. Attachments (like PDFs, images, etc. not intended to be displayed inline in the body) are typically included as separate parts in a `multipart/mixed` container ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,additional%20secret%20content%2C%20and%20more)). The first part of the multipart/mixed is usually the email’s main text (or an alternative block as we built above), and the subsequent parts are attachments. Each attachment part will have:

- **Content-Type:** the MIME type of the file (e.g. `application/pdf`, `image/png`, etc.).
- **Content-Transfer-Encoding:** `base64` (almost always, since attachments are arbitrary binary data) ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,through%20text%20based%20email%20systems)).
- **Content-Disposition:** `attachment; filename="name.ext"` to indicate it’s an attachment and suggest a download filename ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=)).

The go-message `mail.Writer` is designed for this scenario. We start with `mail.CreateWriter`, which by default sets up a `multipart/mixed` container for us ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header%20%3D%20header,modify%20the%20caller%27s%20view)) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header.Set%28%22Content)). We then add the body content and attachments via the writer’s methods.

Let’s build an email with a plain text body and one attachment (for illustration):

```go
import (
    "bytes"
    "log"
    "time"
    "io"

    "github.com/emersion/go-message/mail"
)

func buildEmailWithAttachment() []byte {
    var buf bytes.Buffer

    // 1. Common headers
    h := mail.Header{}
    h.SetDate(time.Now())
    h.SetSubject("Test Email with Attachments")
    h.SetAddressList("From", []*mail.Address{{Name: "Alice", Address: "alice@example.org"}})
    h.SetAddressList("To", []*mail.Address{{Name: "Bob", Address: "bob@example.com"}})
    h.Set("MIME-Version", "1.0")
    // We don't set Content-Type; CreateWriter will make it multipart/mixed.

    // 2. Create a multipart/mixed writer
    mw, err := mail.CreateWriter(&buf, h)
    if err != nil {
        log.Fatal("CreateWriter:", err)
    }

    // 3. Create the body part (plain text inline part)
    var bodyHeader mail.InlineHeader
    bodyHeader.Set("Content-Type", "text/plain; charset=UTF-8")
    bodyPart, err := mw.CreateSingleInline(bodyHeader)
    if err != nil {
        log.Fatal("CreateSingleInline:", err)
    }
    io.WriteString(bodyPart, "Hello Bob,\r\nPlease find the attached document.\r\nRegards,\r\nAlice")
    bodyPart.Close()
    // (We used CreateSingleInline for a single text part. If we wanted a text+HTML body, we could use mw.CreateInline() to start a multipart/alternative part inside the mixed message.)

    // 4. Create an attachment part
    var attHeader mail.AttachmentHeader
    attHeader.Set("Content-Type", "application/pdf")
    attHeader.SetFilename("report.pdf")  // sets Content-Disposition: attachment; filename="report.pdf"
    attPart, err := mw.CreateAttachment(attHeader)
    if err != nil {
        log.Fatal("CreateAttachment:", err)
    }
    // Write the binary content of the attachment.
    // In a real case, we'd read from a file. Here, we simulate with a placeholder:
    io.WriteString(attPart, "%PDF-1.4 binary data ...")
    attPart.Close()

    // (If multiple attachments, repeat the CreateAttachment step for each.)

    // 5. Close the multipart/mixed message writer
    mw.Close()

    return buf.Bytes()
}
```

In this code:

- We call `mail.CreateWriter(&buf, h)` after setting up the header. `CreateWriter` copies our header and adds `Content-Type: multipart/mixed; boundary=...` to it ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header%20%3D%20header,modify%20the%20caller%27s%20view)) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=header.Set%28%22Content)). So the outgoing message is now a multipart/mixed container.
- We add the email body as an inline part. We used `mw.CreateSingleInline(bodyHeader)` to create a single inline part (the plain text body). This is a convenience method equivalent to starting an inline part and writing one subpart ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=%2F%2F%20CreateSingleInline%20creates%20a%20new,part%20with%20the%20provided%20header)) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=h%20%3D%20InlineHeader,modify%20the%20caller%27s%20view)). We set the body’s content type to text/plain. The library will automatically set `Content-Transfer-Encoding: quoted-printable` for this text part and `Content-Disposition: inline` ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=if%20%21h.Has%28%22Content)) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=h.Set%28%22Content)). We write the body text and close the part.
- Next we create an attachment header. We set its Content-Type to application/pdf (pretend we are attaching a PDF file). We then call `attHeader.SetFilename("report.pdf")` which not only stores the filename but also sets the Content-Disposition to attachment ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=%2F%2F%20Create%20an%20attachment%20var,a%20JPEG%20file%20to%20w)). Now, calling `mw.CreateAttachment(attHeader)` returns a WriteCloser for the attachment body. Importantly, the library’s attachment handling will ensure that if Content-Disposition is not “attachment”, it sets it to attachment ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=disp%2C%20_%2C%20_%20%3A%3D%20h)) (so even if we hadn’t called SetFilename, it would default to attachment without a name). It will also default the transfer encoding to base64 for us ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=)) ([go-message/mail/writer.go at v0.18.2 · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/v0.18.2/mail/writer.go#L105#:~:text=if%20%21h.Has%28%22Content)).
- We write the (simulated) PDF content to the attachment writer. In practice, you might do `io.Copy(attPart, file)` after opening the file, or use `ioutil.ReadFile` and write the bytes. We then close the attachment part.

When `mw.Close()` is called, it writes the closing boundary for the mixed content and finishes the message.

The resulting email structure will be:

```
Content-Type: multipart/mixed; boundary="ABC..."; charset=UTF-8
MIME-Version: 1.0
Date: ...
Subject: Test Email with Attachments
From: Alice <alice@example.org>
To: Bob <bob@example.com>

--ABC...
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

Hello Bob,
Please find the attached document.
Regards,
Alice

--ABC...
Content-Type: application/pdf
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="report.pdf"

JVBERi0xLjQg... (base64 encoded PDF content)
...==
--ABC-- 
```

A few details:

- The top-level Content-Type is `multipart/mixed` (with an automatically generated boundary, shown as ABC... above). All parts (body and attachments) are separated by that boundary.
- The text body part appears first. As expected, it’s content type text/plain, and encoded quoted-printable (if needed). The Content-Disposition: inline marks it as the main content.
- The attachment part has `Content-Type: application/pdf` (no charset for binary types), and `Content-Transfer-Encoding: base64`. The go-message library encoded our PDF data to base64 on the fly when we wrote to `attPart`. We didn’t have to manually base64-encode the file; by setting the header and using `CreateAttachment`, the library handled the encoding. (If you peek at the buffer content before base64 encoding, you would see the raw data, but the final `buf.Bytes()` has it base64-encoded with proper 76-character line wrapping.)
- The Content-Disposition is `attachment; filename="report.pdf"`. This tells email clients to show an attachment named “report.pdf” that the user can download ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=)).
- If we had multiple attachments, each would be a new `--boundary` section with its own headers and base64 data. All attachments are peers under the multipart/mixed container.
- The `charset=UTF-8` parameter you see on the Content-Type of multipart/mixed in this output is actually not needed for multipart (multipart media types ignore any charset param), but the library might have carried it over from our mail.Header or default. It doesn’t cause harm. The important parameter for multipart is the boundary.

With this code, we can attach any kind of file. If you attach an image or text file, you would adjust the Content-Type accordingly (e.g., `image/png` or `text/csv`, etc.). The library will still use base64 by default for attachments, which is usually appropriate. (In some cases, textual attachments could use quoted-printable or 7bit if strictly ASCII, but base64 is safe for all and universally used for binary attachments ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=,through%20text%20based%20email%20systems)).)

Our email now has a body and an attachment, fully MIME-compliant. We can retrieve `buf.Bytes()` to get the raw message. This message can be sent via SMTP or saved to a .eml file, etc. In the next section, we’ll look at embedding an image in an HTML email (which is slightly different from an “attachment” meant for download).

## HTML Email with Inline Image (Multipart/Related)

Sometimes you want to include an image in the email content itself, for example a logo or picture in an HTML message, without requiring the recipient to download it as a separate attachment. This is done with **inline images** using a `multipart/related` container ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=There%20is%20also%20multipart%2Frelated%3A%20which,are%20bundled%20together)). The HTML part references the image via a CID (Content-ID), and the image is included as another part in the same related container. Email clients that understand HTML will render the image in place.

To create an email with an inline image, the MIME structure might look like:

- If no plain text alternative: Content-Type: multipart/related; boundary=...; type="text/html"
  - Part 1: Content-Type: text/html; charset=UTF-8; (maybe Content-Transfer-Encoding: quoted-printable); Content-Disposition: inline; *Body HTML with `<img src="cid:someimage">`*
  - Part 2: Content-Type: image/png (or jpeg, etc); Content-Transfer-Encoding: base64; Content-Disposition: inline; filename="image.png"; Content-ID: <someimage>
- If we also include a plain text alternative (recommended in practice): then the top-level would be multipart/alternative, and one of its parts is the above multipart/related (with HTML+image) ([MIME type to satisfy  HTML, email, images and plain text? - Stack Overflow](https://stackoverflow.com/questions/10631856/mime-type-to-satisfy-html-email-images-and-plain-text#:~:text=,)). For brevity, we will illustrate the simpler case with just HTML and image.

The go-message `mail` high-level API does not have a direct helper for multipart/related, so we will use the lower-level `message` package to construct this nested structure manually. We’ll create the HTML part and image part as separate `message.Entity` objects, then combine them with `message.NewMultipart` to form a related entity.

Here’s how we can build an HTML email with an inline image:

```go
import (
    "bytes"
    "log"
    "mime"
    "time"
    "github.com/emersion/go-message"
)

func buildHtmlWithInlineImage() []byte {
    // 1. Create the HTML part entity
    htmlHeader := message.Header{}
    htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
    htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
    htmlBody := `<html><body>
<p>Hello Bob,</p>
<p>Here is an inline image:</p>
<img src="cid:flowerIMG" alt="Flower" />
</body></html>`
    htmlEntity, err := message.New(htmlHeader, bytes.NewReader([]byte(htmlBody)))
    if err != nil {
        log.Fatal("New html entity:", err)
    }

    // 2. Create the image part entity
    imgHeader := message.Header{}
    imgHeader.Set("Content-Type", "image/png")
    imgHeader.Set("Content-Transfer-Encoding", "base64")
    imgHeader.Set("Content-Disposition", "inline; filename=\"flower.png\"")
    imgHeader.Set("Content-ID", "<flowerIMG>")
    // For demonstration, using a small PNG data (base64 encoding will be handled by the library)
    imgData := []byte{ /* ... raw PNG file bytes ... */ }
    imgEntity, err := message.New(imgHeader, bytes.NewReader(imgData))
    if err != nil {
        log.Fatal("New image entity:", err)
    }

    // 3. Combine into a multipart/related entity
    relatedHeader := message.Header{}
    relatedHeader.Set("Content-Type", "multipart/related; boundary=relBound; type=\"text/html\"")
    // 'type="text/html"' is a hint that the root part is HTML.
    relatedEntity, err := message.NewMultipart(relatedHeader, []*message.Entity{htmlEntity, imgEntity})
    if err != nil {
        log.Fatal("NewMultipart related:", err)
    }

    // 4. Set the overall message headers (From, To, Subject, etc) on the related entity
    relatedEntity.Header.Set("MIME-Version", "1.0")
    relatedEntity.Header.Set("Date", time.Now().Format(time.RFC1123Z))
    relatedEntity.Header.Set("Subject", "Test Inline Image Email")
    relatedEntity.Header.Set("From", "Alice <alice@example.org>")
    relatedEntity.Header.Set("To", "Bob <bob@example.com>")

    // 5. Write the entity to a buffer to get raw message
    var buf bytes.Buffer
    if err := relatedEntity.WriteTo(&buf); err != nil {
        log.Fatal("WriteTo:", err)
    }

    return buf.Bytes()
}
```

Let’s explain the steps:

1. **HTML part**: We set up a `message.Header` for the HTML part with content type text/html and charset UTF-8. We also set `Content-Transfer-Encoding: quoted-printable` because the HTML might contain non-ASCII or just to be safe (here it’s actually all ASCII, but QP is fine). We then create the entity with `message.New(htmlHeader, bytes.NewReader([]byte(htmlBody)))`. The htmlBody string contains an `<img>` tag with `src="cid:flowerIMG"`. This “cid” is referencing a Content-ID that we’ll use in the image part. We choose an identifier `flowerIMG` (it could be any unique identifier, often formatted like an email id). We wrap it in angle brackets in the Content-ID header later.
   
2. **Image part**: We prepare the headers for the image. It’s `image/png` in this example. We set `Content-Transfer-Encoding: base64` (we definitely want base64 for binary image data) and `Content-Disposition: inline` with a filename. We also set `Content-ID: <flowerIMG>`. Notice the Content-ID value is `<flowerIMG>` which matches what the HTML’s cid is referencing (including the angle brackets). Now we create the image entity with `message.New(imgHeader, bytes.NewReader(imgData))`. Here `imgData` would be the raw bytes of the PNG file. For example, if we had the image file on disk, we could read it into a byte slice. In our case, we’d plug in the actual bytes or obtain them. The library will see the header says base64 and will handle encoding those bytes to base64 when writing out ([go-message/encoding.go at master · emersion/go-message · GitHub](https://github.com/emersion/go-message/blob/master/encoding.go#:~:text=case%20)). (If the image were large, using an `io.Reader` stream is better than loading all bytes, but we’ll keep it simple.)

3. **Combine into multipart/related**: We create a header for the related container. We set `Content-Type: multipart/related; boundary=relBound; type="text/html"`. We explicitly set a boundary here (`relBound` for illustration) – in real usage, you could omit the boundary and let the library generate one, but it’s often helpful to set it for consistency in examples. We also set the `type="text/html"` parameter to indicate the root part’s MIME type (this is optional but a good hint for clients). Then we call `message.NewMultipart(relatedHeader, []*message.Entity{ htmlEntity, imgEntity })` to create a multipart entity containing the HTML and image parts. Now `relatedEntity` represents the whole MIME body of our email.

4. **Overall headers**: Now we need to add the standard email headers (From, To, Subject, Date, MIME-Version) to the top-level. Since our top-level entity is `relatedEntity`, we add those to its Header. We use `relatedEntity.Header.Set(...)` for each. (We formatted the Date using time.RFC1123Z which is an acceptable full date format; `mail.Header.SetDate` could also be used via a mail.Header, but here we did it manually.)

5. **Serialize**: Finally, we write the `relatedEntity` to a bytes.Buffer using `relatedEntity.WriteTo(&buf)`. The WriteTo method takes care of outputting the headers and encoding the body parts with the boundaries.

The resulting MIME structure will be:

```
MIME-Version: 1.0
Date: Mon, 23 Mar 2025 13:36:11 -0400
Subject: Test Inline Image Email
From: Alice <alice@example.org>
To: Bob <bob@example.com>
Content-Type: multipart/related; boundary="relBound"; type="text/html"

--relBound
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline

<html><body>
<p>Hello Bob,</p>
<p>Here is an inline image:</p>
<img src="cid:flowerIMG" alt="Flower" />
</body></html>

--relBound
Content-Type: image/png
Content-Transfer-Encoding: base64
Content-Disposition: inline; filename="flower.png"
Content-ID: <flowerIMG>

iVBORw0KGgoAAAANS... (base64 encoded image data)
... (more base64 lines) ...
--relBound--
```

As you can see, the HTML part and image part are bundled together inside the `multipart/related`. The HTML references the image via the Content-ID (`cid:flowerIMG`), and the image part’s `Content-ID` header matches that. The email client will typically render the HTML, see the `<img src="cid:flowerIMG">` reference, find the part with Content-ID `<flowerIMG>`, decode it (base64 to binary), and display it in place. Because we used `Content-Disposition: inline` for the image, it hints to the client that this is not a separate attachment to list, but part of the message body ([Working with messages](https://mimekit.net/docs/html/Working-With-Messages.htm#:~:text=The%20Content,two%20values%3A%20inline%20or%20attachment)). The `multipart/related` container itself indicates that the parts are meant to be presented together as a single unit ([What an email really looks like — MailPace](https://mailpace.com/blog/guides/what-an-email-looks-like#:~:text=There%20is%20also%20multipart%2Frelated%3A%20which,are%20bundled%20together)).

In practice, we would likely nest this `multipart/related` inside a `multipart/alternative` with a plain text part for maximum compatibility (so that non-HTML clients have a fallback) ([MIME type to satisfy  HTML, email, images and plain text? - Stack Overflow](https://stackoverflow.com/questions/10631856/mime-type-to-satisfy-html-email-images-and-plain-text#:~:text=,)). That would make the top-level Content-Type `multipart/alternative` with two parts: one text/plain, and one multipart/related (which in turn contains text/html + image). Constructing that programmatically would involve nesting one more level (creating an alternative entity with the text part entity and the related entity). The go-message library supports creating nested structures similarly (we could use `message.NewMultipart` for the alternative, etc.). The code gets more complex, but the principles are the same. The example above covers the core of inline image embedding.

We’ve now created several kinds of emails and obtained their raw `[]byte` representations. All these messages are in proper RFC 5322 format with CRLF line breaks, ready to be sent or stored. Next, we’ll show how to append these raw emails to an IMAP mailbox.

## Appending the Raw Message to an IMAP Mailbox

With our message bytes at hand (for example, from `buildEmailWithAttachment()` or any of the builders above), we can use the `emersion/go-imap` client to append the message to a mailbox on an IMAP server. The IMAP `APPEND` command allows adding a message to a mailbox (like saving a sent message to “Sent” or uploading an email to an account).

First, ensure you have a connected and authenticated IMAP client (`*client.Client`). For example, you might dial and log in like:

```go
c, err := client.DialTLS("imap.example.com:993", nil)
if err != nil {
    log.Fatal(err)
}
defer c.Logout()
if err := c.Login("username", "password"); err != nil {
    log.Fatal(err)
}
```

*(We assume the connection is already established and authenticated; error handling omitted for brevity.)*

To append a message, use the `Client.Append` method:

```go
msgBytes := buildEmailWithAttachment()  // our raw RFC 5322 message bytes
flags := []string{imap.SeenFlag}        // mark as \Seen, for example
mailbox := "Sent"                       // target mailbox (e.g. "INBOX" or "Sent")

// Append the message to the mailbox
err = c.Append(mailbox, flags, time.Now(), bytes.NewReader(msgBytes))
if err != nil {
    log.Fatal("Append:", err)
}
```

The `Append` method takes the mailbox name, a slice of flags, a date (internal timestamp for the message), and an `imap.Literal` containing the full message data ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=,%C2%B6)) ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=%2F%2F%20Append%20it%20to%20INBOX%2C,log.Fatal%28err)). Here we used `bytes.NewReader(msgBytes)` to provide the message. You could also do `c.Append("INBOX", nil, time.Now(), &buf)` if you still have your bytes.Buffer (since `bytes.Buffer` implements `io.Reader`). The `flags` can include `\Seen`, `\Answered`, etc., or be nil/empty. The date is optional; we passed `time.Now()` for “current time” (this becomes the internal date stored on the message in IMAP).

It’s crucial that the byte stream we append is a valid RFC 5322 message with all the headers and CRLFs ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=,%C2%B6)). Our use of go-message ensured that. The IMAP server will take those bytes and store them exactly as given. The snippet above would append the message to the “Sent” mailbox and mark it as read. Adjust the mailbox name or flags as needed. 

After appending, you can `c.Logout()` or continue with other IMAP commands. There’s no SMTP involved in this flow – we’re not “sending” the email to anyone, just storing it. For sending, one would use an SMTP client (or an API) with the same raw bytes.

To verify, you could fetch the message from IMAP or open the mailbox in an email client – you should see the emails we constructed, with proper attachments or inline content. The structure we built (headers, MIME boundaries, encodings) will be preserved exactly. 

---

**Conclusion:** We covered how to use `emersion/go-message` to programmatically build MIME emails of various types, taking care of MIME headers like Content-Type, Content-Disposition, and Content-Transfer-Encoding. We also saw how to get the raw message bytes and use `emersion/go-imap` to append those messages to an IMAP mailbox (e.g., to save a sent message). By handling the MIME construction ourselves, we have fine-grained control over the email content and structure, which is useful in applications like email clients, automated mailers, or mail processing tools. Each example we built can be extended or combined (for instance, adding attachments to the HTML email with inline images scenario would involve nesting the structures accordingly). The Go libraries handle much of the low-level detail (like boundary generation and encoding) for us, while allowing us to set the necessary headers to comply with email standards. With this knowledge, you can construct complex emails and ensure they are correctly formatted for transit or storage. 

