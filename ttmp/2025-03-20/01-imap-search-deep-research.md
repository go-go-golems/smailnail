I will provide a detailed technical breakdown of IMAP SEARCH functionality as defined in the IMAP4rev1 protocol (RFC 3501). This will include an in-depth explanation of how SEARCH operates at the protocol level, supported search keys, logical combinations, and performance considerations. 

Additionally, I will include a comprehensive set of IMAP SEARCH examples ranging from basic queries (e.g., searching by subject, sender, or date) to advanced multi-condition searches using flags, size filters, and complex logical expressions.

I'll let you know once the research is complete.

# IMAP SEARCH in IMAP4rev1 (RFC 3501)

## Overview of the IMAP SEARCH Command

**IMAP SEARCH** is a powerful command in the IMAP4rev1 protocol (defined in RFC 3501) that allows clients to search for messages on the server based on various criteria. The search is performed **server-side**, meaning the IMAP server examines message metadata and content to find matches, and returns the results to the client ([Using Squatter for Faster IMAP SEARCH — Cyrus IMAP 2.5.17 documentation](https://www.cyrusimap.org/2.5/imap/admin/sop/squatter.html#:~:text=IMAP%20SEARCH%2C%20as%20described%20in,the%20response%20to%20the%20client)). This avoids downloading all messages to the client for local search and can be efficient, especially for large mailboxes or limited-bandwidth clients. The SEARCH command does not alter any message state (it does not mark messages as seen or modify flags), it simply queries the mailbox.

**Usage context:** A client can only issue a SEARCH after selecting a mailbox (via `SELECT` or `EXAMINE`), since the search is performed within the currently selected mailbox. The basic syntax is: 

```
C: A001 SEARCH [CHARSET <charset>] <search criteria...>
S: * SEARCH <list of matching message numbers>
S: A001 OK SEARCH completed
```

- The command is prefixed with an IMAP *tag* (`A001` in this example) for tracking the command/response.
- An optional `CHARSET` can be specified if the search involves international characters. If omitted, the default charset is US-ASCII. The server will return a `NO [BADCHARSET]` error if it does not support the requested charset ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=in%20a%20,s%20supported%20by%20the%20server)).
- The **search criteria** consist of one or more *search keys* (detailed below) that specify conditions messages must meet ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=The%20SEARCH%20command%20searches%20the,that%20were%20placed%20in%20the)).
- The server’s response includes an untagged `* SEARCH` line listing the **message sequence numbers** of all matching messages, followed by a tagged OK/NO completion status. For example, `* SEARCH 2 84 882` means messages 2, 84, and 882 match ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=%23%20%20)).
- If no messages match, the `* SEARCH` line is returned with no numbers following it (just `* SEARCH` on its own) ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=Track%20,SEARCH%2043)). An OK completion still indicates the search finished (as opposed to a NO which indicates an error in the query).

**UID SEARCH:** IMAP also allows a UID-specific variant of the search. If the client issues `UID SEARCH <criteria>`, the server will return matching **unique identifiers (UIDs)** instead of sequence numbers ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=occurs%20as%20a%20result%20of,FLAGS%20Response%20Contents)). The search criteria format is the same; the only difference is in the interpretation of numbers and the returned values. For example, `UID SEARCH FROM "alice@example.com"` might return `* SEARCH 1001 1010` meaning UIDs 1001 and 1010 matched (instead of message sequence numbers) ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=occurs%20as%20a%20result%20of,FLAGS%20Response%20Contents)). This is useful when a client wants to work with stable message identifiers across sessions. Internally, there is also a search key named **UID** (described later) that can be used within a normal SEARCH to specify a UID or range to restrict the search ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=UID%20%27message%20set%27)), but using the `UID SEARCH` command prefix is more common when UIDs are needed.

## SEARCH Command Mechanics and Behavior

When the client sends a SEARCH command, the IMAP server evaluates the criteria **against all messages in the selected mailbox by default** (or against a subset if message sequence numbers or UID ranges are specified as part of the criteria). Multiple criteria are **combined by default with a logical AND** – a message must satisfy *all* the given conditions to be returned ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=Crispin%20Standards%20Track%20,be%20a%20parenthesized%20list%20of)). For example, the command:

```
C: A282 SEARCH DELETED FROM "Smith" SINCE 1-Feb-1994
```

asks for messages that are (Deleted **AND** from “Smith” **AND** dated on/after Feb 1 1994). This would return only messages meeting all three conditions ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=Crispin%20Standards%20Track%20,be%20a%20parenthesized%20list%20of)). The order of criteria in the command does not change the logical result (they are an unordered AND combination), though servers may internally optimize the evaluation order for efficiency.

The SEARCH operation is performed on the server’s current view of the mailbox. IMAP guarantees that no messages will be expunged (deleted permanently) mid-search, to avoid disrupting the sequence numbers while the search is in progress ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=a%20subsequent%20command,for%20the%20completion%20result%20response)). In practice, the server locks out expunge notifications during a SEARCH, FETCH, or STORE operation, ensuring a consistent snapshot of the mailbox is searched. Once the search completes, normal notifications (e.g., new mail or expunges) may resume.

**Case-insensitivity and partial matches:** String-based search keys match substrings **case-insensitively**. In all search keys that use text strings, the match is true if the given string is found as a substring of the target field, ignoring ASCII letter case ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=In%20all%20search%20keys%20that,Syntax%20section%20for%20the%20precise)). For example, searching for SUBJECT "hello" will match a message with subject "Hello World" (case-insensitive) and searching TEXT "profit" would match "Profit" or "nonprofit" in the text. The matching is typically not full-regex or prefix-based (unless the server supports extension filters); it’s a simple substring containment test as defined by the IMAP spec.

**Charset considerations:** If search strings include non-ASCII characters (e.g., accented letters or other scripts), the client can specify a CHARSET. IMAP servers are required to support at least US-ASCII and may support other charsets like UTF-8 ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=strings%20in%20%5BRFC,This%20response%20SHOULD%20contain%20the)). The server will decode message data (e.g., MIME encoded words in headers or base64 in bodies) to perform the match in the specified charset ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=OPTIONAL%20,key%20if%20the%20string%20is)). If an unsupported charset is requested, the server responds with `NO` rather than attempting an incorrect search ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=in%20a%20,s%20supported%20by%20the%20server)).

**Result and post-processing:** The untagged SEARCH results provide the list of matching message numbers (or UIDs). It is then up to the client what to do with them – for example, the client might subsequently fetch those messages, mark them, or display summaries. The SEARCH command itself does not convey any data except the matched ids. Notably, performing a search does **not** set the `\Seen` flag on messages, even if it searches the body – the server is explicitly allowed to scan message content without marking it as read.

## Supported Search Keys and Criteria

IMAP4rev1 defines a rich set of search keys to filter messages by flags, headers, content, size, and date. Search criteria can be combined in various ways to form complex queries. Below is a breakdown of **all the supported search keys** defined in RFC 3501, organized by category:

### General and Special Keys

- **<message sequence number set>** – You can directly specify message numbers or ranges as a criteria to limit the scope of the search to certain messages. For example, `SEARCH 1:100 SUBJECT "Report"` would only consider messages 1 through 100 when matching the subject. By default, if no explicit sequence or UID range is given, the search covers all messages in the mailbox. (This is more of a part of the command syntax than a "key name": any numbers at the start of the criteria are taken as the message set to search within ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=syntactic%20definitions%20of%20the%20arguments,RFC%203501%20IMAPv4%20March%202003)).)

- **ALL** – Matches **all messages** in the mailbox ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=corresponding%20to%20the%20specified%20message,RFC%203501%20IMAPv4%20March%202003)). This is the default if you don’t provide any search criteria (i.e., `SEARCH` by itself is equivalent to `SEARCH ALL`). In practice, clients usually specify criteria explicitly, but `ALL` can be used to force inclusion of all messages especially in compound queries.

- **UID** *<uid-set>* – Matches messages with unique identifiers in the specified set ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=UID%20%27message%20set%27)). The `<uid-set>` can be a single UID or a range (e.g., `1000:1500`). This criterion limits the search to particular UIDs. It’s similar to giving a sequence range, but uses UIDs. For example, `SEARCH UID 10:20 UNSEEN` finds unseen messages among UIDs 10 through 20. (Remember, using the `UID SEARCH` prefix is an alternative way to search by UIDs and get UIDs in the result; the `UID` key inside a normal SEARCH just restricts the scope by UID while still returning sequence numbers in a normal SEARCH response.)

### Flag-Based Keys (Message Status)

These keys filter messages based on their flag status (read/unread, flagged, deleted, etc.). Many of them come in pairs – one to find messages with a flag, and the other (prefixed with `UN`-) to find messages *without* that flag:

- **ANSWERED** / **UNANSWERED** – Messages with or without the `\Answered` flag set, respectively ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=identifiers%20corresponding%20to%20the%20specified,UNDRAFT%20Messages)). An ANSWERED message is one that has been replied to (this flag is usually set by the client when the user replies to a message).

- **SEEN** / **UNSEEN** – Messages that have or have not been marked as `\Seen` (read). **SEEN** finds messages the user has already read; **UNSEEN** finds all unread messages ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=set,Crispin%20Standards)) (this is equivalent to what many clients call the "unread" search).

- **FLAGGED** / **UNFLAGGED** – Messages with or without the `\Flagged` flag ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=ranges%20are%20permitted,that%20do%20not%20have%20the)). Typically `\Flagged` means the message is marked as important or starred.

- **DELETED** / **UNDELETED** – Messages with or without the `\Deleted` flag ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=identifiers%20corresponding%20to%20the%20specified,that%20do%20not%20have%20the)). `\Deleted` usually means a message has been marked for deletion (but not yet expunged).

- **DRAFT** / **UNDRAFT** – Messages with or without the `\Draft` flag ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=ranges%20are%20permitted,that%20do%20not%20have%20the)). This is often used to identify draft messages saved in a Drafts folder.

- **RECENT** / **OLD** – Messages that have the `\Recent` flag set, or not set, respectively ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=within%20the%20specified%20date,disregarding%20time%20and)) ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=Crispin%20Standards%20Track%20,RECENT)). `\Recent` is a special flag that the server sets for messages that have arrived since the last time the mailbox was opened. **RECENT** thus finds messages that are newly delivered and not yet accessed. **OLD** is essentially the opposite of RECENT (messages that are not recent) and is functionally equivalent to `NOT RECENT` ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=Crispin%20Standards%20Track%20,is)). Note that `\Recent` is cleared when a mailbox is opened (for the second and subsequent times), so it’s ephemeral; clients often use it only to see what’s "new since last check". In searches, **NEW** is a related key (see next).

- **NEW** – Messages that are both `\Recent` **and not** `\Seen` ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=specified%20keyword%20flag%20set,2003%20NOT%20%20Messages%20that)). In other words, **NEW** = RECENT + UNSEEN. This finds messages that arrived recently and have not been read yet (sometimes considered "new unread" messages). It is functionally equivalent to the compound search `(RECENT UNSEEN)` ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=specified%20keyword%20flag%20set,2003%20NOT%20%20Messages%20that)).

- **KEYWORD** *<flag>* / **UNKEYWORD** *<flag>* – Messages that have (or do not have) a specific **keyword flag** set ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=specified%20field,KEYWORD%20%20Messages%20with%20the)) ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=set,Crispin%20Standards)). IMAP allows client-defined keywords (flags that are not one of the standard ones like \Seen, etc.). For example, if a mailbox supports a custom flag "ProjectX", you could search `KEYWORD ProjectX` to find all messages tagged with that keyword. Conversely, `UNKEYWORD ProjectX` finds messages that do not have that tag.

*(Flags in IMAP are stored per message, so these searches are usually efficient since servers track flags in metadata. Searching by flag typically does not require scanning the message content at all.)*

### Header/Envelope Field Criteria

These keys search for a substring within specific header fields of the message (the “envelope” data):

- **FROM** *"string"* – Messages that contain the specified string in the **From:** header field ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=FROM%20%27string%27)). This is typically used to find messages by the sender’s name or email address. (Example: `FROM "alice@example.com"` or `FROM "Alice"`)

- **TO** *"string"* – Messages with the specified string in the **To:** header field ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=TO%20%27string%27)) (the recipient list). Useful to find emails addressed to a certain person or containing a certain recipient.

- **CC** *"string"* – Messages with the specified string in the **Cc:** (carbon copy) header field ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=CC%20%27string%27)).

- **BCC** *"string"* – Messages with the specified string in the **Bcc:** (blind carbon copy) field of the envelope ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=BCC%20%27string%27)). (Note: Bcc headers are not always stored in the delivered message headers, but IMAP servers often have Bcc information in the envelope structure if the server is the final destination or if the Bcc was included in an SMTP transaction. Not all messages will have a visible Bcc field.)

- **SUBJECT** *"string"* – Messages with the specified string in the **Subject:** header ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=March%202003%20SUBJECT%20%20Messages,contain%20the%20specified%20string%20in)). This searches the subject line for the substring provided (again, case-insensitive). Example: `SUBJECT "meeting"` would match "Meeting Agenda" or "team meeting notes".

- **HEADER** *<field-name> "string"* – A generalized header search. This lets you specify any header field name and a value substring to search for ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=,KEYWORD%20%20Messages%20with%20the)). For example, `HEADER "List-Id" "project-team"` could find messages on a mailing list (looking for "project-team" in the `List-Id:` header), or `HEADER "X-Priority" "1"` could find high-priority messages. If the **"string"** given is empty (`""`), this key will match all messages that have the header field at all, regardless of its value ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=envelope%20structure%27s%20FROM%20field,KEYWORD%20%20Messages%20with%20the)). (For instance, `HEADER "X-Spam-Flag" ""` would match messages that have an `X-Spam-Flag` header, no matter what it’s set to.)

*All these header searches look at the textual content of the header field (what comes after the colon in the raw email header). They do **not** interpret or parse email addresses – they just do substring match on the raw header text. So, searching FROM "Bob" would match an email from "Alice Bobson <alice@example.com>" as well as one from "bob@example.com", since "Bob" appears in the header line in both cases.* 

### Body/Text Content Criteria

These keys search within the content of the message (body or entire message text):

- **BODY** *"string"* – Searches for the specified substring in the **body** of the message ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=BODY%20%27string%27)). This means it searches the text of the message **excluding headers**. It typically covers the text parts of the email message (for MIME messages, servers may limit to searching text/plain or text/html parts and might not search within attachments that are binary or non-text ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=one%20or%20more%20search%20keys,The))).

- **TEXT** *"string"* – Searches for the specified substring in the entire text of the message ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=March%202003%20SUBJECT%20%20Messages,contain%20the%20specified%20string%20in)), including both the headers and the body. This is a broad search: it will match if the string appears *anywhere* in the message. Essentially, `TEXT` = `HEADER <any field>` OR `BODY`, conceptually. For example, `TEXT "invoice"` could match an email where "Invoice" is in the subject, or in the message body.

*(BODY and TEXT are often the most expensive searches performance-wise, because they require scanning the message content. To optimize, some IMAP servers may exclude attachments or non-textual parts from these searches, as allowed by the spec ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=one%20or%20more%20search%20keys,The)), or they might utilize full-text indexes to speed up matching.)*

### Date/Time Criteria

IMAP allows searching by dates. Importantly, IMAP has two sets of date-based keys: one set refers to the message’s **Internal Date** (the date the message was received by the mailbox), and another set refers to the **Date:** header (the date the message was sent, according to the sender).

All date-based searches in IMAP compare **only the date portion** (year-month-day) of the timestamp, ignoring the time of day and timezone differences ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=NEW,2822)). So a message’s internal date of “2025-03-20 15:30 UTC” is considered to have the date March 20, 2025 for searching purposes.

- **BEFORE** *<date>* – Messages whose internal date is **earlier** than the given date ([IMAP Search Commands](https://www.marshallsoft.com/ImapSearch.htm#:~:text=BEFORE%20%27date%27)). For example, `BEFORE 1-Feb-2021` finds messages with an internal timestamp before February 1, 2021 (i.e., 31-Jan-2021 and earlier).  
- **ON** *<date>* – Messages whose internal date is on **the same day** as the given date ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=do%20not%20match%20the%20specified,RECENT)). (E.g., `ON 10-Oct-2022` finds messages dated October 10, 2022.)  
- **SINCE** *<date>* – Messages whose internal date is **on or after** the given date (i.e., *since* that day) ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=Date%3A%20header%20,SINCE%20%20Messages%20whose)). (E.g., `SINCE 1-Dec-2024` finds messages from December 1, 2024 and later.)

*(“Internal date” is typically when the message was added to the mailbox, which usually corresponds to when it was delivered by the server. This can differ from the sent date in the header if, for example, the email was sent earlier but only delivered later, or if an old message was copied/moved into the mailbox.)*

- **SENTBEFORE** *<date>* – Messages whose **Date:** header is earlier than the given date ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=set,disregarding%20time%20and)).  
- **SENTON** *<date>* – Messages whose **Date:** header is on the given date ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=set,disregarding%20time%20and)).  
- **SENTSINCE** *<date>* – Messages whose **Date:** header is on or after the given date ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=set,disregarding%20time%20and)).

These three are analogous to BEFORE/ON/SINCE but apply to the date a message was sent (according to the sender’s email headers) rather than the time it arrived on the server. For instance, `SENTSINCE 1-Jan-2023` finds emails that were sent on or after January 1, 2023 (as per the Date header in the message).

*Date format:* Dates in IMAP search criteria must be given in day-month-year format with the month abbreviated to the first three letters and year as four digits, e.g., `20-Mar-2025`. (This format is the same as used in IMAP FETCH envelope dates.)

### Size Criteria

- **LARGER** *<n>* – Messages with a size (RFC822 size of the entire message) **larger** than *n* bytes ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=specified%20keyword%20flag%20set,RECENT%20UNSEEN)). For example, `LARGER 100000` finds messages bigger than 100,000 bytes (~100 KB). This is useful to locate big emails or attachments.

- **SMALLER** *<n>* – Messages with a size **smaller** than *n* bytes ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=SENTSINCE%20%20Messages%20whose%20%5BRFC,2822%5D%20size%20smaller%20than%20the)). For example, `SMALLER 5000` finds messages under 5 KB.

*(Size refers to the message’s raw size in bytes. The IMAP server usually knows this value for each message without reading the whole message, so size searches are typically efficient. Note that 1000 bytes is about 0.98 KB, not 1 KB exactly, but for practical purposes admins treat these as approximate KB thresholds.)*

### Logical Operators and Grouping

By combining multiple search keys, you implicitly perform a logical **AND** (intersection) of those criteria, as noted earlier. IMAP also provides explicit operators to perform **OR** and **NOT** operations, as well as parentheses for grouping complex expressions.

- **NOT** *<search-key>* – Negates a search criterion, matching messages that **do not** meet the given condition ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=not%20the%20,NOT)). For example, `NOT SEEN` finds messages that are not seen (equivalent to UNSEEN), and `NOT FROM "Bob"` finds messages whose From field does *not* include “Bob”. You can prepend NOT to a parenthesized group as well, e.g., `NOT (FROM "Alice" TO "Alice")` would mean messages neither from nor to Alice.

- **OR** *<search-key1>* *<search-key2>* – Returns messages that match **either** of the two given search keys (the logical OR of two conditions) ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=NEW,disregarding%20time%20and)). For example, `OR FROM "Alice" FROM "Bob"` finds messages that are from Alice **or** from Bob. The OR operator takes exactly two search conditions as arguments. To combine more than two alternatives, OR can be nested or chained with parentheses – for instance, to search for messages from Alice, Bob, or Carol, one could do: `OR FROM "Alice" (OR FROM "Bob" FROM "Carol")`. This is effectively (Alice) OR (Bob or Carol). Parenthesized sub-criteria are treated as a single search key for the purposes of NOT or OR.

- **( ... )** (Parenthesized criteria list) – Used to group one or more criteria into a single compound criterion. Parentheses are especially useful when using OR or NOT to ensure the logical grouping is as intended ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=1,Server)). For example, consider searching for messages that are either from Alice **or** (from Bob **and** are flagged). There’s no single operator for AND (since AND is the default), so you might express this as: `OR FROM "Alice" (FROM "Bob" FLAGGED)`. The parentheses group the `FROM "Bob" FLAGGED` part together as one condition (which itself is an AND of Bob and Flagged) that OR will evaluate against.

Putting it all together, complex queries can be constructed. For instance: 

```
C: A300 SEARCH NOT DELETED SINCE 1-Feb-2021 (FROM "boss@example.com" SUBJECT "Urgent")
```

This would search for messages (since Feb 1, 2021) that are **not** deleted, and (are from boss@example.com **and** have "Urgent" in the subject). Such a query uses both NOT and parentheses with an implicit AND inside the parentheses. The IMAP server will evaluate this according to the defined precedence (parentheses group first, NOT applies to everything after it, and OR if present would only apply to its two specified sub-keys).

## Performance Considerations and Optimizations

The flexibility of IMAP SEARCH comes with performance considerations. A simple search like checking flags or dates can be very fast, whereas complex text searches can be slow on large mailboxes if not optimized:

- **Server-side scanning:** By default, when an IMAP server receives a SEARCH command, it may have to **scan through all messages** in the mailbox to test the criteria. This can be an intensive, time-consuming operation for a mailbox with thousands of emails ([Using Squatter for Faster IMAP SEARCH — Cyrus IMAP 2.5.17 documentation](https://www.cyrusimap.org/2.5/imap/admin/sop/squatter.html#:~:text=IMAP%20SEARCH%2C%20as%20described%20in,the%20response%20to%20the%20client)). For example, a `TEXT "keyword"` search might require reading every message’s content to look for the keyword.

- **Optimizing with indexes:** IMAP servers often employ optimizations to speed up search. Many servers maintain internal indexes or caches of message information:
  - Most servers index flags and internal dates (for example, they can instantly know which messages are \Seen or \Deleted without scanning each message file). So searches on flags or dates can be answered by looking up metadata rather than reading each message.
  - Some servers also build **full-text search indexes** to accelerate BODY/TEXT searches. For instance, the Cyrus IMAP server provides a tool called “squatter” that builds a cache of message content and metadata to allow fast text searches ([Using Squatter for Faster IMAP SEARCH — Cyrus IMAP 2.5.17 documentation](https://www.cyrusimap.org/2.5/imap/admin/sop/squatter.html#:~:text=To%20significantly%20speed%20up%20the,generate%20and%20maintain%20these%20caches)). Dovecot and other modern IMAP servers have similar full-text search (FTS) indexing options. With such an index, a search for a keyword can be answered in near constant time using the index, rather than linear time scanning through messages.
  - If indexes are not built, administrators sometimes impose limits or advise clients to avoid extremely broad content searches on very large mailboxes due to the load it can create.

- **MIME and text parts:** As noted earlier, the IMAP spec allows servers to **exclude non-textual parts** of messages from TEXT/BODY searches ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=one%20or%20more%20search%20keys,The)). This is an optimization – binary attachments (images, PDFs, etc.) need not be searched for text matches, since they likely won’t contain the query string (and if they did, it’d be in some encoded form). By ignoring these parts, the server saves processing time. This means a BODY/TEXT search might not catch a word that only appears in an attachment or an image, but it’s generally an acceptable trade-off for performance.

- **Incremental searches and server support:** Some IMAP extensions (beyond RFC 3501) exist to make searching more efficient or flexible. For example, there are extensions for server-side sorting and searching (e.g., ESEARCH – Extended Search, CONDSTORE and QRESYNC for caching state, etc.). However, those are beyond the base IMAP4rev1 spec. In practice, if a client expects to do a lot of repeated searches, it might benefit from those extensions or from caching results client-side. Nonetheless, for most uses, the base SEARCH with good server indexing is sufficient for “nearly all users and situations” ([email - Is the RFC 3501 SEARCH command used in production in major ways? - Stack Overflow](https://stackoverflow.com/questions/57288909/is-the-rfc-3501-search-command-used-in-production-in-major-ways#:~:text=In%20any%20case%2C%20I%27d%20say,nearly%20all%20users%20and%20situations)).

- **Client strategies:** Clients that want to be efficient often use selective criteria to narrow searches. For example, a mail client might first issue a cheap search for `UNSEEN` to get all unseen message IDs, then fetch those. Or it might use date ranges to limit a text search to recent messages rather than searching 10 years of mail by default. Another strategy is to perform server search only when needed (like a user-initiated search in a mail app) and rely on local caches for frequent filtering (like showing the unread count, etc., which might be tracked without a search). Mobile clients and webmail clients heavily rely on server SEARCH commands since they typically do not store all mail locally.

In summary, the SEARCH command is extremely useful and is widely implemented (server support is mandatory in IMAP4rev1). Many IMAP clients do use it under the hood (for example, to find all unseen messages, or to locate messages by certain criteria without downloading everything) ([email - Is the RFC 3501 SEARCH command used in production in major ways? - Stack Overflow](https://stackoverflow.com/questions/57288909/is-the-rfc-3501-search-command-used-in-production-in-major-ways#:~:text=statistics%2C%20I%20can%20see%20that,commands)). Proper use of the available keys and understanding the server’s capabilities can yield powerful results with good performance.

## Examples of IMAP SEARCH Queries

Below is a set of examples demonstrating various IMAP SEARCH uses, from basic queries to complex combinations. In these examples, `C:` denotes the **client** sending a command, and `S:` denotes the **server** response. (The tags `A001`, `A002`, etc., are arbitrary command identifiers set by the client.)

### Basic Search Examples

**1. Search by sender (FROM)** – Find all messages from a specific sender. For example, to find emails from **alice@example.com**: 

```
C: A001 SEARCH FROM "alice@example.com"
S: * SEARCH 3 5 8 21
S: A001 OK SEARCH completed
```

The server returns the list of message sequence numbers (e.g., messages 3, 5, 8, 21) that have “alice@example.com” in the From header. The match is substring-based, so this would also match if the From header contained `<alice@example.com>` or perhaps `"Alice Johnson" <alice@example.com>`.

**2. Search by recipient (TO)** – Similar to above, but for the **To** field. For instance, to find messages sent *to* bob@example.com:

```
C: A002 SEARCH TO "bob@example.com"
S: * SEARCH 7 15 16 22 23
S: A002 OK SEARCH completed
```

This returns all messages where "bob@example.com" appears in the To header (he could be one of multiple recipients).

**3. Search by subject** – To find messages with a certain word in the Subject, use the **SUBJECT** key. For example:

```
C: A003 SEARCH SUBJECT "Invoice"
S: * SEARCH 4 10 11
S: A003 OK SEARCH completed
```

This might return messages 4, 10, 11 which have “Invoice” in the subject line. It would match subjects like "January Invoice Attached" or "Re: Invoice Questions" (case-insensitive). It would not match if "invoice" appears only in the body but not in the subject.

**4. Search by keywords in the entire message** – Use **TEXT** for a broad search. For example, to find any message that contains the word "database" anywhere (in headers or body):

```
C: A004 SEARCH TEXT "database"
S: * SEARCH 2 9 14 18 30
S: A004 OK SEARCH completed
```

This might return messages that mention "database" either in the body text or perhaps in a header (maybe part of a subject or an email address, etc.). If you wanted to restrict to just the body, you would use `BODY` instead of `TEXT`.

**5. Search by date** – Suppose we want emails on or after March 1, 2025. We can use **SINCE** with a date:

```
C: A005 SEARCH SINCE 1-Mar-2025
S: * SEARCH 20 21 22 23 24 25
S: A005 OK SEARCH completed
```

This returns all messages whose internal date is March 1, 2025 or later (in this hypothetical, messages 20-25). If we wanted messages *before* a date, we’d use **BEFORE**, and **ON** for exactly on a date. For example, `SEARCH ON 20-Mar-2025` would find messages dated March 20, 2025. Remember, the date format is day-Mon-year.

**6. Search by age relative to sent date** – If we want to find older emails by the date they were sent, we can use **SENTBEFORE/SENTSINCE**. For example, to find messages sent before 2020:

```
C: A006 SEARCH SENTBEFORE 1-Jan-2020
S: * SEARCH 1 2 3 4 5 6 7 8 9 10 11
S: A006 OK SEARCH completed
```

This might list a bunch of early messages (1–11) that have Date headers before Jan 2020. By contrast, `SENTSINCE 1-Jan-2020` would find those sent in 2020 or later.

**7. Search by message size** – To find large messages (perhaps to clean up mailbox space), use **LARGER**. For example:

```
C: A007 SEARCH LARGER 1000000
S: * SEARCH 30 31 32
S: A007 OK SEARCH completed
```

This finds messages larger than 1,000,000 bytes (~1 MB). Similarly, `SMALLER 10000` could find messages under 10 KB, etc.

**8. Search by flags (unread or read)** – To get all unread emails, you can search for **UNSEEN**:

```
C: A008 SEARCH UNSEEN
S: * SEARCH 22 23 25 30
S: A008 OK SEARCH completed
```

This returns the messages currently marked as unseen (in this case, e.g., messages 22, 23, 25, 30 are unread). Conversely, `SEARCH SEEN` would list all messages that have been read. You can similarly use `ANSWERED` / `UNANSWERED`, `FLAGGED` / `UNFLAGGED`, etc., in the same way. For instance, `SEARCH FLAGGED` might return all “starred” or important messages.

**9. Search for new messages** – To find messages that are new (recent and unseen), use **NEW**:

```
C: A009 SEARCH NEW
S: * SEARCH 25 30
S: A009 OK SEARCH completed
```

This would return messages that arrived recently and haven’t been marked as seen yet (perhaps 25 and 30 in this example). This is essentially a shortcut for `SEARCH (RECENT UNSEEN)`.

**10. Search using a specific header field** – Suppose we want to find messages that have a specific header or value. Using the **HEADER** key allows this. For example:

```
C: A010 SEARCH HEADER "X-Priority" "1"
S: * SEARCH 5 8
S: A010 OK SEARCH completed
```

This would find messages that have `X-Priority: 1` in their header (often denoting high priority). Another example could be `SEARCH HEADER "Mailing-List" "list@example.com"` to find messages belonging to a certain mailing list, or `SEARCH HEADER "Received" "smtp.example.com"` to find messages that passed through a certain server (though that’s a rather advanced usage). If we wanted to find any message that has an `X-Spam-Flag` header (regardless of its value), we could do: `SEARCH HEADER "X-Spam-Flag" ""` (empty string) which matches presence of that header.

### Advanced and Compound Search Examples

**11. Combining multiple conditions (AND logic)** – As noted, you can simply list multiple criteria and the server will interpret it as an AND. For example, to find messages from alice@example.com **that are unread**:

```
C: A011 SEARCH FROM "alice@example.com" UNSEEN
S: * SEARCH 8 21
S: A011 OK SEARCH completed
```

This returns only the messages that satisfy both conditions (in this case, messages 8 and 21 are from Alice and currently unread). Similarly, you could add more: e.g., `FROM "alice@example.com" UNSEEN SINCE 1-Jan-2025` to further restrict to recent messages from Alice that are unread. Because AND is the default, you just list them separated by space.

**12. Using NOT to exclude** – Suppose you want all messages with "Project X" in the subject **except** those from bob@example.com. You could do:

```
C: A012 SEARCH SUBJECT "Project X" NOT FROM "bob@example.com"
S: * SEARCH 11 14 18
S: A012 OK SEARCH completed
```

Here the search finds messages with "Project X" in the subject, and then the `NOT FROM "bob@example.com"` excludes any of those that were from Bob. The result (messages 11, 14, 18) are Project X related messages that are not sent by Bob. (This works because NOT applies to the single criterion immediately following it. In our case, it's equivalent to saying SUBJECT contains "Project X" AND (NOT (FROM contains "bob@example.com")).)

**13. Using OR for alternatives** – If you want to match one of several conditions, use OR. For example, find messages that are **from Alice or from Bob**:

```
C: A013 SEARCH OR FROM "alice@example.com" FROM "bob@example.com"
S: * SEARCH 3 5 8 21 22 23
S: A013 OK SEARCH completed
```

This returns messages that satisfy either criterion. In this hypothetical result, messages 3,5,8,21 might be from Alice and 22,23 from Bob (or vice versa). The key thing is any message that matches at least one of the two conditions is returned. Another example: `SEARCH OR SUBJECT "Urgent" SUBJECT "Important"` would find messages whose subject contains either "Urgent" or "Important".

**14. Grouping with parentheses** – OR only takes two arguments, but you can chain multiple ORs by grouping. For example, find messages that are from Alice **or** Bob, **and** that have "Report" in the subject. We need to ensure "Report" applies to either case, so we group the OR conditions:

```
C: A014 SEARCH (OR FROM "alice@example.com" FROM "bob@example.com") SUBJECT "Report"
S: * SEARCH 5 8
S: A014 OK SEARCH completed
```

The parentheses around the OR clause make it a single combined condition, which is then ANDed with `SUBJECT "Report"`. So this finds messages whose subject contains "Report" and (the message is from Alice or Bob). The result could be, for example, message 5 (a report from Alice) and message 8 (a report from Bob). Without the parentheses, if you wrote `OR FROM "alice" FROM "bob" SUBJECT "Report"`, the query would be interpreted as `OR (FROM "alice") (FROM "bob" AND SUBJECT "Report")`, which is a different logic (that would return anything from Alice, or those that are from Bob with "Report" in subject).

**15. Complex example** – Now a more complex scenario: find all messages that (are unread **and** from the **example.com** domain) **or** (are flagged **and** larger than 500KB). This is a compound logic: (UNSEEN AND FROM "*@example.com") OR (FLAGGED AND LARGER 512000). In IMAP SEARCH syntax:

```
C: A015 SEARCH OR (UNSEEN FROM "example.com") (FLAGGED LARGER 512000)
S: * SEARCH 8 9 30 31
S: A015 OK SEARCH completed
```

In this made-up result, messages 8 and 9 might be unread notes from someone at example.com, and 30,31 might be big flagged messages. This demonstrates combining multiple criteria in each parenthesized group and then OR-ing the groups. It shows how powerful IMAP search can be, retrieving exactly the set of messages that meet a complex combination of properties.

**16. UID SEARCH example** – Lastly, an example using **UID SEARCH** to clarify its behavior. Let’s say we want to find messages with UIDs between 100 and 150 that have "demo" in the subject. We can use:

```
C: A100 UID SEARCH 100:150 SUBJECT "demo"
S: * SEARCH 142 150
S: A100 OK SEARCH completed
```

Here, because we used `UID SEARCH`, the numbers `142 150` in the untagged SEARCH response are UIDs (not sequence positions). It indicates that the messages with UID 142 and 150 match the criteria. If the client instead had done a normal `SEARCH 100:150 SUBJECT "demo"`, the numbers returned would be message sequence numbers in the range 100–150 that match. Using UID SEARCH ensures the identifiers refer to UIDs, which remain stable even if the mailbox is reopened or if other messages are added/removed.

---

These examples cover common use cases, but the SEARCH command’s full capability extends to any combination of the listed criteria. Clients can mix and match as needed: for instance, searching for unseen messages with certain keywords in specific date ranges, or messages to a list that haven’t been answered yet, and so on. By leveraging server-side search, IMAP clients can efficiently locate messages without transferring large amounts of data. The IMAP SEARCH command, as defined in RFC 3501, provides a comprehensive set of tools to perform these queries, making it a cornerstone of IMAP’s functionality for email management ([www.rfc-editor.org](https://www.rfc-editor.org/rfc/rfc3501.txt#:~:text=The%20SEARCH%20command%20searches%20the,that%20were%20placed%20in%20the)) ([email - Is the RFC 3501 SEARCH command used in production in major ways? - Stack Overflow](https://stackoverflow.com/questions/57288909/is-the-rfc-3501-search-command-used-in-production-in-major-ways#:~:text=reason%20to%20cache%20all%20messages,nearly%20all%20users%20and%20situations)).

