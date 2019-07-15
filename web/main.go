package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"

	pb "github.com/kuwuda/guild_management/api"
	"github.com/kuwuda/guild_management/client"
	"google.golang.org/grpc"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile     = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr = flag.String("server_addr", "127.0.0.1:50051", "The server address in the format of host:port")
	addr       = flag.String("addr", ":8080", "http service address")
)

var indexTemplate = template.Must(template.ParseFiles("web/templates/index.html"))
var conn *grpc.ClientConn

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var deleteRequests []*pb.DeleteRequest
	var deleteRequest pb.DeleteRequest

	deleteRequest.Name = r.FormValue("name")

	deleteRequests = append(deleteRequests, &deleteRequest)

	resp, err := client.DeleteMembers(conn, deleteRequests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(resp)

	http.Redirect(w, r, "/", 301)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := client.GetKeys(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.ParseForm()

	activity := &pb.ActivityItem{Name: r.FormValue("name")}
	activity.Activities = make(map[string]uint32)
	for _, v := range keys {
		var n int
		n, err = strconv.Atoi(r.FormValue(v))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		activity.Activities[v] = uint32(n)
	}

	var activities []*pb.ActivityItem
	activities = append(activities, activity)

	resp, err := client.UpdateMembers(conn, activities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(resp)

	http.Redirect(w, r, "/", 301)
}

func colHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.FormValue("key")
	var keys []string

	keys = append(keys, key)
	resp, err := client.AddColumns(conn, keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(resp)

	http.Redirect(w, r, "/", 301)
}

func rowHandler(w http.ResponseWriter, r *http.Request) {
	keys, err := client.GetKeys(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.ParseForm()

	activity := &pb.ActivityItem{Name: r.FormValue("name")}
	activity.Activities = make(map[string]uint32)
	for _, v := range keys {
		var n int
		n, err = strconv.Atoi(r.FormValue(v))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		activity.Activities[v] = uint32(n)
	}

	var activities []*pb.ActivityItem
	activities = append(activities, activity)

	resp, err := client.WriteMembers(conn, activities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(resp)

	http.Redirect(w, r, "/", 301)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var request pb.ActivityRequest
	activities, err := client.GetActivities(conn, &request)
	err = indexTemplate.Execute(w, activities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	var err error
	conn, err = grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/x-javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)

	fs := http.FileServer(http.Dir("web/static/"))
	http.Handle("/static/", m.Middleware(http.StripPrefix("/static/", fs)))

	// REST api
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/col", colHandler)
	http.HandleFunc("/row", rowHandler)

	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
