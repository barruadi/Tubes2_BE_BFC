package cmd

import (
	"fmt"
	"time"
)

// ---------------- Data Structures ----------------

type bfsNode struct {
	Current 	string
	Tree    	ElementNode
	Visited 	map[string]bool
}

var abaseElements = map[string]bool{
	"water"    : true,
	"fire"     : true,
	"earth"    : true,
	"air"      : true,
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

// ---------------- Helper ----------------

func resolveToBase(recipes RecipeMap, target string) *ElementNode {
	root := bfsShortestPath(recipes, target)
	if root == nil {
		return nil
	}
	expandTree(recipes, root)
	return root
}

func expandTree(recipes RecipeMap, node *ElementNode) {
	if isBase(node.Result) {
		return
	}

	if len(node.Children) == 0 {
		combinations, exists := recipes[node.Result]
		if !exists {
			return
		}

		for _, pair := range combinations {
			left := resolveToBase(recipes, pair[0])
			right := resolveToBase(recipes, pair[1])
			if left != nil && right != nil {
				node.Children = []ElementNode{
					{
						Result:   fmt.Sprintf("%s + %s", pair[0], pair[1]),
						Children: []ElementNode{*left, *right},
					},
				}
				break
			}
		}
	} else {
		for i := range node.Children {
			expandTree(recipes, &node.Children[i])
		}
	}
}

// ---------------- Shortest Path  ----------------

func bfsShortestPath(recipes RecipeMap, target string) *ElementNode {
	queue := []bfsNode{{
		Current: target,
		Tree:    ElementNode{Result: target},
		Visited: map[string]bool{target: true},
	}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if isBase(curr.Current) {
			return &curr.Tree
		}

		combos, exists := recipes[curr.Current]
		if !exists {
			continue
		}

		for _, pair := range combos {
			if curr.Visited[pair[0]] || curr.Visited[pair[1]] {
				continue
			}

			left := ElementNode{Result: pair[0]}
			right := ElementNode{Result: pair[1]}
			recipeStep := ElementNode{
				Result:   fmt.Sprintf("%s + %s", pair[0], pair[1]),
				Children: []ElementNode{left, right},
			}

			newTree := ElementNode{
				Result:   curr.Current,
				Children: []ElementNode{recipeStep},
			}

			newVisited := copyMap(curr.Visited)
			newVisited[pair[0]] = true
			newVisited[pair[1]] = true

			if isBase(pair[0]) && isBase(pair[1]) {
				return &newTree
			}

			queue = append(queue, bfsNode{Current: pair[0], Tree: newTree, Visited: newVisited})
			queue = append(queue, bfsNode{Current: pair[1], Tree: newTree, Visited: newVisited})
		}
	}
	return nil
}

// ---------------- N Recipes ----------------

func FindNPathsBFS(recipes RecipeMap, target string, maxPaths int) BfsResult {
	type state struct {
		current string
		visited map[string]bool
	}

	var results []ElementNode

	queue := []state{{
		current: target,
		visited: map[string]bool{target: true},
	}}

	for len(queue) > 0 && len(results) < maxPaths {
		curr := queue[0]
		queue = queue[1:]

		combos, exists := recipes[curr.current]
		if !exists {
			continue
		}

		for _, pair := range combos {
			if curr.visited[pair[0]] || curr.visited[pair[1]] {
				continue
			}

			left := resolveToBase(recipes, pair[0])
			right := resolveToBase(recipes, pair[1])
			if left == nil || right == nil {
				continue
			}

			step := ElementNode{
				Result:   fmt.Sprintf("%s + %s", pair[0], pair[1]),
				Children: []ElementNode{*left, *right},
			}

			tree := ElementNode{
				Result:   target,
				Children: []ElementNode{step},
			}

			results = append(results, tree)
			if len(results) >= maxPaths {
				break
			}

			newVisited := copyMap(curr.visited)
			newVisited[pair[0]] = true
			newVisited[pair[1]] = true

			queue = append(queue, state{current: pair[0], visited: newVisited})
			queue = append(queue, state{current: pair[1], visited: newVisited})
		}
	}

	var finalResult = BfsResult{target, results, 10, 0}

	return finalResult
}

// ---------------- Main ----------------

func MainBfs(recipes RecipeMap, target string, maxPaths int) BfsResult {
	startTime := time.Now()
	var bfsResult BfsResult = FindNPathsBFS(recipes, target, maxPaths)
	searchTime := time.Since(startTime).Milliseconds()
	bfsResult.SearchTime = float64(searchTime)

	return  bfsResult
}
