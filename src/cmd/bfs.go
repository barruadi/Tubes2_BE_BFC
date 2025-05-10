package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ---------------- Data Structures ----------------

type ElementNode struct {
	Result   	string        `json:"result"`
	Children 	[]ElementNode `json:"children,omitempty"`
}

type RecipeMap  map[string][][]string

type bfsNode struct {
	Current 	string
	Tree    	ElementNode
	Visited 	map[string]bool
}

var baseElements = map[string]bool{
	"water"    : true,
	"fire"     : true,
	"earth"    : true,
	"air"      : true,
}

func isBase(e string) bool {
	return baseElements[e]
}

func copyMap(original map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range original {
		newMap[k] = v
	}
	return newMap
}

func writeAndPrint(data interface{}, filename string) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error encoding result:", err)
		return
	}
	fmt.Println("\nResult:")
	fmt.Println(string(jsonBytes))
	if err := os.WriteFile(filename, jsonBytes, 0644); err != nil {
		fmt.Println("Error writing file:", err)
	}
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

func findNPathsBFS(recipes RecipeMap, target string, maxPaths int) []ElementNode {
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

	return results
}

// ---------------- Main ----------------

func main() {
	data, err := os.ReadFile("./data/alchemy_recipes.json")
	if err != nil {
		fmt.Println("Error reading recipes.json:", err)
		return
	}

	var recipes RecipeMap
	if err := json.Unmarshal(data, &recipes); err != nil {
		fmt.Println("Invalid JSON:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter target element: ")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)

	fmt.Print("Choose mode ('1. shortest', '2. multiple-recipe'): ")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	switch mode {
	case "1":
		result := resolveToBase(recipes, target)
		if result == nil {
			fmt.Println("Not Found.")
			return
		}
		writeAndPrint(result, "result_tree.json")

	case "2":
		fmt.Print("Berapa: ")
		nStr, _ := reader.ReadString('\n')
		nStr = strings.TrimSpace(nStr)
		n, err := strconv.Atoi(nStr)
		if err != nil || n <= 0 {
			fmt.Println("Invalid number.")
			return
		}

		var results []ElementNode
		results = findNPathsBFS(recipes, target, n)

		if len(results) == 0 {
			fmt.Println("Not Found")
			return
		}
		writeAndPrint(results, "result_tree.json")

	default:
		fmt.Println("Invalid")
	}
}
