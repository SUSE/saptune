package system

// defaultCommand contains all available 'command - action' combinations
var defaultCommand = map[string]bool{
	"daemon start":        false,
	"daemon status":       false,
	"daemon stop":         false,
	"service apply":       false,
	"service start":       false,
	"service status":      false,
	"service stop":        false,
	"service restart":     false,
	"service revert":      false,
	"service reload":      false,
	"service takeover":    false,
	"service enable":      false,
	"service disable":     false,
	"service enablestart": false,
	"service disablestop": false,
	"note list":           false,
	"note revertall":      false,
	"note enabled":        false,
	"note applied":        false,
	"note apply":          false,
	"note simulate":       false,
	"note customise":      false,
	"note create":         false,
	"note edit":           false,
	"note revert":         false,
	"note show":           false,
	"note delete":         false,
	"note verify":         false,
	"note rename":         false,
	"solution list":       false,
	"solution verify":     false,
	"solution enabled":    false,
	"solution applied":    false,
	"solution apply":      false,
	"solution change":     false,
	"solution simulate":   false,
	"solution customise":  false,
	"solution create":     false,
	"solution edit":       false,
	"solution revert":     false,
	"solution show":       false,
	"solution delete":     false,
	"solution rename":     false,
	"staging status":      false,
	"staging enable":      false,
	"staging disable":     false,
	"staging is-enabled":  false,
	"staging list":        false,
	"staging diff":        false,
	"staging analysis":    false,
	"staging release":     false,
	"revert all":          false,
	"lock remove":         false,
	"check":               false,
	"status":              false,
	"version":             false,
	"help":                false,
}

// newDefaultCommandsMap creates a new 'command - action' map
func newDefaultCommandsMap() map[string]bool {
	var newCommandMap = make(map[string]bool, len(defaultCommand))
	for key, value := range defaultCommand {
		newCommandMap[key] = value
	}
	return newCommandMap
}

// supportedRACMap contains the supported 'command - action' combinations
// for the json output
func supportedRACMap() map[string]bool {
	var supportedRAC = newDefaultCommandsMap()

	supportedRAC["daemon status"] = true
	supportedRAC["service status"] = true
	supportedRAC["note list"] = true
	supportedRAC["note enabled"] = true
	supportedRAC["note applied"] = true
	supportedRAC["note verify"] = true
	supportedRAC["solution list"] = true
	supportedRAC["solution verify"] = true
	supportedRAC["solution enabled"] = true
	supportedRAC["solution applied"] = true
	supportedRAC["status"] = true
	supportedRAC["version"] = true

	return supportedRAC
}

// lockCommandsMap contains the supported 'command - action' combinations
// for the saptune locking mechanism
func lockCommandsMap() map[string]bool {
	var lockCommand = newDefaultCommandsMap()

	lockCommand["daemon start"] = true
	lockCommand["daemon stop"] = true
	lockCommand["service apply"] = true
	lockCommand["service start"] = true
	lockCommand["service stop"] = true
	lockCommand["service restart"] = true
	lockCommand["service revert"] = true
	lockCommand["service reload"] = true
	lockCommand["service takeover"] = true
	lockCommand["service enablestart"] = true
	lockCommand["service disablestop"] = true
	lockCommand["note revertall"] = true
	lockCommand["note apply"] = true
	lockCommand["note customise"] = true
	lockCommand["note create"] = true
	lockCommand["note edit"] = true
	lockCommand["note revert"] = true
	lockCommand["note delete"] = true
	lockCommand["note rename"] = true
	lockCommand["solution apply"] = true
	lockCommand["solution change"] = true
	lockCommand["solution customise"] = true
	lockCommand["solution create"] = true
	lockCommand["solution edit"] = true
	lockCommand["solution revert"] = true
	lockCommand["solution delete"] = true
	lockCommand["solution rename"] = true
	lockCommand["staging status"] = true
	lockCommand["staging enable"] = true
	lockCommand["staging disable"] = true
	lockCommand["staging is-enabled"] = true
	lockCommand["staging list"] = true
	lockCommand["staging diff"] = true
	lockCommand["staging analysis"] = true
	lockCommand["staging release"] = true
	lockCommand["revert all"] = true

	return lockCommand
}
