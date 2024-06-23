package fileuploader

type ChunkMeta struct {
	FileName string `json:"file_name"`
	MD5Hash  string `json:"md5_hash"`
	Index    int    `json:"index"`
}

type Config struct {
	ChunkSize int
	ServerURL string
}

type DefaultFileChunker struct {
	chunkSize int
}

func (d *DefaultFileChunker) SetChunkSize(chunkSize int) {
	d.chunkSize = chunkSize
}

type DefaultUploader struct {
	serverURL string
}

func (d *DefaultUploader) SetServerURL(serverURL string) {
	d.serverURL = serverURL
}

type DefaultMetadataManager struct{}

type FileChunker interface {
	ChunkFile(filePath string) ([]ChunkMeta, error)
	ChunkLargeFile(filePath string) ([]ChunkMeta, error)
}

type Uploader interface {
	UploadChunk(chunk ChunkMeta) error
}

type MetadataManager interface {
	LoadMetadata(filePath string) (map[string]ChunkMeta, error)
	SaveMetadata(filePath string, metadata map[string]ChunkMeta) error
}
