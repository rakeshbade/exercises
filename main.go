package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Data structures
type Exercise struct {
	Category          string   `json:"category"`
	Description       string   `json:"description"`
	Equipment         []string `json:"equipment"`
	Instructions      []string `json:"instructions"`
	Name              string   `json:"name"`
	PrimaryMuscles    []string `json:"primary_muscles"`
	SecondaryMuscles  []string `json:"secondary_muscles"`
	Video             string   `json:"video"`
	MuscleGroup       string   `json:"muscle_group"`
}

type ExercisesData struct {
	Categories   []string   `json:"categories"`
	Equipment    []string   `json:"equipment"`
	MuscleGroups []string   `json:"muscle_groups"`
	Exercises    []Exercise `json:"exercises"`
}

// Cache structure
type CacheEntry struct {
	Data      []Exercise
	ExpiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	ttl     time.Duration
}

var (
	exercisesData ExercisesData
	cache         *Cache
)

func init() {
	// Initialize cache with 5-minute TTL
	cache = &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     5 * time.Minute,
	}
}

// LoadExercises loads the JSON data from file
func LoadExercises() error {
	data, err := ioutil.ReadFile("exercises.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &exercisesData)
}

// CacheKey generates a unique cache key from query parameters
func CacheKey(muscles []string, equipment, category string) string {
	key := strings.Join(muscles, ",") + "|" + equipment
	if category != "" {
		key += "|" + category
	}
	return key
}

// Get retrieves from cache if valid
func (c *Cache) Get(key string) ([]Exercise, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Data, true
}

// Set stores in cache
func (c *Cache) Set(key string, data []Exercise) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Clean up expired entries periodically
func (c *Cache) CleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// FilterExercises filters exercises based on query parameters
func FilterExercises(muscles []string, equipment, category string) []Exercise {
	// Normalize muscle names to lowercase
	normalizedMuscles := make([]string, len(muscles))
	for i, m := range muscles {
		normalizedMuscles[i] = strings.ToLower(m)
	}

	equipment = strings.ToLower(equipment)
	if equipment == "" {
		equipment = "none"
	}

	var results []Exercise

	for _, exercise := range exercisesData.Exercises {
		// Check equipment filter
		equipmentMatch := false
		for _, eq := range exercise.Equipment {
			if strings.ToLower(eq) == equipment {
				equipmentMatch = true
				break
			}
		}
		if !equipmentMatch {
			continue
		}

		// Check muscle group filter - must match at least one of the required muscle groups
		muscleMatch := false
		for _, reqMuscle := range normalizedMuscles {
			if strings.ToLower(exercise.MuscleGroup) == reqMuscle {
				muscleMatch = true
				break
			}
		}
		if !muscleMatch {
			continue
		}

		// Check category filter if provided
		if category != "" && strings.ToLower(exercise.Category) != strings.ToLower(category) {
			continue
		}

		results = append(results, exercise)
	}

	return results
}

// GetExercisesHandler handles the /exercises endpoint
func GetExercisesHandler(c *gin.Context) {
	// Parse query parameters
	muscleParam := c.Query("muscle_group")
	equipment := c.DefaultQuery("equipment", "none")
	category := c.Query("category")

	// Validate required parameters
	if muscleParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "muscle_group parameter is required",
		})
		return
	}

	// Split multiple muscle groups (comma or space separated)
	muscles := strings.Split(muscleParam, ",")
	for i := range muscles {
		muscles[i] = strings.TrimSpace(muscles[i])
	}

	// Generate cache key
	cacheKey := CacheKey(muscles, equipment, category)

	// Check cache first
	if cached, ok := cache.Get(cacheKey); ok {
		c.JSON(http.StatusOK, gin.H{
			"data":      cached,
			"count":     len(cached),
			"cached":    true,
			"timestamp": time.Now(),
		})
		return
	}

	// Filter exercises
	results := FilterExercises(muscles, equipment, category)

	// Store in cache
	cache.Set(cacheKey, results)

	c.JSON(http.StatusOK, gin.H{
		"data":      results,
		"count":     len(results),
		"cached":    false,
		"timestamp": time.Now(),
	})
}

// GetStatsHandler returns available filters
func GetStatsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"categories":   exercisesData.Categories,
		"equipment":    exercisesData.Equipment,
		"muscle_groups": exercisesData.MuscleGroups,
		"total_exercises": len(exercisesData.Exercises),
	})
}

// HealthCheckHandler for health checks
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now(),
	})
}

func main() {
	// Load exercises data
	if err := LoadExercises(); err != nil {
		log.Fatalf("Failed to load exercises: %v", err)
	}
	log.Printf("Loaded %d exercises", len(exercisesData.Exercises))

	// Start cache cleanup routine
	go cache.CleanupExpired()

	// Setup Gin router
	router := gin.Default()

	// Routes
	router.GET("/health", HealthCheckHandler)
	router.GET("/exercises", GetExercisesHandler)
	router.GET("/stats", GetStatsHandler)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
