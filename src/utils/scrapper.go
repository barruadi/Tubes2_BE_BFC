// package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"regexp"
// 	"strings"
// )

// func main() {
// 	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)" 

// 	resp, err := http.Get(url)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	buf := new(bytes.Buffer)
// 	_, err = buf.ReadFrom(resp.Body)
// 	if err != nil {
// 		panic(err)
// 	}
// 	html := buf.String()

// 	// Inisialisasi recipes sebagai map untuk JSON
// 	recipes := make(map[string][][]string)

// 	// Ambil semua tabel
// 	tableRegex := regexp.MustCompile(`(?is)<table.*?>.*?</table>`)
// 	tables := tableRegex.FindAllString(html, -1)

// 	for _, table := range tables {
// 		// Ambil semua baris <tr>
// 		rowRegex := regexp.MustCompile(`(?is)<tr.*?>.*?</tr>`)
// 		rows := rowRegex.FindAllString(table, -1)

// 		for _, row := range rows {
// 			// Ambil <td>
// 			tdRegex := regexp.MustCompile(`(?is)<td.*?>.*?</td>`)
// 			tds := tdRegex.FindAllString(row, -1)

// 			if len(tds) < 2 {
// 				continue
// 			}

// 			// --- Ambil Element (kolom pertama) ---
// 			elementRegex := regexp.MustCompile(`(?is)<a href="/wiki/[^"]*" title="[^"]*">(.*?)</a>`)
// 			elementMatch := elementRegex.FindStringSubmatch(tds[0])
// 			element := "UNKNOWN"
// 			if len(elementMatch) > 1 {
// 				element = cleanText(elementMatch[1])
// 				// Konversi ke lowercase untuk konsistensi
// 				element = strings.ToLower(element)
// 			}

// 			if element == "unknown" {
// 				continue 
// 			}

// 			// --- Ambil Komposer1 dan Komposer2 (kolom kedua) ---
// 			komposers := extractKomposers(tds[1])

// 			// Tambahkan ke recipes JSON
// 			if len(komposers) > 0 {
// 				if _, exists := recipes[element]; !exists {
// 					recipes[element] = make([][]string, 0)
// 				}

// 				for _, komposer := range komposers {
// 					parts := strings.Split(komposer, " + ")
// 					if len(parts) == 2 {
// 						ingredient1 := strings.ToLower(strings.TrimSpace(parts[0]))
// 						ingredient2 := strings.ToLower(strings.TrimSpace(parts[1]))
// 						recipes[element] = append(recipes[element], []string{ingredient1, ingredient2})
// 					}
// 				}
// 			}
// 		}
// 	}

// 	// Convert ke JSON dengan indentasi
// 	jsonData, err := json.MarshalIndent(recipes, "", "  ")
// 	if err != nil {
// 		panic(err)
// 	}
// 	output := "src/data/alchemy_recipes2.json"
// 	// Tulis ke file JSON
// 	err = os.WriteFile(output, jsonData, 0644)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("Berhasil disimpan di alchemy_recipes.json")
// }

// // Fungsi bantu untuk bersihin spasi/HTML entities kecil
// func cleanText(s string) string {
// 	re := regexp.MustCompile(`\s+`)
// 	return strings.TrimSpace(re.ReplaceAllString(s, " "))
// }

// // Fungsi untuk ekstrak komposer dari <ul><li> dengan mengabaikan <span>
// func extractKomposers(td string) []string {
// 	// Ambil semua <li> dalam <ul>
// 	liRegex := regexp.MustCompile(`(?is)<li[^>]*>.*?</li>`)
// 	liMatches := liRegex.FindAllString(td, -1)

// 	komposers := []string{}

// 	for _, li := range liMatches {
// 		// Hapus semua <span> tags dari string
// 		spanRegex := regexp.MustCompile(`(?is)<span[^>]*>.*?</span>`)
// 		cleanLi := spanRegex.ReplaceAllString(li, "")

// 		// Ambil <a> dalam <li>
// 		aRegex := regexp.MustCompile(`(?is)<a[^>]*>(.*?)</a>`)
// 		aMatches := aRegex.FindAllStringSubmatch(cleanLi, -1)

// 		if len(aMatches) == 2 {
// 			komposer1 := cleanText(aMatches[0][1])
// 			komposer2 := cleanText(aMatches[1][1])
// 			komposers = append(komposers, komposer1+" + "+komposer2)
// 		}
// 	}

// 	return komposers
// }


package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ElementInfo adalah struktur data untuk menyimpan tier dan resep elemen
type ElementInfo struct {
	Tier    int        `json:"tier"`
	Recipes [][]string `json:"recipes"`
}

// cleanText menghilangkan whitespace berlebih
func cleanText(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

// extractTierNumber mengekstrak nomor tier dari string id
func extractTierNumber(id string) (int, error) {
	re := regexp.MustCompile(`Tier_(\d+)`)
	matches := re.FindStringSubmatch(id)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no tier number found in: %s", id)
	}
	return strconv.Atoi(matches[1])
}

// parseRecipes mengekstrak resep dari sel <td>
func parseRecipes(recipeCell *goquery.Selection) [][]string {
	var recipes [][]string
	
	recipeCell.Find("li").Each(func(i int, li *goquery.Selection) {
		var ingredients []string
		
		li.Find("a").Each(func(j int, a *goquery.Selection) {
			ingredient := strings.ToLower(cleanText(a.Text()))
			if ingredient != "" {
				ingredients = append(ingredients, ingredient)
			}
		})
		
		// Recipe harus berisi 2 ingredients
		if len(ingredients) == 2 {
			recipes = append(recipes, ingredients)
		}
	})
	
	return recipes
}

func main() {
	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
	
	// Buat HTTP request dengan User-Agent yang lebih mirip browser
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		fmt.Printf("Status code error: %d %s\n", resp.StatusCode, resp.Status)
		return
	}
	
	// Parse HTML menggunakan goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}
	
	// Inisialisasi map untuk menyimpan informasi elemen
	elements := make(map[string]ElementInfo)
	
	// Tambahkan elemen dasar secara manual
	baseElements := []string{"air", "earth", "fire", "water"}
	for _, base := range baseElements {
		elements[base] = ElementInfo{
			Tier:    0,
			Recipes: [][]string{},
		}
	}
	
	// Cari semua heading tier (h2 dan h3)
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
		tier, err := extractTierNumber(id)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			return
		}
		
		fmt.Printf("Processing Tier %d elements...\n", tier)
		
		// Cari tabel yang mengikuti heading ini
		var table *goquery.Selection
		
		// Cari tabel setelah heading
		next := s.NextAll()
		next.EachWithBreak(func(i int, el *goquery.Selection) bool {
			if el.Is("table") {
				table = el
				return false // break loop
			} else if el.Is("h2, h3") {
				return false // break if we hit another heading first
			}
			return true // continue loop
		})
		
		if table == nil {
			fmt.Printf("Warning: No table found for Tier %d\n", tier)
			return
		}
		
		// Process each row in the table (skip header row)
		table.Find("tr").Each(func(j int, row *goquery.Selection) {
			if j == 0 {
				return // Skip header row
			}
			
			cells := row.Find("td")
			if cells.Length() < 2 {
				return // Skip if not enough cells
			}
			
			// Get element name from first cell
			elementCell := cells.Eq(0)
			elementLink := elementCell.Find("a")
			elementName := ""
			
			if elementLink.Length() > 0 {
				elementName = strings.ToLower(cleanText(elementLink.Text()))
			} else {
				elementName = strings.ToLower(cleanText(elementCell.Text()))
			}
			
			if elementName == "" {
				return // Skip if no element name found
			}
			
			// Get recipes from second cell
			recipeCell := cells.Eq(1)
			recipes := parseRecipes(recipeCell)
			
			if len(recipes) > 0 {
				elements[elementName] = ElementInfo{
					Tier:    tier,
					Recipes: recipes,
				}
			}
		})
	})
	
	// Convert to JSON
	jsonData, err := json.MarshalIndent(elements, "", "  ")
	if err != nil {
		fmt.Println("Error creating JSON:", err)
		return
	}
	
	// Create output directory if it doesn't exist
	err = os.MkdirAll("src/data", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	
	// Write to file
	err = os.WriteFile("src/data/alchemy_recipes.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON file:", err)
		return
	}
	
	fmt.Printf("Successfully processed and saved %d elements (including %d basic elements) to src/data/alchemy_recipes.json\n", 
		len(elements), len(baseElements))
}