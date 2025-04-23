package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)" 

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		panic(err)
	}
	html := buf.String()

	// Inisialisasi recipes sebagai map untuk JSON
	recipes := make(map[string][][]string)

	// Ambil semua tabel
	tableRegex := regexp.MustCompile(`(?is)<table.*?>.*?</table>`)
	tables := tableRegex.FindAllString(html, -1)

	for _, table := range tables {
		// Ambil semua baris <tr>
		rowRegex := regexp.MustCompile(`(?is)<tr.*?>.*?</tr>`)
		rows := rowRegex.FindAllString(table, -1)

		for _, row := range rows {
			// Ambil <td>
			tdRegex := regexp.MustCompile(`(?is)<td.*?>.*?</td>`)
			tds := tdRegex.FindAllString(row, -1)

			if len(tds) < 2 {
				continue
			}

			// --- Ambil Element (kolom pertama) ---
			elementRegex := regexp.MustCompile(`(?is)<a href="/wiki/[^"]*" title="[^"]*">(.*?)</a>`)
			elementMatch := elementRegex.FindStringSubmatch(tds[0])
			element := "UNKNOWN"
			if len(elementMatch) > 1 {
				element = cleanText(elementMatch[1])
				// Konversi ke lowercase untuk konsistensi
				element = strings.ToLower(element)
			}

			if element == "unknown" {
				continue 
			}

			// --- Ambil Komposer1 dan Komposer2 (kolom kedua) ---
			komposers := extractKomposers(tds[1])

			// Tambahkan ke recipes JSON
			if len(komposers) > 0 {
				if _, exists := recipes[element]; !exists {
					recipes[element] = make([][]string, 0)
				}

				for _, komposer := range komposers {
					parts := strings.Split(komposer, " + ")
					if len(parts) == 2 {
						ingredient1 := strings.ToLower(strings.TrimSpace(parts[0]))
						ingredient2 := strings.ToLower(strings.TrimSpace(parts[1]))
						recipes[element] = append(recipes[element], []string{ingredient1, ingredient2})
					}
				}
			}
		}
	}

	// Convert ke JSON dengan indentasi
	jsonData, err := json.MarshalIndent(recipes, "", "  ")
	if err != nil {
		panic(err)
	}
	output := "src/data/alchemy_recipes.json"
	// Tulis ke file JSON
	err = os.WriteFile(output, jsonData, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Berhasil disimpan di alchemy_recipes.json")
}

// Fungsi bantu untuk bersihin spasi/HTML entities kecil
func cleanText(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// Fungsi untuk ekstrak komposer dari <ul><li> dengan mengabaikan <span>
func extractKomposers(td string) []string {
	// Ambil semua <li> dalam <ul>
	liRegex := regexp.MustCompile(`(?is)<li[^>]*>.*?</li>`)
	liMatches := liRegex.FindAllString(td, -1)

	komposers := []string{}

	for _, li := range liMatches {
		// Hapus semua <span> tags dari string
		spanRegex := regexp.MustCompile(`(?is)<span[^>]*>.*?</span>`)
		cleanLi := spanRegex.ReplaceAllString(li, "")

		// Ambil <a> dalam <li>
		aRegex := regexp.MustCompile(`(?is)<a[^>]*>(.*?)</a>`)
		aMatches := aRegex.FindAllStringSubmatch(cleanLi, -1)

		if len(aMatches) == 2 {
			komposer1 := cleanText(aMatches[0][1])
			komposer2 := cleanText(aMatches[1][1])
			komposers = append(komposers, komposer1+" + "+komposer2)
		}
	}

	return komposers
}