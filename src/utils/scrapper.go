package main  
// NTAR GANTI KE PACKAGE SCRAPPER klo mau integrate , func mainny jg apus
// package scrapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ElementInfo adalah struktur untuk menyimpan informasi tentang elemen
type ElementInfo struct {
	Tier    int        `json:"tier"`
	Recipes [][]string `json:"recipes"`
}

// CleanText menghilangkan whitespace berlebih dari string
func CleanText(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

// ExtractTier mengekstrak nomor tier dari id heading
func ExtractTier(id string) (int, error) {
	re := regexp.MustCompile(`Tier_(\d+)`)
	matches := re.FindStringSubmatch(id)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no tier number found in: %s", id)
	}
	return strconv.Atoi(matches[1])
}

// ParseRecipes mengekstrak resep dari elemen <td>
func ParseRecipes(recipeCell *goquery.Selection) [][]string {
	var recipes [][]string
	
	recipeCell.Find("li").Each(func(i int, li *goquery.Selection) {
		var ingredients []string
		
		li.Find("a").Each(func(j int, a *goquery.Selection) {
			ingredient := strings.ToLower(CleanText(a.Text()))
			if ingredient != "" {
				ingredients = append(ingredients, ingredient)
			}
		})
		
		// Resep harus berisi 2 bahan
		if len(ingredients) == 2 {
			recipes = append(recipes, ingredients)
		}
	})
	
	return recipes
}

// ScrapeAlchemyElements melakukan scraping pada wiki Little Alchemy 2
func ScrapeAlchemyElements() (map[string]ElementInfo, error) {
	startTime := time.Now()
	
	// Siapkan HTTP client dengan User-Agent untuk menghindari pemblokiran
	client := &http.Client{}
	req, err := http.NewRequest("GET", 
		"https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("User-Agent", 
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	
	// Kirim request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}
	
	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}
	
	// Inisialisasi map untuk menyimpan data elemen
	elements := make(map[string]ElementInfo)
	
	// Tambahkan elemen dasar secara manual
	baseElements := []string{"air", "earth", "fire", "water"}
	for _, base := range baseElements {
		elements[base] = ElementInfo{
			Tier:    0,
			Recipes: [][]string{},
		}
	}
	
	// Cari semua heading tier
	doc.Find("h2, h3").Each(func(i int, s *goquery.Selection) {
		headlineSpan := s.Find("span.mw-headline")
		if headlineSpan.Length() == 0 {
			return
		}
		
		id, exists := headlineSpan.Attr("id")
		if !exists || !strings.HasPrefix(id, "Tier_") {
			return
		}
		
		// Ekstrak nomor tier
		tier, err := ExtractTier(id)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			return
		}
		
		fmt.Printf("Processing Tier %d elements...\n", tier)
		
		// Cari tabel yang mengikuti heading ini
		var table *goquery.Selection
		
		next := s.NextAll()
		next.EachWithBreak(func(i int, el *goquery.Selection) bool {
			if el.Is("table") {
				table = el
				return false // break loop
			} else if el.Is("h2, h3") {
				return false // break jika bertemu heading lain
			}
			return true // lanjutkan loop
		})
		
		if table == nil {
			fmt.Printf("Warning: No table found for Tier %d\n", tier)
			return
		}
		
		// Proses setiap baris dalam tabel (skip header)
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			if j == 0 {
				return // Skip header row
			}
			
			cells := row.Find("td")
			if cells.Length() < 2 {
				return // Skip jika tidak cukup sel
			}
			
			// Ambil nama elemen dari sel pertama
			elementCell := cells.Eq(0)
			elementLink := elementCell.Find("a")
			elementName := ""
			
			if elementLink.Length() > 0 {
				elementName = strings.ToLower(CleanText(elementLink.Text()))
			} else {
				elementName = strings.ToLower(CleanText(elementCell.Text()))
			}
			
			if elementName == "" {
				return // Skip jika tidak ada nama elemen
			}
			
			// Ambil resep dari sel kedua
			recipeCell := cells.Eq(1)
			recipes := ParseRecipes(recipeCell)
			
			if len(recipes) > 0 {
				elements[elementName] = ElementInfo{
					Tier:    tier,
					Recipes: recipes,
				}
			}
		})
	})
	
	// Tambahkan debug info
	elapsedTime := time.Since(startTime)
	fmt.Printf("Scraping completed in %s\n", elapsedTime)
	fmt.Printf("Found %d elements (including %d base elements)\n", 
		len(elements), len(baseElements))
	
	return elements, nil
}

// SaveElementsToJSON menyimpan data elemen ke file JSON
func SaveElementsToJSON(elements map[string]ElementInfo, filepath string) error {
	// Buat JSON
	jsonData, err := json.MarshalIndent(elements, "", "  ")
	if err != nil {
		return fmt.Errorf("error creating JSON: %w", err)
	}
	
	// Buat direktori jika belum ada
	dir := strings.TrimSuffix(filepath, "/"+filepath)
	if dir != "" && dir != filepath {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}
	
	// Tulis ke file
	err = ioutil.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON file: %w", err)
	}
	
	fmt.Printf("Successfully saved data to %s\n", filepath)
	return nil
}

// LoadElementsFromJSON membaca data elemen dari file JSON
func LoadElementsFromJSON(filepath string) (map[string]ElementInfo, error) {
	// Baca file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON file: %w", err)
	}
	
	// Parse JSON
	var elements map[string]ElementInfo
	err = json.Unmarshal(data, &elements)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	
	return elements, nil
}

// Contoh Fungsi main untuk testing
func main() {
	// Test scraping
	fmt.Println("Starting Little Alchemy 2 scraper...")
	
	// Scrape elemen dari wiki
	elements, err := ScrapeAlchemyElements()
	if err != nil {
		fmt.Printf("Error scraping elements: %v\n", err)
		os.Exit(1)
	}
	
	// Buat direktori untuk output jika belum ada
	err = os.MkdirAll("src/data", 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}
	
	// Simpan ke file JSON
	outputPath := "src/data/alchemy_recipes.json"
	err = SaveElementsToJSON(elements, outputPath)
	if err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
		os.Exit(1)
	}
	
	// Tes membaca dari file JSON
	fmt.Println("\nTesting JSON loading...")
	loadedElements, err := LoadElementsFromJSON(outputPath)
	if err != nil {
		fmt.Printf("Error loading JSON: %v\n", err)
		os.Exit(1)
	}
	
	// Verifikasi data
	fmt.Printf("Successfully loaded %d elements from JSON\n", len(loadedElements))
	
	// Tampilkan beberapa elemen
	fmt.Println("\nSample elements:")
	sampleElements := []string{"brick", "volcano", "human", "time"}
	for _, name := range sampleElements {
		if info, exists := loadedElements[name]; exists {
			fmt.Printf("- %s (Tier %d) has %d recipe(s)\n", 
				name, info.Tier, len(info.Recipes))
			for i, recipe := range info.Recipes {
				if i < 3 { // Tunjukkan max 3 resep
					fmt.Printf("  * %s + %s\n", recipe[0], recipe[1])
				}
			}
			if len(info.Recipes) > 3 {
				fmt.Printf("  * (and %d more recipes)\n", len(info.Recipes)-3)
			}
		} else {
			fmt.Printf("- %s: Not found\n", name)
		}
	}
	
	fmt.Println("\nScraper test completed successfully!")
}