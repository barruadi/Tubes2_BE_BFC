package cmd

import (
	"context"
	"sync"
	"time"
)

// Menyimpan state untuk pencarian dua arah
type BidirectionalState struct {
	ForwardCache  map[string][]*ElementNode
	BackwardCache map[string][][]string
	mu            sync.RWMutex
}

// Inisialisasi state baru
func NewBidirectionalState() *BidirectionalState {
	return &BidirectionalState{
		ForwardCache:  make(map[string][]*ElementNode),
		BackwardCache: make(map[string][][]string),
	}
}

// Membangun jalur dari elemen dasar ke target
func generateBackwardPaths(recipes RecipeMap, tiers TierMap, state *BidirectionalState) {
	queue := make([]string, 0)
	for el := range abaseElements {
		queue = append(queue, el)
		state.mu.Lock()
		state.BackwardCache[el] = [][]string{{el}}
		state.mu.Unlock()
	}

	visited := make(map[string]bool)
	for el := range abaseElements {
		visited[el] = true
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		currTier := tiers[curr]

		for result, combos := range recipes {
			resultTier := tiers[result]
			if currTier >= resultTier {
				continue
			}

			for _, combo := range combos {
				if combo[0] == curr || combo[1] == curr {
					var other string
					if combo[0] == curr {
						other = combo[1]
					} else {
						other = combo[0]
					}

					if tiers[other] >= resultTier {
						continue
					}

					state.mu.RLock()
					_, otherVisited := state.BackwardCache[other]
					state.mu.RUnlock()

					if otherVisited {
						state.mu.Lock()
						state.BackwardCache[result] = append(state.BackwardCache[result], combo)
						if !visited[result] {
							queue = append(queue, result)
							visited[result] = true
						}
						state.mu.Unlock()
					}
				}
			}
		}
	}
}

// Membangun pohon dari target ke elemen dasar
func forwardBuildTree(
	recipes RecipeMap,
	tiers TierMap,
	target string,
	maxPaths int,
	state *BidirectionalState,
	visitedNodes *int,
) []*ElementNode {
	var result []*ElementNode

	if isBase(target) {
		node := &ElementNode{Result: target}
		state.mu.Lock()
		state.ForwardCache[target] = []*ElementNode{node}
		state.mu.Unlock()
		*visitedNodes++
		return []*ElementNode{node}
	}

	state.mu.RLock()
	if cached, ok := state.ForwardCache[target]; ok {
		state.mu.RUnlock()
		return cached
	}
	state.mu.RUnlock()

	state.mu.RLock()
	backwardPaths, exists := state.BackwardCache[target]
	state.mu.RUnlock()

	if exists {
		for _, path := range backwardPaths {
			if len(path) == 1 {
				result = append(result, &ElementNode{Result: target})
				*visitedNodes++
			} else if len(path) == 2 {
				left, right := path[0], path[1]
				if tiers[left] >= tiers[target] || tiers[right] >= tiers[target] {
					continue
				}

				leftTrees := forwardBuildTree(recipes, tiers, left, maxPaths, state, visitedNodes)
				rightTrees := forwardBuildTree(recipes, tiers, right, maxPaths, state, visitedNodes)

				for _, l := range leftTrees {
					for _, r := range rightTrees {
						if len(result) >= maxPaths {
							break
						}
						result = append(result, &ElementNode{
							Result:   target,
							Sources:  path,
							Children: []*ElementNode{l, r},
						})
						*visitedNodes++
					}
					if len(result) >= maxPaths {
						break
					}
				}
			}
			if len(result) >= maxPaths {
				break
			}
		}

		if len(result) > 0 {
			state.mu.Lock()
			state.ForwardCache[target] = result
			state.mu.Unlock()
			return result
		}
	}

	combos, exists := recipes[target]
	if !exists {
		return result
	}

	parentTier := tiers[target]
	queue := make(chan []string, len(combos))
	resultsChan := make(chan *ElementNode, maxPaths)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case pair, ok := <-queue:
				if !ok {
					return
				}

				if isUnbuildable(pair[0], recipes) || isUnbuildable(pair[1], recipes) {
					continue
				}

				if tiers[pair[0]] >= parentTier || tiers[pair[1]] >= parentTier {
					continue
				}

				leftTrees := forwardBuildTree(recipes, tiers, pair[0], maxPaths, state, visitedNodes)
				rightTrees := forwardBuildTree(recipes, tiers, pair[1], maxPaths, state, visitedNodes)

				for _, l := range leftTrees {
					for _, r := range rightTrees {
						select {
						case resultsChan <- &ElementNode{
							Result:   target,
							Sources:  pair,
							Children: []*ElementNode{l, r},
						}:
							*visitedNodes++
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}

	workerCount := 2 + parentTier*2
	if workerCount > 16 {
		workerCount = 16
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		for _, combo := range combos {
			queue <- combo
		}
		close(queue)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for node := range resultsChan {
		result = append(result, node)
		if len(result) >= maxPaths {
			cancel()
			break
		}
	}

	state.mu.Lock()
	state.ForwardCache[target] = result
	state.mu.Unlock()

	return result
}

func MainBidirectionalBfs(recipes RecipeMap, tiers TierMap, target string, maxPaths int) Result {
	var res Result

	state := NewBidirectionalState()
	visited := 0

	generateBackwardPaths(recipes, tiers, state)

	start := time.Now()
	trees := forwardBuildTree(recipes, tiers, target, maxPaths, state, &visited)
	res.SearchTime = float64(time.Since(start).Microseconds())
	res.RecipeTree = flattenTreeList(trees)
	res.VisitedNodes = visited

	return res
}