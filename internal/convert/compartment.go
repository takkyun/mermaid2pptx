package convert

import (
	"math"
	"strconv"
	"strings"
)

// Compartment nodes (classDiagram classes, erDiagram entities) are boxes with
// title / member sections separated by divider lines. They are decomposed
// into: one box shape + exactly-positioned text boxes + divider lines.

// elColor reads a color from the style attribute first, then from the
// presentation attribute (class/er/sequence SVGs use fill="#..." attributes).
func elColor(el *xnode, key string) string {
	if c := styleColor(parseStyleDecls(el.get("style")), key); c != "" {
		return c
	}
	v := el.get(key)
	if v == "" || v == "none" {
		return ""
	}
	return cssColorToHex(v)
}

// pathBBox scans the coordinate pairs of a path's d attribute.
func pathBBox(d string) (Rect, bool) {
	nums := reNumber.FindAllString(d, -1)
	if len(nums) < 4 {
		return Rect{}, false
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for i := 0; i+1 < len(nums); i += 2 {
		x, _ := strconv.ParseFloat(nums[i], 64)
		y, _ := strconv.ParseFloat(nums[i+1], 64)
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minY, maxY = math.Min(minY, y), math.Max(maxY, y)
	}
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}, true
}

// parseCompartmentNode parses one class / entity node into d.Nodes,
// d.TextBoxes and d.Lines.
func parseCompartmentNode(g *xnode, d *Diagram, dx, dy float64) {
	tx, ty, ok := parseTranslate(g.get("transform"))
	if !ok {
		return
	}
	tx, ty = tx+dx, ty+dy
	n := Node{ID: nodeID(g.get("id")), Kind: KindBox}
	var lines []Line

	// geometry: fill path -> box, degenerate paths -> dividers,
	// stroked outline path -> border color
	g.walk(func(k *xnode) {
		if k.tag == "rect" && k.get("width") != "" {
			// entities without attributes are plain rects
			w, _ := strconv.ParseFloat(k.get("width"), 64)
			h, _ := strconv.ParseFloat(k.get("height"), 64)
			x, _ := strconv.ParseFloat(k.get("x"), 64)
			y, _ := strconv.ParseFloat(k.get("y"), 64)
			if n.R.W == 0 {
				n.R = Rect{X: tx + x, Y: ty + y, W: w, H: h}
			}
			st := parseStyleDecls(k.get("style"))
			n.Fill = orDefault(n.Fill, styleColor(st, "fill"))
			n.Stroke = orDefault(n.Stroke, styleColor(st, "stroke"))
			return
		}
		if k.tag != "path" {
			return
		}
		bb, ok := pathBBox(k.get("d"))
		if !ok {
			return
		}
		fill, stroke := elColor(k, "fill"), elColor(k, "stroke")
		switch {
		case bb.W < 0.1 || bb.H < 0.1:
			// divider line (drawn as a hairline path)
			if stroke != "" || k.get("stroke") != "none" {
				lines = append(lines, Line{
					P1: Pt{tx + bb.X, ty + bb.Y}, P2: Pt{tx + bb.X + bb.W, ty + bb.Y + bb.H},
					Color: stroke,
				})
			}
		case fill != "" && n.R.W == 0:
			// first filled area = the box itself
			n.R = Rect{X: tx + bb.X, Y: ty + bb.Y, W: bb.W, H: bb.H}
			n.Fill = fill
		case stroke != "" && n.Stroke == "":
			n.Stroke = stroke
		}
	})
	if n.R.W == 0 || n.R.H == 0 {
		return
	}
	// dividers default to the border color
	for i := range lines {
		if lines[i].Color == "" {
			lines[i].Color = orDefault(n.Stroke, defNodeStroke)
		}
		lines[i].Above = true
	}

	// text labels at their exact positions
	collectCompartmentLabels(g, tx, ty, d)

	d.Nodes = append(d.Nodes, n)
	d.Lines = append(d.Lines, lines...)
}

// collectCompartmentLabels walks label groups accumulating translates and
// emits one TextBox per label.
func collectCompartmentLabels(g *xnode, tx, ty float64, d *Diagram) {
	var visit func(k *xnode, dx, dy float64)
	visit = func(k *xnode, dx, dy float64) {
		if k.tag != "g" && k.tag != "foreignObject" {
			return
		}
		if x, y, ok := parseTranslate(k.get("transform")); ok {
			dx, dy = dx+x, dy+y
		}
		if k.tag == "g" && k.hasClass("label") {
			var fo *xnode
			k.walk(func(f *xnode) {
				if fo == nil && f.tag == "foreignObject" {
					fo = f
				}
			})
			if fo == nil {
				return
			}
			w, _ := strconv.ParseFloat(fo.get("width"), 64)
			h, _ := strconv.ParseFloat(fo.get("height"), 64)
			paras, color := extractLabel(fo)
			if len(paras) == 0 || w <= 0 {
				return
			}
			// class titles carry font-weight: bolder; ER titles have class
			// "label name" -- render both bold
			if strings.Contains(k.get("style"), "font-weight") || k.hasClass("name") {
				for i := range paras {
					for j := range paras[i].Runs {
						paras[i].Runs[j].Bold = true
					}
				}
			}
			for i := range paras {
				paras[i].Align = "l"
			}
			d.TextBoxes = append(d.TextBoxes, TextBox{
				R:     Rect{X: dx, Y: dy, W: w, H: h},
				Color: color,
				Label: paras,
			})
			return
		}
		for _, c := range k.kids {
			visit(c, dx, dy)
		}
	}
	for _, c := range g.kids {
		visit(c, tx, ty)
	}
}
