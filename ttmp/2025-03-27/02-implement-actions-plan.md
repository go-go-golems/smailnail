# Implementation Plan for IMAP DSL Actions

This document outlines a detailed plan for implementing the actions component of the IMAP DSL. The actions allow users to specify operations to perform on messages that match search criteria, such as adding/removing flags, moving/copying messages to other mailboxes, and deleting messages.

## 1. Overview of the Actions Feature

The IMAP DSL actions will allow users to:

- Add or remove flags from messages (seen, flagged, deleted, etc.)
- Move messages to different mailboxes
- Copy messages to different mailboxes
- Delete messages (permanently or move to trash)
- Export messages to files in various formats

Example of an action section in a YAML rule:

```yaml
actions:
  flags:
    add: ["seen", "flagged"]
    remove: ["draft"]
  move_to: "Archive/2025"
  copy_to: "Backup"
```

## 2. Implementation Tasks

### 2.1. Update Type Definitions

- [ ] Add `Actions` struct to the `Rule` struct in `types.go`
- [ ] Define `ActionConfig` struct with fields for each action type
- [ ] Define supporting types for flag operations, move/copy, delete, and export
- [ ] Update the Rule validation method to validate the actions

### 2.2. Update Parser

- [ ] Modify `ParseRuleString` in `parser.go` to handle the actions section
- [ ] Add UnmarshalYAML method for ActionConfig to handle complex structures

### 2.3. Implement Action Execution

- [ ] Create a new file `actions.go` to contain action execution logic
- [ ] Implement `ExecuteActions` function to process actions against fetched messages
- [ ] Implement helper functions for each action type:
  - [ ] `executeFlags` - Add/remove flags
  - [ ] `executeMove` - Move messages to another mailbox
  - [ ] `executeCopy` - Copy messages to another mailbox
  - [ ] `executeDelete` - Delete messages or move to trash
  - [ ] `executeExport` - Save messages to files

### 2.4. Update Processor

- [ ] Modify `ProcessRule` in `processor.go` to apply actions after fetching messages
- [ ] Add appropriate error handling and logging for actions

## 3. Detailed Type Definitions

### 3.1. ActionConfig Struct

```go
// ActionConfig defines actions to perform on matched messages
type ActionConfig struct {
    // Flag operations
    Flags *FlagActions `yaml:"flags,omitempty"`
    
    // Move/Copy operations
    MoveTo string `yaml:"move_to,omitempty"`
    CopyTo string `yaml:"copy_to,omitempty"`
    
    // Delete operation
    Delete interface{} `yaml:"delete,omitempty"` // Can be bool or DeleteConfig
    
    // Export operation
    Export *ExportConfig `yaml:"export,omitempty"`
}

// FlagActions defines add/remove flag operations
type FlagActions struct {
    Add    []string `yaml:"add,omitempty"`
    Remove []string `yaml:"remove,omitempty"`
}

// DeleteConfig provides options for delete operations
type DeleteConfig struct {
    Trash bool `yaml:"trash,omitempty"` // If true, move to trash; if false, delete permanently
}

// ExportConfig defines options for exporting messages
type ExportConfig struct {
    Format           string `yaml:"format,omitempty"`           // eml, mbox
    Directory        string `yaml:"directory,omitempty"`        // Where to save files
    FilenameTemplate string `yaml:"filename_template,omitempty"` // Template for filenames
}
```

### 3.2. Updated Rule Struct

```go
// Rule represents a complete IMAP DSL rule
type Rule struct {
    Name        string       `yaml:"name"`
    Description string       `yaml:"description"`
    Search      SearchConfig `yaml:"search"`
    Output      OutputConfig `yaml:"output"`
    Actions     ActionConfig `yaml:"actions"`
}
```

## 4. Action Implementation Details

### 4.1. Flag Operations

Flags will be added or removed using the IMAP STORE command:

```go
func executeFlags(client *imapclient.Client, messages []*EmailMessage, flagActions *FlagActions) error {
    if flagActions == nil {
        return nil
    }
    
    // Create sequence sets from message UIDs
    var uidSet imap.UIDSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }
    
    // Add flags if specified
    if len(flagActions.Add) > 0 {
        flags := convertToIMAPFlags(flagActions.Add)
        store := &imap.StoreFlags{
            Op:    imap.StoreFlagsAdd,
            Flags: flags,
        }
        
        cmd := client.Store(uidSet, store, nil)
        if err := cmd.Wait(); err != nil {
            return fmt.Errorf("failed to add flags: %w", err)
        }
    }
    
    // Remove flags if specified
    if len(flagActions.Remove) > 0 {
        flags := convertToIMAPFlags(flagActions.Remove)
        store := &imap.StoreFlags{
            Op:    imap.StoreFlagsRemove,
            Flags: flags,
        }
        
        cmd := client.Store(uidSet, store, nil)
        if err := cmd.Wait(); err != nil {
            return fmt.Errorf("failed to remove flags: %w", err)
        }
    }
    
    return nil
}

// convertToIMAPFlags converts string flags to IMAP flag format
func convertToIMAPFlags(flags []string) []imap.Flag {
    imapFlags := make([]imap.Flag, len(flags))
    for i, flag := range flags {
        // Convert to standard IMAP flag format if needed
        flag = strings.ToLower(flag)
        if flag == "seen" {
            imapFlags[i] = "\\Seen"
        } else if flag == "answered" {
            imapFlags[i] = "\\Answered"
        } else if flag == "flagged" {
            imapFlags[i] = "\\Flagged"
        } else if flag == "deleted" {
            imapFlags[i] = "\\Deleted"
        } else if flag == "draft" {
            imapFlags[i] = "\\Draft"
        } else if flag == "recent" {
            imapFlags[i] = "\\Recent"
        } else if flag == "important" {
            imapFlags[i] = "\\Important"
        } else if !strings.HasPrefix(flag, "\\") && !strings.HasPrefix(flag, "$") {
            // Add backslash for standard flags that weren't properly formatted
            imapFlags[i] = imap.Flag(flag)
        } else {
            imapFlags[i] = imap.Flag(flag)
        }
    }
    return imapFlags
}
```

### 4.2. Move/Copy Operations

The MOVE and COPY commands will be used to transfer messages to other mailboxes:

```go
func executeMove(client *imapclient.Client, messages []*EmailMessage, targetMailbox string) error {
    if targetMailbox == "" {
        return nil
    }
    
    var uidSet imap.UIDSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }
    
    moveCmd := client.Move(uidSet, targetMailbox)
    if err := moveCmd.Wait(); err != nil {
        return fmt.Errorf("failed to move messages to %s: %w", targetMailbox, err)
    }
    
    return nil
}

func executeCopy(client *imapclient.Client, messages []*EmailMessage, targetMailbox string) error {
    if targetMailbox == "" {
        return nil
    }
    
    var uidSet imap.UIDSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }
    
    copyCmd := client.Copy(uidSet, targetMailbox)
    if err := copyCmd.Wait(); err != nil {
        return fmt.Errorf("failed to copy messages to %s: %w", targetMailbox, err)
    }
    
    return nil
}
```

### 4.3. Delete Operations

Deleting messages will either mark them with the \Deleted flag and expunge, or move them to a Trash folder:

```go
func executeDelete(client *imapclient.Client, messages []*EmailMessage, deleteConfig interface{}) error {
    if deleteConfig == nil {
        return nil
    }
    
    var moveToTrash bool
    
    // Check if deleteConfig is a boolean or a DeleteConfig struct
    switch config := deleteConfig.(type) {
    case bool:
        // If true, permanently delete
        moveToTrash = false
    case DeleteConfig:
        // Use the trash setting
        moveToTrash = config.Trash
    default:
        return fmt.Errorf("invalid delete configuration")
    }
    
    var uidSet imap.UIDSet
    for _, msg := range messages {
        uidSet.AddNum(uint32(msg.UID))
    }
    
    if moveToTrash {
        // Move to trash folder
        moveCmd := client.Move(uidSet, "Trash")
        if err := moveCmd.Wait(); err != nil {
            return fmt.Errorf("failed to move messages to Trash: %w", err)
        }
    } else {
        // Mark as deleted and expunge
        store := &imap.StoreFlags{
            Op:    imap.StoreFlagsAdd,
            Flags: []imap.Flag{"\\Deleted"},
        }
        
        storeCmd := client.Store(uidSet, store, nil)
        if err := storeCmd.Wait(); err != nil {
            return fmt.Errorf("failed to mark messages as deleted: %w", err)
        }
        
        expungeCmd := client.Expunge()
        if err := expungeCmd.Wait(); err != nil {
            return fmt.Errorf("failed to expunge messages: %w", err)
        }
    }
    
    return nil
}
```

### 4.4. Export Operations (Optional for first implementation)

Exporting messages to files will require fetching the full message content and saving it:

```go
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
    
    // Create directory if it doesn't exist
    if err := os.MkdirAll(exportConfig.Directory, 0755); err != nil {
        return fmt.Errorf("failed to create export directory: %w", err)
    }
    
    // For each message, fetch full content and save to file
    for _, msg := range messages {
        // Implement file export logic here (requires fetching full message content)
        // This is a more complex feature that can be implemented in a later phase
    }
    
    return nil
}
```

### 4.5. Main Action Execution

The overall action execution function will look like:

```go
// ExecuteActions performs the specified actions on the matched messages
func ExecuteActions(client *imapclient.Client, messages []*EmailMessage, actions *ActionConfig) error {
    if actions == nil {
        return nil
    }
    
    // Execute flag operations
    if actions.Flags != nil {
        if err := executeFlags(client, messages, actions.Flags); err != nil {
            return err
        }
    }
    
    // Execute copy operation before move or delete
    if actions.CopyTo != "" {
        if err := executeCopy(client, messages, actions.CopyTo); err != nil {
            return err
        }
    }
    
    // Execute move operation
    if actions.MoveTo != "" {
        if err := executeMove(client, messages, actions.MoveTo); err != nil {
            return err
        }
        // If we've moved the messages, we don't need to delete them separately
        return nil
    }
    
    // Execute delete operation if specified
    if actions.Delete != nil {
        if err := executeDelete(client, messages, actions.Delete); err != nil {
            return err
        }
    }
    
    // Execute export operation if specified
    if actions.Export != nil {
        if err := executeExport(client, messages, actions.Export); err != nil {
            return err
        }
    }
    
    return nil
}
```

## 5. Updated ProcessRule Function

The ProcessRule function in processor.go will be updated to include action execution:

```go
// ProcessRule executes an IMAP rule
func ProcessRule(client *imapclient.Client, rule *Rule) error {
    startTime := time.Now()
    log.Info().
        Str("rule", rule.Name).
        Msg("Processing rule")

    // 1. Fetch messages
    messages, err := rule.FetchMessages(client)
    if err != nil {
        return err
    }

    if len(messages) == 0 {
        log.Warn().
            Str("rule", rule.Name).
            Msg("No messages found matching the criteria")
        return nil
    }

    // 2. Output messages
    outputStartTime := time.Now()
    err = OutputMessages(messages, rule.Output)
    if err != nil {
        return fmt.Errorf("failed to output messages: %w", err)
    }

    log.Info().
        Str("rule", rule.Name).
        Int("messages_output", len(messages)).
        Str("output_duration", time.Since(outputStartTime).String()).
        Msg("Messages output complete")

    // 3. Execute actions
    if !reflect.DeepEqual(rule.Actions, ActionConfig{}) {
        actionsStartTime := time.Now()
        err = ExecuteActions(client, messages, &rule.Actions)
        if err != nil {
            return fmt.Errorf("failed to execute actions: %w", err)
        }

        log.Info().
            Str("rule", rule.Name).
            Str("actions_duration", time.Since(actionsStartTime).String()).
            Msg("Actions executed successfully")
    }

    log.Info().
        Str("rule", rule.Name).
        Int("messages_processed", len(messages)).
        Str("total_duration", time.Since(startTime).String()).
        Msg("Rule processing complete")

    return nil
}
```

## 6. Implementation Order

1. First implement the type definitions in `types.go`
2. Update `Rule.Validate()` for action validation
3. Create the `actions.go` file with the `ExecuteActions` function
4. Implement flag operations, which are the most straightforward
5. Implement move/copy operations
6. Implement delete operations
7. Update the processor to call action execution
8. Add more extensive tests
9. Implement export operations (could be deferred to a later phase)

## 8. Documentation Updates

1. Update the IMAP DSL documentation to include comprehensive examples of actions
2. Add code documentation for new types and functions
3. Include examples of common use cases, such as:
   - Archiving old emails
   - Auto-categorizing messages
   - Cleaning up mailboxes
   - Creating backups of important messages

## 9. Future Enhancements (NOT TODAY)

1. Add support for conditional actions based on message properties
2. Implement templating for export filenames and paths
3. Add support for more export formats
4. Implement asynchronous action execution for large message sets
5. Add transaction-like behavior with rollback on failure
