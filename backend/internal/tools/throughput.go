package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// Iperf3Tool wraps the iperf3 command
type Iperf3Tool struct {
	*BaseTool
}

func NewIperf3Tool() *Iperf3Tool {
	return &Iperf3Tool{
		BaseTool: NewBaseTool(
			"iperf3",
			"iperf3",
			CategoryThroughput,
			"apt-get install iperf3",
			"Network throughput, jitter, and packet loss testing",
		),
	}
}

type Iperf3Stream struct {
	Socket        int     `json:"socket"`
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
	Seconds       float64 `json:"seconds"`
	Bytes         int64   `json:"bytes"`
	BitsPerSecond float64 `json:"bits_per_second"`
	Retransmits   int     `json:"retransmits,omitempty"`
	SndCwnd       int     `json:"snd_cwnd,omitempty"`
	RTT           int     `json:"rtt,omitempty"`
	RTTVar        int     `json:"rttvar,omitempty"`
}

type Iperf3Result struct {
	Start struct {
		Connected     []map[string]interface{} `json:"connected"`
		Version       string                   `json:"version"`
		SystemInfo    string                   `json:"system_info"`
		Timestamp     map[string]interface{}   `json:"timestamp"`
		ConnectingTo  map[string]interface{}   `json:"connecting_to"`
		TestStart     map[string]interface{}   `json:"test_start"`
	} `json:"start"`
	Intervals []struct {
		Streams []Iperf3Stream `json:"streams"`
		Sum     Iperf3Stream   `json:"sum"`
	} `json:"intervals"`
	End struct {
		Streams []struct {
			Sender   Iperf3Stream `json:"sender"`
			Receiver Iperf3Stream `json:"receiver"`
		} `json:"streams"`
		SumSent     Iperf3Stream `json:"sum_sent"`
		SumReceived Iperf3Stream `json:"sum_received"`
		CPUUtilization struct {
			HostTotal    float64 `json:"host_total"`
			HostUser     float64 `json:"host_user"`
			HostSystem   float64 `json:"host_system"`
			RemoteTotal  float64 `json:"remote_total"`
			RemoteUser   float64 `json:"remote_user"`
			RemoteSystem float64 `json:"remote_system"`
		} `json:"cpu_utilization_percent"`
	} `json:"end"`
}

func (t *Iperf3Tool) ParseOutput(output string) (interface{}, error) {
	var result Iperf3Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse iperf3 JSON output: %w", err)
	}
	return &result, nil
}

// Iperf3Client runs iperf3 in client mode
func (t *Iperf3Tool) Iperf3Client(ctx context.Context, server string, port int, duration int, protocol string, reverse bool) (*Iperf3Result, error) {
	args := []string{
		"-c", server,
		"-p", strconv.Itoa(port),
		"-t", strconv.Itoa(duration),
		"-J", // JSON output
	}

	if protocol == "udp" {
		args = append(args, "-u")
	}

	if reverse {
		args = append(args, "-R")
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*Iperf3Result), nil
}

// Iperf3Server runs iperf3 in server mode (daemon)
func (t *Iperf3Tool) Iperf3Server(ctx context.Context, port int, oneOff bool) error {
	args := []string{
		"-s",
		"-p", strconv.Itoa(port),
	}

	if oneOff {
		args = append(args, "-1")
	}

	_, err := t.Execute(ctx, args)
	return err
}

// Register throughput tools
func RegisterThroughputTools(registry *ToolRegistry) {
	registry.Register(NewIperf3Tool())
}
