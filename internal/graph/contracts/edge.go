package contracts

type EdgeType string

const (
	EdgeCalls        EdgeType = "CALLS"
	EdgeImplements   EdgeType = "IMPLEMENTS"
	EdgeUses         EdgeType = "USES"
	EdgeImports      EdgeType = "IMPORTS"
	EdgeBelongsTo    EdgeType = "BELONGS_TO"
	EdgeDependsOn    EdgeType = "DEPENDS_ON"
	EdgeFlowsThrough EdgeType = "FLOWS_THROUGH"
)

type Edge struct {
	FromID   string                 `json:"from_id"`
	ToID     string                 `json:"to_id"`
	Type     EdgeType               `json:"type"`
	Sequence int                    `json:"sequence"`
	Metadata map[string]interface{} `json:"metadata"`
}
