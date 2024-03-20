package main

import (
    "fmt"
    "html/template"
    "net/http"
    "encoding/json"
    "time"
)

var templates = template.Must(template.ParseFiles("templates/accueil.html", "templates/connexion.html", "templates/collection.html", "templates/categorie1.html", "templates/categorie2.html", "templates/categorie3.html", "templates/ressource.html", "templates/favoris.html", "templates/recherche.html", "templates/apropos.html", "templates/erreur404.html",))

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
http.HandleFunc("/connexion", connexionHandler)
http.HandleFunc("/collection", collectionHandler)
http.HandleFunc("/categorie1", categorieHandler("https://api.spaceflightnewsapi.net/v3/articles/"))
http.HandleFunc("/categorie2", categorieHandler("https://api.spaceflightnewsapi.net/v3/reports"))
http.HandleFunc("/categorie3", categorieHandler("https://api.spaceflightnewsapi.net/v3/blogs/"))
http.HandleFunc("/ressource", ressourceHandler)
http.HandleFunc("/favoris", favorisHandler)
http.HandleFunc("/recherche", rechercheHandler) 
http.HandleFunc("/apropos", aproposHandler)
http.HandleFunc("/erreur404", erreur404Handler)

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
        case "https://api.spaceflightnewsapi.net/v3/articles/":
            templateName = "categorie1.html"
        case "https://api.spaceflightnewsapi.net/v3/reports":
            templateName = "categorie2.html"
        case "https://api.spaceflightnewsapi.net/v3/blogs/":
            templateName = "categorie3.html"
        }

        templates.ExecuteTemplate(w, templateName, data)
    }
}

func collectionHandler(w http.ResponseWriter, r *http.Request) {
    // Définissez les URLs des endpoints
    urls := []string{
        "https://api.spaceflightnewsapi.net/v3/articles/",
        "https://api.spaceflightnewsapi.net/v3/reports",
        "https://api.spaceflightnewsapi.net/v3/blogs/",
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
        case "https://api.spaceflightnewsapi.net/v3/articles/":
            data.Articles = tmpData
        case "https://api.spaceflightnewsapi.net/v3/reports":
            data.Reports = tmpData
        case "https://api.spaceflightnewsapi.net/v3/blogs/":
            data.Blogs = tmpData
        }
    }

    // Passez les données collectées au template
    templates.ExecuteTemplate(w, "collection.html", data)
}

func ressourceHandler(w http.ResponseWriter, r *http.Request) {
    // Extraire l'ID de la requête
    ids, ok := r.URL.Query()["id"]
    if !ok || len(ids[0]) < 1 {
        http.Error(w, "ID de la ressource manquant", http.StatusBadRequest)
        return
    }
    id := ids[0]

    // Supposons que vous avez une fonction qui peut déterminer le type de la ressource (article, rapport, blog) et récupérer les détails en fonction de l'ID.
    ressourceDetails, err := fetchRessourceDetails(id)
    if err != nil {
        http.Error(w, "Échec de la récupération des détails de la ressource", http.StatusInternalServerError)
        return
    }

    // Passez les détails récupérés au template de la page de détail
    templates.ExecuteTemplate(w, "ressource.html", ressourceDetails)
}

func fetchRessourceDetails(id string) (*PageData, error) {
    // Liste des endpoints à essayer. Vous devez les remplacer par les vrais endpoints.
    endpoints := []string{
        "https://api.spaceflightnewsapi.net/v3/articles//%s",
        "https://api.spaceflightnewsapi.net/v3/reports/%s",
        "https://api.spaceflightnewsapi.net/v3/blogs//%s",
    }

    httpClient := &http.Client{Timeout: time.Second * 10}
    var data PageData

    // Essayez chaque endpoint jusqu'à ce que vous trouviez la ressource ou que vous épuisiez les options.
    for _, endpoint := range endpoints {
        url := fmt.Sprintf(endpoint, id)
        fmt.Println("Tentative avec URL:", url) // Ajout de log pour débogage
        req, err := http.NewRequest(http.MethodGet, url, nil)
        if err != nil {
            fmt.Println("Erreur lors de la création de la requête :", err)
            continue
        }
    
        res, err := httpClient.Do(req)
        if err != nil || res.StatusCode != http.StatusOK {
            fmt.Println("Erreur ou code d'état HTTP non-OK :", res.StatusCode)
            continue
        }
        defer res.Body.Close()
    
        if err := json.NewDecoder(res.Body).Decode(&data); err == nil {
            return &data, nil
        }
        // Ajoutez un log ici si le décodage échoue, pour voir que la requête a réussi mais le décodage non.
    }
    

    return nil, fmt.Errorf("la ressource avec l'ID %s n'a pas été trouvée", id)
}

func accueilHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "accueil.html", r)
}

func connexionHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "connexion.html", r)
}

func favorisHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "favoris.html", r)
}

func rechercheHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "recherche.html", r)
}

func aproposHandler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "apropos.html", r)
}

func erreur404Handler(w http.ResponseWriter, r *http.Request) {

	templates.ExecuteTemplate(w, "erreur404.html", r)
}
























































































/*

// Struct pour les données combinées de tous les endpoints
type CombinedData struct {
    Data1 *PageData
    Data2 *PageData
    Data3 *PageData
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
    // Gestion des fichiers statiques (assets)
    rootDoc, _ := os.Getwd()
	fileserver := http.FileServer(http.Dir(rootDoc + "/asset"))
	http.Handle("/static/", http.StripPrefix("/static/", fileserver))

	// Analyser tous les fichiers de modèle correspondants à un modèle de nom de fichier particulier
    tmpl, err := template.ParseGlob("asset/templates/*.html")
   	if err != nil {
	    log.Fatal(err)
	}

    // Définir les routes
    http.HandleFunc("/accueil", func(w http.ResponseWriter, r *http.Request) {
        renderTemplate(w, tmpl, "accueil", &PageData{Title: "accueil"})
    })

    http.HandleFunc("/connection", func(w http.ResponseWriter, r *http.Request) {
        renderTemplate(w, tmpl,  "connection", &PageData{Title: "Connection"})
    })

    http.HandleFunc("/collection", func(w http.ResponseWriter, r *http.Request) {
        // Récupérer les données depuis les endpoints API
        combinedData, err := fetchData()
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
            return
        }

        // Afficher les données dans le template
        renderTemplate(w, tmpl, "collection", combinedData.Data1)
    })

    http.HandleFunc("/categorie", func(w http.ResponseWriter, r *http.Request) {
        // Récupérer les données depuis les endpoints API
        combinedData, err := fetchData()
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
            return
        }

        // Afficher les données dans le template
        renderTemplate(w, tmpl, "categorie", combinedData.Data2)
    })

    http.HandleFunc("/ressource", func(w http.ResponseWriter, r *http.Request) {
        // Récupérer les données depuis les endpoints API
        combinedData, err := fetchData()
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
            return
        }

        // Afficher les données dans le template
        renderTemplate(w, tmpl, "ressource", combinedData.Data3)
    })

    http.HandleFunc("/favoris", func(w http.ResponseWriter, r *http.Request) {
        // Récupérer les données depuis les endpoints API
        combinedData, err := fetchData()
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
            return
        }

        // Afficher les données dans le template
        renderTemplate(w, tmpl, "favoris", combinedData.Data1) // Par exemple, vous pouvez utiliser combinedData.Data1 ou combinedData.Data2 ici
    })

    http.HandleFunc("/recherche", func(w http.ResponseWriter, r *http.Request) {
        // Récupérer les données depuis les endpoints API
        combinedData, err := fetchData()
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
            return
        }

        // Afficher les données dans le template
        renderTemplate(w, tmpl, "recherche", combinedData.Data2) // Par exemple, vous pouvez utiliser combinedData.Data1 ou combinedData.Data2 ici
    })

    http.HandleFunc("/apropos", func(w http.ResponseWriter, r *http.Request) {
        renderTemplate(w, tmpl, "accueil", &PageData{Title: "A propos"})
    })

    // Page 404 - si aucune route ne correspond
    http.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
        renderTemplate(w, tmpl,  "accueil", &PageData{Title: "Page non trouvée"})
    })

    // Démarrer le serveur sur le port 7070
    http.ListenAndServe(":7070", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, tmplName string, data *PageData) {
    // Exécuter le template avec les données
    err := tmpl.ExecuteTemplate(w, tmplName+".html", data)
    if err != nil {
        log.Printf("Erreur lors du rendu du template %s: %v", tmplName, err) // Ajoutez ce log pour capturer les erreurs
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

// Fonction pour récupérer les données depuis les endpoints API
func fetchData() (CombinedData, error) {
    // Envoyer une requête HTTP GET à chaque endpoint API
    resp1, err := http.Get("https://api.spaceflightnewsapi.net/v3/articles")
    if err != nil {
        return CombinedData{}, err
    }
    defer resp1.Body.Close()

    resp2, err := http.Get("https://api.spaceflightnewsapi.net/v3/reports")
    if err != nil {
        return CombinedData{}, err
    }
    defer resp2.Body.Close()

    resp3, err := http.Get("https://api.spaceflightnewsapi.net/v3/blogs")
    if err != nil {
        return CombinedData{}, err
    }
    defer resp3.Body.Close()

    // Analyser les réponses JSON et combiner les données
    var data1 PageData
    if err := json.NewDecoder(resp1.Body).Decode(&data1); err != nil {
        return CombinedData{}, err
    }

    var data2 PageData
    if err := json.NewDecoder(resp2.Body).Decode(&data2); err != nil {
        return CombinedData{}, err
    }

    var data3 PageData
    if err := json.NewDecoder(resp3.Body).Decode(&data3); err != nil {
        return CombinedData{}, err
    }

    combinedData := CombinedData{
        Data1: &data1,
        Data2: &data2,
        Data3: &data3,
    }
    return combinedData, nil
}

*/