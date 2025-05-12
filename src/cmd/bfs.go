package cmd

import (
	"time"
	"sync"
	"context"
)

var abaseElements = map[string]bool{
	"water": true,
	"fire":  true,
	"earth": true,
	"air":   true,
}

// ----------------- HELPER -----------------
func isBase(e string) bool {
	return abaseElements[e]
}

func flattenTreeList(trees []*ElementNode) []ElementNode {
	result := make([]ElementNode, len(trees))
	for i, t := range trees {
		result[i] = *t
	}
	return result
}

func isUnbuildable(e string, recipes RecipeMap) bool {
	return !isBase(e) && len(recipes[e]) == 0
}

func countNodes(node *ElementNode) int {
	if node == nil {
		return 0
	}
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}


func bfsBuildTree(recipes RecipeMap, tiers TierMap, target string, maxPaths int) []*ElementNode {
	var result []*ElementNode

	if isBase(target) {
		return []*ElementNode{
			{
				Result:   target,
				Sources:  nil,
				Children: nil,
			},
		}
	}

	combos, exists := recipes[target]
	if !exists {
		return nil
	}

	parentTier := tiers[target]

	queue := make(chan []string, len(combos))
	resultsChan := make(chan *ElementNode, maxPaths)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for pair := range queue {
			if isUnbuildable(pair[0], recipes) || isUnbuildable(pair[1], recipes) {
				continue
			}

			tierA := tiers[pair[0]]
			tierB := tiers[pair[1]]

			if tierA >= parentTier || tierB >= parentTier {
				continue
			}

			leftTrees := bfsBuildTree(recipes, tiers, pair[0], maxPaths)
			rightTrees := bfsBuildTree(recipes, tiers, pair[1], maxPaths)

			for _, left := range leftTrees {
				for _, right := range rightTrees {
					select {
					case resultsChan <- &ElementNode{
						Result:   target,
						Sources:  pair,
						Children: []*ElementNode{left, right},
					}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}

	// --------- Start workers ---------
	workerCount := 2 + parentTier*2
	if workerCount > 16 {
		workerCount = 16
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		for _, pair := range combos {
			queue <- pair
		}
		close(queue)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// --------- results ---------
	for node := range resultsChan {
		result = append(result, node)
		if len(result) >= maxPaths {
			cancel()
			break
		}
	}

	return result
}

func MainBfs(recipes RecipeMap, tiers TierMap, target string, maxPaths int) BfsResult {
	var bfsResult BfsResult
	startTime := time.Now()
	trees := bfsBuildTree(recipes, tiers, target, maxPaths)
	bfsResult.SearchTime = float64(time.Since(startTime).Milliseconds())
	bfsResult.RecipeTree = flattenTreeList(trees)
	totalNodes := 0
	for _, tree := range bfsResult.RecipeTree {
		totalNodes += countNodes(&tree)
	}
	bfsResult.VisitedNodes = totalNodes

	return bfsResult
}
