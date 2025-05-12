// // package main

// // import (
// //     "encoding/json"
// //     "net/http"
// //     "src/cmd"
// //     "os"
// // )

// // type RequestData struct {
// //     ElementTarget string `json:"ElementTarget"`
// //     AlgorithmType string `json:"AlgorithmType"`
// //     Multiple      bool   `json:"Multiple"`
// //     MaxRecipe     int    `json:"MaxRecipe"`
// // }

// // func handleData(w http.ResponseWriter, r *http.Request) {
// //     w.Header().Set("Content-Type", "application/json")
// //     w.Header().Set("Access-Control-Allow-Origin", "*")
// //     w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
// //     w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// //     if r.Method == http.MethodOptions {
// //         return
// //     }

// //     if r.Method != http.MethodPost {
// //         http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
// //         return
// //     }

// //     // SCRAP DATA ----- ganti jadi utils.scrapper --------
// //     scrapData, err := os.ReadFile("./data/alchemy_recipes.json")
// //     if err != nil {
// // 		return
// // 	}

// //     var raw map[string]struct {
// // 		Tier    int          `json:"tier"`
// // 		Recipes [][]string   `json:"recipes"`
// // 	}
// // 	if err := json.Unmarshal(scrapData, &raw); err != nil {
// // 		return 
// // 	}

// //     recipes := make(cmd.RecipeMap)
// // 	tiers := make(cmd.TierMap)

// //     for key, val := range raw {
// // 		recipes[key] = val.Recipes
// // 		tiers[key] = val.Tier
// // 	}

// //     if r.Method == http.MethodPost {
// //         var data RequestData
// //         err := json.NewDecoder(r.Body).Decode(&data)
// //         if err != nil {
// //             http.Error(w, err.Error(), http.StatusBadRequest)
// //             return
// //         }

// //         // var response ResponseData
// //         var results cmd.BfsResult = cmd.MainBfs(recipes, tiers, data.ElementTarget, data.MaxRecipe)

// //         if data.AlgorithmType == "bfs" {
// //             // BFS
            
// //         } else if data.AlgorithmType == "dfs" {
// //             // DFS
// //         } else if data.AlgorithmType == "bidirectional" {
// //             // BIDIRECTIONAL

// //         }

// //         w.WriteHeader(http.StatusOK)
// //         json.NewEncoder(w).Encode(results)

// //     } else {
// //         http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
// //     }
// // }

// // func main() {
// //     http.HandleFunc("/api/data", handleData)
// //     http.ListenAndServe(":8080", nil)
// // }


package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	// "os"
	// "time"

	"src/cmd" 
)

func main() {
	
	recipesPath := "data/alchemy_recipes.json" 
	err := cmd.LoadRecipesFromFile(recipesPath)
	if err != nil {
		log.Fatalf("Error loading recipes: %v", err)
	}

	targetElement := "picnic" 
	maxRecipes :=  1



	result := cmd.DFSMultiple(targetElement, maxRecipes)
	
	
	// result := cmd.DFSSingle(targetElement)

	outputFileName := fmt.Sprintf("%s_recipes_%d.json", targetElement, maxRecipes)
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling result to JSON: %v", err)
	}

	err = ioutil.WriteFile(outputFileName, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file: %v", err)
	}
	fmt.Printf("Results saved to %s\n", outputFileName)

	
	// ===== TEST BIDIRECTIONAL SINGLE =====

	

	singleResult := cmd.BidirectionalSingle(targetElement)

	
	// Save single result to JSON
	singleOutputFile := targetElement + "_pure_bidirectional_single.json"
	singleJsonData, err := json.MarshalIndent(singleResult, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}
	
	err = ioutil.WriteFile(singleOutputFile, singleJsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file: %v", err)
	}


	// ===== TEST BIDIRECTIONAL MULTIPLE =====
	

	multipleResult := cmd.BidirectionalSearch(targetElement, maxRecipes)

	multipleOutputFile := fmt.Sprintf("%s_pure_bidirectional_multiple_%d.json", targetElement, maxRecipes)
	multipleJsonData, err := json.MarshalIndent(multipleResult, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}
	
	err = ioutil.WriteFile(multipleOutputFile, multipleJsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file: %v", err)
	}
	
}
