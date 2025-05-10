// dfs.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
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
	nodesVisited int
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
	nodesVisited++
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
			childNode := buildRecipeTree(ingredient, newVisited)
			
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
	
	// Bangun recipe tree dengan DFS
	return buildRecipeTree(targetElement, make(map[string]bool))
}

// DFSMultiple mencari multiple recipes dengan algoritma DFS
func DFSMultiple(targetElement string, maxRecipes int) []RecipeNode {
	// Jika target adalah elemen dasar, kembalikan langsung
	if isBaseElement(targetElement) {
		return []RecipeNode{{Result: targetElement}}
	}
	
	// Dapatkan semua recipe untuk target element
	directRecipes := getDirectRecipes(targetElement)
	
	// Jika tidak ada recipe, kembalikan node tanpa children
	if len(directRecipes) == 0 {
		return []RecipeNode{{Result: targetElement}}
	}
	
	// Variabel untuk menyimpan hasil
	var results []RecipeNode
	seen := make(map[string]bool)
	
	// Coba setiap direct recipe
	for _, recipe := range directRecipes {
		// Skip jika sudah mencapai maxRecipes
		if len(results) >= maxRecipes {
			break
		}
		
		// Validasi tier
		if !isValidTierCombination(recipe, targetElement) {
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
		
		// Proses setiap ingredient secara rekursif
		allValid := true
		var childNodes []RecipeNode
		
		for _, ingredient := range recipe {
			// Build tree untuk ingredient ini
			childNode := buildRecipeTree(ingredient, make(map[string]bool))
			
			// Jika childNode tidak valid, tandai recipe sebagai tidak valid
			if len(childNode.Children) == 0 && !isBaseElement(childNode.Result) {
				allValid = false
				break
			}
			
			childNodes = append(childNodes, childNode)
		}
		
		// Jika semua ingredient valid
		if allValid {
			recipeNode.Children = childNodes
			
			// Buat root node untuk target element
			rootNode := RecipeNode{
				Result:   targetElement,
				Children: []RecipeNode{recipeNode},
			}
			
			// Generate key untuk mencegah duplikat
			key := recipeStr  // Use recipe string as key, e.g. "mud + fire"
			
			// Jika belum pernah dilihat, tambahkan ke hasil
			if !seen[key] {
				seen[key] = true
				results = append(results, rootNode)
			}
		}
	}
	
	return results
}

// loadRecipesFromFile memuat data recipe dari file JSON
func loadRecipesFromFile(filePath string) error {

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

func main() {
	nodesVisited = 0
	filePath := "src/data/alchemy_recipes.json"
	
	err := loadRecipesFromFile(filePath)
	if err != nil {
		fmt.Printf("Error loading recipes: %v\n", err)
		os.Exit(1)
	}
	element := "planet"


	mode := "single"  

	maxRecipes := 5

	
	outputFile := "output.json"

	startTime := time.Now()
	

	var result interface{}
	
	if mode == "single" {
		result = DFSSingle(element)
	} else {
		result = DFSMultiple(element, maxRecipes)
	}

	timeTaken := time.Since(startTime).String()

	type OutputResult struct {
		RecipeResult interface{} `json:"recipeResult"`
		TimeTaken    string      `json:"timeTaken"`
		NodesVisited int         `json:"nodesVisited"`
		Mode         string      `json:"mode"`
	}
	
	output := OutputResult{
		RecipeResult: result,
		TimeTaken:    timeTaken,
		NodesVisited: nodesVisited,
		Mode:         mode,
	}
	
	jsonResult, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	fmt.Println(string(jsonResult))

	if outputFile != "" {
		err = ioutil.WriteFile(outputFile, jsonResult, 0644)
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
		} else {
			fmt.Printf("Results written to %s\n", outputFile)
		}
	}
}