package main

import (
    "encoding/json"
    "net/http"
    "src/cmd"
    "os"
)

type RequestData struct {
    ElementTarget   string
    AlgorithmType   string
    Multiple        bool
    MaxRecipe       int
}

type Node struct {
    Name            string
    Children        []*Node
}

type ResponseData struct {
    TargetElement   string
    RecipeTree      *Node
    VisitedNodes    int
	SearchTime      float64
	Found           bool
	Steps           int 
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

    // SCRAP DATA
    scrapData, err := os.ReadFile("../data/alchemy_recipes.json")
    if err != nil {
		return
	}

    if r.Method == http.MethodPost {
        var data RequestData
        err := json.NewDecoder(r.Body).Decode(&data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        var response ResponseData

        if data.AlgorithmType == "bfs" {
            // BFS
            
        } else if data.AlgorithmType == "dfs" {
            // DFS
            // cmd.DFS()
        } else if data.AlgorithmType == "bidirectional" {
            // BIDIRECTIONAL

        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)

    } else {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}

func main() {
    http.HandleFunc("/api/data", handleData)
    http.ListenAndServe(":8080", nil)
}
