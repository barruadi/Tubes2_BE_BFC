package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)


type Node struct {
	Element    string   // Nama elemen
	Components []*Node  // Elemen-elemen yang dikombinasikan untuk membuat elemen ini
	Depth      int      // Kedalaman dari root
}

// DFSResult menyimpan hasil pencarian DFS
type DFSResult struct {
	TargetElement string
	RecipeTree    *Node
	VisitedNodes  int
	SearchTime    float64
	Found         bool
	Steps         int 
}

type RecipeData map[string][][]string

// Daftar elemen dasar
var baseElements = map[string]bool{
	"air":   true,
	"earth": true,
	"fire":  true,
	"water": true,
}

// DFS mencari satu recipe untuk elemen target
func DFS(targetElement string, recipeData map[string][][]string) DFSResult {
	startTime := time.Now()
	
	// Buat root node
	root := &Node{Element: targetElement, Depth: 0}
	
	// Jika target adalah elemen dasar, kembalikan langsung
	if baseElements[targetElement] {
		return DFSResult{
			TargetElement: targetElement,
			RecipeTree:    root,
			VisitedNodes:  1,
			SearchTime:    float64(time.Since(startTime).Milliseconds()),
			Found:         true,
			Steps:         0,
		}
	}
	
	// Jika tidak ada recipe untuk target element
	recipes, exists := recipeData[targetElement]
	if !exists || len(recipes) == 0 {
		return DFSResult{
			TargetElement: targetElement,
			RecipeTree:    root,
			VisitedNodes:  1,
			SearchTime:    float64(time.Since(startTime).Milliseconds()),
			Found:         false,
			Steps:         0,
		}
	}
	
	// Stack untuk DFS (implementasi iteratif DFS)
	type StackItem struct {
		Node       *Node
		Path       []string
		Visited    map[string]bool
	}
	
	var stack []StackItem
	visitedCount := 1 // Mulai dari 1 untuk root
	
	// Coba setiap recipe untuk target element
	for _, recipe := range recipes {
		// Buat nodes untuk elemen dalam recipe
		child1 := &Node{Element: recipe[0], Depth: 1}
		child2 := &Node{Element: recipe[1], Depth: 1}
		
		// Inisialisasi visited map
		visited := make(map[string]bool)
		visited[targetElement] = true
		
		// Cek apakah kedua elemen adalah elemen dasar
		if baseElements[recipe[0]] && baseElements[recipe[1]] {
			// Jika kedua elemen adalah elemen dasar, kita temukan recipe
			root.Components = []*Node{child1, child2}
			return DFSResult{
				TargetElement: targetElement,
				RecipeTree:    root,
				VisitedNodes:  3, // Root + 2 basic elements
				SearchTime:    float64(time.Since(startTime).Milliseconds()),
				Found:         true,
				Steps:         1,
			}
		}
		
		// Tambahkan ke stack untuk ditelusuri
		if !baseElements[recipe[0]] {
			stack = append(stack, StackItem{
				Node:       child1,
				Path:       []string{targetElement, recipe[0]},
				Visited:    visited,
			})
		}
		
		if !baseElements[recipe[1]] {
			stack = append(stack, StackItem{
				Node:       child2,
				Path:       []string{targetElement, recipe[1]},
				Visited:    visited,
			})
		}
	}
	
	// Lakukan DFS iteratif
	for len(stack) > 0 {
		// Pop dari stack
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		
		// Increment visitedCount
		visitedCount++
		
		// Skip jika sudah dikunjungi
		if current.Visited[current.Node.Element] {
			continue
		}
		
		// Tandai sebagai dikunjungi
		current.Visited[current.Node.Element] = true
		
		// Cari recipe untuk elemen ini
		elementRecipes, exists := recipeData[current.Node.Element]
		if !exists || len(elementRecipes) == 0 {
			continue
		}
		
		// Coba setiap recipe
		for _, elementRecipe := range elementRecipes {
			// Buat nodes untuk elemen dalam recipe
			elementChild1 := &Node{Element: elementRecipe[0], Depth: current.Node.Depth + 1}
			elementChild2 := &Node{Element: elementRecipe[1], Depth: current.Node.Depth + 1}
			
			// Cek apakah kedua elemen adalah elemen dasar
			if baseElements[elementRecipe[0]] && baseElements[elementRecipe[1]] {
				// Jika kedua elemen adalah elemen dasar, kita temukan recipe untuk current node
				current.Node.Components = []*Node{elementChild1, elementChild2}
				
				// Trace back ke root untuk membangun recipe tree
				if len(current.Path) > 1 {
					// Cari parent dari current node
					parent := current.Path[len(current.Path)-2]
					
					// Cari recipe untuk parent ke current
					for _, recipeItem := range recipeData[parent] {
						if recipeItem[0] == current.Node.Element || recipeItem[1] == current.Node.Element {
							// Temukan recipe, bangun tree
							// Untuk sederhananya kita asumsikan parent ada di root
							for _, component := range root.Components {
								if component.Element == parent {
									component.Components = []*Node{elementChild1, elementChild2}
									break
								}
							}
							break
						}
					}
				}
				
				// Buat recipe tree
				found := true
				for i := len(current.Path) - 1; i > 0; i-- {
					// Cek apakah ada recipe untuk elemen ini
					parentElement := current.Path[i-1]
					childElement := current.Path[i]
					
					// Cari recipe
					recipes, exists := recipeData[parentElement]
					if !exists {
						found = false
						break
					}
					
					// Cari recipe yang menggunakan childElement
					elementFound := false
					for _, recipe := range recipes {
						if recipe[0] == childElement || recipe[1] == childElement {
							elementFound = true
							break
						}
					}
					
					if !elementFound {
						found = false
						break
					}
				}
				
				if found {
					// Kita temukan recipe lengkap
					return DFSResult{
						TargetElement: targetElement,
						RecipeTree:    root,
						VisitedNodes:  visitedCount,
						SearchTime:    float64(time.Since(startTime).Milliseconds()),
						Found:         true,
						Steps:         calculateSteps(root),
					}
				}
			}
			
			// Jika tidak keduanya elemen dasar, tambahkan ke stack untuk ditelusuri
			if !baseElements[elementRecipe[0]] && !current.Visited[elementRecipe[0]] {
				// Clone visited map
				visitedCopy := make(map[string]bool)
				for k, v := range current.Visited {
					visitedCopy[k] = v
				}
				
				// Clone path
				pathCopy := make([]string, len(current.Path))
				copy(pathCopy, current.Path)
				pathCopy = append(pathCopy, elementRecipe[0])
				
				stack = append(stack, StackItem{
					Node:       elementChild1,
					Path:       pathCopy,
					Visited:    visitedCopy,
				})
			}
			
			if !baseElements[elementRecipe[1]] && !current.Visited[elementRecipe[1]] {
				// Clone visited map
				visitedCopy := make(map[string]bool)
				for k, v := range current.Visited {
					visitedCopy[k] = v
				}
				
				// Clone path
				pathCopy := make([]string, len(current.Path))
				copy(pathCopy, current.Path)
				pathCopy = append(pathCopy, elementRecipe[1])
				
				stack = append(stack, StackItem{
					Node:       elementChild2,
					Path:       pathCopy,
					Visited:    visitedCopy,
				})
			}
		}
	}
	
	// Jika sampai di sini, berarti tidak menemukan recipe
	return DFSResult{
		TargetElement: targetElement,
		RecipeTree:    root,
		VisitedNodes:  visitedCount,
		SearchTime:    float64(time.Since(startTime).Milliseconds()),
		Found:         false,
		Steps:         0,
	}
}

// FindMultipleRecipesDFS mencari beberapa recipe dengan multithreading
func FindMultipleRecipesDFS(targetElement string, recipeData map[string][][]string, maxRecipes int) []DFSResult {
	// Jika target adalah elemen dasar, return langsung
	if baseElements[targetElement] {
		root := &Node{Element: targetElement, Depth: 0}
		return []DFSResult{{
			TargetElement: targetElement,
			RecipeTree:    root,
			VisitedNodes:  1,
			SearchTime:    0,
			Found:         true,
			Steps:         0,
		}}
	}
	
	// Dapatkan semua kombinasi untuk elemen target
	recipes, exists := recipeData[targetElement]
	if !exists || len(recipes) == 0 {
		return []DFSResult{}
	}
	
	// Batasi maxRecipes ke jumlah kombinasi yang tersedia
	// if maxRecipes > len(recipes) {
	// 	maxRecipes = len(recipes)
	// }
	
	// Hasil dan mutex untuk sinkronisasi
	var results []DFSResult
	var resultsMutex sync.Mutex
	
	// Gunakan channel untuk mengumpulkan hasil
	resultChan := make(chan DFSResult, maxRecipes)
	
	// Gunakan WaitGroup untuk sinkronisasi
	var wg sync.WaitGroup
	
	// Batasi jumlah goroutines yang berjalan secara bersamaan
	maxGoroutines := min(len(recipes), 10)
	semaphore := make(chan struct{}, maxGoroutines)
	
	// Luncurkan goroutine untuk setiap kombinasi
	for i, combo := range recipes {
		if i >= maxRecipes {
			break
		}
		
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		
		go func(combination []string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore
			
			startTime := time.Now()
			
			// Buat root node
			root := &Node{Element: targetElement, Depth: 0}
			
			// Buat nodes untuk elemen dalam kombinasi
			child1 := &Node{Element: combination[0], Depth: 1}
			child2 := &Node{Element: combination[1], Depth: 1}
			
			// Set components pada root
			root.Components = []*Node{child1, child2}
			
			// Lakukan DFS untuk kedua elemen
			visitedCount := 1 // Mulai dari 1 untuk root
			
			// Stack untuk DFS (implementasi iteratif DFS)
			type StackItem struct {
				Node       *Node
				Path       []string
				Visited    map[string]bool
			}
			
			// Flag untuk menandakan apakah kedua child berhasil
			found1 := false
			found2 := false
			
			// Cek apakah child1 adalah elemen dasar
			if baseElements[combination[0]] {
				found1 = true
			} else {
				// Lakukan DFS untuk child1
				var stack1 []StackItem
				visited1 := make(map[string]bool)
				visited1[targetElement] = true
				
				stack1 = append(stack1, StackItem{
					Node:       child1,
					Path:       []string{targetElement, combination[0]},
					Visited:    visited1,
				})
				
				// Lakukan DFS iteratif
				for len(stack1) > 0 && !found1 {
					// Pop dari stack
					current := stack1[len(stack1)-1]
					stack1 = stack1[:len(stack1)-1]
					
					// Increment visitedCount
					visitedCount++
					
					// Skip jika sudah dikunjungi
					if current.Visited[current.Node.Element] {
						continue
					}
					
					// Tandai sebagai dikunjungi
					current.Visited[current.Node.Element] = true
					
					// Cari recipe untuk elemen ini
					elementRecipes, exists := recipeData[current.Node.Element]
					if !exists || len(elementRecipes) == 0 {
						continue
					}
					
					// Coba setiap recipe
					for _, elementRecipe := range elementRecipes {
						// Buat nodes untuk elemen dalam recipe
						elementChild1 := &Node{Element: elementRecipe[0], Depth: current.Node.Depth + 1}
						elementChild2 := &Node{Element: elementRecipe[1], Depth: current.Node.Depth + 1}
						
						// Cek apakah kedua elemen adalah elemen dasar
						if baseElements[elementRecipe[0]] && baseElements[elementRecipe[1]] {
							// Jika kedua elemen adalah elemen dasar, kita temukan recipe untuk current node
							current.Node.Components = []*Node{elementChild1, elementChild2}
							
							// Update child1 dengan recipe yang ditemukan
							child1.Components = []*Node{elementChild1, elementChild2}
							found1 = true
							break
						}
						
						// Jika tidak keduanya elemen dasar, tambahkan ke stack untuk ditelusuri
						if !baseElements[elementRecipe[0]] && !current.Visited[elementRecipe[0]] {
							// Clone visited map
							visitedCopy := make(map[string]bool)
							for k, v := range current.Visited {
								visitedCopy[k] = v
							}
							
							// Clone path
							pathCopy := make([]string, len(current.Path))
							copy(pathCopy, current.Path)
							pathCopy = append(pathCopy, elementRecipe[0])
							
							stack1 = append(stack1, StackItem{
								Node:       elementChild1,
								Path:       pathCopy,
								Visited:    visitedCopy,
							})
						}
						
						if !baseElements[elementRecipe[1]] && !current.Visited[elementRecipe[1]] {
							// Clone visited map
							visitedCopy := make(map[string]bool)
							for k, v := range current.Visited {
								visitedCopy[k] = v
							}
							
							// Clone path
							pathCopy := make([]string, len(current.Path))
							copy(pathCopy, current.Path)
							pathCopy = append(pathCopy, elementRecipe[1])
							
							stack1 = append(stack1, StackItem{
								Node:       elementChild2,
								Path:       pathCopy,
								Visited:    visitedCopy,
							})
						}
					}
					
					if found1 {
						break
					}
				}
			}
			
			// Cek apakah child2 adalah elemen dasar
			if baseElements[combination[1]] {
				found2 = true
			} else {
				// Lakukan DFS untuk child2
				var stack2 []StackItem
				visited2 := make(map[string]bool)
				visited2[targetElement] = true
				
				stack2 = append(stack2, StackItem{
					Node:       child2,
					Path:       []string{targetElement, combination[1]},
					Visited:    visited2,
				})
				
				// Lakukan DFS iteratif
				for len(stack2) > 0 && !found2 {
					// Pop dari stack
					current := stack2[len(stack2)-1]
					stack2 = stack2[:len(stack2)-1]
					
					// Increment visitedCount
					visitedCount++
					
					// Skip jika sudah dikunjungi
					if current.Visited[current.Node.Element] {
						continue
					}
					
					// Tandai sebagai dikunjungi
					current.Visited[current.Node.Element] = true
					
					// Cari recipe untuk elemen ini
					elementRecipes, exists := recipeData[current.Node.Element]
					if !exists || len(elementRecipes) == 0 {
						continue
					}
					
					// Coba setiap recipe
					for _, elementRecipe := range elementRecipes {
						// Buat nodes untuk elemen dalam recipe
						elementChild1 := &Node{Element: elementRecipe[0], Depth: current.Node.Depth + 1}
						elementChild2 := &Node{Element: elementRecipe[1], Depth: current.Node.Depth + 1}
						
						// Cek apakah kedua elemen adalah elemen dasar
						if baseElements[elementRecipe[0]] && baseElements[elementRecipe[1]] {
							// Jika kedua elemen adalah elemen dasar, kita temukan recipe untuk current node
							current.Node.Components = []*Node{elementChild1, elementChild2}
							
							// Update child2 dengan recipe yang ditemukan
							child2.Components = []*Node{elementChild1, elementChild2}
							found2 = true
							break
						}
						
						// Jika tidak keduanya elemen dasar, tambahkan ke stack untuk ditelusuri
						if !baseElements[elementRecipe[0]] && !current.Visited[elementRecipe[0]] {
							// Clone visited map
							visitedCopy := make(map[string]bool)
							for k, v := range current.Visited {
								visitedCopy[k] = v
							}
							
							// Clone path
							pathCopy := make([]string, len(current.Path))
							copy(pathCopy, current.Path)
							pathCopy = append(pathCopy, elementRecipe[0])
							
							stack2 = append(stack2, StackItem{
								Node:       elementChild1,
								Path:       pathCopy,
								Visited:    visitedCopy,
							})
						}
						
						if !baseElements[elementRecipe[1]] && !current.Visited[elementRecipe[1]] {
							// Clone visited map
							visitedCopy := make(map[string]bool)
							for k, v := range current.Visited {
								visitedCopy[k] = v
							}
							
							// Clone path
							pathCopy := make([]string, len(current.Path))
							copy(pathCopy, current.Path)
							pathCopy = append(pathCopy, elementRecipe[1])
							
							stack2 = append(stack2, StackItem{
								Node:       elementChild2,
								Path:       pathCopy,
								Visited:    visitedCopy,
							})
						}
					}
					
					if found2 {
						break
					}
				}
			}
			
			// Jika kedua jalur berhasil, tambahkan ke hasil
			if found1 && found2 {
				result := DFSResult{
					TargetElement: targetElement,
					RecipeTree:    root,
					VisitedNodes:  visitedCount,
					SearchTime:    float64(time.Since(startTime).Milliseconds()),
					Found:         true,
					Steps:         calculateSteps(root),
				}
				
				resultChan <- result
			}
		}(combo)
	}
	
	// Goroutine untuk menutup channel setelah semua pencarian selesai
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Kumpulkan hasil
	for result := range resultChan {
		resultsMutex.Lock()
		results = append(results, result)
		resultsMutex.Unlock()
	}
	
	return results
}

// Helper functions

// calculateSteps menghitung kedalaman maksimum dari tree
func calculateSteps(node *Node) int {
	if node == nil {
		return 0
	}
	
	if node.Components == nil || len(node.Components) == 0 {
		return 0
	}
	
	// Hitung kedalaman maksimum dari children
	depth1 := calculateSteps(node.Components[0])
	depth2 := calculateSteps(node.Components[1])
	
	if depth1 > depth2 {
		return depth1 + 1
	}
	return depth2 + 1
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// printRecipeTree mencetak tree recipe
func printRecipeTree(node *Node, indent string) {
	if node == nil {
		return
	}
	
	// Cetak elemen saat ini
	fmt.Print(indent)
	fmt.Print(node.Element)
	
	// Tandai elemen dasar
	if baseElements[node.Element] {
		fmt.Print(" (base element) ✓")
	}
	fmt.Println()
	
	// Jika tidak memiliki components, berhenti
	if node.Components == nil || len(node.Components) == 0 {
		return
	}
	
	// Cetak children
	for _, child := range node.Components {
		printRecipeTree(child, indent+"  ")
	}
}

func main() {
	// Load data dari file JSON
	jsonData, err := ioutil.ReadFile("src/data/alchemy_recipes.json")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	var recipeData RecipeData
	err = json.Unmarshal(jsonData, &recipeData)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}
	
	// Element yang ingin dicari
	targetElement := "brick"
	
	// 1. Test DFS untuk satu recipe
	fmt.Printf("\n===== TESTING DFS FOR '%s' =====\n", targetElement)
	startTime := time.Now()
	result := DFS(targetElement, recipeData)
	searchTime := time.Since(startTime).Milliseconds()
	
	fmt.Printf("Target Element: %s\n", result.TargetElement)
	fmt.Printf("Found: %v\n", result.Found)
	fmt.Printf("Visited Nodes: %d\n", result.VisitedNodes)
	fmt.Printf("Search Time: %d ms\n", searchTime)
	fmt.Printf("Steps: %d\n", result.Steps)
	
	if result.Found {
		fmt.Println("\n--- RECIPE TREE ---")
		printRecipeTree(result.RecipeTree, "")
	} else {
		fmt.Println("No recipe found.")
	}
	
	// 2. Test FindMultipleRecipesDFS
	maxRecipes := 20 // Jumlah recipe yang ingin dicari
	fmt.Printf("\n===== TESTING MULTIPLE RECIPES (%d) FOR '%s' =====\n", maxRecipes, targetElement)
	startTime = time.Now()
	multipleResults := FindMultipleRecipesDFS(targetElement, recipeData, maxRecipes)
	searchTime = time.Since(startTime).Milliseconds()
	
	fmt.Printf("Found %d recipes\n", len(multipleResults))
	fmt.Printf("Search Time: %d ms\n", searchTime)
	
	// Print setiap recipe tree yang ditemukan
	for i, result := range multipleResults {
		fmt.Printf("\n--- RECIPE %d ---\n", i+1)
		fmt.Printf("Steps: %d\n", result.Steps)
		fmt.Printf("Visited Nodes: %d\n", result.VisitedNodes)
		printRecipeTree(result.RecipeTree, "")
	}
}