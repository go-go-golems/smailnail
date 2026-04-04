# Tasks

## Current Phase

- [ ] Review the detailed guide in `design/01-inbox-browser-and-time-based-exploration-guide.md`
- [ ] Confirm the first implementation slice: messages first, threads first, or shared browsing primitives first

## Backend Query Surface

- [ ] Add list/detail endpoints for messages
- [ ] Add list/detail endpoints for threads
- [ ] Extend sender endpoints with search and richer filters
- [ ] Add shared filter parsing for mailbox, sender, tag, annotation state, and time range
- [ ] Decide whether to expose thread/message identifiers as composite keys or synthetic IDs at the API layer
- [ ] Add endpoints or query options for timeline-based aggregation
- [ ] Add tests for pagination, search, and filter combinations

## Frontend View Primitives

- [ ] Define reusable filter state for inbox exploration
- [ ] Define reusable table/list row types for messages and threads
- [ ] Create an embeddable message list view
- [ ] Create an embeddable thread list view
- [ ] Create an embeddable message detail view
- [ ] Create an embeddable thread detail view
- [ ] Create a sender-scoped message view
- [ ] Create a tag-scoped message view
- [ ] Create a time-range exploration view for messages and annotations

## Search And Filtering

- [ ] Add sender search text input and domain/mailbox filters
- [ ] Add message search using FTS-backed search text where available
- [ ] Add thread search by subject, participants, and annotation state
- [ ] Add time-range filters that work consistently across message, sender, thread, and annotation contexts
- [ ] Add filter chips or summary state so embedded views stay understandable when reused

## Annotation Integration

- [ ] Show annotation counts in message/thread list rows where useful
- [ ] Allow browsing all messages for a sender from sender detail
- [ ] Allow browsing all messages for a tag from an annotation/tag context
- [ ] Allow viewing annotations and messages together over a time period

## Validation

- [ ] Add backend tests for list/detail/search behavior
- [ ] Add frontend tests for reusable view composition
- [ ] Add a manual playbook for browsing sender, tag, and time-slice contexts

## Documentation

- [ ] Keep the diary updated during implementation
- [ ] Relate core backend and frontend files to the focused design doc
- [ ] Run `docmgr doctor --ticket SMN-20260403-INBOX-BROWSER`
- [ ] Upload the document bundle to reMarkable after the guide is finalized
