package enrich

type Options struct {
	AccountKey string
	Mailbox    string
	Rebuild    bool
	DryRun     bool
	BatchSize  int
}

type ThreadsReport struct {
	MessagesProcessed int
	ThreadsCreated    int
	ThreadsUpdated    int
	ElapsedMS         int64
}

type SendersReport struct {
	SendersCreated    int
	SendersUpdated    int
	MessagesTagged    int
	PrivateRelayCount int
	ElapsedMS         int64
}

type UnsubscribeReport struct {
	SendersWithUnsubscribe int
	MailtoLinks            int
	HTTPLinks              int
	OneClickLinks          int
	ElapsedMS              int64
}

type AllReport struct {
	Senders     SendersReport
	Threads     ThreadsReport
	Unsubscribe UnsubscribeReport
	ElapsedMS   int64
}
