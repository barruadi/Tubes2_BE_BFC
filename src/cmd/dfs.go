package cmd

import (
	"context"
	"sync"
	"time"
)

var (
	baseElements     = map[string]bool{"water": true, "fire": true, "earth": true, "air": true}

)


func isBaseElement(e string) bool {
	return baseElements[e]
}


func dfsBuildTree(
	recipes RecipeMap,
	tiers TierMap,
	target string,
	maxPaths int,
	visited map[string]bool,
	depth int,
	maxDepth int,
	memo *MemoCache, 
) []*ElementNode {

	if isBaseElement(target) {
		return []*ElementNode{
			{
				Result:   target,
				Sources:  nil,
				Children: nil,
			},
		}
	}

	memo.mu.Lock()
	if val, ok := memo.store[target]; ok {
		memo.mu.Unlock()
		return val
	}
	memo.mu.Unlock()

	if visited[target] || depth > maxDepth {
		return []*ElementNode{{Result: target}}
	}

	combos, exists := recipes[target]
	if !exists || len(combos) == 0 {
		return []*ElementNode{{Result: target}}
	}

	newVisited := make(map[string]bool)
	for k, v := range visited {
		newVisited[k] = v
	}
	newVisited[target] = true

	parentTier := tiers[target]

	resultsChan := make(chan *ElementNode, maxPaths)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	comboChan := make(chan []string, len(combos))

	var wg sync.WaitGroup
	var result []*ElementNode
	var resultMutex sync.Mutex

	worker := func() {
		defer wg.Done()
		for combo := range comboChan {
			// Cek limit sebelum lanjut
			resultMutex.Lock()
			if len(result) >= maxPaths {
				resultMutex.Unlock()
				return
			}
			resultMutex.Unlock()
	
			if isUnbuildable(combo[0], recipes) || isUnbuildable(combo[1], recipes) {
				continue
			}
	
			tierA := tiers[combo[0]]
			tierB := tiers[combo[1]]
			if tierA >= parentTier || tierB >= parentTier {
				continue
			}
	
			leftTrees := dfsBuildTree(recipes, tiers, combo[0], maxPaths, newVisited, depth+1, maxDepth, memo)
			rightTrees := dfsBuildTree(recipes, tiers, combo[1], maxPaths, newVisited, depth+1, maxDepth, memo)
	
			for _, left := range leftTrees {
				for _, right := range rightTrees {
					resultMutex.Lock()
					if len(result) >= maxPaths {
						resultMutex.Unlock()
						cancel()
						return
					}
					newNode := &ElementNode{
						Result:   target,
						Sources:  combo,
						Children: []*ElementNode{left, right},
					}
					result = append(result, newNode)
					resultMutex.Unlock()
				}
			}
		}
	}
	

	workerCount := 4
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		for _, combo := range combos {
			comboChan <- combo
		}
		close(comboChan)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for node := range resultsChan {
		resultMutex.Lock()
		result = append(result, node)
		if len(result) >= maxPaths {
			cancel()
		}
		resultMutex.Unlock()
	}


	memo.mu.Lock()
	memo.store[target] = result
	memo.mu.Unlock()

	return result
}


func MainDfs(recipes RecipeMap, tiers TierMap, targetElement string, maxRecipes int) Result {

	// atomic.StoreInt64(&visitedNodeCount, 0)

	startTime := time.Now()
	cache := &MemoCache{store: make(map[string][]*ElementNode)}

	if isBaseElement(targetElement) {
		return Result{
			TargetElement: targetElement,
			RecipeTree:    []ElementNode{{Result: targetElement}},
			VisitedNodes:  1,
			SearchTime:    0,
		}
	}
	
	trees := dfsBuildTree(recipes, tiers, targetElement, maxRecipes, make(map[string]bool), 0, 15, cache)

	searchTime := float64(time.Since(startTime).Microseconds())
	
    totalNodes := 0
    for _, tree := range trees {
        totalNodes += countNodes(tree)
    }
	result := Result{
		TargetElement: targetElement,
		RecipeTree:    flattenTreeList(trees),
		VisitedNodes:  totalNodes,
		SearchTime:    searchTime,
	}
	
	return result
}

