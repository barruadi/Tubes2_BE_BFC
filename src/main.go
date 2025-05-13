package main

import (
    "encoding/json"
    "net/http"
    "tubes2_be_bfc/src/cmd"
	"tubes2_be_bfc/src/utils"
	"log"
    // "os"
)

var (
    Recipes cmd.RecipeMap
    Tiers   cmd.TierMap
)

type RequestData struct {
    ElementTarget string `json:"ElementTarget"`
    AlgorithmType string `json:"AlgorithmType"`
    Multiple      bool   `json:"Multiple"`
    MaxRecipe     int    `json:"MaxRecipe"`
}
func init() {

  
    scrapData, err := utils.ScrapeAlchemyElements()
    if err != nil {
        log.Fatalf("Gagal melakukan scraping: %v", err)
    }
    
    log.Printf("Berhasil scraping %d elemen", len(scrapData))
    
    Recipes = make(cmd.RecipeMap)
    Tiers = make(cmd.TierMap)
 
    for key, val := range scrapData {
        Recipes[key] = val.Recipes
        Tiers[key] = val.Tier
    }
    
    log.Println("Persiapan data selesai, siap menerima permintaan.")
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
    if r.Method == http.MethodPost {
        var data RequestData
        err := json.NewDecoder(r.Body).Decode(&data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // var response ResponseData
        var results cmd.Result

        if data.AlgorithmType == "bfs" {
            // BFS
            results = cmd.MainBfs(Recipes, Tiers, data.ElementTarget, data.MaxRecipe)
        } else if data.AlgorithmType == "dfs" {
            // DFS
            results = cmd.MainDfs(Recipes, Tiers, data.ElementTarget, data.MaxRecipe)
        } else if data.AlgorithmType == "bidirectional" {
            // BIDIRECTIONAL
            // results = cmd.BidirectionalBfs(recipes, tiers, data.ElementTarget, data.MaxRecipe)
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
