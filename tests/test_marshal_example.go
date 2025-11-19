package main

import (
	"fmt"
	"log"

	"github.com/tnfssc/goon/pkg/toon"
)

// User represents a user in the system
type User struct {
	ID       int      `toon:"id"`
	Name     string   `toon:"name"`
	Email    string   `toon:"email"`
	Active   bool     `toon:"active"`
	Tags     []string `toon:"tags"`
	Profile  Profile  `toon:"profile"`
	ignored  string   // unexported fields are ignored
	Excluded string   `toon:"-"` // fields with - tag are ignored
}

// Profile represents a user's profile information
type Profile struct {
	Bio       string `toon:"bio"`
	Website   string `toon:"website"`
	AvatarURL string `toon:"avatar_url"`
}

func main() {
	fmt.Println("=== Marshal/Unmarshal Example ===\n")

	// Create a User struct
	user := User{
		ID:     123,
		Name:   "Alice Johnson",
		Email:  "alice@example.com",
		Active: true,
		Tags:   []string{"developer", "golang", "toon"},
		Profile: Profile{
			Bio:       "Software engineer passionate about Go",
			Website:   "https://alice.dev",
			AvatarURL: "https://example.com/avatar.jpg",
		},
		ignored:  "this will not be marshaled",
		Excluded: "this will not be marshaled either",
	}

	// Marshal to TOON
	fmt.Println("1. Marshaling struct to TOON:")
	toonData, err := toon.Marshal(user, toon.EncodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatalf("Marshal failed: %v", err)
	}
	fmt.Println(string(toonData))
	fmt.Println()

	// Unmarshal back to struct
	fmt.Println("2. Unmarshaling TOON back to struct:")
	var decodedUser User
	err = toon.Unmarshal(toonData, &decodedUser, toon.DecodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatalf("Unmarshal failed: %v", err)
	}

	fmt.Printf("ID: %d\n", decodedUser.ID)
	fmt.Printf("Name: %s\n", decodedUser.Name)
	fmt.Printf("Email: %s\n", decodedUser.Email)
	fmt.Printf("Active: %v\n", decodedUser.Active)
	fmt.Printf("Tags: %v\n", decodedUser.Tags)
	fmt.Printf("Profile.Bio: %s\n", decodedUser.Profile.Bio)
	fmt.Printf("Profile.Website: %s\n", decodedUser.Profile.Website)
	fmt.Printf("Profile.AvatarURL: %s\n", decodedUser.Profile.AvatarURL)
	fmt.Println()

	// Also works with maps
	fmt.Println("3. Marshal/Unmarshal with map:")
	data := map[string]interface{}{
		"title":   "TOON Format",
		"version": 1.0,
		"features": []string{
			"human-readable",
			"indentation-based",
			"no-braces",
		},
	}

	toonMap, err := toon.Marshal(data, toon.EncodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatalf("Marshal map failed: %v", err)
	}
	fmt.Println(string(toonMap))
	fmt.Println()

	var decodedMap map[string]interface{}
	err = toon.Unmarshal(toonMap, &decodedMap, toon.DecodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatalf("Unmarshal map failed: %v", err)
	}
	fmt.Printf("Decoded map: %+v\n", decodedMap)
	fmt.Println()

	fmt.Println("✓ All examples completed successfully!")
}
