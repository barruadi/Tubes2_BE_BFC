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
