package cmd

import (
	"time"
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
	type state struct {
		current string
	}

	var results []ElementNode
	queue := []state{{current: target}}

	for len(queue) > 0 && len(results) < maxPaths {
		curr := queue[0]
		queue = queue[1:]

		combos, exists := recipes[curr.current]
		if !exists {
			continue
		}

		parentTier := tiers[curr.current]

		for _, pair := range combos {
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

			tree := ElementNode{
				Result:   target,
				Sources:  pair,
				Children: []*ElementNode{left, right},
			}

			results = append(results, tree)

			if len(results) >= maxPaths {
				break
			}
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