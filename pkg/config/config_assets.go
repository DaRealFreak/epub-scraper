package config

// Assets contains the included assets in each added section
type Assets struct {
	CSS  Asset `yaml:"css"`
	Font Asset `yaml:"fonts"`
}

// Asset contains the path on the host system and after being added the internal path of the epub
type Asset struct {
	HostPath     string `yaml:"path"`
	InternalPath string
}
