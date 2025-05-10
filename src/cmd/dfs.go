package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "os"
	"sync"
	// "time"
)

// ElementData menyimpan informasi tier dan resep elemen
type ElementData struct {
	Tier    int        `json:"tier"`
	Recipes [][]string `json:"recipes"`
}

// AlchemyRecipes menyimpan seluruh data resep
type AlchemyRecipes map[string]ElementData

// RecipeNode adalah node dalam tree recipe
type RecipeNode struct {
	Result   string       `json:"result"`
	Children []RecipeNode `json:"children,omitempty"`
}

// Global variables
var (
	recipesDB    AlchemyRecipes
	baseElements = []string{"air", "earth", "fire", "water"}
	NodesVisited int
	nodeMutex    sync.Mutex
)

// isBaseElement memeriksa apakah elemen adalah elemen dasar
func isBaseElement(element string) bool {
	for _, baseElement := range baseElements {
		if element == baseElement {
			return true
		}
	}
	return false
}

// isValidTierCombination memeriksa apakah kombinasi tier valid
func isValidTierCombination(ingredients []string, result string) bool {
	resultTier := recipesDB[result].Tier
	
	for _, ingredient := range ingredients {
		if recipesDB[ingredient].Tier >= resultTier {
			return false
		}
	}
	
	return true
}

// getDirectRecipes mengembalikan semua recipe langsung untuk elemen tertentu
func getDirectRecipes(element string) [][]string {
	if elem, exists := recipesDB[element]; exists {
		return elem.Recipes
	}
	return nil
}

// incrementNodeCount menambah counter node yang dikunjungi
func incrementNodeCount() {
	nodeMutex.Lock()
	NodesVisited++
	nodeMutex.Unlock()
}
// buildRecipeTree membangun tree recipe dari elemen
func buildRecipeTree(element string, visited map[string]bool) RecipeNode {
	incrementNodeCount()
	
	// Mencegah infinite recursion
	if visited[element] {
		return RecipeNode{Result: element}
	}
	
	// Jika elemen dasar, kembalikan langsung
	if isBaseElement(element) {
		return RecipeNode{Result: element}
	}
	
	// Tambahkan elemen ke visited
	newVisited := make(map[string]bool)
	for k, v := range visited {
		newVisited[k] = v
	}
	newVisited[element] = true
	
	// Dapatkan semua recipe untuk elemen ini
	recipes := getDirectRecipes(element)
	
	// Jika tidak ada recipe, kembalikan node tanpa children
	if len(recipes) == 0 {
		return RecipeNode{Result: element}
	}
	
	// Coba setiap recipe sampai menemukan yang valid
	for _, recipe := range recipes {
		// Validasi tier
		if !isValidTierCombination(recipe, element) {
			continue
		}
		
		// Format recipe string
		recipeStr := recipe[0]
		if len(recipe) > 1 {
			recipeStr += " + " + recipe[1]
		}
		
		// Buat node untuk recipe ini
		recipeNode := RecipeNode{
			Result: recipeStr,
		}
		
		// Bangun children untuk setiap ingredient
		allValid := true
		var childNodes []RecipeNode
		
		for _, ingredient := range recipe {
			childNode := buildRecipeTree(ingredient, newVisited) //recursive 
			
			// Jika childNode tidak valid, tandai recipe sebagai tidak valid
			if len(childNode.Children) == 0 && !isBaseElement(childNode.Result) {
				allValid = false
				break
			}
			
			childNodes = append(childNodes, childNode)
		}
		
		// Jika semua ingredient valid, kembalikan recipe tree
		if allValid {
			recipeNode.Children = childNodes
			
			return RecipeNode{
				Result:   element,
				Children: []RecipeNode{recipeNode},
			}
		}
	}
	
	// Jika tidak ada recipe yang valid, kembalikan node tanpa children
	return RecipeNode{Result: element}
}

// DFSSingle mencari satu recipe dengan algoritma DFS
func DFSSingle(targetElement string) RecipeNode {
	// Jika target adalah elemen dasar, kembalikan langsung
	if isBaseElement(targetElement) {
		return RecipeNode{Result: targetElement}
	}
	return buildRecipeTree(targetElement, make(map[string]bool))
}

func DFSMultiple(targetElement string, maxRecipes int) []RecipeNode {
	// Reset counter node yang dikunjungi
	NodesVisited = 0
	// Jika target adalah elemen dasar, kembalikan langsung
		
	if isBaseElement(targetElement) {
		return []RecipeNode{{Result: targetElement}}
	}

	directRecipes := getDirectRecipes(targetElement)
	
	var results []RecipeNode
	seen := make(map[string]bool)
	
	var resultMutex sync.Mutex
	var wg sync.WaitGroup
	
	// Gunakan channel untuk mengumpulkan hasil
	resultChan := make(chan RecipeNode, maxRecipes)
	
	// Buat signal untuk memberitahu goroutine berhenti
	done := make(chan struct{})
	defer close(done)
	
	// Olah setiap recipe secara parallel
	for _, recipe := range directRecipes {
		// Skip jika sudah mencapai maxRecipes
		resultMutex.Lock()
		if len(results) >= maxRecipes {
			resultMutex.Unlock()
			break
		}
		resultMutex.Unlock()
		
		// Validasi tier
		if !isValidTierCombination(recipe, targetElement) {
			continue
		}
		
		// Format recipe string
		recipeStr := recipe[0]
		if len(recipe) > 1 {
			recipeStr += " + " + recipe[1]
		}
		
		// Skip jika recipe ini sudah pernah dilihat
		if seen[recipeStr] {
			continue
		}
		seen[recipeStr] = true
		
		// Buat node untuk recipe ini
		recipeNode := RecipeNode{
			Result: recipeStr,
		}
		
		wg.Add(1)
		go func(recipe []string, recipeNode RecipeNode) {
			defer wg.Done()
			
			// Proses setiap ingredient secara DFS
			allValid := true
			var childNodes []RecipeNode
			
			for _, ingredient := range recipe {
				select {
				case <-done:
					return // Berhenti jika sudah cukup recipe
				default:
					// Bangun tree untuk ingredient ini
					childNode := buildRecipeTree(ingredient, make(map[string]bool))
					
					// Jika childNode tidak valid, tandai recipe sebagai tidak valid
					if len(childNode.Children) == 0 && !isBaseElement(childNode.Result) {
						allValid = false
						break
					}
					
					childNodes = append(childNodes, childNode)
				}
			}
			
			// Jika semua ingredient valid
			if allValid {
				recipeNode.Children = childNodes
				
				// Buat root node untuk target element
				rootNode := RecipeNode{
					Result:   targetElement,
					Children: []RecipeNode{recipeNode},
				}
				
				// Kirim ke channel hasil
				select {
				case <-done:
					return
				case resultChan <- rootNode:
				}
			}
		}(recipe, recipeNode)
	}
	
	// Goroutine untuk mengumpulkan hasil
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Ambil hasil dari channel
	for rootNode := range resultChan {
		resultMutex.Lock()
		results = append(results, rootNode)
		if len(results) >= maxRecipes {
			close(done) // Signal semua goroutine untuk berhenti
			resultMutex.Unlock()
			break
		}
		resultMutex.Unlock()
	}
	return results
}

func LoadRecipesFromFile(filePath string) error {

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	err = json.Unmarshal(data, &recipesDB)
	if err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}
	
	return nil
}