package cmd

import (
	"time"
)

type pair [2]string

func buildInverseRecipeMap(recipes RecipeMap) map[string][]pair {
	inv := make(map[string][]pair)
	for result, combos := range recipes {
		for _, c := range combos {
			inv[result] = append(inv[result], pair{c[0], c[1]})
		}
	}
	return inv
}

func combineElements(known map[string]*ElementNode, recipes map[string][]pair) map[string]*ElementNode {
	newElements := make(map[string]*ElementNode)

	for e1, n1 := range known {
		for e2, n2 := range known {
			if e1 > e2 {
				continue 
			}

			for result, pairs := range recipes {
				for _, p := range pairs {
					if (p[0] == e1 && p[1] == e2) || (p[0] == e2 && p[1] == e1) {
						if _, exists := known[result]; !exists {
							newElements[result] = &ElementNode{
								Result:   result,
								Sources:  []string{e1, e2},
								Children: []*ElementNode{n1, n2},
							}
						}
					}
				}
			}
		}
	}
	return newElements
}

func resolveBackward(
	recipes RecipeMap,
	tiers TierMap,
	target string,
	visited map[string]*ElementNode,
) {
	if _, ok := visited[target]; ok || isBase(target) {
		if isBase(target) {
			visited[target] = &ElementNode{Result: target}
		}
		return
	}

	for _, pair := range recipes[target] {
		resolveBackward(recipes, tiers, pair[0], visited)
		resolveBackward(recipes, tiers, pair[1], visited)

		left := visited[pair[0]]
		right := visited[pair[1]]

		if left != nil && right != nil {
			visited[target] = &ElementNode{
				Result:   target,
				Sources:  pair,
				Children: []*ElementNode{left, right},
			}
			return
		}
	}
}

func mergeTrees(mid string, forward *ElementNode, backward *ElementNode) *ElementNode {
	// mid is the connecting node
	// backward will include target and up, forward includes base and up
	if forward == nil {
		return backward
	}
	if backward == nil {
		return forward
	}
	return &ElementNode{
		Result:   mid,
		Sources:  forward.Sources,
		Children: forward.Children,
	}
}

func BidirectionalBfs(
	recipes RecipeMap,
	tiers TierMap,
	target string,
	maxPaths int,
) Result {
	start := time.Now()

	forwardVisited := map[string]*ElementNode{}
	backwardVisited := map[string]*ElementNode{}

	queue := []string{}
	for base := range abaseElements {
		forwardVisited[base] = &ElementNode{Result: base}
		queue = append(queue, base)
	}

	resolveBackward(recipes, tiers, target, backwardVisited)

	if _, ok := forwardVisited[target]; ok {
		return Result{
			TargetElement: target,
			RecipeTree:    []ElementNode{{Result: target}},
			VisitedNodes:  1,
			SearchTime:    float64(time.Since(start).Milliseconds()),
		}
	}

	inv := buildInverseRecipeMap(recipes)

	found := []ElementNode{}

	for len(queue) > 0 && len(found) < maxPaths {
		currentKnown := make(map[string]*ElementNode)
		for _, e := range queue {
			currentKnown[e] = forwardVisited[e]
		}
		queue = nil

		newElements := combineElements(currentKnown, inv)
		for k, v := range newElements {
			if _, ok := forwardVisited[k]; ok {
				continue
			}
			forwardVisited[k] = v
			queue = append(queue, k)

			if backNode, ok := backwardVisited[k]; ok {
				merged := mergeTrees(k, v, backNode)
				found = append(found, *merged)
				if len(found) >= maxPaths {
					break
				}
			}
		}
	}

	// Count unique nodes
	visited := make(map[string]bool)
	count := 0
	for _, tree := range found {
		count += countUniqueNodes(&tree, visited)
	}

	return Result{
		TargetElement: target,
		RecipeTree:    found,
		VisitedNodes:  count,
		SearchTime:    float64(time.Since(start).Milliseconds()),
	}
}

func countUniqueNodes(node *ElementNode, visited map[string]bool) int {
	if node == nil || visited[node.Result] {
		return 0
	}
	visited[node.Result] = true
	count := 1
	for _, child := range node.Children {
		count += countUniqueNodes(child, visited)
	}
	return count
}
