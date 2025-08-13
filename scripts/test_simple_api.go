package main

import (
	"fmt"
	"net/http"
)

func main() {
	baseURL := "http://localhost:3000"

	fmt.Println("🧪 Simple API Test")
	fmt.Println("==================")

	// Test blood types endpoint
	fmt.Println("\nTesting /api/lab/blood-types:")
	resp, err := http.Get(baseURL + "/api/lab/blood-types")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Status: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("✅ Route is working!")
		} else {
			fmt.Printf("❌ Unexpected status: %d\n", resp.StatusCode)
		}
		resp.Body.Close()
	}

	fmt.Println("\n✅ Test completed!")
}
