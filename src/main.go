// package main
// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"github.com/barruadi/tubes2_be_bfc/src/cmd"
// 	// "time" // Uncomment kalau kamu mau pakai time measurement
// )


// func main() {
// 	cmd.NodesVisited = 0
// 	filePath := "src/data/alchemy_recipes.json"
	
// 	err := cmd.LoadRecipesFromFile(filePath)
// 	if err != nil {
// 		fmt.Printf("Error loading recipes: %v\n", err)
// 		os.Exit(1)
// 	}
// 	element := "picnic"


// 	mode := "multiple"  

// 	maxRecipes := 100
	
// 	outputFile := "outputpicnicold.json"

// 	// startTime := time.Now()
	

// 	var result interface{}
	
// 	if mode == "single" {
// 		result = cmd.DFSSingle(element)
// 	} else {
// 		result = cmd.DFSMultiple(element, maxRecipes)
// 	}

// 	// timeTaken := time.Since(startTime).String()

// 	type OutputResult struct {
// 		RecipeResult interface{} `json:"recipeResult"`
// 		TimeTaken    string      `json:"timeTaken"`
// 		NodesVisited int         `json:"nodesVisited"`
// 		Mode         string      `json:"mode"`
// 	}
	
// 	// output := OutputResult{
// 	// 	RecipeResult: result,
// 	// 	TimeTaken:    timeTaken,
// 	// 	NodesVisited: nodesVisited,
// 	// 	Mode:         mode,
// 	// }
	
// 	jsonResult, err := json.MarshalIndent(result, "", "  ")
// 	if err != nil {
// 		fmt.Println("Error marshaling JSON:", err)
// 		return
// 	}

// 	fmt.Println(string(jsonResult))

// 	if outputFile != "" {
// 		err = ioutil.WriteFile(outputFile, jsonResult, 0644)
// 		if err != nil {
// 			fmt.Printf("Error writing to output file: %v\n", err)
// 		} else {
// 			fmt.Printf("Results written to %s\n", outputFile)
// 		}
// 	}
// }
package main

import (
    "encoding/json"
    "net/http"
    "src/cmd"
    "os"
)

type RequestData struct {
    ElementTarget string `json:"ElementTarget"`
    AlgorithmType string `json:"AlgorithmType"`
    Multiple      bool   `json:"Multiple"`
    MaxRecipe     int    `json:"MaxRecipe"`
}

func handleData(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == http.MethodOptions {
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
        return
    }

    // SCRAP DATA ----- ganti jadi utils.scrapper --------
    scrapData, err := os.ReadFile("./data/alchemy_recipes.json")
    if err != nil {
		return
	}

    var raw map[string]struct {
		Tier    int          `json:"tier"`
		Recipes [][]string   `json:"recipes"`
	}
	if err := json.Unmarshal(scrapData, &raw); err != nil {
		return 
	}

    recipes := make(cmd.RecipeMap)
	tiers := make(cmd.TierMap)

    for key, val := range raw {
		recipes[key] = val.Recipes
		tiers[key] = val.Tier
	}

    if r.Method == http.MethodPost {
        var data RequestData
        err := json.NewDecoder(r.Body).Decode(&data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // var response ResponseData
        var results cmd.BfsResult = cmd.MainBfs(recipes, tiers, data.ElementTarget, data.MaxRecipe)

        if data.AlgorithmType == "bfs" {
            // BFS
            
        } else if data.AlgorithmType == "dfs" {
            // DFS
        } else if data.AlgorithmType == "bidirectional" {
            // BIDIRECTIONAL

        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(results)

    } else {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}

func main() {
    http.HandleFunc("/api/data", handleData)
    http.ListenAndServe(":8080", nil)
}
