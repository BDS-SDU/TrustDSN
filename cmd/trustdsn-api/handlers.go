package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMiners(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	miners, err := s.collectMinerInfos()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MinerListResponse{Miners: miners})
}

func (s *Server) handleProofs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	proofs, err := s.collectProofInfos()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ProofListResponse{Proofs: proofs})
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	files, err := loadAndNormalizeFileNames(filepath.Join(s.cfg.RepoRoot, "filenames.log"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, FileListResponse{Files: files})
}

func (s *Server) handleDeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req DealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	fileName, err := sanitizeFileName(req.FileName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	inputPath := filepath.Join(s.cfg.RepoRoot, fileName)
	if _, err := os.Stat(inputPath); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("file not found: %s", inputPath))
		return
	}

	output, err := s.runLotusCommand(
		r.Context(),
		"bftdsn", "deal",
		"-k", "3",
		"-m", "1",
		"--keep-chunks",
		inputPath,
	)
	if err != nil {
		writeJSON(w, http.StatusOK, CommandResponse{
			Success: false,
			Message: "deal failed",
			Output:  output,
		})
		return
	}

	writeJSON(w, http.StatusOK, CommandResponse{
		Success: true,
		Message: "deal finished",
		Output:  output,
	})
}

func (s *Server) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req RetrieveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	fileName, err := sanitizeFileName(req.FileName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	outputName := req.OutputName
	if strings.TrimSpace(outputName) == "" {
		outputName = fileName
	}

	outputName, err = sanitizeFileName(outputName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	outputPath := filepath.Join(s.cfg.RepoRoot, outputName)

	output, err := s.runLotusCommand(
		r.Context(),
		"bftdsn", "retrieve",
		"-k", "3",
		"-m", "1",
		fileName,
		outputPath,
	)
	if err != nil {
		writeJSON(w, http.StatusOK, CommandResponse{
			Success:    false,
			Message:    "retrieve failed",
			Output:     output,
			OutputPath: outputPath,
		})
		return
	}

	writeJSON(w, http.StatusOK, CommandResponse{
		Success:    true,
		Message:    "retrieve finished",
		Output:     output,
		OutputPath: outputPath,
	})
}

func (s *Server) collectMinerInfos() ([]MinerInfo, error) {
	var out []MinerInfo

	localInfo, err := readMinerInfoFile(filepath.Join(s.cfg.RepoRoot, "miner_info.log"))
	if err == nil {
		out = append(out, localInfo)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	localIP := localInfo.NodeIP

	ips, err := loadAndNormalizeIPs(filepath.Join(s.cfg.RepoRoot, "miner_ip.log"))
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		if ip == "" {
			continue
		}
		if localIP != "" && ip == localIP {
			continue
		}

		info, err := s.fetchRemoteMinerInfo(ip)
		if err != nil {
			continue
		}

		out = append(out, info)
	}

	return out, nil
}

func (s *Server) collectProofInfos() ([]ProofInfo, error) {
	var out []ProofInfo

	localIP := ""
	localMinerInfo, err := readMinerInfoFile(filepath.Join(s.cfg.RepoRoot, "miner_info.log"))
	if err == nil {
		localIP = localMinerInfo.NodeIP
	}

	localProofs, err := readProofInfoFile(filepath.Join(s.cfg.RepoRoot, "proof_info.log"))
	if err == nil {
		for _, p := range localProofs {
			if strings.TrimSpace(p.Status) == "" {
				continue
			}
			p.NodeIP = localIP
			out = append(out, p)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	ips, err := loadAndNormalizeIPs(filepath.Join(s.cfg.RepoRoot, "miner_ip.log"))
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		if ip == "" {
			continue
		}
		if localIP != "" && ip == localIP {
			continue
		}

		proofs, err := s.fetchRemoteProofInfo(ip)
		if err != nil {
			continue
		}

		for _, p := range proofs {
			if strings.TrimSpace(p.Status) == "" {
				continue
			}
			p.NodeIP = ip
			out = append(out, p)
		}
	}

	return out, nil
}

func (s *Server) fetchRemoteMinerInfo(ip string) (MinerInfo, error) {
	url := fmt.Sprintf("http://%s:%d/miner_info.log", ip, s.cfg.MinerHTTPPort)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return MinerInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MinerInfo{}, fmt.Errorf("http status %s", resp.Status)
	}

	info, err := parseMinerInfo(resp.Body)
	if err != nil {
		return MinerInfo{}, err
	}

	if info.NodeIP == "" {
		info.NodeIP = ip
	}

	return info, nil
}

func (s *Server) fetchRemoteProofInfo(ip string) ([]ProofInfo, error) {
	url := fmt.Sprintf("http://%s:%d/proof_info.log", ip, s.cfg.MinerHTTPPort)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %s", resp.Status)
	}

	return parseProofInfo(resp.Body)
}

func readMinerInfoFile(path string) (MinerInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return MinerInfo{}, err
	}
	defer f.Close()

	return parseMinerInfo(f)
}

func parseMinerInfo(r io.Reader) (MinerInfo, error) {
	var info MinerInfo

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "ip":
			info.NodeIP = val
		case "miner":
			info.Index = val
		case "quality_adjusted_power":
			info.StoragePower = val
		case "committed_space":
			info.CommittedSpace = val
		case "user_data_size":
			info.UserDataSize = val
		}
	}

	if err := scanner.Err(); err != nil {
		return MinerInfo{}, err
	}

	return info, nil
}

func readProofInfoFile(path string) ([]ProofInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseProofInfo(f)
}

func parseProofInfo(r io.Reader) ([]ProofInfo, error) {
	records := map[string]*ProofInfo{
		"storage_proof": {
			ProofType:                  "storage_proof",
			GenerateDurationSeconds:    "0",
			VerifyDurationMilliseconds: "0",
		},
		"window_post": {
			ProofType:                  "window_post",
			GenerateDurationSeconds:    "0",
			VerifyDurationMilliseconds: "0",
		},
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "storage_proof_type":
			records["storage_proof"].ProofType = val
		case "storage_proof_status":
			records["storage_proof"].Status = val
		case "storage_proof_timestamp":
			records["storage_proof"].Timestamp = val
		case "storage_proof_generate_duration_seconds":
			records["storage_proof"].GenerateDurationSeconds = val
		case "storage_proof_verify_duration_milliseconds":
			records["storage_proof"].VerifyDurationMilliseconds = val
		case "storage_proof_verify_duration_seconds":
			records["storage_proof"].VerifyDurationMilliseconds = val
		case "window_post_type":
			records["window_post"].ProofType = val
		case "window_post_status":
			records["window_post"].Status = val
		case "window_post_timestamp":
			records["window_post"].Timestamp = val
		case "window_post_generate_duration_seconds":
			records["window_post"].GenerateDurationSeconds = val
		case "window_post_verify_duration_milliseconds":
			records["window_post"].VerifyDurationMilliseconds = val
		case "window_post_verify_duration_seconds":
			records["window_post"].VerifyDurationMilliseconds = val
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []ProofInfo{
		*records["storage_proof"],
		*records["window_post"],
	}, nil
}

func loadAndNormalizeIPs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var ordered []string
	seen := make(map[string]struct{})

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip == "" {
			continue
		}
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		ordered = append(ordered, ip)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	content := strings.Join(ordered, "\n")
	if content != "" {
		content += "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}

	return ordered, nil
}

func loadAndNormalizeFileNames(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var ordered []string
	seen := make(map[string]struct{})

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		ordered = append(ordered, name)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	content := strings.Join(ordered, "\n")
	if content != "" {
		content += "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}

	return ordered, nil
}

func sanitizeFileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("file_name is required")
	}
	if filepath.Base(name) != name {
		return "", fmt.Errorf("only plain file names are allowed")
	}
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid file name")
	}
	return name, nil
}

func (s *Server) runLotusCommand(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.cfg.LotusBin, args...)
	cmd.Dir = s.cfg.RepoRoot

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
