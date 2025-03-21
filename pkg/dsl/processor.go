package dsl

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/rs/zerolog/log"
)

// FetchMessages retrieves messages from IMAP server based on the rule
func (rule *Rule) FetchMessages(client *imapclient.Client) ([]*EmailMessage, error) {
	startTime := time.Now()
	defer func() {
		log.Debug().
			Str("rule", rule.Name).
			Str("duration", time.Since(startTime).String()).
			Msg("FetchMessages completed")
	}()

	log.Debug().
		Str("rule", rule.Name).
		Interface("search_config", rule.Search).
		Interface("output_config", rule.Output).
		Msg("Starting message fetch operation")

	// 1. Build search criteria
	criteriaStartTime := time.Now()
	criteria, options, err := BuildSearchCriteria(rule.Search, &rule.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to build search criteria: %w", err)
	}
	log.Debug().
		Str("rule", rule.Name).
		Str("duration", time.Since(criteriaStartTime).String()).
		Interface("search_options", options).
		Msg("Built search criteria and options")

	// 2. Execute search
	searchStartTime := time.Now()
	searchCmd := client.Search(criteria, options)
	searchData, err := searchCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	searchDuration := time.Since(searchStartTime)

	// 3. Check if we have results
	seqNums := searchData.AllSeqNums()
	totalFound := len(seqNums)
	if searchData.Count > 0 {
		// If we have count from the server, use that as the total
		totalFound = int(searchData.Count)
	}

	log.Debug().
		Str("rule", rule.Name).
		Str("duration", searchDuration.String()).
		Int("total_messages_found", totalFound).
		Int("seqnums_returned", len(seqNums)).
		Uint32("min", searchData.Min).
		Uint32("max", searchData.Max).
		Uint32("count", searchData.Count).
		Bool("uid", searchData.UID).
		Msg("Search completed")

	if totalFound == 0 {
		return nil, nil
	}

	// If no sequence numbers were returned but we have a count,
	// we need to fetch the most recent messages manually
	if len(seqNums) == 0 && totalFound > 0 {
		log.Debug().
			Str("rule", rule.Name).
			Int("total_count", totalFound).
			Msg("No sequence numbers returned but count > 0, fetching most recent messages")

		// Create a sequence set for the most recent messages
		var manualSeqSet imap.SeqSet
		startSeq := totalFound // Most recent message
		endSeq := 1            // First message

		// Apply limit if set
		limit := totalFound
		if rule.Output.Limit > 0 && rule.Output.Limit < limit {
			limit = rule.Output.Limit
		}

		// Apply offset if specified
		offset := rule.Output.Offset
		if offset > totalFound {
			log.Warn().
				Str("rule", rule.Name).
				Int("offset", offset).
				Int("total_messages", totalFound).
				Msg("Offset exceeds total messages count, no messages will be fetched")
			return nil, nil
		}

		// Calculate range based on offset and limit
		startSeq = totalFound - offset
		endSeq = startSeq - limit + 1
		if endSeq < 1 {
			endSeq = 1
		}

		log.Debug().
			Str("rule", rule.Name).
			Int("start_seq", startSeq).
			Int("end_seq", endSeq).
			Int("will_fetch", startSeq-endSeq+1).
			Msg("Fetching messages by sequence range")

		manualSeqSet.AddRange(uint32(endSeq), uint32(startSeq))

		// Fetch message UIDs first
		var uidFetchOptions imap.FetchOptions
		uidFetchOptions.UID = true

		uidMessages, err := client.Fetch(manualSeqSet, &uidFetchOptions).Collect()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch message UIDs: %w", err)
		}

		log.Debug().
			Str("rule", rule.Name).
			Int("messages_fetched", len(uidMessages)).
			Msg("Fetched UIDs for messages")

		// Convert to sequence numbers (just for consistency with the existing code path)
		for _, msg := range uidMessages {
			seqNums = append(seqNums, msg.SeqNum)
		}

		// Sort sequence numbers in descending order (newest first)
		for i, j := 0, len(seqNums)-1; i < j; i, j = i+1, j-1 {
			seqNums[i], seqNums[j] = seqNums[j], seqNums[i]
		}
	}

	// 4. Create sequence set from results, respecting the limit and offset if set
	var seqSet imap.SeqSet
	limit := len(seqNums)
	if rule.Output.Limit > 0 && rule.Output.Limit < limit {
		limit = rule.Output.Limit
	}

	// Apply offset if specified
	offset := rule.Output.Offset
	if offset > len(seqNums) {
		log.Warn().
			Str("rule", rule.Name).
			Int("offset", offset).
			Int("total_messages", len(seqNums)).
			Msg("Offset exceeds total messages count, no messages will be fetched")
		offset = len(seqNums)
	}

	// Use the most recent messages first (highest sequence numbers)
	startIdx := len(seqNums) - 1 - offset
	endIdx := startIdx - limit + 1
	if endIdx < 0 {
		endIdx = 0
	}

	log.Debug().
		Str("rule", rule.Name).
		Int("offset", offset).
		Int("limit", limit).
		Int("start_idx", startIdx).
		Int("end_idx", endIdx).
		Int("will_fetch", startIdx-endIdx+1).
		Msg("Pagination parameters")

	if startIdx >= 0 && startIdx < len(seqNums) {
		for i := startIdx; i >= endIdx; i-- {
			seqSet.AddNum(seqNums[i])
		}
	} else {
		log.Warn().
			Str("rule", rule.Name).
			Int("start_idx", startIdx).
			Int("total_messages", len(seqNums)).
			Msg("Invalid start index, no messages will be fetched")
		return nil, nil
	}

	// 5. Build initial fetch options for metadata and structure
	fetchOptionsStartTime := time.Now()
	fetchOptions, err := BuildFetchOptions(rule.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to build fetch options: %w", err)
	}
	log.Debug().
		Str("rule", rule.Name).
		Str("duration", time.Since(fetchOptionsStartTime).String()).
		Interface("fetch_options", fetchOptions).
		Msg("Built fetch options")

	fetchOptions.BodySection = []*imap.FetchItemBodySection{}

	// 6. First fetch: get metadata and structure
	firstFetchStartTime := time.Now()
	messages, err := client.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	log.Debug().
		Str("rule", rule.Name).
		Str("duration", time.Since(firstFetchStartTime).String()).
		Int("messages_fetched", len(messages)).
		Msg("Completed first fetch (metadata and structure)")

	// 7. Process messages in batches to reduce round trips
	result := make([]*EmailMessage, 0, len(messages))

	// First pass: determine all MIME parts we need to fetch
	type MessageFetchInfo struct {
		Message          *imapclient.FetchMessageBuffer
		MimePartMetadata []MimePartMetadata
		Index            int
	}

	messagesToFetch := make([]MessageFetchInfo, 0, len(messages))

	for msgIdx, msg := range messages {
		log.Debug().
			Str("rule", rule.Name).
			Int("msg_index", msgIdx).
			Uint32("seq_num", msg.SeqNum).
			Str("uid", fmt.Sprintf("%d", msg.UID)).
			Msg("Analyzing message structure")

		// Determine required body sections based on structure
		bodyStructure := msg.BodyStructure
		mimePartMetadata, err := determineRequiredBodySections(bodyStructure, rule.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to determine required body sections: %w", err)
		}

		// Only add to fetch list if it has MIME parts to fetch
		if len(mimePartMetadata) > 0 {
			messagesToFetch = append(messagesToFetch, MessageFetchInfo{
				Message:          msg,
				MimePartMetadata: mimePartMetadata,
				Index:            msgIdx,
			})
		} else {
			// If no MIME parts to fetch, process it immediately
			email, err := NewEmailMessageFromIMAP(msg, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert message: %w", err)
			}
			email.TotalCount = uint32(totalFound)
			result = append(result, email)

			log.Debug().
				Str("rule", rule.Name).
				Int("msg_index", msgIdx).
				Str("uid", fmt.Sprintf("%d", msg.UID)).
				Msg("Processed message (no MIME parts)")
		}
	}

	// Skip the batch fetch if no messages need MIME parts
	if len(messagesToFetch) == 0 {
		log.Debug().
			Str("rule", rule.Name).
			Msg("No MIME parts needed for any message, skipping content fetch")
		return result, nil
	}

	// Second pass: batch fetch MIME parts for all messages
	batchFetchStartTime := time.Now()

	// Create a combined sequence set with all messages that need MIME parts
	var batchSeqSet imap.SeqSet
	allFetchSections := []*imap.FetchItemBodySection{}

	// Map to track which sections belong to which message
	sectionToMessageMap := make(map[string]MessageFetchInfo)

	for _, msgInfo := range messagesToFetch {
		// Add message to the sequence set
		batchSeqSet.AddNum(msgInfo.Message.SeqNum)

		// Add all sections for this message, with a mapping back to the message
		for _, metadata := range msgInfo.MimePartMetadata {
			// Create a unique identifier for this section
			sectionKey := fmt.Sprintf("%d:%v", msgInfo.Message.SeqNum, metadata.Path)

			// Store the mapping
			sectionToMessageMap[sectionKey] = msgInfo

			// Add the section to the fetch request
			allFetchSections = append(allFetchSections, metadata.FetchSection)
		}
	}

	// Create batch fetch options
	batchFetchOptions, err := BuildFetchOptions(rule.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to build batch fetch options: %w", err)
	}
	batchFetchOptions.BodyStructure = &imap.FetchItemBodyStructure{}
	batchFetchOptions.BodySection = allFetchSections

	log.Debug().
		Str("rule", rule.Name).
		Int("messages_to_fetch", len(messagesToFetch)).
		Int("total_sections", len(allFetchSections)).
		Msg("Starting batch fetch for MIME parts")

	// Execute the batch fetch
	batchFetchCmd := client.Fetch(batchSeqSet, batchFetchOptions)
	defer batchFetchCmd.Close()

	// Process the batch fetch results
	contentMap := make(map[string][]byte)

	for {
		fetchedMsg := batchFetchCmd.Next()
		if fetchedMsg == nil {
			break
		}

		for {
			item := fetchedMsg.Next()
			if item == nil {
				break
			}

			if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
				if data.Literal == nil {
					log.Warn().
						Str("rule", rule.Name).
						Uint32("seq_num", fetchedMsg.SeqNum).
						Str("section", fmt.Sprintf("%v", data.Section)).
						Msg("No literal found for body section")
					continue
				}

				// Read the body content
				content, err := io.ReadAll(data.Literal)
				if err != nil {
					return nil, fmt.Errorf("failed to read body section: %w", err)
				}

				// Create a key from the sequence number and section
				sectionKey := fmt.Sprintf("%d:%v", fetchedMsg.SeqNum, data.Section.Part)
				contentMap[sectionKey] = content
			}
		}
	}

	err = batchFetchCmd.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close batch fetch command: %w", err)
	}

	log.Debug().
		Str("rule", rule.Name).
		Int("sections_fetched", len(contentMap)).
		Str("duration", time.Since(batchFetchStartTime).String()).
		Msg("Completed batch fetch for MIME parts")

	// Third pass: process all messages with their fetched content
	processStartTime := time.Now()

	// Group content by message sequence number
	messageContents := make(map[uint32]map[string][]byte)
	for sectionKey, content := range contentMap {
		parts := strings.SplitN(sectionKey, ":", 2)
		if len(parts) != 2 {
			continue
		}

		seqNum := uint32(0)
		_, err := fmt.Sscanf(parts[0], "%d", &seqNum)
		if err != nil {
			continue
		}

		if _, exists := messageContents[seqNum]; !exists {
			messageContents[seqNum] = make(map[string][]byte)
		}

		messageContents[seqNum][parts[1]] = content
	}

	// Process each message with its content
	for _, msgInfo := range messagesToFetch {
		msgStartTime := time.Now()
		seqNum := msgInfo.Message.SeqNum

		// Get content for this message
		msgContent, exists := messageContents[seqNum]
		if !exists {
			log.Warn().
				Str("rule", rule.Name).
				Uint32("seq_num", seqNum).
				Msg("No content found for message in batch fetch results")

			// Create a message without content
			email, err := NewEmailMessageFromIMAP(msgInfo.Message, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to convert message: %w", err)
			}
			email.TotalCount = uint32(totalFound)
			result = append(result, email)
			continue
		}

		// Create MIME parts from fetched content
		var mimeParts []MimePart
		for _, metadata := range msgInfo.MimePartMetadata {
			pathKey := fmt.Sprintf("%v", metadata.Path)
			content, exists := msgContent[pathKey]

			if !exists {
				log.Warn().
					Str("rule", rule.Name).
					Uint32("seq_num", seqNum).
					Str("path", pathKey).
					Msg("MIME part not found in fetch results")
				continue
			}

			mimePart := MimePart{
				Type:     metadata.Type,
				Subtype:  metadata.Subtype,
				Content:  string(content),
				Size:     uint32(len(content)),
				Charset:  metadata.Params["charset"],
				Filename: metadata.Filename,
			}
			mimeParts = append(mimeParts, mimePart)
		}

		// Convert to our internal format
		email, err := NewEmailMessageFromIMAP(msgInfo.Message, mimeParts)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}

		// Set the total count field
		email.TotalCount = uint32(totalFound)

		result = append(result, email)

		log.Debug().
			Str("rule", rule.Name).
			Int("msg_index", msgInfo.Index).
			Str("uid", fmt.Sprintf("%d", msgInfo.Message.UID)).
			Int("mime_parts_processed", len(mimeParts)).
			Str("duration", time.Since(msgStartTime).String()).
			Msg("Processed message with content")
	}

	log.Debug().
		Str("rule", rule.Name).
		Int("messages_processed", len(result)).
		Str("duration", time.Since(processStartTime).String()).
		Msg("Finished processing all messages")

	log.Info().
		Str("rule", rule.Name).
		Int("total_messages_found", totalFound).
		Int("messages_fetched", len(messages)).
		Int("messages_processed", len(result)).
		Str("duration", time.Since(startTime).String()).
		Msg("Fetch messages operation complete")

	return result, nil
}

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
		Str("total_duration", time.Since(startTime).String()).
		Msg("Rule processing complete")

	return nil
}
