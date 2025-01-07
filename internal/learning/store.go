package learning

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Interaction represents a single user interaction with the agent
type Interaction struct {
	Query           string             `json:"query"`
	Response        string             `json:"response"`
	ChunksUsed      []string           `json:"chunks_used"` // IDs of chunks used
	Successful      bool               `json:"successful"`  // Whether the response was helpful
	Timestamp       time.Time          `json:"timestamp"`
	Context         map[string]string  `json:"context"`          // Additional context (command, namespace, etc.)
	FeedbackApplied map[string]float32 `json:"feedback_applied"` // Adjustments made based on this interaction
}

// LearningStore manages the agent's learning data
type LearningStore struct {
	Interactions   []Interaction       `json:"interactions"`
	ChunkScores    map[string]float32  `json:"chunk_scores"`    // Learned relevance adjustments per chunk
	QueryPatterns  map[string][]string `json:"query_patterns"`  // Mapping of canonical forms to variations
	CommandAliases map[string][]string `json:"command_aliases"` // Alternative ways commands are referenced
	path           string              // Path to the learning data file
}

// New creates a new learning store
func New(dataPath string) (*LearningStore, error) {
	store := &LearningStore{
		Interactions:   make([]Interaction, 0),
		ChunkScores:    make(map[string]float32),
		QueryPatterns:  make(map[string][]string),
		CommandAliases: make(map[string][]string),
		path:           dataPath,
	}

	// Load existing data if available
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// RecordInteraction records a new interaction and updates learning data
func (s *LearningStore) RecordInteraction(interaction Interaction) error {
	s.Interactions = append(s.Interactions, interaction)

	// Update chunk scores based on success
	multiplier := float32(1.0)
	if interaction.Successful {
		multiplier = 1.1 // Boost successful chunks
	} else {
		multiplier = 0.9 // Reduce score for unsuccessful chunks
	}

	for _, chunkID := range interaction.ChunksUsed {
		currentScore := s.ChunkScores[chunkID]
		s.ChunkScores[chunkID] = currentScore*0.9 + multiplier*0.1 // Exponential moving average
	}

	return s.save()
}

// GetChunkScore returns the learned relevance adjustment for a chunk
func (s *LearningStore) GetChunkScore(chunkID string) float32 {
	score, ok := s.ChunkScores[chunkID]
	if !ok {
		return 1.0 // Default score
	}
	return score
}

// AddQueryPattern records a new way of asking about something
func (s *LearningStore) AddQueryPattern(canonicalForm, variation string) error {
	patterns := s.QueryPatterns[canonicalForm]
	for _, p := range patterns {
		if p == variation {
			return nil // Already exists
		}
	}
	s.QueryPatterns[canonicalForm] = append(patterns, variation)
	return s.save()
}

// load reads the learning data from disk
func (s *LearningStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}

// save persists the learning data to disk
func (s *LearningStore) save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}
