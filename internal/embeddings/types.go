package embeddings

// Chunk represents a piece of documentation with its embedding
type Chunk struct {
	Content   string    // The actual text content
	Embedding []float32 // The vector embedding of the content
	Metadata  Metadata  // Additional information about the chunk
}

// Metadata contains information about where the chunk came from
type Metadata struct {
	Source    string      // The source file
	StartLine int         // Starting line number in source
	EndLine   int         // Ending line number in source
	Command   string      // Related command (e.g., "logs", "pods")
	Topic     string      // Section topic (e.g., "Usage", "Examples")
	Type      SectionType // Type of section (usage, example, flags, etc.)
	IsTable   bool        // Whether this chunk contains a table
	IsCode    bool        // Whether this chunk contains code
	TableCols []string    // Column headers if this is a table
}

// SectionType represents the type of content in a section
type SectionType string

const (
	SectionTypeUsage    SectionType = "usage"
	SectionTypeExample  SectionType = "example"
	SectionTypeFlags    SectionType = "flags"
	SectionTypeCommand  SectionType = "command"
	SectionTypeOverview SectionType = "overview"
)

// EmbeddingStore defines the interface for storing and retrieving embeddings
type EmbeddingStore interface {
	// Store saves a chunk and its embedding
	Store(chunk *Chunk) error

	// Search finds the most relevant chunks for a query
	Search(query string, limit int) ([]*Chunk, error)

	// Load initializes the store from a file
	Load(path string) error

	// Save persists the store to a file
	Save(path string) error
}

// Processor handles document processing and chunking
type Processor interface {
	// Process takes a document and returns chunks
	Process(content string, metadata Metadata) ([]*Chunk, error)
}

// Generator handles creating embeddings for text
type Generator interface {
	// Generate creates an embedding for the given text
	Generate(text string) ([]float32, error)

	// GenerateBatch creates embeddings for multiple texts
	GenerateBatch(texts []string) ([][]float32, error)
}
