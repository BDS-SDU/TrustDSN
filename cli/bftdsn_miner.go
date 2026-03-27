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

type minerInfoRecord struct {
	IP                   string
	Miner                string
	QualityAdjustedPower string
	CommittedSpace       string
	UserDataSize         string
}

var BftDsnListMinerCmd = &cli.Command{
	Name:      "list-miner",
	Usage:     "List miner information from local and remote miner_info.log files",
	ArgsUsage: "",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "port",
			Value: 18080,
			Usage: "HTTP port used by remote miners to expose miner_info.log",
		},
	},
	Action: func(cctx *cli.Context) error {
		var rows []minerInfoRecord

		localInfo, err := readMinerInfoFile("miner_info.log")
		if err == nil {
			rows = append(rows, localInfo)
		} else if !os.IsNotExist(err) {
			return xerrors.Errorf("read local miner_info.log: %w", err)
		}

		ips, err := loadAndNormalizeMinerIPs("miner_ip.log")
		if err != nil {
			return xerrors.Errorf("read miner_ip.log: %w", err)
		}

		localIP := ""
		if localInfo.IP != "" {
			localIP = localInfo.IP
		}

		client := &http.Client{
			Timeout: 3 * time.Second,
		}

		for _, ip := range ips {
			if ip == "" {
				continue
			}
			if ip == localIP {
				continue
			}

			info, err := fetchRemoteMinerInfo(client, ip, cctx.Int("port"))
			if err != nil {
				continue
			}

			rows = append(rows, info)
		}

		if len(rows) == 0 {
			fmt.Println("No miner information found.")
			return nil
		}

		printMinerInfoTable(rows)
		return nil
	},
}

func readMinerInfoFile(path string) (minerInfoRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return minerInfoRecord{}, err
	}
	defer f.Close()

	return parseMinerInfo(f)
}

func fetchRemoteMinerInfo(client *http.Client, ip string, port int) (minerInfoRecord, error) {
	url := fmt.Sprintf("http://%s:%d/miner_info.log", ip, port)

	resp, err := client.Get(url)
	if err != nil {
		return minerInfoRecord{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return minerInfoRecord{}, fmt.Errorf("http status %s", resp.Status)
	}

	info, err := parseMinerInfo(resp.Body)
	if err != nil {
		return minerInfoRecord{}, err
	}

	if info.IP == "" {
		info.IP = ip
	}

	return info, nil
}

func parseMinerInfo(r io.Reader) (minerInfoRecord, error) {
	var info minerInfoRecord

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
			info.IP = val
		case "miner":
			info.Miner = val
		case "quality_adjusted_power":
			info.QualityAdjustedPower = val
		case "committed_space":
			info.CommittedSpace = val
		case "user_data_size":
			info.UserDataSize = val
		}
	}

	if err := scanner.Err(); err != nil {
		return minerInfoRecord{}, err
	}

	return info, nil
}

func loadAndNormalizeMinerIPs(path string) ([]string, error) {
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

func printMinerInfoTable(rows []minerInfoRecord) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "Node IP\tIndex\tStorage Power\tCommitted Space\tUser Data Size")
	for _, row := range rows {
		fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%s\n",
			emptyOrDash(row.IP),
			emptyOrDash(row.Miner),
			emptyOrDash(row.QualityAdjustedPower),
			emptyOrDash(row.CommittedSpace),
			emptyOrDash(row.UserDataSize),
		)
	}

	_ = tw.Flush()
}

func emptyOrDash(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}
