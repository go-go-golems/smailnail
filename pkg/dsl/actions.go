package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/rs/zerolog/log"
)

// ExecuteActions performs the specified actions on the matched messages
func ExecuteActions(client *imapclient.Client, messages []*EmailMessage, actions *ActionConfig) error {
	if actions == nil || reflect.DeepEqual(*actions, ActionConfig{}) {
		return nil
	}

	startTime := time.Now()
	log.Debug().
		Int("message_count", len(messages)).
		Msg("Starting to execute actions on messages")

	// Execute flag operations
	if actions.Flags != nil {
		if err := executeFlags(client, messages, actions.Flags); err != nil {
			return fmt.Errorf("failed to execute flag actions: %w", err)
		}
	}

	// Execute copy operation before move or delete
	if actions.CopyTo != "" {
		if err := executeCopy(client, messages, actions.CopyTo); err != nil {
			return fmt.Errorf("failed to copy messages to %s: %w", actions.CopyTo, err)
		}
	}

	// Execute move operation
	if actions.MoveTo != "" {
		if err := executeMove(client, messages, actions.MoveTo); err != nil {
			return fmt.Errorf("failed to move messages to %s: %w", actions.MoveTo, err)
		}
		// If we've moved the messages, we don't need to delete them separately
		log.Debug().
			Str("duration", time.Since(startTime).String()).
			Msg("Actions executed successfully")
		return nil
	}

	// Execute delete operation if specified
	if actions.Delete != nil {
		if err := executeDelete(client, messages, actions.Delete); err != nil {
			return fmt.Errorf("failed to delete messages: %w", err)
		}
	}

	// Execute export operation if specified
	if actions.Export != nil {
		if err := executeExport(client, messages, actions.Export); err != nil {
			return fmt.Errorf("failed to export messages: %w", err)
		}
	}

	log.Debug().
		Str("duration", time.Since(startTime).String()).
		Msg("Actions executed successfully")

	return nil
}

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

		_, err := client.Store(uidSet, storeFlags, nil).Collect()
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

		_, err := client.Store(uidSet, storeFlags, nil).Collect()
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

		_, err := client.Store(uidSet, storeFlags, nil).Collect()
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

// isStandardFlag checks if a flag is one of the standard IMAP flags
func isStandardFlag(flag string) bool {
	flag = strings.ToLower(flag)
	standardFlags := []string{
		"seen", "answered", "flagged", "deleted", "draft", "recent",
	}

	for _, std := range standardFlags {
		if flag == std {
			return true
		}
	}

	return false
}
