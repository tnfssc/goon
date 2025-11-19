package main

import (
	"fmt"
	"log"

	"github.com/tnfssc/goon/pkg/toon"
)

// This demonstrates how the Marshal/Unmarshal functions solve the issue:
// "Without a way to Marshal or unmarshal the given toon string directly into
// the go code like we do with json, this library isn't really useful"

type AppConfig struct {
	AppName     string   `toon:"app_name"`
	Version     string   `toon:"version"`
	Port        int      `toon:"port"`
	EnableDebug bool     `toon:"enable_debug"`
	Endpoints   []string `toon:"endpoints"`
}

func main() {
	fmt.Println("=== Demonstrating Marshal/Unmarshal (like encoding/json) ===\n")

	// Before: You could only work with map[string]interface{}
	fmt.Println("❌ OLD WAY (without Marshal/Unmarshal):")
	fmt.Println("You had to use map[string]interface{} and type assertions:")
	fmt.Println(`
  toonStr := "app_name: MyApp\nversion: 1.0.0"
  data, _ := toon.Decode(toonStr, toon.DecodeOptions{})
  appName := data.(map[string]interface{})["app_name"].(string) // ugh!
`)

	// Now: You can marshal/unmarshal directly with structs!
	fmt.Println("\n✅ NEW WAY (with Marshal/Unmarshal):")
	fmt.Println("Now you can work directly with Go structs, just like encoding/json!\n")

	// Create a config struct
	config := AppConfig{
		AppName:     "MyAwesomeApp",
		Version:     "2.1.0",
		Port:        8080,
		EnableDebug: true,
		Endpoints:   []string{"/api/users", "/api/posts", "/api/comments"},
	}

	fmt.Println("1. Marshal struct to TOON (like json.Marshal):")
	toonBytes, err := toon.Marshal(config, toon.EncodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(toonBytes))

	fmt.Println("\n2. Unmarshal TOON to struct (like json.Unmarshal):")
	var loadedConfig AppConfig
	err = toon.Unmarshal(toonBytes, &loadedConfig, toon.DecodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   Loaded Config:\n")
	fmt.Printf("   - App Name: %s\n", loadedConfig.AppName)
	fmt.Printf("   - Version: %s\n", loadedConfig.Version)
	fmt.Printf("   - Port: %d\n", loadedConfig.Port)
	fmt.Printf("   - Debug: %v\n", loadedConfig.EnableDebug)
	fmt.Printf("   - Endpoints: %v\n", loadedConfig.Endpoints)

	fmt.Println("\n✨ Now the library is as useful as encoding/json! ✨")
}
