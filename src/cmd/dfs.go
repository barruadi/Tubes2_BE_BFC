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

// DFSMultiple mencari multiple recipes dengan algoritma DFS
func DFSMultiple(targetElement string, maxRecipes int) []RecipeNode {
    if isBaseElement(targetElement) {
        return []RecipeNode{{Result: targetElement}}
    }
    
    type StackItem struct {
        Element   string
        Path      []string
        RecipeMap map[string][]string  // Menyimpan recipe yang digunakan untuk setiap elemen
    }
    
    // Stack untuk DFS
    stack := []StackItem{
        {
            Element:   targetElement,
            Path:      []string{},
            RecipeMap: make(map[string][]string),
        },
    }
    
    // Set untuk mencegah duplikasi recipe
    seen := make(map[string]bool)
    
    // Hasil recipe trees
    var results []RecipeNode
    
    // Visited set untuk menghindari cycle
    visited := make(map[string]bool)
    
    // Lakukan DFS sampai stack kosong atau maxRecipes terpenuhi
    for len(stack) > 0 && len(results) < maxRecipes {
        // Pop dari stack
        current := stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        incrementNodeCount()
        
        // Jika elemen sudah dikunjungi dalam path ini, lewati (hindari cycle)
        currentPathKey := current.Element
        for _, elem := range current.Path {
            currentPathKey += ":" + elem
        }
        
        if visited[currentPathKey] {
            continue
        }
        
        // Tandai sebagai dikunjungi
        visited[currentPathKey] = true
        
        // Jika elemen dasar, lanjutkan
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
            
            // Buat copy dari RecipeMap
            newRecipeMap := make(map[string][]string)
            for k, v := range current.RecipeMap {
                newRecipeMap[k] = v
            }
            
            // Simpan recipe untuk elemen ini
            newRecipeMap[current.Element] = recipe
            
            // Cek apakah semua ingredients adalah elemen dasar atau sudah memiliki recipe valid
            allIngredientsValid := true
            var nonBaseIngredients []string
            
            for _, ingredient := range recipe {
                if !isBaseElement(ingredient) && newRecipeMap[ingredient] == nil {
                    allIngredientsValid = false
                    nonBaseIngredients = append(nonBaseIngredients, ingredient)
                }
            }
            
            // Jika ada ingredient yang bukan elemen dasar dan belum punya recipe
            if !allIngredientsValid {
                // Tambahkan semua non-base ingredients ke stack
                for _, ingredient := range nonBaseIngredients {
                    newPath := make([]string, len(current.Path))
                    copy(newPath, current.Path)
                    newPath = append(newPath, current.Element)
                    
                    stack = append(stack, StackItem{
                        Element:   ingredient,
                        Path:      newPath,
                        RecipeMap: newRecipeMap,
                    })
                }
                continue
            }
            
            // Jika semua ingredients valid, kita punya recipe lengkap
            // Bangun recipe tree dari RecipeMap
            recipeTree := buildRecipeTree(targetElement, newRecipeMap)
            
            // Buatkan key unik untuk tree ini
            key := generateTreeKey(recipeTree)
            
            // Jika belum pernah dilihat, tambahkan ke hasil
            if !seen[key] {
                seen[key] = true
                results = append(results, recipeTree)
                
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

	element := "picnic"
	mode := "multiple"  
	maxRecipes := 999999999
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