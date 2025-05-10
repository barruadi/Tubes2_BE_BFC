package main

import (
    "encoding/json"
    "net/http"
    "src/cmd"
    "os"
    "fmt"
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

    var recipes cmd.RecipeMap
	if err := json.Unmarshal(scrapData, &recipes); err != nil {
		fmt.Println("Invalid JSON:", err)
		return
	}

    if r.Method == http.MethodPost {
        var data RequestData
        err := json.NewDecoder(r.Body).Decode(&data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // var response ResponseData
        var results cmd.BfsResult
        if data.AlgorithmType == "bfs" {
            // BFS
            
        } else if data.AlgorithmType == "dfs" {
            // DFS
            results = cmd.FindNPathsBFS(recipes, data.ElementTarget, data.MaxRecipe)
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
