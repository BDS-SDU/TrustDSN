package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

type proofInfoRecord struct {
	IP                         string
	ProofType                  string
	Status                     string
	Timestamp                  string
	GenerateDurationSeconds    string
	VerifyDurationMilliseconds string
}

var BftDsnListProofCmd = &cli.Command{
	Name:      "list-proof",
	Usage:     "List proof information from local and remote proof_info.log files",
	ArgsUsage: "",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "port",
			Value: 18080,
			Usage: "HTTP port used by remote miners to expose proof_info.log",
		},
	},
	Action: func(cctx *cli.Context) error {
		var rows []proofInfoRecord

		localIP := readLocalMinerIP()
		localProofs, err := readProofInfoFile("proof_info.log")
		if err == nil {
			rows = append(rows, annotateAndFilterProofRecords(localProofs, localIP)...)
		} else if !os.IsNotExist(err) {
			return xerrors.Errorf("read local proof_info.log: %w", err)
		}

		ips, err := loadAndNormalizeMinerIPs("miner_ip.log")
		if err != nil {
			return xerrors.Errorf("read miner_ip.log: %w", err)
		}

		client := &http.Client{Timeout: 3 * time.Second}

		for _, ip := range ips {
			if ip == "" {
				continue
			}
			if localIP != "" && ip == localIP {
				continue
			}

			proofs, err := fetchRemoteProofInfo(client, ip, cctx.Int("port"))
			if err != nil {
				continue
			}

			rows = append(rows, annotateAndFilterProofRecords(proofs, ip)...)
		}

		if len(rows) == 0 {
			fmt.Println("No proof information found.")
			return nil
		}

		printProofInfoTable(rows)
		return nil
	},
}

func readLocalMinerIP() string {
	info, err := readMinerInfoFile("miner_info.log")
	if err != nil {
		return ""
	}
	return info.IP
}

func readProofInfoFile(path string) ([]proofInfoRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseProofInfo(f)
}

func fetchRemoteProofInfo(client *http.Client, ip string, port int) ([]proofInfoRecord, error) {
	url := fmt.Sprintf("http://%s:%d/proof_info.log", ip, port)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %s", resp.Status)
	}

	return parseProofInfo(resp.Body)
}

func parseProofInfo(r io.Reader) ([]proofInfoRecord, error) {
	records := map[string]*proofInfoRecord{
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

	return []proofInfoRecord{
		*records["storage_proof"],
		*records["window_post"],
	}, nil
}

func printProofInfoTable(rows []proofInfoRecord) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "Node IP\tProof Type\tStatus\tTimestamp\tGenerate (s)\tVerify (ms)")
	for _, row := range rows {
		fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			emptyOrDash(row.IP),
			emptyOrDash(row.ProofType),
			emptyOrDash(row.Status),
			emptyOrDash(row.Timestamp),
			emptyOrDash(row.GenerateDurationSeconds),
			emptyOrDash(row.VerifyDurationMilliseconds),
		)
	}

	_ = tw.Flush()
}

func annotateAndFilterProofRecords(records []proofInfoRecord, ip string) []proofInfoRecord {
	var filtered []proofInfoRecord
	for _, rec := range records {
		if strings.TrimSpace(rec.Status) == "" {
			continue
		}
		rec.IP = ip
		filtered = append(filtered, rec)
	}
	return filtered
}
