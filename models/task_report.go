package models

// TaskReportTreeNode is a tree of task report lines (mirrors the Node TreeNode shape).
// GetTaskReport returns nil when there are no done descendant tasks under the root — not an error.
type TaskReportTreeNode struct {
	Name     string               `json:"name"`
	Depth    int                  `json:"depth"`
	Children []*TaskReportTreeNode `json:"children,omitempty"`
}
