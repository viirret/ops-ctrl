package service

type Argument string

const (
	Binary           Argument = "binary"           // Program binary path
	ID               Argument = "id"               // Unique indentifier for the service
	Alias            Argument = "alias"            // Alias for binary path, aliases found in config.toml
	Envs             Argument = "env"              // Environment variables for the program binary
	ProgramArguments Argument = "program_argument" // Arguments for the program binary
	PID              Argument = "pid"              // PID number for the service
	WorkingDir       Argument = "working_dir"      // Working directory for the program
)

func (m Argument) IsValid() bool {
	switch m {
	case Binary, ID, Alias, Envs, ProgramArguments, PID, WorkingDir:
		return true
	}
	return false
}

func (m Argument) SupportsArrays() bool {
	switch m {
	case Envs, ProgramArguments:
		return true
	}
	return false
}

type Mergeable interface {
	Merge(other interface{}) interface{}
}

// Base merge function that both Int and String will use
func mergeValues[T any](a, b interface{}) interface{} {
	valA, okA := a.(T)
	valB, okB := b.(T)
	if okA && okB {
		switch x := any(valA).(type) {
		case int:
			return x + any(valB).(int)
		case string:
			return x + any(valB).(string)
		}
	}
	return a
}

// Int type that implements Mergeable interface
type mergeInt int

func (a mergeInt) Merge(other interface{}) interface{} {
	return mergeValues[mergeInt](a, other)
}

// String type that implements Mergeable interface
type mergeString string

func (s mergeString) Merge(other interface{}) interface{} {
	return mergeValues[mergeString](s, other)
}

func CheckArguments(args []string) map[Argument]interface{} {
	validArgs := make(map[Argument]interface{})

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

	envValues := map[string]bool{
		"-e":    true,
		"--env": true,
	}
	handleArguments(args, validArgs, envValues, Envs)

	programArgValues := map[string]bool{
		"-arg":  true,
		"--arg": true,
	}
	handleArguments(args, validArgs, programArgValues, ProgramArguments)

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

func checkArgument(args []string, targetValues map[string]bool, targetType Argument) map[Argument]interface{} {
	validArgs := make(map[Argument]interface{})
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if targetValues[arg] {
			validArgs[targetType] = args[i+1]
		}
	}
	return validArgs
}

func addArguments(original map[Argument]interface{}, new map[Argument]interface{}) {
	for key, value := range new {
		if originalVal, ok := original[key]; ok {
			if mergeable, canMerge := originalVal.(Mergeable); canMerge {
				original[key] = mergeable.Merge(value)
			}
		} else {
			original[key] = value
		}
	}
}

func handleArguments(args []string, originalArguments map[Argument]interface{}, values map[string]bool, argumentType Argument) {
	newArguments := checkArgument(args, values, argumentType)
	addArguments(originalArguments, newArguments)
}
