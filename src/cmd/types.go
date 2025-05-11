package cmd

type ElementNode struct {
	Result   string         `json:"name"`     // The resulting element
	Sources  []string       `json:"sources"`  // Ingredients used to make this element
	Children []*ElementNode `json:"children"` // Trees of the ingredients
}

type RecipeMap map[string][][]string
type TierMap map[string]int

type BfsResult struct {
	TargetElement string        `json:"targetElement"`
	RecipeTree    []ElementNode `json:"tree"`
	VisitedNodes  int           `json:"nodes"`
	SearchTime    float64       `json:"time"`
}
