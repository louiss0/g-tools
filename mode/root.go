package mode

// buildMode is a package-level variable that can be set at compile time using -ldflags.
// Example: go build -ldflags "-X louiss0/github.com/g-tools/mode.buildMode=production"
var buildMode string

// mode is an unexported type alias for string, used to define distinct operational modes.

const (
	// DEVELOPMENT represents the development operational mode.
	// This is the default mode if no specific build flag is set or if the flag is unrecognized.
	DEVELOPMENT = "development"
	// PRODUCTION represents the production operational mode.
	PRODUCTION = "production"
)

// ModeOperator defines the interface for querying the application's operational mode.
// The mode is intended to be set at compile time via build flags and should not be
// modifiable at runtime by external users of this package.
type ModeOperator interface {
	// GetMode returns the current operational mode as a string.
	GetMode() string
	// IsDevelopmentMode checks if the application is running in development mode.
	// This is the default mode if no specific build flag is set or recognized.
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
// If 'buildMode' is empty or not one of the predefined constants (PRODUCTION, DEBUG),
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
