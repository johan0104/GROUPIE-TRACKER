package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

var templates = template.Must(template.ParseFiles("templates/accueil.html", "templates/collection.html", "templates/categorie1.html", "templates/categorie2.html", "templates/categorie3.html", "templates/ressource.html", "templates/favoris.html", "templates/recherche.html", "templates/apropos.html", "templates/erreur404.html"))

const favoritesFilePath = "favorites.json"

type Favorites struct {
    Articles []string `json:"articles"` // IDs des articles favoris
}

// Struct pour les données que vous souhaitez passer à vos templates
type PageData struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	ImageURL    string `json:"imageUrl"`
	NewsSite    string `json:"newsSite"`
	Summary     string `json:"summary"`
	PublishedAt string `json:"publishedAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func main() {
	// Gère la route "/static/" pour servir des fichiers statiques depuis le dossier "static"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/accueil", accueilHandler)
	http.HandleFunc("/collection", collectionHandler)
	http.HandleFunc("/categorie1", categorieHandler("https://api.spaceflightnewsapi.net/v3/articles?_limit=50"))
	http.HandleFunc("/categorie2", categorieHandler("https://api.spaceflightnewsapi.net/v3/reports?_limit=50"))
	http.HandleFunc("/categorie3", categorieHandler("https://api.spaceflightnewsapi.net/v3/blogs?_limit=50"))
	http.HandleFunc("/ressource", ressourceHandler)
	http.HandleFunc("/add-to-favorites", addToFavoritesHandler)
	http.HandleFunc("/remove-from-favorites", removeFromFavoritesHandler)
	http.HandleFunc("/favoris", favorisHandler)
	http.HandleFunc("/recherche", rechercheHandler)
	http.HandleFunc("/apropos", aproposHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        erreur404Handler(w, r)
    })

	// Démarrez le serveur web
	fmt.Println("Server started at http://localhost:7070")
	http.ListenAndServe(":7070", nil)
}

func categorieHandler(url string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpClient := &http.Client{Timeout: time.Second * 10}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			http.Error(w, "Unable to create request", http.StatusInternalServerError)
			return
		}

		res, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		var data []PageData
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			http.Error(w, "Failed to decode data", http.StatusInternalServerError)
			return
		}

		templateName := ""
		switch url {
		case "https://api.spaceflightnewsapi.net/v3/articles?_limit=50":
			templateName = "categorie1.html"
		case "https://api.spaceflightnewsapi.net/v3/reports?_limit=50":
			templateName = "categorie2.html"
		case "https://api.spaceflightnewsapi.net/v3/blogs?_limit=50":
			templateName = "categorie3.html"
		}

		templates.ExecuteTemplate(w, templateName, data)
	}
}

func collectionHandler(w http.ResponseWriter, r *http.Request) {
	// Définissez les URLs des endpoints
	urls := []string{
		"https://api.spaceflightnewsapi.net/v3/articles?_limit=50",
		"https://api.spaceflightnewsapi.net/v3/reports?_limit=50",
		"https://api.spaceflightnewsapi.net/v3/blogs?_limit=50",
	}

	httpClient := &http.Client{Timeout: time.Second * 10}

	// Créez un conteneur pour les données de chaque catégorie
	var data struct {
		Articles []PageData
		Reports  []PageData
		Blogs    []PageData
	}

	// Parcourez chaque URL et récupérez les données
	for _, url := range urls {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			http.Error(w, "Unable to create request", http.StatusInternalServerError)
			return
		}

		res, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		// Utilisez le bon conteneur en fonction de l'URL
		var tmpData []PageData
		if err := json.NewDecoder(res.Body).Decode(&tmpData); err != nil {
			http.Error(w, "Failed to decode data", http.StatusInternalServerError)
			return
		}

		switch url {
		case "https://api.spaceflightnewsapi.net/v3/articles?_limit=50":
			data.Articles = tmpData
		case "https://api.spaceflightnewsapi.net/v3/reports?_limit=50":
			data.Reports = tmpData
		case "https://api.spaceflightnewsapi.net/v3/blogs?_limit=50":
			data.Blogs = tmpData
		}
	}

	// Passez les données collectées au template
	templates.ExecuteTemplate(w, "collection.html", data)
}

func ressourceHandler(w http.ResponseWriter, r *http.Request) {
	// Extraire le type et l'ID de la requête
	query := r.URL.Query()
	typeRes := query.Get("type")
	id := query.Get("id")

	fmt.Println(typeRes)

	if typeRes == "" || id == "" {
		http.Error(w, "Type ou ID de la ressource manquant", http.StatusBadRequest)
		return
	}
	ressourceDetails, err := fetchRessourceDetails(typeRes, id)
	if err != nil {
		http.Error(w, "Échec de la récupération des détails de la ressource", http.StatusInternalServerError)
		return
	}

	// Passez les détails récupérés au template de la page de détail
	templates.ExecuteTemplate(w, "ressource.html", ressourceDetails)
}

func fetchRessourceDetails(typeRes, id string) (*PageData, error) {
	// Liste des endpoints à essayer. Vous devez les remplacer par les vrais endpoints.
	endpoints := fmt.Sprintf("https://api.spaceflightnewsapi.net/v3/%s/%s", typeRes, id)

	httpClient := &http.Client{Timeout: time.Second * 10}
	var data PageData

	// Essayez chaque endpoint jusqu'à ce que vous trouviez la ressource ou que vous épuisiez les options.
	fmt.Println("Tentative avec URL:", endpoints) // Ajout de log pour débogage
	req, err := http.NewRequest(http.MethodGet, endpoints, nil)
	if err != nil {
		fmt.Println("Erreur lors de la création de la requête :", err)
	}

	res, err := httpClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		fmt.Println("Erreur ou code d'état HTTP non-OK :", res.StatusCode)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&data); err == nil {
		return &data, nil
	}
	// Ajoutez un log ici si le décodage échoue, pour voir que la requête a réussi mais le décodage non.

	return nil, fmt.Errorf("la ressource avec l'ID %s n'a pas été trouvée", id)
}

func rechercheHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Requête de recherche vide", http.StatusBadRequest)
		return
	}

	// Ici, implémentez la logique pour récupérer toutes vos ressources.
	// Pour cet exemple, nous faisons semblant avec une fonction fetchAllResources()
	allResources, err := fetchAllResources()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des ressources", http.StatusInternalServerError)
		return
	}

	var results []PageData
	for _, resource := range allResources {
		if strings.Contains(strings.ToLower(resource.Title), strings.ToLower(query)) {
			results = append(results, resource)
		}
	}

	// Assurez-vous d'avoir un template `recherche.html` capable d'afficher les résultats
	templates.ExecuteTemplate(w, "recherche.html", results)
}

// Implémentez fetchAllResources selon votre source de données
func fetchAllResources() ([]PageData, error) {
	urls := []string{
		"https://api.spaceflightnewsapi.net/v3/articles?_limit=50",
		"https://api.spaceflightnewsapi.net/v3/reports?_limit=50",
		"https://api.spaceflightnewsapi.net/v3/blogs?_limit=50",
	}

	var allResources []PageData
	httpClient := &http.Client{Timeout: time.Second * 10}

	for _, url := range urls {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request failed: %v", err)
		}

		res, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching data failed: %v", err)
		}

		var resources []PageData
		if err := json.NewDecoder(res.Body).Decode(&resources); err != nil {
			res.Body.Close() // Assurez-vous de fermer le corps de la réponse avant de retourner une erreur
			return nil, fmt.Errorf("decoding data failed: %v", err)
		}
		res.Body.Close() // Fermez le corps de la réponse après l'avoir lu

		allResources = append(allResources, resources...)
	}

	return allResources, nil
}

func addToFavoritesHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }

    resourceID := r.FormValue("resourceID")
    if resourceID == "" {
        http.Error(w, "Resource ID is required", http.StatusBadRequest)
        return
    }

    favorites, err := loadFavorites() // Charger les favoris existants
    if err != nil && !os.IsNotExist(err) {
        http.Error(w, "Failed to load favorites", http.StatusInternalServerError)
        return
    }

    // Ajouter l'ID de ressource à la liste des favoris, en évitant les doublons
    for _, id := range favorites.Articles {
        if id == resourceID {
            http.Redirect(w, r, "/favoris", http.StatusFound) // ID déjà dans les favoris
            return
        }
    }
    favorites.Articles = append(favorites.Articles, resourceID)

    if err := saveFavorites(favorites); err != nil { // Utiliser saveFavorites pour sauvegarder les modifications
        http.Error(w, "Failed to save favorites", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/favoris", http.StatusFound)
}



func favorisHandler(w http.ResponseWriter, r *http.Request) {
    favorites, err := loadFavorites()
    if err != nil {
        http.Error(w, "Failed to load favorites", http.StatusInternalServerError)
        return
    }

    // Supposons que vous ayez une fonction pour récupérer les détails des articles par ID
    var articlesDetails []PageData
    for _, id := range favorites.Articles {
        article, err := fetchArticleDetails(id) // Vous devez implémenter cette fonction
        if err == nil {
            articlesDetails = append(articlesDetails, article)
        }
    }

    // Passe les détails des articles favoris au template
    templates.ExecuteTemplate(w, "favoris.html", articlesDetails)
}



func fetchArticleDetails(articleID string) (PageData, error) {
    var article PageData
    url := fmt.Sprintf("https://api.spaceflightnewsapi.net/v3/articles/%s", articleID)
    
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return article, err
    }

    resp, err := client.Do(req)
    if err != nil {
        return article, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return article, fmt.Errorf("failed to fetch article with ID %s, status code: %d", articleID, resp.StatusCode)
    }

    err = json.NewDecoder(resp.Body).Decode(&article)
    if err != nil {
        return article, err
    }

    return article, nil
}

func removeFromFavoritesHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return
    }

    resourceID := r.FormValue("resourceID")
    if resourceID == "" {
        http.Error(w, "Resource ID is required", http.StatusBadRequest)
        return
    }

    favorites, err := loadFavorites() // Charger les favoris existants
    if err != nil {
        http.Error(w, "Failed to load favorites", http.StatusInternalServerError)
        return
    }

    // Retirer l'ID des favoris
    for i, id := range favorites.Articles {
        if id == resourceID {
            favorites.Articles = append(favorites.Articles[:i], favorites.Articles[i+1:]...)
            break
        }
    }

    if err := saveFavorites(favorites); err != nil { // Utiliser saveFavorites pour sauvegarder les modifications
        http.Error(w, "Failed to save updated favorites", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/favoris", http.StatusFound)
}

func saveFavorites(favorites Favorites) error {
    file, err := json.MarshalIndent(favorites, "", " ")
    if err != nil {
        return err
    }
    return ioutil.WriteFile(favoritesFilePath, file, 0644)
}

func loadFavorites() (Favorites, error) {
    var favorites Favorites
    file, err := ioutil.ReadFile(favoritesFilePath)
    if err != nil {
        return favorites, err
    }
    err = json.Unmarshal(file, &favorites)
    return favorites, err
}

func erreur404Handler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound) // Important : Définir le statut HTTP à 404
    templates.ExecuteTemplate(w, "erreur404.html", nil)
}

func accueilHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "accueil.html", r)
}

func aproposHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "apropos.html", r)
}