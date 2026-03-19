import { useEffect } from "react";
import { useAppDispatch, useAppSelector } from "../../store";
import { MailboxSidebar } from "./MailboxSidebar";
import {
  clearSelectedMessage,
  closeExplorer,
  fetchMailboxes,
  fetchMessageDetail,
  fetchMessages,
  selectMailbox,
} from "./mailboxSlice";
import { MessageDetail } from "./MessageDetail";
import { MessageList } from "./MessageList";

interface MailboxExplorerProps {
  accountId: string;
  accountLabel: string;
  onBack: () => void;
}

export function MailboxExplorer({
  accountId,
  accountLabel,
  onBack,
}: MailboxExplorerProps) {
  const dispatch = useAppDispatch();
  const {
    mailboxes,
    mailboxLoadState,
    selectedMailbox,
    messages,
    messageLoadState,
    totalCount,
    offset,
    limit,
    selectedMessage,
    messageDetailLoadState,
    error,
  } = useAppSelector((s) => s.mailbox);

  useEffect(() => {
    dispatch(fetchMailboxes(accountId));
    return () => {
      dispatch(closeExplorer());
    };
  }, [dispatch, accountId]);

  // Auto-select INBOX when mailboxes load
  useEffect(() => {
    if (mailboxLoadState === "loaded" && !selectedMailbox && mailboxes.length > 0) {
      const inbox = mailboxes.find(
        (m) => m.name.toUpperCase() === "INBOX",
      );
      const target = inbox ?? mailboxes[0];
      if (target) {
        dispatch(selectMailbox(target.name));
      }
    }
  }, [dispatch, mailboxLoadState, selectedMailbox, mailboxes]);

  // Fetch messages when mailbox changes
  useEffect(() => {
    if (selectedMailbox) {
      dispatch(
        fetchMessages({ accountId, mailbox: selectedMailbox, offset: 0, limit }),
      );
    }
  }, [dispatch, accountId, selectedMailbox, limit]);

  function handleSelectMailbox(name: string) {
    dispatch(selectMailbox(name));
  }

  function handleSelectMessage(uid: number) {
    if (selectedMailbox) {
      dispatch(fetchMessageDetail({ accountId, mailbox: selectedMailbox, uid }));
    }
  }

  function handlePageChange(newOffset: number) {
    if (selectedMailbox) {
      dispatch(
        fetchMessages({
          accountId,
          mailbox: selectedMailbox,
          offset: newOffset,
          limit,
        }),
      );
    }
  }

  function handleBack() {
    dispatch(closeExplorer());
    onBack();
  }

  return (
    <div data-widget="mailbox-explorer" data-part="root">
      {/* Header */}
      <div className="d-flex align-items-center mb-3">
        <button
          className="btn btn-link text-decoration-none p-0 me-2 text-body-secondary"
          onClick={handleBack}
        >
          &larr;
        </button>
        <h2 className="h5 mb-0">{accountLabel}</h2>
      </div>

      {error && (
        <div className="alert alert-danger mb-3" role="alert">
          {error}
        </div>
      )}

      {/* Message detail view (full width when viewing a message) */}
      {selectedMessage ? (
        <MessageDetail
          message={selectedMessage}
          loading={messageDetailLoadState === "loading"}
          onBack={() => dispatch(clearSelectedMessage())}
        />
      ) : (
        <div className="row g-0">
          {/* Sidebar */}
          <div className="col-4 col-md-3 border-end" style={{ minHeight: 300 }}>
            <div className="p-2">
              <strong className="small text-body-secondary">Mailboxes</strong>
            </div>
            <MailboxSidebar
              mailboxes={mailboxes}
              selected={selectedMailbox}
              loading={mailboxLoadState === "loading"}
              onSelect={handleSelectMailbox}
            />
          </div>

          {/* Message list */}
          <div className="col-8 col-md-9">
            {selectedMailbox ? (
              <MessageList
                messages={messages}
                loading={messageLoadState === "loading"}
                totalCount={totalCount}
                offset={offset}
                limit={limit}
                selectedUid={null}
                onSelect={handleSelectMessage}
                onPageChange={handlePageChange}
              />
            ) : (
              <p className="text-body-secondary small text-center py-4">
                Select a mailbox to view messages.
              </p>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
