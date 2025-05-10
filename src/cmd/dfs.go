package main
// package dfs
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

// DFSSingle mencari satu recipe dengan algoritma DFS
func DFSSingle(targetElement string) RecipeNode {
	// Jika target adalah elemen dasar, kembalikan langsung
	if isBaseElement(targetElement) {
		incrementNodeCount()
		return RecipeNode{Result: targetElement}
	}
	
	// Stack untuk DFS
	type StackItem struct {
		Element  string
		Path     []string
		Visited  map[string]bool
	}

	// Mulai dari target element
	stack := []StackItem{
		{
			Element:  targetElement,
			Path:     []string{},
			Visited:  make(map[string]bool),
		},
	}
	
	// Map untuk menyimpan recipe yang digunakan
	recipeMap := make(map[string][]string)
	
	// Lakukan DFS
	for len(stack) > 0 {
		// Pop dari stack
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		incrementNodeCount()
		
		// Jika sudah pernah dikunjungi, lewati
		if current.Visited[current.Element] {
			continue
		}
		
		// Tandai sebagai dikunjungi
		current.Visited[current.Element] = true
		
		// Jika elemen dasar, lanjutkan ke elemen berikutnya
		if isBaseElement(current.Element) {
			continue
		}
		
		// Dapatkan semua recipe untuk elemen ini
		recipes := getDirectRecipes(current.Element)
		
		// Coba setiap recipe
		for _, recipe := range recipes {
			// Validasi tier
			if !isValidTierCombination(recipe, current.Element) {
				continue
			}
			
			// Simpan recipe untuk elemen ini
			recipeMap[current.Element] = recipe
			
			// Tambahkan setiap ingredient ke stack
			for _, ingredient := range recipe {
				newPath := make([]string, len(current.Path))
				copy(newPath, current.Path)
				newPath = append(newPath, current.Element)
				
				newVisited := make(map[string]bool)
				for k, v := range current.Visited {
					newVisited[k] = v
				}
				
				stack = append(stack, StackItem{
					Element: ingredient,
					Path:    newPath,
					Visited: newVisited,
				})
			}
		}
	}
	
	// Bangun tree dari recipeMap untuk fe
	return buildRecipeTree(targetElement, recipeMap)
}

// buildRecipeTree membangun tree recipe dari recipeMap // untuk tree front end
func buildRecipeTree(element string, recipeMap map[string][]string) RecipeNode {
	// Jika elemen dasar, kembalikan langsung
	if isBaseElement(element) {
		return RecipeNode{Result: element}
	}
	
	// Dapatkan recipe untuk elemen ini
	recipe, found := recipeMap[element]
	
	// Jika tidak ada recipe, kembalikan node tanpa children
	if !found || len(recipe) == 0 {
		return RecipeNode{Result: element}
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
	for _, ingredient := range recipe {
		childNode := buildRecipeTree(ingredient, recipeMap)
		recipeNode.Children = append(recipeNode.Children, childNode)
	}
	
	// Buat root node untuk elemen ini
	return RecipeNode{
		Result:   element,
		Children: []RecipeNode{recipeNode},
	}
}

// DFSMultiple mencari multiple recipes dengan algoritma DFS
func DFSMultiple(targetElement string, maxRecipes int) []RecipeNode {
	// Variabel untuk menyimpan hasil
	var results []RecipeNode
	
	// Jika target adalah elemen dasar, kembalikan langsung
	if isBaseElement(targetElement) {
		return []RecipeNode{{Result: targetElement}} 
	}
	
	// Dapatkan semua recipe untuk target element
	allRecipes := getDirectRecipes(targetElement)
	
	// Jika tidak ada recipe, kembalikan node tanpa children
	if len(allRecipes) == 0 {
		return []RecipeNode{{Result: targetElement}}
	}
	
	// Map untuk mencegah duplikat
	seen := make(map[string]bool)
	
	// Coba setiap recipe
	for _, recipe := range allRecipes {
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
		
		// Flag untuk memeriksa validitas recipe
		allIngredientsValid := true
		
		// Children nodes untuk recipe ini
		var children []RecipeNode
		
		// Proses setiap ingredient
		for _, ingredient := range recipe {
			// Gunakan DFS untuk ingredient ini
			childNode := DFSSingle(ingredient)
			
			// Jika childNode tidak valid, tandai recipe sebagai tidak valid
			if !isValidTree(childNode) {
				allIngredientsValid = false
				break
			}
			
			children = append(children, childNode)
		}
		
		// Jika semua ingredient valid
		if allIngredientsValid {
			recipeNode.Children = children
			
			// Buat root node untuk target element
			rootNode := RecipeNode{
				Result:   targetElement,
				Children: []RecipeNode{recipeNode},
			}
			
			// Buatkan key unik untuk tree ini
			key := generateTreeKey(rootNode)
			
			// Jika belum pernah dilihat, tambahkan ke hasil
			if !seen[key] {
				seen[key] = true
				results = append(results, rootNode)
				
				// Jika sudah mencapai maxRecipes, keluar dari loop
				if len(results) >= maxRecipes {
					break
				}
			}
		}
	}
	
	return results
}


// isValidTree memeriksa apakah semua leaf dalam tree adalah elemen dasar
func isValidTree(node RecipeNode) bool {
	// Jika node tidak memiliki children (leaf)
	if len(node.Children) == 0 {
		// Jika leaf, harus elemen dasar
		return isBaseElement(node.Result)
	}
	
	// Jika node memiliki children, periksa semua children
	for _, child := range node.Children {
		if !isValidTree(child) {
			return false
		}
	}
	
	return true
}


// Key ini hanya mencegah recipe yang benar-benar identik (sama persis)
func generateTreeKey(node RecipeNode) string {
    if len(node.Children) == 0 {
        return node.Result
    }
    
    key := node.Result + "("
    
    // Untuk node dengan children, gunakan Result dari child pertama sebagai bagian dari key
    // Ini memastikan bahwa recipe yang berbeda (mud+fire vs clay+fire) dianggap berbeda
    if len(node.Children) > 0 {
        key += node.Children[0].Result
    }
    
    key += ")"
    
    return key
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}




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


// UJI COBA 
func main() {
	nodesVisited = 0
	filePath := "src/data/alchemy_recipes.json"
	err := loadRecipesFromFile(filePath)
	if err != nil {
		fmt.Printf("Error loading recipes: %v\n", err)
		os.Exit(1)
	}

	element := "sun"
	mode := "multiple"  
	maxRecipes := 20
	outputFile := "output"
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
	

	//output json
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