package mode

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	testifyassert "github.com/stretchr/testify/assert"
)

func TestMode(t *testing.T) {
	RunSpecs(t, "Mode Suite")
}

// Helper function to reset buildMode after each test that modifies it.
// This is crucial for isolating tests that rely on a global package variable.
func resetBuildMode() {
	buildMode = "" // Reset to default (development)
}

var describeNewModeOperator = Describe("NewModeOperator", func() {
	var assertions *testifyassert.Assertions

	BeforeEach(func() {
		assertions = testifyassert.New(GinkgoT())
	})

	AfterEach(resetBuildMode)

	It("defaults to development mode when buildMode is empty", func() {
		buildMode = ""

		op := NewModeOperator()
		assertions.True(op.IsDevelopmentMode())
		assertions.Equal(DEVELOPMENT, op.GetMode())
	})

	It("uses production mode when buildMode is production", func() {
		buildMode = "production"

		op := NewModeOperator()
		assertions.True(op.IsProductionMode())
		assertions.Equal(PRODUCTION, op.GetMode())
	})

	It("falls back to development mode on invalid buildMode", func() {
		buildMode = "unrecognized"

		op := NewModeOperator()
		assertions.True(op.IsDevelopmentMode())
		assertions.Equal(DEVELOPMENT, op.GetMode())
	})
})

var describeModeOperator = Describe("ModeOperator", func() {
	var assertions *testifyassert.Assertions

	BeforeEach(func() {
		assertions = testifyassert.New(GinkgoT())
	})

	AfterEach(resetBuildMode)

	DescribeTable("mode checks", func(setBuildMode string, isDev bool, isProd bool) {
		buildMode = setBuildMode
		op := NewModeOperator()

		assertions.Equal(isDev, op.IsDevelopmentMode())
		assertions.Equal(isProd, op.IsProductionMode())
	},
		Entry("default (empty)", "", true, false),
		Entry("development", "development", true, false),
		Entry("production", "production", false, true),
		Entry("staging (unrecognized)", "staging", true, false),
	)

	Describe("ExecuteIfModeIsProduction", func() {
		It("executes the callback in production mode", func() {
			buildMode = "production"
			op := NewModeOperator()

			executed := false
			op.ExecuteIfModeIsProduction(func() {
				executed = true
			})

			assertions.True(executed)
		})

		It("skips the callback in development mode", func() {
			buildMode = "development"
			op := NewModeOperator()

			executed := false
			op.ExecuteIfModeIsProduction(func() {
				executed = true
			})

			assertions.False(executed)
		})
	})
})

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
