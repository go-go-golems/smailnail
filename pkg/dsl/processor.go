package dsl

import (
	"fmt"
	"io"
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

	// 7. Process each message
	result := make([]*EmailMessage, 0, len(messages))
	for msgIdx, msg := range messages {
		msgStartTime := time.Now()
		log.Debug().
			Str("rule", rule.Name).
			Int("msg_index", msgIdx).
			Uint32("seq_num", msg.SeqNum).
			Str("uid", fmt.Sprintf("%d", msg.UID)).
			Msg("Processing message")

		// Determine required body sections based on structure
		bodyStructure := msg.BodyStructure
		mimePartMetadata, err := determineRequiredBodySections(bodyStructure, rule.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to determine required body sections: %w", err)
		}

		log.Debug().
			Str("rule", rule.Name).
			Int("msg_index", msgIdx).
			Str("uid", fmt.Sprintf("%d", msg.UID)).
			Int("mime_parts_count", len(mimePartMetadata)).
			Msg("Determined required MIME parts")

		var mimeParts []MimePart
		var fetchSections []*imap.FetchItemBodySection

		// Collect all fetch sections
		for _, metadata := range mimePartMetadata {
			fetchSections = append(fetchSections, metadata.FetchSection)
		}

		// If we need body sections, do a second fetch
		if len(fetchSections) > 0 {
			secondFetchStartTime := time.Now()

			// Create a sequence set for just this message
			msgSeqSet := imap.SeqSetNum(msg.SeqNum)

			// Second fetch: get required body sections
			bodyFetchOptions, err := BuildFetchOptions(rule.Output)
			if err != nil {
				return nil, fmt.Errorf("failed to build fetch options: %w", err)
			}
			bodyFetchOptions.BodyStructure = &imap.FetchItemBodyStructure{}
			bodyFetchOptions.BodySection = fetchSections

			fetchCmd := client.Fetch(msgSeqSet, bodyFetchOptions)
			defer fetchCmd.Close()

			fetchedMsg := fetchCmd.Next()
			if fetchedMsg == nil {
				return nil, fmt.Errorf("failed to fetch message body")
			}

			// Create a map to store content for each path
			contentMap := make(map[string][]byte)

			for {
				item := fetchedMsg.Next()
				if item == nil {
					break
				}

				if data, ok := item.(imapclient.FetchItemDataBodySection); ok {
					// Read the body content
					content, err := io.ReadAll(data.Literal)
					if err != nil {
						return nil, fmt.Errorf("failed to read body section: %w", err)
					}

					// Create a key from the section
					pathKey := fmt.Sprintf("%v", data.Section.Part)
					contentMap[pathKey] = content
				}
			}

			err = fetchCmd.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to close fetch command: %w", err)
			}

			// Create MimeParts using metadata and content
			for _, metadata := range mimePartMetadata {
				pathKey := fmt.Sprintf("%v", metadata.Path)
				content := contentMap[pathKey]

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

			log.Debug().
				Str("rule", rule.Name).
				Int("msg_index", msgIdx).
				Str("uid", fmt.Sprintf("%d", msg.UID)).
				Int("mime_parts_fetched", len(mimeParts)).
				Str("duration", time.Since(secondFetchStartTime).String()).
				Msg("Completed second fetch (message content)")
		}

		// Convert to our internal format
		email, err := NewEmailMessageFromIMAP(msg, mimeParts)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}

		// Set the total count field on each message
		email.TotalCount = uint32(totalFound)

		result = append(result, email)

		log.Debug().
			Str("rule", rule.Name).
			Int("msg_index", msgIdx).
			Str("uid", fmt.Sprintf("%d", msg.UID)).
			Str("duration", time.Since(msgStartTime).String()).
			Msg("Finished processing message")
	}

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
