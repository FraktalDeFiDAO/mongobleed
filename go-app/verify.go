// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("MongoBleed Go Implementation - Build Verification")
	fmt.Println("================================================")
	fmt.Println()

	// Check if Go is installed
	fmt.Print("Checking Go installation... ")
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		fmt.Println("NOT FOUND")
		fmt.Println("Error: Go is not installed. Please install Go 1.18 or higher.")
		fmt.Println("Visit: https://golang.org/dl/")
		os.Exit(1)
	}
	fmt.Println("OK")

	// Check Go version
	fmt.Print("Checking Go version... ")
	versionCmd := exec.Command("go", "version")
	versionOutput, _ := versionCmd.Output()
	fmt.Printf("%s", versionOutput)

	// Try to build
	fmt.Print("Testing build... ")
	buildCmd := exec.Command("go", "build", "-o", "mongobleed-test", "mongobleed.go")
	if err := buildCmd.Run(); err != nil {
		fmt.Println("FAILED")
		fmt.Printf("Build error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")

	// Clean up test binary
	os.Remove("mongobleed-test")

	fmt.Println()
	fmt.Println("✓ All checks passed!")
	fmt.Println("✓ Ready to build with: make build")
	fmt.Println("✓ Or directly: go build -o mongobleed mongobleed.go")
	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Println("  ./mongobleed --host localhost")
	fmt.Println("  ./mongobleed --host 192.168.1.100 --max-offset 50000")
	fmt.Println("  ./examples.sh")
}
