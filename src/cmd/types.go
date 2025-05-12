package cmd


import (
	"sync"
)
type RecipeMap  map[string][][]string
type TierMap 	map[string]int

type BfsResult struct {
	TargetElement string
	RecipeTree    []ElementNode
	VisitedNodes  int
	SearchTime    float64
}

// type BidirectionalResult struct {
// 	TargetElement string        `json:"targetElement"`
// 	RecipeTree    []ElementNode `json:"tree"`
// 	VisitedNodes  int           `json:"nodes"`
// 	SearchTime    float64       `json:"time"`
// }
type ElementNode struct {
	Result   string         `json:"name"`     
	Sources  []string       `json:"sources"`  
	Children []*ElementNode `json:"children"` 
}

type RecipeContextDFS struct {
	RecipeCount   *int32             // Counter recipe yang ditemukan
	MaxCount      int32              // Jumlah maksimum recipe yang dicari
	ResultChan    chan<- ElementNode // Channel untuk mengirim hasil
	Done          <-chan struct{}    // Channel untuk signal berhenti
	TargetElement string             // Elemen target yang dicari
	SeenMutex     *sync.Mutex        // Mutex untuk map seen
	MaxWorkers    int                // Jumlah maksimum worker goroutine
}
type DfsResult struct {
	TargetElement string        `json:"targetElement"`
	RecipeTree    []ElementNode `json:"tree"`
	VisitedNodes  int           `json:"nodes"`
	SearchTime    float64       `json:"time"` 
	CacheStats    interface{}   `json:"cacheStats,omitempty"` 
}


type ElementInfo struct {
	Tier    int        `json:"tier"`
	Recipes [][]string `json:"recipes"`
}


type AlchemyRecipes map[string]ElementInfo

