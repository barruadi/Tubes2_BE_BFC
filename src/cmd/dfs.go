// package cmd

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"sort"
// 	"strings"
// 	"sync"
// 	"sync/atomic"
// 	"time"
// )


// var (
// 	recipesDB        AlchemyRecipes
// 	baseElements     = []string{"air", "earth", "fire", "water"}
// 	visitedNodeCount int64
// )






// func incrementNodeCount() {
// 	atomic.AddInt64(&visitedNodeCount, 1)
// }

// func recipeToKey(recipe []string) string {
// 	sortedRecipe := make([]string, len(recipe))
// 	copy(sortedRecipe, recipe)
// 	sort.Strings(sortedRecipe)
// 	return strings.Join(sortedRecipe, "|||")
// }

// func addRecipeToResults(recipe []string, childNodes []*ElementNode, ctx RecipeContextDFS) bool {
// 	if atomic.LoadInt32(ctx.RecipeCount) >= ctx.MaxCount {
// 		return false
// 	}
// 	rootNode := ElementNode{
// 		Result:   ctx.TargetElement,
// 		Sources:  recipe,
// 		Children: childNodes,
// 	}
	
// 	atomic.AddInt32(ctx.RecipeCount, 1)
// 	select {
// 	case <-ctx.Done:
// 		return false
// 	case ctx.ResultChan <- rootNode:
// 		return true
// 	default:
// 		return false
// 	}
// }


// func buildRecipeTreeNode(element string, visited map[string]bool) ElementNode {
// 	incrementNodeCount()
// 	if visited[element] {
// 		return ElementNode{Result: element}
// 	}
	
// 	if isBaseElementElement(element) {
// 		return ElementNode{Result: element}
// 	}

// 	newVisited := make(map[string]bool)
// 	for k, v := range visited {
// 		newVisited[k] = v
// 	}
// 	newVisited[element] = true

// 	recipes := getDirectRecipes(element)
// 	if len(recipes) == 0 {
// 		return ElementNode{Result: element}
// 	}
// 	for _, recipe := range recipes {
// 		if !isValidTierCombination(recipe, element) {
// 			continue
// 		}
// 		allValid := true
// 		var childNodes []*ElementNode
// 		for _, ingredient := range recipe {
// 			childNode := buildRecipeTreeNode(ingredient, newVisited)
			
// 			if len(childNode.Children) == 0 && !isBaseElementElement(childNode.Result) {
// 				allValid = false
// 				break
// 			}
// 			childNodeCopy := childNode
// 			childNodes = append(childNodes, &childNodeCopy)
// 		}
		
// 		if allValid {
// 			return ElementNode{
// 				Result:   element,
// 				Sources:  recipe,
// 				Children: childNodes,
// 			}
// 		}
// 	}
	
// 	return ElementNode{Result: element}
// }


// func processRecipe(recipe []string, element string, seen map[string]bool, ctx RecipeContextDFS) {

// 	if atomic.LoadInt32(ctx.RecipeCount) >= ctx.MaxCount {
// 		return
// 	}
	
// 	select {
// 	case <-ctx.Done:
// 		return
// 	default:
		
// 	}
	
// 	if !isValidTierCombination(recipe, element) {
// 		return
// 	}
	
// 	recipeKey := recipeToKey(recipe)
	
// 	ctx.SeenMutex.Lock()
// 	seenBefore := seen[recipeKey]
// 	if !seenBefore {
// 		seen[recipeKey] = true
// 	}
// 	ctx.SeenMutex.Unlock()
	
// 	if seenBefore {
// 		return
// 	}
	

// 	var childNodes []*ElementNode
// 	allValid := true
	
// 	for _, ingredient := range recipe {
	
// 		visitedMap := make(map[string]bool)
// 		childNode := buildRecipeTreeNode(ingredient, visitedMap)
		

// 		if len(childNode.Children) == 0 && !isBaseElementElement(childNode.Result) {
// 			allValid = false
// 			break
// 		}
// 		childNodeCopy := childNode
// 		childNodes = append(childNodes, &childNodeCopy)
// 	}

// 	if allValid {
// 		addRecipeToResults(recipe, childNodes, ctx)
// 	}
// }

// func exploreDFSPaths(element string, ctx RecipeContextDFS, seen map[string]bool) {
// 	if atomic.LoadInt32(ctx.RecipeCount) >= ctx.MaxCount {
// 		return
// 	}
	
// 	select {
// 	case <-ctx.Done:
// 		return
// 	default:
		
// 	}
	

// 	recipes := getDirectRecipes(element)

// 	if len(recipes) == 0 {
// 		return
// 	}
	

// 	recipeChan := make(chan []string, len(recipes))
	

// 	numRecipes := len(recipes)
// 	numWorkers := ctx.MaxWorkers
// 	if numRecipes < numWorkers {
// 		numWorkers = numRecipes
// 	}
// 	if numWorkers < 1 {
// 		numWorkers = 1
// 	}
	

// 	var wg sync.WaitGroup
	

// 	for i := 0; i < numWorkers; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
			
// 			for recipe := range recipeChan {
// 				if atomic.LoadInt32(ctx.RecipeCount) >= ctx.MaxCount {
// 					return
// 				}
				
// 				select {
// 				case <-ctx.Done:
// 					return
// 				default:

// 					processRecipe(recipe, element, seen, ctx)
// 				}
// 			}
// 		}()
// 	}
	
// 	recipeLoop:
// 	for _, recipe := range recipes {
		
// 		if atomic.LoadInt32(ctx.RecipeCount) >= ctx.MaxCount {
// 			break
// 		}
		
// 		select {
// 		case <-ctx.Done:
// 			break recipeLoop
// 		case recipeChan <- recipe:
// 		}
// 	}
	
	
// 	close(recipeChan)

// 	wg.Wait()
// }

// func LoadRecipesFromFile(filePath string) error {
// 	data, err := ioutil.ReadFile(filePath)
// 	if err != nil {
// 		return fmt.Errorf("error reading file: %v", err)
// 	}

// 	err = json.Unmarshal(data, &recipesDB)
// 	if err != nil {
// 		return fmt.Errorf("error parsing JSON: %v", err)
// 	}
	
// 	return nil
// }


// func GetVisitedNodeCount() int64 {
// 	return atomic.LoadInt64(&visitedNodeCount)
// }


// func DFSSingle(targetElement string) DfsResult {
// 	startTime := time.Now()

// 	atomic.StoreInt64(&visitedNodeCount, 0)
	

// 	if isBaseElementElement(targetElement) {
// 		return DfsResult{
// 			TargetElement: targetElement,
// 			RecipeTree:    []ElementNode{{Result: targetElement}},
// 			VisitedNodes:  1,
// 			SearchTime:    float64(time.Since(startTime).Milliseconds()),
// 		}
// 	}
	
	
// 	rootNode := buildRecipeTreeNode(targetElement, make(map[string]bool))
	
// 	return DfsResult{
// 		TargetElement: targetElement,
// 		RecipeTree:    []ElementNode{rootNode},
// 		VisitedNodes:  int(atomic.LoadInt64(&visitedNodeCount)),
// 		SearchTime:    float64(time.Since(startTime).Milliseconds()),
// 	}
// }


// func DFSMultiple(targetElement string, maxRecipes int) DfsResult {
// 	startTime := time.Now()
	
// 	atomic.StoreInt64(&visitedNodeCount, 0)
	

// 	if isBaseElementElement(targetElement) {
// 		return DfsResult{
// 			TargetElement: targetElement,
// 			RecipeTree:    []ElementNode{{Result: targetElement}},
// 			VisitedNodes:  1,
// 			SearchTime:    float64(time.Since(startTime).Milliseconds()),
// 		}
// 	}

// 	var results []ElementNode

// 	resultChan := make(chan ElementNode, maxRecipes)

// 	done := make(chan struct{})
	
// 	var doneIsClosed bool
// 	var doneCloseMutex sync.Mutex

// 	safeCloseDone := func() {
// 		doneCloseMutex.Lock()
// 		defer doneCloseMutex.Unlock()
		
// 		if !doneIsClosed {
// 			close(done)
// 			doneIsClosed = true
// 		}
// 	}
	

// 	defer safeCloseDone()
	

// 	var recipeCount int32 = 0
// 	var seenMutex sync.Mutex
	
// 	ctx := RecipeContextDFS{
// 		RecipeCount:   &recipeCount,
// 		MaxCount:      int32(maxRecipes),
// 		ResultChan:    resultChan,
// 		Done:          done,
// 		TargetElement: targetElement,
// 		SeenMutex:     &seenMutex,
// 		MaxWorkers:    100, // Jumlah worker yang diinginkan, bisa disesuaikan
// 	}
	
	
// 	go func() {
// 		seen := make(map[string]bool) // Untuk deduplikasi
// 		exploreDFSPaths(targetElement, ctx, seen)
// 		close(resultChan) // Tutup channel hasil setelah selesai
// 	}()
	

// 	for rootNode := range resultChan {
// 		results = append(results, rootNode)
// 		if len(results) >= maxRecipes {
// 			safeCloseDone() 
// 			break
// 		}
// 	}
	
// 	return DfsResult{
// 		TargetElement: targetElement,
// 		RecipeTree:    results,
// 		VisitedNodes:  int(atomic.LoadInt64(&visitedNodeCount)),
// 		SearchTime:    float64(time.Since(startTime).Milliseconds()),
// 	}
// }


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


func flattenTreeList(trees []*ElementNode) []ElementNode {
	result := make([]ElementNode, len(trees))
	for i, t := range trees {
		result[i] = *t
	}
	return result
}


func isUnbuildable(e string, recipes RecipeMap) bool {
	return !isBaseElement(e) && len(recipes[e]) == 0
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


func MainDfs(recipes RecipeMap, tiers TierMap, targetElement string, maxRecipes int) DfsResult {

	atomic.StoreInt64(&visitedNodeCount, 0)

	startTime := time.Now()

	if isBaseElement(targetElement) {
		return DfsResult{
			TargetElement: targetElement,
			RecipeTree:    []ElementNode{{Result: targetElement}},
			VisitedNodes:  1,
			SearchTime:    0,
		}
	}
	
	trees := dfsBuildTree(recipes, tiers, targetElement, maxRecipes, make(map[string]bool), 0, 15)

	searchTime := float64(time.Since(startTime).Milliseconds())
	

	result := DfsResult{
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
