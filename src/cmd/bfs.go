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

func isBase(e string) bool {
	return abaseElements[e]
}

func copyMap(original map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range original {
		newMap[k] = v
	}
	return newMap
}

// ---------------- Memoized Resolver ----------------

func resolveToBase(recipes RecipeMap, tiers TierMap, target string, maxTier int, cache map[string]*ElementNode) *ElementNode {
	if node, found := cache[target]; found {
		return node
	}

	if isBase(target) {
		base := &ElementNode{
			Result:   target,
			Sources:  nil,
			Children: nil,
		}
		cache[target] = base
		return base
	}

	combos, exists := recipes[target]
	if !exists {
		return nil
	}

	targetTier := tiers[target]
	if targetTier >= maxTier {
		return nil
	}

	for _, pair := range combos {
		tierA := tiers[pair[0]]
		tierB := tiers[pair[1]]

		if tierA >= targetTier || tierB >= targetTier {
			continue
		}

		left := resolveToBase(recipes, tiers, pair[0], targetTier, cache)
		right := resolveToBase(recipes, tiers, pair[1], targetTier, cache)

		if left != nil && right != nil {
			node := &ElementNode{
				Result:   target,
				Sources:  pair,
				Children: []*ElementNode{left, right},
			}
			cache[target] = node
			return node
		}
	}

	return nil
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

func FindNPathsBFS(recipes RecipeMap, tiers TierMap, target string, maxPaths int) BfsResult {
	type job struct {
		pair []string
	}

	type resultNode struct {
		tree ElementNode
	}

	combos, exists := recipes[target]
	if !exists {
		return BfsResult{TargetElement: target}
	}

	parentTier := tiers[target]
	jobs := make(chan job, len(combos))
	resultsChan := make(chan resultNode, maxPaths)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ---- Determine worker count based on tier ----
	workerCount := 2 + parentTier*2
	if workerCount > 16 {
		workerCount = 16
	}

	var wg sync.WaitGroup

	// ---- Worker function ----
	worker := func() {
		defer wg.Done()
		for j := range jobs {
			pair := j.pair
			tierA := tiers[pair[0]]
			tierB := tiers[pair[1]]

			if tierA >= parentTier || tierB >= parentTier {
				continue
			}

			cache := make(map[string]*ElementNode)
			left := resolveToBase(recipes, tiers, pair[0], parentTier, cache)
			right := resolveToBase(recipes, tiers, pair[1], parentTier, cache)

			if left == nil || right == nil {
				continue
			}

			select {
			case resultsChan <- resultNode{
				tree: ElementNode{
					Result:   target,
					Sources:  pair,
					Children: []*ElementNode{left, right},
				},
			}:
			case <-ctx.Done():
				return
			}
		}
	}

	// ---- Start workers ----
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	// ---- Send jobs ----
	go func() {
		for _, pair := range combos {
			jobs <- job{pair: pair}
		}
		close(jobs)
	}()

	// ---- Collect results ----
	var results []ElementNode
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for node := range resultsChan {
		results = append(results, node.tree)
		if len(results) >= maxPaths {
			cancel()
			break
		}
	}

	return BfsResult{
		TargetElement: target,
		RecipeTree:    results,
	}
}

func MainBfs(recipes RecipeMap, tiers TierMap, target string, maxPaths int) BfsResult {
	startTime := time.Now()
	bfsResult := FindNPathsBFS(recipes, tiers, target, maxPaths)
	bfsResult.SearchTime = float64(time.Since(startTime).Milliseconds())

	totalNodes := 0
	for _, tree := range bfsResult.RecipeTree {
		totalNodes += countNodes(&tree)
	}
	bfsResult.VisitedNodes = totalNodes

	return bfsResult
}