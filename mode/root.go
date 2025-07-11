// Package mode provides a way to determine the application's operational mode
// (e.g., development, production) at compile time using build flags.
//
// The mode is set using the -ldflags option during the go build command. For example:
//
//	go build -ldflags "-X github.com/yourusername/yourproject/mode.buildMode=production" main.go
//
// The ModeOperator interface provides methods to query the current mode.
package mode

import (
	"fmt"
)

// buildMode is a package-level variable that is set at compile time using -ldflags.
// It determines the application's operational mode.
var buildMode string

const (
	// DEVELOPMENT represents the development operational mode.
	// This is the default mode if no build flag is set, or if an invalid flag is used.
	DEVELOPMENT = "development"

	// PRODUCTION represents the production operational mode.
	PRODUCTION = "production"
)

// ModeOperator defines the interface for querying the application's operational mode.
//
// Implementations of this interface provide methods to determine the current
// operational mode of the application (e.g., development, production). The
// mode is intended to be set at compile time via build flags and should not be
// modifiable at runtime.
type ModeOperator interface {
	// GetMode returns the current operational mode as a string.
	GetMode() string

	// IsDevelopmentMode checks if the application is running in development mode.
	IsDevelopmentMode() bool

	// IsProductionMode checks if the application is running in production mode.
	IsProductionMode() bool

	// ExecuteIfModeIsProduction executes the provided callback function
	// only if the application is running in production mode.
	ExecuteIfModeIsProduction(cb func())
}

// modeOperator is the concrete implementation of the ModeOperator interface.
// Its name is unexported (lowercase) to signal that consumers should
// interact with the ModeOperator interface rather than this concrete type directly.
type modeOperator struct {
	currentMode string
}

// NewModeOperator initializes and returns a new ModeOperator interface.
// The concrete modeOperator struct's current mode is determined by the 'buildMode'
// package variable, which is expected to be set at compile time via -ldflags.
// If 'buildMode' is empty or not one of the predefined constants (PRODUCTION),
// it defaults to DEVELOPMENT mode.
func NewModeOperator() ModeOperator {
	op := &modeOperator{}

	switch buildMode {
	case PRODUCTION:
		op.currentMode = PRODUCTION
	default: // Includes "" (not set) and any unrecognized string
		op.currentMode = DEVELOPMENT
	}
	return op
}

// GetMode returns the current operational mode of the application.
func (o modeOperator) GetMode() string {
	return o.currentMode
}

// IsDevelopmentMode checks if the application is running in development mode.
// This is the default mode if no specific build flag is set or recognized.
func (o modeOperator) IsDevelopmentMode() bool {
	return o.currentMode == DEVELOPMENT
}

// IsProductionMode checks if the application is running in production mode.
func (o modeOperator) IsProductionMode() bool {
	return o.currentMode == PRODUCTION
}

// ExecuteIfModeIsProduction executes the provided callback function
// only if the application is running in production mode.
func (o modeOperator) ExecuteIfModeIsProduction(cb func()) {
	if o.IsProductionMode() {
		cb()
	}
}

func init() {

	if buildMode != "" {

		allowedModes := []string{PRODUCTION, DEVELOPMENT}
		isValid := false
		for _, mode := range allowedModes {
			if mode == buildMode {
				isValid = true
				break
			}
		}

		if !isValid {
			panic(fmt.Sprintf("Invalid build mode: %s it's supposed to be one of %v", buildMode, allowedModes))
		}

	}

}
