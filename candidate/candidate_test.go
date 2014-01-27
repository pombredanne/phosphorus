package candidate

import (
	"container/heap"
	"testing"
)

var candidateSets = [][]uint{
	[]uint{50, 40, 100, 70, 60, 20, 80},
	[]uint{80, 40, 50, 90},
	[]uint{100, 80, 50, 60, 70, 30, 40},
	[]uint{30, 40, 90},
	[]uint{10, 30, 60, 70, 40},
	[]uint{100, 20, 10, 30, 70},
	[]uint{80, 10}}

func TestCandidate (t *testing.T) {
	var candidateHeaps []*CandidateHeap
	for i, candidateSet := range candidateSets {
		candidateHeap := &CandidateHeap{}
		heap.Init(candidateHeap)
		for _, recordId := range candidateSet {
			newCandidate := &Candidate{recordId, i}
			heap.Push(candidateHeap, newCandidate)
		}

		candidateHeaps = append(candidateHeaps, candidateHeap)
	}

	_ = MergeCandidateHeap(candidateHeaps...)
}
