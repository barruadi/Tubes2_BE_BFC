package cmd

type RecipeMap map[string][][]string
type TierMap map[string]int

type Result struct {
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