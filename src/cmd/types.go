package cmd

type RecipeMap  map[string][][]string
type TierMap 	map[string]int

type BfsResult struct {
	TargetElement string
	RecipeTree    []ElementNode
	VisitedNodes  int
	SearchTime    float64
}

type BidirectionalResult struct {
	TargetElement string        `json:"targetElement"`
	RecipeTree    []ElementNode `json:"tree"`
	VisitedNodes  int           `json:"nodes"`
	SearchTime    float64       `json:"time"`
}
type ElementNode struct {
	Result   string         `json:"name"`     
	Sources  []string       `json:"sources"`  
	Children []*ElementNode `json:"children"` 
}
// RecipeContextDFS menyimpan konteks pencarian DFS
// type RecipeContextDFS struct {
// 	RecipeCount   *int32
// 	MaxCount      int32
// 	ResultChan    chan ElementNode
// 	Done          chan struct{}
// 	TargetElement string
// 	SeenMutex     *sync.Mutex
// 	MaxWorkers    int
// }

type DfsResult struct {
	TargetElement string        `json:"targetElement"`
	RecipeTree    []ElementNode `json:"tree"`
	VisitedNodes  int           `json:"nodes"`
	SearchTime    float64       `json:"time"` 
}


type ElementInfo struct {
	Tier    int        `json:"tier"`
	Recipes [][]string `json:"recipes"`
}


type AlchemyRecipes map[string]ElementInfo


