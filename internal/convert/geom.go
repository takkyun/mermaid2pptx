package convert

import "math"

// side of a shape an edge attaches to.
type side int

const (
	sideTop side = iota
	sideLeft
	sideBottom
	sideRight
)

// cxnIdx maps a side to the preset-geometry connection site index
// (rect / roundRect / diamond / can all enumerate top, left, bottom, right).
func (s side) cxnIdx() int { return int(s) }

func (s side) vertical() bool { return s == sideTop || s == sideBottom }

// outward is +1 when leaving the shape means increasing coordinate
// (bottom/right), -1 otherwise.
func (s side) outward() float64 {
	if s == sideBottom || s == sideRight {
		return 1
	}
	return -1
}

// sideOf classifies which side of rect r the boundary point p lies on.
func sideOf(r Rect, p Pt) side {
	dx := (p.X - r.Cx()) / math.Max(r.W/2, 1e-6)
	dy := (p.Y - r.Cy()) / math.Max(r.H/2, 1e-6)
	if math.Abs(dx) > math.Abs(dy) {
		if dx > 0 {
			return sideRight
		}
		return sideLeft
	}
	if dy > 0 {
		return sideBottom
	}
	return sideTop
}

// tangentSideOut classifies which side of the source shape an edge leaves,
// based on the direction v of its first segment.
func tangentSideOut(v Pt) side {
	if math.Abs(v.X) > math.Abs(v.Y) {
		if v.X > 0 {
			return sideRight
		}
		return sideLeft
	}
	if v.Y > 0 {
		return sideBottom
	}
	return sideTop
}

// tangentSideIn classifies which side of the target shape an edge enters,
// based on the direction v of its last segment (moving right enters the
// left side, moving down enters the top, ...).
func tangentSideIn(v Pt) side {
	if math.Abs(v.X) > math.Abs(v.Y) {
		if v.X > 0 {
			return sideLeft
		}
		return sideRight
	}
	if v.Y > 0 {
		return sideTop
	}
	return sideBottom
}

// segmentDir returns the first non-zero segment direction from the start
// (forward) or the end (backward) of the polyline.
func segmentDir(pts []Pt, fromEnd bool) Pt {
	if fromEnd {
		for i := len(pts) - 1; i > 0; i-- {
			v := Pt{pts[i].X - pts[i-1].X, pts[i].Y - pts[i-1].Y}
			if v.X != 0 || v.Y != 0 {
				return v
			}
		}
	} else {
		for i := 1; i < len(pts); i++ {
			v := Pt{pts[i].X - pts[0].X, pts[i].Y - pts[0].Y}
			if v.X != 0 || v.Y != 0 {
				return v
			}
		}
	}
	return Pt{}
}

// connGeom describes a preset connector placement in EMU.
type connGeom struct {
	prst         string
	offX, offY   int64
	cx, cy       int64
	rot          int64 // 1/60000 degree, clockwise
	flipH, flipV bool
	ok           bool
}

const alignEps = 1.5 // px: treat as straight when perpendicular offset is below

// connectorGeom computes the xfrm of a preset connector whose local start
// maps to S and local end (arrow side) maps to E, in px coordinates.
//
// Local geometries (bounding box w x h, before flips/rotation):
//   - straightConnector1: (0,0) -> (w,h)
//   - curvedConnector2:   (0,0) tangent +x -> (w,h) tangent +y
//   - curvedConnector3:   (0,0) tangent +x -> (w,h) tangent +x (S-curve)
//
// flipH/flipV mirror inside the box; rot rotates about the box center.
func connectorGeom(S, E Pt, sideS, sideE side, toEMU func(Pt) (int64, int64)) connGeom {
	g := connGeom{}
	rotate := false // true: exit direction is vertical -> rotate frame 90° cw

	switch {
	case sideS.vertical() != sideE.vertical():
		g.prst = "curvedConnector2"
		rotate = sideS.vertical()
	default:
		// same axis: must travel the same direction on that axis
		var travel, along, perp float64
		if sideS.vertical() {
			travel, along, perp = sideS.outward(), E.Y-S.Y, E.X-S.X
		} else {
			travel, along, perp = sideS.outward(), E.X-S.X, E.Y-S.Y
		}
		// arriving into sideE means moving opposite to sideE's outward
		if -sideE.outward() != travel || along*travel <= 0 {
			return g // U-turn shape: not representable, use freeform fallback
		}
		if math.Abs(perp) < alignEps {
			g.prst = "straightConnector1"
		} else {
			g.prst = "curvedConnector3"
		}
		rotate = sideS.vertical() && g.prst != "straightConnector1"
	}

	sx, sy := toEMU(S)
	ex, ey := toEMU(E)
	minX, maxX := min64(sx, ex), max64(sx, ex)
	minY, maxY := min64(sy, ey), max64(sy, ey)
	w, h := maxX-minX, maxY-minY

	if !rotate {
		g.offX, g.offY, g.cx, g.cy = minX, minY, w, h
		g.flipH = sx > ex
		g.flipV = sy > ey
	} else {
		// pre-rotation frame: rotate world by -90° about the box center
		cx2, cy2 := minX+w/2, minY+h/2 // center (x2 to avoid shadowing)
		g.cx, g.cy = h, w
		g.offX, g.offY = cx2-h/2, cy2-w/2
		g.rot = 5400000 // 90° cw
		// S_pre = c + Rinv(S-c), Rinv(x,y) = (y,-x)
		spx := cx2 + (sy - cy2)
		spy := cy2 - (sx - cx2)
		epx := cx2 + (ey - cy2)
		epy := cy2 - (ex - cx2)
		g.flipH = spx > epx
		g.flipV = spy > epy
	}
	g.ok = true
	return g
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
