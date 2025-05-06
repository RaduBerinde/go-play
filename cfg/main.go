package main

import (
	"encoding/json"
	"github.com/dustin/go-humanize"
	"github.com/hjson/hjson-go/v4"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
	"time"
)

type Config struct {
	Logs    LogConfig     `yaml:"logs" json:"logs" toml:"logs"`
	Storage StorageConfig `yaml:"storage" json:"storage" toml:"storage"`
}

type LogConfig struct {
	FileDefaults FileDefaults `yaml:"file-defaults,omitempty" json:"file-defaults" toml:"file-defaults"`
	Buffering    LogBuffering `yaml:"buffering,omitempty" json:"buffering" toml:"buffering"`
}

type FileDefaults struct {
	BufferedWrites *bool `yaml:"buffered-writes,omitempty" json:"buffered-writes" toml:"buffered-writes" comment:"Note that the sink itself can still be buffered."`
}

type LogBuffering struct {
	MaxStaleness     *time.Duration `yaml:"max-staleness" json:"nax-staleness,omitempty" toml:"max-staleness"`
	FlushTriggerSize *ByteSize      `yaml:"flush-trigger-size" json:"flush-trigger-size,omitempty" toml:"flush-trigger-size"`
	MaxBufferSize    *ByteSize      `yaml:"max-buffer-size" json:"max-buffer-size,omitempty" toml:"max-buffer-size"`
}

type StorageConfig struct {
	StorePaths      []string `yaml:"store-paths" json:"store-paths" toml:"store-paths"`
	WALFailoverMode string   `yaml:"wal-failover-mode" json:"wal-failover-mode" toml:"wal-failover-mode" comment:"See xyz for details on WAL failover mode."`
}

type StoreConfig struct {
	Attrs           string                `yaml:"attrs" json:"attrs" toml:"attrs"`
	Size            ByteSize              `yaml:"size" json:"size" toml:"size"`
	BallastSize     ByteSize              `yaml:"ballast" json:"ballast" toml:"ballast"`
	ProvisionedRate string                `yaml:"provisioned-rate" json:"provisioned-rate" toml:"provisioned-rate"`
	Pebble          []PebbleConfigSection `yaml:"pebble" json:"pebble" toml:"pebble" comment:"See xyz for details on Pebble options."`
}

type PebbleConfigSection struct {
	Section  string   `yaml:"section" toml:"section" json:"section"`
	Settings []string `yaml:"settings" json:"settings" toml:"settings"`
}

func main() {
	c := Config{
		Logs: LogConfig{
			FileDefaults: FileDefaults{
				BufferedWrites: new(bool),
			},
			Buffering: LogBuffering{
				MaxStaleness:     func() *time.Duration { s := time.Second; return &s }(),
				FlushTriggerSize: func() *ByteSize { s := ByteSize(256 << 10); return &s }(),
				MaxBufferSize:    func() *ByteSize { s := ByteSize(50 << 20); return &s }(),
			},
		},
		Storage: StorageConfig{
			StorePaths:      []string{"/mnt/data1", "/mnt/data2"},
			WALFailoverMode: "AMONG_STORES",
		},
	}

	sc := StoreConfig{
		Attrs:           "vol-0b7bdb6d48162294a:type_gp3:bw_125_mibps:iops_3000:size_256_gib",
		Size:            100 << 30,
		BallastSize:     1 << 30,
		ProvisionedRate: "bandwidth=100",
		//		Pebble: strings.Split(`[Options]
		//mem_table_stop_writes_threshold=100
		//[Level "0"]
		//index_block_size=256000000
		//[Level "1"]
		//index_block_size=256000000
		//target_file_size=2000000
		//[Level "2"]
		//index_block_size=256000000
		//target_file_size=4000000`, "\n"),
		Pebble: []PebbleConfigSection{
			{
				Section:  `Options`,
				Settings: []string{"mem_table_stop_writes_threshold=100"},
			},
			{
				Section:  `Level "0"`,
				Settings: []string{"index_block_size=256000000"},
			},
			{
				Section:  `Level "1"`,
				Settings: []string{"index_block_size=256000000", "target_file_size=2000000"},
			},
			{
				Section:  `Level "2"`,
				Settings: []string{"index_block_size=256000000", "target_file_size=4000000"},
			},
		},
	}

	writeHJSON(c, "out/1-crdb-config.hjson")
	writeHJSON(sc, "out/1-crdb-store.hjson")

	writeTOML(c, "out/2-crdb-config.toml")
	writeTOML(sc, "out/2-crdb-store.toml")

	writeJSON(c, "out/3-crdb-config.json")
	writeJSON(sc, "out/3-crdb-store.json")

	writeYAML(c, "out/4-crdb-config.yaml")
	writeYAML(sc, "out/4-crdb-store.yaml")
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func writeTOML(v interface{}, path string) {
	var buf strings.Builder
	enc := toml.NewEncoder(&buf)
	enc.SetArraysMultiline(true)
	checkErr(enc.Encode(v))
	checkErr(os.WriteFile(path, []byte(buf.String()), 0666))
}

func writeHJSON(v interface{}, path string) {
	opts := hjson.DefaultOptions()
	opts.EmitRootBraces = false
	b, err := hjson.MarshalWithOptions(v, opts)
	checkErr(err)
	checkErr(os.WriteFile(path, b, 0666))
}

func writeJSON(v interface{}, path string) {
	b, err := json.MarshalIndent(v, "", "  ")
	checkErr(err)
	checkErr(os.WriteFile(path, b, 0666))
}

func writeYAML(v interface{}, path string) {
	b, err := yaml.Marshal(v)
	checkErr(err)
	checkErr(os.WriteFile(path, b, 0666))
}

type ByteSize int

// String implements the fmt.Stringer interface.
func (x ByteSize) String() string {
	return strings.ReplaceAll(humanize.IBytes(uint64(x)), " ", "")
}

func (x ByteSize) IsZero() bool { return x == 0 }

func (x ByteSize) MarshalYAML() (interface{}, error) {
	return x.String(), nil
}
func (x ByteSize) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

func (x ByteSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}
