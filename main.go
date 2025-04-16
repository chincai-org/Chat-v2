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
		tmpl.Execute(w, data)
	}
	sendMessageHandler := func(w http.ResponseWriter, r *http.Request) {
		messageBox := r.PostFormValue("message-box")
		tmpl := template.Must(template.ParseFiles("./templates/main.html"))
		message := Message{Id: 1, Text: messageBox}
		data["Messages"] = append(data["Messages"], message)
		tmpl.ExecuteTemplate(w, "message-element", message)
	}

	//call to root path
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/send-message", sendMessageHandler)

	log.Println("App running on 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
