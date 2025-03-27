> https://chatgpt.com/share/67e555d6-221c-8012-b58f-c026540b9e40

Great. I’ll create a very detailed, in-depth, and technical tutorial that covers all IMAP verbs and functionalities using the Go libraries [`emersion/go-imap`](https://github.com/emersion/go-imap) and [`emersion/go-message`](https://github.com/emersion/go-message). This will include short, focused code examples illustrating how to use each verb or feature, suitable for a senior developer.

I'll let you know as soon as it's ready for your review.

# Comprehensive Guide to IMAP in Go with `go-imap` and `go-message`

This tutorial provides a deep dive into using the Go libraries **`emersion/go-imap`** (an IMAP client library) and **`emersion/go-message`** (a MIME message parsing library) to implement full IMAP client functionality. We’ll cover all major IMAP commands (LOGIN, SELECT, FETCH, SEARCH, STORE, COPY, MOVE, APPEND, EXPUNGE, IDLE, NOOP, LIST, LSUB, STATUS, CREATE, DELETE, RENAME, SUBSCRIBE, UNSUBSCRIBE, and their UID variants) with code examples in Go. We’ll also explore parsing email messages (MIME format) using `go-message` for decoding headers, attachments, and nested multipart content. Along the way, we’ll explain IMAP flags (like `\Seen`, `\Answered`), the difference between sequence numbers and UIDs, mailbox state transitions in IMAP, and provide tips on efficient message retrieval, concurrency, and folder synchronization. This guide is written for experienced developers who want a thorough understanding of IMAP client programming in Go.

**Prerequisites:** We assume you have already established a connection to an IMAP server and have an `imap.Client` instance (from `go-imap`) ready to use. We won't cover server connection setup or authentication methods beyond the basic `LOGIN` command. All code examples focus on the IMAP client logic using `go-imap`, and parsing emails with `go-message`. 

Let's get started.

## IMAP Client States and `go-imap` Overview

Before diving into commands, it's important to understand IMAP session states and how `go-imap` models them. IMAP defines distinct **connection states**:

- **Not Authenticated:** Initial state after connecting. The client must log in (authenticate) before most commands are allowed ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=%2F%2F%20In%20the%20not%20authenticated,authenticated)).
- **Authenticated:** Logged in, but no mailbox selected. In this state, you can manage mailboxes (LIST, CREATE, DELETE, etc.) or set up idle, but to retrieve or modify messages you must select a mailbox ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=%2F%2F%20In%20the%20authenticated%20state%2C,before%20commands%20that%20affect%20messages)).
- **Selected:** A mailbox is selected (e.g. via SELECT or EXAMINE) and message-related commands (FETCH, SEARCH, STORE, COPY, etc.) can be used on that mailbox ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=%2F%2F%20In%20a%20selected%20state%2C,a%20mailbox%20has%20been%20successfully)). Only one mailbox can be selected per connection at a time.
- **Logout:** The session is ending (after LOGOUT command).

Some commands cause state transitions. For example, a successful LOGIN moves you from Not Authenticated to Authenticated. SELECT moves from Authenticated to Selected. A CLOSE (or the non-standard UNSELECT extension) will return the session to Authenticated state (deselecting the mailbox). The `go-imap` client tracks its state; you can query `c.State()` to check if the client is in NotAuthenticated, Authenticated, or Selected state.

**Concurrency note:** IMAP (and `go-imap`) generally expects one command at a time on a given connection. The `go-imap` client is **not safe for concurrent use from multiple goroutines** ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=It%20is%20not%20safe%20to,commands%20on%20the%20same%20connection)). This means you should serialize command calls on a single `Client` or use separate connections if you need to perform operations in parallel. Some commands (like FETCH or LIST) stream responses asynchronously via channels, which we'll see in examples. Always consume these response channels promptly to avoid blocking the IMAP connection. The `Client.Updates` channel can be used to receive asynchronous server updates (expunges, new mail notifications, etc.), but if you use it, make sure to read from it to prevent blocking the client ([client package - gopkg.in/emersion/go-imap.v1/client - Go Packages](https://pkg.go.dev/gopkg.in/emersion/go-imap.v1/client#:~:text=type%20Client%20struct%20)).

With these concepts in mind, let's walk through each IMAP command, demonstrating usage with `go-imap`, and then cover message parsing with `go-message`.

## LOGIN – Authenticating with the IMAP Server

The `LOGIN` command authenticates with a username and password. Using `go-imap`, you call the `Client.Login` method. This moves the client to the Authenticated state if successful.

```go
err := c.Login("username", "password")
if err != nil {
    // handle authentication error
}
```

This corresponds to the IMAP command `A001 LOGIN username password`. After `Login`, the server will allow authenticated-state commands. (For OAuth or other SASL authentication, `Client.Authenticate` would be used instead, but we focus on simple LOGIN here.)

Once logged in, you might want to check server capabilities (using `c.Capability()`), but for brevity we will proceed to mailbox operations.

## LIST – Listing All Mailboxes

`LIST` retrieves mailboxes on the server. In IMAP, mailboxes are folders (e.g., "INBOX", "Sent", "Archive/2021", etc.). With `go-imap`, `Client.List(reference, pattern, chan)` is used. This command is asynchronous: the library will send `LIST` to the server and stream the results (each mailbox) into a channel of `*imap.MailboxInfo`. You supply a channel and typically run the `List` call in a goroutine.

Example – list all mailboxes in the account:

```go
mailboxes := make(chan *imap.MailboxInfo, 50)
go func() {
    err = c.List("", "*", mailboxes)
}()
fmt.Println("Mailboxes:")
for m := range mailboxes {
    fmt.Println(" -", m.Name)
}
if err != nil {
    // handle error from List
}
``` 

In the above snippet, the `reference` is `""` (the root reference) and the `pattern` is `"*"` to match all mailboxes. The server’s responses are read from the `mailboxes` channel. We used a buffered channel (size 50) to avoid blocking the server’s sending of data. After the loop, when the channel closes, we check the error to ensure the command completed successfully ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=%2F%2F%20List%20mailboxes%20mailboxes%20%3A%3D,mailboxes%29)) ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=if%20err%20%3A%3D%20%3C,log.Fatal%28err%29)).

The `MailboxInfo` struct contains fields like `Name` (the mailbox name), `Delimiter` (hierarchy delimiter, often "/" or "."), and `Attributes` (flags like `\HasNoChildren`, `\Noselect`, etc. indicating mailbox properties).

## LSUB – Listing Subscribed Mailboxes

`LSUB` is similar to LIST, but it only lists mailboxes marked as “subscribed” (a user preference on which folders to show). Use `Client.Lsub` in `go-imap`:

```go
subs := make(chan *imap.MailboxInfo, 20)
go func() {
    err = c.Lsub("", "*", subs)
}()
for m := range subs {
    fmt.Println("Subscribed:", m.Name)
}
if err != nil {
    // handle error
}
```

This will list all subscribed mailboxes matching the pattern. If a mailbox is subscribed, it means the user has marked it for inclusion in listings (common in some email clients to show only a subset of all folders).

## CREATE – Creating a Mailbox

To create a new mailbox (folder) on the server, use `Client.Create`. This corresponds to the IMAP `CREATE` command. Simply provide the name of the new mailbox:

```go
if err := c.Create("NewProject"); err != nil {
    // handle error (e.g., mailbox already exists or invalid name)
}
```

This will create a top-level mailbox named "NewProject". You can also create nested mailboxes by including the hierarchy delimiter (often `/`). For example: `c.Create("Projects/GoIMAPDemo")` might create a folder "GoIMAPDemo" under "Projects", if the server uses `/` as the delimiter.

**Note:** IMAP server hierarchy delimiters vary (you can get it from `MailboxInfo.Delimiter`). Also, some servers may not allow certain characters in names. Always check for errors.

## DELETE – Deleting a Mailbox

Use `Client.Delete("MailboxName")` to delete a mailbox. This issues the IMAP `DELETE` command. For example:

```go
if err := c.Delete("OldStuff"); err != nil {
    // handle error (e.g., mailbox not empty on some servers, or permissions)
}
```

Be careful: deleting a mailbox usually also deletes all messages in it on the server. Some servers might not allow deleting special mailboxes like "INBOX".

## RENAME – Renaming a Mailbox

`Client.Rename(oldName, newName)` issues the IMAP `RENAME` command, changing the name of a mailbox (and effectively moving it in the hierarchy if new name implies a different path). For example:

```go
err := c.Rename("Projects/OldName", "Projects/RenamedProject")
if err != nil {
    // handle error
}
```

If the rename is successful, the mailbox "Projects/OldName" will now appear as "Projects/RenamedProject". Renaming is atomic on the server if supported. Some servers may not support renaming certain folders (again, "INBOX" cannot typically be renamed per IMAP spec).

## SUBSCRIBE & UNSUBSCRIBE – Managing Subscriptions

IMAP allows clients to “subscribe” to mailboxes of interest. Subscribing doesn’t affect the mailbox’s existence or messages; it’s a client-side preference stored on the server. To subscribe or unsubscribe a mailbox, use `Client.Subscribe(name)` or `Client.Unsubscribe(name)`:

```go
if err := c.Subscribe("Newsletter"); err != nil {
    // handle error
}
// ... later ...
if err := c.Unsubscribe("Newsletter"); err != nil {
    // handle error
}
```

Subscribing a mailbox means it will show up in LSUB results (and possibly in the folder list in some email clients), whereas unsubscribing hides it from LSUB (though it will still appear in a full LIST). This is often used to declutter folder views.

## SELECT – Opening a Mailbox to Access Messages

Before retrieving or modifying messages, you must select a mailbox. The `SELECT` command makes a mailbox the current context (selected state). In `go-imap`, use `Client.Select(name, readOnly)` which returns an `*imap.MailboxStatus`. For example:

```go
mbox, err := c.Select("INBOX", false)
if err != nil {
    // handle error (e.g., mailbox doesn't exist or permission denied)
}
fmt.Println("Mailbox selected:", mbox.Name)
fmt.Println("Flags:", mbox.Flags)
fmt.Println("Total Messages:", mbox.Messages, "Recent:", mbox.Recent, "Unseen:", mbox.Unseen)
```

If `readOnly` is false, the mailbox is opened in read-write mode (the server will allow flag changes, deletions, etc.). If `readOnly` is true (which corresponds to the IMAP `EXAMINE` command), the mailbox is opened in read-only mode (no changes allowed, often used if you just want to read without marking messages as seen). The `MailboxStatus` (`mbox` above) provides information about the mailbox at select time, such as:  
- `Flags`: the list of flags the mailbox supports (e.g., `\Seen`, `\Answered`, custom keywords, etc.) ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=%2F%2F%20Select%20INBOX%20mbox%2C%20err,mbox.Flags)).  
- `Messages`: total number of messages in the mailbox.  
- `Recent`: number of messages with the `\Recent` flag (new messages since last time this mailbox was opened).  
- `Unseen`: number of messages not marked `\Seen` (the count of unread messages).  
- `UidNext`: the next UID that will be assigned to a new message (useful for synchronization).  
- `UidValidity`: an identifier for the mailbox’s UID numbering (if this changes, previously cached UIDs are invalid).

After `Select`, the client is in **Selected** state for that mailbox. Now we can fetch and search messages, set flags, etc., in this mailbox context.

## STATUS – Querying Mailbox Status Without Selecting

What if you want to know how many unread messages are in a mailbox without selecting it (which might be slow or have side effects)? IMAP provides `STATUS` for that. In `go-imap`, use `Client.Status(mailboxName, []StatusItem)` to fetch specific info about a mailbox. Example:

```go
items := []imap.StatusItem{imap.StatusMessages, imap.StatusUnseen}
status, err := c.Status("INBOX", items)
if err != nil {
    // handle error
}
fmt.Printf("Inbox has %d messages, %d unseen\n", status.Messages, status.Unseen)
```

Here we asked for total messages and unseen count. Other `StatusItem` values include `StatusRecent`, `StatusUidNext`, `StatusUidValidity`, etc. You can request multiple items in one call. This is useful for getting mailbox stats (new mail count, etc.) without disrupting the currently selected mailbox or changing state. `STATUS` can be used in Authenticated state (no mailbox selected).

## FETCH – Retrieving Messages (Headers, Body, Flags)

`FETCH` is one of the most commonly used IMAP commands – it retrieves data about messages (envelope, flags, body parts, etc.). With `go-imap`, `Client.Fetch(seqset, items, msgsChan)` is used. 

- **Sequence Set:** specifies which messages to fetch. It can be identified by sequence numbers (relative position in the mailbox) or, if using the UID variant, by UIDs (permanent IDs). We’ll discuss UIDs shortly. `imap.SeqSet` is used to build the set. You can add individual IDs or ranges. For example:
  ```go
  seqset := new(imap.SeqSet)
  seqset.AddNum(1)            // message #1
  seqset.AddRange(10, 15)     // messages #10 through #15
  ```
  Sequence numbers refer to the currently selected mailbox. `1` is the first message (usually oldest), and a sequence can also use `*` to mean the last message.

- **Items:** specify what data to fetch. `go-imap` defines constants for common items:
  - `imap.FetchEnvelope` – fetches the envelope (from, to, subject, date, etc. metadata).
  - `imap.FetchFlags` – fetches flags set on the message.
  - `imap.FetchInternalDate` – gets the internal received date.
  - `imap.FetchRFC822` – gets the full raw message (all headers and body) as a single item.
  - `imap.FetchRFC822Header` – gets just the full headers.
  - `imap.FetchBody` or `imap.FetchBodyStructure` – gets the body structure metadata.
  - You can also specify body sections like `"BODY[TEXT]"` or `"BODY[HEADER]"` or `"BODY[1.2]"` (to fetch a specific MIME part). In `go-imap`, you supply these as strings via `imap.FetchItem` (which is just `string` type).  
  For example, to fetch the entire raw email without marking it seen, you might use: `imap.FetchItem("BODY.PEEK[]")` (the `PEEK` avoids setting the `\Seen` flag).

- **Channel:** where fetched messages will be delivered. The library streams `*imap.Message` objects into this channel.

**Example 1: Fetch envelopes (metadata) of the last N messages.** This is useful to list subjects or senders without downloading full emails:

```go
// Determine the last 10 messages in the mailbox
seqset := new(imap.SeqSet)
start := uint32(1)
if mbox.Messages > 10 {
    start = mbox.Messages - 9  // last 10 messages
}
seqset.AddRange(start, mbox.Messages)

// Prepare channel and fetch
messages := make(chan *imap.Message, 10)
go func() {
    err = c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}, messages)
}()
for msg := range messages {
    fmt.Printf("#%d: %s | Flags: %v\n", msg.SeqNum, msg.Envelope.Subject, msg.Flags)
}
if err != nil {
    // handle error
}
``` 

This will output something like:
```
#5: Re: Meeting Notes | Flags: [\Seen]
#6: Fwd: Annual Report | Flags: [\Seen \Answered]
#7: [No Subject] | Flags: []
...
``` 
We requested each message’s `Envelope` and `Flags`. The `Envelope` includes fields like `Envelope.From`, `Envelope.To`, `Envelope.Subject`, `Envelope.Date`, etc. The `Flags` slice might contain `\Seen`, `\Answered`, `\Flagged`, custom tags, etc. Fetching flags in the same query is more efficient than separate calls.

**Example 2: Fetch a full message by UID and parse it.** Suppose we want to download an email (including attachments) to process it:

```go
// Assume we have a known UID of a message (e.g., from a search or previous fetch)
uid := uint32(1234)
seqset := new(imap.SeqSet)
seqset.AddNum(uid)

// We will fetch the full message body without marking it seen:
item := imap.FetchItem("BODY.PEEK[]")
msgCh := make(chan *imap.Message, 1)
if err := c.UidFetch(seqset, []imap.FetchItem{item}, msgCh); err != nil {
    log.Fatal(err)
}
msg := <-msgCh
r := msg.GetBody("BODY[]")  // get the raw body IMAP literal
if r == nil {
    log.Fatal("Server did not return message body")
}
// Now use go-message to parse r (see next section)
```

Here we used `UidFetch` (since we had a UID) and requested the entire message (`BODY[]`). We used `PEEK` to avoid setting the `\Seen` flag on the message when fetching ([How to read an email's message body using the *imap.Message (emersion/go-imap) - Stack Overflow](https://stackoverflow.com/questions/70540892/how-to-read-an-emails-message-body-using-the-imap-message-emersion-go-imap#:~:text=Check%20first%20if%20emersion%2Fgo,306%20would%20apply)). The `msg.GetBody("BODY[]")` gives us an `io.Reader` for the raw message content. We will parse this with `go-message` in the next section.

**Tip:** If you only need certain parts of the email (say just the text body and not attachments), you can fetch specific body sections. First, fetch the `BODYSTRUCTURE` (which gives a tree of MIME parts) and find the part number you want. Then fetch `BODY[<part specifier>]`. For example, if the BODYSTRUCTURE shows part 1.2 is the HTML body, you could do `Fetch(seqset, []imap.FetchItem{imap.FetchBody + "[1.2]"}, ...)`. This can avoid downloading large attachments when not needed. However, this is advanced usage – many clients simply fetch the whole message or at least the whole BODY[] and use MIME parsing to extract what they need. Keep in mind that fetching BODYSTRUCTURE and then specific parts requires multiple round trips. Use it only if bandwidth or memory usage of attachments is a concern ([How to get the message content? · Issue #72 · emersion/go-imap · GitHub](https://github.com/emersion/go-imap/issues/72#:~:text=If%20you%20only%20want%20the,1.3.2)).

**Processing `imap.Message`:** The `*imap.Message` object returned by fetch may contain multiple fields depending on what was requested. For example:
- `msg.Envelope` (if `FetchEnvelope` was in items).
- `msg.Flags` (if flags were fetched).
- `msg.BodyStructure` (if requested).
- `msg.Body` – a map of section name to literal data (if any BODY[...] sections were fetched). `msg.GetBody(sectionName)` is a helper to get the `io.Reader` for a section.

**Marking messages Seen:** By default, using `BODY[]` in a fetch (without `PEEK`) will mark messages as `\Seen` (read). If you want to retrieve the message without marking it read, always use `BODY.PEEK[...]` ([How to read an email's message body using the *imap.Message (emersion/go-imap) - Stack Overflow](https://stackoverflow.com/questions/70540892/how-to-read-an-emails-message-body-using-the-imap-message-emersion-go-imap#:~:text=Check%20first%20if%20emersion%2Fgo,306%20would%20apply)). The `ENVELOPE`, `FLAGS`, etc., can be fetched without affecting `\Seen`. Only BODY and BODYSTRUCTURE are sensitive to `PEEK`.

## SEARCH – Finding Messages by Criteria

The `SEARCH` command lets you query messages in the selected mailbox by various criteria (sender, subject, flags, dates, etc.). In `go-imap`, use `Client.Search(criteria)`. The criteria are built using the `imap.SearchCriteria` struct. You can combine multiple criteria.

Example: find all unseen messages (not `\Seen`) in the current mailbox:

```go
criteria := imap.NewSearchCriteria()
criteria.WithoutFlags = []string{imap.SeenFlag}  // messages that do NOT have \Seen
ids, err := c.Search(criteria)
if err != nil {
    // handle error
}
fmt.Printf("Unseen message sequence numbers: %v\n", ids)
```

If successful, `ids` will be a slice of **sequence numbers** of messages that match. You could then fetch those messages or parts of them. Common search criteria include:
- `criteria.WithFlags` / `WithoutFlags` – filter by flags present or absent (e.g., `imap.SeenFlag`, `imap.AnsweredFlag`, etc.).
- `criteria.Header` – a map of header fields to match values (e.g., `criteria.Header.Add("FROM", "alice@example.com")` to find messages from Alice).
- `criteria.Text` or `Body` – strings to search in the entire message or just body.
- `criteria.Since` / `Before` – filter by internal date.
- `criteria.Uid` or `SeqNum` – you can directly specify a set of UIDs or sequence numbers to search within.
- `criteria.Or` / `criteria.Not` – combine conditions logically.

For example, to search for messages from Alice that are unseen:
```go
crit := imap.NewSearchCriteria()
crit.WithoutFlags = []string{imap.SeenFlag}
crit.Header = textproto.MIMEHeader{}
crit.Header.Add("From", "alice@example.com")
ids, err := c.Search(crit)
// ...
```
The `ids` returned by `c.Search` are sequence numbers by default. If you want UIDs, use `c.UidSearch(crit)` which returns UIDs.

**Using UIDs in search:** Often you'll use UidSearch so that you get stable identifiers. For example:
```go
uids, err := c.UidSearch(criteria)
```
Now `uids` is a slice of UID numbers, which you can use with `UidFetch` or other UID commands.

**Note:** IMAP search is case-insensitive for text and can be quite powerful. However, not all servers support complex searches efficiently. Some servers might not support searching the BODY text unless explicitly allowed. `go-imap` just sends the SEARCH query; the server does the heavy lifting.

## STORE – Setting and Clearing Flags on Messages

The `STORE` command modifies message flags (read/unread, answered, etc.) or other message data. Typically, it's used to add or remove flags like `\Seen` (mark as read), `\Flagged` (star), `\Deleted` (mark for deletion), etc., or to set custom keywords.

In `go-imap`, `Client.Store(seqset, item, value, msgChan)` is used. The **`item`** specifies the operation (flags add/remove/replace). You can build it with `imap.FormatFlagsOp(op, silent)` where `op` is one of:
- `imap.AddFlags` (add the flags),
- `imap.RemoveFlags` (remove the flags),
- `imap.SetFlags` (replace with these flags exactly).

The `silent` boolean, if true, uses the `.SILENT` variant to not return updated flags (to avoid unsolicited FETCH responses).

The **`value`** is typically a slice of flag names (strings) to add/remove. System flags are represented by constants: e.g., `imap.SeenFlag = "\Seen"`, `imap.AnsweredFlag = "\Answered"`, `imap.FlaggedFlag = "\Flagged"`, `imap.DeletedFlag = "\Deleted"`, `imap.DraftFlag = "\Draft"`, etc. ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=SeenFlag%20%20%20%20,)). You can also use custom flags (keywords) as strings (servers may allow defining new ones).

**Example: Mark a message as read (\Seen).**

```go
seqset := new(imap.SeqSet)
seqset.AddNum(5)  // message #5
item := imap.FormatFlagsOp(imap.AddFlags, true)      // +Flags.SILENT
flags := []interface{}{imap.SeenFlag}
if err := c.Store(seqset, item, flags, nil); err != nil {
    fmt.Println("Error setting flag:", err)
}
```

This will add the `\Seen` flag to message 5. We used the silent variant (`true` for the second argument), so the server won’t send back an update to the flags (we don’t need a response). If we wanted to get the updated flags, we would set `silent=false` and provide a channel to receive the updated message data:

```go
item = imap.FormatFlagsOp(imap.RemoveFlags, false)   // -Flags (not silent)
flags = []interface{}{imap.FlaggedFlag}
updated := make(chan *imap.Message, 1)
go func() {
    _ = c.Store(seqset, item, flags, updated)
}()
msg := <-updated
fmt.Println("Updated flags for UID 42:", msg.Flags)
```

This example removes the `\Flagged` flag (star) from a message and prints the updated flags. We assumed `seqset` corresponds to the message of interest (e.g., we might have built it with a UID in UidStore scenario).

**Flags and their meaning:** Standard IMAP flags include ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=SeenFlag%20%20%20%20,)):
- `\Seen` – Message has been read.
- `\Answered` – Message has been replied to.
- `\Flagged` – Message is flagged (starred/important).
- `\Deleted` – Message is marked for deletion (will be removed on EXPUNGE).
- `\Draft` – Message is marked as a draft.
- `\Recent` – Message is recent (this flag is set by server on new messages and is special: it’s not persisted across sessions, and you **cannot set or unset \Recent** manually).

You can define custom flags (keywords) like `$Urgent` or any arbitrary string without `\`. Many servers allow it, and `MailboxStatus.PermanentFlags` will indicate which flags are allowed or if new ones can be created (some servers advertise a special `\*` in PermanentFlags meaning any new flag can be used) ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=const%20TryCreateFlag%20%3D%20)).

**Marking messages deleted:** In IMAP, deletion is a two-step process: you mark a message with the `\Deleted` flag via STORE, and then later use EXPUNGE to permanently remove all `\Deleted` messages. For example:
```go
seqset := new(imap.SeqSet); seqset.AddNum(10)
item := imap.FormatFlagsOp(imap.AddFlags, true)
_ = c.Store(seqset, item, []interface{}{imap.DeletedFlag}, nil)
```
Now message 10 is marked deleted (it may disappear from some clients’ view immediately, depending on client). It’s not gone until expunged (see EXPUNGE section).

**UID version:** There is also `Client.UidStore`, which works the same but `seqset` should contain UIDs. Use that if you identified messages by UID.

## COPY – Duplicating Messages to Another Mailbox

The `COPY` command copies messages from the current mailbox to another mailbox. It’s like a “copy-paste” for emails on the server (originals remain in source mailbox). In `go-imap`, use `Client.Copy(seqset, destMailbox)`.

Example: copy messages 1-5 to the "Archive" mailbox:

```go
seqset := new(imap.SeqSet)
seqset.AddRange(1, 5)
if err := c.Copy(seqset, "Archive"); err != nil {
    fmt.Println("Copy failed:", err)
}
```

If the server succeeds, messages 1 through 5 of the current mailbox are now duplicated into "Archive". Their flags (except `\Recent`) are typically preserved in the target mailbox. The `seqset` here is sequence numbers in the source mailbox context. If you want to copy by UID, use `Client.UidCopy(uidSeqset, dest)`.

One common use of COPY is to move messages if the server doesn’t support the MOVE command (clients used to simulate “move” by COPY then marking original `\Deleted`). We’ll discuss MOVE next.

## MOVE – Moving Messages to Another Mailbox (Atomic)

`MOVE` is an IMAP extension (RFC 6851) that atomically moves messages to a new mailbox (copy + delete original in one step). Many modern servers support it (and `go-imap` supports it if the server does). Use `Client.Move(seqset, destMailbox)`:

```go
uids := new(imap.SeqSet)
uids.AddNum(1001, 1002, 1003)  // message UIDs to move
if err := c.UidMove(uids, "Trash"); err != nil {
    fmt.Println("Move not supported or failed, falling back to copy+delete")
    // If MOVE not supported, we could fallback:
    // c.UidCopy(uids, "Trash"); then c.UidStore(uids, +Deleted)
}
```

In this snippet, we attempted to move messages with UIDs 1001-1003 to "Trash". We used `UidMove` because we had UIDs (you could also do `c.Move` with sequence numbers if you have them from a select session). If the server doesn’t support `MOVE`, the command might result in an error (and indeed, `go-imap` will return an error). We then show a fallback: copying the messages then marking them deleted in the source. (After that, you would expunge to remove them.) You can check server capabilities via `c.Support("MOVE")` to decide whether to use move or fallback, but often just trying `Move` and catching error is fine.

When `MOVE` succeeds, the messages appear in the destination mailbox and are removed from the source mailbox in one server roundtrip. This simplifies client logic and is usually faster.

## APPEND – Adding a New Message to a Mailbox

`APPEND` allows the client to upload a message to a mailbox (useful for saving sent mail to a "Sent" folder, importing messages, or adding to drafts). With `go-imap`, use `Client.Append(mbox, flags []string, date time.Time, msg imap.Literal)`.

Parameters:
- **mbox:** Name of the mailbox to append into (must already exist, or some servers allow creating on append with a special flag).
- **flags:** Initial flags for the message (e.g., `\Seen` if you want the message to be marked as read immediately, or `\Draft` if saving a draft). You can pass nil or an empty slice for no flags.
- **date:** The internal date timestamp for the message (often you use `time.Now()` or the original date of the email). This is the date that will show as the "received" date in that mailbox. If you don’t care, you can use `time.Now()` or `imap.TimeNow` helper.
- **msg:** The raw message to append, as an `imap.Literal`. `imap.Literal` is basically an `io.Reader` or byte slice with a known length. You can use `imap.NewLiteral([]byte)` or simply pass an object that implements the `imap.Literal` interface. E.g., `bytes.NewReader(emailBytes)` often suffices (go-imap has an internal wrapper for readers to make them Literal).

Example: Append a message to Sent folder.

```go
raw := []byte("From: me@example.com\r\nTo: you@example.com\r\nSubject: Hello\r\n\r\nThis is the body.")
// We have raw message bytes (with CRLF line endings as required by IMAP).
err := c.Append("Sent", []string{imap.SeenFlag}, time.Now(), imap.NewLiteral(raw))
if err != nil {
    fmt.Println("Append failed:", err)
}
```

This will upload the given message to the "Sent" mailbox, mark it as \Seen (so it doesn’t show as unread in Sent), with the current time as the internal date. If the server returns OK, the message is now stored on the server in that folder. The server will assign it a UID in that mailbox (which you could obtain via an `UidNext` query in STATUS or SELECT after append, or some servers support an `APPENDUID` response if they have UIDPLUS extension).

**Note:** Ensure the message is well-formed with proper MIME headers and CRLF newlines. The server will not modify the content; it stores it as-is. Using the `go-message` library (or Go’s `net/mail`) to construct the MIME text can be helpful if generating emails.

## EXPUNGE – Permanently Removing Deleted Messages

The `EXPUNGE` command permanently deletes all messages marked with `\Deleted` in the currently selected mailbox. After an EXPUNGE, those message UIDs/sequence numbers become invalid and are freed up. In IMAP, when an expunge happens, the server typically sends unsolicited `EXPUNGE` responses to inform the client which messages were removed (often by sequence number). This causes remaining messages’ sequence numbers to shift down to fill the gaps.

In `go-imap`, `Client.Expunge(expungeChan)` triggers an expunge. You can pass a channel to receive expunged message sequence numbers (each expunged message will be reported). The command is asynchronous like FETCH/LIST:

```go
expungeChan := make(chan uint32)
go func() {
    _ = c.Expunge(expungeChan)
}()
for seq := range expungeChan {
    fmt.Println("Message with sequence number", seq, "was expunged")
}
```

If you don’t care to get the list of expunged messages, you can call `c.Expunge(nil)` or simply ignore the channel results. After expunge, those messages are gone from the mailbox on the server.

**Important:** If other clients are connected to the same mailbox, they will be notified of the expunge too (via their `Client.Updates` channel as `ExpungeUpdate` events in `go-imap`). If you are maintaining a local cache of UIDs, you need to remove the expunged ones. If you only track by UIDs, note that when messages are expunged, their UIDs are gone from the mailbox (but other UIDs remain the same; UIDs are not reused as long as `UidValidity` stays the same).

If the mailbox was opened read-write, an `EXPUNGE` is allowed. If it was opened read-only (`Select(..., true)` or `Examine`), the server will typically reject expunge (you’d have to re-select in write mode or the server might allow but it’s not standard).

**CLOSE vs. EXPUNGE:** Issuing a `Client.Close()` will also expunge and then close the mailbox. IMAP `CLOSE` command implies an expunge of deleted messages then leaving the mailbox. In `go-imap`, `Client.Close()` can be used instead of separate Expunge + ending the session (this returns to authenticated state). If you want to deselect without expunging, use `Client.Unselect()` (if server supports the `UNSELECT` extension ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=Extensions)) ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=%2A%20MOVE%20%2A%20SASL,USE%20%2A%20UNSELECT))). Unselect cleanly deselects the mailbox without removing `\Deleted` messages.

## IDLE – Real-time Mailbox Updates (Push Notifications)

The `IDLE` command (RFC 2177) allows the client to tell the server it’s ready to receive updates. While idling, the server can push updates like new emails (`EXISTS` responses indicating new message count, or `RECENT` count changes), flag changes, expunges, etc., without the client polling. 

`go-imap` supports IDLE via `Client.Idle(stopChan, options)`. Typically, you provide a channel (`stopChan`) that you can close or send to in order to terminate the idle, and an `IdleOptions` (which can be nil or used to set a custom timeout for auto-resume behavior).

Basic usage example:

```go
stop := make(chan struct{})
// Start idling in a goroutine
go func() {
    if err := c.Idle(stop, nil); err != nil {
        fmt.Println("Idle error:", err)
    }
}()
fmt.Println("Client is now IDLE, waiting for updates...")

// ... elsewhere, when you want to stop idling (or after some time):
close(stop) // this will terminate the Idle
```

When `Idle` is active, the server will send unsolicited responses for new messages, expunges, or flag changes. These will be delivered to `c.Updates` channel in `go-imap` (if you have set it up). For example, you can set `c.Updates = make(chan client.Update)` before calling idle, and then listen on that channel:

```go
for update := range c.Updates {
    switch u := update.(type) {
    case *client.MailboxUpdate:
        // Mailbox status updated (e.g., new message arrived)
        fmt.Println("Mailbox update:", u.Mailbox.Name, "exists:", u.Mailbox.Messages)
    case *client.MessageUpdate:
        // Message flags updated
        fmt.Println("Message", u.Message.Uid, "flags updated to", u.Message.Flags)
    case *client.ExpungeUpdate:
        // Message expunged
        fmt.Println("Message expunged, SeqNum:", u.SeqNum)
    }
}
```

Each update type corresponds to a server event. For instance, if a new email arrives, the server might send an "* n EXISTS" message, which `go-imap` translates into a `MailboxUpdate` (with Mailbox status containing new message count). An expunge would produce an `ExpungeUpdate`. By handling these, you can keep your client’s state in sync in real-time.

**IDLE cycle:** IMAP servers typically require that the client come out of IDLE periodically (at least every 29 minutes as per RFC, or sooner, some servers send a periodic keepalive). `go-imap` will end the `Idle` call when you close the `stop` channel. You can also manually break idle by sending the "DONE" command. The pattern often used is:
- Start Idle.
- Wait until an update is received (or a timeout).
- When an update comes, exit idle (stop channel) so you can issue a FETCH or whatever to get the new data.
- Go back into IDLE after processing, to continue getting updates.

There are helper packages like `go-imap-idle` that implement an idle loop with fallback to noop, but you can also implement it yourself.

**No Concurrent Commands in Idle:** Remember, while the client is idling, you shouldn’t send other commands on the same connection until you terminate the idle. If you need to perform other actions in parallel, either break out of idle or use a second connection.

## NOOP – Keeping the Connection Alive / Checking for Updates

`NOOP` is a do-nothing command that the server will simply acknowledge. Its main use is to allow the server to send any pending unilateral updates. Clients often send NOOP when idle (not in the IDLE command, but being inactive) to poll for changes and to keep the TCP connection from timing out.

Usage in `go-imap` is trivial:

```go
if err := c.Noop(); err != nil {
    fmt.Println("Noop error:", err)
}
```

If any messages arrived or were expunged since the last command, the server may send untagged `EXISTS`, `EXPUNGE`, etc., during the NOOP. `go-imap` will deliver those to the `c.Updates` channel or update the `MailboxStatus` in the `Client.Mailbox` if one is selected. So NOOP is a way to flush those updates. It’s common to do a NOOP every few minutes if not using IDLE, to check for new mail.

## UIDs vs Sequence Numbers – Using UID Commands for Stability

Throughout the above sections, we mentioned sequence numbers and UIDs. Here’s a summary and best practices:

- **Sequence Numbers** are 1-based indexes of messages in the currently selected mailbox. They can change whenever messages are added or removed. For example, if you delete message 3, then what was message 4 becomes message 3, etc. Sequence numbers are only valid during a session and while no intervening changes happen.
- **UIDs (Unique IDs)** are assigned to messages by the server and do not change for the life of the message in that mailbox. Even if other messages are deleted, each message keeps its UID. UIDs are guaranteed to be unique within a mailbox and never reused as long as the mailbox’s `UIDVALIDITY` stays the same. If the server resets the UIDVALIDITY (say the mailbox was reconstructed), then all previous UIDs are invalid.

For reliable reference to messages across sessions or after expunges, use UIDs. For example, if you’re syncing emails, you’d record the last seen UID and next time start from the next one.

**Using UIDs in `go-imap`:** The library provides parallel methods for all message-specific commands:
- `Client.UidFetch` instead of `Fetch`
- `Client.UidSearch` instead of `Search`
- `Client.UidStore` instead of `Store`
- `Client.UidCopy` instead of `Copy`
- `Client.UidMove` instead of `Move`

These take the same arguments, except the sequence set values are interpreted as UIDs. Also, when you do a regular `Search`, the server by default returns sequence numbers. `UidSearch` tells the server to return UIDs. Similarly, `Fetch` vs `UidFetch` changes whether the numbers in the `SeqSet` are seqnums or UIDs and whether the `Message.Uid` field is populated. Typically:
- If you call `UidFetch`, you should interpret any `msg.SeqNum` with caution (it’s the transient sequence at that time), and use `msg.Uid` (which the server will include since you did a UID fetch, it usually sends `UID xyz` in the response). In `go-imap`, if you do a UID command, it automatically sets the `Message.Uid` for you. If you do a non-UID fetch and want the UIDs, include `imap.FetchUid` in the item list to have the server include UIDs ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=FetchUid%20%20%20%20,)).
- `UidStore` will apply flags to messages identified by UIDs.
- `UidCopy` copies by UID.

**Example: Fetch by UID after a search.**

```go
// Search for unseen messages and get UIDs
criteria := imap.NewSearchCriteria()
criteria.WithoutFlags = []string{imap.SeenFlag}
unseenUIDs, err := c.UidSearch(criteria)
if err != nil {
    // handle error
}
fmt.Println("Unseen message UIDs:", unseenUIDs)

// Fetch the envelopes of those unseen messages by UID
uidSet := new(imap.SeqSet)
for _, uid := range unseenUIDs {
    uidSet.AddNum(uid)
}
msgs := make(chan *imap.Message, len(unseenUIDs))
go func() { _ = c.UidFetch(uidSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags}, msgs) }()
for msg := range msgs {
    fmt.Printf("UID %d: %s (Flags: %v)\n", msg.Uid, msg.Envelope.Subject, msg.Flags)
}
```

By using UIDs, we ensure that even if another client expunges some messages, we are still referencing the correct ones. 

**When to use sequence numbers:** If you are iterating over messages in one mailbox session (like fetching 1:* in one go, or processing batch by batch), sequence numbers are fine *within that session*. They might be slightly faster on the server (no need to map to UIDs). But for any persistent reference (like storing “last read message” or syncing in an app), always use UIDs.

## Parsing MIME Messages with `go-message`

Once you have fetched a message’s raw bytes or a section of it, you need to parse it to extract headers and body parts. The `emersion/go-message` library is a streaming MIME parser that makes this easier than manually splitting by boundaries. It deals with decoding transfer encodings (base64, quoted-printable) and charsets, and provides easy access to common header fields.

For email parsing, we particularly use the subpackage `mail` in `go-message` (`github.com/emersion/go-message/mail`). This provides high-level types for mail messages.

### Reading a Message and Headers

Suppose we have an `io.Reader` `r` containing the raw email (for example, from `msg.GetBody("BODY[]")` as shown earlier). We can create a `mail.Reader`:

```go
mr, err := mail.CreateReader(r)
if err != nil {
    log.Fatal("Failed to create mail reader:", err)
}
// The mail.Reader now parses the message header
header := mr.Header
from, _ := header.AddressList("From")
subject, _ := header.Subject()
date, _ := header.Date()
// Print header info:
fmt.Println("From:", from)
fmt.Println("Subject:", subject)
fmt.Println("Date:", date)
```

`mail.CreateReader(r)` reads the message headers and prepares to read the body. We can then query common headers:
- `header.AddressList(key)` to parse addresses (returns a list of `*mail.Address` with Name and Address fields).
- `header.Get(key)` to get a header value as string (for multiple values, use `header.Values(key)`).
- `header.Subject()` convenience method to get and decode the Subject.
- `header.Date()` to get the Date as `time.Time`.
- There are also methods like `header.MessageID()`, `header.AddressList("To")`, etc., for common fields. These handle any RFC2047 encoded-words and such, returning decoded values ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=,Header%29%20SetDate%28t%20time.Time)) ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=,Header%29%20SetDate%28t%20time.Time)).

### Iterating through Body Parts

After reading the headers, you can iterate through the message parts using `mr.NextPart()`:

```go
for {
    p, err := mr.NextPart()
    if err == io.EOF {
        break // no more parts
    }
    if err != nil {
        log.Fatal("Failed to read part:", err)
    }

    // Check the part's header to see if it's an attachment or inline content
    switch h := p.Header.(type) {
    case *mail.InlineHeader:
        // This is an inline part (text or HTML content)
        b, _ := io.ReadAll(p.Body)
        // You might check Content-Type to decide how to use it:
        ct := h.Get("Content-Type")
        fmt.Printf("Got inline part: Content-Type=%s, %d bytes\n", ct, len(b))
        // If it's text, you could decode to string (assuming UTF-8 or known charset)
        // mail.InlineHeader provides convenience if needed, e.g. no filename to extract.
    case *mail.AttachmentHeader:
        // This is an attachment
        filename, _ := h.Filename()
        fmt.Println("Got attachment:", filename)
        data, _ := io.ReadAll(p.Body)
        fmt.Printf("Attachment %s is %d bytes\n", filename, len(data))
        // You may want to save data to file or process it.
    }
}
```

Using `mail.Reader`, each part’s Header is either an `InlineHeader` (for inline body parts, often the main text or HTML parts) or an `AttachmentHeader` (for attachments, which typically have a filename) ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=switch%20h%20%3A%3D%20p.Header.%28type%29%20,v%5Cn%22%2C%20filename%29%20%7D)). We read all data from `p.Body` here for simplicity, but be mindful that attachments could be large. You might stream them to disk instead of reading fully into memory.

The library takes care of decoding the transfer encoding. So `p.Body` gives you the raw content of that part (e.g., the actual text, or binary data of an image, etc.). If a part is text and in a charset like iso-8859-1, `CreateReader` will handle charset conversion to UTF-8 by default (if possible), so `p.Body` yields UTF-8 text bytes.

For attachments, you often want the filename (`AttachmentHeader.Filename()`), and possibly the MIME type (which you can get from the part header’s Content-Type, via `h.ContentType()` or `h.Get("Content-Type")`). 

**Nested Multiparts:** The code above flattens the structure one level: `mail.CreateReader` will handle typical multipart emails by returning each subpart in sequence. For example, if an email is `multipart/alternative` (text and HTML), and the whole email is enclosed in `multipart/mixed` with an attachment, the reader might return parts in this order:
1. text/plain (InlineHeader)
2. text/html (InlineHeader)
3. attachment (AttachmentHeader)

Actually, the `mail` package docs state it assumes one or more text parts and attachments ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=Package%20mail%20implements%20reading%20and,writing%20mail%20messages)). It will present each alternative text as a separate inline part. If the structure is more complex (e.g., nested multiparts like an email with inline images which is `multipart/related` inside the HTML alternative), how does `mail.Reader` handle it? In practice, it might flatten some but not all. If it encounters a `message/rfc822` part (an email forwarded as attachment), that might come as an attachment part as well (with its own structure to parse recursively by creating another `mail.Reader` from that part’s body). 

For full control of MIME structure, you can use the lower-level `message` package. For example, `message.Read(r)` gives a `*message.Entity` which can be recursively inspected: it has `e.Header` and `e.MultipartReader()` if it’s multipart. There’s also `message.Walk` to recursively walk through MIME parts ([message package - github.com/emersion/go-message - Go Packages](https://pkg.go.dev/github.com/emersion/go-message#:~:text=type%20WalkFunc%20func%28path%20,99)). But for most email use cases, `mail.CreateReader` and iterating parts as above suffices and is more convenient.

### Decoding Attachments and Content

In the above loop, we used `io.ReadAll` to get the content. Real-world usage:
- **Text parts:** You may decode them to string. If you know the MIME type, e.g., content-type text/plain vs text/html, you can handle accordingly. `go-message` does not parse HTML for you (that’s up to you if needed), but it gives the raw HTML text. For text, it will handle quoted-printable or base64 decoding automatically.
- **Attachments:** If an attachment is base64 encoded (which most are), `p.Body` is already the decoded binary data. You can directly write it to a file. The Content-Type header will tell you the MIME type (e.g., image/png, application/pdf, etc.), and AttachmentHeader.Filename gives the suggested name. 

Also note, headers like `Content-Disposition` and `Content-Type` parameters can indicate if something is an attachment or inline. The `mail.Reader` logic uses `Content-Disposition: attachment` to decide to give you `AttachmentHeader`. Inline images might come as InlineHeader but with a Content-ID. If you need to handle inline images (for an HTML email), you might see them as attachments in this API (since they have filename perhaps). You can differentiate by checking if the disposition is "inline" vs "attachment".

### Handling Email Encodings Gotchas

- **Character sets:** If you see an error about unknown charset, `CreateReader` may still return a Reader with a special error that you can inspect (it tries to handle what it can) ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=func%20CreateReader%28r%20io%20,Reader%20%2C%20%2089)) ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=NextPart%20returns%20the%20next%20mail,EOF%20is%20returned%20as%20error)). Common charsets like UTF-8, ISO-8859-*, etc., are supported.
- **Large attachments:** Avoid reading all at once if not necessary. Stream to file.
- **Multiple attachments or nested parts:** Approach is similar – the loop will iterate through them.
- **Encrypted or Signed emails:** Those come as `multipart/signed` or application/pgp etc. `go-message` doesn’t verify signatures or decrypt, but you can extract the parts (e.g., one part will be the signed data, another the signature).

### Example: Putting it together

Suppose we fetched an email and want to print out basic info and save attachments:

```go
msgReader, err := mail.CreateReader(r)
if err != nil { ... }

header := msgReader.Header
from, _ := header.AddressList("From")
subject, _ := header.Subject()
fmt.Printf("From: %s\nSubject: %s\n", from[0].Address, subject)

for {
    part, err := msgReader.NextPart()
    if err == io.EOF {
        break
    } else if err != nil {
        log.Fatal(err)
    }

    switch h := part.Header.(type) {
    case *mail.InlineHeader:
        contentType := h.Get("Content-Type")
        body, _ := io.ReadAll(part.Body)
        if strings.HasPrefix(contentType, "text/plain") {
            fmt.Println("Body text:", string(body))
        } else if strings.HasPrefix(contentType, "text/html") {
            fmt.Println("HTML body length:", len(body))
            // (Maybe save or render HTML)
        }
    case *mail.AttachmentHeader:
        filename, _ := h.Filename()
        fmt.Println("Saving attachment:", filename)
        f, _ := os.Create(filename)
        io.Copy(f, part.Body)
        f.Close()
    }
}
```

This would output the sender and subject, print the plain text body (if present), note the length of HTML (if present), and save any attachments to files.

## Tips for Efficient IMAP Usage and Synchronization

Finally, let's discuss some best practices and gotchas when using `go-imap` in a real application:

- **Use UIDs for Sync:** As mentioned, when building an email client or sync service, always track messages by UID. For example, to sync a mailbox, you might:
  1. Perform an `UID SEARCH ALL` to get all UIDs, or use `UID SEARCH SINCE <date>` for incremental.
  2. Compare with locally stored UIDs to find new ones or ones that disappeared (expunged).
  3. Fetch new messages by UID as needed.
  4. If you need to sync flag changes, the IMAP `FETCH UID … FLAGS` or `UID SEARCH FLAGGED` etc., can be used, or use the IDLE updates.
  Also watch for `UIDVALIDITY`: if it changes, you should invalidate your cache of that mailbox (it means the server reset UIDs, so all bets are off – typically you then re-download or reconcile carefully).

- **Batch requests:** IMAP allows requesting multiple items in one command. `go-imap` makes this easy (just add more to the `SeqSet`). For example, fetching 100 messages in one `Fetch` is much more efficient than 100 separate `Fetch` calls. The server will stream them and you handle them in one go. The example above fetching a range is a good pattern. You could fetch by batches of UIDs if you don't want one giant fetch.

- **Avoiding blocking and concurrency issues:** Because `go-imap` is not thread-safe for concurrent commands on one connection ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=It%20is%20not%20safe%20to,commands%20on%20the%20same%20connection)), structure your code to send one command at a time and wait for it to complete (or use channels as shown). If you need concurrency (e.g., fetching two mailboxes simultaneously), open two `Client` connections. Each can log in (possibly same account or different) and operate independently. Just be mindful of server connection limits (some servers allow only a few concurrent connections per account).

- **Processing while fetching:** The streaming design of `go-imap` lets you start processing messages as they come in through the channel. For example, you could spawn a goroutine to parse each message as it arrives from a `Fetch`, overlapping network receive and processing. But do not forget to read from the channel or you’ll stall the fetch. Using a buffered channel (as shown in examples) gives some leeway if your processing is a bit slower than the network.

- **Memory usage:** Fetching a very large email (`BODY[]`) will allocate an `imap.Literal`. By default, `go-imap` may buffer the literal entirely in memory. Keep an eye on memory if downloading huge attachments. It might be better to fetch headers first, decide if you need the attachment, etc. There is an IMAP extension `BINARY` for chunking, but not widely used. In practice, if you anticipate huge messages, make sure you have enough memory or handle in chunks (the library as of v1 doesn’t stream partial content out, it waits until it's fully received to give you the `io.Reader`). In `go-imap v2` this might improve, but v1 will buffer the literal.

- **New mail detection:** The most efficient way to stay updated is to use `IDLE`. It avoids constant polling. You could combine IDLE with an occasional NOOP (some servers disconnect idle after a timeout without telling the client). If not using IDLE, a common pattern is to do a `STATUS` or `NOOP` every X minutes to see if messages or unseen count changed, then do a `FETCH` or `SEARCH` for new messages. IDLE simplifies this by pushing updates to you. We saw how to use IDLE with the Updates channel to handle new messages.

- **Server support and extensions:** Not all servers support all commands. For instance, `MOVE`, `UNSELECT`, `UIDPLUS` (which provides `APPENDUID` and `UID EXPUNGE`), etc., may or may not be available. You can check capabilities:
  ```go
  caps, err := c.Capability()
  if err == nil {
      if caps["MOVE"] {
          // server supports MOVE
      }
  }
  ```
  `Capability()` returns a map of capability -> bool ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=,Handler)). This must be called in Authenticated state (or some servers allow in Not Authenticated as well, typically servers send capabilities upon connect and after login automatically). `c.Support("MOVE")` is a convenience that likely does similar check ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=,Client%29%20SupportStartTLS%28%29%20%28bool%2C%20error)). Plan fallbacks for missing features (as we did for move).

- **Handling server bugs or quirks:** Some IMAP servers have quirks. A robust client might need to handle:
  - Gmail IMAP has some custom flags like `\Important` (and an `$Important` keyword) ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=const%20ImportantAttr%20%3D%20)), and it doesn't support `UID EXPUNGE` (it uses X-GM-EXT-1 extensions for advanced search and labels, which is beyond standard IMAP).
  - Some servers might not update `Unseen` count reliably; it’s safer to do a search for unseen if critical.
  - `\Recent` flag is of limited use; it’s session-specific. Often clients ignore it and track seen/unseen themselves.
  - If you try to fetch an empty sequence set or invalid UIDs, the server might return an error or empty result; handle both cases.
  - Be mindful of line length limits; extremely long headers or lines could be an issue, but `go-message` should handle most RFC-compliant messages.

- **Logout and cleanup:** Always `Logout` when done. `c.Logout()` sends the `LOGOUT` command (transition to Logout state). The server will typically say bye and close. In code, you might use `defer c.Logout()` after connecting and logging in ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=%2F%2F%20Don%27t%20forget%20to%20logout,Logout)). This ensures the session ends cleanly and resources are freed on the server.

- **Testing:** You can test against real servers or an in-memory server. The `emersion/go-imap/server` package combined with `backend/memory` (as shown in the `go-imap` README ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=func%20main%28%29%20,New))) can be used to simulate an IMAP server locally (for unit tests). This is handy to test your client logic without needing external connectivity.

- **Updates channel usage:** If you set `c.Updates`, remember the warning: don't block it. The internal `go-imap` goroutine will try to send updates there; if you never read from it and it's unbuffered, you can deadlock the entire client ([client package - gopkg.in/emersion/go-imap.v1/client - Go Packages](https://pkg.go.dev/gopkg.in/emersion/go-imap.v1/client#:~:text=type%20Client%20struct%20)). Use a buffered channel or quickly consume in a separate goroutine. If you don't care about real-time updates, you can leave `c.Updates` nil and just rely on explicit commands to get state (then the library will just drop unsolicited responses or update internal state silently).

- **Mailbox state changes:** If the selected mailbox is deleted by another client, your next command might fail with a "Mailbox not found" or so. Be prepared for such errors. Similarly, if the mailbox is closed (e.g., server unselects you if you did a `DELETE` on it), you might need to select a different one.

In summary, `emersion/go-imap` provides a powerful low-level API to IMAP. It may require careful handling of goroutines and channels, but it gives you full control. Meanwhile, `emersion/go-message` greatly simplifies handling the MIME format of emails, which can otherwise be tedious to parse correctly. By combining these, a Go developer can build an email client or mail processing service that handles everything from mailbox management, message retrieval, flagging, to parsing message content and attachments.

With the patterns and examples provided above, you should be well-equipped to implement a robust IMAP client in Go that leverages these libraries to their fullest extent. Happy coding!

**Sources:**

- go-imap v1 README and docs for usage examples ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=%2F%2F%20Login%20if%20err%20%3A%3D,Logged%20in)) ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=%2F%2F%20Select%20INBOX%20mbox%2C%20err,mbox.Flags)) ([GitHub - emersion/go-imap at v1](https://github.com/emersion/go-imap/tree/v1#:~:text=seqset%20%3A%3D%20new%28imap))  
- go-imap client documentation and source for command methods and concurrency notes ([client package - github.com/emersion/go-imap/client - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap/client#:~:text=It%20is%20not%20safe%20to,commands%20on%20the%20same%20connection)) ([client package - gopkg.in/emersion/go-imap.v1/client - Go Packages](https://pkg.go.dev/gopkg.in/emersion/go-imap.v1/client#:~:text=type%20Client%20struct%20))  
- IMAP RFC 3501 for understanding flags and states ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=SeenFlag%20%20%20%20,)) ([imap package - github.com/emersion/go-imap - Go Packages](https://pkg.go.dev/github.com/emersion/go-imap#:~:text=%2F%2F%20In%20the%20authenticated%20state%2C,before%20commands%20that%20affect%20messages))  
- Stack Overflow answer on fetching full message with BODY.PEEK ([How to read an email's message body using the *imap.Message (emersion/go-imap) - Stack Overflow](https://stackoverflow.com/questions/70540892/how-to-read-an-emails-message-body-using-the-imap-message-emersion-go-imap#:~:text=Check%20first%20if%20emersion%2Fgo,306%20would%20apply))  
- go-message/mail package documentation and examples for MIME parsing ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=%2F%2F%20Create%20a%20new%20mail,log.Fatal%28err%29)) ([mail package - github.com/emersion/go-message/mail - Go Packages](https://pkg.go.dev/github.com/emersion/go-message/mail#:~:text=switch%20h%20%3A%3D%20p.Header.%28type%29%20,v%5Cn%22%2C%20filename%29%20%7D))  
- Maintainer comments on selective body fetching and IMAP usage ([How to get the message content? · Issue #72 · emersion/go-imap · GitHub](https://github.com/emersion/go-imap/issues/72#:~:text=would%20get%20the%20full%20MIME,you%20also%20download%20all%20attachments)) ([How to get the message content? · Issue #72 · emersion/go-imap · GitHub](https://github.com/emersion/go-imap/issues/72#:~:text=If%20you%20only%20want%20the,1.3.2))

