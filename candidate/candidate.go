package candidate

import (
	"container/heap"
)

type Candidate struct {
	recordId uint
	source   int
}

type CandidateHeap []*Candidate

func (h CandidateHeap) Len() int           { return len(h) }
func (h CandidateHeap) Less(i, j int) bool { return h[i].recordId < h[j].recordId }
func (h CandidateHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *CandidateHeap) Push(c interface{}) {
	*h = append(*h, c.(*Candidate))
}

func (h *CandidateHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type WeightedCandidate struct {
	recordId           uint
	matchingSignatures int
}

type WeightedCandidateHeap []*WeightedCandidate

func (h WeightedCandidateHeap) Len() int { return len(h) }
func (h WeightedCandidateHeap) Less(i, j int) bool {
	return h[i].matchingSignatures > h[j].matchingSignatures
}
func (h WeightedCandidateHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *WeightedCandidateHeap) Push(c interface{}) {
	*h = append(*h, c.(*WeightedCandidate))
}

func (h *WeightedCandidateHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func MergeCandidateHeap(sourceHeaps ...*CandidateHeap) []WeightedCandidate {
	mergeHeap := &CandidateHeap{}
	heap.Init(mergeHeap)

	// Grab the minimum item from each source heap and push it onto mergeHeap
	var sourceHeap *CandidateHeap
	for _, sourceHeap = range sourceHeaps {
		candidate := heap.Pop(sourceHeap).(*Candidate)
		heap.Push(mergeHeap, candidate)
	}

	// While the mergeHeap still has candidates:
	// - pop the minimum item off mergeHeap and print it
	// - push an item from the same source heap as the item we just pushed
	//   onto the heap

	weightHeap := &WeightedCandidateHeap{}
	heap.Init(weightHeap)

	var lastRecordId uint
	currentMagnitude := 0
	for len(*mergeHeap) > 0 {
		candidate := heap.Pop(mergeHeap).(*Candidate)
		if lastRecordId == candidate.recordId {
			currentMagnitude++
		} else {
			if lastRecordId != 0 {
				heap.Push(weightHeap, &WeightedCandidate{
					lastRecordId, currentMagnitude})

			}
			lastRecordId = candidate.recordId
			currentMagnitude = 1
		}

		if len(*sourceHeaps[candidate.source]) > 0 {
			nextCandidate := heap.Pop(sourceHeaps[candidate.source]).(*Candidate)
			heap.Push(mergeHeap, nextCandidate)
		} else {
		}
	}

	cs := make([]WeightedCandidate, 0, len(*weightHeap))

	for len(*weightHeap) > 0 {
		wc := heap.Pop(weightHeap).(*WeightedCandidate)
		cs = append(cs, *wc)
	}

	return cs
}
