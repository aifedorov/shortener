package repository

type BatchURLInput struct {
	CID         string
	OriginalURL string
}

type BatchURLOutput struct {
	CID      string
	ShortURL string
}

type URLOutput struct {
	ShortURL    string
	OriginalURL string
}
