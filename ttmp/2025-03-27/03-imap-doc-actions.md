# IMAP Actions Implementation with go-imap/v2

This document provides the correct implementation details for IMAP actions using the go-imap/v2 library, focusing on the operations needed for the IMAP DSL: setting flags, moving and copying messages, deleting messages, and exporting message content.

## Table of Contents

1. [Understanding go-imap/v2 API Changes](#understanding-go-imapv2-api-changes)
2. [General Patterns for IMAP Commands](#general-patterns-for-imap-commands)
3. [Flag Manipulation (STORE)](#flag-manipulation-store)
4. [Message Copy (COPY)](#message-copy-copy)
5. [Message Move (MOVE)](#message-move-move)
6. [Message Deletion (STORE + EXPUNGE)](#message-deletion-store--expunge)
7. [Message Export (FETCH)](#message-export-fetch)
8. [Complete Action Implementation Example](#complete-action-implementation-example)

## Understanding go-imap/v2 API Changes

The go-imap/v2 API has significant changes compared to v1, particularly:

- Command methods return command objects that you must interact with to get results (e.g., using `.Wait()` or `.Collect()`)
- Different structs and type hierarchies for commands and data
- Different parameter requirements for most commands

Key differences that caused our errors:

1. **StoreOptions vs StoreFlags**: In v2, the `Store` command takes a `*imap.StoreFlags` struct and an optional `*imap.StoreOptions` struct, not a single options struct.
2. **NumSet vs SeqSet**: Methods like `Store`, `Copy`, and `Move` work with `imap.NumSet` type, which can be either a `imap.SeqSet` or `imap.UIDSet`.
3. **Command Return Values**: Commands return command objects with methods like `Wait()` that return multiple values - both the data and an error.
4. **FetchMessageBuffer Structure**: The structure for message data from fetch operations has changed, including how literal data is accessed.

## General Patterns for IMAP Commands

In go-imap/v2, most commands follow this pattern:

```go
// 1. Call the command which returns a command object
cmd := client.CommandName(args...)

// 2. Get results using Wait() or Collect()
result, err := cmd.Wait() // for single result commands
// OR
results, err := cmd.Collect() // for commands that stream multiple items
```

For commands that stream data (like Fetch), you can also use the Next/Close pattern:

```go
cmd := client.Fetch(...)
defer cmd.Close() // Important: always close commands to avoid leaking goroutines

for {
    msg := cmd.Next()
    if msg == nil {
        break
    }
    // Process msg
}
if err := cmd.Close(); err != nil {
    // Handle error
}
```

## Flag Manipulation (STORE)

To add or remove flags from messages, use the `Store` method with `StoreFlags` struct:

```go
// Create a sequence set from message UIDs
var seqSet imap.SeqSet
for _, msg := range messages {
    seqSet.AddNum(uint32(msg.UID))
}

// Add flags
storeFlags := &imap.StoreFlags{
    Op:     imap.StoreFlagsAdd,  // Use StoreFlagsAdd to add flags
    Silent: true,                // Silent mode doesn't return updated flags
    Flags:  []imap.Flag{imap.FlagSeen, imap.FlagFlagged},
}

// Execute STORE
_, err := client.Store(seqSet, storeFlags, nil).Wait()
if err != nil {
    return fmt.Errorf("failed to add flags: %w", err)
}

// Remove flags
storeFlags = &imap.StoreFlags{
    Op:     imap.StoreFlagsDel,  // Use StoreFlagsDel to remove flags
    Silent: true,
    Flags:  []imap.Flag{imap.FlagDraft},
}

_, err = client.Store(seqSet, storeFlags, nil).Wait()
if err != nil {
    return fmt.Errorf("failed to remove flags: %w", err)
}
```

**Important details**:
- The `Op` field determines if you're adding (`StoreFlagsAdd`), removing (`StoreFlagsDel`), or replacing (`StoreFlagsSet`) flags
- `Silent` set to `true` means the server won't send updated flags back (more efficient)
- Flags are predefined in the `imap` package (e.g., `imap.FlagSeen`, `imap.FlagDeleted`, etc.)

## Message Copy (COPY)

To copy messages to another mailbox:

```go
// Create a sequence set from message UIDs
var seqSet imap.SeqSet
for _, msg := range messages {
    seqSet.AddNum(uint32(msg.UID))
}

// Execute COPY
copyData, err := client.Copy(seqSet, targetMailbox).Wait()
if err != nil {
    return fmt.Errorf("failed to copy messages to %s: %w", targetMailbox, err)
}

// Optional: Use copyData.DestUIDs to track the UIDs of the copied messages
```

**Important details**:
- Unlike in v1, there are no options here - just the sequence set and destination mailbox
- The command returns a `*imap.CopyData` struct, which can contain the UIDs of the messages in the destination mailbox if the server supports UIDPLUS

## Message Move (MOVE)

To move messages to another mailbox:

```go
// Create a sequence set from message UIDs
var seqSet imap.SeqSet
for _, msg := range messages {
    seqSet.AddNum(uint32(msg.UID))
}

// Check if MOVE is supported
if !client.Caps().Has(imap.CapMove) {
    // Manual fallback (go-imap v2 handles this automatically, but for illustration):
    
    // 1. Copy the messages 
    if _, err := client.Copy(seqSet, targetMailbox).Wait(); err != nil {
        return err
    }
    
    // 2. Mark original messages as deleted
    delFlags := &imap.StoreFlags{
        Op:     imap.StoreFlagsAdd,
        Silent: true,
        Flags:  []imap.Flag{imap.FlagDeleted},
    }
    if _, err := client.Store(seqSet, delFlags, nil).Wait(); err != nil {
        return err
    }
    
    // 3. Expunge the deleted messages
    if err := client.Expunge().Close(); err != nil {
        return err
    }
    
    return nil
}

// Server supports MOVE, use it directly
moveData, err := client.Move(seqSet, targetMailbox).Wait()
if err != nil {
    return fmt.Errorf("failed to move messages to %s: %w", targetMailbox, err)
}
```

**Important details**:
- The `Move` method automatically falls back to COPY+DELETE+EXPUNGE if the server doesn't support MOVE
- The command returns a `*MoveData` struct, similar to CopyData

## Message Deletion (STORE + EXPUNGE)

To delete messages (either permanently or by moving to Trash):

```go
// Create a sequence set from message UIDs
var seqSet imap.SeqSet
for _, msg := range messages {
    seqSet.AddNum(uint32(msg.UID))
}

if moveToTrash {
    // Move to Trash folder (uses the MOVE command)
    _, err := client.Move(seqSet, "Trash").Wait()
    if err != nil {
        return fmt.Errorf("failed to move messages to Trash: %w", err)
    }
} else {
    // Mark as deleted and expunge
    storeFlags := &imap.StoreFlags{
        Op:     imap.StoreFlagsAdd,
        Silent: true,
        Flags:  []imap.Flag{imap.FlagDeleted},
    }
    
    _, err := client.Store(seqSet, storeFlags, nil).Wait()
    if err != nil {
        return fmt.Errorf("failed to mark messages as deleted: %w", err)
    }
    
    // Expunge the messages
    err = client.Expunge().Close()
    if err != nil {
        return fmt.Errorf("failed to expunge messages: %w", err)
    }
}
```

**Important details**:
- The standard IMAP deletion process is a two-step operation: mark with `\Deleted` flag, then `EXPUNGE`
- Moving to Trash is a convenience operation that simply uses `MOVE` to a designated Trash folder
- `Expunge().Close()` properly cleans up the command and returns any error

## Message Export (FETCH)

To fetch and export message content:

```go
// Create a sequence set for one message
var seqSet imap.SeqSet
seqSet.AddNum(uint32(msg.UID))

// Create fetch options for the entire message
fetchOptions := &imap.FetchOptions{
    UID: true,
    BodySection: []*imap.FetchItemBodySection{
        {Peek: true}, // Peek=true means don't mark as seen
    },
}

// Fetch the message
fetchCmd := client.Fetch(seqSet, fetchOptions)
fetchedMsgs, err := fetchCmd.Collect()
if err != nil {
    return fmt.Errorf("failed to fetch message for export: %w", err)
}

if len(fetchedMsgs) == 0 {
    return fmt.Errorf("no messages returned in fetch")
}

fetchedMsg := fetchedMsgs[0]

// Access the message body from the BodySection field
if len(fetchedMsg.BodySection) == 0 || len(fetchedMsg.BodySection[0].Bytes) == 0 {
    return fmt.Errorf("empty message body")
}

// Get the raw message content
messageContent := fetchedMsg.BodySection[0].Bytes

// Write to file
err = os.WriteFile(filepath, messageContent, 0644)
if err != nil {
    return fmt.Errorf("failed to write message to file: %w", err)
}
```

**Important details**:
- The `Collect()` method reads all messages into memory - suitable for normal emails, but be careful with large attachments
- Access the message body through `fetchedMsg.BodySection[i].Bytes`
- `Peek: true` fetches the message without marking it as seen

## Complete Action Implementation Example

Here's a complete implementation of the `executeFlags`, `executeCopy`, `executeMove`, and `executeDelete` functions for go-imap/v2:

```go
// executeFlags adds or removes flags from messages
func executeFlags(client *imapclient.Client, messages []*EmailMessage, flagActions *FlagActions) error {
    if flagActions == nil || (len(flagActions.Add) == 0 && len(flagActions.Remove) == 0) {
        return nil
    }

    // Create sequence set from message UIDs
    var uidSet imap.SeqSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }

    // Add flags if specified
    if len(flagActions.Add) > 0 {
        flags := convertToIMAPFlags(flagActions.Add)
        
        log.Debug().
            Strs("flags", flagActions.Add).
            Int("message_count", len(messages)).
            Msg("Adding flags to messages")
            
        storeFlags := &imap.StoreFlags{
            Op:     imap.StoreFlagsAdd,
            Silent: true,
            Flags:  flags,
        }
        
        _, err := client.Store(uidSet, storeFlags, nil).Wait()
        if err != nil {
            return fmt.Errorf("failed to add flags: %w", err)
        }
    }

    // Remove flags if specified
    if len(flagActions.Remove) > 0 {
        flags := convertToIMAPFlags(flagActions.Remove)
        
        log.Debug().
            Strs("flags", flagActions.Remove).
            Int("message_count", len(messages)).
            Msg("Removing flags from messages")
            
        storeFlags := &imap.StoreFlags{
            Op:     imap.StoreFlagsDel,
            Silent: true,
            Flags:  flags,
        }
        
        _, err := client.Store(uidSet, storeFlags, nil).Wait()
        if err != nil {
            return fmt.Errorf("failed to remove flags: %w", err)
        }
    }

    return nil
}

// executeCopy copies messages to another mailbox
func executeCopy(client *imapclient.Client, messages []*EmailMessage, targetMailbox string) error {
    if targetMailbox == "" {
        return nil
    }

    log.Debug().
        Str("target_mailbox", targetMailbox).
        Int("message_count", len(messages)).
        Msg("Copying messages to target mailbox")

    var uidSet imap.SeqSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }
    
    _, err := client.Copy(uidSet, targetMailbox).Wait()
    if err != nil {
        return fmt.Errorf("failed to copy messages to %s: %w", targetMailbox, err)
    }

    return nil
}

// executeMove moves messages to another mailbox
func executeMove(client *imapclient.Client, messages []*EmailMessage, targetMailbox string) error {
    if targetMailbox == "" {
        return nil
    }

    log.Debug().
        Str("target_mailbox", targetMailbox).
        Int("message_count", len(messages)).
        Msg("Moving messages to target mailbox")

    var uidSet imap.SeqSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }
    
    // The Move method automatically handles the fallback if server 
    // doesn't support MOVE capability
    _, err := client.Move(uidSet, targetMailbox).Wait()
    if err != nil {
        return fmt.Errorf("failed to move messages to %s: %w", targetMailbox, err)
    }

    return nil
}

// executeDelete marks messages as deleted and optionally expunges them or moves them to Trash
func executeDelete(client *imapclient.Client, messages []*EmailMessage, deleteConfig interface{}) error {
    if deleteConfig == nil {
        return nil
    }

    var moveToTrash bool

    // Check if deleteConfig is a boolean or a DeleteConfig struct
    switch config := deleteConfig.(type) {
    case bool:
        moveToTrash = false
    case map[string]interface{}:
        // Try to extract trash setting from the map
        if trashVal, ok := config["trash"]; ok {
            if trash, ok := trashVal.(bool); ok {
                moveToTrash = trash
            }
        }
    case DeleteConfig:
        moveToTrash = config.Trash
    default:
        return fmt.Errorf("invalid delete configuration type: %T", deleteConfig)
    }

    log.Debug().
        Bool("move_to_trash", moveToTrash).
        Int("message_count", len(messages)).
        Msg("Deleting messages")

    var uidSet imap.SeqSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }

    if moveToTrash {
        // Move to trash folder using the MOVE command
        _, err := client.Move(uidSet, "Trash").Wait()
        if err != nil {
            return fmt.Errorf("failed to move messages to Trash: %w", err)
        }
    } else {
        // Mark as deleted and expunge
        storeFlags := &imap.StoreFlags{
            Op:     imap.StoreFlagsAdd,
            Silent: true,
            Flags:  []imap.Flag{imap.FlagDeleted},
        }
        
        _, err := client.Store(uidSet, storeFlags, nil).Wait()
        if err != nil {
            return fmt.Errorf("failed to mark messages as deleted: %w", err)
        }
        
        // Expunge the messages
        err = client.Expunge().Close()
        if err != nil {
            return fmt.Errorf("failed to expunge messages: %w", err)
        }
    }

    return nil
}
```

For the `executeExport` function, here's the correct implementation:

```go
// executeExport exports messages to files
func executeExport(client *imapclient.Client, messages []*EmailMessage, exportConfig *ExportConfig) error {
    if exportConfig == nil {
        return nil
    }

    // Validate export configuration
    if exportConfig.Directory == "" {
        exportConfig.Directory = "."
    }
    if exportConfig.Format == "" {
        exportConfig.Format = "eml"
    }
    if exportConfig.Format != "eml" && exportConfig.Format != "mbox" {
        return fmt.Errorf("unsupported export format: %s", exportConfig.Format)
    }

    log.Debug().
        Str("directory", exportConfig.Directory).
        Str("format", exportConfig.Format).
        Int("message_count", len(messages)).
        Msg("Exporting messages")

    // Create directory if it doesn't exist
    if err := os.MkdirAll(exportConfig.Directory, 0755); err != nil {
        return fmt.Errorf("failed to create export directory: %w", err)
    }

    // For each message, fetch full content and save to file
    for i, msg := range messages {
        // Create a sequence set for this message
        var uidSet imap.SeqSet
        uidSet.AddNum(uint32(msg.UID))
        
        // Fetch the full message content
        fetchOptions := &imap.FetchOptions{
            UID: true,
            BodySection: []*imap.FetchItemBodySection{
                {Peek: true}, // Fetch the entire message without marking as seen
            },
        }
        
        fetchedMsgs, err := client.Fetch(uidSet, fetchOptions).Collect()
        if err != nil {
            return fmt.Errorf("failed to fetch message %d for export: %w", i, err)
        }
        
        if len(fetchedMsgs) == 0 {
            log.Warn().
                Uint32("uid", msg.UID).
                Msg("Could not fetch message for export, skipping")
            continue
        }
        
        fetchedMsg := fetchedMsgs[0]
        
        // Get the message body
        if len(fetchedMsg.BodySection) == 0 {
            log.Warn().
                Uint32("uid", msg.UID).
                Msg("Message body section is empty, skipping export")
            continue
        }
        
        messageContent := fetchedMsg.BodySection[0].Bytes
        if len(messageContent) == 0 {
            log.Warn().
                Uint32("uid", msg.UID).
                Msg("Message body is empty, skipping export")
            continue
        }
        
        // Determine the filename
        var filename string
        if exportConfig.FilenameTemplate != "" {
            // TODO: Implement template parsing for filenames
            filename = fmt.Sprintf("%s-%d.%s", 
                strings.ReplaceAll(msg.Envelope.Subject, "/", "_"), 
                msg.UID, 
                exportConfig.Format)
        } else {
            filename = fmt.Sprintf("message-%d.%s", msg.UID, exportConfig.Format)
        }
        
        // Create the output file
        filePath := filepath.Join(exportConfig.Directory, filename)
        if err := os.WriteFile(filePath, messageContent, 0644); err != nil {
            return fmt.Errorf("failed to write message to file %s: %w", filePath, err)
        }
        
        log.Debug().
            Str("filename", filename).
            Uint32("uid", msg.UID).
            Msg("Exported message to file")
    }

    return nil
}
```

Finally, for the `convertToIMAPFlags` function:

```go
// convertToIMAPFlags converts string flags to IMAP flag format
func convertToIMAPFlags(flags []string) []imap.Flag {
    imapFlags := make([]imap.Flag, len(flags))
    for i, flag := range flags {
        // Standardize flag names
        flag = strings.ToLower(flag)
        switch flag {
        case "seen":
            imapFlags[i] = imap.FlagSeen
        case "answered":
            imapFlags[i] = imap.FlagAnswered
        case "flagged":
            imapFlags[i] = imap.FlagFlagged
        case "deleted":
            imapFlags[i] = imap.FlagDeleted
        case "draft":
            imapFlags[i] = imap.FlagDraft
        default:
            // Add backslash if it's a standard flag without one
            if isStandardFlag(flag) && !strings.HasPrefix(flag, "\\") {
                // Convert first character to uppercase
                if len(flag) > 0 {
                    imapFlags[i] = imap.Flag("\\" + strings.ToUpper(flag[:1]) + flag[1:])
                } else {
                    imapFlags[i] = imap.Flag(flag)
                }
            } else {
                imapFlags[i] = imap.Flag(flag)
            }
        }
    }
    return imapFlags
}
```

This implementation properly handles all the IMAP operations needed for the DSL actions component, and should fix the errors that were occurring in the previous implementation. 