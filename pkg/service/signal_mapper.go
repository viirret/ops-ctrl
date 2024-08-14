package service

import (
	"fmt"
	"os"
	"syscall"
)

func GetSignal(signalName string) (os.Signal, error) {
	// Map of signal names to their corresponding os.Signal values
	signalMap := map[string]os.Signal{
		"SIGHUP":    syscall.SIGHUP,
		"SIGINT":    syscall.SIGINT,
		"SIGQUIT":   syscall.SIGQUIT,
		"SIGILL":    syscall.SIGILL,
		"SIGTRAP":   syscall.SIGTRAP,
		"SIGABRT":   syscall.SIGABRT,
		"SIGBUS":    syscall.SIGBUS,
		"SIGFPE":    syscall.SIGFPE,
		"SIGKILL":   syscall.SIGKILL,
		"SIGUSR1":   syscall.SIGUSR1,
		"SIGSEGV":   syscall.SIGSEGV,
		"SIGUSR2":   syscall.SIGUSR2,
		"SIGPIPE":   syscall.SIGPIPE,
		"SIGALRM":   syscall.SIGALRM,
		"SIGTERM":   syscall.SIGTERM,
		"SIGCHLD":   syscall.SIGCHLD,
		"SIGCONT":   syscall.SIGCONT,
		"SIGSTOP":   syscall.SIGSTOP,
		"SIGTSTP":   syscall.SIGTSTP,
		"SIGTTIN":   syscall.SIGTTIN,
		"SIGTTOU":   syscall.SIGTTOU,
		"SIGURG":    syscall.SIGURG,
		"SIGXCPU":   syscall.SIGXCPU,
		"SIGXFSZ":   syscall.SIGXFSZ,
		"SIGVTALRM": syscall.SIGVTALRM,
		"SIGPROF":   syscall.SIGPROF,
		"SIGWINCH":  syscall.SIGWINCH,
		"SIGPOLL":   syscall.SIGPOLL,
		"SIGPWR":    syscall.SIGPWR,
		"SIGSYS":    syscall.SIGSYS,
	}

	// Check if the signal name exists in the map
	if sig, exists := signalMap[signalName]; exists {
		return sig, nil
	}

	// Return an error if the signal name does not exist
	return nil, fmt.Errorf("invalid signal name: %s", signalName)
}
