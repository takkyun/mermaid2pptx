package convert

// Pt is a point in SVG pixel coordinates (y-down).
type Pt struct {
	X, Y float64
}

// Rect is a rectangle in SVG pixel coordinates, top-left based.
type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Cx() float64 { return r.X + r.W/2 }
func (r Rect) Cy() float64 { return r.Y + r.H/2 }

func (r Rect) contains(p Pt) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}

// Run is a text run within a paragraph.
type Run struct {
	Text   string
	Bold   bool
	Italic bool
}

// Para is one paragraph of a label.
type Para struct {
	Runs  []Run
	Align string // "" = center, "l" = left
}

func (p Para) empty() bool {
	for _, r := range p.Runs {
		if r.Text != "" {
			return false
		}
	}
	return true
}

// ShapeKind selects the PowerPoint preset geometry for a node.
type ShapeKind int

const (
	KindRect     ShapeKind = iota // roundRect
	KindDiamond                   // diamond
	KindCylinder                  // can
	KindEllipse                   // ellipse (mermaid circle/ellipse nodes)
	KindHexagon                   // hexagon
	KindPolygon                   // custGeom polygon (trapezoid, parallelogram, ...)
	KindBox                       // plain square rect (class / ER boxes)
)

// Node is a diagram node (flowchart node, state, class box, actor, note, ...).
type Node struct {
	ID        string
	Kind      ShapeKind
	R         Rect
	Fill      string // "RRGGBB", "" = mermaid default
	Stroke    string
	TextColor string
	Adj       int  // preset adjust: "can" ry ratio / roundRect corner (0 = default)
	Poly      []Pt // world-coordinate vertices for KindPolygon
	Label     []Para
}

// Cluster is a subgraph / composite-state container.
type Cluster struct {
	ID        string
	R         Rect
	Fill      string
	Stroke    string
	TextColor string
	Label     []Para
}

// Edge is a link between two nodes, routed along Points.
type Edge struct {
	ID         string
	From       string // node ID, may be "" when unresolved
	To         string
	Points     []Pt // routing waypoints incl. endpoints (from data-points)
	Dashed     bool
	Thick      bool
	StartArrow string // OOXML line-end type at the start ("" = none)
	EndArrow   string // OOXML line-end type at the end ("" = none)
}

// EdgeLabel is a label box with translucent background placed on an edge.
type EdgeLabel struct {
	C         Pt // center
	W, H      float64
	TextColor string
	Label     []Para
}

// Line is a plain line segment (lifelines, compartment dividers, messages).
type Line struct {
	P1, P2     Pt
	Color      string  // "" = default edge color
	WidthPx    float64 // 0 = 1px
	Dashed     bool
	StartArrow string
	EndArrow   string
	Above      bool // draw above node shapes (dividers, messages)
}

// TextBox is a free-standing borderless text (message labels, members, ...).
type TextBox struct {
	R     Rect
	Color string
	Label []Para
}

// Diagram is the parsed mermaid diagram.
type Diagram struct {
	Type       string  // "flowchart", "state", "class", "er", "sequence"
	MinX, MinY float64 // viewBox origin
	W, H       float64 // viewBox size in px
	Clusters   []Cluster
	Nodes      []Node
	Edges      []Edge
	Labels     []EdgeLabel
	Lines      []Line
	TextBoxes  []TextBox
}

// Mermaid default theme colors (from the embedded CSS of the SVG output).
const (
	defNodeFill    = "ECECFF"
	defNodeStroke  = "9370DB"
	defTextColor   = "333333"
	defClusterFill = "FFFFDE"
	defClusterLine = "AAAA33"
	defEdgeColor   = "333333"
	defLabelBg     = "E8E8E8"
)

// clusterDefaults returns the default fill/stroke of subgraph containers,
// which differ between diagram types.
func (d *Diagram) clusterDefaults() (string, string) {
	if d.Type == "state" {
		return defNodeFill, defNodeStroke
	}
	return defClusterFill, defClusterLine
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
