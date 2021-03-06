package types

import "encoding/json"

type VimInfo struct {
	WindowNumber int64  `json:"winnr"`
	BufferNumber int64  `json:"bufnr"`
	TabNumber    int64  `json:"tabnr"`
	Mode         string `json:"mode"`
}

// Payload is the data that will be passed down to the plugin
type Payload struct {
	Function string            `json:"function"`
	Args     *json.RawMessage  `json:"args"`
	Env      map[string]string `json:"env"`
	Cwd      string            `json:"cwd"`
	Vim      *VimInfo          `json:"vim"`
}

// This represents the object powerline will read as
// the segment data (for one segment)
// Taken from https://powerline.readthedocs.io/en/master/develop/segments.html
type PowerlineReturn struct {
	Content               string   `json:"contents,omitempty"`
	HighlightGroup        []string `json:"highlight_groups,omitempty"`
	DrawInnerDivider      bool     `json:"draw_inner_divider,omitempty"`
	DrawSoftDivider       bool     `json:"draw_soft_divider,omitempty"`
	DrawHardDivider       bool     `json:"draw_hard_divider,omitempty"`
	DividerHighlightGroup string   `json:"divider_highlight_group,omitempty"`
}

// Contains registration data like the functions
// that the plugin maps to
type PluginStartData struct {
	// Metadata about the plugin that just was
	// loaded
	Metadata PluginMetadata
}

type FunctionDescriptor struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Parameters  map[string]string `json:"parameters" yaml:"parameters"`
}

type PluginMetadata struct {
	Description string               `json:"description" yaml:"description"`
	Functions   []FunctionDescriptor `json:"functions" yaml:"functions"`
	Author      string               `json:"author" yaml:"author"`
	Version     string               `json:"version" yaml:"version"`
}

// ServerVersioInfo contains various infos about the server
// such as build version, date, arch and OS
type ServerVersionInfo struct {
	Version         string `json:"version" yaml:"version"`
	BuildHost       string `json:"build_host" yaml:"build_host"`
	BuildTime       string `json:"build_date" yaml:"build_time"`
	GitHash         string `json:"git_hash" yaml:"git_hash"`
	Architecture    string `json:"architecture" yaml:"architecture"`
	OperatingSystem string `json:"operating_system" yaml:"operating_system"`
}
