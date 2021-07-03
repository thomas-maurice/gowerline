package types

import "encoding/json"

// Payload is the data that will be passed down to the plugin
type Payload struct {
	Function string            `json:"function"`
	Args     *json.RawMessage  `json:"args"`
	Env      map[string]string `json:"env"`
}

// This represents the object powerline will read as
// the segment data (for one segment)
// Taken from https://powerline.readthedocs.io/en/master/develop/segments.html
type PowerlineReturn struct {
	Content               string   `json:"contents"`
	HighlightGroup        []string `json:"highlight_groups"`
	DrawInnerDivider      bool     `json:"draw_inner_divider"`
	DrawSoftDivider       bool     `json:"draw_soft_divider"`
	DrawHardDivider       bool     `json:"draw_hard_divider"`
	DividerHighlightGroup string   `json:"divider_highlight_group"`
}

// Contains registration data like the functions
// that the plugin maps to
type PluginStartData struct {
	Functions []string `json:"functions"`
}
