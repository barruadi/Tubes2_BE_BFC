package cmd

type ElementNode struct {
	Result    string   			`json:"name"`
	Children []ElementNode  	`json:"children"`
}

type RecipeMap  map[string][][]string
type TierMap 	map[string]int

type BfsResult struct {
	TargetElement string			`json:"targetElement"`
	RecipeTree    []ElementNode		`json:"tree"`
	VisitedNodes  int				`json:"nodes"`
	SearchTime    float64			`json:"time"`
}