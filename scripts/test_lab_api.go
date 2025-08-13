package main

import (
	"fmt"
	"net/http"
)

func main() {
	baseURL := "http://localhost:3000"

	fmt.Println("🧪 Testing Lab API Endpoints")
	fmt.Println("=============================")

	// Test 1: Check if server is running
	fmt.Println("\n1. Testing server connectivity")
	resp, err := http.Get(baseURL + "/")
	if err != nil {
		fmt.Printf("❌ Error connecting to server: %v\n", err)
		return
	}
	fmt.Printf("✅ Server is running (Status: %d)\n", resp.StatusCode)

	// Test 2: Test the lab test endpoint
	fmt.Println("\n2. Testing GET /api/test-lab")
	resp, err = http.Get(baseURL + "/api/test-lab")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Status: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("✅ Test endpoint working")
		}
	}

	// Test 3: Test blood types endpoint
	fmt.Println("\n3. Testing GET /api/lab/blood-types")
	resp, err = http.Get(baseURL + "/api/lab/blood-types")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Status: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("✅ Blood types endpoint working")
		} else {
			fmt.Printf("❌ Unexpected status: %d\n", resp.StatusCode)
		}
	}

	// Test 4: Test blood types by category endpoint
	fmt.Println("\n4. Testing GET /api/lab/blood-types/category/CBC")
	resp, err = http.Get(baseURL + "/api/lab/blood-types/category/CBC")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Status: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("✅ Blood types by category endpoint working")
		} else {
			fmt.Printf("❌ Unexpected status: %d\n", resp.StatusCode)
		}
	}

	// Test 5: Test swab types endpoint
	fmt.Println("\n5. Testing GET /api/lab/swab-types")
	resp, err = http.Get(baseURL + "/api/lab/swab-types")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Status: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("✅ Swab types endpoint working")
		} else {
			fmt.Printf("❌ Unexpected status: %d\n", resp.StatusCode)
		}
	}

	// Test 6: Test urine types endpoint
	fmt.Println("\n6. Testing GET /api/lab/urine-types")
	resp, err = http.Get(baseURL + "/api/lab/urine-types")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Status: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("✅ Urine types endpoint working")
		} else {
			fmt.Printf("❌ Unexpected status: %d\n", resp.StatusCode)
		}
	}

	fmt.Println("\n✅ Lab API testing completed!")
}
