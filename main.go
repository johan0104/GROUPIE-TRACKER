package main

import (
	"html/template"
	"net/http"
)

// Struct pour les données que vous souhaitez passer à vos templates
type PageData struct {
	Title string
}

func main() {
	// Gestion des fichiers statiques (assets)
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Définir les routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "accueil", PageData{Title: "Accueil"})
	})

	http.HandleFunc("/connection", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "connection", PageData{Title: "Connexion"})
	})

	http.HandleFunc("/collection", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "collection", PageData{Title: "Collection"})
	})

	http.HandleFunc("/categorie", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "categorie", PageData{Title: "Catégorie"})
	})

	http.HandleFunc("/ressource", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "ressource", PageData{Title: "Ressource"})
	})

	http.HandleFunc("/favoris", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "favoris", PageData{Title: "Favoris"})
	})

	http.HandleFunc("/recherche", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "recherche", PageData{Title: "Recherche"})
	})

	http.HandleFunc("/apropos", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "apropos", PageData{Title: "À Propos"})
	})

	// Page 404 - si aucune route ne correspond
	http.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		renderTemplate(w, "404", PageData{Title: "Page non trouvée"})
	})

	// Démarrer le serveur sur le port 7070
	http.ListenAndServe(":7070", nil)
}

func renderTemplate(w http.ResponseWriter, tmplName string, data PageData) {
	// Analyser le fichier de template
	tmpl, err := template.ParseFiles(tmplName + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Exécuter le template avec les données
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}