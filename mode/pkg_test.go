package mode

import (
	"fmt"
	"testing"
)

// Helper function to reset buildMode after each test that modifies it.
// This is crucial for isolating tests that rely on a global package variable.
func resetBuildMode() {
	buildMode = "" // Reset to default (development)
}

func TestNewModeOperator_Default(t *testing.T) {
	// Ensure buildMode is not set for this test
	defer resetBuildMode()
	buildMode = ""

	op := NewModeOperator()
	if !op.IsDevelopmentMode() {
		t.Errorf("NewModeOperator() with empty buildMode: expected IsDevelopmentMode() to be true, got %t", op.IsDevelopmentMode())
	}
	if op.GetMode() != DEVELOPMENT {
		t.Errorf("NewModeOperator() with empty buildMode: expected mode %q, got %q", DEVELOPMENT, op.GetMode())
	}
}

func TestNewModeOperator_Production(t *testing.T) {
	defer resetBuildMode()
	buildMode = "production"

	op := NewModeOperator()
	if !op.IsProductionMode() {
		t.Errorf("NewModeOperator() with buildMode 'production': expected IsProductionMode() to be true, got %t", op.IsProductionMode())
	}
	if op.GetMode() != PRODUCTION {
		t.Errorf("NewModeOperator() with buildMode 'production': expected mode %q, got %q", PRODUCTION, op.GetMode())
	}
}

func TestNewModeOperator_Invalid(t *testing.T) {
	defer resetBuildMode()
	buildMode = "unrecognized" // An invalid mode string

	op := NewModeOperator()
	if !op.IsDevelopmentMode() {
		t.Errorf("NewModeOperator() with invalid buildMode: expected IsDevelopmentMode() to be true, got %t", op.IsDevelopmentMode())
	}
	if op.GetMode() != DEVELOPMENT {
		t.Errorf("NewModeOperator() with invalid buildMode: expected mode %q, got %q", DEVELOPMENT, op.GetMode())
	}
}

func TestModeOperator_ModeChecks(t *testing.T) {
	tests := []struct {
		name         string
		setBuildMode string
		isDev        bool
		isProd       bool
	}{
		{"Default (empty)", "", true, false},
		{"Development", "development", true, false},
		{"Production", "production", false, true},
		{"Staging (unrecognized)", "staging", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetBuildMode()
			buildMode = tt.setBuildMode
			op := NewModeOperator()

			if op.IsDevelopmentMode() != tt.isDev {
				t.Errorf("IsDevelopmentMode() for %q: expected %t, got %t", tt.setBuildMode, tt.isDev, op.IsDevelopmentMode())
			}
			if op.IsProductionMode() != tt.isProd {
				t.Errorf("IsProductionMode() for %q: expected %t, got %t", tt.setBuildMode, tt.isProd, op.IsProductionMode())
			}
		})
	}
}

func TestModeOperator_ExecuteIfModeIsProduction(t *testing.T) {
	var executed bool

	// Case 1: Production mode - callback should execute
	t.Run("ProductionMode", func(t *testing.T) {
		defer resetBuildMode()
		buildMode = "production"
		op := NewModeOperator()

		executed = false
		op.ExecuteIfModeIsProduction(func() {
			executed = true
		})
		if !executed {
			t.Errorf("Callback not executed in production mode")
		}
	})

	// Case 2: Development mode - callback should NOT execute
	t.Run("DevelopmentMode", func(t *testing.T) {
		defer resetBuildMode()
		buildMode = "development"
		op := NewModeOperator()

		executed = false
		op.ExecuteIfModeIsProduction(func() {
			executed = true
		})
		if executed {
			t.Errorf("Callback executed in development mode")
		}
	})
}

func ExampleNewModeOperator() {
	// To run this example with a specific mode, use -ldflags:
	//  go run -ldflags "-X github.com/yourusername/yourproject/mode.buildMode=production" your_main_file.go
	// Or for documentation generation:
	//  go doc -all github.com/yourusername/yourproject/mode

	// Simulate setting buildMode for the example's execution environment
	// In a real application, this would be set by the build command.
	originalBuildMode := buildMode
	defer func() { buildMode = originalBuildMode }() // Restore original value
	buildMode = "production"                         // Simulate being compiled for production

	op := NewModeOperator()
	fmt.Printf("Is Production Mode: %t\n", op.IsProductionMode())

	buildMode = "development" // Simulate being compiled for development
	op = NewModeOperator()    // Re-initialize to pick up new buildMode
	fmt.Printf("Is Development Mode: %t\n", op.IsDevelopmentMode())

	// Output:
	// Is Production Mode: true
	// Is Development Mode: true
}

func ExampleModeOperator_ExecuteIfModeIsProduction() {
	// Simulate setting buildMode for the example's execution environment
	originalBuildMode := buildMode
	defer func() { buildMode = originalBuildMode }() // Restore original value

	// Scenario 1: Simulate Development build
	buildMode = "development"
	opDev := NewModeOperator()
	opDev.ExecuteIfModeIsProduction(func() {
		fmt.Println("This should not be printed in development.")
	})

	// Scenario 2: Simulate Production build
	buildMode = "production"
	opProd := NewModeOperator()
	opProd.ExecuteIfModeIsProduction(func() {
		fmt.Println("Performing production-only task.")
	})

	// Output:
	// Performing production-only task.
}

func ExampleModeOperator_is_modes() {
	// Simulate setting buildMode for the example's execution environment
	originalBuildMode := buildMode
	defer func() { buildMode = originalBuildMode }() // Restore original value

	// Test Default/Development Mode
	buildMode = "" // Default
	op := NewModeOperator()
	fmt.Printf("Default - Is Dev: %t, Is Prod: %t\n", op.IsDevelopmentMode(), op.IsProductionMode())

	// Test Production Mode
	buildMode = "production"
	op = NewModeOperator()
	fmt.Printf("Production - Is Dev: %t, Is Prod: %t\n", op.IsDevelopmentMode(), op.IsProductionMode())

	// Output:
	// Default - Is Dev: true, Is Prod: false
	// Production - Is Dev: false, Is Prod: true
}
