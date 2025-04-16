package main

import (
	"html/template"
	"log"
	"net/http"
)

type Message struct {
	Id   int
	Text string
}

func main() {

	data := map[string][]Message{
		"Messages": {
			Message{Id: 1, Text: "Hello"},
		},
	}
	//static asset correctly resolve in static directory, server know how to resolve requests to image and css files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		//retrive a file from disc
		tmpl := template.Must(template.ParseFiles("./templates/index.html"))
		//compile it and send it as a response
		tmpl.Execute(w, nil)
	}
	sendMessageHandler := func(w http.ResponseWriter, r *http.Request) {
		messageBox := r.PostFormValue("message-box")
		tmpl := template.Must(template.ParseFiles("./templates/main.html"))
		message := Message{Id: 1, Text: messageBox}
		data["Messages"] = append(data["Messages"], message)
		tmpl.ExecuteTemplate(w, "message-element", message)
	}
	signInHandler := func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./templates/signin.html"))
		tmpl.Execute(w, nil)
	}
	signUpHandler := func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./templates/signup.html"))
		tmpl.Execute(w, nil)
	}
	mainHandler := func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("./templates/main.html"))
		tmpl.Execute(w, data)
	}

	//call to root path
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/send-message", sendMessageHandler)
	http.HandleFunc("/signin", signInHandler)
	http.HandleFunc("/signup", signUpHandler)
	http.HandleFunc("/main", mainHandler)

	log.Println("App running on 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
