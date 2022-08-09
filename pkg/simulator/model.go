package simulator

type StartScan struct {
	Data string
}

type FetchScanResults struct {
	JobId string
}

type ScanResults struct {
	IsDone bool
	Data   string
}
