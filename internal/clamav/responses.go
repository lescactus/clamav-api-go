package clamav

// The ClamavResponse represents Clamd responses
// to commands over a tcp connection.
type ClamavResponse []byte

var (
	RespPing                         ClamavResponse = []byte("PONG")
	RespReload                       ClamavResponse = []byte("RELOADING")
	RespScan                         ClamavResponse = []byte("stream: OK")
	RespErrUnknownCommand            ClamavResponse = []byte("UNKNOWN COMMAND")
	RespErrScanFileSizeLimitExceeded ClamavResponse = []byte("INSTREAM size limit exceeded. ERROR")
)
