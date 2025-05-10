package cmd

type ElementNode struct {
	Result    string   
	Children []ElementNode  
}

type RecipeMap  map[string][][]string

type BfsResult struct {
	TargetElement string
	RecipeTree    []ElementNode
	VisitedNodes  int
	SearchTime    float64
}