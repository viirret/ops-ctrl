package service

import "fmt"

type Argument string

const (
	Binary     Argument = "binary"
	ID         Argument = "id"
	Alias      Argument = "alias"
	PID        Argument = "pid"
	WorkingDir Argument = "working_dir"
	Other      Argument = "other"
)

func (m Argument) IsValid() bool {
	switch m {
	case Binary, ID, Alias, PID, WorkingDir, Other:
		return true
	}
	return false
}

func CheckArguments(args []string) map[Argument]string {
	validArgs := make(map[Argument]string)

	binaryValues := map[string]bool{
		"-b":    true,
		"--bin": true,
	}
	handleArguments(args, validArgs, binaryValues, Binary)

	idValues := map[string]bool{
		"-i":     true,
		"-id":    true,
		"--id":   true,
		"--name": true,
	}
	handleArguments(args, validArgs, idValues, ID)

	aliasValues := map[string]bool{
		"-a":      true,
		"--alias": true,
	}
	handleArguments(args, validArgs, aliasValues, Alias)

	pidValues := map[string]bool{
		"-p":    true,
		"--pid": true,
		"-P":    true,
		"-PID":  true,
		"--PID": true,
	}
	handleArguments(args, validArgs, pidValues, PID)

	workingDirValues := map[string]bool{
		"-w":            true,
		"--working_dir": true,
		"--work_dir":    true,
	}
	handleArguments(args, validArgs, workingDirValues, WorkingDir)

	return validArgs
}

// NewMode creates a Mode from a string, validating it against known modes
func NewArgument(modeStr string) (Argument, error) {
	mode := Argument(modeStr)
	if !mode.IsValid() {
		return "", fmt.Errorf("invalid mode: %s", modeStr)
	}
	return mode, nil
}

func checkArgument(args []string, targetValues map[string]bool, targetType Argument) map[Argument]string {
	validArgs := make(map[Argument]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if targetValues[arg] {
			validArgs[targetType] = args[i+1]
		}
	}
	return validArgs
}

func mapAdd(original map[Argument]string, new map[Argument]string) {
	for key, value := range new {
		original[key] += value
	}
}

func handleArguments(args []string, originalArguments map[Argument]string, values map[string]bool, argumentType Argument) {
	newArguments := checkArgument(args, values, argumentType)
	mapAdd(originalArguments, newArguments)
}
