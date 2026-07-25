package assets

import _ "embed"

//go:embed template.mm
var TemplateMindMap []byte

// TemplateMindMapFileName is the default destination filename for mm+.
const TemplateMindMapFileName = "template.mm"
