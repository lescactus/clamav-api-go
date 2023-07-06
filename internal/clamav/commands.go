package clamav

// The ClamavCommand represents Clamd commands
// over a tcp connection.
//
// It's recommended to prefix clamd commands with the letter z (eg. zSCAN)
// to indicate that the command will be delimited by a NULL character and
// that clamd should continue reading command data until a NULL character is read.
// The null delimiter assures that the complete command and its entire argument
// will be processed as a single command. Alternatively commands may be prefixed
// with the letter n (e.g. nSCAN) to use a newline character as the delimiter.
// Clamd replies will honour the requested terminator in turn. If clamd doesn't
// recognize the command, or the command doesn't follow the requirements specified below,
// it will reply with an error message, and close the connection.
//
// More information on clamd(8)
type ClamavCommand []byte

var (
	CmdPing            ClamavCommand = []byte("zPING\000")
	CmdVersionBytes    ClamavCommand = []byte("zVERSION\000")
	CmdReload          ClamavCommand = []byte("zRELOAD\000")
	CmdInstream        ClamavCommand = []byte("zINSTREAM\000")
	CmdStats           ClamavCommand = []byte("zSTATS\000")
	CmdVersionCommands ClamavCommand = []byte("nVERSIONCOMMANDS\n") // From https://linux.die.net/man/8/clamd, it is recommended to use nVERSIONCOMMANDS.
	CmdShutdown        ClamavCommand = []byte("zSHUTDOWN\000")
)
