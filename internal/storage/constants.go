package storage

// Config directory internal files
const (
	ConfigFileName    = "config.json"
	StateFileName     = "state.json"
	StateLockFileName = "state.json.lock"
	StateTmpFileName  = "state.json.tmp"
	ConfigTmpFileName = "config.json.tmp"
	LogFileName       = "gistsync.log"
	LogsDirName       = "logs"
)

// GetIgnoredConfigFiles returns a list of files in the config directory that 
// should be ignored for hashing and watcher events to prevent feedback loops.
func GetIgnoredConfigFiles() []string {
	return []string{
		StateFileName,
		StateLockFileName,
		StateTmpFileName,
		ConfigTmpFileName,
		LogFileName,
		LogsDirName,
	}
}

// IsIgnoredConfigFile checks if a file name should be ignored in the config directory.
func IsIgnoredConfigFile(name string) bool {
	ignored := GetIgnoredConfigFiles()
	for _, f := range ignored {
		if f == name {
			return true
		}
	}
	return false
}
