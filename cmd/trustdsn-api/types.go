package main

type Config struct {
	Addr          string
	RepoRoot      string
	LotusBin      string
	MinerHTTPPort int
}

type MinerInfo struct {
	NodeIP         string `json:"node_ip"`
	Index          string `json:"index"`
	StoragePower   string `json:"storage_power"`
	CommittedSpace string `json:"committed_space"`
	UserDataSize   string `json:"user_data_size"`
}

type ProofInfo struct {
	NodeIP                     string `json:"node_ip"`
	ProofType                  string `json:"proof_type"`
	Status                     string `json:"status"`
	Timestamp                  string `json:"timestamp"`
	GenerateDurationSeconds    string `json:"generate_duration_seconds"`
	VerifyDurationMilliseconds string `json:"verify_duration_milliseconds"`
}

type FileListResponse struct {
	Files []string `json:"files"`
}

type MinerListResponse struct {
	Miners []MinerInfo `json:"miners"`
}

type ProofListResponse struct {
	Proofs []ProofInfo `json:"proofs"`
}

type DealRequest struct {
	FileName string `json:"file_name"`
}

type RetrieveRequest struct {
	FileName   string `json:"file_name"`
	OutputName string `json:"output_name"`
}

type CommandResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Output     string `json:"output"`
	OutputPath string `json:"output_path,omitempty"`
}
