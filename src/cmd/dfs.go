package cmd

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

var (
	baseElements     = map[string]bool{"water": true, "fire": true, "earth": true, "air": true}
	visitedNodeCount int64
)


func isBaseElement(e string) bool {
	return baseElements[e]
}

// dfsBuildTree membangun pohon recipe menggunakan DFS
func dfsBuildTree(recipes RecipeMap, tiers TierMap, target string, maxPaths int, visited map[string]bool, depth int, maxDepth int) []*ElementNode {

	if isBaseElement(target) {
		return []*ElementNode{
			{
				Result:   target,
				Sources:  nil,
				Children: nil,
			},
		}
	}
	
	// Cek apakah sudah dikunjungi atau melebihi kedalaman maksimum
	if visited[target] || depth > maxDepth {
		return []*ElementNode{
			{
				Result: target,
			},
		}
	}
	
	// Cek apakah elemen memiliki resep
	combos, exists := recipes[target]
	if !exists || len(combos) == 0 {
		return []*ElementNode{
			{
				Result: target,
			},
		}
	}
	

	newVisited := make(map[string]bool)
	for k, v := range visited {
		newVisited[k] = v
	}
	newVisited[target] = true
	
	// Dapatkan tier elemen target
	parentTier := tiers[target]
	

	resultsChan := make(chan *ElementNode, maxPaths)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Saluran untuk resep
	comboChan := make(chan []string, len(combos))
	
	var wg sync.WaitGroup
	var result []*ElementNode
	var resultMutex sync.Mutex

	worker := func() {
		defer wg.Done()
		for combo := range comboChan {
			// Cek apakah sudah mencapai batas maksimum
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
			
	
			leftTrees := dfsBuildTree(recipes, tiers, combo[0], maxPaths, newVisited, depth+1, maxDepth)
			rightTrees := dfsBuildTree(recipes, tiers, combo[1], maxPaths, newVisited, depth+1, maxDepth)
			
			// Kombinasikan sub-pohon
			for _, left := range leftTrees {
				for _, right := range rightTrees {
					select {
					case resultsChan <- &ElementNode{
						Result:   target,
						Sources:  combo,
						Children: []*ElementNode{left, right},
					}:
					case <-ctx.Done():
						return
					}
					
					// Cek apakah sudah mencapai batas maksimum
					resultMutex.Lock()
					if len(result) >= maxPaths {
						resultMutex.Unlock()
						return
					}
					resultMutex.Unlock()
				}
			}
		}
	}
	
	// Mulai worker
	workerCount := 4 
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}
	
	// Kirim resep ke worker
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
	
	return result
}


func MainDfs(recipes RecipeMap, tiers TierMap, targetElement string, maxRecipes int) Result {

	atomic.StoreInt64(&visitedNodeCount, 0)

	startTime := time.Now()

	if isBaseElement(targetElement) {
		return Result{
			TargetElement: targetElement,
			RecipeTree:    []ElementNode{{Result: targetElement}},
			VisitedNodes:  1,
			SearchTime:    0,
		}
	}
	
	trees := dfsBuildTree(recipes, tiers, targetElement, maxRecipes, make(map[string]bool), 0, 15)

	searchTime := float64(time.Since(startTime).Milliseconds())
	

	result := Result{
		TargetElement: targetElement,
		RecipeTree:    flattenTreeList(trees),
		VisitedNodes:  int(atomic.LoadInt64(&visitedNodeCount)),
		SearchTime:    searchTime,
	}
	
	return result
}

func GetVisitedNodeCount() int64 {
	return atomic.LoadInt64(&visitedNodeCount)
}

func ResetVisitedNodeCount() {
	atomic.StoreInt64(&visitedNodeCount, 0)
}
