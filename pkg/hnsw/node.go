package hnsw

import "math"

type Vector []float64

type NodeID = uint32

type Node struct {
	// Index of the vector
	id NodeID
	v  Vector

	layer int

	// Layered list of neighbors,
	// each layer is a slice of NodeIds
	friends [][]NodeID
}

func NewNode(id NodeID, v Vector) *Node {
	return &Node{
		id,
		v,
		-1,
		make([][]NodeID, 0),
	}
}

func (n0 *Node) VecDistFromVec(v1 Vector) float64 {
	v0 := n0.v

	return EuclidDist(v0, v1)
}

func (n0 *Node) VecDistFromNode(n1 *Node) float64 {
	// pull vec from nodes
	v0 := n0.v
	v1 := n1.v

	return EuclidDist(v0, v1)
}

func EuclidDist(v0, v1 Vector) float64 {
	// check if vector dimensionality is correct
	if len(v0) != len(v1) {
		panic("invalid lengths")
	}

	var sum float64

	for i := range v0 {
		delta := float64(v0[i] - v1[i])
		sum += delta * delta
	}

	return math.Sqrt(sum)
}

// NearlyEqual is sourced from scalar package written by gonum
// https://pkg.go.dev/gonum.org/v1/gonum/floats/scalar#EqualWithinAbsOrRel
func NearlyEqual(a, b float64) bool {
	return EqualWithinAbs(a, b) || EqualWithinRel(a, b)
}

// EqualWithinAbs returns true when a and b have an absolute difference
// not greater than tol.
func EqualWithinAbs(a, b float64) bool {
	return a == b || math.Abs(a-b) <= 1e-6
}

// minNormalFloat64 is the smallest normal number. For 64 bit IEEE-754
// floats this is 2^{-1022}.
const minNormalFloat64 = 0x1p-1022

// EqualWithinRel returns true when the difference between a and b
// is not greater than tol times the greater absolute value of a and b,
//
//	abs(a-b) <= tol * max(abs(a), abs(b)).
func EqualWithinRel(a, b float64) bool {
	if a == b {
		return true
	}
	delta := math.Abs(a - b)
	if delta <= minNormalFloat64 {
		return delta <= 1e-6*minNormalFloat64
	}
	// We depend on the division in this relationship to identify
	// infinities (we rely on the NaN to fail the test) otherwise
	// we compare Infs of the same sign and evaluate Infs as equal
	// independent of sign.
	return delta/math.Max(math.Abs(a), math.Abs(b)) <= 1e-6
}
